# STG 実データ運用テスト — 城東 + 八王子

更新日: 2026-08-26（Lane 2 コードは `4703cf3e9` で main 入り。M-3 の canonical runbook 追記は未適用）  
本書は STG 並行運用から本番移行日までの実行計画の正本。実行 SoT は Linear（入口 [`todo.md`](todo.md)）。old_db 側の producer 作業は sibling `old_db/todo.md`。手順の詳細（F6 変数、backup、失敗時）は末尾の runbook を参照し、ここに複製しない。

---

## 対応状況（2026-08-26）

エージェント実装（Lane 2）は完了。残作業は USER / old_db。**2026-08-26:** USER ローカルログイン確認済み。城東の 2 回目 `make reset` import（preflight/apply/verify）**PASS**（`jouto-intake-20260822-01`）。Lane 3 STG 投入は引き続き **USER のみ**（H3 は投入するまで付けない）。

| レーン | 状態 | メモ |
|--------|------|------|
| Lane 0 入力 | 進行中 | H0-1 済み。H0-3a/H0-4 城東済み。H0-5: ローカル attach **PASS**（staff_count=414 / digest `a6ea3cdd6d2b36bea961dfd00faadbfc88fb4e0b59b194d02baf5b6bb1ed71ff`）。secrets は gitignored。`set_active=false` のまま。**2026-08-26 USER ローカルログイン確認済み**（画面で is_active 対応）。STG attach 未。未: H0-2 八王子 |
| Lane 2 コード | **完了** | `4703cf3e9`。SKELETON / IMPORT / STAFF / Make の DB_HOST・SSL・sentinel 転送 |
| Lane 1 ローカル証明 | 城東 rehearsal 済み | H1-1a write-0 PASS。H1-1b: 1 回目（2026-08-25）と **2 回目（2026-08-26 ~10:18–10:27 JST）** とも USER `make reset` → `csv-import` apply **PASS** + verify **PASS**（`jouto-intake-20260822-01`）。ローカル rehearsal 経路であり STG UAT `stg-uat-csv-import` apply ではない。八王子未 |
| Lane 3 STG 投入 | 未着手 | **USER のみ。** エージェントは pscale / `make stg-uat-*` apply を実行しない。城東先行。ローカル証明済みでも本レーンの H3 は付けない |
| Lane 4 並行運用 | 未着手 | 城東投入後。STG 入力は本番に移さない |
| 任意・後追い | 未着手 | H3-10 SHIFT、H1-3/H1-4、M-3 runbook 1 行（明示時のみ）、AE-SEED-RETIRE-DEMO |

claim 枝（削除は USER。merge / abandon のあと `git branch -D`）:

`claim/AE-STG-UAT-SKELETON` · `IMPORT` · `STAFF` · `OPS-SHEET` · `MAKE-REMOTE` · `H0-JOU-CHECK` · `H1-JOU-PREFLIGHT` · `H1-LAND` · `JOU-PAY-SNAP` · `JOU-PAY-GRAPH` · `H0-5-ROSTER` · `STAFF-ATTACH-LOCAL` · `LANE3-PREP`

---

## 0. 目的

**STG に八王子・城東の旧システム実データを投入し、各医院のスタッフが現行業務と並行して新システムでも同じ登録を行い、ブロッキングな不具合が消えるまで検証する。その後に本番移行日を決め、その日に旧データの出力と新システムへの移行を行う。**

| 項目 | 内容 |
|------|------|
| 何のため | 本番切替の前に、実患者・実カルテ・実会計の上で現場が新システムを回せることを確認する。デモデータでは足りない |
| 誰が使う | 八王子・城東の **各医院スタッフ**（受付・獣医・会計など、現行業務を担っている人）。開発者の代行操作ではない |
| どこで | 共有 STG（`stg.noah-karte.com` / PlanetScale `animalekarte-stg`）。院内から URL で触る |
| 並行期間の正本 | **旧システム。** 現行業務はこれまでどおり旧システムで続ける |
| STG の役割 | 同じ内容を新システムにも登録する検証環境。本番の下書きではない |
| いつ終わるか | 必須 4 業務（検索・受付・カルテ・会計）を対象医院で連続 5 営業日 STG でも再現でき、現場が切替に進んでよいと判断したとき。上限 8 週 |
| その後 | 本番移行日を決める。その日に旧データを出力し、新システムの本番へ移行する（§0.2） |

この並行登録は製品の二重入力機能ではない。移行検証のための **期限付き工程** であり、本番稼働後は残さない。

### 0.1 二段階

```
第1段階  STG 並行運用（本書の主対象）
  旧システムの履歴を STG へ投入
  各医院スタッフが旧システム + STG の両方に同じ業務を登録
  不具合は Linear で直し、ブロッキングが消えるまで続ける

第2段階  本番移行日（本書の後続。第1段階と混ぜない）
  移行日を決める
  その日の旧データを出力する（STG 上の並行登録は使わない）
  新システムの本番へ F6 正式 cutover（PASS / TRUSTED_CANDIDATE）
  以後の正本は新システム。旧システムと STG 並行登録は捨てる
```

### 0.2 STG のデータは本番にしない

並行期間中にスタッフが STG へ入れた予約・カルテ・会計は **検証用** である。本番移行日の入力は、その時点の旧システム出力である。

- STG DB を本番へコピーしない
- STG の並行登録を本番の差分として足さない
- 本番 cutover は第1段階の rehearsal 合格を「本番 F6 の証跡」に使わない。本番は別 bundle（`PASS` / `TRUSTED_CANDIDATE`）

### 0.3 第1段階で検証する業務 / しない業務

必須（各医院スタッフが旧システムと同じように登録する）:

- 飼主・ペットの検索と参照
- 受付、カルテ記載、会計（preflight が通る範囲）
- ログインは移行 `staffs.id` に紐づく本人アカウント

任意（やるなら別途投入が要る。21 表に無い）:

- 今日の枠での新規予約（未来日シフト）
- LINE / LIFF
- 院内検査機器
- 検査画像（R2）

やらない:

- デモ臨床での代用
- 本番環境への先行投入
- STG を正本にした現行業務（現行の正本は旧システムのまま）

### 0.4 第1段階の完了 / 第2段階の開始

第1段階が終わったと言える条件:

- 第1段階開始時に対象となっている医院のスタッフが自分の医院でログインできる（開始時は城東のみでも可）
- 必須 4 業務（検索・受付・カルテ・会計）を、対象医院で連続 5 営業日 STG でも再現できている
- 現場がブロッキングと判断した不具合が解消している
- 切替印は **両院が STG に入ったあと** に付ける。上限 8 週
- 本番 F6 の TRUSTED 契約は緩めていない

第2段階は、その印のあとで移行日を切る別作業である。

---

## 1. 決めたこと

| 項目 | 決定 |
|------|------|
| 対象医院 | **両方。** 八王子 `clinic_id=1`（ordinal 1、ID band 0–10M）、城東 `clinic_id=2`（ordinal 2、ID band 10M–20M） |
| 同一 STG DB | 載せる。band が医院で分かれている |
| PHI | 共有 STG に飼主・ペット・カルテを置いてよい。残置期限と破棄担当を作業票に書く |
| STG DB | デモと混ぜない。承認済みで PlanetScale を再作成してよい |
| `003_demo` / `004_staging` | **STG に使わない。載せない。** `APP_ENV=staging` の migrate も読まない |
| `002_master` | 残す（種・会社・LSTEP 接頭辞）。21 表の `FALLBACK_ANIMAL_SPECIES_ID` の供給源 |
| 業務データの正本 | old_db の医院別 21 CSV + manifest。F6 `csv-import` |
| `_old_db_handoff/` | ローカル隔離だけ。`cmd/migrate` は読まない。STG 投入の運搬路にしない |
| ログイン | 各医院の **現行スタッフほぼ全員** を、移行 `staffs.csv` の `id` に account 後付けする。新規 staff の複製はしない（§7）。CSV 上の履歴・退職行は対象外。email は送信しないデモアドレス（`stg-staff-{id}@example.test`） |
| 並行期間の正本 | 旧システム。STG は検証用の二重登録先 |
| 本番 cutover | 第2段階。従来どおり `PASS` / `TRUSTED_CANDIDATE` のみ。緩和しない。STG データを昇格しない |
| STG 並行運用 | 第1段階。本番 cutover とは別レーン（§2） |
| 旧会計の異常値 | 補正せずそのまま載せる（`todo.md` AE-MIG-NEG-*） |

`003_demo` はデモ臨床と既知パスワードの admin を含む。STG に載せると実患者と混ざる、unique 衝突で import が落ちる、テスト用パスワードが共有環境に残る。投入対象ではない。`004_staging` は 003 の trimming 明細の残りで、STG 用ではない。

`_old_db_handoff/` は `.git/info/exclude` 済みであること（`git check-ignore -q --no-index backend/migrations/seeds/_old_db_handoff/`）。CSV / manifest / 行値 / 初期パスワードを git・Issue・チャット・ログに出さない。

---

## 2. 本番 cutover と STG 運用テストを分ける

| レーン | 入力 | 目的 |
|--------|------|------|
| **第1段階 STG 並行運用** | 医院 identity 付き 21 CSV。現状は城東が `REHEARSAL_ONLY`、八王子は現行 checkout に 21 CSV がない | 各医院スタッフが旧システムと STG に同じ業務を登録し、ブロッキング不具合を潰す |
| **第2段階 本番 cutover** | 移行日時点の旧データ出力。`status=PASS` かつ `handoffEligibility=TRUSTED_CANDIDATE` | 医院カットオーバー。既存 F6 を緩めない。STG の並行登録は使わない |

TRUSTED 待ちだけにすると、城東は Z5（税、八王子未証明）と G2 で BLOCKED、八王子は KNJO ページ破損で 21 CSV 自体が無い。運用テスト開始が producer 完了まで止まる。

そのため **STG 運用テストは REHEARSAL_ONLY を、USER の明示承認があるときだけ** 共有 STG に載せる。これは本番 F6 の `--allow-local-rehearsal` を staging で有効化することではない。ローカル reset 用フラグを共有環境へ流用しない。STG 専用の承認付き手順を別途切る。

承認する人が引き受ける事実:

- 現場は移行完了まで **二重登録** する（旧システムが正本、STG は検証）。製品機能として残さない
- 会計は機械分解しきれない行が残る（城東は `needs_review` が残っている）
- 負額請求はそのまま載る（補正しない）
- 八王子は KNJO 不完全な rehearsal になり得る
- 証跡は「STG 並行運用」であり、本番 cutover の合格証跡には使わない
- PHI が共有 STG に載る。並行期間中にスタッフが新規登録した PHI も STG に残る
- 並行期間中の STG 書き込みは本番へ持っていかない。移行日に旧システムから出し直す

承認が取れないなら STG には載せない。ローカル / A4 disposable で画面確認に留める。

`pscale connect` で `localhost` に見せかけて `make csv-import` を通すのは使わない（Compose は `DB_HOST=db` 固定、SSL 契約、`--confirm-target-host db` と食い違う）。

---

## 3. いまの在庫（2026-08-25）

正本の状態は old_db の医院別ドキュメント。ここは AnimalEkarte 側の要約（件数のみ。行値は書かない）。

| 医院 | producer 21 CSV | 資格 | AnimalEkarte 側 |
|------|-----------------|------|-----------------|
| 城東 `jouto` | `jouto-intake-20260822-01` あり。飼主約 4.3 万 / ペット約 7.0 万 / カルテ約 114 万 / 会計明細約 498 万 / 検査結果約 797 万 / 支払約 88 万 / 分割約 89 万（header-only **ではない**）。スタッフ 414 行は全員 `is_active=false`。CSV 合計およそ 1.8GB | `REHEARSAL_ONLY` / `UNVERIFIED`。TRUSTED は JOU-CSV-1 で BLOCKED | `_old_db_handoff/jouto/` にローカル rehearsal あり。STG 未投入 |
| 八王子 `hachioji` | 現行 checkout に正式 21 CSV なし。直下 7 CSV は医院 identity 無しで **使わない** | KNJO 破損。inventory `INCOMPLETE`。旧 partial rehearsal は payments=0 | `_old_db_handoff/hachioji/` は空ディレクトリ |

[CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) の「KNJO 支払が header-only」は、**城東のこの run には当たらない。** 行はある。足りないのは producer の正式認定である。不足は「行が無い」と「認定が無い」を混ぜない。

STG の migrate 後に自動で入るのは `002_master` だけ。**clinics / accounts / 支払方法 / 検査種別 / 予約種別 / 権限グループは無い。** 21 表だけを流し込んでもログインできないし、F6 の 6 seed binding も満たせない。

handoff の `staffs.csv` 列は `id, clinic_id, name, license_number, is_active, reservation_visible` だけ。email / パスワード / 権限 / `staff_type` は無い。氏名は一意でない。

21 表に無いため、運用テスト範囲に応じて別途用意する:

- シフト / 予約枠（カレンダー・LIFF 予約を試す場合）
- LINE / LIFF 実アカウント（STG は mock 禁止）
- 検査機器マスタ（院内機器テストをする場合）
- 検査画像（R2）。CSV だけでは来ない

カルテ閲覧・患者検索・会計参照が主なら、骨格 + 21 表 + ログインで足りる。

---

## 4. やってはいけないこと

- `_old_db_handoff` を git に入れる、Docker イメージに焼く、`003_demo` にコピーする
- `003_demo` / `004_staging` を STG の migrate に乗せる、または `\copy` する
- `APP_ENV=staging` の migrate に handoff を読ませる
- export 直下の旧 7 CSV を八王子または城東の成果物として使う
- `REHEARSAL_ONLY` を `TRUSTED_CANDIDATE` に書き換えて F6 を通す
- `--allow-local-rehearsal` を STG / 本番で付ける
- `make csv-import` の `DB_HOST` を PlanetScale ホスト名にする
- `pscale` + `\copy` で 21 表を手投入する（band / SHA / 単一 transaction / 非PHI report を飛ばす）
- 氏名から email を推測してスタッフ 414 人分のアカウントを自動発行する
- `staff-provision` で移行済み staff と **同じ人の新規行** を作る（カルテ `doctor_id` がずれる）
- band 占有後に手書き DELETE して再投入する。やり直しは backup 復元か承認済み DB 再構築
- 片方の医院 import 失敗を、成功した医院の band を消して取り繕う
- 現行 Cloudflare `backend-deploy.yml` に無い `db_reset` を捏造する
- 行値・パスワードを本書・Issue・PR・ログに書く

---

## 5. 確定した判断

2026-08-25 USER 回答で確定。Q2=B と Q5=A は衝突したため、**城東で第1段階を開始**し、八王子 21 表は後追い必須（HAC-CSV-1。identity 無しの 7 CSV は使わない）と読む。

| ID | 判断 | 決定 | 印 |
|---|---|---|---|
| D1 | 城東 rehearsal を STG に載せるか | **載せる。** STG 専用フラグを新設。本番 F6 は緩めない。`--allow-local-rehearsal` は流用しない | 確定 |
| D2 | 第1段階の開始 | **B. 城東単独で開始。** 現場へ「城東のみ。八王子はまだ旧システムのまま」と明示。八王子 apply は H3-7 maintenance window | 確定 |
| D3 | ログイン対象 | **各医院の現行スタッフほぼ全員**を移行 `staffs.id` に後付け。CSV の履歴・退職行は対象外。414 行の一括発行はしない。誰が現行かは名簿で渡す | 確定 |
| D4 | email 方針 | **デモ専用。SMTP には使わない。** 形式 `stg-staff-{id}@example.test`。実メールは使わない | 確定 |
| D5 | 権限グループ | 現行役割ごとに **明示 ID**。role 名からの推論はしない。具体 ID は名簿で渡す | 方針確定・ID は名簿待ち |
| D6 | 必須業務 | 検索・受付・カルテ・会計（preflight が通る範囲）。未来日予約・LIFF・機器・画像は必須にしない | 確定 |
| D7 | 並行期間と破棄 | **開始:** 城東投入後。八王子は後追い。**終了:** 必須 4 業務を対象医院で連続 5 営業日 STG 再現、かつ現場が切替可と判断。上限 8 週。切替印は両院投入後。破棄日は第2段階の前に書く | 確定 |

---

## 6. 作業レーン（この順。逆行しない）

工程を足す前に、存在すべきでない経路（migrate に handoff を載せる、demo CSV に実データを混ぜる）は作らない。
着手前に `git branch --list 'claim/<TASK-ID>'`。空なら `git branch claim/<TASK-ID>`。claim の削除は USER。

### Lane 0 — 入力を揃える（USER / old_db）

- [x] **H0-1** D1–D4, D6–D7 確定（2026-08-25）。D5 の権限 ID と現行スタッフ名簿は repo 外で後から渡す
- [ ] **H0-2** 八王子の医院 identity 付き 21 表 CSV + manifest を old_db から出す（HAC-CSV-1。rehearsal 可。旧 7 CSV は使わない）。各 CSV が 512MiB 未満であることを確認する
- [x] **H0-3a** 城東: `CLINIC_CODE=jouto MIGRATION_RUN_ID=jouto-intake-20260822-01 make old-db-handoff-check` PASS（配置済み。`old-db-handoff-stage` は再実行しない）
- [ ] **H0-3b** 八王子: 同じ check（H0-2 待ち。現行 `hachioji/` に manifest なし）
- [x] **H0-4** 城東 manifest SHA-256（`backend/migrations/seeds/_old_db_handoff/jouto/manifest.json`）: `7bbda50f06f7d0acac6711d1a73b78ca68b835ee9be5df6cd04f3e6a5094a405`（八王子は H0-2 後に別途転記。2026-08-25 Class A: payments snapshot、completed+nonzero 欠 graph を pending 再分類、同一 medical_record_id の余剰 billing 192 行のリンク解除。billings SHA `c04d05a014d7ae58cc9103e347ffe42100e7c84bede9a72c4bee0a1f3ca780be`）
- [x] **H0-5** ログイン名簿（repo 外・mode 0600）: roster `sensitive-local/stg-uat-staff-roster.json` + secrets `sensitive-local/stg-uat-staff-secrets.json`（どちらも gitignored / 0600）。**2026-08-26 ローカル attach PASS:** preflight+apply `staff_count=414` digest `a6ea3cdd6d2b36bea961dfd00faadbfc88fb4e0b59b194d02baf5b6bb1ed71ff`（`DB_HOST=db` / remote sentinel 未設定）。email は `stg-staff-{id}@example.test`。`set_active=false` のため is_active は CSV のまま。**2026-08-26 USER ローカルログイン確認済み**（2 回目 `make reset` 後に staff-attach は消えたが、画面で is_active を扱いログインできた）。STG attach は未実行。SMTP には使わない

配置先:

```text
backend/migrations/seeds/_old_db_handoff/hachioji/
backend/migrations/seeds/_old_db_handoff/jouto/
```

同一医院で電話 unique 用の sanitize が要る場合は `<clinic>-local/` を優先する（ローカル `make reset` の既存契約）。

### Lane 1 — ローカル証明（Lane 2 の後・STG より先）

現行 `make reset` はローカル rehearsal 経路で、STG リモートゲート（AE-STG-UAT-IMPORT）の証明には使えない。Lane 2 が終わってから、そのゲートをローカルまたは disposable で通す。

- [x] **H1-1a** 城東 preflight（ローカル `db` / write-0）: `make stg-uat-csv-import-preflight`（`CLINIC_CODE=jouto` / ordinal 2 / clinic id 2 / SHA `42cb5f6755d2e4539253365d8975fc74fe633a44be6b784360cafc001bb71ef0` / `STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND`）。**結果 (2026-08-25):** PAY-SNAP 後 fail-closed `completed billing is missing its payment graph`。診断 counts-only: `completed_nonzero_without_payment=200582` / `completed_zero_without_payment=54603` / payments rows 883361 不変。Class A で completed+nonzero 欠 graph を `pending` + `completed_at=""` へ再分類（支払行は捏造せず。zero-without は importer 許容のまま）。rewrite 後 artifact 掃除（0600 / bak 除去）のうえ preflight **PASS**（`CSV STG UAT cutover preflight PASS` / tables=21 / apply 未実行）。seed IDs clinic2: species=1 exam=11009 trimming=59 cash=5 credit=6
- [x] **H1-1b** 城東 apply → verify（ローカル rehearsal）: USER `make reset`（`--allow-local-rehearsal` / `cmd/csv-import`）。**結果 (2026-08-25 23:36–23:37):** `CSV cutover apply PASS` clinic_code=jouto run_id=`jouto-intake-20260822-01` → `CSV cutover verification PASS` → `imported jouto/jouto-intake-20260822-01` → reset complete。**2 回目 (2026-08-26 ~10:18–10:27 JST):** 同じ run_id で preflight/apply/verify 再 **PASS**（reset 後の再 import。先行 staff-attach は wipe）。STG UAT `make stg-uat-csv-import` apply は未実行（このローカル DB は band 占有済み）。durable: old_db が 1 カルテ 1 billing リンク、completed-without-payment を出さないこと
- [ ] **H1-2** 八王子: 同じ。bundle が来てから
- [ ] **H1-3** 画面確認が要るなら [A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md)。通常 `csv-import-*` は使わない
- [ ] **H1-4** 失敗側は [F8_G4_FAILURE_REHEARSAL.md](docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md)（本番 CSV は渡さない）
- [ ] **H1-5** ここで落ちる bundle を STG に送らない

### Lane 2 — AnimalEkarte 実装

**コード完了**（main `4703cf3e9`）。残るのは名簿と USER の投入（Lane 1 → Lane 3）。PHI をテストフィクスチャに入れない。ローカル go / npm は使わない。検証は Docker のスコープ限定。`make db` / STG 破壊操作は USER 手動。

#### AE-STG-UAT-SKELETON — STG 用医院骨格

STG / 本番 migrate が読んでよい最小データ。臨床行（owners / pets / カルテ / 会計 / 予約）は入れない。既知デモパスワードの system admin も入れない。

**21 表（`staffs` を含む cutover 対象）へは行を作らない。** 八王子 band は 0–10M で、21 表の行が空であること。骨格の staff 行が id=1 になると `CUTOVER_REF_BAND_OCCUPIED` で八王子 preflight が落ちる。ops bootstrap は `accounts` のみ（staffs 行なし）。21 表へ行が必要な場合は、先に該当 sequence を `applicationIDFloor`（1,000,000,000）以上へ `setval` してから作る。

最低限（21 表の外、または sequence 前進後）:

- `clinics` id=1 八王子、id=2 城東（`002_master` の company 1 配下。clinics は 21 表に含まれない）
- 各 clinic の `payment_methods`（`cash` / `credit_card`。ensure-if-missing。clinic INSERT trigger が seed。F6 必須。21 表外）
- 各 clinic の `exam_types` name=`検査`（21 表外）
- 各 clinic の `reservation_types` category=`trimming`（21 表外）
- 権限グループと rule（現行スタッフを載せる受け皿。21 表外）
- `clinic_settings`（締め時刻など、運用テストで触る画面があるなら。21 表外）

**実装済み（`4703cf3e9`）:** `backend/cmd/stg-uat-skeleton` + `make stg-uat-skeleton`。投入は USER。エージェントは実行しない。

- 既定はローカル `DB_HOST=db` / `DB_SSL_MODE=disable`。Compose の `up` は引き続き `DB_HOST:db` 固定で、`docker compose run -e` だけが上書きする
- Make は `-e STG_UAT_SKELETON_ALLOW_REMOTE` と `-e DB_HOST` / `DB_PORT` / `DB_SSL_MODE` を転送する。リモート STG は USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE`（`require`|`verify-ca`|`verify-full`）と `STG_UAT_SKELETON_ALLOW_REMOTE=YES_I_UNDERSTAND` をセットしてから `make stg-uat-skeleton`。エージェントは実行しない
- `pscale connect` で `localhost` に見せかけて通す経路は禁止（§2）

ローカル `import-old-db-handoffs-on-reset.sh` と同等のグラフを、**STG では自動 reset に載せない。**

ログイン用 account は骨格にデモユーザーを固めない。ops 用の最初の 1 account だけがブートストラップに必要なら、secret-managed で **accounts のみ**発行し Git に書かない。対応する `staffs` 行は作らない（後付けは移行 CSV の id に対して行う）。

#### AE-STG-UAT-IMPORT — STG 運用テスト用 import ゲート

本番 F6 の `cmd/csv-import` / `buildTargetPoolConfig` ローカルガードは緩めない。STG 経路は別エントリ。

**実装済み（`4703cf3e9`）:** `backend/cmd/csv-import-stg-uat` + `make stg-uat-csv-import-preflight` / `make stg-uat-csv-import` / `make stg-uat-csv-import-verify`。投入は USER。エージェントは実行しない。

- 本番 F6 の TRUSTED 契約は変えない
- Make レシピは `APP_ENV=staging` と `STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND` を強制。report lane は `stg-uat-rehearsal`
- **STG では `--allow-local-rehearsal` 禁止**（ローカル `make reset` 経路を流用しない）
- 順序は preflight（write 0）→ apply（単一 transaction）→ verify（read-only）。病院は 1 clinic ずつ。band が空でなければ fail-closed。置換しない
- report は件数と 6 seed ID のみ。`sensitive-local/` へ 0600 / no-clobber
- 21 CSV は repo 外を read-only mount。git / CI / イメージに焼かない
- リモート SSL は `require` / `verify-ca` / `verify-full`
- Make は `DB_HOST` / `DB_PORT` / `DB_SSL_MODE` を転送し、apply の `--confirm-target-host` は `STG_UAT_CSV_IMPORT_CONFIRM_HOST`（既定 `db`）経由。`DB_HOST` への自動 alias はしない。ローカルは既定のまま。リモート STG は USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE` と一致する `STG_UAT_CSV_IMPORT_CONFIRM_HOST`、および `STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND` をセットしてから `make stg-uat-csv-import-*`。エージェントは実行しない
- Compose の通常 `up` は `DB_HOST:db` のまま。`pscale connect` → `localhost` 偽装は禁止（§2）
- 実行場所は STG から見える one-shot（対象環境コンテナ、または TTL `pscale role` 直結）。Hyperdrive 不可（advisory locks）

#### AE-STG-UAT-STAFF — 移行 staff への account 後付け

`staff-provision` は新規 staff を作るため、カルテの `doctor_id` とログイン中 ID がずれる。移行済み id には `CreateStaff` を使わない。CSV に無い現行スタッフだけ、新規発行の補助として `staff-provision` を使ってよい。

**実装済み（`4703cf3e9`）:** `backend/cmd/stg-uat-staff-attach` + `make stg-uat-staff-attach-preflight` / `make stg-uat-staff-attach`。**2026-08-26 ローカル apply PASS**（414）。STG リモート attach は USER / Lane 3。

- 入力キーは移行済み `staffs.id`
- 既存 staff 行を複製しない。`accounts` を作り `staffs.account_id` を張る
- 対象は名簿に書いた id だけ
- 後付け時に並行登録する人だけ `is_active=true` にしてよい（CSV は全員 false）
- email は `stg-staff-{id}@example.test`（送信しない）。パスワードは repo 外 0600。stdout は digest / count / staff_id のみ
- 権限グループは明示 ID。無い group は fail-closed
- Make は `STG_UAT_STAFF_ATTACH_ROSTER` / `STG_UAT_STAFF_ATTACH_SECRETS` を bind-mount し、`-e STG_UAT_STAFF_ATTACH_ALLOW_REMOTE` と `-e DB_HOST` / `DB_PORT` / `DB_SSL_MODE` を渡す。名簿と secrets は repo 外 0600
- ローカルは `DB_HOST=db` / SSL `disable` 既定。リモート STG は USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE`（`require`|`verify-ca`|`verify-full`）と `STG_UAT_STAFF_ATTACH_ALLOW_REMOTE=YES_I_UNDERSTAND` をセットしてから `make stg-uat-staff-attach*`。エージェントは実行しない。ローカル専用ガードを本番 F6 と共用して緩めない

名簿の最小列（repo 外）: `staff_id`、`clinic_id`（両院なら `clinic_ids`）、`email`、`secret_ref`、`permission_group_ids`、`set_active`。

#### AE-STG-UAT-JOU / AE-STG-UAT-HAC — 投入作業（USER。コードなし）

骨格の 6 seed を run sheet に固定する。manifest SHA-256 は producer report から別経路。ディレクトリ自己申告値を使わない。

- 城東: `jouto-intake-20260822-01`（または後継）。clinic 2 / ordinal 2 / band 10M–20M
- 八王子: 21 CSV が producer に存在するまで着手しない。clinic 1 / ordinal 1 / band 0–10M

#### AE-STG-UAT-SHIFT — 予約を試す場合のみ

未来日の `shift_entries`（またはテンプレート展開）。003 の陳腐化した絶対日付をコピーしない。21 表の過去予約だけでは「今日の枠」は埋まらない。

#### AE-SEED-RETIRE-DEMO — 003 / 004 の退役（STG の後）

STG 運用テストのブロッカーではない。ローカル E2E / lint / `make reset` が 003 固定のあいだは repo に残す。STG が実データで回り、ローカルも handoff または薄い骨格で足りたら削除する。

削除時に直すもの: `BundleOrder`、`local-db-reset-contract`、`demo_account_label_drift`、`seed_future_horizon`、E2E の固定 pet ID、UAT シナリオの「環境: 003_demo」。

### Lane 3 — STG 投入（USER。エージェントは DB を触らない）

[STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊境界に従う。各 CSV は 512MiB 上限。城東の大きい表は約 370MB（exam_results / billing_items）。八王子は H0-2 で上限確認する。

- [ ] **H3-1** 他エンジニアが STG を使っていないことを確認。並行運用の開始を医院へ伝える
- [ ] **H3-2** 検証済み full backup。復元担当を作業票に書く。現行データの破棄は USER 承認
- [ ] **H3-3** TTL 付き `pscale role` で直結。必要なら `DROP SCHEMA public CASCADE` と `CREATE SCHEMA public` を **対で**実行
- [ ] **H3-4** `gh workflow run backend-deploy.yml --ref staging`。migrate 成功。この時点は `002_master` のみ
- [ ] **H3-5** `make stg-uat-skeleton`（clinic 1=八王子 / 2=城東、6 seed、権限グループ）。**21 表に行を作っていないこと。** ID を作業票へ転記（PHI なし）。Make は `STG_UAT_SKELETON_ALLOW_REMOTE` と `DB_*` を転送する。ローカル既定は `DB_HOST=db` / SSL `disable`。リモートは USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE`（`require`|`verify-ca`|`verify-full`）と `STG_UAT_SKELETON_ALLOW_REMOTE=YES_I_UNDERSTAND` をセットしてから `make`。エージェントは実行しない
- [ ] **H3-6** 各院の 10M ID band が 21 表すべて空であること（staffs 含む）。骨格の bootstrap が band を占有していないこと
- [ ] **H3-7** 医院ごとに `make stg-uat-csv-import-preflight` → 承認 → `make stg-uat-csv-import` → `make stg-uat-csv-import-verify`。`STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND` 必須。**`--allow-local-rehearsal` 禁止。** band は重ならない。**城東を先に投入して第1段階を開始する。** 八王子（2 院目）は maintenance window。城東へ事前告知し、apply 中は STG 入力を止める（table lock は城東の並行登録も止める）。Make は `DB_*` を転送し、`--confirm-target-host` は `STG_UAT_CSV_IMPORT_CONFIRM_HOST`（既定 `db`）。リモートは USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE` と一致する `STG_UAT_CSV_IMPORT_CONFIRM_HOST`、sentinel をセットしてから `make`（`DB_HOST` 自動 alias なし。`pscale connect`→localhost 偽装は禁止）。backup / `DROP SCHEMA` 境界は [STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md)。エージェントは実行しない
- [ ] **H3-8** 片方成功・片方失敗なら、成功側は残し、失敗側だけ原因を直す。成功側 band は触らない
- [ ] **H3-9** `make stg-uat-staff-attach-preflight` → `make stg-uat-staff-attach`。`STG_UAT_STAFF_ATTACH_ROSTER` / `STG_UAT_STAFF_ATTACH_SECRETS`（repo 外 0600）必須。Make は `STG_UAT_STAFF_ATTACH_ALLOW_REMOTE` と `DB_*` を転送する。ローカル既定は `DB_HOST=db` / SSL `disable`。リモートは USER が `DB_HOST` / `DB_PORT` / `DB_SSL_MODE` と `STG_UAT_STAFF_ATTACH_ALLOW_REMOTE=YES_I_UNDERSTAND` をセットしてから `make`。初期パスワードは別経路。エージェントは実行しない
- [ ] **H3-10** 予約を試すなら未来日シフト（AE-STG-UAT-SHIFT）
- [ ] **H3-11** `stg.noah-karte.com` で対象医院に切り替え、飼主検索が実データで見えることだけ確認（行の中身はログに残さない）

### Lane 4 — 第1段階 並行運用

現場への伝え方: 「旧システムは今までどおり正本。STG にも同じ内容を入れて、動かないところを報告する。STG の入力は本番に移さない。」

- [ ] 期間中は STG の DB reset をしない（並行登録が消える）
- [ ] `staging` への backend デプロイは、**追記型 DDL のみ可。** 001 への統合・適用済み migration 編集・seed 差替えは並行期間中禁止（checksum mismatch で再構築が要る）。不具合修正のデプロイは現場に日時を知らせる
- [ ] スタッフが STG に新規作成した予約・カルテ・会計は移行 band の上に乗る。**本番へは持っていかない。** 終了後は D7 の破棄日で schema 再作成する
- [ ] 現行業務の欠測・遅延を起こさない。STG が落ちても旧システムで業務を続ける
- [ ] Workers Logs に氏名が混ざったら止めてクエリ/ログを直す。業務監査の正本は `audit_logs`
- [ ] 不具合は Linear。本ファイルには行値を書かない
- [ ] 必須 4 業務を連続 5 営業日再現し、両院が入り、現場が切替可と判断したら第2段階（移行日の決定）へ進む。上限 8 週

現場に渡してよい範囲:

| 業務 | rehearsal 城東が preflight を通った場合 | 八王子 bundle 未着 |
|---|---|---|
| 飼主・ペット検索、カルテ閲覧・記載 | 可（必須） | 不可（その医院） |
| 受付 | 可（必須） | 不可 |
| 予約参照 | 可 | 不可 |
| 今日の枠での新規予約 | シフト投入後（任意） | 同じ |
| 会計・入金・締め | preflight 次第（必須にしたいが通らなければ延期） | 不可 |
| 現行スタッフでのログイン | 後付けした id のみ | 後付けした id のみ |
| 検査画像（R2） | 欠ける前提（任意） | 同じ |

---

## 7. ログイン設計

| 方式 | 採用 | 理由 |
|---|---|---|
| 氏名から 414 アカウントを自動生成 | しない | 同名、全員 inactive、権限不明 |
| `staff-provision` で移行スタッフと同一人物の新規行を作る | しない | カルテ `doctor_id` とログイン ID がずれる |
| 移行 `staffs.id` に account を後付け | **する** | 現行スタッフほぼ全員。CSV の id が安定キー |
| email | `stg-staff-{id}@example.test` | ログイン専用。SMTP 送信には使わない |
| CSV に無い人の新規発行 | 補助として可 | 移行 CSV に居ないが現行の人だけ |
| 骨格にデモ admin を固める | しない | 既知パスワードを共有 STG に残さない |

---

## 8. 規模・失敗時

- 城東単独で CSV 約 1.8GB、検査結果約 800 万行、会計明細約 500 万行、合計約 2100 万行
- 八王子は bundle 未着のため未計測。同規模以上を仮定してメンテナンス窓を取る
- csv-import の CSV 上限は 512MiB/ファイル。現行城東の最大は約 370MB（exam_results / billing_items）で上限内。八王子は H0-2 で確認
- PlanetScale 接続枠とコンテナ 1GiB を超えないこと
- 長時間 apply は単一 transaction + table lock。**対象医院だけでなく、同じ STG 上の他医院の並行登録も止まる。** 2 院目は H3-7 の maintenance window で行う
- 並行期間中の STG reset は禁止（Lane 4）。reset すると 1 院目の並行登録も消える
- preflight 失敗: 書き込みなし。原因解消まで apply しない
- apply 中失敗: 自動 rollback。再実行は新しい作業確認のあと
- commit 応答不明: 再実行も restore も止め、read-only verify と DB 照合が先

---

## 9. old_db 側（本 repo では実行しない）

正本: sibling `old_db/todo.md` · `old_db/docs/by-clinic/jouto.md` · `old_db/docs/by-clinic/hachioji.md`

| ID | 内容 | STG 運用テスト | 本番 cutover |
|----|------|----------------|--------------|
| 城東 JOU-CSV-1 | TRUSTED 21 CSV | 不要（UAT は rehearsal 可） | **必須** |
| 城東 Z5 / G2 | 税セマンティクス・timing | UAT では待たない | **必須** |
| 八王子 HAC-CSV-1 | 医院 identity 付き 21 CSV + manifest | **必須**（rehearsal 可） | TRUSTED が必須 |
| 八王子 KNJO | 完全な TBL_KNJO_DATA または新規クリーン BAK | rehearsal でも品質が落ちる | **必須** |

live Postgres は一度に 1 医院。城東 live を八王子 load で上書きしない（old_db AGENTS.md）。

---

## 10. 着手順

```
D1–D7 確定（D2=B 城東先行。Q5 の「八王子 CSV 必須」は後追い HAC-CSV-1 として残す）
Lane 2: AE-STG-UAT-SKELETON / IMPORT / STAFF / MAKE-REMOTE  ← 完了（4703cf3e9）
次: 城東ローカル attach 済み。H3 / STG は USER。八王子は H0-2 待ち。follow-up: old_db が completed-without-payment を出さないこと、同一 medical_record に複数 billing を出さないこと。ログインする人は `set_active` / is_active を画面で有効化
Lane 3: AE-STG-UAT-JOU → 第1段階開始（現場へ「城東のみ」）  ← USER。未着手
old_db: 八王子 21 CSV（HAC-CSV-1。rehearsal 可。医院 identity 必須）
Lane 3 続き: AE-STG-UAT-HAC（H3-7 maintenance window。城東の STG 入力を止める）
（予約も並行するなら）AE-STG-UAT-SHIFT
第1段階: 各医院スタッフが旧システムと STG に同じ業務を登録
必須 4 業務 × 連続 5 営業日、上限 8 週、両院投入後に切替判断
第2段階: 本番移行日を決める → その日の旧データ出力 → 本番 F6（PASS / TRUSTED）
（後で）AE-SEED-RETIRE-DEMO
第2段階は JOU-CSV-1 + 八王子 TRUSTED。STG 並行登録は入力にしない
```

---

## 11. 完了条件

第1段階（STG 並行運用）:

- **開始:** 城東の 21 表が band に入り verify が通っている（rehearsal なら report にその旨）。現場へ「城東のみ」と伝える
- **八王子:** HAC-CSV-1 後、maintenance window で投入。切替判断の前に両院が入っていること
- 003 由来のデモ臨床が無い。骨格が 21 表の band を占有していない
- 対象医院の現行スタッフほぼ全員が自分の医院でログインし、必須 4 業務を STG でも登録できている
- 必須 4 業務を対象医院で連続 5 営業日再現し、上限 8 週以内に切替判断する
- スタッフが「STG の入力は本番に移らない。正本は旧システムのまま」と理解している
- 本番 F6 の TRUSTED 契約が残っている
- 行値・パスワードが Git / 本ファイル / 標準ログに無い

第2段階（本番移行。第1段階の印のあと）:

- 移行日が決まっている
- その日の旧データ出力と本番 F6（`PASS` / `TRUSTED_CANDIDATE`）の run sheet がある
- 入力は移行日の旧システム出力であり、STG の並行登録ではない

---

## 12. 次の一手

1. ~~Lane 2 コードを commit~~ → 済み（`4703cf3e9`）
2. ~~H0-5 ローカル attach~~ → 済み（414 / digest `a6ea3cdd6d2b36bea961dfd00faadbfc88fb4e0b59b194d02baf5b6bb1ed71ff`）。パスワードは `sensitive-local/stg-uat-staff-secrets.json`（gitignored）。STG attach は Lane 3。**2026-08-26 USER ローカルログイン確認済み**（is_active は画面で有効化）
3. ~~城東 handoff check + manifest SHA-256 転記（H0-3a / H0-4）~~ → 済み（`jouto-intake-20260822-01` / SHA `7bbda50f06f7d0acac6711d1a73b78ca68b835ee9be5df6cd04f3e6a5094a405`。Class A: payments snapshot、pending 再分類、billing×medical_record unique 余剰 192 リンク解除。stage 再実行なし）。八王子は H0-2 後
4. ~~Lane 1 で城東をローカル証明~~ → 済み（1 回目 2026-08-25 + **2 回目 2026-08-26** `make reset` / `csv-import` apply+verify PASS）。次は USER が Lane 3 で城東を先に STG へ投入し第1段階開始（`stg-uat-csv-import`。ローカル rehearsal とは別ゲート。H3 は未チェックのまま）
5. 並行して old_db が八王子 21 表（HAC-CSV-1）を出す。出来次第 maintenance window で投入
6. M-3 の runbook 1 行は、明示指示があるときだけ
7. claim 枝の削除は merge / abandon のあと USER が行う

---

## 13. 参照

| 文書 | 役割 |
|---|---|
| [docs/ops/deploy/CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) | F6 21 表 cutover |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | `_old_db_handoff` のローカル隔離 |
| [docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md](docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) | seed と 21 表の境界、`APP_ENV` ゲート |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 再作成・直結・破壊境界 |
| [docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md](docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md) | STG は master-only。本書の実データ UAT は例外（canonical 側への 1 行追記は M-3、未適用） |
| [docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) | 新規 staff 発行（後付けではない） |
| [docs/ops/deploy/A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) | 隔離画面確認 |
| [docs/ops/deploy/LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) | ローカル reset |
| [docs/ops/infra/staging/runbook.md](docs/ops/infra/staging/runbook.md) | STG 障害初動 |
