CREATE TABLE IF NOT EXISTS posts (
  id serial PRIMARY KEY,
  title varchar NOT NULL,
  content text NOT NULL,
  tags text[] NOT NULL,
  created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_posts_tags_gin ON posts USING GIN (tags);

CREATE TABLE IF NOT EXISTS activity_logs (
  id serial PRIMARY KEY,
  action varchar NOT NULL,
  post_id integer REFERENCES posts(id) ON DELETE CASCADE,
  logged_at timestamptz DEFAULT now()
);

INSERT INTO posts (title, content, tags, created_at)
VALUES
  (
    'Go tips',
    'Tips for Go developers: struct tags, error wrapping and small interfaces.',
    ARRAY['go', 'programming'],
    now()
  ),
  (
    'Docker trick',
    'How to use docker-compose effectively for local development.',
    ARRAY['devops', 'docker'],
    now()
  ),
  (
    'Testing in Go',
    'Use httptest for handler tests and table-driven tests for logic.',
    ARRAY['go', 'testing'],
    now()
  );

INSERT INTO activity_logs (action, post_id, logged_at)
SELECT 'new_post', p.id, now()
FROM posts p
WHERE p.id IN (
  (SELECT max(id) FROM posts) - 2,
  (SELECT max(id) FROM posts) - 1,
  (SELECT max(id) FROM posts)
);
