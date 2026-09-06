# ADR-007: 城東検査機器 — 受信永続化・1画面・即 persist + Undo

**Status**: Accepted; local receive-agent choice partially superseded by [ADR-008](008-local-lab-device-agent.md)

**Implementation**: **code complete** for BRT-95〜98 paths on main (decoder · masters · board · commit allowlist). **Not** hospital release-ready by itself — real-device UAT, agent ops, and clinic rollout remain human gates (see BRT-94 / LAB_DEVICE_CONNECTIVITY).
**Date**: 2026-08-19 (status split 2026-08-30)
**Deciders**: PO（MinoruSoga）
**Relates to**: ADR-002（clinic_id）、ADR-006（`medicalrecord` write owner）、Linear [BRT-100](https://linear.app/baritechllc/issue/BRT-100) / [BRT-94](https://linear.app/baritechllc/issue/BRT-94)
**接続・機器正本**: [`docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md`](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md)
**Historical external evidence**: `old_db/docs/lab-go/go-impl/device-serial-adapter.md` and `old_db/docs/lab-go/go-impl/REVIEW-FABLE-2026-08-19-AE-LAB-UX.md` belonged to an external archive and are not paths in this repository. Current repository-accessible connectivity authority is `LAB_DEVICE_CONNECTIVITY.md` above.

## Context

現行 `lab_import_jobs` は行カウンタと状態だけを持つ。`Preview` はステートレス、`Commit` はクライアントが `LabExamPersistInput`（`exam_type_id` 等）を渡すワンショット。受信フレームをペット選択まで置くテーブルが無い。

**Original BRT-96 scope (historical):** 城東3台（NX600 / AU10V / PU-4010）は検査用 Mac の待機ページが有線シリアルを読む案だった。ファイルアップロード・常駐デーモン・`drwan` を製品経路にしない判断のうち、常駐受信を避ける部分は ADR-008 が supersede し、現在は user LaunchAgent を採用する。

Fable UX（YES-WITH-FIXES）: 日常は本日診療中カルテを1回選んで16項目手打ちを消す（ペット検索はしない。正本は LAB_DEVICE_CONNECTIVITY）。ただし前の子の待機が残ったまま次送信が届く誤紐付けを、確認ダイアログなしで塞ぐまで医院に出さない。

### 実践ゲート

| 項 | 内容 |
| --- | --- |
| 責任者 | MinoruSoga |
| 業務目的 | NX600 16項目の手打ち転記（40操作超+転記ミス）を消す |
| 削除する工程 | 手打ち、ファイル作成/選択、確認ダイアログ、待機と未紐付けの往復、日常のポート選択 |
| 残す操作 | 本日診療中カルテ選択1回。送信は機器側（従来どおり） |
| メトリクス | 追加操作は本日診療中カルテ選択1回。Undo / 待機解除 / 後付けは各1操作 |
| やらない | アップロード、取込値エディタ、遠隔待機起動、検体ID自動紐付け、未紐付け検索。常駐アプリを避ける判断は historical で、ADR-008 が user LaunchAgent 採用へ supersede |

## Decision

### 1. 1フレーム = 1測定 = 1 `lab_import_jobs` 行

新しい「受信ヘッダ」テーブルは作らない。既存ジョブを **1受信フレーム** に使う。

`lab_import_jobs` に足す列（device 行だけ埋める。fixture は NULL / 空のまま）:

| 列 | 内容 |
| --- | --- |
| `pet_id` | 未紐付けは NULL。clinic 内 pets への FK |
| `measured_at` | 電文日時。検査日の正 |
| `received_at` | デコード成功時刻。監査のみ。検査日にしない |
| `device_hint` | `NX600` / `AU10V` / `PU-4010` / `VetLab` |
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

Original BRT-96 scope では3値だった。現行 `001_init.sql` の `CREATE TYPE lab_import_source_type` は `fuji_nx600` / `fuji_au10v` / `arkray_pu4010` / `idexx_vetlab` を含み、current decoder/persist path も IDEXX を実装する。

テストはサニタイズ合成バイトのみ。実 `.raw` は上記 external archive にだけ残し、AnimalEkarte へ複製しない。

### 3. device persist は `job_id` + `pet_id`。既存 Commit は触らない

`POST /api/v1/lab-imports`（fixture + `LabExamPersistInput`）に機器 `source_type` を allowlist しない。クライアントは `exam_type_id` も測定値も送らない。

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

device attach/detach は現行4種（`fuji_nx600`, `fuji_au10v`, `arkray_pu4010`, `idexx_vetlab`）だけを受け、それ以外は 400。

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

### 7. マスタと 1測定=mapped exam_type 数分の exam

`lab_device_item_masters`: `(clinic_id, source_type, device_item_code)` unique。`exam_type_field_id` は空で投入可。`legacy_name_candidate` は列にしない。

未知コードはマスタへ自動追加しない。`needs_review`。

~~persist 時、マスタ対応行の `exam_type_id` が2種以上なら **保存拒否**（ジョブは `needs_review`）。~~

**2026-08-20 追記（T001 / VetLab 複数 exam_type 対応）**: IDEXX VetLab は複数機器の結果を1電文で送る端末（送信口）であり、1受信フレームに複数の `exam_type` にまたがる項目が含まれる。

- `device_item_code` → `exam_type_field` → `exam_type` で検査を決める（スタッフが機器を選ぶのではない）。
- マップ済み項目の `exam_type_id` が N 種なら `exams` を N 行作成する（1種なら従来どおり1行）。これは1フレームの種別分割であり、機器グルーピング schema は作らない。
- 保存拒否条件「`exam_type_id` 2種以上」は廃止。`AssertSingleExamType` の呼び出しを削除。
- detach/undo は `job_id` 由来の exams をすべて取り消す（既存 `DetachDeviceJob` がすでに全件取り消しに対応済み）。
- `lab_devices.exam_type_id` の複数化 schema は作らない。`LAB_DEVICE_CONNECTIVITY.md` は触らない。

対応できた行だけ `exam_results`。未対応はジョブ項目に残す。

日常の送信経路にマスタ画面を出さない。

### 8. 公式リカバリ

届いていない → 機器の送信をもう一度押す。3台とも再押下可能を観測済み。指紋が二重取込を防ぐ。カードが出るか、または「再送（取込済み）」で分かる。

医院セットアップ: 口→機器プロファイルを1回。以後は許可済みポートを自動再オープン。日常のポート選択はゼロ。取り違えは文字化けとして失敗し `invalid_payload`（サイレントにしない）。スリープ無効・固定タブは運用メモ。メニューバー常駐を作らないという当初判断は historical。ADR-008 が user LaunchAgent の `lab-device-agent` 採用へ部分的に supersede した。

ポート許可オブジェクトはブラウザ（Web Serial `getPorts`）にしか無い。サーバが持つのは論理スロット→`source_type`。電文が機器を自己申告できる場合はデコーダ側を正とし、プロファイルはシリアル設定用。

`lab_device_station_settings`（`clinic_id` PK）: `wait_ttl_seconds`、論理スロット JSON。`clinic_settings`（締め/CPM）には載せない。write owner は `medicalrecord`（lab 専用事実）。

### 9. 権限・隔離

以下は採用時案。現行の認証・権限契約は[現行 HTTP 境界の補足](#current-http-boundary)を参照する。

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
- BRT-98: `job_id`+`pet_id` の exam 書き込み。fixture Commit は機器 `source_type` を受けない。コード済み。Done は人間ゲート
- 医院公開は子 Issue 全部 Done のあと

**しない**

- 既存 `Commit` の allowlist
- 確認ダイアログ
- アップロード / `drwan` / 7000V。デーモンを作らない判断は ADR-008 が supersede。IDEXX decoder/PIMS reply code は実装済みだが、医院での物理接続・PIMS応答・UAT・rollout は別 gate
- 検体ID照合
- 未紐付けの検索・ページング
- ADR の書き換えで方針変更（覆すなら新番号）

**既存への影響**

- fixture preview/commit/revert は維持
- `exams.job_id` を device persist でも使う
- detach は `reverted` を増やさない

## Historical implementation gate（BRT-95 着手前に充足済み）

BRT-95 着手前は BRT-100 の人間決定を gate とした。現行 header のとおり BRT-95〜98 の code path は完了しており、この節は active prohibition ではない。物理機器 UAT、agent 運用、医院 rollout は引き続き人間 gate である。

<a id="current-http-boundary"></a>

## 現行 HTTP 境界の補足（2026-09-06）

上記の採用時案のうち、§3 の API 名と §9 の認証・権限は現在の `backend/internal/medicalrecord/routes_lab.go` / `backend/docs/api.yaml` を優先する。共通 prefix は `/api/v1`。

| Endpoint 群 | 現行 `lab-import` action |
| --- | --- |
| `POST /lab-device/frames`、`PUT/DELETE /lab-device/wait`、`GET /lab-device/board`、`GET /lab-device/agent-consumer` | `create` |
| `GET /lab-device/unlinked`、`GET /lab-device/station`、device / item-master 一覧 | `view` |
| station 更新、device / item-master 更新、item-master ensure、job attach/detach/revert | `edit` |
| `POST /lab-devices` | `create` |

§9 の「clinic は JWT のみ」は client body を authority にしないという採用時の要約であり、JWT snapshot を最終 authority とする契約ではない。現行 staff route は認証 middleware が `current_access_service` で再解決した clinic 集合と `X-Clinic-ID` を検証し、その trusted clinic scope を handler へ渡す（[auth.md](../auth.md)）。

§8 のブラウザ `getPorts` は Web Serial 経路の説明である。ADR-008 の local agent 経路は Mac の許可済みポートを agent が監視し、Frontend が consumer lease/token でキューを読み出す。機器対応コード、医院での配線確認、UAT、公開判断は別々に確認する。
