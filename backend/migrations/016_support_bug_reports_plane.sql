-- support_bug_reports に Plane 連携状態を保持するカラムを追加する。
-- 起票は報告保存後に best-effort で実行し、失敗時は報告自体を巻き戻さず
-- plane_sync_error に記録する（管理画面からの手動再送は
-- POST /support/bug-reports/:id/plane-ticket）。
-- Do not edit applied migrations. User must run: make migrate

ALTER TABLE support_bug_reports
    ADD COLUMN plane_issue_id   text,
    ADD COLUMN plane_issue_url  text,
    ADD COLUMN plane_sync_error text;

COMMENT ON COLUMN support_bug_reports.plane_issue_id IS 'Plane ワークアイテムの UUID（起票成功時のみ）。二重起票防止のクレーム対象';
COMMENT ON COLUMN support_bug_reports.plane_issue_url IS 'Plane チケットの表示用 URL（app.plane.so/<workspace>/browse/<IDENT>/）';
COMMENT ON COLUMN support_bug_reports.plane_sync_error IS '直近の Plane 起票失敗理由（成功・未試行時は NULL）。手動再送の手掛かり';
