package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// mount the HTTP handlers on the provided router
func RegisterRoutes(r *gin.Engine) {
	r.POST("/posts", createPostHandler)
	r.GET("/posts/:id", getPostHandler)
	r.PUT("/posts/:id", updatePostHandler)
	r.GET("/posts/search-by-tag", searchByTagHandler)
	r.GET("/posts/search", searchHandler)
	r.GET("/posts", listPostsHandler)
	r.POST("/internal/reindex", reindexHandler)
}

// -- create post: transaction + ES indexing
func createPostHandler(c *gin.Context) {
	var req struct {
		Title   string   `json:"title" binding:"required"`
		Content string   `json:"content" binding:"required"`
		Tags    []string `json:"tags"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		newPost := Post{Title: req.Title, Content: req.Content, Tags: req.Tags}
		if err := tx.Create(&newPost).Error; err != nil {
			return err
		}
		logEntry := ActivityLog{Action: "new_post", PostID: newPost.ID}
		if err := tx.Create(&logEntry).Error; err != nil {
			return err
		}
		// try to index in ES; if it fails, return error to roll back DB as well
		if err := indexPostToES(&newPost); err != nil {
			return err
		}
		c.JSON(http.StatusCreated, newPost)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// -- get post: cache-aside
/* func getPostHandler(c *gin.Context) {
	idStr := c.Param("id")
	cacheKey := "post:" + idStr

	// try redis first
	val, err := RedisClient.Get(AppCtx, cacheKey).Result()
	if err == nil {
		var p Post
		if err := json.Unmarshal([]byte(val), &p); err == nil {
			c.JSON(http.StatusOK, p)
			return
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var p Post
	if err := DB.First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	b, _ := json.Marshal(p)
	RedisClient.Set(AppCtx, cacheKey, b, cacheTTL)
	c.JSON(http.StatusOK, p)
} */

// -- get similar posts: return the post and a small list of related posts by tags
func getPostHandler(c *gin.Context) {
	idStr := c.Param("id")
	cacheKey := "post:" + idStr

	// try redis first
	val, err := RedisClient.Get(AppCtx, cacheKey).Result()
	var p Post
	if err == nil {
		// cache hit
		if err := json.Unmarshal([]byte(val), &p); err == nil {
			// get related posts and return combined
			related, _ := findRelatedPostsByTags(p.Tags, p.ID, 5)
			c.JSON(http.StatusOK, gin.H{"post": p, "related": related})
			return
		}
	}

	// cache miss -> load from DB
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}
	if err := DB.First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	b, _ := json.Marshal(p)
	RedisClient.Set(AppCtx, cacheKey, b, cacheTTL)

	related, _ := findRelatedPostsByTags(p.Tags, p.ID, 5)

	c.JSON(http.StatusOK, gin.H{"post": p, "related": related})
}

// -- update post: invalidate cache + reindex
func updatePostHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Title   *string  `json:"title"`
		Content *string  `json:"content"`
		Tags    []string `json:"tags"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var p Post
	if err := DB.First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.Content != nil {
		p.Content = *req.Content
	}
	if req.Tags != nil {
		p.Tags = req.Tags
	}

	if err := DB.Save(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// invalidate cache
	RedisClient.Del(AppCtx, "post:"+idStr)

	// reindex
	if err := indexPostToES(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "updated but failed to index to ES: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// -- search by tag using Postgres array operator
func searchByTagHandler(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag query param required"})
		return
	}

	var posts []Post
	if err := DB.Where("? = ANY (tags)", tag).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// -- full text search via ES
func searchHandler(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q required"})
		return
	}
	query := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{"query": q, "fields": []string{"title", "content"}},
		},
	}
	b, _ := json.Marshal(query)
	res, err := ESClient.Search(ESClient.Search.WithBody(bytes.NewReader(b)), ESClient.Search.WithIndex(esIndex), ESClient.Search.WithSize(20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer res.Body.Close()

	var esResp map[string]any
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid es response"})
		return
	}

	hits := []any{}
	if hitsObj, ok := esResp["hits"].(map[string]any); ok {
		if hitArr, ok := hitsObj["hits"].([]any); ok {
			for _, h := range hitArr {
				if hm, ok := h.(map[string]any); ok {
					if src, exists := hm["_source"]; exists {
						hits = append(hits, src)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": hits})
}

// -- list posts
func listPostsHandler(c *gin.Context) {
	var posts []Post
	if err := DB.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// -- reindex handler: index all posts to ES
func reindexHandler(c *gin.Context) {
	var posts []Post
	if err := DB.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var failed int
	for _, p := range posts {
		if err := indexPostToES(&p); err != nil {
			failed++
		}
	}
	c.JSON(http.StatusOK, gin.H{"indexed": len(posts) - failed, "failed": failed})
}
