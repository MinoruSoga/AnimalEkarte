# BE-114: owner / pet 検索の ひらがな⇔カタカナ正規化

**Status**: Closed (2026-04-14)
**Priority**: Medium
**Affects**: `owner_repository.FindAll`, `pet_repository.FindAll`
**Date Created**: 2026-04-14
**Related**: BUG-375, BE-113, FE-251

## Summary

owner / pet の `FindAll` 検索で `translate()` を用いて検索クエリと DB 値の双方を正規化し、
**ひらがな入力でもカタカナ入力でも同じヒット結果**を返すようにする。
さらに owner 検索対象に `name_kana` を追加し、pet 検索対象に `owners.name_kana` を追加する。

## 現状のコード

### Owner FindAll（`name_kana` 検索対象外）

```go
// backend/internal/repository/owner_repository.go:35-49
func (r *ownerRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
    owners := make([]model.Owner, 0)
    var total int64

    buildBase := func() *gorm.DB {
        q := r.db.WithContext(ctx).Model(&model.Owner{}).Scopes(clinicScope(clinicID))
        if search != "" {
            pattern := "%" + escapeLike(search) + "%"
            q = q.Where(
                `(name ILIKE ? ESCAPE '\' OR phone ILIKE ? ESCAPE '\' OR email ILIKE ? ESCAPE '\')`,
                pattern, pattern, pattern,
            )
        }
        return q
    }
    // ...
}
```

### Pet FindAll（`name_kana` は検索対象だが正規化なし）

```go
// backend/internal/repository/pet_repository.go:31-50
func (r *petRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
    // ...
    if search != "" {
        escaped := escapeLike(search)
        pattern := "%" + escaped + "%"
        q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.deleted_at IS NULL").
            Where(
                `(pets.name ILIKE ? ESCAPE '\' OR pets.name_kana ILIKE ? ESCAPE '\' OR owners.name ILIKE ? ESCAPE '\')`,
                pattern, pattern, pattern,
            )
    }
    // ...
}
```

## 必要な変更

### 1. Owner FindAll — 検索条件に `name_kana` 追加 + translate 正規化

```go
// backend/internal/repository/owner_repository.go
func (r *ownerRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
    owners := make([]model.Owner, 0)
    var total int64

    buildBase := func() *gorm.DB {
        q := r.db.WithContext(ctx).Model(&model.Owner{}).Scopes(clinicScope(clinicID))
        if search != "" {
            pattern := "%" + escapeLike(search) + "%"
            // BUG-375: name_kana はひらがな⇔カタカナ正規化して比較
            q = q.Where(
                `(name ILIKE ? ESCAPE '\'`+
                    ` OR translate(name_kana, ?, ?) ILIKE translate(? ESCAPE '\', ?, ?)`+
                    ` OR phone ILIKE ? ESCAPE '\'`+
                    ` OR email ILIKE ? ESCAPE '\')`,
                pattern,
                kanaSourceChars, kanaTargetChars, pattern, kanaSourceChars, kanaTargetChars,
                pattern, pattern,
            )
        }
        return q
    }
    // ... rest unchanged
}
```

### 2. Pet FindAll — translate 正規化 + `owners.name_kana` 追加

```go
// backend/internal/repository/pet_repository.go
func (r *petRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
    // ...
    if search != "" {
        escaped := escapeLike(search)
        pattern := "%" + escaped + "%"
        q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.deleted_at IS NULL").
            Where(
                `(pets.name ILIKE ? ESCAPE '\'`+
                    ` OR translate(pets.name_kana, ?, ?) ILIKE translate(? ESCAPE '\', ?, ?)`+
                    ` OR owners.name ILIKE ? ESCAPE '\'`+
                    ` OR translate(owners.name_kana, ?, ?) ILIKE translate(? ESCAPE '\', ?, ?))`,
                pattern,
                kanaSourceChars, kanaTargetChars, pattern, kanaSourceChars, kanaTargetChars,
                pattern,
                kanaSourceChars, kanaTargetChars, pattern, kanaSourceChars, kanaTargetChars,
            )
    }
    // ... rest unchanged
}
```

### 3. 共通定数（`backend/internal/repository/kana_normalize.go` 新規）

```go
package repository

// BUG-375: ひらがな⇔カタカナ正規化用の文字マッピング。
// translate(s, kanaSourceChars, kanaTargetChars) で カタカナ → ひらがな に正規化する。
// 86 文字対応: U+30A1 (ァ) 〜 U+30F6 (ヶ) → U+3041 (ぁ) 〜 U+3096 (ゖ)
const (
    kanaSourceChars = "ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"
    kanaTargetChars = "ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"
)
```

### 4. テスト追加

```go
// backend/internal/repository/owner_repository_test.go (新規 or 既存に追記)
func TestOwnerRepository_FindAll_KanaSearch(t *testing.T) {
    // ひらがな登録 → カタカナ検索でヒット
    // カタカナ登録 → ひらがな検索でヒット（マイグレーション後はひらがな前提）
    // 文字種混在も OK
}
```

## API レスポンス形式

変更なし。

## フロントエンド影響

- 検索ボックスで「ハヤシ」「はやし」どちらを入力しても同じ結果
- FE-251 で placeholder のみ更新

## 完了条件

- [ ] 共通定数 `kanaSourceChars` / `kanaTargetChars` 追加
- [ ] owner_repository.FindAll: name_kana を検索対象に追加 + translate 正規化適用
- [ ] pet_repository.FindAll: translate 正規化適用 + owners.name_kana を検索対象に追加
- [ ] テスト追加: ひらがな↔カタカナ相互ヒット検証
- [ ] 既存テスト全件パス
- [ ] `go vet ./...` パス
- [ ] golangci-lint エラーなし

## リスク

| リスク | 影響 | 対策 |
|--------|------|------|
| translate がインデックス無効化 → Seq Scan | 中 | owner/pet マスタは数千件規模で許容。性能劣化時に functional index `(translate(name_kana, kanaSource, kanaTarget))` を別 issue で追加 |
| translate の文字列リテラル長すぎ → 可読性低下 | 低 | 共通定数化で対応済み |
| escape sequence と translate の併用で SQL 構文エラー | 中 | placeholder 順序を厳密に管理。テストで網羅 |

## 参照

- BUG-375: 全体タスク
- BE-113: マイグレーション（前提）
- PostgreSQL `translate(string, from, to)` 公式: https://www.postgresql.org/docs/current/functions-string.html
