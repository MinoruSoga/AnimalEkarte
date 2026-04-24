# TASK-248: checkup_handler.go — ListGlobalCheckups が gin.H でラップ + RequirePermission 欠落

## 優先度
High

## 対象ファイル
- `backend/internal/handler/checkup_handler.go`

## 問題概要

### 問題1: gin.H でのレスポンスラップ（他エンドポイントと不統一）
`ListGlobalCheckups`（行192付近）が `gin.H{"data": mapSlice(...)}` という形式でレスポンスを返している。
同ファイルの `ListCheckups`（行31付近）は生配列 `mapSlice(checkups, toCheckupGlobalResponse)` を直接返す。
マスタ系 List エンドポイントは全て生配列を返す規約に反する。

### 問題2: RequirePermission 未設定
対応するルート登録で `ListGlobalCheckups` に `RequirePermission` が設定されていない可能性がある。

## 現状コード（行192付近）

```go
// ❌ gin.H でラップ（他と不統一）
c.JSON(http.StatusOK, gin.H{"data": mapSlice(checkups, toCheckupGlobalResponse)})
```

## 比較（正しい実装）

```go
// ✅ 他マスタと統一
c.JSON(http.StatusOK, mapSlice(checkups, toCheckupGlobalResponse))
```

## あるべき姿

```go
// handler
c.JSON(http.StatusOK, mapSlice(checkups, toCheckupGlobalResponse))

// routes
checkups.GET("", h.RequirePermission(string(model.ResourceCheckups), "view"), h.ListGlobalCheckups)
```

## 完了条件
- [ ] `gin.H{"data": ...}` を `mapSlice(...)` の直接返却に変更
- [ ] ルート登録に `RequirePermission("view")` を追加（未設定の場合）
- [ ] フロントエンドが `response.data` でアクセスしている場合は合わせて修正
- [ ] `go test ./backend/internal/...` がパス
