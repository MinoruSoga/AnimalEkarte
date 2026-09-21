# UAT-Q2-TREATMENTS-IMPORT: 処置・注射・処方の移行契約差分

状態: **全種類・全期間の移行契約設計 READY／実装・取込未実施**。2026-09-19に依頼者が「マスタ/患者別処置履歴/対象期間」の問いへ「全部」と回答した。処置名・料金マスタと患者ごとの処置履歴を今期対象とし、任意の年数/種類で減らさない。2026-09-22: 下記「契約ドラフト（未解決フィールド）」「合成 fixture 設計」「理由コード付き保留ルール」を repo 根拠のみで追記（21 表 hash / producer / DB / 実データは未変更）。要件・状態の正本は [TODO](../../../todo-issue.md#uat-q2-treatments-import)、以前の判断経緯は [医院 UAT 記録](../stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補)。

## 現行契約と不足分

| 対象 | 旧側で確認できる根拠 | 現行の送出・取込先 | 未確定の差分 |
| --- | --- | --- | --- |
| 処置・注射のマスタ候補 | `../old_db/sql/table_structure_sqlserver.sql:517-539` の `MST_SAL_INFO` → `020_canonical.sql:1715-1755` `legacy_canonical.sale_items` → `030_stage.sql:239-270` が `sal_tri_kbn='1'` のみ `animalekarte_stage.procedures` へ。`legacy_sale_no` / `source_pk` を系譜保持。 | `backend/internal/csvimport/cutover_contract.go:205` の `procedures`（21 表内）。列 `id, clinic_id, name, price, is_active, description, duration, anesthesia, parent_id, tax_type, tax_rate, sort_order, is_surgery`。 | 注射を含める旧種類・コード範囲、旧マスタキーと新 ID の確定 crosswalk、採用/除外ルールは **UNKNOWN**。`SalTanka_Num` は canonical にあるが stage `price` は `NULL` 固定。現行マスタ送出は患者実施履歴ではない。 |
| 来院ごとの処置・注射実施 | `table_structure_sqlserver.sql:1538-1573` の `TBL_TRI_DATA`（PK `(TPK_PET_No, TReat_Sno, TGyo_Num)`）→ `020_canonical.sql:1911-1980` は全行を `billing_items` に正規化。`030_stage.sql` に `treatments` CREATE/INSERT は **無い**。 | `cutover_contract.go:203-227` に `treatments` 表仕様は **無い**（21 表外）。AE 実行 DDL は `backend/migrations/001_init.sql:1587-1620` に `treatments` があるが cutover 非契約。 | 臨床実施として別表へ出す契約は未成立。会計 `billing_items` から実施行を捏造しない。候補列・FK・重複キーは下表（未採用は UNKNOWN）。 |
| 処方履歴 | `TBL_YBS_HIST`（`020_canonical.sql:3637+`）は予防接種処方箋ストリーム。一般処方薬の独立 producer INSERT は `030_stage.sql` に無い。`1660` の `prescriptions` 言及は YBS lot 空文字コメントのみ。 | `cutover_contract.go:203-227` に `prescriptions` **無し**。AE DDL `001_init.sql:1436-1452` にヘッダ表のみ（明細子表なし）。 | 一般処方の旧表・元列、薬剤マスタ、用量/用法/日数、発行日時、FK、重複キーは **UNKNOWN**。YBS を一般処方へ流用しない。 |
| 会計と過去カルテ表示 | `030_stage.sql:1053-1108` / `020_canonical.sql:1795-1835` は会計明細に `billing_id` 必須。`TReat_Sno` は精算シリアル（§31）でありカルテ番号ではない。 | 21 表に `billings` / `billing_items`（`cutover_contract.go:215-216`）。 | 会計を再計上せず臨床履歴としてどう表示するか、欠損と未移行の区別、読み取り専用表示は **UNKNOWN**。`billing_items` 取込完了 ≠ 処置/処方履歴完了。 |

`030_stage.sql:1072-1108` の会計明細は `unit_price_raw` を数値判定して丸め、`quantity_raw` を正数判定して小数第 1 位へ丸め、それ以外は既定値にする。これは**現行の会計変換**であり、臨床履歴の元数量・元金額を保存する方針ではない。`procedures` 側の価格は `030_stage.sql:239-270` で `NULL`。会計金額と臨床実施値の対応・差額照合は別途決める必要がある。

## 旧列 → producer → AE（repo 根拠。不明列は UNKNOWN）

根拠ファイル（読取のみ）:

| 層 | パス |
| --- | --- |
| 旧物理 DDL | `../old_db/sql/table_structure_sqlserver.sql`（`MST_SAL_INFO`, `TBL_TRI_DATA`, `TBL_YBS_HIST`） |
| canonical | `../old_db/sql/migration/020_canonical.sql`（`sale_items`, `billings`, `billing_items`） |
| stage producer | `../old_db/sql/migration/030_stage.sql` |
| AE 21 表契約 | `backend/internal/csvimport/cutover_contract.go`（`CutoverTableSpecs` / placeholder / hash 定数） |
| AE 実行 schema（非 cutover） | `backend/migrations/001_init.sql`（`treatments`, `prescriptions`, `medicines`） |
| 参考（非現行 producer） | `backend/cmd/_archive/seed-old-db/transform.go` — 採用根拠にしない。候補メモのみ |

21 表 hash は観測のみ（**変更しない**）: `cutoverStageMappingSHA256=0d7f089990079af28c2ba454ee48188d55a67beab19a9f16a86dee93fec80597`、`cutoverCSVContractSHA256=19b2c5c270058b20c1fa816679c0430f236a0a181165ae1ed2257c64c83f6671`（`cutover_contract.go:17-18`）。`CutoverTableSpecs` は 21 表で `treatments` / `prescriptions` / `medicines` を含まない（観測: staffs…vaccinations）。

過去分の既定: **参照専用・自動再請求なし**（`todo-issue.md#uat-q2-treatments-import`）。`billing_items` 取込完了を処置/処方履歴完了とみなさない。氏名一致結合はしない。

### procedures（21 表に存在。`CutoverTableSpecs` `cutover_contract.go:205`）

Producer: `animalekarte_stage.procedures` INSERT `030_stage.sql:239-269`。FROM `legacy_canonical.sale_items` WHERE `sal_tri_kbn = '1'`（`246-247`）。ルータコメント `209-216` / canonical `1710-1714`: `SalTri_Kbn` のみ。`1=procedure`, `2=merchandise`。未分類は canonical のみ。注射専用コード範囲は **UNKNOWN**。`sal_tri_kbn='1'` を全処置/全注射としない。

#### 物理 → canonical（マスタ）

| 旧 `MST_SAL_INFO`（`table_structure_sqlserver.sql:517-539`） | `legacy_canonical.sale_items`（`020_canonical.sql:1715-1755`） | 注 |
| --- | --- | --- |
| `PK_Sal_No` | `legacy_sale_no` / `source_pk` | 唯一承認の item crosswalk キー（canonical コメント） |
| （生成） | `sale_item_key` IDENTITY | stage/AE `id` の元 |
| `SalTri_Data` | `primary_name` | 空名は stage で `needs_review` |
| `SalTriNm_Data` / `SalTriNm2_Data` | `alternate_name` / `alternate_name2` | description 折り畳み |
| `SalTri_Kbn` | `sal_tri_kbn` | `'1'`→procedures / `'2'`→merchandise |
| `SalTanka_Num` | `unit_price_raw` | **stage procedures の `price` には載せない**（下表） |
| `SalRv_Kbn` | `sal_rv_kbn` | merchandise 側税区分に出る。procedures INSERT には無い |
| `SalTri_KbnNm` | `category_label` | |
| `SalZkei_Data` / `SalTouyo_Data` | `dosage_form` / `administration` | 剤形/投与。臨床用量ではない |
| `SalBiko_Data` | `remarks` | |
| `SalSotUch_Flg` / `SalGyoSei_Flg` | `outsourced_flg` / `insurance_flg` | procedures INSERT 未使用 |
| `SalSdtlNo` | `sort_no` | |
| `SalGenka_Num`, `SalPoint_Num`, `SalNbr_Num`, `Salnbk_Num`, `SalBar_Data`, `SalSmy_Kbn`, `SalUri_No` | canonical INSERT に **無い** | 用途 **UNKNOWN**（推定しない） |

#### stage → AE 21 表

| 旧 / canonical | producer `animalekarte_stage.procedures` | AE 21 表 `procedures` | FK / dedup | 注 |
| --- | --- | --- | --- | --- |
| `sale_item_key` | `id = sale_item_key + :stage_id_offset` | `id`（band `id`,`parent_id`） | dedup 候補: `legacy_sale_no` UNIQUE / `PK_Sal_No`。AE 側は band 済み `id` | crosswalk 候補。確定採用/除外は **UNKNOWN** |
| （定数 `1`。CSV で `{{CLINIC_ID}}`） | `clinic_id` | `clinic_id`（placeholder `cutover_contract.go:175`） | FK → `clinics` | 旧支店コードは clinic_id に使わない |
| `primary_name`（重複時 `legacy_sale_no` 接尾） | `name` | `name`（force-not-null） | 名前衝突は接尾で回避 | 空名は `''` + `needs_review`。CSV 除外はコメントのみ |
| `unit_price_raw`（`SalTanka_Num`） | `price = NULL`（INSERT リテラル） | `price` | — | マスタ料金の現行送出値は NULL。原価列の採用は **UNKNOWN** |
| なし（`true`） | `is_active` | `is_active` | — | 旧列なし |
| `remarks` + 別名 + 分類 + 剤形 | `description`（改行連結） | `description`（force-not-null） | — | `administration`（`SalTouyo_Data`）は description 連結に **含まれない**（030 INSERT 未参照） |
| なし（`NULL`） | `duration` | `duration` | — | **UNKNOWN** |
| なし（`'none'`） | `anesthesia` | `anesthesia` | — | **UNKNOWN** |
| なし（`NULL`） | `parent_id` | `parent_id`（self FK band） | FK → `procedures.id` | 旧親子列 **UNKNOWN** |
| なし（`'excluded'` / `0.10`） | `tax_type`, `tax_rate` | `tax_type`, `tax_rate` | — | `sal_rv_kbn` は procedures INSERT に無い |
| `sort_no` | `sort_order = coalesce(sort_no, 0)` | `sort_order` | — | |
| なし（`false`） | `is_surgery` | `is_surgery` | — | **UNKNOWN** |
| `legacy_sale_no` | `legacy_item_code` | 21 表 CSV 列に無い | 系譜 | 非契約 |
| `'MST_SAL_INFO'`, `source_pk` | `source_table`, `source_pk` | 21 表 CSV 列に無い | 系譜 | |

### treatments（21 表に無い。`cutover_contract.go:203-227` に表名なし）

`030_stage.sql` に `CREATE TABLE animalekarte_stage.treatments` / `INSERT INTO …treatments` は **ABSENT**。現行 producer は `TBL_TRI_DATA` 全行を会計 `billing_items` へ送るのみ（`020_canonical.sql:1795-1799`, `030_stage.sql:1072-1108`）。

AE 実行表（cutover 非契約）: `001_init.sql:1587-1620` / model `backend/internal/model/treatment.go:30-60`。

#### 候補: 物理 TRI →（未実装 producer）→ AE treatments

| 旧 `TBL_TRI_DATA`（`table_structure_sqlserver.sql:1538-1573`） | canonical 上の位置（現行は billing 系） | AE `treatments`（DDL・非 cutover） | FK / dedup | 契約状態 |
| --- | --- | --- | --- | --- |
| PK `(TPK_PET_No, TReat_Sno, TGyo_Num)` | `billing_items.source_pk = pet\|sno\|line`（`020_canonical.sql:1979`） | `id`（新採番） | **dedup 候補** 旧複合 PK / `source_pk`。会計行と同一キー共有可否は **UNKNOWN** | producer ABSENT |
| `TPK_PET_No` | `legacy_pet_no` → `xw_pet` | （直接列なし。親 `medical_records.pet_id`） | FK 候補: pet crosswalk。氏名結合禁止 | 経路はカルテ解決依存 |
| `TReat_Sno` | 精算シリアル（= `KNJO.JKnjo_Sno`）。**カルテ番号ではない**（`020_canonical.sql:1801-1807` §31） | — | カルテ FK に直結させない | 誤結合禁止は確定 |
| （KNJO `JKrt_No` via settlement） | `billings.record_key` | `medical_record_id` NOT NULL | FK → `medical_records` | 解決不能時は保留。黙って落とさない |
| `TChri_No` | `sale_no` → `sale_item_key`（jouto 特殊コードは NULL） | `procedure_id` / `medicine_id`（item_type 依存） | FK → `procedures` / `medicines` | `medicines` は 21 表外。マスタ未移行時 **UNKNOWN**/保留 |
| `Ttri_Kbn` / `TRv_Kbn` | canonical `category_kbn`（並列分類。会計 category の正本は `SalTri_Kbn`） | `item_type`（`procedure`/`medicine`/`consultation`/`other`） | CHECK `chk_treatment_item_ref` | 値域マッピングは **UNKNOWN**。archive `seed-old-db` は `'1'→procedure,'2'→medicine` 提案のみで現行非採用 |
| `TreatNm_Data` / `Treat_Data` / `TbunNm_Data` | `item_name` / raw 列 | `content` | — | coalesce 規則は会計側と同一候補 |
| `Ttanka_Kin` | `unit_price_raw`（text 保持） | `unit_price` bigint | — | 会計丸め（stage round）を臨床へ流用しない。原精度保持方針は未契約 |
| `Tsryo_Num` | `quantity_raw` | `quantity` numeric | CHECK `quantity > 0` | 同上 |
| `Tnebi_Kin` | `discount_raw` | `discount_amount` | — | stage billing_items は割引列を **送出しない** |
| `Tkin_Kin` | `line_amount_raw` | （AE treatments に金額合計列なし） | — | 表示/監査用の保持先 **UNKNOWN** |
| `Ttreat_Date` | header `billings.treat_date` = min | （treatments に実施日列なし。親カルテ `date`） | — | 診療日/実施日/会計日の混同禁止。行単位日時の置き場 **UNKNOWN** |
| `TTouyo_Data` | （billing_items INSERT 未送出） | `admin_route` 候補 | — | 対応は **UNKNOWN** |
| `TRyou_Num` / `Tkaisu_Num` | （未送出） | （dose_* は medicine 用スナップショット） | — | 処方用量との関係 **UNKNOWN**。黙って dose に入れない |
| `TGyoSei_Flg` | `insurance_flg` | `is_insurance` | — | `'1'/'01'` 真の規則は会計側と揃える候補 |
| `TZei_Flg` | `tax_flg` | （treatments に tax 列なし） | — | 臨床表へ載せるか **UNKNOWN** |
| `TSeisan_Flg` | header `settle_flg` | — | — | 取消/精算状態の臨床反映 **UNKNOWN** |
| `Tgenka_Num`, `Tpoint_Num`, `Tnbirt_Num`, `Turi_*`, `TSet_*`, `TGrp_No`, `TZaikei_Data` 等 | 一部未 canonical 化 | — | — | **UNKNOWN**（列を発明して埋めない） |
| clinic | 旧支店コードあり得るが clinic は定数置換 | treatments は自前 `clinic_id` なし（親カルテ経由） | — | clinic は親/マスタ側で検証 |

| 旧 | producer | AE 21 表 | 注 |
| --- | --- | --- | --- |
| 臨床実施の送出 | **ABSENT** | **ABSENT**（cutover） | 会計行を treatments 行へコピーしない。新契約+hash 変更は別ゲート |
| 重複キー | **UNKNOWN**（候補: `TPK_PET_No\|TReat_Sno\|TGyo_Num`） | — | 再取込で増えないこと |
| 参照欠損 vs 未移行 | 判定手続きは下記「理由コード付き保留ルール」 | — | コード自体はドラフト。値域写像など個別 UNKNOWN は残す |
| 自動再請求 | 禁止（既定） | — | 過去分 read-only。`billing_items.treatment_id`（DDL にあるが CSV 非契約）を埋めない |

### prescriptions（21 表に無い。`cutover_contract.go:203-227` に表名なし）

`030_stage.sql` に一般処方の `CREATE`/`INSERT` は **ABSENT**。`1660` は YBS 接種 INSERT の lot 空文字コメント（`-- YBS has no lot columns (prescriptions)`）であり処方送出ではない。

AE DDL `001_init.sql:1436-1452`: `id, clinic_id, owner_id, pet_id, medical_record_id, prescribed_at, duration_days`（ヘッダのみ。明細子表なし）。`medicines` マスタは `001_init.sql:740+` だが 21 表外。

| 旧 | producer | AE（DDL / 21 表） | FK / dedup | 注 |
| --- | --- | --- | --- | --- |
| 一般処方の元表・元列 | **ABSENT** / **UNKNOWN** | 21 表 **ABSENT**；DDL ヘッダのみ | — | 復元根拠のない用量を作らない |
| `TBL_YBS_HIST`（予防処方箋） | vaccinations_ybs → stage vaccinations | `vaccinations`（21 表） | 別ユニット | 一般 `prescriptions` へ流用しない（canonical コメント） |
| `TBL_TRI_DATA.TRyou_Num` / `Tkaisu_Num` / `TTouyo_Data` | billing 経路で未送出 | prescriptions / treatments.dose_* | — | 処方ヘッダ/明細との対応 **UNKNOWN** |
| 薬剤マスタ | `MST_SAL_INFO` の一部? / 別表? | `medicines`（非 cutover） | FK **UNKNOWN** | 採用範囲未確定 |
| `prescribed_at` / `duration_days` | **UNKNOWN** | DDL 必須/既定 | — | YBS の start/end を流用しない |
| clinic / owner / pet / record FK | **UNKNOWN** | DDL FK あり | 氏名結合禁止 | |
| 重複キー | **UNKNOWN** | — | — | |

### billing_items（21 表に存在。会計明細。`cutover_contract.go:216`）

Producer: `animalekarte_stage.billing_items` INSERT `030_stage.sql:1072-1108`。FROM `legacy_canonical.billing_items`。`source_table` リテラル `'TBL_TRI_DATA'`。stage CREATE に `clinic_id` 列は無い（`1053-1071`）。

#### 物理 → canonical（会計明細）

| 旧 `TBL_TRI_DATA` | `legacy_canonical.billing_items`（`020_canonical.sql:1911-1980`） | 注 |
| --- | --- | --- |
| `TPK_PET_No` | `legacy_pet_no` | |
| `TReat_Sno` | `legacy_treat_sno` | 精算シリアル |
| `TGyo_Num` | `legacy_line_no` | |
| （join） | `billing_key` NOT NULL | 親 `billings` UNIQUE `(legacy_pet_no, legacy_treat_sno)` |
| coalesce(TreatNm, Treat, Tbun) | `item_name` | |
| `Ttanka_Kin` / `Tsryo_Num` / `Tnebi_Kin` / `Tkin_Kin` | `*_raw` text | 符号・NULL 保持（canonical） |
| `TZei_Flg` | `tax_flg` | |
| `TRv_Kbn`/`Ttri_Kbn` | `category_kbn` | 会計 category 正本ではない |
| `TGyoSei_Flg` | `insurance_flg` | |
| `TChri_No` | `sale_no` → `sale_item_key` / `sal_tri_kbn` | unmatched は NULL/`other` |
| 複合 | `source_pk = pet\|sno\|line` | **現行会計 dedup/系譜キー** |

#### stage → AE 21 表

| 旧（canonical） | producer `animalekarte_stage.billing_items` | AE 21 表 `billing_items` | FK / dedup | 注 |
| --- | --- | --- | --- | --- |
| `item_key` | `id = item_key + :stage_id_offset` | `id`（band `id`,`billing_id`） | dedup: `source_pk` | |
| `billing_key` | `billing_id` NOT NULL | `billing_id` | FK → `billings` | |
| なし（stage 列なし） | **UNKNOWN** in INSERT | `clinic_id`（placeholder `cutover_contract.go:186`） | — | CSV 置換経路はこの INSERT に無い |
| `sal_tri_kbn` | `category`: `'1'`→`procedure`、`'2'`/else→`other` | `category` | — | 注射専用値 **UNKNOWN** |
| `item_name` | `name` | `name`（force-not-null） | — | |
| `unit_price_raw` | round→bigint / else `0` | `unit_price` | — | **会計変換**。臨床原値ではない |
| `quantity_raw` | round 1 桁 / else `1` | `quantity` | — | 臨床原値ではない |
| `tax_flg` | excluded/included / needs_review | `tax_type` | — | |
| `insurance_flg` | `is_insurance_applicable` | `is_insurance_applicable` | — | |
| `legacy_line_no` | `sort_order` | `sort_order` | — | |
| `source_pk`, `legacy_pet_no`, `legacy_treat_sno` | 同左（`legacy_record_no` 名で treat_sno） | 21 表 CSV 列に無い | 系譜 | 臨床実施キーとしては未契約 |
| `discount_raw` | stage 未送出 | AE DDL に `discount_*` あり、**CSV 非契約** | — | |
| AE `treatment_id` 等 | stage 未送出 | DDL にあるが `cutover_contract.go:216` 列に無い | FK → treatments | 過去分で埋めない（再請求導線を作らない） |

親 `billings`（会計ヘッダ。臨床実施表ではない）: stage `030_stage.sql:908-949`。`scheduled_date` は `treat_date::date`（1900-01-01 超のみ）。AE 列 `id, clinic_id, medical_record_id, owner_id, pet_id, total_amount, status, scheduled_date, completed_at`（`cutover_contract.go:215`）。dedup: `(legacy_pet_no, legacy_treat_sno)`。`legacy_treat_sno` はヘッダ系譜であり treatments 行キーではない。

見積明細 `estimate_items`（`030_stage.sql:1385-1427`、`TBL_MIT_DATA`）は `sal_tri_kbn='1'` で `procedure_id` を付けるが、患者実施履歴ではない。本票の treatments 契約に流用しない。

### FK / dedup 要約（確定 vs UNKNOWN）

| 対象 | 確定できるキー / FK | UNKNOWN / 禁止 |
| --- | --- | --- |
| procedures | 旧 `PK_Sal_No` UNIQUE → `legacy_sale_no`；AE `id` band；`clinic_id` placeholder；`parent_id` self | 注射範囲、price 採用、parent 旧列、採用/除外 |
| billings | `(TPK_PET_No, TReat_Sno)` UNIQUE；`medical_record_id` は KNJO `JKrt_No` 経由 | `TReat_Sno` をカルテ番号とみなすこと（禁止） |
| billing_items | `source_pk = pet\|sno\|line`；`billing_id` 必須；category は `SalTri_Kbn` | 会計丸め値の臨床流用；`treatment_id` 埋め |
| treatments（将来） | 候補 dedup 旧複合 PK；候補 FK pet/record via §31、procedure via `TChri_No`；missing vs not-yet-migrated の理由コード草案あり | producer 自体、item_type 写像、実施日列、dose、会計との二重計上、AE 正規化承認値 |
| prescriptions（将来） | AE DDL FK 形状のみ観測；`HOLD_SOURCE_ABSENT_PRESCRIPTION` で作成しない | 旧ソース表、明細、用量、YBS 流用 |

## 確定した範囲と設計担当の調査項目

| 判断 | 現状 |
| --- | --- |
| 業務目的、要件責任者、受入責任者 | 目的は過去診療を欠落なく参照すること。出所は9月19日の依頼者回答。個人名を伴う要件/受入参照は契約変更前に実行票へ記録する。 |
| 採用する種類と用途 | 処置マスタ＋患者ごとの処置履歴を全種類。注射・薬剤/処方等の関連原本も棚卸しし、含めない関連行があれば理由と判断を残す。過去分は参照専用/自動再請求なしを設計の安全側既定とする。 |
| 対象期間と日付基準 | 提供原本の最古から最終抽出まで全部。種類/年数の任意filterを設けない。診療日/実施日/会計日/処方日は意味ごとに保持し、日付不明行を黙って除外しない。 |
| 元列、変換、FK/crosswalk、重複キー | 本票で物理→canonical→stage→AE を埋めた。未実装の treatments/prescriptions は候補と UNKNOWN を分離。氏名一致で結合しない。 |
| 元数量・単位・金額、処方用量/用法/日数、訂正と復旧 | 会計丸めを臨床へ流用しない（確定方針）。具体写像値は **UNKNOWN**。 |
| 合格条件 | 下記の原本全行照合、参照/原値保持、重複なし、読み取り専用表示、再請求なし。数量/金額の無承認丸めは不可。 |

## 次に作る成果物と順序

1. producer担当は旧schema/canonical定義と `030_stage.sql` から、`source table/column / 種類 / 旧clinic-pet-record-row key / 日時の意味 / 元数量・単位・単価・金額 / 取消・訂正状態 / 根拠` を一覧化する。原本全種類の分類値と件数照合方法を定義し、既存 `sal_tri_kbn='1'` だけで「全処置」としない。会計行しかないものを臨床実施として捏造しない。
2. AE担当は同じ行へ `producer出力 / consumer field / FK / 重複キー / 欠損時の保留理由 / 表示先 / audit・復旧` を追加。旧値と正規化値を分け、原本から確定できない単位/用量/実施日を既定値で埋めない。既存21表のマスタ/会計と、新しい履歴契約の差分を明示する。
3. 両担当は医院・種類・年別に `原本行数 = 取込行数 + 理由付き保留行数 + 根拠付き重複除外行数` を検証できる合成fixtureと期待値を作る。原本の存在する金額/数量は原精度で照合し、変換差を個別説明する。保留を隠して全件移行済みとしない。
4. 契約案・合成期待値・バージョン移行/互換性・復旧方式をレビューしてから、producerとconsumerを同じ契約で実装する。DB schema変更やcodegenが必要なら人間実行ゲートへ分離。契約案レビュー前に21表/hash/生成物を変更しない。

設計単位の完了は、全種類が対応表にあり、実装可能な列と原本不足で保留する列、テスト期待値、担当範囲が区別できたこと。原本の追加提供が必要な行は根拠付きでまとめて依頼し、既回答の「全部か/何年か」は聞き直さない。

## 合成 fixture と照合案（概要・未実施）

詳細は下記「合成 fixture 設計」へ。概要のみ:

1. 同一医院・同一ペット・同一旧カルテに、処置と注射のマスタ参照、実施行、処方行、会計行を持つ合成例を用意する。他医院・他ペットの旧ID衝突例も入れ、相互参照しないことを検査する。
2. 元数量・単位・元金額と変換後値を別列で照合し、丸め・単位換算・欠損値の扱いを個別に承認する。会計明細との金額差も記録する。
3. 同じ入力を二度取り込み、旧行単位の重複キーにより件数と参照が増えないことを確認する。参照先が存在しないケースと、その参照先がまだ移行されていないケースを分けて停止/保留結果を照合する。
4. 臨床履歴の表示で過去カルテ詳細へ到達し、元来院・処置・注射・処方が読み取り専用で表示されることを確認する。履歴の表示が新たな `billing_items` や請求を生成しないことを検査する。
5. 訂正・復旧ケースで旧行と新行の対応、件数・金額、監査手段を照合する。承認された disposable rehearsal の保存→再読込→表示はこの後の別ゲートとする。

---

## 契約ドラフト（未解決フィールド — レビュー用）

目的: 既存マッピングを作り直さず、**treatments 履歴契約で未決の FK / 分類 / 日時 / dose / price / dedup** だけをレビュー可能な案にする。確定済み禁止事項（氏名結合禁止、`TReat_Sno`≠カルテ番号、会計丸めの臨床流用禁止、自動再請求なし、YBS→一般処方流用禁止、会計行の臨床捏造禁止）は維持する。臨床値・件数・医院固有コードは発明しない。`cutoverStageMappingSHA256` / `cutoverCSVContractSHA256` と 21 表 producer は観測のみ（変更しない）。

### D0. 契約境界（確定観測）

| 項目 | 観測 | 契約含意 |
| --- | --- | --- |
| `treatments` / `prescriptions` / `medicines` | `CutoverTableSpecs`（`cutover_contract.go:203-225`）に **表名なし** | 臨床履歴は **新契約候補**。現行 21 表 cutover 完了 ≠ treatments 完了 |
| stage producer | `030_stage.sql` に `animalekarte_stage.treatments` CREATE/INSERT **ABSENT** | 実装前に producer+consumer 同一契約レビューが必須 |
| AE 実行 DDL | `001_init.sql:1587-1616` `treatments`；dose 列は同ファイル後半 `#201` 追加 | cutover 非契約。DDL 形状は候補列の参照先 |
| 会計経路 | `TBL_TRI_DATA` → canonical `billing_items` → stage `billing_items`（`030_stage.sql:1072-1108`） | 会計は並行維持。臨床行へコピーしない |
| hash（観測） | `0d7f0899…` / `19b2c5c2…`（`cutover_contract.go:17-18`） | 本ドラフトで変更しない |

### D1. FK（attribution / master refs）

| フィールド | 解決経路（repo） | ドラフト規則 | 未決 / UNKNOWN |
| --- | --- | --- | --- |
| 医院帰属 | AE `treatments` に `clinic_id` なし。親 `medical_records.clinic_id` とマスタ `procedures.clinic_id`（placeholder `{{CLINIC_ID}}`）で検証 | 取込バンドルの clinic 定数と親カルテ clinic が一致しない行は投入しない。旧支店コードを clinic_id に使わない | 複数支店原本を 1 clinic に畳む規則は **UNKNOWN**（別 clinic 決定票） |
| pet 帰属 | `TPK_PET_No` → `xw_pet` → 親カルテ `pet_id` | 氏名/カナ一致結合は禁止。pet crosswalk 欠落は保留 | — |
| `medical_record_id` | §31: `TReat_Sno` = 精算シリアル（`KNJO.JKnjo_Sno`）。カルテは `KNJO.JKrt_No` 経由（`020_canonical.sql:1801-1806,1837-1843`） | `TReat_Sno` をカルテ番号として直結しない。解決不能は黙って落とさず保留 | 1 settlement に複数 karte 候補が残る場合の採用規則は **UNKNOWN**（canonical は distinct 1 件のみ昇格） |
| `procedure_id` | `TChri_No` → `sale_no` → `sale_item_key`；`sal_tri_kbn='1'` かつ procedures 21 表へ載るもの | マスタ行が同一 clinic の procedures に存在するときのみ FK 設定 | 注射を procedures に含める旧コード範囲・採用/除外は **UNKNOWN** |
| `medicine_id` | `medicines` は 21 表外（`001_init.sql:740+`） | 現状 FK を埋めない（保留）。YBS/vaccinations を流用しない | 薬剤マスタ元表・crosswalk 全体が **UNKNOWN** |
| `consultation_id` / `inventory_id` | 旧 TRI に対応列なし | 過去分は常に NULL | — |
| 会計 `billing_items.treatment_id` | DDL にあるが CSV 非契約（`cutover_contract.go:216`） | 過去分では埋めない（再請求導線を作らない） | 将来の双方向リンク要否は **UNKNOWN** |

### D2. 分類（`item_type` / category）

| 入力 | 現行会計での扱い | 臨床ドラフト | 未決 / UNKNOWN |
| --- | --- | --- | --- |
| `MST_SAL_INFO.SalTri_Kbn` | `'1'`→procedures / billing `procedure`；`'2'`→merchandise / billing `other`（`030_stage.sql:246-247,1079-1083`） | 会計 category の正本候補。臨床 `item_type` への機械写像は未承認 | 注射専用区分の有無 **UNKNOWN** |
| `Ttri_Kbn` / `TRv_Kbn` | canonical `category_kbn`（並列。会計正本ではない） | 監査・保留理由の根拠列として保持。`item_type` 既定値にしない | 値域→`procedure`/`medicine`/`consultation`/`other` 写像は **UNKNOWN**。archive `seed-old-db` の `'1'→procedure,'2'→medicine` は現行非採用 |
| unmatched `TChri_No` | `sale_item_key` NULL / `sal_tri_kbn` NULL → billing `other` | 臨床は `item_type='other'` 候補 + マスタ FK NULL。名称は raw `content` | 「other」確定か保留かは **UNKNOWN**（件数根拠待ち） |
| `chk_treatment_item_ref` | DDL: item_type と FK の組合せ制約 | 契約は制約を破る行を作らない。未解決 FK は type=`other` または保留行へ | 保留行の物理格納（本表 vs hold 表）は実装ゲート |

### D3. 日時

| 旧 / 中間 | 意味 | ドラフト | 未決 / UNKNOWN |
| --- | --- | --- | --- |
| `Ttreat_Date` | 行の実施/処置日時候補（物理 `table_structure_sqlserver.sql:1542`） | 行単位の実施日時として系譜保持。診療日・会計日と混同しない | AE `treatments` に実施日列なし。置き場（親 `medical_records.date` のみ / 別列 / memo系譜）は **UNKNOWN** |
| header `billings.treat_date` | 同一 `(pet,sno)` の min `Ttreat_Date` | 会計ヘッダ用。臨床行日時の代替にしない | — |
| `medical_records.date` | カルテ日付（予約/来院側） | 親カルテ帰属の日付。TRI 行日時の代用に黙って使わない | 行日時欠損時に親日付へフォールバックするかは **UNKNOWN**（フォールバックするなら理由コード必須） |
| `prescribed_at` / `duration_days` | AE `prescriptions` DDL 必須/既定 | 一般処方ソース未確定のため埋めない。YBS start/end 流用禁止 | 処方日時の元列 **UNKNOWN** |

### D4. dose / 用法

| 旧列 | AE 候補 | ドラフト | 未決 / UNKNOWN |
| --- | --- | --- | --- |
| `TTouyo_Data` | `admin_route`（varchar） | 対応は未証明。空文字既定で埋めない | 写像規則 **UNKNOWN**。billing 経路は未送出 |
| `TRyou_Num` / `Tkaisu_Num` | `dose_*`（`dose_weight_kg` / `dose_amount_mg` 等） | `#201` dose 列は **medicine の体重あたり自動計算スナップショット**（`treatment.go:50-56`）。旧用量文字列を黙って dose に入れない | 処方用量との関係・単位正規化は **UNKNOWN** |
| `MST_SAL_INFO.SalTouyo_Data` / `SalZkei_Data` | procedures `description` 連結の一部候補 | stage procedures INSERT は `administration` を **未参照**（本票マスタ節）。臨床用量ではない | — |

### D5. price / qty / 金額保持

| 値 | 会計 stage（現行） | 臨床ドラフト | 未決 / UNKNOWN |
| --- | --- | --- | --- |
| `Ttanka_Kin` → `unit_price_raw` | 数値なら `round`→bigint、否则 `0`（`030_stage.sql:1085`） | **原 text/精度を系譜保持**。会計丸め値を臨床 `unit_price` の唯一根拠にしない | AE `unit_price` bigint への正規化承認値は **UNKNOWN** |
| `Tsryo_Num` → `quantity_raw` | 正数なら小数 1 位丸め、否则 `1`（`030_stage.sql:1086-1087`） | 原値保持。既定 `1` への黙殺置換禁止 | AE `quantity`（DDL `numeric(10,1)` / model `numeric(10,2)`）への写像と CHECK `quantity > 0` の 0/負数扱いは **UNKNOWN** |
| `Tnebi_Kin` / `Tkin_Kin` | discount は stage 未送出；line_amount は会計側 | 臨床合計列は AE treatments に無し。監査用に raw 保持 | 表示用保持先（系譜 JSON / hold payload / 別監査表）は **UNKNOWN** |
| procedures `price` | stage リテラル `NULL`（`SalTanka_Num` 非送出） | マスタ料金採用は別判断。履歴行の単価は TRI raw を優先 | マスタ価格を履歴へ補完するか **UNKNOWN** |

### D6. dedup / 再取込

| キー | 役割 | ドラフト |
| --- | --- | --- |
| 候補 natural key | `TPK_PET_No \| TReat_Sno \| TGyo_Num`（= canonical `billing_items.source_pk` 形式） | **臨床行の dedup 候補**。再取込で同一キーの treatments 行を増やさない |
| 会計 `source_pk` | 同一文字列を billing_items 系譜でも使用 | キー文字列の一致 ≠ 同一エンティティ。会計行 ID と臨床行 ID を共有/上書きしない |
| 再取込 | 同一入力 2 回 | 取込済みキーは no-op または idempotent upsert。件数・表示参照が増えないこと |
| 取消/精算 | `TSeisan_Flg` 等 | 臨床反映規則は **UNKNOWN**。不明行を削除済み扱いしない |

### D7. prescriptions（本ドラフトの扱い）

一般処方の旧ソース表・明細・用量が **UNKNOWN / producer ABSENT** のため、本ドラフトは **treatments 履歴を主対象**とする。`prescriptions` は AE DDL ヘッダ形状の観測に留め、復元根拠のない用量/日数/発行日を作らない。処方を含む合成ケースは「ソース不足 → 理由付き保留」として fixture に含める（下表 FX-P*）。

### D8. レビュー承認前にやらないこと

- 21 表 / hash / `030_stage.sql` producer / migrations / 実データ取込の変更
- 会計 `billing_items` を treatments へコピーする実装
- dose_* / prescribed_at / duration_days への推測埋め
- 保留を「データなし」に置換して全件移行済みと報告すること

---

## 合成 fixture 設計（期待値・未実行）

目的: 同一医院の pet/record 帰属、元 qty/amount 保持、再取込非重複、欠損参照 vs 未移行の区別を **合成データだけで**検証可能にする。実原本件数・実臨床値は入れない（UNKNOWN）。実行は disposable / 別承認ゲート。

### 共通前提

| 項目 | 値 |
| --- | --- |
| clinic | 合成 `clinic_id=C1`（CSV placeholder 置換後）。旧支店コードは使わない |
| 対照 clinic | `C2`（衝突検査専用。C1 バンドルから参照しない） |
| pet/record | C1 内: `P1`+`R1`（本線）、`P2`+`R2`（別患者）。C2: 同形の旧番号文字列を故意に衝突させ、crosswalk は clinic スコープで分離 |
| マスタ | C1 `procedures` に `legacy_sale_no=S-PROC-1`（sal_tri_kbn='1'）のみ confirmed。注射専用マスタ行の採否は **UNKNOWN** のため「未分類マスタ」行を別キーで用意し hold 対象にする |
| 会計並行 | 同一 TRI 行から billing_items が既に作れる状態を許可。臨床取込は会計 ID を書き換えない |
| 過去分 | read-only。表示・再取込が新規請求/`billing_items.treatment_id` 埋めを起こさない |

### 行セット（合成 ID はテスト用記号。臨床的意味を持たせない）

| ID | 旧キー (pet\|sno\|line) | ねらい | 期待 disposition | 期待理由コード |
| --- | --- | --- | --- | --- |
| FX-OK-1 | P1\|SNO1\|001 | 同一医院・同一ペット・§31 経由で R1 に帰属。`TChri_No=S-PROC-1`。raw qty/amount あり | import | （なし） |
| FX-OK-2 | P1\|SNO1\|002 | 同一 settlement の 2 行目。別 line。親 record 同一 | import | （なし） |
| FX-AMT-1 | P1\|SNO2\|001 | raw `quantity_raw`/`unit_price_raw` が会計 stage 丸めと食い違う合成値（例: 小数桁差）。臨床側は原値保持列/系譜で照合 | import（原値保持）または `HOLD_PRICE_PRECISION_PENDING`（正規化未承認時） | 後者なら `HOLD_PRICE_PRECISION_PENDING` |
| FX-DEDUP-1 | P1\|SNO1\|001 | FX-OK-1 と同一キーを二度投入 | 2 回目 no-op。treatments 件数+0、billing_items 件数+0 | （idempotent） / 衝突時のみ `HOLD_DEDUP_COLLISION` |
| FX-HOLD-REC | P1\|SNO9\|001 | settlement はあるが KNJO.`JKrt_No` 欠落で record 解決不能 | hold | `HOLD_REF_MISSING_RECORD` |
| FX-HOLD-PET | P9\|SNO1\|001 | pet crosswalk なし | hold | `HOLD_REF_MISSING_PET` |
| FX-HOLD-PROC-MISS | P1\|SNO3\|001 | `TChri_No` がマスタにも原本 sale_items にも無い | hold または item_type=other+FK NULL（レビュー選択）。黙って procedures を捏造しない | `HOLD_REF_MISSING_PROCEDURE` |
| FX-HOLD-PROC-NYM | P1\|SNO3\|002 | sale_items 上は存在するが **C1 procedures 未移行**（意図的に stage/CSV から除外） | hold | `HOLD_REF_NOT_YET_MIGRATED_PROCEDURE` |
| FX-HOLD-MED-NYM | P1\|SNO4\|001 | 薬剤参照が必要に見える行だが medicines 非 21 表・未移行 | hold | `HOLD_REF_NOT_YET_MIGRATED_MEDICINE` |
| FX-HOLD-CLASS | P1\|SNO5\|001 | `Ttri_Kbn` が写像表外 | hold | `HOLD_CLASSIFICATION_UNKNOWN` |
| FX-HOLD-DOSE | P1\|SNO6\|001 | `TRyou_Num`/`Tkaisu_Num` のみ埋まり dose 写像なし | 行の非 dose 部が import 可能でも dose_* は埋めない。用量を必須とする方針なら hold | `HOLD_DOSE_UNMAPPED`（方針待ちは UNKNOWN） |
| FX-HOLD-DT | P1\|SNO7\|001 | `Ttreat_Date` NULL かつ親日付フォールバック未承認 | hold | `HOLD_DATETIME_UNPLACED` |
| FX-ISO-C2 | （C2 側で P1\|SNO1\|001 と同文字列） | 他医院の同形旧 ID | C1 取込結果から参照されない / C2 行は C2 のみ | `HOLD_CROSS_CLINIC_FORBIDDEN`（誤結合時） |
| FX-P-ABSENT | — | 一般処方行を「ソース表 ABSENT」として 1 ケース登録 | hold（作らない） | `HOLD_SOURCE_ABSENT_PRESCRIPTION` |
| FX-NO-REBILL | P1\|SNO1\|001 表示 | 履歴 UI / 再読込 | 新規 billing_items 0、`treatment_id` 埋め 0、自動請求 0 | 違反時は FAIL（コードというより受入） |

### 照合式（医院・種類・年別）

`原本合成行数 = import 行数 + 理由付き hold 行数 + 根拠付き dedup 除外行数`

- FX-OK-* と（承認済み）FX-AMT-1 が import。
- FX-HOLD-* / FX-P-ABSENT が hold（理由コード必須）。
- FX-DEDUP-1 の 2 回目は dedup 除外または no-op。
- 会計側 billing_items 行数は臨床 import の成否で増減しない（並行契約）。

### 元 qty / amount 保持の期待

| チェック | 期待 |
| --- | --- |
| raw 保存 | `quantity_raw` / `unit_price_raw` / `discount_raw` / `line_amount_raw` が合成入力とビット一致（text） |
| 会計差分 | 同一キーの billing_items.unit_price/quantity（丸め後）との差を別欄に記録。差の存在自体は FAIL にしない |
| 正規化列 | AE `unit_price`/`quantity` を埋める場合は変換式と承認 ID を fixture 注記に固定。未承認なら正規化列は NULL または hold |
| ゼロ埋めたて禁止 | 会計 stage の「非数値→0/1」規則を臨床 raw に適用しない |

### 再取込非重複の期待

1. 初期取込後の treatments 件数 `N`（import のみ）を固定。
2. 同一バンドル再取込後も `N`。hold 行の理由コード分布も不変。
3. 既存行の `id` が変わって表示リンクが壊れないこと（upsert 方針は実装ゲートで選択；本設計は「増えない・リンクが壊れない」のみ要求）。

---

## 理由コード付き保留ルール（欠損参照 vs 未移行）

目的: 保留を隠さず、**原本に参照が無い/壊れている**ことと、**参照先エンティティは計画上存在するがまだ移行されていない**ことを機械的に区別する。コードはレビュー用ドラフト。実装テーブル名は未固定。

### 判定順序（先に当たった理由で停止）

1. ソース契約欠如（処方など）
2. clinic スコープ違反
3. pet / record 解決
4. dedup
5. 分類写像
6. マスタ FK（missing vs not-yet-migrated）
7. datetime / dose / price の未契約埋め防止

### コード一覧

| code | 区分 | いつ付けるか（証拠条件） | やってはいけないこと |
| --- | --- | --- | --- |
| `HOLD_SOURCE_ABSENT_PRESCRIPTION` | source absent | 一般処方として扱う行だが producer/旧表根拠が ABSENT（本票 prescriptions 節） | 空の `prescriptions` ヘッダや YBS 流用で件数を作る |
| `HOLD_CROSS_CLINIC_FORBIDDEN` | isolation | 解決した pet/record/procedure の clinic がバンドル clinic と不一致 | 他医院 ID の読替・氏名結合 |
| `HOLD_REF_MISSING_PET` | **missing ref** | `TPK_PET_No` が当該 clinic の pet crosswalk に存在しない | 同名ペットへ接続 |
| `HOLD_REF_MISSING_RECORD` | **missing ref** | §31 経路で `JKrt_No` 欠落、または karte crosswalk 無し | `TReat_Sno` をカルテ番号として直結 |
| `HOLD_REF_MISSING_PROCEDURE` | **missing ref** | `TChri_No` が原本 `sale_items` / `MST_SAL_INFO` に存在しない（unmatched） | 同名処置マスタの新規捏造 |
| `HOLD_REF_MISSING_MEDICINE` | **missing ref** | 薬剤キーが原本マスタ集合に存在しない（元表確定後） | medicines を推測作成 |
| `HOLD_REF_NOT_YET_MIGRATED_PROCEDURE` | **not yet migrated** | 原本 sale_items 上は存在するが、対象 clinic の procedures CSV/stage に未載、または mapping_status が importable でない | missing と同一コードに畳む；「マスタ無し」と報告する |
| `HOLD_REF_NOT_YET_MIGRATED_MEDICINE` | **not yet migrated** | 薬剤マスタ移行が 21 表外で未実施、または当該キーが今後の medicines 取込対象と識別できる | 処方/投薬履歴完了とみなす |
| `HOLD_REF_NOT_YET_MIGRATED_RECORD` | **not yet migrated** | カルテ行は原本にあるが medical_records が未 importable（親 21 表の順序待ち） | record missing と同一視 |
| `HOLD_CLASSIFICATION_UNKNOWN` | classification | `Ttri_Kbn`/`TRv_Kbn`/`SalTri_Kbn` が承認写像表外。注射範囲未確定行を含む | 黙って `other`/`procedure` に落とす（落とすなら別コードと件数） |
| `HOLD_DATETIME_UNPLACED` | datetime | 行実施日時の置き場未契約、または必須日時が NULL でフォールバック未承認 | 会計日/親カルテ日へ黙って置換 |
| `HOLD_DOSE_UNMAPPED` | dose | 用量・用法列があるが dose_*/admin_route 写像未承認 | `#201` dose スナップショットへ文字列を流し込む |
| `HOLD_PRICE_PRECISION_PENDING` | price | raw 金額/数量はあるが AE 型への正規化が未承認。または会計丸めとの差を臨床値にできない | stage の 0/1 既定や round 結果を臨床唯一値にする |
| `HOLD_DEDUP_COLLISION` | dedup | 同一 natural key で系譜内容が不一致（真の衝突） | 行を複製して両方 import |
| `HOLD_REBILL_GUARD` | safety | 過去分取込が billing 更新/`treatment_id` 埋め/請求生成を要求する経路に入った | 自動再請求を続行 |

### missing vs not-yet-migrated の判定手続き

| 参照種別 | missing ref | not yet migrated |
| --- | --- | --- |
| procedure | 原本 `MST_SAL_INFO`/`sale_items` に `PK_Sal_No`/`TChri_No` が無い | 原本にはあるが、当該 clinic の procedures 21 表出力にキーが無い、または `mapping_status` が confirmed/inferred 外 |
| medicine | 確定した薬剤マスタ元集合にキーが無い（元集合自体が UNKNOWN の間は medicine FK を埋めず `HOLD_REF_NOT_YET_MIGRATED_MEDICINE` または source 欠如を優先） | medicines 契約/取込が未実施で、キーが将来対象と識別できる |
| record | KNJO/`JKrt_No`/karte 原本側が欠落 | karte 原本はあるが medical_records が未取込 |
| pet | 原本ペット自体が無い/ crosswalk 不能 | ペット stage 行が未 importable |

`not yet migrated` は **再実行で解消し得る**（親マスタ/カルテ取込後）。`missing ref` は **原本追加または明示棄却**が必要。集計では両コードを分け、`import + hold(missing) + hold(not_yet) + dedup` を公開する。

### 出力契約（保留行）

保留行は少なくとも次を残す（個人情報を共有表へ増やさない範囲で）:

- `reason_code`（上表）
- `legacy_source_pk`（pet\|sno\|line）
- `clinic_id`（バンドル）
- 欠ける参照の種別と、missing / not_yet の判定に使った証拠フラグ（例: `sale_item_present=yes`, `procedure_csv_present=no`）
- raw qty/amount/datetime/dose 列（存在する入力のみ。無いものを空で埋めない）

### 本票更新範囲の再確認

この追記（契約ドラフト / 合成 fixture 設計 / 理由コード）は設計成果のみ。21 表契約・hash・producer・DB・実データ・root `todo*.md` は変更しない。共有環境の実データ投入・migrate・push・Linear Done は別承認。

移行全体の完了には原本全期間の突合、保留解消または対象データ責任者による明示的な処置、再取込で増えないこと、医院/患者分離と履歴表示の受入が必要。原本不足を「データなし」に置換しない。
