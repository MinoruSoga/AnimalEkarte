# TASK-039: BE 未登録エンドポイント対応の判断と実装

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 中
**領域**: Backend

---

## 概要

service 層に実装・テストが存在するにもかかわらず、対応する HTTP エンドポイントが router に未登録のメソッドが3つある。
「エンドポイントとして公開する」または「service/repository から削除する」のいずれかを判断し実装する。

---

## 対象メソッド

### 1. `MedicalRecordService.GetByRecordNo`

```go
// service/medical_record_service.go L43
GetByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error)
```

- 実装あり（L75）
- テストあり（`medical_record_service_test.go:291`）
- handler のモックにも定義あり
- **対応する HTTP エンドポイントが未登録**（例: `GET /v1/medical-records?record_no=xxx`）

**判断**: カルテを診療券番号で検索するユースケースとして有用か？

---

### 2. `PermissionGroupService.GetByID`

```go
// service/permission_group_service.go
GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
```

- 実装あり
- テストあり
- **`GET /v1/permission-groups/:id` が `RegisterPermissionGroupRoutes` に未登録**

**判断**: 権限グループ詳細取得 API として公開するか？（現 UI では一覧取得のみ使用）

---

### 3. `BillingItemService.GetByID`

```go
// service/billing_item_service.go
GetByID(ctx context.Context, id uint64) (*model.BillingItem, error)
```

- 実装あり
- テストあり
- **単体取得エンドポイント未実装**

**判断**: 会計明細の単体取得 API が必要か？（現状は一覧・作成・更新・削除のみ）

---

## 対応方針（実装者が選択）

各メソッドについて以下のいずれかを選択する：

**A. エンドポイントを追加する**
1. handler に GET /:id ハンドラを実装
2. RegisterXxxRoutes に登録
3. フロントエンドから利用する場合は API hook も追加

**B. service / repository から削除する**
1. service interface からメソッドを削除
2. service 実装を削除
3. repository メソッドも削除（他から使われていない場合）
4. テストファイルのモック定義も削除

---

## 推奨判断

| メソッド | 推奨 | 理由 |
|---------|------|------|
| `GetByRecordNo` | B（削除）| フロントエンドは ID でカルテを取得している。診療券番号検索は現状不要 |
| `GetByID`（PermissionGroup）| A（追加）| 権限グループ編集時に単体取得が必要になる可能性が高い |
| `GetByID`（BillingItem）| B（削除）| 会計明細は一覧から操作するため単体取得 API は現時点では不要 |

---

## 受入条件

- [ ] 3メソッドについて A/B を決定し、対応が完了している
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./...` 全テストパス
