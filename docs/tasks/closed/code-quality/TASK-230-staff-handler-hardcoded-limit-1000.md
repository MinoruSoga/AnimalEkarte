# TASK-230: staff_handler.go — ListStaffs が limit=1000 ハードコードで全件返却

## 優先度
High

## 対象ファイル
- `backend/internal/handler/staff_handler.go`

## 問題概要
`ListStaffs`（行28）が `h.svc.Staff.List(ctx, clinicID, 1, 1000)` と page=1/limit=1000 を固定で呼び出している。
スタッフ数が増加した場合にメモリ・DBクエリのパフォーマンスが低下する。
また、全 List エンドポイントの設計と一貫性がなく、将来的なページネーション対応の妨げになる。

## 現状コード（行28）

```go
// NOTE: pagination パラメータは無視（全件返却）
staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, 1, 1000)
```

## あるべき姿

スタッフは現実的に多くても数十〜数百名であるため、全件返却自体は許容できる。
ただし、ハードコード値を定数化し、将来的なページネーション対応への移行を容易にする。

```go
// handler/staff_handler.go
const staffListMaxLimit = 1000  // 全件返却用の上限定数

func (h *Handler) ListStaffs(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, 1, staffListMaxLimit)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, mapSlice(staffs, toStaffResponse))
}
```

または、他ドメインと同様に `parsePagination(c)` でリクエストから取得する形に統一する。

## 完了条件
- [ ] `1000` を定数 `staffListMaxLimit` に置き換えるか、`parsePagination` で対応
- [ ] `go test ./backend/internal/...` がパス
