# TASK-210: reservation_type_liff — Delete の FK チェックが service + repository で二重検知

## 優先度
Medium

## 対象ファイル
- `backend/internal/service/reservation_type_liff_service.go`
- `backend/internal/repository/reservation_type_liff_repository.go`

## 問題概要
`Delete` の FK 依存チェックが service 層と repository 層の両方で実装されており二重になっている。

- **Service 層**: `ExistsByReservationTypeID` で FK 参照存在確認 → `WrapConflict` を返す
- **Repository 層**: `Delete` 内で FK 制約エラーを `isFKConstraintErr` でキャッチ → 別途 Conflict エラーを返す

二重検知は以下の問題を生む:
1. エラーメッセージが発生箇所によって異なる可能性がある
2. FK 制約エラーのキャッチを repository 層で行うと、service/repository の責任分離が崩れる
3. race condition（service チェック通過後に別トランザクションが FK レコードを作成）時は
   repository の FK エラーでフォールバックできるため、実害はないが実装の意図が不明確

## 推奨修正方針

**Service 層チェックのみに一本化**:

```go
// service 層（維持）: 存在確認
count, err := s.repo.CountUsageByReservationTypeLiffID(ctx, id)
if count > 0 {
    return apperrors.WrapConflict("この予約コースは使用中のため削除できません")
}

// repository 層（削除）: isFKConstraintErr のキャッチを除去
// → FK エラーは DB 制約として残るが、service 層で先行チェック済みのため通常は発生しない
```

ただし race condition への防御として repository 側も残す場合は、
両者のエラーメッセージを統一し、コメントで設計意図を明記すること。

## 完了条件
- [ ] repository 層の `isFKConstraintErr` 処理を除去、または設計意図をコメントで明記
- [ ] エラーメッセージが一箇所から返ることを確認
- [ ] `go test ./backend/internal/...` がパス
