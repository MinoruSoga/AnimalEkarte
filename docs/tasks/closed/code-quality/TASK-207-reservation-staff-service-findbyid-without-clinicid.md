# TASK-207: reservation_staff_service.go — Update / PatchStatus 後の再取得で FindByID（clinicID なし）を使用

## 優先度
Medium

## 対象ファイル
`backend/internal/service/reservation_staff_service.go`

## 問題概要
`Update`（行138-140付近）と `PatchStatus`（行173-174付近）が
操作後の結果取得に `s.repo.FindByID(ctx, id)` を使っており、`clinicID` を渡していない。

`FindByIDAndClinicID` が Interface に存在しているにも関わらず、
テナント境界なしの `FindByID` を再取得に使用している。

操作前の存在確認では `FindByIDAndClinicID` を使用しているため、
更新・パッチ自体はテナント安全だが、再取得のテナント境界が外れている。

```go
// 現状（NG）— 行138-140付近
updated, err := s.repo.FindByID(ctx, id)  // clinicID 欠落

// 現状（NG）— 行173-174付近
result, err := s.repo.FindByID(ctx, id)   // clinicID 欠落
```

```go
// あるべき姿
updated, err := s.repo.FindByIDAndClinicID(ctx, clinicID, id)

result, err := s.repo.FindByIDAndClinicID(ctx, clinicID, id)
```

## 完了条件
- [ ] `Update` 後の再取得を `FindByIDAndClinicID` に変更
- [ ] `PatchStatus` 後の再取得を `FindByIDAndClinicID` に変更
- [ ] `go test ./backend/internal/...` がパス
