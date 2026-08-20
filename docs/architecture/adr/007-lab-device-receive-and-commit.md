# ADR-007: 城東検査機器 — 受信永続化・1画面・即 persist + Undo

**Status**: Accepted（設計。実装は BRT-95〜98）
**Date**: 2026-08-19
**Deciders**: PO（MinoruSoga）
**Relates to**: ADR-002（clinic_id）、ADR-006（`medicalrecord` write owner）、Linear [BRT-100](https://linear.app/baritechllc/issue/BRT-100) / [BRT-94](https://linear.app/baritechllc/issue/BRT-94)
**接続・機器正本**: [`docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md`](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md)
**電文正本**: `old_db/docs/lab-go/go-impl/device-serial-adapter.md`
**体験正本**: `old_db/docs/lab-go/go-impl/REVIEW-FABLE-2026-08-19-AE-LAB-UX.md`

## Context

現行 `lab_import_jobs` は行カウンタと状態だけを持つ。`Preview` はステートレス、`Commit` はクライアントが `LabExamPersistInput`（`exam_type_id` 等）を渡すワンショット。受信フレームをペット選択まで置くテーブルが無い。

城東3台（NX600 / AU10V / PU-4010）は検査用 Mac の待機ページが有線シリアルを読む。ファイルアップロード・常駐デーモン・`drwan` は製品経路にしない。

Fable UX（YES-WITH-FIXES）: 日常は本日診療中カルテを1回選んで16項目手打ちを消す（ペット検索はしない。正本は LAB_DEVICE_CONNECTIVITY）。ただし前の子の待機が残ったまま次送信が届く誤紐付けを、確認ダイアログなしで塞ぐまで医院に出さない。

### 実践ゲート

| 項 | 内容 |
| --- | --- |
| 責任者 | MinoruSoga |
| 業務目的 | NX600 16項目の手打ち転記（40操作超+転記ミス）を消す |
| 削除する工程 | 手打ち、ファイル作成/選択、確認ダイアログ、待機と未紐付けの往復、日常のポート選択 |
| 残す操作 | 本日診療中カルテ選択1回。送信は機器側（従来どおり） |
| メトリクス | 追加操作は本日診療中カルテ選択1回。Undo / 待機解除 / 後付けは各1操作 |
| やらない | アップロード、取込値エディタ、遠隔待機起動、検体ID自動紐付け、未紐付け検索、常駐アプリ（試用で再送が週に何度も要ると観測されるまで） |

## Decision

### 1. 1フレーム = 1測定 = 1 `lab_import_jobs` 行

新しい「受信ヘッダ」テーブルは作らない。既存ジョブを **1受信フレーム** に使う。

`lab_import_jobs` に足す列（device 行だけ埋める。fixture は NULL / 空のまま）:

| 列 | 内容 |
| --- | --- |
| `pet_id` | 未紐付けは NULL。clinic 内 pets への FK |
| `measured_at` | 電文日時。検査日の正 |
| `received_at` | デコード成功時刻。監査のみ。検査日にしない |
| `device_hint` | `NX600` / `AU10V` / `PU-4010` |
| `specimen_id_raw` | 表示専用。紐付けキーにしない。ログに出さない |
| `unmapped_item_count` | マスタ未対応行数 |

項目は新テーブル `lab_import_job_items`:

| 列 | 内容 |
| --- | --- |
| `clinic_id` | NOT NULL。親ジョブと同じ |
| `job_id` | FK → `lab_import_jobs` |
| `device_item_code` | 電文コードのまま（`Na-P` を削らない） |
| `value_raw` | 文字列。`<3.75` 等を残す |
| `unit` / `flag` | 電文のまま。空可 |
| `exam_type_field_id` | マスタ解決後。未対応は NULL |
| `needs_review` | 未対応・欠測記号など |
| `sort_order` | 表示順 |

生バイト列は保存しない。デコード後に捨てる。

冪等:

```
UNIQUE (clinic_id, source_type, source_fingerprint)
  WHERE source_fingerprint <> ''
```

衝突は新規行を作らず、既存ジョブを「再送（取込済み）」として返す。

### 2. デコード DTO は fixture バッチと分ける

`LabInboundResultRow`（`old_pet_key` 等）に機器項目を載せない。BRT-95 は純関数:

```
bytes + 機器ヒント（任意） → LabDeviceFrame | invalid_payload
```

`LabDeviceFrame`: `source_type`, `source_fingerprint`, `measured_at`, `specimen_id_raw`, `device_hint`, `items[]`, `warnings[]`。

Postgres enum は BRT-95 では触らない。fresh `001_init.sql` の `CREATE TYPE lab_import_source_type` に `fuji_nx600` / `fuji_au10v` / `arkray_pu4010` を含む（2026-08-19 統合。旧 F9 の ADD VALUE 分割は incremental 時代の制約）。使う側は BRT-96 以降。

テストはサニタイズ合成バイトのみ。実 `.raw` は old_db に残し、AnimalEkarte へ複製しない。

### 3. device persist は `job_id` + `pet_id`。既存 Commit は触らない

`POST /api/v1/lab-imports`（fixture + `LabExamPersistInput`）に3種を allowlist しない。クライアントは `exam_type_id` も測定値も送らない。

新口（名前は実装時に OpenAPI へ）:

| 口 | 役割 |
| --- | --- |
| `POST /lab-device/frames` | 待機ページが payload（上限 8 KiB）を送る。サーバがデコードしジョブ+項目を書く。バイトはログしない |
| `PUT /lab-device/wait` | ペットを選ぶ。医院あたり **有効待機は1件**。別の子を選ぶと置き換える |
| `DELETE /lab-device/wait` | 待機解除（1操作） |
| `GET /lab-device/board` | 1画面用: 有効待機、未紐付け、直近の保存カード |
| `POST /lab-imports/:job_id/attach` | `{ pet_id }` だけ。サーバが保存済み項目 + マスタで persist |
| `POST /lab-imports/:job_id/detach` | ［取り消す］。下記 §5 |
| `GET /lab-device/unlinked` | 診察端末バナー用。件数は 0〜2 想定。page/q は作らない |

`source_type` が3種以外の attach/detach は 400。

Write owner は既存どおり `medicalrecord`。`internal/lab` は作らない。

### 4. 待機はサーバ保持。期限切れは未紐付けへ落ちる

`lab_device_waits`:

| 列 | 内容 |
| --- | --- |
| `clinic_id` | NOT NULL |
| `pet_id` | NOT NULL。clinic 内 pets |
| `staff_id` | 選んだスタッフ |
| `expires_at` | 期限 |
| `cleared_at` | 解除または期限切れ処理済み |

有効待機: `cleared_at IS NULL`。部分 unique `(clinic_id) WHERE cleared_at IS NULL`。v1 は検査用 Mac 1台。

フレーム到着時:

1. 有効待機を行ロックする
2. `expires_at <= now` なら `cleared_at` を立て、未紐付けとしてジョブを残す（`pet_id` NULL）
3. 有効ならその `pet_id` で即 persist（マスタ対応行）

同じ子の NX600→AU10V→尿は、解除・別の子・期限切れまで待機を持続する。この持続は §5 の Undo と期限切れなしでは成立させない。

TTL の数値は本 ADR で製品 KPI にしない。医院セットアップの1項目。シード初期値は実装時に1つ置く。数値チューニング UI は作らない。

### 5. ［取り消す］は detach。終端 `reverted` にはしない

既存 `POST /lab-imports/:id/revert` は fixture 補償用の **終端** `persisted → reverted` のまま残す。device の Undo に使うと指紋 unique が再 attach を殺す。

device ［取り消す］は **detach**:

- 流用: exam の retraction スナップショット、`usage_tracking_started` 必須、usage receipt / confirmed / 確定カルテの 409（`assertRevertSafe` 相当）
- 変える: 状態は `persisted → received`、`pet_id` を NULL。指紋はそのまま。同じジョブを未紐付け欄から付け直せる
- すぐ Undo する分には usage receipt はまだ無い想定。診察画面を開いたあと（`examination_detail` receipt）は 409。そのときは検査画面で直す。確認ダイアログは足さない

### 6. UI は1画面。未紐付け欄は一級

待機ページは操作画面ではなく、機器の前から読める掲示板。

状態:

1. 開く → 受信中（ペット未選択）+ 「受信中/切断」と最終受信（クライアント表示。サーバは知らない）
2. ペット選択 → 待機中（名前をページ最大）
3. 電文（待機中）→ 保存カード + ［取り消す］
4. 電文（未選択/期限切れ）→ 未紐付け欄
5. 未紐付けにペットを付ける → 保存カード
6. 取り消し → 未紐付けへ戻る
7. 待機解除は1操作
8. 未対応行は欄内チップ → マスタ該当行。対応後そのジョブだけ再 attach/再 persist

モードAは「未紐付け欄にペットが事前に載っている」最適化。Bを例外画面に分けない。

診察端末: その子の検査画面に「未紐付けの受信あり」。1クリックで `attach`（値は編集しない）。カルテから待機を遠隔起動しない。

確認ダイアログ禁止。

### 7. マスタと 1測定=1 exam

`lab_device_item_masters`: `(clinic_id, source_type, device_item_code)` unique。`exam_type_field_id` は空で投入可。`legacy_name_candidate` は列にしない。

未知コードはマスタへ自動追加しない。`needs_review`。

persist 時、マスタ対応行の `exam_type_id` が2種以上なら **保存拒否**（ジョブは `needs_review`）。1測定=1 `exams` 行。対応できた行だけ `exam_results`。未対応はジョブ項目に残す。

日常の送信経路にマスタ画面を出さない。

### 8. 公式リカバリ

届いていない → 機器の送信をもう一度押す。3台とも再押下可能を観測済み。指紋が二重取込を防ぐ。カードが出るか、または「再送（取込済み）」で分かる。

医院セットアップ: 口→機器プロファイルを1回。以後は許可済みポートを自動再オープン。日常のポート選択はゼロ。取り違えは文字化けとして失敗し `invalid_payload`（サイレントにしない）。スリープ無効・固定タブは運用メモ。メニューバー常駐は今は作らない。

ポート許可オブジェクトはブラウザ（Web Serial `getPorts`）にしか無い。サーバが持つのは論理スロット→`source_type`。電文が機器を自己申告できる場合はデコーダ側を正とし、プロファイルはシリアル設定用。

`lab_device_station_settings`（`clinic_id` PK）: `wait_ttl_seconds`、論理スロット JSON。`clinic_settings`（締め/CPM）には載せない。write owner は `medicalrecord`（lab 専用事実）。

### 9. 権限・隔離

- 受信・待機・board: `lab-import:create` または既存 lab-import 相当。clinic は JWT のみ
- attach/detach: `lab-import:edit`
- `pet_id` は request 由来 FK。同 clinic の生存ペット以外は 400
- 全クエリは `clinic_id` 必須。count も同様
- 生ペイロード・接続文字列・`specimen_id_raw` をログ/handoff/Linear に出さない

## Consequences

**する**

- BRT-95: `LabDeviceFrame` デコーダ + 合成バイト。DB なし
- BRT-96: マスタ。DDL は `001_init.sql` セクション13（2026-08-19 統合）
- BRT-97: 待機/board/frames/wait/UI。persist は §3 の口を呼ぶ
- BRT-98: `job_id`+`pet_id` の exam 書き込み。fixture Commit は3種を受けない。コード済み。Done は人間ゲート
- 医院公開は子 Issue 全部 Done のあと

**しない**

- 既存 `Commit` の allowlist
- 確認ダイアログ
- アップロード / デーモン / `drwan` / IDEXX / 7000V / ACK
- 検体ID照合
- 未紐付けの検索・ページング
- ADR の書き換えで方針変更（覆すなら新番号）

**既存への影響**

- fixture preview/commit/revert は維持
- `exams.job_id` を device persist でも使う
- detach は `reverted` を増やさない

## 実装順（コードを書いてよい条件）

BRT-95 のコードは **本 ADR を人間が BRT-100 Done にしたあと**。0 のまま 95 を始めない。
