#!/usr/bin/env bash
set -e
BASE="http://localhost:8080"

curl -s -X POST "$BASE/posts" -H "Content-Type: application/json" -d '{"title":"Go tips","content":"Tips for Go developers","tags":["go","programming"]}' | jq || true
curl -s -X POST "$BASE/posts" -H "Content-Type: application/json" -d '{"title":"Docker trick","content":"How to use docker-compose","tags":["devops","docker"]}' | jq || true
curl -s -X POST "$BASE/posts" -H "Content-Type: application/json" -d '{"title":"Testing in Go","content":"httptest and best practices","tags":["go","testing"]}' | jq || true

echo "seeded posts"
