# R-05 Channel Secret single-SoT Phase B

> 日付: 2026-07-31
> Binding decision: `reports/2026-07-31-line-residual-po-decisions-FINAL.md` / Phase A residual
> SoT: `clinic_integrations` / `service=lstep` / `key_name=line_channel_secret`
> 実装状態: **CODE GREEN (reservation write path) / DROP HOLD / composition residual**

## 結論

reservation 設定 API から legacy Channel Secret の **受付・再暗号・upsert 書込** を撤去した。

```text
reservation settings Save
  -> request: line_channel_secret を DTO から削除（未知 JSON は無視）
  -> service: LineChannelSecret を input / model に載せない
  -> repository OnConflict: line_channel_secret を更新列から除外
  -> DB 既存列値は温存（presence SELECT 用）
  -> 正規 write owner は引き続き L-step 設定 API → clinic_integrations
```

Webhook verifier は Phase A のまま **clinic_integrations-only**。本 Phase は dual-write / dual-fallback を再導入しない。

## Residual matrix

| Surface | 状態 | Disposition | Evidence |
|---|---|---|---|
| reservation request `line_channel_secret` | 削除済み | **DONE (Phase B)** | `line_reservation_setting_request.go` |
| reservation service encrypt/decrypt fallback (secret) | 削除済み | **DONE (Phase B)** | `line_reservation_setting_service.go` |
| reservation repository upsert column | 除外済み | **DONE (Phase B)** | `lineReservationSettingUpdatableColumns()` |
| presence SELECT `(line_channel_secret <> '')` | 維持 | **KEEP** | `FindWebhookRouteByLineBotUserID` |
| `line_bot_user_id` routing | 維持 | **KEEP** | same repository method |
| response 非漏洩 (secret/token 非出力) | 維持 | **KEEP** | `line_reservation_setting_response.go` |
| LineAccessToken empty-fallback + encrypt | 維持 | **KEEP (co-travel)** | service Save |
| model / DB column `line_channel_secret` | 残存 | **HOLD** | `model.LineReservationSetting` |
| DROP migration (column 削除) | 未実施 | **HOLD** | inventory zero 後の別 packet |
| `cmd/api/composition_reservation_test.go` | 旧期待（rewired secret） | **RESIDUAL** | allowlist 外。merge 前に assertion 更新要 |
| legacy inventory empty/equal/mismatch | 運用 | **HOLD** | Phase A rollout gate 継続 |

## Fail-closed / safety

- Verifier へ reservation secret を再配線しない。
- 本経路の log に credential 値を出さない（既存どおり clinic_id のみ）。
- 空の `line_channel_secret` JSON を送っても binding は成功し、値は破棄される（未知キー）。
- update 時に zero model を渡しても OnConflict が secret 列を触らない（repository test で証明）。
- DROP / `make migrate` / production 操作 / push は実施していない。

## TDD evidence

### 変更ファイル（allowlist）

1. `backend/internal/reservation/line_reservation_setting_request.go`
2. `backend/internal/reservation/line_reservation_setting_request_test.go`
3. `backend/internal/reservation/line_reservation_setting_service.go`
4. `backend/internal/reservation/line_reservation_setting_service_test.go`
5. `backend/internal/reservation/line_reservation_setting_repository.go`
6. `backend/internal/reservation/line_reservation_setting_repository_test.go`
7. `reports/2026-07-31-r05-single-sot-phase-b.md`

### 期待テスト

- request: secret を toServiceInput しない / 未知 JSON 無視
- service: secret を write model に載せない / access token の encrypt・empty preserve は継続
- repository: OnConflict で secret 列を潰さない / updatable list に secret が無い

### 検証コマンド（worktree mount）

```text
# reservation LineReservation* — exit 0
docker compose -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go \
  -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-r05b/backend:/app backend \
  test ./internal/reservation/ -count=1 -run 'TestLineReservationSetting'
# → ok github.com/animal-ekarte/backend/internal/reservation 0.549s

# lstep verifier-focused — exit 0
... test ./internal/lstep/ -count=1 -run 'TestVerifySignature|TestCachedDecrypt|TestLstepSettingsRepository_FindCredential|TestLineLinkService_HandleWebhook'
# → ok github.com/animal-ekarte/backend/internal/lstep 0.331s

# combined broad -run (includes baseline atomicity AutoMigrate fail) — exit 1 baseline
# TestLstepSettingsService_UpdateSettings_RollsBackCredentialsWhenClinicConfigFails
# SQLSTATE 22007 on clinic_settings time default — Phase A と同型の baseline residual

# composition residual (allowlist 外) — exit 1 expected
... test ./cmd/api/ -count=1 -run 'TestNewReservationComposition_InjectsLineCredentialCipherClosures'
# expected "rewired:secret" actual "" — Phase B 契約へ assertion 更新が必要
```

## 次の packet（HOLD 解除条件）

1. clinic ごとの legacy presence inventory が **ゼロ**（値・digest は保存しない）。
2. composition wiring test を Phase B 契約へ更新（allowlist 拡張 or 別 packet）。
3. 上記後に限り numbered DROP migration を提案（本 packet では作成しない）。
