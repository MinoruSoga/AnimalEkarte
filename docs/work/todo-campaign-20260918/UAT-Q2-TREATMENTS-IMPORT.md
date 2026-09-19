# UAT-Q2-TREATMENTS-IMPORT: 処置・注射・処方の移行契約差分

状態: **全種類・全期間の移行契約設計 READY／実装・取込未実施**。2026-09-19に依頼者が「マスタ/患者別処置履歴/対象期間」の問いへ「全部」と回答した。処置名・料金マスタと患者ごとの処置履歴を今期対象とし、任意の年数/種類で減らさない。要件・状態の正本は [TODO](../../../todo-issue.md#uat-q2-treatments-import)、以前の判断経緯は [医院 UAT 記録](../stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補)。

## 現行契約と不足分

| 対象 | 旧側で確認できる根拠 | 現行の送出・取込先 | 未確定の差分 |
| --- | --- | --- | --- |
| 処置・注射のマスタ候補 | `old_db/sql/migration/030_stage.sql:209-270` は `MST_SAL_INFO` 由来の `legacy_canonical.sale_items` を `sal_tri_kbn='1'` で `animalekarte_stage.procedures` へ振り分ける。`legacy_sale_no`、`source_pk` を系譜として保持する。 | `backend/internal/csvimport/cutover_contract.go:205` の `procedures` は 21 表契約内のマスタ。`id`, `clinic_id`, `name`, `price` 等を持つ。 | 注射を含める旧種類・コード範囲、旧マスタキーと新 ID の確定 crosswalk、採用/除外ルールは **UNKNOWN**。現行のマスタ送出は患者ごとの実施履歴を意味しない。 |
| 来院ごとの処置・注射実施 | `030_stage.sql:1051-1108` の `legacy_canonical.billing_items` は `TBL_TRI_DATA` の会計明細として送出され、`sal_tri_kbn='1'` は `procedure` 区分になる。`legacy_treat_sno` は `legacy_record_no` に記録される。 | `cutover_contract.go:216` の `billing_items` は `billing_id`, `category`, `name`, `unit_price`, `quantity` 等の会計明細。`cutover_contract.go:203-227` に `treatments` 表仕様はない。 | 臨床的実施の元表・元列、旧カルテ/ペット/医院 FK、実施日時、処置マスタ FK、数量・単位・金額の原値、重複判定キー、訂正/取消、参照欠損時の扱いは **UNKNOWN**。会計明細から実施履歴を再構成できるとは確定していない。 |
| 処方履歴 | `030_stage.sql:1660` の `prescriptions` はワクチン lot 欠損のコメントであり、処方 INSERT の根拠ではない。 | `cutover_contract.go:203-227` に `prescriptions` 表仕様はない。 | 処方の旧表・元列、薬剤マスタ対応、用量・単位換算・用法・日数、発行/実施日時、旧カルテ/ペット/医院 FK、重複キー、欠損参照の扱いは **UNKNOWN**。 |
| 会計と過去カルテ表示 | `030_stage.sql:1053-1108` は会計明細に `billing_id` を必須とし、価格・数量を変換している。 | 現行 21 表に `billings` と `billing_items` がある（`cutover_contract.go:215-216`）。[UAT 記録](../stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補) は処置が一覧にないとカルテが空に見える問題を記録する。 | 会計を再計上せず臨床履歴としてどう表示するか、欠損と未移行の区別、読み取り専用表示の仕様、既存の詳細導線との接続は **UNKNOWN**。`billing_items` の存在を処置/処方履歴の取込完了とみなさない。 |

`030_stage.sql:1072-1108` の会計明細は `unit_price_raw` を数値判定して丸め、`quantity_raw` を正数判定して小数第 1 位へ丸め、それ以外は既定値にする。これは**現行の会計変換**であり、臨床履歴の元数量・元金額を保存する方針ではない。`procedures` 側の価格は `030_stage.sql:239-270` で `NULL` としている。会計金額と臨床実施値の対応・差額照合は別途決める必要がある。

## 旧列 → producer → AE（repo 根拠。不明列は UNKNOWN）

根拠はこの票では `old_db/sql/migration/030_stage.sql` と `backend/internal/csvimport/cutover_contract.go` のみ。物理表 `MST_SAL_INFO` / `TBL_TRI_DATA` は stage の `source_table` リテラルであり、この 2 ファイル内では FROM していない。列を推定しない。21 表 hash は観測のみ（変更しない）: `cutoverStageMappingSHA256=0d7f089990079af28c2ba454ee48188d55a67beab19a9f16a86dee93fec80597`、`cutoverCSVContractSHA256=19b2c5c270058b20c1fa816679c0430f236a0a181165ae1ed2257c64c83f6671`（`cutover_contract.go:17-18`）。

過去分の既定: **参照専用・自動再請求なし**（`todo-issue.md#uat-q2-treatments-import`）。`billing_items` 取込完了を処置/処方履歴完了とみなさない。氏名一致結合はしない。

### procedures（21 表に存在。`CutoverTableSpecs` `cutover_contract.go:205`）

Producer: `animalekarte_stage.procedures` INSERT `030_stage.sql:239-269`。FROM `legacy_canonical.sale_items` WHERE `sal_tri_kbn = '1'`（`246-247`）。ルータコメント `209-216`: `SalTri_Kbn` のみ。`1=procedure`, `2=merchandise`。未分類は canonical のみ。注射専用コード範囲は **UNKNOWN**。`sal_tri_kbn='1'` を全処置/全注射としない。

| 旧（030_stage で参照された列） | producer `animalekarte_stage.procedures` | AE 21 表 `procedures` | 注 |
| --- | --- | --- | --- |
| `sale_item_key` | `id = sale_item_key + :stage_id_offset` | `id`（band `id`,`parent_id`） | crosswalk 候補。確定採用/除外ルールは **UNKNOWN** |
| （定数 `1`。`030_stage.sql:22-23` は CSV で `{{CLINIC_ID}}` 置換） | `clinic_id` | `clinic_id`（placeholder `cutover_contract.go:175`） | 旧支店コードは clinic_id に使わない（コメント） |
| `primary_name`, 重複時 `legacy_sale_no` 接尾 | `name` | `name`（force-not-null） | 空名は `''` のまま `needs_review`。CSV 除外はコメント `214-216` のみ |
| なし（INSERT は `NULL`） | `price` | `price` | マスタ料金の旧列はこの INSERT に無い。**UNKNOWN** |
| なし（`true`） | `is_active` | `is_active` | 既定。旧列なし |
| `remarks`, `alternate_name`, `alternate_name2`, `category_label`, `dosage_form` | `description`（改行連結） | `description`（force-not-null） | 剤形はマスタ説明へ折り畳み。臨床用量ではない |
| なし（`NULL`） | `duration` | `duration` | 旧列 **UNKNOWN** |
| なし（`'none'`） | `anesthesia` | `anesthesia` | 既定。旧列 **UNKNOWN** |
| なし（`NULL`） | `parent_id` | `parent_id` | 旧列 **UNKNOWN** |
| なし（`'excluded'` / `0.10`） | `tax_type`, `tax_rate` | `tax_type`, `tax_rate` | 会計側 `sal_rv_kbn` は merchandise INSERT に出るが procedures INSERT には無い |
| `sort_no` | `sort_order = coalesce(sort_no, 0)` | `sort_order` | |
| なし（`false`） | `is_surgery` | `is_surgery` | 旧列 **UNKNOWN** |
| `legacy_sale_no` | `legacy_item_code` | 21 表 CSV 列に無い（`cutover_contract.go:205`） | 系譜。AE 取込列としては非契約 |
| リテラル `'MST_SAL_INFO'`, `source_pk` | `source_table`, `source_pk` | 21 表 CSV 列に無い | 系譜 |
| `mapping_status` / `confidence` | stage のみ | 21 表 CSV 列に無い | 空名は `needs_review` |

### treatments（21 表に無い。`cutover_contract.go:203-227` に表名なし）

`030_stage.sql` に `CREATE TABLE animalekarte_stage.treatments` / `INSERT INTO …treatments` は無い。

| 旧 | producer | AE 21 表 | 注 |
| --- | --- | --- | --- |
| 臨床実施の元表・元列 | **ABSENT** / **UNKNOWN** | **ABSENT** | 会計 `billing_items` から実施行を捏造しない |
| 旧 clinic / owner / pet / record FK | **UNKNOWN** | **UNKNOWN** | 氏名結合禁止 |
| 実施日時 | **UNKNOWN** | **UNKNOWN** | 診療日/実施日/会計日を混同しない。日付不明行を黙って除外しない |
| 処置マスタ FK | **UNKNOWN** | **UNKNOWN** | `procedures.id` との対応はマスタ側 crosswalk とは別 |
| 元数量・単位・単価・金額 | **UNKNOWN** | **UNKNOWN** | 会計丸め値を流用しない |
| 重複キー / 訂正・取消 | **UNKNOWN** | **UNKNOWN** | |
| 参照欠損 vs 未移行 | **UNKNOWN** | **UNKNOWN** | 保留理由を残す |

### prescriptions（21 表に無い。`cutover_contract.go:203-227` に表名なし）

`030_stage.sql` に `CREATE`/`INSERT` の `prescriptions` は無い。`1660` は YBS 接種 INSERT の lot 空文字コメント（`-- YBS has no lot columns (prescriptions)`）であり処方送出ではない。

| 旧 | producer | AE 21 表 | 注 |
| --- | --- | --- | --- |
| 処方の元表・元列 | **ABSENT** / **UNKNOWN** | **ABSENT** | 復元根拠のない用量・単位を作らない |
| 薬剤マスタ参照 | **UNKNOWN** | **UNKNOWN** | |
| 用量・単位・用法・日数・旧表現 | **UNKNOWN** | **UNKNOWN** | |
| 発行/実施日時、clinic/pet/record FK、重複キー | **UNKNOWN** | **UNKNOWN** | |

### billing_items（21 表に存在。会計明細。`cutover_contract.go:216`）

Producer: `animalekarte_stage.billing_items` INSERT `030_stage.sql:1072-1108`。FROM `legacy_canonical.billing_items`。`source_table` リテラル `'TBL_TRI_DATA'`。stage CREATE に `clinic_id` 列は無い（`1053-1071`）。

| 旧（030_stage で参照された列） | producer `animalekarte_stage.billing_items` | AE 21 表 `billing_items` | 注 |
| --- | --- | --- | --- |
| `item_key` | `id = item_key + :stage_id_offset` | `id`（band `id`,`billing_id`） | |
| `billing_key` | `billing_id`（NOT NULL。`1051`） | `billing_id` | 親は `billings`（`cutover_contract.go:215`） |
| なし（stage 表に列なし） | **UNKNOWN** in this INSERT | `clinic_id`（placeholder `cutover_contract.go:186`） | CSV 置換経路はこの 2 ファイルの INSERT に無い |
| `sal_tri_kbn` | `category`: `'1'`→`procedure`、`'2'` または else→`other` | `category` | NULL/その他も `other`。注射専用値は **UNKNOWN** |
| `item_name` | `name` | `name`（force-not-null） | coalesce なし |
| `unit_price_raw` | `unit_price`: 数値なら round→bigint、それ以外は `0` | `unit_price` | **会計変換**。臨床原値ではない |
| `quantity_raw` | `quantity`: 正数なら round 1 桁、それ以外は `1` | `quantity` | dest `numeric(10,1)`。臨床原値ではない |
| `tax_flg` | `tax_type`: `1/01` excluded、`0/00`/blank included、その他 excluded + `needs_review` | `tax_type` | |
| `insurance_flg` | `is_insurance_applicable` | `is_insurance_applicable` | `'1'/'01'` のみ true |
| `legacy_line_no` | `sort_order` | `sort_order` | 非数字は `0` |
| `source_pk`, `legacy_pet_no`, `legacy_treat_sno` | `source_pk`, `legacy_pet_no`, `legacy_record_no`（`1067` は treat_sno と同義） | 21 表 CSV 列に無い | 会計系譜。臨床実施キーではない |
| 実施日時・単位・取消フラグ・マスタ FK | この INSERT に無い | 21 表 `billing_items` にも無い | **UNKNOWN** |

親 `billings`（会計ヘッダ。臨床実施表ではない）: stage `030_stage.sql:908-949`。`scheduled_date` は `treat_date::date`（`948-949`、1900-01-01 超のみ）。AE `billings` 列は `id, clinic_id, medical_record_id, owner_id, pet_id, total_amount, status, scheduled_date, completed_at`（`cutover_contract.go:215`）。`legacy_treat_sno` はヘッダ系譜（`923`）であり treatments 行キーではない。

見積明細 `estimate_items`（`030_stage.sql:1385-1427`、`TBL_MIT_DATA`）は `sal_tri_kbn='1'` で `procedure_id` を付けるが、患者実施履歴ではない。本票の treatments 契約に流用しない。

## 確定した範囲と設計担当の調査項目

| 判断 | 現状 |
| --- | --- |
| 業務目的、要件責任者、受入責任者 | 目的は過去診療を欠落なく参照すること。出所は9月19日の依頼者回答。個人名を伴う要件/受入参照は契約変更前に実行票へ記録する。 |
| 採用する種類と用途 | 処置マスタ＋患者ごとの処置履歴を全種類。注射・薬剤/処方等の関連原本も棚卸しし、含めない関連行があれば理由と判断を残す。過去分は参照専用/自動再請求なしを設計の安全側既定とする。 |
| 対象期間と日付基準 | 提供原本の最古から最終抽出まで全部。種類/年数の任意filterを設けない。診療日/実施日/会計日/処方日は意味ごとに保持し、日付不明行を黙って除外しない。 |
| 元列、変換、FK/crosswalk、重複キー | 両repoの設計担当が調べる次作業。旧IDと同一医院/飼主/ペット/旧カルテの対応を根拠付きで埋め、元列不明のものだけ保留。氏名一致で結合しない。 |
| 元数量・単位・金額、処方用量/用法/日数、訂正と復旧 | **UNKNOWN**。会計明細の丸めや既定値を臨床履歴へ流用しない。 |
| 合格条件 | 下記の原本全行照合、参照/原値保持、重複なし、読み取り専用表示、再請求なし。数量/金額の無承認丸めは不可。要件未回答を理由に、この合成受入設計を止めない。 |

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
