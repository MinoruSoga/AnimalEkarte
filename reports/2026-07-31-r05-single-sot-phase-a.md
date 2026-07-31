# R-05 Channel Secret single-SoT Phase A

> 日付: 2026-07-31
> Binding decision: `reports/2026-07-31-line-residual-po-decisions-FINAL.md:55-80`
> SoT: `clinic_integrations` / `service=lstep` / `key_name=line_channel_secret`
> 実装状態: **CODE GREEN / ROLLOUT HOLD**

## 結論

Webhook verifier を、旧 reservation credential payload を読む経路から次へ切り替えた。

```text
destination
  -> line_reservation_settings: clinic_id + legacy credential presence のみ
  -> legacy presence=true なら HOLD（canonical lookup、decrypt、HMAC はゼロ）
  -> clinic_integrations: clinic_id + service + key_name で1件取得
  -> identity 再検証
  -> non-nil cipher で strict AES-GCM decrypt 最大1回
  -> HMAC 最大1回
```

旧値/raw payload fallback、複数 credential 試行、全 clinic scan、自動 winner 選択はない。cipher 欠落または復号失敗時は HMAC を実行しない。短 TTL cache は有効期限切れまたは暗号文変更を検出すると、旧平文 entry を復号より前に削除し、復号成功後だけ置換する。

## Code inventory

| Surface | 現在の owner / 動作 | Phase A disposition | Evidence |
|---|---|---|---|
| canonical write | L-step 設定 API が入力を暗号化し、`clinic_integrations` を `(clinic_id, service, key_name)` で upsert | 唯一の正規 write owner として維持 | `backend/internal/lstep/lstep_settings_update.go:12-43` |
| canonical settings read | L-step 設定 service が clinic + service の設定を読み、masked response を構築 | 既存 UI 用として維持 | `backend/internal/lstep/lstep_settings_service.go:182-205` |
| canonical verifier read | webhook route の clinic 確定後、service/key を含む clinic-scoped query で1件だけ取得 | **追加** | `backend/internal/lstep/lstep_settings_repository.go:27-45`; `backend/internal/lstep/line_link_service.go:425-447` |
| legacy request | reservation 設定 request は旧 field を引き続き受け入れる | **Phase B residual** | `backend/internal/reservation/line_reservation_setting_request.go:37-40` |
| legacy service read/write | reservation Save は空入力時に既存旧 credential を読み、再暗号化して保存する | **Phase B residual** | `backend/internal/reservation/line_reservation_setting_service.go:140-160,193-200` |
| legacy repository write | upsert column list に旧 credential column が残る | **Phase B residual** | `backend/internal/reservation/line_reservation_setting_repository.go:92-127` |
| destination routing | 旧 row 全体ではなく `clinic_id` と旧 credential の存在フラグだけを SELECT/return | **Phase A で置換** | `backend/internal/reservation/line_reservation_setting_repository.go:40-65` |
| webhook credential verification | canonical row identity を再検証し、strict decrypt に成功した canonical payload のみを HMAC | **Phase A で置換** | `backend/internal/lstep/line_link_service.go:116-140,425-447` |
| decrypted-secret cache | 有効期限切れまたは暗号文不一致の entry を復号前に削除し、成功時だけ短 TTL で置換 | **review repair** | `backend/internal/lstep/line_link_service.go:469-510` |

Production Go source の明示的な旧 credential 参照は上表の request/service/repository と presence query に残る。Phase A は verifier cutover までであり、reservation writer removal や旧 column DROP を完了扱いにしない。

## Fail-closed contract

- destination 欠落・不正、route not found/DB error、clinic ID 不正は不成立。
- 旧 credential が存在する clinic は mismatch/equality を推測せず、canonical lookup より前に不成立。
- canonical row 欠落・DB error・空 payload・nil cipher・AES-GCM decrypt failure は不成立。
- canonical decrypt は generic helper の plaintext/raw fallback を使用せず、失敗値を HMAC または log へ渡さない。
- canonical row の clinic/service/key identity が要求値と一致しなければ不成立。
- route DB error は canonical lookup/decrypt/HMAC を実行せず不成立。
- clinic A query に対して clinic B にしか一致 row がない場合は NotFound となる。
- verifier は `FindAll` を呼ばず、credential query/decrypt/HMAC はそれぞれ最大1回。
- error、log、本 report に credential 値、可逆値、比較 digest を出さない。

## TDD evidence

### RED

Commit: `73eaa2559 test: specify R-05 canonical credential cutover`

```text
docker compose -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-r05/backend:/app backend test ./internal/lstep ./internal/reservation -run 'TestVerifySignatureAnyClinic_(LegacyCredentialPresent|CanonicalCredentialIdentityMismatch|MissingCanonicalCredential|ValidDestination)|TestLstepSettingsRepository_FindCredentialByClinicServiceKey|TestLineReservationSettingRepository_FindWebhookRouteByLineBotUserID' -count=1 -p 1
```

Exit `1`: constructor/struct に canonical reader がなく、両 repository method も未定義だった。

### GREEN

同一 command: exit `0`。

```text
ok github.com/animal-ekarte/backend/internal/lstep 0.155s
ok github.com/animal-ekarte/backend/internal/reservation 0.184s
```

追加回帰:

- LINE link 関連 test: exit `0`, `ok .../internal/lstep 0.352s`
- reservation package 全体: exit `0`, `ok .../internal/reservation 10.580s`
- `go build ./internal/lstep ./internal/reservation`: exit `0`
- `go vet ./internal/lstep ./internal/reservation`: exit `0`
- changed Go files: Docker `gofmt` 済み

### R-SEC adversarial repair

Review 後に、generic `DecryptLineCredential` の legacy plaintext fallback が canonical verifier に残っていた点と、期限切れ/暗号文不一致 cache entry の復号前 eviction が未証明だった点を追加で falsify した。

RED commits / observed failures:

- `ffc451601 test: reject malformed canonical LINE credential`: malformed ciphertext で期待 HMAC `0` に対し実測 `1`、exit `1`。
- `6fa5e7f72 test: require stale LINE credential cache eviction`: expired / ciphertext mismatch の両 case で decrypt 中にも旧 entry が存在し、exit `1`。
- `e22a6af7d test: harden canonical LINE credential proofs`: nil cipher で期待 HMAC `0` に対し実測 `1`、exit `1`。同 commit で正常 routing fixture を実 AES-GCM ciphertext に変更し、route DB error と foreign-only clinic row の決定的 proof を追加した。

GREEN implementation: `e0076e722 fix: fail closed on canonical LINE decrypt errors`

```text
docker compose -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-r05/backend:/app backend test ./internal/lstep -run '^(TestVerifySignatureAnyClinic_(MalformedCanonicalCiphertext_FailsBeforeHMAC|NilCipherFailsBeforeHMAC|FindWebhookRouteError)|TestCachedDecryptChannelSecret_EvictsStaleEntryBeforeDecrypt)$' -count=1 -p 1
```

Exit `0`: `ok github.com/animal-ekarte/backend/internal/lstep 0.003s`。

Canonical routing + clinic-isolation focused matrix:

```text
docker compose -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml run --rm --no-deps -T --entrypoint go -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-r05/backend:/app backend test ./internal/lstep ./internal/reservation -run 'TestVerifySignatureAnyClinic_(LegacyCredentialPresent|CanonicalCredentialIdentityMismatch|MissingCanonicalCredential|MalformedCanonicalCiphertext|NilCipher|ValidDestination|FindWebhookRouteError)|TestCachedDecryptChannelSecret_EvictsStaleEntryBeforeDecrypt|TestLstepSettingsRepository_FindCredentialByClinicServiceKey|TestLineReservationSettingRepository_FindWebhookRouteByLineBotUserID' -count=1 -p 1
```

Exit `0`: L-step `0.198s`; reservation `0.171s`。この run は foreign-only clinic credential case も test-name prefix で含む。

追加 review regression:

- LINE link 関連 test: exit `0`, `ok .../internal/lstep 0.385s`
- `go vet ./internal/lstep ./internal/reservation`: exit `0`
- `go build ./internal/lstep ./internal/reservation ./cmd/api`: exit `0`
- review repair 対象 Go files: Docker `gofmt` 済み

指定の combined package command は exit `1`。変更箇所の compile/assertion failure ではなく、既存の L-step transaction atomicity 3 tests が test DB AutoMigrate の `ClinicSettings` time default を `timestamptz` として扱い SQLSTATE `22007` で失敗した。reservation package は同じ run で GREEN。Coordinator が untouched main `e1b51ceae` でも同じ command を実行し、同じ3 failures と reservation GREEN を再現したため、baseline-relative failure と確認済み。したがって combined package 全体は **SCOPED-GREEN / PACKAGE-GREEN 未達** と記録し、全体 GREEN を主張しない。

Review repair では、この既知 baseline failure の full L-step package command は再実行せず、上記の falsification-focused matrix と build/vet を実行した。

## Rollout gate

現在の旧 column に credential が残る clinic は verifier が意図的に HOLD する。よってこの commit をそのまま production deploy してよいとは判定しない。

Phase B 前提:

1. clinic ごとに `empty/equal/mismatch` の状態だけを inventory する。値・暗号文・digest は保存しない。
2. 旧値のみ・mismatch は自動採用せず、正規 L-step 設定 UI と actor/audit を伴う再設定へ送る。
3. reservation request/service/repository の旧 credential read/write を撤去する。
4. verifier/runtime 証拠と source inventory がゼロになった後だけ、別 packet で numbered DROP migration を提案する。

## Safety / operations

- migration/DROP は作成していない。`make migrate`、DB reset、production LINE 操作、push は実施していない。
- 最初の formatter one-off は app の既定 entrypoint を迂回しておらず migration runner が起動したが、migration summary は `applied=0` で全件 skip、既存 seed checksum mismatch で command が停止した。以後は全 command に `--entrypoint gofmt` または `--entrypoint go` を明示した。
- Initial implementation commit: `c45eea3f0 fix: cut webhook verification to canonical credential`
- Review repair implementation commit: `e0076e722 fix: fail closed on canonical LINE decrypt errors`
