# QA-UAT-S09-FIXTURE — `completed_at` fixture 設計

状態: **HTTP・CLI・原子性・staff/支払/明細/cleanup 実装済み / ブラウザ再実行と S09 PASS は未**  
正本シナリオ: [S09-closing-time-boundaries.md](./scenarios/S09-closing-time-boundaries.md)  
設計・package/HTTP test だけでは S09 を PASS にしない。ブラウザ #2–#6 の再実行証跡が必要。

## 現行 package と残る接続（2026-09-08 ソース照合）

- `CreateSyntheticClosingFixture` は transaction 内で新規 company/clinic/settings/staff(account, system admin)/owner/species/pet/会計ヘッダ/明細/payment/payment_splits と指定 5 時刻の completed billing を作る。支払方法は clinic INSERT の `trg_create_default_payment_methods` が入れた `cash` を再利用し、testdb のように trigger が無いときだけ INSERT する。既存 billing ID の指定を拒否し、既存会計の UPDATE はしない。平日以外の対象日は拒否する。
- `synthetic_closing_env.go` は `APP_ENV=test/development/local/dev`、DB host `db/localhost/127.0.0.1`、HTTP host `backend/localhost/127.0.0.1` を fail-closed で判定する。接続済み DB の hostname 同一性検証は呼び出し側が渡す `DB_HOST` に依存する。
- HTTP: `POST /api/v1/uat/synthetic-closings` と `DELETE /api/v1/uat/synthetic-closings/:clinic_id`（`X-UAT-Cleanup-Token`）。staging/production と許可外 HTTP host は 404。ログインパスワードは応答に出さず `UAT_SYNTHETIC_CLOSING_PASSWORD` からハッシュする。
- CLI: `backend/cmd/synthetic-closing-fixture` の `setup` / `teardown`。
- ブラウザ UAT（S09 #2–#6）と締めプレビュー集計の目視は未。S09 は **BLOCKED を維持**する。

## 対象 scenario

S09 #2〜#6（帰属プレビュー）。#1・#7〜#10 は既存会計を改変せず実施可能な範囲として、helper 承認後に再実行する。

必要な合成会計は **新規 5 件**。完了時刻は JST で 10:00 / 13:30 / 14:00 / 20:00 / 翌 02:00。対象日は平日かつ休診日でない使い捨て日。

## 使い捨て環境・clinic

| 項目 | 値 |
|------|-----|
| 環境 | ローカル Docker のみ。hostname が compose の `backend` / `localhost` 以外なら拒否 |
| `APP_ENV` | `test` または `development` / `local` / `dev` のみ。`staging` / `production` / 空 / 不明は拒否 |
| DB | compose の DB。package が受理する host は `db` / `localhost` / `127.0.0.1`。`ekarte_db` は DB/container 名であり host allowlist の値ではない |
| clinic | 新規合成 clinic。八王子 `1`・城東 `2`・既存 UAT clinic は使わない |
| アカウント | その clinic に attach した合成 staff（cash-register-close） |
| 締め設定 | AM 開始 09:00、境界 13:30、平日終了 19:00 |

## 合成データ

helper が **新規** に作るものだけを使う。

- company / clinic / staff / owner / pet / 支払方法 / 会計ヘッダ / 明細 1 行以上
- `completed_at` は helper 内部で設定する。クライアント generic PATCH / legacy create では設定できない（現行ゲート: `accounting_service_core.go`）
- 既存 `billings.id` の UPDATE はしない
- 行値・PII・credential は証跡に残さない

## 実装検証（2026-09-08）

入れたもの:

- transaction 付き fixture、staff/account、支払方法、明細、payment、payment_splits、cleanup token
- `POST/DELETE /api/v1/uat/synthetic-closings` と `cmd/synthetic-closing-fixture`
- HTTP host allowlist と staging/production の 404

未実施（PASS にしない）:

- ブラウザでの S09 #2–#6
- `make codegen`
- Linear Done


## 推奨実装（到達目標）

**scoped UAT HTTP helper** を選んだ。ブラウザ UAT が #2〜#6 を実行するため。既存会計の testdb 直書き改変はしない。

| 項目 | 内容 |
|------|------|
| 経路 | `POST /api/v1/uat/synthetic-closings` / `DELETE /api/v1/uat/synthetic-closings/{clinic_id}` |
| 入力 | 対象日（JST 暦日）。既存 billing ID があれば拒否。完了時刻は helper 固定 |
| 出力 | 使い捨て `clinic_id`、loginEmail、5 件の billing id、`completed_at`、cleanup token。パスワードは返さない |
| 内部 | 新規 completed billing を transaction で INSERT。generic PATCH は使わない。本番 complete に clock seam は入れない |
| 登録 | router には載る。staging/production と許可外 HTTP host は handler が 404 |

## 変更ファイル

- `backend/internal/billing/synthetic_closing_*.go`
- `backend/cmd/synthetic-closing-fixture/`
- `backend/cmd/api/composition_runtime.go`
- `backend/docs/api.yaml`
- `backend/internal/apicontract/openapi_route_drift_test.go`

シナリオ本文は編集しない。

## 許可操作 / 禁止操作

許可:

- 新規合成 clinic / 会計の作成
- 締めプレビュー GET
- 対象 clinic に限った cleanup DELETE

禁止:

- 既存会計の改変
- `UPDATE billings SET completed_at`
- システム時計変更、コンテナ TZ 改変
- 共有 STG / PROD / PlanetScale
- helper 不在での S09 PASS 宣言

## 環境誤指定時の停止

起動またはリクエスト時に次のいずれかなら **即 404/拒否**。部分成功を残さない。

1. `APP_ENV` が allowlist 外
2. DB host が `db` / `localhost` / `127.0.0.1` 以外
3. 入力に既存 `billing_id` がある
4. clinic が予約済み本番/STG ID（1, 2 および設定された除外集合）

## cleanup / 失敗時回収

- 応答の cleanup token で、作成した clinic 配下を一括削除する
- プロセス異常時は clinic 単位で再 sweep
- append-only 締めを #7 まで進めた場合は、その clinic ごと破棄する（close の reverse API は無い）
- sweep 失敗は BLOCKED。共有 DB の手削除はしない

## 局所検証コマンド（実装後）

```bash
docker compose exec backend go test ./internal/billing/... -count=1 -run 'TestCreateSyntheticClosingFixture|TestAllowUATSyntheticClosing|TestRejectExistingBillingIDs|TestRejectReservedClinicID|TestRegisterUATRoutes|TestSyntheticClosingCleanupToken'
docker compose exec backend go test ./cmd/synthetic-closing-fixture/ -count=1
docker compose exec backend go test ./internal/apicontract/ -count=1 -run 'TestOpenAPIRouteDrift_MatchesAllowlist'
```

フロントを触った場合のみ:

```bash
docker compose exec frontend npx vitest run <変更spec>
```

全体 `go test ./...`、migration apply、shared STG、S09 ブラウザ再実行は自動実行しない。

## 承認が必要な実行操作

| 操作 | 承認 |
|------|------|
| 本設計 | source 定義済み。受入 sign-off とは別 |
| package helper と局所テスト | 実装済み（staff/支払/明細/原子性/cleanup を含む） |
| HTTP/CLI・回収経路 | 実装済み。`make codegen` は未実行（USER） |
| S09 #2〜#6 のブラウザ UAT 再実行 | ローカル stack 起動後の別承認。Docker 停止中は実行しない |
| UAT 集計の PASS 更新 | 再実行証跡後。helper だけでは更新しない |
| Linear Done | USER |
