# QA-UAT-S09-FIXTURE — `completed_at` fixture 設計

状態: **設計完了 / package helper 実装済み / HTTP・CLI・ブラウザ再実行は未 / UAT は BLOCKED**  
正本シナリオ: [S09-closing-time-boundaries.md](./scenarios/S09-closing-time-boundaries.md)  
設計・package test だけでは S09 を PASS にしない。以下の「推奨実装」「cleanup」は到達目標であり、現行 package がすべて実装しているという意味ではない。

## 現行 package と残る接続（2026-09-06 ソース照合）

- `backend/internal/billing/synthetic_closing_fixture.go` の `CreateSyntheticClosingFixture` は、新規 company/clinic/settings/owner/species/pet と指定 5 時刻の completed billing **ヘッダ**を作る。既存 billing ID の指定を拒否し、既存会計の UPDATE はしない。
- `synthetic_closing_env.go` は `APP_ENV=test/development/local/dev` と DB host `db/localhost/127.0.0.1` の入力値を検証する。HTTP hostname の検証や接続済み DB の同一性検証は未実装。呼び出し側から渡す値だけで実接続の安全を証明しない。
- 現行 helper は staff/account、支払方法、明細、payment splits、cleanup token を作らず、transaction wrapper・teardown も持たない。途中失敗時の原子性と回収は未確立。
- HTTP route/CLI はなく、`POST /api/v1/uat/synthetic-closings` は下記の提案名のまま。ブラウザ用 identity と支払方法別の理論売上 fixture も未接続のため、S09 #2–#6 は **BLOCKED を維持**する。
- `synthetic_closing_fixture_test.go` は unsafe request 拒否と 5 ヘッダの時刻/settings を検証する。HTTP/CLI・締めプレビュー集計・ブラウザ UAT・cleanup の成功証拠ではない。

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

## 推奨実装（承認後）

**scoped UAT HTTP helper** を選ぶ。ブラウザ UAT が #2〜#6 を実行するため。Go `testdb` 直書きは「直接 DB 更新」に当たるため採用しない。

| 項目 | 内容 |
|------|------|
| 経路 | `POST /api/v1/uat/synthetic-closings`（仮。実装時に OpenAPI へ載せる） |
| 入力 | 対象日（JST 暦日）、5 時刻、clinic 作成フラグ |
| 出力 | 使い捨て `clinic_id`、5 件の billing id、`completed_at`、cleanup token。金額の実値は最小合成値 |
| 内部 | 現行 `POST /accountings/complete` 相当のトランザクションを呼び、`completed_at` だけテスト用引数で上書きする。generic PATCH は使わない |
| 登録 | `APP_ENV` ゲートを通ったときだけ router に載せる。staging/production バイナリに経路が残っても handler が 404 |

代替（局所 Go 証明のみ）: `testdb` に `MakeCompletedBillingAt` を足し、`billing` パッケージの帰属テストで半開区間を固定する。これは S09 ブラウザ受入の代替ではない。

## 変更ファイル（実装承認後）

- `backend/internal/billing/` — complete 内部の時刻注入口（本番 complete はサーバ時刻のまま）
- `backend/internal/` の UAT helper 登録（env gate）
- fail-closed テスト（staging DSN / 既存 billing id / 時計変更フラグ）
- `backend/docs/api.yaml`（経路を載せる場合）
- 本ファイルの「実装検証」節（承認後に追記）

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
docker compose exec backend go test ./internal/billing/... -count=1 -run 'TestCreateSyntheticClosingFixture|TestAllowUATSyntheticClosing|TestRejectExistingBillingIDs|TestRejectReservedClinicID'
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
| package helper と局所テスト | 実装済み。HTTP/CLI・原子性・支払明細・cleanup は残る |
| ブラウザ接続と回収経路の追加 | 残る仕様と scope を確定した後 |
| 追加 HTTP/CLI・回収経路の main 統合 | USER（既存 package helper は `d4c870f9e` で source に収録済み） |
| S09 #2〜#6 の UAT 再実行 | 統合後の別承認 |
| UAT 集計の PASS 更新 | 再実行証跡後。設計だけでは更新しない |
| Linear Done | USER |
