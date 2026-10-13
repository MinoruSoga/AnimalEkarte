# EMR-364/365: 現行経路の実測記録（CURRENT-PATHS）

**状態: 調査記録・実装なし。** この文書は現行コードの実際の経路を path:line で記録するものであり、契約・CSV・コードの変更提案・採用を意味しない。提案の正本は単一 artifact [DESIGN.md §8（履歴税率）](../emr-342-medicines-contract-design/DESIGN.md)・§9（ワクチン単価の精度）に置き、producer/consumer が同一 artifact に receipt で合意するまで採用されない。本書は競合する契約 shape・pin・profile を持たない。

- 正本の契約実行面: `backend/internal/csvimport/cutover_contract.go`（21 表、本票は第 2 の実行契約を作らない）。
- 権威ある ticket（Plane）: EMR-364・EMR-365（consumer 契約確認）、MIG-44（税・率・総額 source）、MIG-31（ワクチン価格写像）、MIG-37（台帳/pin）、MIG-38（税復号実測）、MIG-39・MIG-41（生成）、MIG-40（merchandise 負値の別件）、MIG-15（treatments 表の別提案）。
- 上記 ticket のローカルスナップショット（private 文脈、任意）: `/private/tmp/ae-csv-fusion-20261013/evidence/{EMR-364,EMR-365,MIG-15,MIG-31,MIG-37,MIG-38,MIG-39,MIG-40,MIG-41,MIG-44}.md`。観測日 2026-10-13、workspace baseline `ed04c0237f81ddc2890d39b528d3608a58180c8b`。
- producer source（READ ONLY、実測・編集とも不可）: `/Users/minoru/Dev/Case/AnimalHospital/old_db/scripts/lib/stage-csv-columns.mjs`（HEAD `b1590fd93ea7d4b729dfdabe066fb3b1b9d918bf`）。
- 記録方法: コード検査と ticket 記録にもとづく静的記録。live DB・実 CSV・実行 apply の実測は行っていない（NOT RUN）。

## 1. EMR-364: `billing_items` の現行 10 列契約と `tax_rate` 不在経路

### 1.1 producer 側（stage → CSV）

- `targetColumns("billing_items")` は 10 列: `id, clinic_id, billing_id, category, name, unit_price, quantity, tax_type, is_insurance_applicable, sort_order`（`stage-csv-columns.mjs:415-417`）。`tax_rate` は CSV header にも stage 写像にも無い。
- `selectListSql("billing_items")` も `tax_rate` を SELECT しない（`stage-csv-columns.mjs:553-559`）。stage 側に `billing_items.tax_rate` の source 列自体が存在しないため、producer は `COALESCE(tax_rate,…)` のような既存列への既定適用すら書けない（COALESCE は実在する列の NULL しか救えない）。列追加には producer 側の stage 写像・source SQL の変更が要り、それは MIG-44 側の未実装範囲。
- 対照的に `estimate_items` は stage 側の `tax_rate` 列を `COALESCE(tax_rate,0.10)` で出している（`stage-csv-columns.mjs:430-434`・`:593-601`）— 既存列への既定適用であり、`billing_items` には同等の列・同等の行ともに無い。
- `TBL_BILLING_ITEMS` の stage 参照は `billings` FK 検証にのみ現れる（`stage-csv-columns.mjs:297-301`）。税フラグ `TZei_Flg`・精算日の写像は MIG-44 側の stage SQL であり、本 producer 候補では `tax_rate` の生成は未実装（EMR-364 最新コメント: 「tax_ratestage/CSV列・5/8/10日付境界はまだUNIMPLEMENTED」）。

### 1.2 consumer 契約（実行正本）

- `CutoverTableSpecs()` の `billing_items` spec は同じ 10 列、`tax_rate` 無し（`backend/internal/csvimport/cutover_contract.go:256`）。
- 兄弟の金銭表は `tax_type` 直後に `tax_rate` を既に持つ: `procedures`（`cutover_contract.go:245`）、`merchandise_items`（`:246`）、`estimate_items`（`:260`）。
- compiled pin: `cutoverCSVContractSHA256 = 9e02897b86c7…`（`cutover_contract.go:19`）。manifest の `csvContractSha256` がこれと一致しないと formal/staging は拒否（`cutover_contract_validate.go:313-316`・`:355-358`）。
- rehearsal のみの容認: `rehearsalContractDrift`（`cutover_contract.go:56-58`）は REHEARSAL_ONLY artifact かつ rehearsal provenance の場合だけ、manifest の `csvContractSha256` 差異（`cutover_contract_validate.go:400-405`）、tables[] 順（`:270-284`）、CSV header 順（`:609-631`）、importablePredicate・placeholder 差異（`:188-203`）を記録付きで許す。**これは現行 artifact 向けの既存機構であって、列集合の増減（10→11 列）は permutation の範囲外 — 未知列・欠列は rehearsal でも拒否される**（`cutover_contract_validate.go:621-630`）。「rehearsal で受理された」は新 profile 採用の証拠に使えない。

### 1.3 preflight（source 側信頼境界）

- `PreflightCutoverBundle` は manifest sha256 一致 → manifest decode/検証 → per-CSV 検証の順（`cutover_contract_validate.go:24-63`）。
- `validateCutoverCSV`: header を読み row arity を header 長に固定し（`:558-561`）、`cutoverColumnIndexes` が列集合を検査 — formal は順序厳格、rehearsal は同一集合の順序のみ許容、未知列・重複・欠列は常に fail-closed（`:609-631`）。各行で `id` の整数性・一意性（`:581-589`）、placeholder の位置・形式（`:634-653`）、band 列の整数性・範囲（`:654-670`）を見て、行数と manifest `sha256` まで一致させる（`:594-600`）。
- **`unit_price`/`tax_rate` 等の金銭セルは preflight で数値解析されない**。placeholder・band 検査に通る任意文字列はそのまま COPY に流れ、PostgreSQL の型変換に委ねられる（→ §1.4）。

### 1.4 apply（transform → COPY → insert）

- `transformCutoverCSV` は安定 open（symlink/所有者/TOCTOU、`:558-601`）の同じバイトを hash しながら、rehearsal 時のみ header の列順を正規化し（`:514-532`）、placeholder を置換（`:533-537`）して temp CSV へ写す。preflight 以降に sha256 が変われば拒否（`:551-554`）。
- COPY は `spec.Columns` だけを列挙した列リストで実行（`cutover_import_copy.go:195-207`）。**契約に無い列は一切送らない**ため、`billing_items.tax_rate` は COPY に現れない。
- staging 一時表から本表への `INSERT … SELECT` も同じ `spec.Columns` を使う（`cutover_import_copy.go:214-230`）。`tax_rate` は select されず、挿入列にも入らない。
- target の必須列は `information_schema` で検査（`cutover_import.go:273-286` — NOT NULL・no-default・非 identity・非 generated の列のみ fail）。`tax_rate` は契約列ではなく nullable + default 付きでもあるため、どちらの面でも参照・警告されない。

### 1.5 DB 既定と下流

- DDL: `billing_items.tax_rate numeric(3,2) DEFAULT 0.10`（**nullable**。`backend/migrations/001_init.sql:1921`）。INSERT に列を含めないので PostgreSQL の DEFAULT が全移行行に `0.10` を入れる。なお、仮に将来列を追加しても**明示的に空セルを出せば NULL になる**（`NULL ''` 規約）ため、列を足すだけでは値の保証にならず、提案の列挙検証が別途要る（[DESIGN.md §8](../emr-342-medicines-contract-design/DESIGN.md)）。
- `billing_items` 表自体は初期 DDL（`001_init.sql:1913-1936`）に加えて後段で `clinic_id`・`vaccination_id`・`exam_id`・`other_reason`・`created_by` が追加される（`001_init.sql:4033-4035`・`:4075-4079`・`:6168-6177`）。CSV の `clinic_id` は挿入されるが、trigger `trg_billing_items_sync_clinic_id` が INSERT/UPDATE 時に親 `billings.clinic_id` で必ず上書きする（`001_init.sql:6210-6237`）。`vaccination_id`/`exam_id`/`other_reason`/`created_by`/`source` は契約に無い列なので NULL・既定値のまま — **移行明細は vaccination/exam provenance を持たず、`source` は既定 `'manual'` になる**。
- 結果として、移行された `billing_items` 行の `tax_rate` は 5%/8% 時代の明細を含め一律 `0.10` で保存される。**以後にこの値を読む経路（明細の税額表示・再計算・帳票の集計ロジック）は保存値 0.10 を前提に動く**。なお旧 DB の source 行・旧側の実績値そのものはここでは消えない（消えるのは新 DB への輸送上の区別だけ）し、過去の保存済 `tax_amount`/`subtotal`/支払行は再計算されない — 「全帳票が 0.10 で再計算済み」や「source 値が恒久的に失われた」とは主張しない。
- MIG-44 の決定（`Jseikyu_Kin` claim gross、`TZei_Flg` 復号、日付境界）は producer 側 stage の決定で、consumer 側は本票の時点で `tax_rate` を受け取る経路を持たない。

### 1.6 verify

- `verifyCutoverRows` は行数一致と `clinic_id` 分離のみ検査（`cutover_import_copy.go:603-636`）。`tax_rate` の値検証は存在しない — 0.10 への均一化はどの検査でも発見されない設計。

### 1.7 まとめ（EMR-364 現行）

| 段 | 実際 |
| --- | --- |
| stage/CSV | `tax_rate` 列なし（10 列 header、stage にも source 列無し） |
| preflight | 列集合・placeholder・band・sha256 のみ。金銭値は未解析 |
| COPY/INSERT | 契約列のみ。`tax_rate` は送られない |
| DB | `DEFAULT 0.10` が全行に入る |
| verify | 行数・clinic 分離のみ。率の検証なし |
| 下流 | 読み取り・再計算経路は保存値 0.10 を見る |

補足: cutover とは別に、使い捨て DB 向けの legacy demo import 経路（`backend/internal/csvimport/import.go`）がある。こちらも `billing_items` を 9 列で写像し `tax_rate` を持たず（`import.go:201-202`）、`unit_price` は `strconv.ParseInt` なので小数セルは parse エラーで落ちる（`import.go:247-255`）。いずれの消費経路も履歴率・小数を輸送しない。

新しい 11 列 profile・version・pin・legacy bundle 扱いの提案は [DESIGN.md §8](../emr-342-medicines-contract-design/DESIGN.md)（PROPOSED/UNAPPROVED）を参照。本書では権威ある列定義を複写しない。

## 2. EMR-365: ワクチン単価の現行全経路

### 2.1 legacy DB（旧DB canonical）

- canonical `sale_items.unit_price_raw` は **numeric** で小数を保持（MIG-37 記録の DB 列）。リンクは `MST_YOB_INFO.YobTri_No` / `MST_YSS_INFO.YssTri_No` → `MST_SAL_INFO.PK_Sal_No`（MIG-31）。
- リンクされた単価は**税抜**として扱う既存決定（MIG-38 の税復号の逆転を反映）。**÷1.1 による整数化は行わない**（旧提案は撤回済み、MIG-31 最新コメント）。
- リンク解決の実測値は ticket 記録どおり: `TBL_TRI_DATA.Yob` 79 行全件 YOB マッチ（MIG-31）、`MST_YSS_INFO` の `YssTri_No` 非空白 95 件中 79 件マッチ（MIG-37）。本票では再計測していない。
- MIG-31 の価格ルール（**既存の決定**。本票では変更しない）: 旧税コード blank/1（external-tax）→ 旧単価そのまま。**0・負・NULL・空・リンク無し・未知税・2/3（exempt）は bare 空 CSV セル → 新 DB では NULL**（後述 §4.2）。正の小数は正のまま保持。

### 2.2 producer stage → CSV

- stage `vaccines` に `price` の source 写像は無い。`selectListSql("vaccines")` は `NULL::bigint AS price` を明示出力（`stage-csv-columns.mjs:610-614`）。
- `targetColumns("vaccines")` は 11 列で `price` を含む（`stage-csv-columns.mjs:440-442`）— **列は存在するが常に NULL で出る**。

### 2.3 CSV consumer → COPY → DB

- consumer spec も 11 列で `price` を含み、band 列は `id` のみ（`cutover_contract.go:263`）。
- COPY の NULL 規約: `NULL ''` オプションにより**空文字セル → SQL NULL**（`cutover_import_copy.go:198`）。producer が `NULL::bigint` を出すので、CSV の `price` セルは空 → 新 DB の `vaccines.price` は **NULL** になる。
- 金銭セルは preflight で数値解析されないため（§1.3）、仮に `"0.4"` のような小数が来れば preflight を通過し、COPY で `bigint` 変換エラー（SQLSTATE `22P02` 等）として sanitize されて落ちる — **現行は小数を輸送できない**。
- DDL: `vaccines.price bigint`（NULL 許容）（`backend/migrations/001_init.sql:720-735` 付近）。

### 2.4 Go model / JSON / API

- `model.Vaccine.Price *int64`（`backend/internal/model/vaccine.go:21`）、`price,omitempty`。
- 作成/更新 request: `Price *int64`（`backend/internal/medicalrecord/vaccine_request.go:21`・`:45`）。**JSON の `0.4` は int64 フィールドへの unmarshal で拒否される**（Go `encoding/json` は小数リテラルを整数型へ入れない）。数値の小数輸送経路は現行に存在しない。
- サービス検証: create/update とも `validateNonNegativePrice`（`backend/internal/medicalrecord/vaccine_service.go:109`・`:163` → `backend/internal/sharedkernel/validators.go:98`）で負は拒否、`nil` は通る。書き込みは直接の整数 write（`vaccine_service.go:54-55`）。
- 応答も `*int64`（`backend/internal/medicalrecord/vaccine_response.go:14`）。

### 2.5 マスタ UI

- 予防接種タブは治療プラン一覧の一タブ（`frontend/src/features/master/routes/treatment-plan-master-model.ts:17-23`・`:82-88`）。
- `transformVaccine`: `price: v.price ?? 0`（`frontend/src/lib/transforms/treatment.ts:80`）— **NULL と 0 が UI では区別されない**。
- 一覧の価格セル: `formatCurrencyOrDash`（`frontend/src/features/master/components/TreatmentPlanRows.tsx:179-185` → `frontend/src/lib/format/number.ts:38-41`）— `price<=0` は `-` 表示。
- 編集: `TreatmentItemSidePanel` の `MoneyInput`（`frontend/src/features/master/components/TreatmentItemSidePanel.tsx:155-160`）。`<input type="number">` で `e.target.valueAsNumber` を読む（`frontend/src/components/shared/SidePeek/MoneyInput.tsx:30-39`）— ブラウザ number input では小数も打てるが、`buildVaccineCreateRequest` が `price: data.price` を JSON 送信（`treatment-plan-master-model.ts:157-165`）し、BE が `*int64` で拒否するため**小数は保存できない**。`0` は空表示され NaN は 0 に潰れる（`MoneyInput.tsx:35-38`）。

### 2.6 接種イベント → 未請求 candidate

- `FindUnbilledVaccinationItemsByPetID` は `v.price AS unit_price` を取り、`vaccine_id/name/price` の NULL・負の単価を **unbillable として skip** し blocking warning `vaccination_master_unbillable` に数える（`backend/internal/billing/billing_item_repository_unbilled.go:17-108`、特に `:34`・`:83-91`）。
- 正常 candidate: `UnitPrice: *row.UnitPrice`（int64）、`Quantity: 1`、`TaxType: "excluded"`、`TaxRate: sharedkernel.DefaultTaxRate` = `0.10`（`sharedkernel/validators.go:15`）、`Source: "medical_record"`、`vaccination_id` 付き（同ファイル `:93-105`）。
- **現行 path では DB→model→candidate まで一貫して整数円であり、小数の保持はどの段でも無い**。unbillable warning があると精算完了をブロック（`billing_item_unbilled.go` 系の warning 処理）。

### 2.7 FE 会計画面（表示・ローカル再計算）

- 未請求詳細: `GET /v1/billing-items/unbilled-details`、ワクチン candidate は `vaccination_<id>` キー（`frontend/src/features/accounting/api/get-unbilled-items.ts:63-72`）。
- 変換: `unitPrice = item.unit_price ?? 0`、`taxRate = item.tax_rate ?? DEFAULT_STANDARD_TAX_RATE`（`frontend/src/features/accounting/api/transforms.ts:12-42`）— 欠落値は 0 / 0.10 に潰れる。
- 手入力明細: `parseInt(price, 10)` で整数化（`frontend/src/features/accounting/hooks/use-accounting-item-actions.ts:171`）— `"0.4"` は `0` になる。
- **FE の画面内再計算 `recomputeLineAmounts`（`use-accounting-item-actions.ts:63-76`）は BE と同一ではない**: FE は `Math.max(Math.round(unitPrice * quantity) - discountAmount, 0)` で**先に `単価×数量` を丸めてから割引を引き**、その丸め後の subtotal に税率を掛けて `Math.round` する。BE の合計側 `billing.CalculateBillingTotals` は subtotal については同じ式（`round(u*q)−discount` の 0 クランプ、`billing_service.go:37`）だが、**税額は `model.BillingItem.CalculateTaxAmount`（`model/accounting.go:162-171`）が独立に計算**し、課税ベースは `max(float64(u*q) − discount, 0)` の**未丸めの乗算結果−割引**。FE は税ベースを丸めた subtotal、BE は未丸めの `u*q−discount` から直接税額を作るため、同じ入力で税額がずれうる（§2.9 の匿名例）。
- 表示: `formatCurrency(item.unitPrice)`・小計/税/合計（`frontend/src/features/accounting/components/ItemListCard.tsx:342`・`:499-501`）。`formatCurrency` は `toLocaleString("ja-JP")`（`frontend/src/lib/format/number.ts:28-32`）。
- 確定時はサーバ側が明細作成→`CalculateBillingTotals` で正本を再計算する（§2.8）ため、画面の optimistic 値がそのまま保存値になるわけではない。

### 2.8 BE 明細作成 → 合計

- 完了経路: `createCompleteItems` → `CreateItemForComplete` → `newBillingItemFromCreateInput` → `applyVaccinationBillingItemCreate`（`backend/internal/billing/accounting_complete_tx.go:214-247`、`billing_item_service_create.go:220-247`・`:249-280`）。ワクチンリンク明細は `category=vaccine`、`source=medical_record`、`vaccination_id` 保持を強制される。
- 価格検証: create request の `unit_price int64` は `binding:"min=0"`（`backend/internal/billing/billing_item_request.go:40`）で非負整数のみ受理。小数・負は binding 時に拒否される。
- 税計算には 2 つの関数がある。**helper `billing.CalculateTaxAmount(unitPrice, quantity, taxType, taxRate)`（`billing_service.go:15-27`）は割引を引数に取らず `u*q*rate` を丸める**。実際の明細の税額は **`model.BillingItem.CalculateTaxAmount`（`model/accounting.go:162-171`）** が担当し、課税ベース `max(float64(u*q) − discount, 0)` に税率を掛けて `math.Round` → `int64`。helper と model の課税ベースは異なるので混同しない。
- 合計: `CalculateBillingTotals` は行 subtotal `round(u*q)−discount` を 0 未満にクランプして集計し、`totalAmount = subtotal + 外税sum`（内税は subtotal に内包）（`billing_service.go:33-46`）。**subtotal は「丸めた u*q − discount」、税額は「未丸め u*q − discount」の 2 系統が同居**している現行実装。
- 支払: `billingAmount = total - insurance - discount` 等、全て **int64 円**（`accounting_complete_tx.go:276` 付近）。未収は整数支払合計ベース（`backend/internal/billing/unpaid_amount.go:18-30`）。
- `payments`・`payment_splits` の CSV も整数円（producer `stage-csv-columns.mjs:560-582`、consumer `cutover_contract.go:258-259`）。

### 2.9 丸めモード・既存の FE/BE 差分・精度喪失点（現行の証拠）

- 丸めモード: Go `math.Round` は half away from zero、JS `Math.round` は half 方向 +∞（round half up）。**正の値では両者一致**、負の半値で符号方向にずれるが、比較対象の subtotal/課税ベースはいずれも `max(…, 0)` で負に落ちないため、この符号差は実害を持たない。丸め段階はいずれも「明細行ごと」。
- 既存の FE/BE 差分（例）: `unitPrice=11`、`quantity=0.5`、`discountAmount=1`、`taxRate=0.10`（excluded）。
  - FE: `u*q=5.5` → `round(5.5)=6` → subtotal `max(6−1,0)=5` → tax `round(5×0.1)=1`。
  - BE: subtotal は合計側で同じく `max(round(5.5)−1,0)=5`、ただし税額は `max(5.5−1,0)=4.5` → `round(4.5×0.1)=round(0.45)=0`。
  - **同じ入力で FE 税額 1 / BE 税額 0**。確定時の保存値は BE。これは現行実装の記録であって、どちらが正かの判断・修正権限は owner（本票は変更しない）。
- 精度喪失点（小数を仮に入れてもどこで潰れるか）:
  1. stage: `NULL::bigint AS price`（source 精度を取らない）— `stage-csv-columns.mjs:612`
  2. CSV: `''` → NULL。小数文字列は COPY で型エラー — `cutover_import_copy.go:198`
  3. DB: `price bigint` — `001_init.sql` vaccines DDL
  4. Go: `*int64`・JSON 小数拒否 — `model/vaccine.go:21`・`vaccine_request.go`
  5. FE: `transformVaccine` `?? 0`・`parseInt`・`MoneyInput` number input — `transforms/treatment.ts:80`・`use-accounting-item-actions.ts:171`・`MoneyInput.tsx`
  6. 計算: float64 → `math.Round` int64（10 進小数を正確に扱う decimal 型は無い）— `billing_service.go`・`model/accounting.go`
  7. unbilled: `UnitPrice int64`・Quantity 1・0.10 固定 — `billing_item_repository_unbilled.go`

## 3. 提案の参照先（本書は現行経路の記録のみ）

- 履歴税率 `tax_rate` の新列・列挙・version/pin・互換性・検証ケースの提案: [DESIGN.md §8](../emr-342-medicines-contract-design/DESIGN.md)（PROPOSED/UNAPPROVED）。
- ワクチン単価の保存精度・輸送・表示・計算・匿名固定ケースの提案: [DESIGN.md §9](../emr-342-medicines-contract-design/DESIGN.md)（PROPOSED/UNAPPROVED）。
- 両者とも未承認であり、本書の現行記録をもって採用・実装・profile 受理の根拠にしない。

## 4. 保留の owner 決定・receipt（両票共通）

- **EMR-364**: version 名・ordered 11 列・nullability・rate enum・pin（producer source 全ファイル SHA-256）の同一 artifact 合意。legacy 10 列 bundle の扱い。medicines-only profile との採用順序。MIG-15 treatments 22 表案は本票では未承認。
- **EMR-365**: 保存型・scale・表示桁・JSON 表現（decimal 文字列）・税/合計の丸め順・数量の扱い・旧整数 bundle 互換の同一 artifact 合意。小数対応の採用自体が未承認。FE/BE の既存の丸め差分（§2.9）の是正有無・正規規則も owner 決定。
- **共通**: 実装 GO・DB migration・STG/F6・PR merge・deploy は別承認・別 session。本票は調査・提案文書のみ。
- producer 側 owner: MIG-44（税・率・総額 source）、MIG-31（ワクチン価格写像）、MIG-37（台帳/pin）、MIG-41（単一実 CSV 生成）、EMR-349 系（受領 import）。
