-- 006_add_campaign_master.sql
-- #81: プレミアムフライデー等の割引キャンペーンマスタ。
-- payment_methods / trimming_course_types と同型の拡張可能マスタ + 対象指定の中間テーブル。
-- Q1=D: 割引対象は「カテゴリ単位指定」+「個別商品指定」の併用。

CREATE TYPE campaign_discount_type AS ENUM ('rate', 'amount');

CREATE TABLE campaigns (
    id              BIGSERIAL              PRIMARY KEY,
    clinic_id       bigint                 NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name            text                   NOT NULL DEFAULT '',
    start_date      date                   NOT NULL,
    end_date        date                   NOT NULL,
    discount_type   campaign_discount_type NOT NULL DEFAULT 'rate',
    discount_value  numeric(12,2)          NOT NULL DEFAULT 0,  -- rate: 率(%), amount: 額(円)
    is_active       boolean                NOT NULL DEFAULT true,
    sort_order      integer                NOT NULL DEFAULT 0,
    created_at      timestamptz            NOT NULL DEFAULT now(),
    updated_at      timestamptz            NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    CONSTRAINT chk_campaigns_period CHECK (end_date >= start_date)
);

-- Q1=D: カテゴリ単位の対象指定
CREATE TABLE campaign_target_categories (
    id          BIGSERIAL     PRIMARY KEY,
    campaign_id bigint        NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    category    item_category NOT NULL,
    CONSTRAINT uq_campaign_target_categories UNIQUE (campaign_id, category)
);

-- Q1=D: 個別商品の対象指定
CREATE TABLE campaign_target_items (
    id                  BIGSERIAL PRIMARY KEY,
    campaign_id         bigint    NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    merchandise_item_id bigint    NOT NULL REFERENCES merchandise_items(id) ON DELETE CASCADE,
    CONSTRAINT uq_campaign_target_items UNIQUE (campaign_id, merchandise_item_id)
);

CREATE INDEX idx_campaigns_clinic ON campaigns(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaigns_clinic_period ON campaigns(clinic_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaign_target_categories_campaign ON campaign_target_categories(campaign_id);
CREATE INDEX idx_campaign_target_items_campaign ON campaign_target_items(campaign_id);
CREATE INDEX idx_campaign_target_items_merchandise ON campaign_target_items(merchandise_item_id);

COMMENT ON TABLE campaigns IS '#81: 割引キャンペーンマスタ(期間・割引種別/値)';
COMMENT ON TABLE campaign_target_categories IS '#81: キャンペーン対象カテゴリ(Q1=D カテゴリ単位指定)';
COMMENT ON TABLE campaign_target_items IS '#81: キャンペーン対象商品(Q1=D 個別商品指定)';
