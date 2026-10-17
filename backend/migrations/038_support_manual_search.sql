-- EMR-410: サポートチャット ハイブリッド検索（追加のみ・既存 migration 不変）
-- 再生成可能な派生キャッシュ + quota 状態。共有マニュアル由来のため clinic_id なし。
-- 本文・質問・スタッフ情報は保存しない（R3/R7）。

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
