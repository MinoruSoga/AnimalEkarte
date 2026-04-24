# BUG-381: trimming_master_handler の Create エンドポイントで Location ヘッダーが欠如

## 概要
`trimming_master_handler.go` の `CreateTrimmingCourse` と `CreateTrimmingOption` は、201 Created レスポンスを返す際に `Location` ヘッダーを設定していない。他の全マスタ Create ハンドラが `Location` ヘッダーを返しているため、REST 規約の一貫性が破られている。

## 再現手順
1. `POST /v1/masters/trimming-courses` を実行
2. **結果**: `201 Created` だが `Location` ヘッダーなし
3. 比較: `POST /v1/masters/exam-types` → `201 Created` + `Location: /v1/masters/exam-types/123`

## 期待する動作
- 全 Create エンドポイントで `Location: /v1/masters/{resource}/{id}` ヘッダーを返すこと

## 現状コード

### `backend/internal/handler/trimming_master_handler.go:75`（Location ヘッダーなし）
```go
c.JSON(http.StatusCreated, toTrimmingCourseResponse(course))
```

### `backend/internal/handler/trimming_master_handler.go:209`（Location ヘッダーなし）
```go
c.JSON(http.StatusCreated, toTrimmingOptionResponse(option))
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/handler/exam_type_handler.go:75-76
c.Header("Location", fmt.Sprintf("/v1/masters/exam-types/%d", examType.ID))
c.JSON(http.StatusCreated, toExamTypeResponse(examType))

// backend/internal/handler/vaccine_handler.go:83-84
c.Header("Location", fmt.Sprintf("/v1/masters/vaccines/%d", vaccine.ID))
c.JSON(http.StatusCreated, toVaccineResponse(vaccine))
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/trimming_master_handler.go:75` | CreateTrimmingCourse の Location ヘッダー欠如 | 要修正 |
| `backend/internal/handler/trimming_master_handler.go:209` | CreateTrimmingOption の Location ヘッダー欠如 | 要修正 |

## 修正方針

### 1. `backend/internal/handler/trimming_master_handler.go:75`
```go
// 修正前
c.JSON(http.StatusCreated, toTrimmingCourseResponse(course))

// 修正後
c.Header("Location", fmt.Sprintf("/v1/masters/trimming-courses/%d", course.ID))
c.JSON(http.StatusCreated, toTrimmingCourseResponse(course))
```

### 2. `backend/internal/handler/trimming_master_handler.go:209`
```go
// 修正前
c.JSON(http.StatusCreated, toTrimmingOptionResponse(option))

// 修正後
c.Header("Location", fmt.Sprintf("/v1/masters/trimming-options/%d", option.ID))
c.JSON(http.StatusCreated, toTrimmingOptionResponse(option))
```

※ API パスの正確なリソース名は `cmd/api/main.go` のルーティング定義を確認して合わせること。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md` — Endpoint Design
> Use RESTful conventions. Return consistent error response format. Use proper HTTP status codes.

### RFC 7231 Section 7.1.2 — Location ヘッダー
> 201 Created レスポンスには作成されたリソースの URI を Location ヘッダーで返すことが推奨される。

### プロジェクト内参照実装
`backend/internal/handler/exam_type_handler.go:75-76` — Location ヘッダーの正しい実装

## 優先度
**Low** — REST 規約の一貫性問題。既存機能に影響なし。フロントエンドが Location ヘッダーを使用していれば影響があるが、現状では影響軽微。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/trimming_master_handler.go:75,209` — 問題箇所
- `backend/cmd/api/main.go` — ルーティング定義（パス名確認用）
