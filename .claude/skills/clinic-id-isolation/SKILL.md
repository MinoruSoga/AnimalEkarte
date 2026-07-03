---
name: clinic-id-isolation
description: repository/service で Preload・Where・FindByID・Create/Update・Count を変更した際の clinic_id 越境監査ルータ。実チェックは clinic-isolation-auditor エージェントに委譲する。
---

# Clinic ID Isolation（ルータ）

このskillはチェックロジックを持たない薄いルータである。実際の機械的チェックリストは `clinic-isolation-auditor` エージェントが担当する。ここで判定ロジックを重複実装しない。

## いつ発動するか

`backend/internal/repository/*.go` または `backend/internal/service/*.go` で以下のいずれかを変更した時:
- `Preload` 呼び出し
- `Where` / `FindByID` 呼び出し
- `Create` / `Update` 呼び出し
- `Count` / `Exists` 系クエリ

## 3規則の要約

1. **read: clinic-scoped master の Preload に clinic_id述語**——`Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID)` のように、clinic_idを持つマスタ/区分テーブルへのPreloadには必ずclinic_id述語を付ける。Staff関連（Doctor等）は多医院所属のため単純スコープ禁止の例外
2. **parent-FK: clinic-less子の親をFindByID(clinicID, parentID)で検証**——親テーブル経由でしかclinic_idを持たない子レコードを書き込む前に、親の所有クリニックを検証する
3. **master-FK write: request由来FKをwrite前に検証**——request由来の `XxxID`（`vaccine_id`/`medicine_id`等）をそのままCreate/Updateに渡さず、`FindByID(ctx, clinicID, id)` で当該クリニック所有か検証してから永続化する。ネストしたDTO内の子フィールドのFK漏れ（#124の再発パターン）にも注意

## 実チェックの委譲

詳細な機械チェック・CRITICAL/HIGH/MEDIUM判定・承認基準は本skillでは再実装しない。`clinic-isolation-auditor` エージェントを起動して実施する。

```
Task(subagent_type: clinic-isolation-auditor)
```

## 完了条件

- 変更経路が上記3規則いずれかで隔離されていることを確認した、または
- `clinic-isolation-auditor` エージェントの監査結果（Approve/Warning/Block）が添付されている

## 出典

memory: `cross_tenant_read_idor_audit_20260629` / `cross_tenant_write_audit_20260629` / `cross_tenant_master_fk_write_audit_20260629` / `preload_clinic_scope_lint_p0_20260630`

## 関連エージェント

- `clinic-isolation-auditor`（`.claude/agents/clinic-isolation-auditor.md`）: 本skillが委譲する実チェック本体
