# STG デモデータ seed（004_demo CSV bundle）

> **目的**: クライアントが STG で主要業務を試せるよう、デモデータを CSV seed bundle
> （`backend/migrations/seeds/004_demo/`）としてリポジトリに同梱し、STG の DB reset 後の
> `cmd/migrate` が自動適用する。個別の投入作業は不要。
> **読者**: STG 運用担当・承認済み operator。
> **タイミング**: STG の承認済み DB reset（再構築）の実行前。bundle は migrate のロード
> 順に従い、login seed の後で自動適用される。

## 方式

- デモデータは **CSV seed bundle**（`backend/migrations/seeds/004_demo/` の各 CSV +
  `manifest.json`）。旧 `003_demo`（2026-08 退役）と同じ形式の後継だが、schema が古い
  ため再利用せず、現在の schema から新規生成する。
- **credentials は含まない**。`accounts` / `staffs` / 職員割当は login seed（phase 3、
  `APP_ENV=staging` のみ）が所有し、デモ行はその固定 band id の staff を参照する。
  004_demo は migrate の **login seed の後段 phase** でロードされる（staff FK のため）。
- **適用環境**: `APP_ENV=staging` のみ。production は常に master のみ（デモデータが
  production に乗る経路はない）。local dev も既定ではロードしない。
- 対象医院: clinic 1（八王子病院）。専用デモ医院は作らない。
- 全行に `【デモ】` の名前 prefix。実在の氏名・電話・メール・住所は使わない。

## 内容（W3a + W3b 相当）

- 医院スコープ master（procedures / medicines / vaccines / insurances / diagnosis
  types・names / chief complaints / cages / hospitalization plans / inventory /
  inquiry templates / checkup types / reservation type groups）
- 飼主 3 件・ペット 4 件・予約 3 件
- 診療計画・バイタル・処置・見積・会計 3 種（未納 waiting / 完了 completed / 部分入金待ち
  waiting）
- 予防接種・健診・検査（pending）・トリミング（confirmed）・入院（admitted → cleanup で
  discharged 履歴）
- 在庫の入出庫: 移動履歴 API がないため seed しない（数量は画面から確認）

詳細なテーブル一覧・行数は **`backend/migrations/seeds/004_demo/manifest.json` が正本**。

## 生成（`scripts/tools/gen-demo-seed-bundle.sh`）

デモデータは検証済みの API seeder（`scripts/tools/seed-demo-data.py`）をジェネレータに、
使い捨てコンテナで再生成する（CSV は実 DB の COPY dump が正本 — 手編集禁止のため、
再生成は常に本スクリプト経由で行う）:

1. 独立 `postgres:18-alpine` + `golang:1.26-alpine` コンテナ（専用 network、終了時削除）
2. `cmd/migrate`（001_init + 002_master + 003_login。使い捨てパスワードで login seed を
   適用し staff FK を作る — この行は dump されない）
3. `cmd/api` を起動 → `seed-demo-data.py --apply`（clinic 1）
4. テーブル別件数 diff でデモ所有テーブルを特定 → FK 位相ソート → PK 順 COPY dump +
   `manifest.json` 作成

既存の Compose スタック・local DB volume・STG・production には接触しない。

```bash
bash scripts/tools/gen-demo-seed-bundle.sh
```

**STG 反映直前には必ず再生成すること**: 予約日時等の業務日付は生成日基準のため、古い
bundle では「本日の予約」シナリオが過去日になる。

## STG 反映手順（bundle は自動適用）

1. 承認済み revision を固定し、その checkout を実行元にする
   （[STG_PLANETSCALE_SEED_RUNBOOK.md](STG_PLANETSCALE_SEED_RUNBOOK.md) §6.0）。
2. `scripts/old-db/stg-rebuild.sh` で reset + migrate を実行
   （`Migration key coverage` に `seeds/004_demo` が含まれることを確認）。
3. 保持データ restore + verify（[STG-PRESERVE-DATA.md](STG-PRESERVE-DATA.md)）。
   restore は業務テーブル（owners/pets/reservations/records）を含まないため、デモ行と
   衝突しない。
4. demo login 解錠（`SKIP_RESET=1` + `SEEDLOGIN_DEMO_PASSWORD` で migrate 再実行）。

reset → migrate（004_demo 含む）→ restore → 解錠の順。途中で pull・checkout 切替・
ファイル編集を行わない（§6.0 準拠）。

## API seeder の役割（ジェネレータ）

- `scripts/tools/seed-demo-data.py` は STG へ直接実行しない。bundle 再生成時の使い捨て
  DB への投入のみに使う。
- dry-run / apply / cleanup の各オプションは従来通り（再生成時のローカル検証用）。

## 注意

- 予約の日時は JST の当日 10:00 から 1 日ずつずらす。再生成タイミングで変動する。
- 部分入金は `PATCH /accountings/{id}` では確定できず、`POST /accountings/complete` が
  全額決済を要求する。デモでは waiting の未納として扱う。
- トリミングは reservation_type_id / staff capability / start_time / end_time が必要。
  seeder は一時的に capability を追加する（使い捨て DB 内なので後始末は不要）。
- 入院シナリオは cleanup で discharged の履歴として残す（削除しない）。
- 検査は先頭 examination-type を pending で作成する。
- 2026-10-19 にローカル API で W3a〜W3b-3 の apply → cleanup を EXIT 0 で確認済み。
  bundle 化は EMR-446 のモデル整合後（2026-10-20 現在 BLOCKED）に生成・コミット予定。
