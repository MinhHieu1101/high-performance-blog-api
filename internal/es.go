package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// index a post document to the configured ES index
func indexPostToES(p *Post) error {
	body := map[string]any{
		"id":      p.ID,
		"title":   p.Title,
		"content": p.Content,
		"tags":    p.Tags,
	}
	b, _ := json.Marshal(body)
	res, err := ESClient.Index(esIndex, bytes.NewReader(b), ESClient.Index.WithDocumentID(fmt.Sprint(p.ID)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index error: %s", res.String())
	}
	return nil
}

// return `size` posts that share any of the provided tags, excluding the post with excludeID
func findRelatedPostsByTags(tags []string, excludeID uint, size int) ([]Post, error) {
	if len(tags) == 0 || size <= 0 {
		return nil, nil
	}

	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{"terms": map[string]any{"tags": tags}},
				},
				"must_not": []any{
					map[string]any{"term": map[string]any{"id": excludeID}},
				},
				"minimum_should_match": 1,
			},
		},
		"size": size,
	}

	b, _ := json.Marshal(query)
	res, err := ESClient.Search(
		ESClient.Search.WithBody(bytes.NewReader(b)),
		ESClient.Search.WithIndex(esIndex),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("es search error: %s", res.String())
	}

	var esResp map[string]any
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	var related []Post
	if hitsObj, ok := esResp["hits"].(map[string]any); ok {
		if hitArr, ok := hitsObj["hits"].([]any); ok {
			for _, h := range hitArr {
				if hm, ok := h.(map[string]any); ok {
					if src, exists := hm["_source"]; exists {
						// marshal then unmarshal into Post for safe typed conversion
						bs, _ := json.Marshal(src)
						var p Post
						if err := json.Unmarshal(bs, &p); err == nil {
							related = append(related, p)
						}
					}
				}
			}
		}
	}

	return related, nil
}
