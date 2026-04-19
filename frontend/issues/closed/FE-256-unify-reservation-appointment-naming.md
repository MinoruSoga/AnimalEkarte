# FE-256: reservation / appointment 命名を統一

**Status**: Open  
**Priority**: Medium  
**Type**: Refactor  
**Date Created**: 2026-04-19  
**Related**: BE-256（BE 側対応と同時進行を推奨）

## 背景

「予約」エンティティを `reservation` と `appointment` の2語で呼び分けている。
BE のモデル名・DB テーブルは `Appointment` / `appointments` だが、
API ルートと FE は `reservations` を使用しており、コードを読む際に混乱が生じている。

## 現状の不統一

| 場所 | 語 |
|------|-----|
| DB テーブル | `appointments` |
| BE Go struct | `type Appointment struct` |
| BE handler ファイル名 | `appointment_handler.go` |
| BE API ルートパス | `/v1/reservations`（`handler.go`） |
| FE feature ディレクトリ | `features/reservations/` |
| FE `paths.ts` キー | `reservations` |
| FE URL パス | `/reservations` |
| FE 型名（一部） | `interface Appointment`（`types/index.ts`） |

## 対応方針

**`reservation` に統一する**（API パス・FE 語彙を基準）。

理由: ユーザー可視 URL が `/reservations` で既に公開されており、変更コストが高い。
BE 内部の `Appointment` struct / table は DB マイグレーションが不要なため
「内部実装は `appointment` のまま、外部公開語彙は `reservation`」でも許容できるが、
将来的な混乱を避けるため struct 名も `Reservation` へのリネームを推奨する。

## 変更対象

### FE 側
- `features/reservations/api/types.ts` 内の `Appointment` 型参照を `Reservation` に統一
- `types/index.ts` の `interface Appointment` を `interface Reservation` にリネーム（影響範囲確認必要）
- `features/reservations/` 内部で `appointment` を使っている変数名・関数名を `reservation` に統一

### BE 側（BE-256 として別起票）
- `appointment_handler.go` → `reservation_handler.go` リネーム（内部のみ）
- Go struct `Appointment` → `Reservation`（GORM TableName で `appointments` テーブルは維持）
- `AppointmentService` → `ReservationService`
- `AppointmentRepository` → `ReservationRepository`

## 完了条件

- [ ] FE の型名・変数名が `reservation` に統一される
- [ ] `paths.reservations` との整合性確認
- [ ] lint / 型チェック / ビルドが通る

## 注意事項

- DB テーブル名 `appointments` は変更しない（マイグレーション不要）
- BE の API ルートパス `/v1/reservations` は変更しない（クライアント互換維持）
