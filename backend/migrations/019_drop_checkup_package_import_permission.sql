-- EMR-249: checkup package versioned import API は 89ac9590c で削除済みだが、
-- 新規クリニック作成時の defaultPermissionRuleTable が付与していた
-- resource='checkup-package-import' の default-deny 行が permission_group_rules に
-- 残存し得る。リソース自体を model.AllResources から除去するため残存行を物理削除する。
-- （uk_permission_group_rules は deleted_at を含まない一意制約であり、論理削除では
-- 行が残り一意キーを占有し続けるため物理削除とする。dead resource の deny 行は
-- 復元価値を持たない。）
-- Do not edit applied migrations. User must run: make migrate

DELETE FROM permission_group_rules
 WHERE resource = 'checkup-package-import';
