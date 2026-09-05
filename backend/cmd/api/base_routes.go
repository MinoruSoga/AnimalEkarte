package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const uploadsDirectory = "/app/uploads"

func healthOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// registerBaseRoutes installs the non-domain HTTP surface. Domain routes are
// registered separately after auth creates the protected API group.
//
// /uploads is a public unauthenticated static surface for local/dev media only
// (STORAGE_TYPE empty). Release forces STORAGE_TYPE=s3 so StaticFS is not mounted.
//
// /_internal/scheduled-jobs is an application-level privileged surface (DEC-36):
// callers must present X-Scheduler-Token matching SCHEDULER_INTERNAL_TOKEN.
// Edge topology alone is not sufficient defense-in-depth.
func registerBaseRoutes(
	router *gin.Engine,
	scheduledBatch scheduledBatchService,
) error {
	if router == nil {
		return fmt.Errorf("base route engine is required")
	}

	router.GET("/health", healthOK)
	router.GET("/api/v1/health", healthOK)
	// CMD-05: do not expose local upload PHI via StaticFS when object storage is configured.
	if os.Getenv("STORAGE_TYPE") != "s3" {
		router.StaticFS("/uploads", gin.Dir(uploadsDirectory, false))
	}
	registerScheduledJobRoutes(router, scheduledBatch, os.Getenv("SCHEDULER_INTERNAL_TOKEN"))
	return nil
}
