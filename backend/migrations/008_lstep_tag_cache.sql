-- LSTEP-BE-020 (DB part): Lステップタグキャッシュテーブル
-- カルテDB側にタグキャッシュを持つことでLステップ外部API依存を最小化し、
-- タグ集計・タグ別飼い主検索を高速に実行できるようにする。

CREATE TABLE lstep_tag_cache (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   BIGINT      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    owner_id    BIGINT      NOT NULL REFERENCES owners(id)  ON DELETE CASCADE,
    tag_name    VARCHAR(100) NOT NULL,
    category    VARCHAR(20) NOT NULL DEFAULT 'auto'
                CHECK (category IN ('auto', 'manual')),
    synced_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (clinic_id, owner_id, tag_name)
);

-- タグ名での集計・検索用（BE-020 集計API）
CREATE INDEX idx_lstep_tag_cache_clinic_tag ON lstep_tag_cache (clinic_id, tag_name);
-- 飼い主のタグ一覧取得用（BE-019 タグCRUD）
CREATE INDEX idx_lstep_tag_cache_owner ON lstep_tag_cache (clinic_id, owner_id);

COMMENT ON TABLE lstep_tag_cache IS 'Lステップタグのカルテ側キャッシュ。タグ操作ごとにupsert/deleteして同期を保つ。';
COMMENT ON COLUMN lstep_tag_cache.category IS 'auto=各Sync*メソッドが自動付与, manual=BE-019経由でスタッフが手動付与';
