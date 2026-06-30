-- =============================================================================
-- 010_add_checkup_packages.sql
-- #211 検査・健診パッケージ化 — 型付きフィールド機構（垂直スライス: 歯科検診）
--
-- 設計: examination ドメイン（exam_types → exam_type_fields → exam_results,
--       001_init.sql:692/1680）のパターンを踏襲し、checkup 用に正規化する。
--   - anchor は既存 checkup_types を「パッケージ」として拡張（新テーブルは作らない）。
--   - checkup_type_fields  : パッケージのフィールド定義（型付き）。
--   - checkup_field_results: 健診記録（checkups）に紐づく結果値。
--
-- マルチテナント: 両テーブルとも clinic_id NOT NULL を持ち、RLS は
--   tenant_clinic_id 直接ポリシーで保護する（001_init の clinic_id 自動ループは
--   既適用済みのため、後発の本マイグレーションで明示的に apply_rls_policy する）。
--
-- CASCADE 判断: migrations/CLAUDE.md の「純粋従属子行は CASCADE 許容例外」に従う。
--   - checkup_type_fields.checkup_type_id  → exam_type_fields.exam_type_id と同型（構成要素）
--   - checkup_field_results.checkup_id     → exam_results.exam_id と同型（純粋従属の結果行）。
--     RESTRICT にすると既存 medical_records → checkups CASCADE 連鎖を壊すため CASCADE 必須。
--   - checkup_field_results.checkup_type_field_id は nullable + ON DELETE SET NULL とし、
--     field_name/field_type/unit を非正規化スナップショットとして結果行に保持する
--     （exam_results.exam_type_field_id と同型。フィールド定義削除後も結果が自己記述的）。
-- =============================================================================

-- ------------------------------------
-- フィールド型 ENUM（6種）
-- ------------------------------------
CREATE TYPE checkup_field_type AS ENUM (
    'number',
    'single_select',
    'multi_select',
    'boolean',
    'checklist',
    'text'
);

-- ------------------------------------
-- checkup_type_fields（健診パッケージのフィールド定義マスタ）
-- ------------------------------------
CREATE TABLE checkup_type_fields (
    id              BIGSERIAL          PRIMARY KEY,
    clinic_id       bigint             NOT NULL REFERENCES clinics(id)       ON DELETE RESTRICT,
    checkup_type_id bigint             NOT NULL REFERENCES checkup_types(id) ON DELETE CASCADE,
    name            text               NOT NULL,
    field_type      checkup_field_type NOT NULL,
    unit            text               NOT NULL DEFAULT '',
    -- number 型の異常値判定基準（EXAM-001 と同じく min/max は任意）
    min_value       decimal(10,4),
    max_value       decimal(10,4),
    -- single_select / multi_select / checklist の選択肢定義: [{"value":"...","label":"..."}]
    options         jsonb              NOT NULL DEFAULT '[]'::jsonb,
    -- 暫定 seed フラグ（確定値は Notion 反映の別タスク）
    is_provisional  boolean            NOT NULL DEFAULT false,
    sort_order      integer            NOT NULL DEFAULT 0,
    created_at      timestamptz        NOT NULL DEFAULT now(),
    updated_at      timestamptz        NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- ------------------------------------
-- checkup_field_results（健診結果値: checkups の純粋従属子）
-- ------------------------------------
CREATE TABLE checkup_field_results (
    id                    BIGSERIAL          PRIMARY KEY,
    clinic_id             bigint             NOT NULL REFERENCES clinics(id)   ON DELETE RESTRICT,
    checkup_id            bigint             NOT NULL REFERENCES checkups(id)  ON DELETE CASCADE,
    checkup_type_field_id bigint                      REFERENCES checkup_type_fields(id) ON DELETE SET NULL,
    -- 非正規化スナップショット（フィールド定義削除後も結果が自己記述的であるため）
    field_name            text               NOT NULL DEFAULT '',
    field_type            checkup_field_type NOT NULL,
    unit                  text               NOT NULL DEFAULT '',
    -- 型別の値カラム（field_type に応じてサーバが該当列のみ書き込む）
    value_number          decimal(10,4),
    value_text            text               NOT NULL DEFAULT '',
    value_bool            boolean,
    value_list            text[]             NOT NULL DEFAULT '{}',
    -- number 型の異常値判定（EXAM-001 機構を再利用。exam_result_status を共用）
    ref_min               decimal(10,4),
    ref_max               decimal(10,4),
    is_abnormal           boolean            NOT NULL DEFAULT false,
    status                exam_result_status NOT NULL DEFAULT 'normal',
    sort_order            integer            NOT NULL DEFAULT 0,
    created_at            timestamptz        NOT NULL DEFAULT now(),
    updated_at            timestamptz        NOT NULL DEFAULT now()
);

-- ------------------------------------
-- インデックス（clinic_id を含む複合 + FK）
-- ------------------------------------
CREATE INDEX idx_checkup_type_fields_clinic_id        ON checkup_type_fields(clinic_id);
CREATE INDEX idx_checkup_type_fields_checkup_type_id  ON checkup_type_fields(checkup_type_id);
CREATE INDEX idx_checkup_type_fields_clinic_type_sort ON checkup_type_fields(clinic_id, checkup_type_id, sort_order) WHERE deleted_at IS NULL;

-- FindByCheckupID / ReplaceForCheckup はともに WHERE clinic_id = ? AND checkup_id = ? を発行する。
-- clinic_id 先頭（等値）+ checkup_id（等値）の複合で両クエリを単一インデックスで賄う
-- （migrations/CLAUDE.md「clinic_id を含む複合インデックス」規約）。clinic_id 単独はこの複合の前方一致で代替できるため作らない。
CREATE INDEX idx_checkup_field_results_clinic_checkup ON checkup_field_results(clinic_id, checkup_id);
-- checkup_id 単独は FindByPetID の JOIN（checkups.id = checkup_field_results.checkup_id）用に保持する。
CREATE INDEX idx_checkup_field_results_checkup_id      ON checkup_field_results(checkup_id);
-- checkup_type_field_id は ON DELETE SET NULL で NULL 行が増えるため部分インデックス。
CREATE INDEX idx_checkup_field_results_field_id        ON checkup_field_results(checkup_type_field_id) WHERE checkup_type_field_id IS NOT NULL;

-- ------------------------------------
-- RLS（clinic_id 直接ポリシー。001_init の自動ループ相当を後発で明示適用）
-- ------------------------------------
SELECT app_private.apply_rls_policy(
    'checkup_type_fields',
    'tenant_checkup_type_fields_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'checkup_field_results',
    'tenant_checkup_field_results_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

-- ------------------------------------
-- COMMENT
-- ------------------------------------
COMMENT ON TABLE checkup_type_fields    IS '健診パッケージの型付きフィールド定義マスタ（#211）';
COMMENT ON TABLE checkup_field_results  IS '健診結果値（checkups の純粋従属子・#211）';
COMMENT ON COLUMN checkup_type_fields.options       IS 'select/checklist の選択肢定義 [{"value","label"}]';
COMMENT ON COLUMN checkup_type_fields.is_provisional IS '暫定 seed フラグ（確定値は別タスク）';
COMMENT ON COLUMN checkup_field_results.field_type  IS 'checkup_type_fields.field_type の非正規化スナップショット';

-- =============================================================================
-- 歯科検診パッケージの暫定 seed（provisional）
--   既存の '歯科検診' checkup_type を持つ各クリニックにフィールドを投入する。
--   名前マッチ + NOT EXISTS で冪等・環境非依存（demo 未投入環境では no-op）。
--   確定の選択肢・基準値は Notion「健康診断」反映の別タスク（is_provisional=true）。
-- =============================================================================
DO $$
DECLARE
    ct RECORD;
BEGIN
    FOR ct IN
        SELECT id, clinic_id FROM checkup_types
        WHERE name = '歯科検診' AND deleted_at IS NULL
    LOOP
        -- 歯石除去必要の有無（boolean）
        INSERT INTO checkup_type_fields (clinic_id, checkup_type_id, name, field_type, is_provisional, sort_order)
        SELECT ct.clinic_id, ct.id, '歯石除去必要の有無', 'boolean', true, 1
        WHERE NOT EXISTS (
            SELECT 1 FROM checkup_type_fields f
            WHERE f.checkup_type_id = ct.id AND f.name = '歯石除去必要の有無' AND f.deleted_at IS NULL
        );

        -- 歯石付着度スコア（number・0〜4。max 超過で異常値判定）
        INSERT INTO checkup_type_fields (clinic_id, checkup_type_id, name, field_type, unit, min_value, max_value, is_provisional, sort_order)
        SELECT ct.clinic_id, ct.id, '歯石付着度スコア', 'number', '', 0, 4, true, 2
        WHERE NOT EXISTS (
            SELECT 1 FROM checkup_type_fields f
            WHERE f.checkup_type_id = ct.id AND f.name = '歯石付着度スコア' AND f.deleted_at IS NULL
        );

        -- 歯科ケアアドバイス（multi_select・クライアント文面由来の暫定選択肢）
        -- 暫定 seed のため value == label（日本語）とし、結果値（value_list）が
        -- そのまま飼主レポートで人間可読に表示されるようにする。確定キー体系は
        -- マスタ管理 UI イテレーションで value≠label を導入する（その際は owner-report 側で
        -- options からのラベル逆引きが必要・follow-up）。
        INSERT INTO checkup_type_fields (clinic_id, checkup_type_id, name, field_type, options, is_provisional, sort_order)
        SELECT ct.clinic_id, ct.id, '歯科ケアアドバイス', 'multi_select',
            '[{"value":"毎日の歯磨き","label":"毎日の歯磨き"},
              {"value":"デンタルガムの利用","label":"デンタルガムの利用"},
              {"value":"定期的なスケーリング","label":"定期的なスケーリング"},
              {"value":"歯科用フードへの切替","label":"歯科用フードへの切替"},
              {"value":"経過観察","label":"経過観察"}]'::jsonb,
            true, 3
        WHERE NOT EXISTS (
            SELECT 1 FROM checkup_type_fields f
            WHERE f.checkup_type_id = ct.id AND f.name = '歯科ケアアドバイス' AND f.deleted_at IS NULL
        );
    END LOOP;
END;
$$;
