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

3. **Seed data (for testing)**  
   ```sh
   chmod +x seeder.sh && ./seeder.sh
   ```

## Testing

You can test the API using `curl` or any API client (e.g., Postman):

- **Create a post:**  
  ```sh
  curl -X POST http://localhost:8080/posts \
    -H "Content-Type: application/json" \
    -d '{"title":"Howdy there","content":"Great day for fishing aint it?","tags":["note","greetings"]}'
  ```

- **Get a post:**  
  ```sh
  curl http://localhost:8080/posts/1
  ```

- **Update a post:**  
  ```sh
  curl -X PUT http://localhost:8080/posts/1 \
    -H "Content-Type: application/json" \
    -d '{"title":"Updated","content":"Fairly adjusted content","tags":["note","update"]}'
  ```

- **Search by tag:**  
  ```sh
  curl "http://localhost:8080/posts/search-by-tag?tag=note"
  ```

- **Full-text search (Elasticsearch):**  
  ```sh
  curl "http://localhost:8080/posts/search?q=fishing"
  ```

- **List all posts:**  
  ```sh
  curl http://localhost:8080/posts
  ```

- **Reindex all posts to Elasticsearch:**  
  ```sh
  curl -X POST http://localhost:8080/internal/reindex
  ```

## Notes

- To change the schema or reset the DB, remove the Docker volume:
  ```sh
  docker compose down
  docker volume rm blog-api_pgdata
  docker compose up -d
  ```

- All services (Postgres, Redis, Elasticsearch) must be running for the API to work.

- Verify only Docker Postgres is running
  ```sh
  netstat -ano | findstr 5432
  ```
