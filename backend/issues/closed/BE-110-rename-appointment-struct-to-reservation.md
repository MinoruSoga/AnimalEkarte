# BE-110: Appointment 構造体・サービス・リポジトリを Reservation に統一

**Status**: Open  
**Priority**: Medium  
**Type**: Refactor  
**Date Created**: 2026-04-19  
**Related**: FE-256

## 背景

予約エンティティを BE 内部で `Appointment`、API ルート・FE で `reservation` と呼び分けている。
コードを読む際の認知負荷を下げるため、公開 API パス（`/v1/reservations`）に合わせて
BE 内部の命名も `Reservation` に統一する。

## 現状の不統一

| 場所 | 現在の語 |
|------|---------|
| DB テーブル | `appointments`（変更しない） |
| Go struct | `type Appointment struct`（`reservation.go`） |
| Go struct | `type AppointmentStatus string` など |
| ハンドラ | `appointment_handler.go` 内の関数名 `ListReservations` など混在 |
| サービス I/F | `AppointmentService` |
| リポジトリ I/F | `AppointmentRepository` |
| API ルートパス | `/v1/reservations`（変更しない） |

## 対応方針

1. Go struct `Appointment` → `Reservation`  
   - GORM `TableName()` で `"appointments"` を返し続けることで DB 変更なし  
2. `AppointmentStatus` → `ReservationStatus` など関連型も同様にリネーム  
3. `appointment_handler.go` 内の関数名を `ListReservations`, `GetReservation` など統一  
4. `AppointmentService` → `ReservationService`  
5. `AppointmentRepository` → `ReservationRepository`  

## 変更対象ファイル

- `backend/internal/model/reservation.go`（struct 名・型名リネーム）
- `backend/internal/handler/appointment_handler.go`（→ `reservation_handler.go` にリネーム）
- `backend/internal/service/appointment_service.go`（→ `reservation_service.go`）
- `backend/internal/repository/appointment_repository.go`（→ `reservation_repository.go`）
- `backend/cmd/api/main.go` / DI 配線の参照更新

## 完了条件

- [ ] struct / 型名が `Reservation` に統一
- [ ] ファイル名が `reservation_*.go` に統一
- [ ] `go build ./...` が通る
- [ ] 既存テスト（`*_test.go`）がパス
- [ ] DB テーブル名 `appointments` は変更していない

## 注意事項

- DB テーブル名 `appointments` は変更しない（マイグレーション不要）
- API ルートパス `/v1/reservations` は変更しない（クライアント互換維持）
- `appointment_trimming_details` テーブルおよび関連 struct は別途検討
