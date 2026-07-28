package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const uploadsDirectory = "/app/uploads"

// registerBaseRoutes installs the non-domain HTTP surface. Domain routes are
// registered separately after auth creates the protected API group.
func registerBaseRoutes(
	router *gin.Engine,
	scheduledBatch scheduledBatchService,
) error {
	if router == nil {
		return fmt.Errorf("base route engine is required")
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.StaticFS("/uploads", gin.Dir(uploadsDirectory, false))
	registerScheduledJobRoutes(router, scheduledBatch)
	return nil
}
