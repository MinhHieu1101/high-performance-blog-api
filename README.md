# blog-api

## Quick Start

1. **Start services**  
   ```sh
   docker compose up -d
   ```

2. **Run the API server**  
   ```sh
   go run ./cmd
   ```

3. **Seed data (optional)**  
   ```sh
   ./seeder.sh
   ```
   Or manually run the SQL in [migrations/init.sql](migrations/init.sql).

4. **Reindex posts to Elasticsearch (after seeding SQL directly):**  
   ```sh
   curl -X POST http://localhost:8080/internal/reindex
   ```

## API Endpoints

- **Create Post:**  
  `POST /posts`  
  Body: `{ "title": "...", "content": "...", "tags": ["tag1", "tag2"] }`

- **Get Post (with Redis cache):**  
  `GET /posts/:id`

- **Update Post (invalidates cache, reindexes ES):**  
  `PUT /posts/:id`  
  Body: `{ "title": "...", "content": "...", "tags": ["tag1"] }`

- **Search by Tag (Postgres):**  
  `GET /posts/search-by-tag?tag=go`

- **Full-text Search (Elasticsearch):**  
  `GET /posts/search?q=your+query`

- **List All Posts:**  
  `GET /posts`

- **Reindex All Posts to Elasticsearch:**  
  `POST /internal/reindex`

## Testing

You can test the API using `curl` or any API client (e.g., Postman):

- **Create a post:**  
  ```sh
  curl -X POST http://localhost:8080/posts \
    -H "Content-Type: application/json" \
    -d '{"title":"Test","content":"Hello","tags":["go"]}'
  ```

- **Get a post:**  
  ```sh
  curl http://localhost:8080/posts/1
  ```

- **Update a post:**  
  ```sh
  curl -X PUT http://localhost:8080/posts/1 \
    -H "Content-Type: application/json" \
    -d '{"title":"Updated"}'
  ```

- **Search by tag:**  
  ```sh
  curl "http://localhost:8080/posts/search-by-tag?tag=go"
  ```

- **Full-text search:**  
  ```sh
  curl "http://localhost:8080/posts/search?q=go"
  ```

## Notes

- If you change the schema or want to reset the DB, remove the Docker volume:
  ```sh
  docker compose down
  docker volume rm blog-api_pgdata
  docker compose up -d
  ```

- All services (Postgres, Redis, Elasticsearch) must be running for the API to work.
