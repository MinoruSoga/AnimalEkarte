package support

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// supportManualSearchDDL mirrors the support_manual_embeddings / support_llm_budgets
// tables consolidated into migrations/001_init.sql (formerly 038_support_manual_search.sql).
// The shared test DB is built with GORM AutoMigrate, so these derived-cache tables
// must be created explicitly for the real-DB search tests.
const supportManualSearchDDL = `
CREATE TABLE IF NOT EXISTS support_manual_embeddings (
    cache_key       text PRIMARY KEY,
    model           text NOT NULL,
    task_type       text NOT NULL,
    chunker_version integer NOT NULL,
    dimensions      integer NOT NULL,
    embedding       real[] NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT support_manual_embeddings_dimensions_chk CHECK (dimensions = 1536),
    CONSTRAINT support_manual_embeddings_length_chk CHECK (
        array_ndims(embedding) = 1 AND cardinality(embedding) = 1536 AND array_lower(embedding, 1) = 1
    ),
    CONSTRAINT support_manual_embeddings_task_chk CHECK (task_type = 'RETRIEVAL_DOCUMENT'),
    CONSTRAINT support_manual_embeddings_version_chk CHECK (chunker_version = 1)
);

CREATE TABLE IF NOT EXISTS support_llm_budgets (
    scope            text NOT NULL,
    model            text NOT NULL,
    next_allowed_at  timestamptz NOT NULL,
    PRIMARY KEY (scope, model)
);
`

func ensureSupportManualSearchTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(supportManualSearchDDL).Error)
}
