# EMR-257: 入院会計の前金・複数入金対応（預り金充当）設計書

作成: 2026-10-02 / 状態: 設計提案（レビュー用）・実装未着手 / 対象: 設計のみ（production code・migration・OpenAPI は本書では変更しない）

親: EMR-253（waiting 会計確定経路の修復）の業務フロー調査から分離した設計課題。
関連: EMR-66（1カルテ/1入院=1会計の 409 契約）、EMR-188（月末未納者一覧・期間繰越）、EMR-196②（unbilled 集約の版照合）、EMR-228（期間内売上のある飼主に限定する未納一覧の絞り込み）、BUG-018（complete の原子確定）、BUG-007（未収残高定義）。

## 1. 背景と要件

- 入院の支払いタイミングは **入院前・入院時・退院時・後日** すべてに実在し、1 入院に **複数入金（前金＋退院時精算）** が実際に発生する（オーナー回答 2026-10-02、EMR-253 へ記録済み）。
- 旧 vendor システムの精算画面は **預り金・前回繰越金・今回繰越金** を持ち、残額繰越と預かり充当を処理していた。新システムには該当概念が未実装。
- 現行「1入院=1会計」の 409 契約（EMR-66）と「complete=新規作成 or 既存 waiting 行の takeover（EMR-253）」の構成では、入院中の前金会計と退院時精算会計の併存を表せない。
- 完了条件: 入院会計の支払いタイミング全パターン（入院前/時・退院時・後日・前金+精算）を記録できる設計が確定し、必要な契約変更が列挙されていること。実装は設計確定後の別チケットでよい。

## 2. 現行実装の要点（本設計が拡張/維持する対象）

| 領域 | 現行契約 | 根拠 |
|---|---|---|
| 会計スロット | `billings.medical_record_id` は全行で UNIQUE（soft-delete 済みも占有）。`billings.hospitalization_id` は `deleted_at IS NULL` 限定の部分 UNIQUE —— **1入院=1 active 会計**。 | `backend/migrations/001_init.sql` L2197, L2201-2203 |
| 409 契約 | `FindCompleteConflict` が medical_record_id / hospitalization_id スロット占有者を検出し、`ACCOUNTING_ALREADY_COMPLETED`（409）+ 既存会計を返す。 | `backend/internal/billing/accounting_repository_complete.go` L21-63、`accounting_complete.go` L83-106, L330-351 |
| complete | `POST /accountings/complete` が header/items/totals/payment/splits/監査を単一 tx で原子確定。`Idempotency-Key`(UUID) + payload digest で冪等。`payment_splits` 合計は `billing_amount`(= total − 保険 − 割引) と完全一致が必須。 | `accounting_complete.go` L228-303、`accounting_complete_tx.go` L248-347、`accounting_service_builders.go` L78-103 |
| takeover（EMR-253） | `billing_id` 明示 or `hospitalization_id` スロットの waiting 占有者を FOR UPDATE でロックして in-place 確定（既存明細は request items で置換）。cancelled 占有者は同一 tx で soft-delete してスロット解放。completed 占有者は 409。 | `accounting_complete_tx.go` L54-61, L380-462 |
| 退院時自動作成 | `POST /hospitalizations/:id/discharge-with-billing`（`create_accounting`）が care_plan_items 全量を `source=hospitalization` の明細へ転記して waiting billing を作成。EMR-253 の no-duplicate ガードで既存非取消 billing は再利用（二重作成しない）。 | `backend/internal/medicalrecord/hospitalization_discharge_tx.go` L57-92, L94-141、`routes_hospitalization.go` L15 |
| 支払い記録 | `payments.billing_id` は UNIQUE —— **1会計=1 payment 行**。payment は請求スナップショット（subtotal/tax/total/insurance/discount/billing_amount）＋収受額（received/change: 現金のお預かり・お釣り）。`payment_splits` はその 1 入金イベントの手段別内訳。 | `backend/migrations/001_init.sql` L1963-1983、`backend/internal/model/accounting.go` L174-224 |
| 未納・未収集約 | `unpaidAmountSQL`: payment 無し waiting は `total_amount` 全額、payment 有り waiting/completed は `patient_due − payments.billing_amount` の残差（0 下限）。pending/cancelled は未納に計上しない。未納者一覧・飼主残高・月末繰越（EMR-188/228）がこの式を共有。 | `backend/internal/billing/unpaid_amount.go` L18-30、`accounting_repository_unpaid.go` |
| 未請求集約 | `aggregateUnbilled` のソースは treatments / trimming / vaccinations / exams のみ。**care_plan_items は入院中の unbilled 集約に入らない**（退院時に一括転記される設計）。よって入院中の中間請求は現状手入力/見積ベース。 | `backend/internal/billing/billing_item_unbilled.go` L21-80 |
| 預り金・前金 | **存在しない**。`received_amount` は「お預かり（支払時に受け取った現金）」であり、預託残高の概念ではない。 | grep 全件確認済み（backend/docs に預り金残・前金エンティティなし） |

## 3. 支払いタイミング全パターンと記録経路

採用案（§4）に基づく各パターンの記録先。「入金イベント」は金銭の受領・充当・返金の記録であり、「会計（billing）」は請求書・領収書たる確定書類である。この区別が本設計の基軸。

| パターン | 発生時点 | 記録先エンティティ | API 経路 | 備考 |
|---|---|---|---|---|
| 入院前（予約時・入院日前の前金） | 入院レコード作成前後 | `advance_deposits`（預り金。`hospitalization_id` は `reserved` 入院行へリンク可、未作成なら NULL で owner スコープ預り） | `POST /hospitalizations/:id/deposits`（新設）。入院行が無いケースは先に `POST /hospitalizations`（status=reserved）で作成する既存経路を使う | 預り金は売上・請求ではなく前受金。領収書ではなく預り証の扱い（要 PO 確認 §8） |
| 入院時（チェックイン時の前金・保証金） | `admitted` 入院行 | 同上 `advance_deposits` | 同上 | 複数回・複数手段を許容（入金イベントは N 件） |
| 入院中前金（入院中の一部入金・追加前金） | 入院中任意のタイミング | 同上 `advance_deposits` | 同上 | 入金に明細（billing_items）は不要。中間「請求書」が必要かは別論点（§7.3, §8） |
| 退院時精算 | 退院処理〜会計確定 | `discharge-with-billing` が作る waiting billing（EMR-253 ガードで再利用含む）→ `POST /accountings/complete` の takeover で確定。預り金は `deposit_applications` で充当 | 既存 `POST /hospitalizations/:id/discharge-with-billing` + 拡張 `POST /accountings/complete`（`billing_id` + `deposit_applications` + `deferred_amount`） | 充当額 + 当日収受（payment_splits）+ 繰越残（deferred）= 患者請求額 |
| 後日（退院後の支払い・請求書払い） | 精算日以降 | 精算 billing を waiting のまま残す（既存未納一覧に載る）→ 入金日に complete takeover で確定。または確定時に `deferred_amount` を明示して残額を未収として残す | 既存 takeover 経路 + 拡張 complete | 現行も waiting 残留で未納追跡できるが、部分入金済み残は表現不可 → §6 の complete 契約拡張で解消 |
| 残額繰越（退院時に一部のみ回収、残は未収） | 退院時精算 | completed billing の `payments.billing_amount` に「当日収受 + 預り金充当」のみ計上し、残差は未収残高として未納一覧・月末繰越へ自動集約 | 拡張 `POST /accountings/complete` | `unpaidAmountSQL` の残差分岐（completed + payment 有り）がそのまま残額繰越を拾う。SQL 変更不要 |
| 預り金繰越（充当しきれなかった預り残） | 精算時に預り金 > 充当額 | `advance_deposits` の未充当残（amount − Σapplications − Σrefunds）を owner 預り残として保持し、次回会計へ充当 or 返金 | 新設 `GET /owners/:id/deposit-balance` 等 | 旧 vendor の「今回繰越金」相当 |

## 4. 推奨設計（採用案）

### 4.1 結論: 1入院=1会計を維持し、「預り金（advance deposit）エンティティ + 充当記録」を導入する

**推奨**: 入院に対する入金は会計（billing）ではなく **預り金イベント** として記録し、退院時の唯一の精算 billing に `deposit_applications` で充当する。「1入院=複数会計」は採らない（§5.1）。

この選択の理由:

1. **意味の正しさ**: 前金・預り金は確定した請求ではなく前受金であり、金額は見積・暫定でよい。billing は `items ≥ 1`・server 再計算 total・領収書発行を伴う確定書類であり、「明細のない入金」を billing で表すと帳票の請求金額が虚偽になる。
2. **契約の爆発半径**: `idx_billings_hospitalization_id_unique`・EMR-66 の 409（既存会計を返す UX）・EMR-253 の takeover / 退院 no-duplicate ガード・`FindByHospitalizationID` の単一占有前提・`payments.billing_id` UNIQUE・未納集約 —— これら全てが「1入院=1会計/1会計=1 payment」を前提に組まれている。預り金を別エンティティにすればこれらは一切変更不要（§6.2）。
3. **売上の二重計上回避**: 前金を completed billing で記録すると deposit 会計 + 精算会計で日次/月次/LTV/レポートが二重計上になる。除外フラグを全集計クエリへ入れるより、入金イベントを売上集計の対象外に置く方が誤用に強い。
4. **旧 vendor との対応関係**: 預り金残高・充当額・今回繰越がそのままエンティティの状態として表せる。

### 4.2 新エンティティ（設計のみ — migration 実装は別チケット）

`advance_deposits`（預り金）:

| 列（案） | 意味 |
|---|---|
| `id`, `clinic_id` | PK・テナントキー（clinic スコープ必須・RLS 方針は payments と同型） |
| `owner_id` | NOT NULL。預かり主体（飼主）。充当は同一 owner の billing に限定 |
| `pet_id`, `hospitalization_id` | NULL 可。意図した入院へのリンク（reserved/admitted 問わず）。未確定の預りは NULL で owner スコープ |
| `amount` | 預り額（円・正） |
| `method` / `payment_method_id` | 支払手段（payment_method ENUM + マスタ dual maintain、payments/payment_splits と同規約） |
| `received_on` | 入金日（date。レジ締め・日次集計の計上日） |
| `status` | `received` / `partially_applied` / `applied` / `refunded` / `cancelled` |
| `request_id` | 冪等キー（Idempotency-Key と同型 UUID。complete の completion_request_id 規約に準拠） |
| `memo`, `paid_by`, `created_at`/`updated_at`/`deleted_at` | 補足・担当者・標準監査列 |

`deposit_applications`（預り金充当）:

| 列（案） | 意味 |
|---|---|
| `deposit_id`, `billing_id` | 充当元預り金・充当先会計（複合で一意性は持たせない: 部分充当を許す） |
| `amount` | 充当額（円・正）。1 預り金の Σ applications ≤ amount |
| `applied_by`, `created_at`, `deleted_at` | 監査列 |

不変条件（service 層で tx 内検証）: 充当は同一 clinic・同一 owner、対象 billing の `status ∈ {waiting}` への complete 確定 tx 内でのみ行う（completed への事後充当は返金経路に限定）。

### 4.3 API 追加・変更（設計のみ）

新設:

- `POST /hospitalizations/:id/deposits` — 入院前・入院時・入院中前金の記録。`Idempotency-Key` 必須、冪等 replay 規約は complete と同型。`hospitalization_id` はパスから、`owner_id`/`pet_id` は入院行から server 解決（矛盾送信は 400）。
- `GET /hospitalizations/:id/deposits` — その入院の預り金一覧と残高（amount − applied − refunded）。
- `GET /owners/:id/deposit-balance` — 飼主の預り残高（今回繰越の源泉）。未納残高 `GET /accountings/unpaid-balance` と対になる表示。
- `POST /deposits/:id/refund` — 未充当分の返金（部分返金可・`billing_refunds` と別物: 会計の返金ではなく預り金の返却）。
- `POST /deposits/:id/cancel` — 誤記録の取消（取消権限の新設か `accounting-cancel` 流用かは §8 未決）。

変更（`POST /accountings/complete` の拡張）:

- request に `deposit_applications: [{deposit_id, amount}]` と `deferred_amount`（円・後日回収する残額繰越）を追加。両者は冪等 digest（`ComputeCompleteAccountingDigest`）の入力に含める —— 操作内容の一部であり、retry 同一性判定に必要。
- tx 内で: 対象 deposit を FOR UPDATE ロック → clinic/owner/入院整合・未充当残 ≥ 充当額を検証 → `deposit_applications` 行を作成し deposit status を遷移。
- 金額等式の変更: 現行 `Σ payment_splits.amount == billing_amount(= total − 保険 − 割引)` を、**`Σ splits + Σ 充当 + deferred_amount == patient_due`** へ拡張。`payments.billing_amount` には `Σ splits + Σ 充当`（= 実収受額）を保存する → 未収残高は `patient_due − billing_amount = deferred_amount` となり、既存 `unpaidAmountSQL` の残差分岐がそのまま残額繰越を拾う（`unpaid_amount.go` L18-30、SQL 変更なし）。
- fail-closed: `deferred_amount` を省略した場合は従来どおり全額決済必須（`Σ splits + Σ 充当 == due`）。残額を残すには明示送信を要求し、「splits 不足が黙って未収になる」事故を防ぐ。`deferred_amount` と実計算の不一致は 400。
- 預り金のみで全額充当（当日収受 0）は `splits` 空 + `deferred_amount=0` で表現（現行の `billingAmount>0 かつ splits 空 → 400` は「充当後の残余>0 かつ splits 空かつ deferred 未指定」へ条件変更）。

### 4.4 預り金の充当表現（画面・帳票）

- 精算画面（`11-accounting-detail.md` 右カラム相当）に「預り金充当」領域を追加: 充当可能な預り金一覧（入院紐付け優先、その他 owner 預り）、充当額入力、「お預かり(預り金)」「充当後の請求残」「今回収受」「残額繰越」「預り金残（今回繰越）」の表示 —— 旧 vendor 精算画面の預り金・前回繰越金・今回繰越金に相当。
- 「明細兼領収書」（`AccountingDocument`）に充当行を追加: 預り金充当額・当日収受額・残額（未収繰越）・預り金残。預り金の受領そのものは別途「預り証」として印字するかは PO 判断（§8）。
- 入院詳細画面に預り金一覧・残高表示を追加（`08-hospitalization-detail.md` 系のパネル）。

### 4.5 監査・権限・レジ締め

- deposit の create/refund/cancel と complete 内の充当は既存監査（`auditTx.LogEntryTx` 系）へ記録。complete 監査の NewValue に `applied_deposit_total`・`deferred_amount` を含める。
- 権限: 作成は `accounting` create、返金・取消は専用権限か `accounting-cancel` / `accounting-post-close-edit` 相当 —— §8 未決。
- レジ締め: 預り金入金は現金実残に影響するため日次レジ締めの集計対象。ただし売上ではなく前受金として「預り金入金」を別区分で表示する（`daily-summary` / 締め帳票の応答 shape 変更が必要 —— §6.1 / §8）。
- 締め後日付の預り金記録・充当は `post_close_reason` 規約を踏襲（complete の post-close 経路が既に `resolvePostCloseInTx` で検証済みなのと同列の扱い）。

## 5. 却下した代替案とトレードオフ

### 5.1 却下案A: 1入院=複数会計（前金会計 completed + 精算会計 waiting→completed）

一見既存 complete の再利用で済むが、以下の契約を同時に壊すため不採用:

- `idx_billings_hospitalization_id_unique`（部分 UNIQUE）の緩和が必須 → EMR-66 の 409 は「どの会計がスロット占有者か」が一意に定まらず、`ACCOUNTING_ALREADY_COMPLETED` + 既存会計返却の UX 契約が曖昧化する。
- `FindByHospitalizationID`（EMR-253 の takeover 解決と退院 no-duplicate ガード共有）の単一占有前提が崩れ、「どの行を takeover/reuse するか」の選定規則を新設しなければならない。
- 前金会計は completed = 領収書発行済みの売上書類となるため、items ≥ 1 の制約上ダミー/見積明細を確定書面に載せることになり、日次集計・月次レポート・LTV・未納集約の全クエリに「前金会計除外」述語を配線しないと二重計上になる（爆発半径が大きい）。
- 精算時の充当を別途表す仕組み（負明細・値引充当）は結局必要になり、預り金エンティティの仕事が消えない。

採用見送りの帰結として、**「1入院=1会計」の 409 契約（EMR-66）は維持する**（§6.2）。

### 5.2 却下案B: 1会計 + payments 複数行化（payments.billing_id UNIQUE 撤廃）

「1会計に複数入金イベント」を payments テーブルの複数行で実現する案。不採用理由:

- `payments` 行は請求スナップショット（subtotal/tax/total/insurance/discount/billing_amount）を持つ「1会計の精算結果」であり、前金時点では請求額が未確定でスナップショットが書けない。
- `payments.billing_id` UNIQUE の撤廃は、`Payments[0]` を使う全 reader（`OutstandingAmount`、response builder、帳票、未納 SQL の LEFT JOIN 意味論）へ多重行対応を強いる。`leftJoinPaymentsSQL` は 1 行前提で残差計算しており、複数行化は未収定義（BUG-007）の再設計を要求する。
- 目的（入金イベントの複数記録）は §4 の別エンティティでより小さい変更量で達成できる。

### 5.3 却下案C: hospitalization_id 非束縛の中間会計（現行でも可能な workaround）

入院中の中間請求を `hospitalization_id=NULL` の手入力 billing で切る案（スロットを占有しないため契約変更ゼロ）。次の理由で「前金対応の答え」にはならない:

- 入金と入院の紐付けがなく、預り残高・充当・繰越を表せない（帳票・未納・owner 残高のどこにも現れない）。
- 退院時自動作成は care_plan_items を全量転記するため、中間会計で請求済みの項目との二重請求を手動で防ぐ運用負荷が残る。
- ただし「入院中に正式な請求書を発行したい」要件が別途ある場合の逃げ道としては存在を認める（§8 未決事項参照）。

## 6. 必要な契約変更の列挙

### 6.1 変更するもの

| 層 | 変更 |
|---|---|
| schema | `advance_deposits`・`deposit_applications` 新設（別 migration。`idx_billings_hospitalization_id_unique`・`payments.billing_id` UNIQUE は**不変**） |
| OpenAPI (`backend/docs/api.yaml`) | deposit 系 endpoint、`CompleteAccountingRequest` への `deposit_applications`/`deferred_amount` 追加、応答への充当内訳・預り残高フィールド |
| `POST /accountings/complete` | 金額等式を `Σ splits + Σ 充当 + deferred == patient_due` へ拡張（`validatePaymentSplits` の呼出側）。digest 入力に新フィールド追加。tx 内で deposit lock・充当作成・status 遷移 |
| 未納集約 | `unpaidAmountSQL` 自体は不変（`payments.billing_amount` に充当込みの実収受を入れるため残差が自動的に繰越額になる）。**注意**: 充当を billing_amount に含めない実装を取ると残差が二重計上される —— 実装チケットで検証必須 |
| 退院時自動作成 | 不変（EMR-253 ガード維持）。精算 billing が預り金の充当先になるだけで、作成経路は変えない |
| 日次/月次・未納帳票 | 預り金入金を「売上」と別区分で集計・表示（前受金）。未納者一覧の行定義は不変だが、残額繰越が増えるため表示文言（内訳に預り金充当・繰越を含む旨）の整理が必要 |
| 帳票（明細兼領収書・預り証） | 充当額・当日収受・残額繰越・預り金残の行追加。預り証発行の要否は §8 |
| 監査・権限 | deposit 系操作の audit action 追加、権限割当（§8 未決） |
| FE | 精算画面の預り金充当 UI、入院詳細の預り金パネル、会計一覧/未納一覧の残額繰越表示整合 |

### 6.2 明示的に変更しないもの（維持する契約）

- **1入院=1会計**: `idx_billings_hospitalization_id_unique`、`FindCompleteConflict` の意味論、`ACCOUNTING_ALREADY_COMPLETED` 409（EMR-66）は不変。前金は billing を作らないためスロットを占有しない。
- **EMR-253 takeover**: `billing_id` 明示・入院スロット waiting 占有者の in-place 確定、cancelled 占有者の soft-delete 解放は不変。退院自動作成の no-duplicate ガードも不変。
- **payments 1:1**: `payments.billing_id` UNIQUE 維持。入金イベントの複数性は deposit/application 側に置く。
- **unbilled 集約**: `aggregateUnbilled` のソースに care_plan_items を追加しない（§7.3 参照。別設計課題）。
- **冪等・409・409 UNBILLED_ITEMS_CHANGED（EMR-196②）の振る舞い**: takeover 時の revision 非必須規約を含め不変。

## 7. 設計が答える論点（チケットの検討要件への回答）

### 7.1 1入院=複数会計か、1会計に複数入金か

→ **1会計に複数入金イベント（預り金モデル）を採用**（§4.1、却下理由は §5.1/§5.2）。「会計」は退院時の唯一の請求書、「入金」は時系列の前受金イベント、という責務分離。

### 7.2 預り金/前金の最終精算への充当方法

→ `deposit_applications` 充当行で表現（充当元預り金・充当先 billing・額）。画面は充当額入力＋「預り金・充当後残・今回収受・残額繰越・預り残（今回繰越）」を分離表示（§4.4）。充当は complete tx 内で原子的に確定し、`payments.billing_amount` は充当込みの実収受額とする。

### 7.3 入院中会計が手入力/見積ベースである点との整合

→ 預り金は「入金の記録」であり明細を要求しないため、care_plan_items が unbilled 集約に入らない現行設計と矛盾しない。**入金だけ先に記録し、請求の確定は退院時に一元化する**のが本設計の整合方針。入院中に正式な請求書（確定 billing）を発行したい需要が残る場合は別課題（§8: care_plan_items の unbilled ソース化 or 中間 billing の hospitalization 紐付けと二重請求防止）。

### 7.4 退院時自動作成・409 契約との関係

→ 預り金は billing ではないため EMR-253 の no-duplicate ガード・EMR-66 の 409 には一切干渉しない。退院時の waiting billing 作成・再利用・takeover 確定は現行のまま動き、complete に充当処理が乗るだけ（§6.2）。

### 7.5 未収・繰越の管理と未納者一覧の整合

→ 残額繰越 = completed billing の `patient_due − billing_amount`（`deferred_amount` と一致）として既存 `unpaidAmountSQL` が拾い、未納者一覧（会計単位/飼主単位）・月末未納者一覧（期間前繰越/期間内未納/期末繰越、EMR-188/EMR-228）へそのまま載る。預り金側の繰越（未充当残）は未納ではなく前受金として別管理（`GET /owners/:id/deposit-balance`）。pending ステータスは未納集約対象外の現行契約を維持する（後日払いは waiting 残留 or completed+残差で表し、pending に新しい金銭意味を持たせない）。

## 8. 未決・要承認事項（open items）

以下は人/PO の判断が必要な未確定事項。設計本体はこれらの結論を待たずに確定可能だが、実装チケット化の前に回答が必要:

1. **預り金の充当スコープ**: 本書は「同一 owner の billing へ充当可（hospitalization 紐付けは表示・優先順位に使う）」を推奨とした。入院スロット内だけに限定するか（厳格だが 今回繰越 を別入院/外来へ回せない）、owner スコープにするか —— PO 承認が必要。
2. **預り証・帳票**: 預り金受領時の預り証発行要否、税務上の扱い（前受金の非課税/課税時点）、「明細兼領収書」への充当行の確定文言 —— PO/税務確認。
3. **レジ締め・日次集計**: 預り金入金を daily-summary / 締め帳票へ別区分で出すか、現金残の一部としてだけ出すか —— 表示仕様の確定。
4. **入院中の中間「請求書」需要**: 預り金（入金記録）で足りるか、確定 billing としての中間請求書も必要か。後者なら §5.3 の workaround か、care_plan_items の unbilled ソース化（`aggregateUnbilled` 拡張）が別途必要 —— PO の業務確認。
5. **預り金操作の権限モデル**: 作成/返金/取消に `accounting` create・`accounting-cancel`・新設権限のどれを割り当てるか。
6. **上限・整合ルール**: 預り金の金額上限、充当順序（FIFO/指定）、預り金 > 請求額時の取り扱い（即時返金か預り残繰越か —— 本書は未充当残保持を推奨）。
7. **deferred_amount（残額繰越）の許可条件**: 誰でも残額繰越で complete できるか、権限・理由入力（post_close_reason 類似のメモ必須化）が要るか —— 未納放置の抑止設計として要 PO 判断。
8. **既存 pending ステータスとの関係**: 後日払いを waiting 残留＋complete で扱う本設計で pending を使わないことを確認（pending に「後日精算予約」の意味を新設しない —— 未納集約対象外のため）。

## 9. 検証済み事実と参照先

- complete/takeover/409: `backend/internal/billing/accounting_complete.go`、`accounting_complete_tx.go`、`accounting_repository_complete.go`（`FindCompleteConflict` L21、`FindByHospitalizationID` L70、`SoftDeleteCancelled` L91）
- 退院時会計作成と no-duplicate ガード: `backend/internal/medicalrecord/hospitalization_discharge_tx.go` L57-92
- 未納・繰越集約: `backend/internal/billing/unpaid_amount.go`、`accounting_repository_unpaid.go`（`FindUnpaidByBilling`/`FindUnpaidByOwner`/`SumUnpaidByOwner`/`FindPeriodUnpaidCarryover`）
- エンティティ・制約: `backend/internal/model/accounting.go`（Billing L80-116, Payment L174-203 — `billing_id` uniqueIndex L176, PaymentSplit L205-224）、`backend/migrations/001_init.sql` L1963-1983（payments 1:1）, L2197-2203（会計スロット UNIQUE）
- 支払内訳検証: `backend/internal/billing/accounting_service_builders.go` L78-121（`validatePaymentSplits`/`validateCashSplit`）
- 未請求集約ソース: `backend/internal/billing/billing_item_unbilled.go` L21-80
- API 契約: `backend/docs/api.yaml` `CompleteAccountingRequest`（L2278-）
- 画面契約: `docs/spec/screens/11-accounting-detail.md`（精算・保留・未納導線）、`docs/spec/screens/30-unpaid-list.md`（未納/月末繰越）、`docs/spec/screens/08-hospitalization-detail.md`（退院プロセスと会計連携）
- チケット: Plane `EMR-257`（親 EMR-253）、オーナー裁定 2026-10-02
