# STG デモデータ seed

> **目的**: クライアントが STG で主要業務を試せるよう、合成のマスタ・飼主・ペット・予約を API 経由で投入する。
> **読者**: STG 運用担当・承認済み operator。
> **タイミング**: schema 再構築と保持データ restore の後。

## 現在の範囲（W3a）

- 医院スコープ master: procedures / medicines / vaccines / insurances / diagnosis types・names / chief complaints / cages / hospitalization plans / inventory / inquiry templates / checkup types / reservation type groups
- 合成データ: 飼主3件、ペット4件、予約3件
- W3b-1: 先頭予約のカルテに診療計画・バイタル・処置2件を追加。見積1件、会計3件（未納 waiting / 完了 completed / 部分入金待ち waiting）を作成する。
- W3b-2: 予防接種、健診、検査（pending）、トリミング（confirmed）を追加する。
- W3b-3: 入院（admitted）を追加する。cleanup では discharged に変更し、その飼主・ペット・予約・カルテと一緒に履歴として残す（削除しない）。
- 対象医院: `--clinic-id`（既定 1）。専用デモ医院は作らない。
- すべて `【デモ】` prefix と `EMR443-DEMO` run tag を付ける。実在の氏名・電話・メール・住所は使わない。

**在庫の入出庫**: 移動履歴 API がないため seed しない。在庫 item の数量更新は画面から確認できる。レポートは既存のカルテ・会計データで確認する。

## 前提

- API の demo/operator アカウントを使う。credential は環境変数で渡し、コマンドライン引数・ログ・git へ出さない。
- `--apply` を付けるまで API へ書き込まない（既定は dry-run）。
- 実行状態は `sensitive-local/demo-seed/EMR443-DEMO/state.json`（git 外、0600）に記録する。
- cleanup は **この run が作成した ID だけ**を削除する。既存の同名 master を更新した場合は削除しない。
- 実行は USER または承認済み operator。agent は実行しない。

## Dry-run

```bash
DEMO_SEED_EMAIL='<demo-or-operator-email>' \
DEMO_SEED_PASSWORD='<password>' \
python3 scripts/tools/seed-demo-data.py --clinic-id 1
```

投入対象の件数と API path を表示する。API へは接続しない。

## Apply

```bash
DEMO_SEED_EMAIL='<demo-or-operator-email>' \
DEMO_SEED_PASSWORD='<password>' \
python3 scripts/tools/seed-demo-data.py --clinic-id 1 --apply
```

- 既存 master は名前一致で PATCH、なければ POST。
- 飼主・ペットは名前一致で PATCH/POST。
- 予約は state file の記録で再実行を skip する。
- 失敗時は HTTP status と response の先頭 300 文字を出して停止する。

## Cleanup

```bash
DEMO_SEED_EMAIL='<demo-or-operator-email>' \
DEMO_SEED_PASSWORD='<password>' \
python3 scripts/tools/seed-demo-data.py --clinic-id 1 --cleanup --apply
```

予約 → ペット → 飼主 → デモ master の順に削除する。既存データに参照されている場合は削除に失敗するため、失敗を消さずに報告する。

## ローカル検証

STG 適用前に、同じ script をローカル API へ向けて dry-run / apply / cleanup で確認する。

```bash
DEMO_SEED_BASE_URL='http://localhost:8080' \
DEMO_SEED_EMAIL='...' \
DEMO_SEED_PASSWORD='...' \
python3 scripts/tools/seed-demo-data.py --clinic-id 1 --apply
```

## 注意

- 予約の日時は JST の当日 10:00 から1日ずつずらす。クライアントのテスト日時と重なる場合は `--clinic-id` と合わせて調整する。
- master の `【デモ】` 名は cleanup の識別子でもある。手動で名前を変えると cleanup 対象から外れる。
- 予約作成時にカルテが自動生成される。cleanup は、この run が作成した予約に紐づく自動生成カルテを先に削除してから予約・ペット・飼主を削除する。
- 2026-10-19 にローカル API で W3a〜W3b-3 の apply → cleanup を実行し、両方 EXIT 0 を確認した。cleanup 後、入院シナリオ（飼主・ペット・予約・カルテ・入院）は discharged の履歴として残り、それ以外のデモ行は削除される。STG 実適用は未検証。初回 apply は必ずローカルまたは承認済みの window で行う。
- 部分入金は `PATCH /accountings/{id}` では確定できず、`POST /accountings/complete` が全額決済を要求する。デモでは waiting の未納として扱い、partial の warning を許容する。
- トリミングは reservation_type_id / staff capability / start_time / end_time が必要。seeder は一時的に capability を追加し、cleanup で元へ戻す。
- 入院を残すため、cleanup は入院シナリオの飼主・ペット・予約・カルテを削除しない。このシナリオが参照する master は 409 の WARN として残る。
- 検査は `/masters/examination-types` の先頭種別を pending で作成する。
