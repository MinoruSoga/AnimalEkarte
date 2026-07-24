package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/config"
)

func TestNewHTTPServerBoundsRequestHeaders(t *testing.T) {
	server := newHTTPServer(
		&config.Config{Port: "8080"},
		gin.New(),
	)

	assert.Equal(t, 32<<10, server.MaxHeaderBytes)
}
