---
name: golang-patterns
description: Idiomatic Go パターン。Gin/GORM/apperrors を使ったこのプロジェクト固有のベストプラクティス。Go コード作成・レビュー時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# Go 開発パターン

このプロジェクト（Go 1.25 / Gin / GORM / PostgreSQL 18）で使用するイディオマティックな Go パターン。

## When to Activate

- 新規 Go コードの作成
- Go コードのレビュー・リファクタリング
- アーキテクチャ設計

## Handler → Service → Repository パターン

### Handler（HTTP 層）
```go
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    owner, err := h.service.GetOwner(c.Request.Context(), uint(id))
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, owner)
}
```

### Service（ビジネスロジック層）
```go
func (s *OwnerService) GetOwner(ctx context.Context, id uint) (*model.Owner, error) {
    slog.InfoContext(ctx, "getting owner", "id", id)

    owner, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get owner")
    }
    return owner, nil
}
```

### Repository（データアクセス層）
```go
func (r *OwnerRepository) GetByID(ctx context.Context, id uint) (*model.Owner, error) {
    var owner model.Owner
    if err := r.db.WithContext(ctx).
        Where("clinic_id = ? AND id = ?", r.clinicID, id).
        First(&owner).Error; err != nil {
        return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
    }
    return &owner, nil
}
```

## エラーハンドリングパターン

### apperrors の使い分け
```go
// Repository: GORM エラー変換
return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))

// Service: コンテキスト付きラップ
return nil, apperrors.Wrap(err, "failed to create owner")

// Service: 競合エラー（マスタ削除時）
return apperrors.WrapConflict("この項目は使用中のため削除できません")

// Handler: 統一レスポンス
RespondError(c, err)
```

### マスタ削除の FK チェック
```go
func (s *MedicineMasterService) Delete(ctx context.Context, id uint) error {
    count, err := s.repo.CountUsageByMedicineID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to count usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この薬剤は使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

## PATCH パターン（ポインタ型 + UpdateFields）

```go
// Request DTO（ポインタ型でゼロ値問題を回避）
type UpdateOwnerInput struct {
    Name  *string `json:"name"`
    Email *string `json:"email"`
    Phone *string `json:"phone"`
}

// Service: フィールドマップを構築
func buildOwnerUpdateFields(input UpdateOwnerInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil { fields["name"] = *input.Name }
    if input.Email != nil { fields["email"] = *input.Email }
    if input.Phone != nil { fields["phone"] = *input.Phone }
    return fields
}

// Repository: Updates で部分更新
func (r *OwnerRepository) UpdateFields(ctx context.Context, id uint, fields map[string]any) (*model.Owner, error) {
    var owner model.Owner
    return &owner, r.db.WithContext(ctx).
        Model(&model.Owner{}).
        Where("clinic_id = ? AND id = ?", r.clinicID, id).
        Updates(fields).
        First(&owner).Error
}
```

## 並行処理（errgroup）

```go
import "golang.org/x/sync/errgroup"

func (s *OwnerService) CreateWithPets(ctx context.Context, input CreateOwnerInput) error {
    owner, err := s.repo.CreateOwner(ctx, &model.Owner{Name: input.Name})
    if err != nil {
        return apperrors.Wrap(err, "failed to create owner")
    }

    g, ctx := errgroup.WithContext(ctx)
    for _, pet := range input.Pets {
        g.Go(func() error {
            return s.petRepo.Create(ctx, &model.Pet{OwnerID: owner.ID, Name: pet.Name})
        })
    }
    if err := g.Wait(); err != nil {
        return apperrors.Wrap(err, "failed to create pets")
    }
    return nil
}
```

## Interface 設計（最小化）

```go
// ✅ 必要なメソッドのみ定義
type OwnerRepository interface {
    GetByID(ctx context.Context, id uint) (*model.Owner, error)
    Create(ctx context.Context, owner *model.Owner) error
    UpdateFields(ctx context.Context, id uint, fields map[string]any) (*model.Owner, error)
    Delete(ctx context.Context, id uint) error
}

// ❌ 巨大インターフェース禁止
type BigRepository interface {
    // 20個以上のメソッド...
}
```

## slog（構造化ログ）

```go
// ✅ Service 層のみ記述
slog.InfoContext(ctx, "creating owner", "name", input.Name, "clinic_id", input.ClinicID)
slog.ErrorContext(ctx, "failed to create owner", "error", err, "name", input.Name)

// ❌ Handler/Repository には書かない
// ❌ パスワード・トークンをログに含めない
```

## Core Principles（ECC golang-patterns より）

1. **Accept interfaces, return structs** — 入力は interface、出力は concrete type
2. **Make zero value useful** — `bytes.Buffer`, `sync.Mutex` など初期化不要な設計
3. **Context first** — 全メソッドの第一引数は `ctx context.Context`
4. **Wrap errors** — `fmt.Errorf("context: %w", err)` でコンテキスト情報を保持
5. **Early return** — `if err != nil { return } ` パターンで深いネストを避ける

## 監査ログの原子化（実績パターン）

状態変更（返金・削除等）と監査ログ書込は**同一トランザクションで原子化**し、監査書込失敗時は本処理も巻き戻す（fail-closed）。`AuditTxLogger` / `LogEntryTx` が dbOrTx を受け取る実装が先例。
- 実例: refund_service の Create（commit 6f432912）、clinical系（commit fe04b460）
- 注意: 現状の監査書込は大半が tx外 best-effort（auditRepository.Create が dbOrTx 非対応）。横断是正は docs/tasks/open/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md 参照
（出典: memory issue211_refund_audit_tx_atomicity_20260701 / issue211_audit_tx_atomicity_verify_20260630）

## 対称定数の単一ソース化（実績パターン）

sync/trigger のように2箇所で一致が必要な定数は単一ソースに集約し、alignment test で固定する。片側だけの変更は静かに機能を無効化する（配信が発火しない等）。
（出典: memory feedback_constants_dual_source）
