# CODE-QUALITY-214: マスタ系 その他構造的問題まとめ

## 概要

前回レビュー（CODE-QUALITY-207）から追加で発見した構造的問題をまとめる。
各単独では軽微だが、放置するとコードベースの一貫性・保守性が損なわれる。

## 優先度

MEDIUM

---

## 問題一覧

### 1. diagnosis_handler.go — ReorderDiagnosisNames エンドポイントが欠落

**ファイル**: `backend/internal/handler/diagnosis_handler.go`

`DiagnosisType` には `ReorderDiagnosisTypes`（L131）があるが、
`DiagnosisName` には対応する `ReorderDiagnosisNames` が存在しない。
`DiagnosisNameService` インターフェースに `Reorder` があるなら、Handler と Route 登録も必要。

**修正方針**: `DiagnosisNameService` インターフェースの `Reorder` メソッドの存在を確認し、
あればハンドラと `PATCH /diagnosis-names/reorder` ルートを追加する。
不要なら Service インターフェースから `Reorder` を削除する。

---

### 2. permission_group_service.go:101 — Update で validateOptionalName 未呼び出し

**ファイル**: `backend/internal/service/permission_group_service.go`

```go
// 現状（validateOptionalName なし）
func (s *permissionGroupService) Update(...) (*model.PermissionGroup, error) {
    fields := buildPermissionGroupUpdateFields(input)
    if len(fields) == 0 { ... }
    // ← validateOptionalName(input.Name) がない
```

他のマスタ（occupation, chief_complaint 等）は `validateOptionalName` を先に呼び出している。
空文字の権限グループ名を Update で設定できてしまう。

**修正方針**:
```go
func (s *permissionGroupService) Update(...) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildPermissionGroupUpdateFields(input)
    ...
```

---

### 3. payment_method_master_handler.go — Location ヘッダーのパスが /v1/masters/ 以外

**ファイル**: `backend/internal/handler/payment_method_master_handler.go`

```go
// 現状
c.Header("Location", fmt.Sprintf("/v1/payment-methods/%d", m.ID))
```

他の全マスタは `/v1/masters/{resource}/{id}` の形式で Location ヘッダーを設定している。
`payment_method_master` は `/v1/payment-methods` に登録されているため機能的には正しいが、
URL 規約が他と異なる。

**修正方針**: ルーティング設計を確認し、意図的な分離なら `response.go` に `paymentMethodMasterRoutePrefix` 定数を定義してコメントで明記する。統一すべきなら `PATCH /v1/masters/payment-methods` に移動する。

---

### 4. medicine_service.go:247 — ErrMsgIDsNotEmpty を定数未使用でリテラル直書き

**ファイル**: `backend/internal/service/medicine_service.go`

```go
// 現状（リテラル）
return apperrors.WrapInvalidInput("並び順のIDリストが空です")

// 修正後（定数使用）
return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
```

他の全サービスは `ErrMsgIDsNotEmpty` 定数を使用している。

---

### 5. reservation_type_occupation_repository.go:53/69 — Preload の deleted_at IS NULL 欠落

**ファイル**: `backend/internal/repository/reservation_type_occupation_repository.go`

```go
// 現状（deleted_at IS NULL なし）
Preload("Occupation", "clinic_id = ?", clinicID)

// 修正後
Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID)
```

論理削除済みの Occupation が Preload されてしまう可能性がある。

---

### 6. diagnosis_service.go — ListNames デッドコードの明示

**ファイル**: `backend/internal/service/diagnosis_service.go`

`DiagnosisNameService.ListNames` がインターフェースに定義されているが Handler から未呼び出し。
CODE-QUALITY-212 で扱うが、判断待ちの間は以下のコメントを追加して意図を明示する。

```go
// TODO: ListDiagnosisNamesAll エンドポイント実装時に接続すること。
// 現在ハンドラから未呼び出し。不要なら削除する。
ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
```

---

### 7. trimming_course_response.go:12 — TargetSize の omitempty と型の不整合

**ファイル**: `backend/internal/handler/trimming_course_response.go`

```go
// 現状
TargetSize string `json:"target_size,omitempty"`
```

`string` + `omitempty` のため、空文字 `""` でもフィールドが JSON から消える。
フロントエンドが `undefined` と `""` を区別できない。
`toTrimmingCourseResponse` では nil 時に空文字を設定しているため、
nil の場合も空文字の場合もどちらも JSON 出力が省略される。

**修正方針**: `TargetSize *string json:"target_size"` に変更（ポインタで nil/空文字を区別）するか、
`omitempty` を外して常に出力する。

---

## 規約参照

- `.claude/CLAUDE.md`: エラーメッセージ / 定数使用 / 責任分離

## テスト

- permission_group の Update で空文字名を設定しようとした場合に 400 が返ることを確認
- trimming_course の TargetSize が null / 空文字で正しく返ることを確認
