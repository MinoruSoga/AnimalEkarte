# BE-044: 診断病名マスタ レスポンス形式修正（paginated → plain array）

**Status**: Open
**Priority**: High
**Affects**: 診断病名マスタ設定ページ (`/settings/diagnosis`)
**Date Created**: 2026-03-18
**Related**: TASK-016

## Summary

`ListDiagnosisCategories` と `ListDiagnosisNames` がページネーション付きレスポンス `{data: [...], total, page, limit}` を返しているが、フロントエンドはプレーン配列 `[...]` を期待しているためデータが表示されない。

## 現状のコード

### Handler（ページネーション付き — 不正）

```go
// backend/internal/handler/diagnosis_handler.go:48-53
// ListDiagnosisCategories
categories, total, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID, page, limit)
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusOK, newPaginatedResponse(toDiagnosisCategoryResponseList(categories), total, page, limit))
// ❌ フロントエンドはプレーン配列を期待

// :190-205
// ListDiagnosisNames（2箇所）
names, total, svcErr := h.svc.DiagnosisName.ListByCategoryID(...)
resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)
// ...
names, total, svcErr := h.svc.DiagnosisName.List(...)
resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)
```

### フロントエンド（プレーン配列を期待 — 正しい）

```typescript
// frontend/src/features/master/api/diagnosis.ts:76-81
export async function listDiagnosisCategories(): Promise<DiagnosisCategory[]> {
  const { data } = await axios.get<ModelDiagnosisCategory[]>(
    "/v1/masters/diagnosis-categories",
  );
  return data.map(transformDiagnosisCategory);
  // ❌ data がオブジェクト {data, total, ...} の場合 .map() が失敗
}
```

### DB確認済み

```
diagnosis_categories: 8件存在（シードデータ）
diagnosis_names: 20件存在（シードデータ）
```

## 必要な変更

### 1. ListDiagnosisCategories（:48-53）

```go
// Before
categories, total, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID, page, limit)
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusOK, newPaginatedResponse(toDiagnosisCategoryResponseList(categories), total, page, limit))

// After
categories, _, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID, page, limit)
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusOK, toDiagnosisCategoryResponseList(categories))
```

### 2. ListDiagnosisNames（:190-205、2箇所）

```go
// Before（1箇所目: ListByCategoryID）
names, total, svcErr := h.svc.DiagnosisName.ListByCategoryID(...)
resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)

// After
names, _, svcErr := h.svc.DiagnosisName.ListByCategoryID(...)
resp = toDiagnosisNameResponseList(names)

// Before（2箇所目: List）
names, total, svcErr := h.svc.DiagnosisName.List(...)
resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)

// After
names, _, svcErr := h.svc.DiagnosisName.List(...)
resp = toDiagnosisNameResponseList(names)
```

### 3. resp 変数の型変更

`resp` 変数の型を `any` に変更するか、直接 `c.JSON()` を呼ぶように分岐を修正する（`newPaginatedResponse` の戻り値型と `[]diagnosisNameResponse` は型が異なるため）。

## フロントエンド影響

- 変更なし。フロントエンドの `listDiagnosisCategories()` と `listDiagnosisNames()` は正しくプレーン配列を期待している。

## 完了条件

- [ ] `ListDiagnosisCategories` がプレーン配列を返している
- [ ] `ListDiagnosisNames` がプレーン配列を返している（ListByCategoryID と List の2箇所）
- [ ] `docker compose exec backend go build ./cmd/api` パス
- [ ] `/settings/diagnosis` ページで診断病名カテゴリ一覧（8件）が表示される
- [ ] `/settings/diagnosis` ページで診断病名一覧（20件）が表示される
- [ ] カテゴリでフィルタリングした場合も正しく動作する
