# UAT-Q2-TREATMENTS-IMPORT: 処置・注射・処方の移行契約差分

状態: **全種類・全期間の移行契約設計 READY／実装・取込未実施**。2026-09-19に依頼者が「マスタ/患者別処置履歴/対象期間」の問いへ「全部」と回答した。処置名・料金マスタと患者ごとの処置履歴を今期対象とし、任意の年数/種類で減らさない。要件・状態の正本は [TODO](../../../todo-issue.md#uat-q2-treatments-import)、以前の判断経緯は [医院 UAT 記録](../stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補)。

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
| 参照欠損 vs 未移行 | **UNKNOWN** | — | 保留理由を残す |
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
| treatments（将来） | 候補 dedup 旧複合 PK；候補 FK pet/record via §31、procedure via `TChri_No` | producer 自体、item_type 写像、実施日列、dose、会計との二重計上 |
| prescriptions（将来） | AE DDL FK 形状のみ観測 | 旧ソース表、明細、用量、YBS 流用 |

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

## 合成 fixture と照合案（未実施）

1. 同一医院・同一ペット・同一旧カルテに、処置と注射のマスタ参照、実施行、処方行、会計行を持つ合成例を用意する。他医院・他ペットの旧ID衝突例も入れ、相互参照しないことを検査する。
2. 元数量・単位・元金額と変換後値を別列で照合し、丸め・単位換算・欠損値の扱いを個別に承認する。会計明細との金額差も記録する。
3. 同じ入力を二度取り込み、旧行単位の重複キーにより件数と参照が増えないことを確認する。参照先が存在しないケースと、その参照先がまだ移行されていないケースを分けて停止/保留結果を照合する。
4. 臨床履歴の表示で過去カルテ詳細へ到達し、元来院・処置・注射・処方が読み取り専用で表示されることを確認する。履歴の表示が新たな `billing_items` や請求を生成しないことを検査する。
5. 訂正・復旧ケースで旧行と新行の対応、件数・金額、監査手段を照合する。承認された disposable rehearsal の保存→再読込→表示はこの後の別ゲートとする。

移行全体の完了には原本全期間の突合、保留解消または対象データ責任者による明示的な処置、再取込で増えないこと、医院/患者分離と履歴表示の受入が必要。原本不足を「データなし」に置換しない。この票の更新は21表契約・producer・DB・実データを変更しない。共有環境の実データ投入は別承認。
