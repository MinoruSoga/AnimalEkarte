package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const uploadsDirectory = "/app/uploads"

func healthOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// healthDB pings the database through the connection pool. The STG keep-alive
// cron hits this on a sub-ConnMaxIdleTime cadence so at least one pooled
// connection survives between ticks (fresh TLS+auth setup is otherwise paid by
// the first real request after an idle gap). The response body deliberately
// carries no detail — DB internals are not disclosed.
func healthDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
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
	db *gorm.DB,
) error {
	if router == nil {
		return fmt.Errorf("base route engine is required")
	}

	router.GET("/health", healthOK)
	router.GET("/api/v1/health", healthOK)
	router.GET("/health/db", healthDB(db))
	// CMD-05: do not expose local upload PHI via StaticFS when object storage is configured.
	if os.Getenv("STORAGE_TYPE") != "s3" {
		router.StaticFS("/uploads", gin.Dir(uploadsDirectory, false))
	}
	registerScheduledJobRoutes(router, scheduledBatch, os.Getenv("SCHEDULER_INTERNAL_TOKEN"))
	return nil
}
