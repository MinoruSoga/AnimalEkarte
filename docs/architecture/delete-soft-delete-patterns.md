# Hard Delete / Soft Delete 設計パターン

> **目的**: Hard Delete と Soft Delete の使い分け・FK 制約との関係を定義する。
> **読者**: 削除機能の実装者。
> **タイミング**: 削除/論理削除の機能を実装する時。

> **Animal Ekarte**: Hard Delete と Soft Delete の使い分け、FK 制約との関係、実装時チェックリスト
> **最新更新**: 2026-07-10 | **目的**: STG-001 の根本原因解消と今後の実装指針

---

## 1. Hard Delete と Soft Delete の定義

### Hard Delete（物理削除）
- **定義**: データベースから物理的にレコードを削除する操作
- **SQL**: `DELETE FROM table WHERE id = ?`
- **GORM**: `db.Delete(&entity)` （Unscoped 時の既定動作）
- **FK 制約**: 削除時点で FK チェックが発生し、子レコード存在時は DELETE 失敗
- **復旧**: 不可能（バックアップ復旧のみ）

### Soft Delete（論理削除）
- **定義**: `deleted_at` タイムスタンプカラムを SET し、物理的には行を残す
- **SQL**: `UPDATE table SET deleted_at = NOW() WHERE id = ?`
- **GORM**: `db.Delete(&entity)` （`gorm.DeletedAt` フィールド定義時の既定動作）
- **FK 制約**: **物理行が存在するため FK 制約に違反する** ⚠️
- **復旧**: 可能（`deleted_at` を NULL に更新）
- **クエリ**: `gorm.DeletedAt` 定義時、`db.Where()` は自動的に `deleted_at IS NULL` を付加

---

## 2. AnimalEkarte における使い分け方針

| シーン | 削除方式 | 理由 |
|--------|---------|------|
| **監査対象データ（医療記録・会計・スタッフ）** | Soft Delete | コンプライアンス / 取引実績追跡 / 金銭出納記録必須 |
| **テストデータ（STG 環境）** | Hard Delete | 環境リセット / データ干渉排除 / FK制約競合の原因排除 |
| **権限グループ・設定値** | 状況依存 | 現役運用中は Soft Delete；マスタ変更後の古い値は Hard Delete（下記参照） |
| **子 = 親に強従属（FK 必須）** | 親に追従 | 親削除時に自動清掃必須 |
| **子 = 親に弱従属（FK 任意）** | 独立判定 | 子の独立性により柔軟に判定 |

---

## 3. Hard Delete が妥当なケース

### 3.1 テストデータ・一時データ
- **STG 環境の Smoke テスト用データ**
  - API テスト時の一時 clinic / staff / permission_group
  - テスト後、API DELETE で完全削除（FK チェック後）
  - 物理削除により FK 制約競合を排除

### 3.2 Soft Delete 済み子レコードの親削除
- **パターン**: 親 = Hard Delete 対象、子 = Soft Delete 済み
- **問題**: Soft Delete 済み子は物理行が存在し、FK 制約に違反
- **解決**: 親削除前に子を Unscoped Hard Delete
- **例**: STG-001（下記 §5.2 参照）

### 3.3 マスタデータの古い値
- **シーン**: permission_group / clinic_settings の履歴化・廃止
- **判定**: 新値への置き換えが完了し、監査対象外の場合のみ
- **例**: V1 トリガー設定値が CPM V2 完全移行後に廃止される場合

### 3.4 トランザクション内での一括削除
- **シーン**: 一括処理やロールバック時のクリーンアップ
- **条件**: FK チェック完了済み、監査追跡不要
- **実装**: Repository の `DeleteHardInTx()` メソッド

---

## 4. Soft Delete が妥当なケース

### 4.1 監査対象データ
- **医療記録（medical_records / vital_records）**
  - 診療実績、投薬履歴、検査結果
  - コンプライアンス要件で削除履歴が必須
  - UI では削除済みを非表示だが、内部 API では trackable
  - `medical_record_addenda`（追記）は `deleted_at` カラムを持たない Append-only（削除メソッド自体が存在しない）。Soft Delete 対象ではない

- **会計データ（billings / billing_items / accounting）**
  - 領収書・請求書実績の追跡
  - 金銭出納記録法的義務

- **スタッフ・飼い主マスタ（staffs / owners）**
  - 過去の診療実績者との紐付け
  - 復職者の再登録との区別

### 4.2 運用中の設定値
- **Permission Groups**
  - 現役での権限運用中は削除ではなく無効化
  - Soft Delete で historical audit trail を保持
  - 新規グループ追加時に廃止手続き

- **Clinic Settings**
  - マルチテナント環境で clinic 単位の設定変更
  - `clinic_settings` は `clinic_id` を主キーとする単一行構成で `deleted_at` カラムを持たない。UPDATE による上書きのみで、Soft Delete による変更履歴保持は行われていない

---

## 5. FK 制約と Soft Delete の注意点

### 5.1 Soft Delete は物理行を残す
```
問題パターン：
┌─────────────────────────────────────────┐
│ clinics (hard delete 対象)              │
│ id=100, name='Clinic A', deleted_at=NULL│
└─────────────────────────────────────────┘
         ↓ FK constraint
┌─────────────────────────────────────────┐
│ permission_groups (soft delete)          │
│ id=1, clinic_id=100, deleted_at=2026... │ ← 物理行が残る！
│ id=2, clinic_id=100, deleted_at=NULL    │
└─────────────────────────────────────────┘

DELETE FROM clinics WHERE id=100
→ ❌ FK violation（soft-deleted permission_groups が原因）
```

### 5.2 STG-001 事例：PR #64 修正内容

**問題発生**:
- `clinics` hard delete 実行時、soft-deleted `permission_groups` が FK を引き上げていた
- `CountBlockingReferencesByClinicID` はアクティブ行のみカウント（soft-deleted 行をスキップ）
- 結果として DELETE 実行時に予期しない FK エラー発生

**原因**:
- Repository の FK チェックで `gorm.Unscoped()` を使わずにソフト削除済み子をフィルタリング
- Service での事前チェック（409 返す）と Repository での実際の削除（500 FK エラー）のズレ

**修正（commit 1755c193）**:
1. clinic delete 前に Unscoped で soft-deleted permission_groups を全件硬削除
2. 削除操作をトランザクション内でラップ
3. FK エラーは 409 Conflict に mapping（レスポンスポリシー参照）

**修正コード概要**:
```go
// Before: チェック漏れ
activeCount := db.Where("clinic_id = ?", clinicID).
  Count(&count)
// soft-deleted 行が見えないため undercounted

// After: Unscoped でソフト削除済み子を確認して明示的に削除
softDeletedChildren := db.Unscoped().
  Where("clinic_id = ?", clinicID).
  Where("deleted_at IS NOT NULL").
  Delete(&PermissionGroup{})

// 親削除
db.Delete(&clinic)
```

---

## 6. 親 Hard Delete / 子 Soft Delete パターン

### 6.1 実装パターン（推奨）

#### サービス層：事前チェック
```go
// 1. アクティブな子レコードを確認（409 返す）
activeCount := svc.repo.CountActiveChildrenByParentID(parentID)
if activeCount > 0 {
  return apperrors.NewConflict("active_children_exist", 
    "cannot delete parent with active children")
}
```

#### リポジトリ層：Unscoped 削除
```go
// 2. Soft-deleted 子を Hard Delete
func (r *Repo) DeleteParentWithSoftDeletedChildren(ctx context.Context, 
  parentID string) error {
  return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // ① Soft-deleted 子を物理削除（FK 競合排除）
    if err := tx.Unscoped().
      Where("parent_id = ? AND deleted_at IS NOT NULL", parentID).
      Delete(&ChildEntity{}).Error; err != nil {
      return err
    }
    
    // ② 親を削除（アクティブな子は存在しないと Service で確認済み）
    return tx.Delete(&ParentEntity{ID: parentID}).Error
  })
}
```

### 6.2 エラーハンドリング
- **409 Conflict**: Service での事前チェック（アクティブな子が存在）
- **500+ Internal Error**: 予期しない FK エラー（設計バグ）
- **不可の返却**: 400 Bad Request として FK 情報を隠さない（ロギング / 調査の為に 409 を返す）

---

## 7. 親子両者 Soft Delete のケース

### 7.1 適用シーン
- **スタッフ権限グループ**: 親 `permission_groups` と子 `permission_group_rules` / 割当はいずれも監査対象として Soft Delete が妥当（§4.2）
- **医療記録本体**: 親 `medical_records` は Soft Delete（§4.1）
- **対象外 — `medical_record_addenda`**: 追記は Append-only。`deleted_at` カラムを持たず、削除 API / Soft Delete も存在しない（§4.1 と一致）。親子 Soft Delete の例に使わない

### 7.2 実装上の留意点
- **FK チェック**: 親 soft delete 時に、アクティブ子を確認（409）
- **Append-only 子**: `medical_record_addenda` のように子に `deleted_at` が無い場合、親 Soft Delete 時の子 cleanup / 子 Soft Delete は適用しない。親側の active 判定と clinic 軸 FK のみを扱う
- **クエリの明示性**: soft delete 後のジョイン時に `Unscoped()` を明示的に記載
  ```go
  // ❌ 暗黙：WHERE 句に deleted_at 混在判断があいまい
  db.Where("parent_id = ?", id).Find(&children)
  
  // ✅ 明示：soft-deleted も含めて JOIN したい意図を明確化
  db.Unscoped().Where("parent_id = ?", id).Find(&children)
  ```

---

## 8. 子 Hard Delete ケース

### 8.1 適用パターン
- **テストデータの削除**: STG 環境での smoke テスト後の一括削除
- **一時的な関連データ**: 親削除後の orphaned 行の清掃
- **レガシーの廃棄**: V1 トリガー設定など旧バージョンのマスタ削除

### 8.2 実装時の検査
- FK 逆向き確認：child を delete する場合、parent が存在するか確認（緩い確認でよい）
- Parent 側で reference counting をしていないか確認
- Audit log に Hard Delete 痕跡を残す（子が重要データでない場合は不要）

---

## 9. Unscoped() の条件と禁止事項

### 9.1 使用可能なケース
- ✅ **FK チェック目的**: soft-deleted 子の有無を確認してから親削除
- ✅ **Hard Delete 実行**: soft-deleted 子を物理削除（FK 競合排除）
- ✅ **監査ログ記録**: soft-deleted 行の最後のユーザーや削除日時を取得

### 9.2 禁止事項
- ❌ **通常クエリでの使用禁止**: 意図せず soft-deleted 行を返さない
  ```go
  // ❌ 禁止：通常 GET で soft-deleted 行を返却すると UI に重複表示
  db.Unscoped().Where("id = ?", staffID).First(&staff)
  
  // ✅ OK：API は soft-deleted 行を除外
  db.Where("id = ?", staffID).First(&staff) // auto `deleted_at IS NULL`
  ```

- ❌ **マルチテナント隔離の穴に**: clinic_id scoping を失わない
  ```go
  // ❌ 禁止：他の clinic_id の soft-deleted 行も見える
  db.Unscoped().Where("id = ?", id).First(&entity)
  
  // ✅ OK：clinic_id で二重に scoped
  db.Unscoped().
    Where("id = ? AND clinic_id = ?", id, clinicID).
    First(&entity)
  ```

---

## 10. API レスポンスポリシー

### 10.1 ステータスコード期待値

| 操作 | ステータス | 条件 | 例 |
|------|-----------|------|-----|
| DELETE (成功) | **204 No Content** | アクティブ子なし | テスト用 permission_group 削除 |
| DELETE (失敗 FK) | **409 Conflict** | アクティブ子が存在 / FK 制約違反 | clinic 削除時に active staff あり |
| DELETE (soft-deleted 子) | **409 Conflict** | soft-deleted 子で FK 引き上げ | STG-001 パターン |
| 権限不足 | **403 Forbidden** | clinic scope 外 / 権限なし | 他院 clinic の削除試行 |
| バリデーション失敗 | **400 Bad Request** | 必須パラメータ欠損 | `clinic_id` 指定漏れ |
| 予期しない FK エラー | **409 Conflict** または **500 Internal Error** | 設計バグ / データ不整合 | 手動 SQL での FK 違反作成後のデバッグ |

### 10.2 エラーレスポンス
```json
{
  "code": "RESOURCE_CONFLICT",
  "message": "Cannot delete clinic: 3 active staff members exist",
  "details": {
    "blocking_type": "staff",
    "blocking_count": 3,
    "action": "Please delete or reassign staff members first"
  }
}
```

---

## 11. UI 表示ポリシー

### 11.1 Soft-Deleted 行の非表示
- **表示対象**: `deleted_at IS NULL` または GORM 既定 scoping のみ
- **例外**: 管理画面（admin 権限専用）で soft-deleted 履歴を表示（`Unscoped()` 利用）
- **ユーザー影響**: 削除後は一覧から消える；復旧は admin のみ

### 11.2 Hard Delete の痕跡
- **表示**: 検索結果から完全に除去；復旧不可
- **テストデータの場合**: smoke テスト後に DELETE 実行、清掃完了を記録

### 11.3 マルチテナント表示境界
- **clinic A のスタッフ**: clinic A の soft-deleted スタッフも見えない（既定）
- **clinic B のスタッフ**: clinic A データに一切アクセス不可
- **Soft-deleted 行でも clinic_id で完全隔離**

---

## 12. 監査ログ整合性

### 12.1 Hard Delete 時のログ
```json
{
  "action": "DELETE",
  "resource_type": "permission_group",
  "resource_id": "pg-123",
  "deleted_at": "2026-05-26T10:30:00Z",
  "staff_id": "staff-001",
  "deletion_type": "hard",
  "reason": "STG smoke test cleanup",
  "clinic_id": "clinic-100"
}
```

### 12.2 Soft Delete 時のログ
```json
{
  "action": "DELETE",
  "resource_type": "medical_record",
  "resource_id": "mr-456",
  "soft_deleted_at": "2026-05-26T10:31:00Z",
  "staff_id": "staff-002",
  "deletion_type": "soft",
  "recovery_possible": true,
  "clinic_id": "clinic-100"
}
```

### 12.3 Unscoped Hard Delete の追記
- Soft-deleted 行の hard delete（FK 競合排除）は audit_logs に個別記録
- 親削除と同一トランザクション内で実行されても個別ログ行として区別
- 理由：FK cleanup の explicit 証跡

### 12.4 臨床結果レコード物理削除と監査ログの原子性（#211）
- **背景**: `exam_results`, `checkup_field_results` などの臨床結果レコードは物理削除（Hard Delete）されますが、削除後の監査ログ（`audit_logs`）の書き込みが別トランザクションで行われると、監査ログ書き込みが失敗した際に物理削除をロールバックできず、監査トレイルが失われる危険性があります。
- **Fail-closed 設計**: 臨床結果レコードの置換（物理削除＋再作成）時の物理削除と監査ログ書き込みは、必ず同一のトランザクション（ambient tx / `dbOrTx`）内で実行しなければなりません。監査ログの作成に失敗した場合はトランザクション全体をロールバックし、削除を元に戻します。
- **CI ゲートテスト (`audit_tx_inventory_lint_test.go`)**: 臨床結果レコードを物理削除するリポジトリ関数を監査対象の allowlist（許容リスト）で管理し、新規の物理削除が発生した場合は CI で検知・制限します。
- **返金処理の同一トランザクション化**: `refund_service.go`（返金処理）においても、返金処理と監査ログ作成を同一のトランザクション内で実行し、監査ログ作成に失敗した場合は返金処理もロールバックされるように実装します。
- **監査ログの定義拡張**: #211 の置換処理に伴い、監査ログに記録するリソース定数 `checkup_field_result` およびアクション定数 `checkup_field_result.replace` を新たに定義し、監査証跡の追跡性を保証しています。
- **並行競合時の無監査物理削除の防止 (READ COMMITTED 対策)**: 置換（Replace）処理において、削除判定を「スナップショット時の事前取得件数」ではなく、DELETE文実行後の実際の `RowsAffected` (物理削除数) を用いて監査ログ作成の要否をゲートします。これにより、READ COMMITTED 分離レベル下で事前取得時点では 0 件であっても、直後に並行処理でコミットされた新規データを削除してしまい、監査ログが出力されない（無監査物理削除）という競合を防ぎます。
- **移行ステータス**:
  - `checkup_field_results` (完了 / `statusAuditedTxInternal`): ambient tx を用いてトランザクション内で削除と監査を実施。実行時テスト `checkup_field_result_tx_atomicity_test.go` により担保。
  - `exam_results` (完了 / `statusAuditedTxInternal`): `examinationRepository.ReplaceItemsByExamID` が `dbOrTx` ambient tx を用いてトランザクション内で削除と監査を実施するよう移行済み。実行時テスト `examination_repository_tx_atomicity_test.go` により担保。

---

## 13. マルチテナント Clinic_ID 隔離

### 13.1 Soft Delete 下での隔離
- Soft Delete でも `clinic_id` 制約は生存
- Unscoped query 時も必ず `WHERE clinic_id = ?` で二重 scoping
  ```go
  // ✅ 正：clinic_id スコープ必須
  db.Unscoped().
    Where("clinic_id = ? AND id = ?", clinicID, id).
    Delete(&entity)
  ```

### 13.2 Hard Delete 下での隔離
- clinic A が clinic B のデータを hard delete できないか確認
- persistence pathでの `clinic_id` WHERE句は削除時も必須
- テスト項目：clinic A の staff が clinic B の staff を削除できないことを確認

### 13.3 FK で clinic_id の一貫性を検証
- FK 制約は clinic_id の文字列値まで厳密にチェック（型安全）
- STG テスト時：異なる clinic でレコード作成後、FK violation を意図的に引き起こして検証

---

## 14. 実装チェックリスト

### 14.1 新しい削除機能を追加する際

- [ ] **仕様判定**
  - [ ] 対象エンティティが監査対象か確認（soft delete / hard delete）
  - [ ] FK 関係を mapper（外部キー制約）で確認
  - [ ] 親子関係の削除順序を定義

- [ ] **Application invariant**
  - [ ] 削除前チェック：アクティブな子レコード数をカウント（FK チェック）
  - [ ] 409 Conflict を返す条件を明示
  - [ ] トランザクション制御：begin / commit / rollback の実装

- [ ] **Persistence**
  - [ ] FK チェック済み前提で削除実行
  - [ ] Hard delete の場合：clinic_id スコープを必ず含める
  - [ ] Soft delete の場合：audit log タイムスタンプを記録
  - [ ] Unscoped() の使用箇所を限定（FK cleanup のみ）
  - [ ] コメント：Unscoped 使用理由を記載（保守性のため）

- [ ] **HTTP boundary**
  - [ ] DELETE ハンドラの実装
  - [ ] 409 vs 400 vs 403 のステータスコード判定
  - [ ] エラーレスポンスに blocking_type / blocking_count を含める

- [ ] **テスト**
  - [ ] FK アクティブ子の 409 テスト
  - [ ] clinic_id scoping の 403 テスト
  - [ ] 成功時 204 テスト
  - [ ] soft-deleted 子を含むテスト（ある場合）

### 14.2 既存削除機能を修正する際

- [ ] **FK チェックの検証**
  - [ ] `CountBlockingReferences*` メソッドが Unscoped を使っているか確認
  - [ ] soft-deleted 子を正しくカウントしているか確認

- [ ] **エラーマッピング**
  - [ ] 予期しない FK エラーが 500 で返っていないか
  - [ ] 409 Conflict で適切にマップされているか

- [ ] **トランザクション**
  - [ ] 親削除と soft-deleted 子 hard delete が同じトランザクション内か
  - [ ] ロールバック時の整合性を確認

- [ ] **マルチテナント隔離**
  - [ ] clinic_id WHERE 句が削除時も必須か確認
  - [ ] テスト：他院データ削除が 403 で拒否されるか確認

---

## 15. テストチェックリスト

### 15.1 削除機能の単体テスト

```go
// ✅ 成功ケース
func TestDeletePermissionGroup_Success(t *testing.T) {
  // Arrange: アクティブな子（permission_group_assignments）なし
  group := createTestGroup(t, db)
  
  // Act
  err := svc.DeletePermissionGroup(ctx, group.ID, clinicID)
  
  // Assert
  assert.NoError(t, err)
  
  // Get で 404 確認
  _, err := repo.GetPermissionGroup(ctx, group.ID)
  assert.NotNil(t, err)
}

// ✅ FK 保護テスト
func TestDeletePermissionGroup_WithActiveAssignments_409(t *testing.T) {
  group := createTestGroup(t, db)
  assignment := createAssignment(t, db, group.ID)
  
  err := svc.DeletePermissionGroup(ctx, group.ID, clinicID)
  
  assert.Error(t, err)
  assert.Equal(t, apperrors.ConflictError, err.Type())
}

// ✅ マルチテナント隔離テスト
func TestDeletePermissionGroup_OtherClinic_403(t *testing.T) {
  group := createTestGroup(t, db, clinicID="clinic-A")
  
  err := svc.DeletePermissionGroup(ctx, group.ID, clinicID="clinic-B")
  
  assert.Error(t, err)
  assert.Equal(t, apperrors.ForbiddenError, err.Type())
}
```

### 15.2 統合テスト（父子削除フロー）

```go
// ✅ Soft-deleted 子を含む親削除
func TestDeleteClinic_WithSoftDeletedPermissionGroups(t *testing.T) {
  clinic := createTestClinic(t, db)
  group := createTestGroup(t, db, clinic.ID)
  
  // 子を soft delete
  db.Delete(&group)
  
  // 親削除（soft-deleted 子を hard delete する内部処理）
  err := svc.DeleteClinic(ctx, clinic.ID)
  
  assert.NoError(t, err)
  
  // 親と子両方削除されたことを確認
  _, err := repo.GetClinic(ctx, clinic.ID)
  assert.Error(t, err)
  
  _, err := repo.GetPermissionGroup(ctx, group.ID)
  assert.Error(t, err)
  
  // Unscoped で物理削除も確認
  var count int64
  db.Unscoped().Where("id = ?", group.ID).Count(&count)
  assert.Equal(t, int64(0), count)
}
```

### 15.3 E2E テスト（API 観点）

```bash
# ✅ 削除成功（204）
curl -X DELETE "https://api.stg.noah-karte.com/api/v1/masters/permission-groups/pg-123" \
  -b "access_token=${TOKEN}"
# 期待: 204 No Content

# ✅ FK 違反（409）
curl -X DELETE "https://api.stg.noah-karte.com/api/v1/clinics/clinic-100" \
  -b "access_token=${TOKEN}"
# 期待: 409 Conflict
# {
#   "code": "RESOURCE_CONFLICT",
#   "message": "Cannot delete clinic: 3 active staff members exist"
# }

# ✅ マルチテナント隔離（403）
curl -X DELETE "https://api.stg.noah-karte.com/api/v1/clinics/clinic-B-id" \
  -b "access_token=${TOKEN_CLINIC_A}"
# 期待: 403 Forbidden
```

---

## 参考資料

- **デプロイメント運用**: [`docs/ops/deploy/README.md`](../ops/deploy/README.md)
- **STG テストデータライフサイクル**: [`docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md`](../ops/deploy/STG-DEMO-DATA-LIFECYCLE.md)（テスト削除ポリシー参照）
- **PR #64 修正内容**: commit `1755c193` 他
- **Go アーキテクチャ規約**: [`backend/CLAUDE.md`](../../backend/CLAUDE.md)
- **API 設計**: [`backend/docs/api.yaml`](../../backend/docs/api.yaml)（contract 正本）
