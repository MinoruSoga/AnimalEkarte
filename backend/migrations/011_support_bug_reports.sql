-- support_bug_reports: アプリ内サポートウィジェットから送信されるバグ報告。
-- スクショ画像は FileUploader（ローカル /app/uploads または S3）に保存し、
-- このテーブルにはオブジェクト key のみ保持する（公開 URL は保存しない）。
-- Do not edit applied migrations. User must run: make migrate

CREATE TABLE support_bug_reports (
    id                BIGSERIAL   PRIMARY KEY,
    clinic_id         bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    reporter_staff_id bigint      NOT NULL REFERENCES staffs(id)  ON DELETE RESTRICT,
    title             text        NOT NULL,
    detail            text        NOT NULL DEFAULT '',
    page_url          text        NOT NULL DEFAULT '',
    route_path        text        NOT NULL DEFAULT '',
    user_agent        text        NOT NULL DEFAULT '',
    viewport          text        NOT NULL DEFAULT '',
    app_version       text        NOT NULL DEFAULT '',
    screenshot_key    text,
    status            varchar(20) NOT NULL DEFAULT 'open',  -- open | resolved
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

CREATE INDEX idx_support_bug_reports_clinic_created
    ON support_bug_reports (clinic_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_support_bug_reports_clinic_status
    ON support_bug_reports (clinic_id, status)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE support_bug_reports IS 'サポートウィジェットからのバグ報告（スクショ・画面文脈つき）';
COMMENT ON COLUMN support_bug_reports.screenshot_key IS 'FileUploader のオブジェクト key。配信は GetSignedURL 経由（公開 URL は保存しない）';
