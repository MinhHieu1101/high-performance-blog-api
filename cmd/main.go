package main

import (
    "log"

    "blog-api/internal"
    "github.com/gin-gonic/gin"
)

func main() {
    // initialize services (DB, Redis, ES)
    if err := internal.InitServices(); err != nil {
        log.Fatalf("failed to init services: %v", err)
    }

    // register HTTP routes
    r := gin.Default()
    internal.RegisterRoutes(r)

    log.Println("listening :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
