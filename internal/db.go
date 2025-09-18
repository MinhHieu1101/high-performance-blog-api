package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
	ESClient    *elasticsearch.Client
	AppCtx      = context.Background()
	cacheTTL    = 5 * time.Minute
	esIndex     = "posts"
)

// set up DB, Redis and Elasticsearch clients and runs simple migrations
func InitServices() error {
	// Postgres DSN
	dsn := "host=localhost user=dev password=dev dbname=blogdb port=5432 sslmode=disable"
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	// automigrate
	if err := gdb.AutoMigrate(&Post{}, &ActivityLog{}); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	// ensure GIN index for tags
	if err := gdb.Exec(`CREATE INDEX IF NOT EXISTS idx_posts_tags_gin ON posts USING GIN (tags);`).Error; err != nil {
		log.Printf("warn: create index: %v", err)
	}

	DB = gdb

	// Redis
	RedisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := RedisClient.Ping(AppCtx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	// Elasticsearch
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("es client: %w", err)
	}
	ESClient = es

	return nil
}
