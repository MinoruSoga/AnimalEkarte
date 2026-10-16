package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/config"
)

func TestNewSupportEmbedderGates(t *testing.T) {
	base := &config.Config{
		SupportHybridSearchEnabled:  true,
		SupportExternalDataApproved: true,
		SupportEmbeddingBaseURL:     "https://api.sarashina.ai/v1",
		SupportEmbeddingAPIKey:      "k",
		SupportEmbeddingModel:       "m",
		SupportEmbeddingTimeoutMS:   1000,
	}

	tests := map[string]struct {
		mutate func(*config.Config)
		want   bool
	}{
		"approved sarashina endpoint builds embedder": {
			mutate: func(*config.Config) {},
			want:   true,
		},
		"hybrid disabled": {
			mutate: func(c *config.Config) { c.SupportHybridSearchEnabled = false },
			want:   false,
		},
		"external data unapproved": {
			mutate: func(c *config.Config) { c.SupportExternalDataApproved = false },
			want:   false,
		},
		"non sarashina endpoint": {
			mutate: func(c *config.Config) { c.SupportEmbeddingBaseURL = "https://api.openai.com/v1" },
			want:   false,
		},
		"sarashina lookalike host": {
			mutate: func(c *config.Config) { c.SupportEmbeddingBaseURL = "https://api.sarashina.ai.evil.com/v1" },
			want:   false,
		},
		"empty key still nil": {
			mutate: func(c *config.Config) { c.SupportEmbeddingAPIKey = "" },
			want:   false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := *base
			tt.mutate(&cfg)
			if tt.want {
				assert.NotNil(t, newSupportEmbedder(&cfg))
			} else {
				assert.Nil(t, newSupportEmbedder(&cfg))
			}
		})
	}
}
