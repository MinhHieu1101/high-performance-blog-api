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
  <img width="851" height="281" alt="image" src="https://github.com/user-attachments/assets/c5f35c9c-4e8f-4288-b1dd-fe53ff0d1acc" />

- **Get a post:**
  ```sh
  curl http://localhost:8080/posts/1
  ```
  <img width="850" height="293" alt="image" src="https://github.com/user-attachments/assets/3eacfd77-0ce7-489f-b44d-d03a0fa3b3d3" />

- **Update a post:**
  ```sh
  curl -X PUT http://localhost:8080/posts/1 \
    -H "Content-Type: application/json" \
    -d '{"title":"Updated","content":"Fairly adjusted content","tags":["note","update"]}'
  ```
  <img width="845" height="284" alt="image" src="https://github.com/user-attachments/assets/c1ab6ff7-d3c0-4fc5-8e35-f4ff31d0ffba" />

- **Search by tag:**
  ```sh
  curl "http://localhost:8080/posts/search-by-tag?tag=note"
  ```
  <img width="853" height="480" alt="image" src="https://github.com/user-attachments/assets/3e5cc324-7abb-4ec2-be67-57a4306b3a1e" />

- **Full-text search (Elasticsearch):**
  ```sh
  curl "http://localhost:8080/posts/search?q=fishing"
  ```
  <img width="854" height="331" alt="image" src="https://github.com/user-attachments/assets/cfb08b91-c35c-4ad3-aa79-9f8cac6c8cdb" />

- **List all posts:**
  ```sh
  curl http://localhost:8080/posts
  ```
  <img width="850" height="789" alt="image" src="https://github.com/user-attachments/assets/bdae392d-eeb1-4152-a88a-2ee1574f10f2" />

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
