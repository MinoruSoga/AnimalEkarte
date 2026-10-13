# EMR-342: 取込契約に薬剤マスタの表を足す設計案

作成: 2026-10-11 / 状態: レビュー用提案、実装未着手。

対象は旧カルテ移行の producer（old_db）→ AnimalEkarte の取込契約に、薬剤マスタ（`medicines`）の表を足す設計。
現行の取込処理、CSV、manifest、契約のハッシュ、DB は変更しない。
[Plane EMR-342](https://app.plane.so/baritechllc/projects/6a009324-8a8a-455c-b16a-41b3f6207b0e/work-items/bcb78166-ae47-4ba6-8291-087a9eca9753/) を進捗の正本とする。

## 1. 決定と範囲

| 決定（2026-10-11、曽我稔） | 内容 |
| --- | --- |
| Q1 = 案A | 旧分類のうち薬剤に明らかに当たる 9 分類（下の §3.1）だけを自動で薬剤マスタへ取り込む。判断が割れる分類は診療項目のまま移行し、一覧を医院に確認してもらう |
| Q2 = 案A | 10/20 の本番環境の構築開始までには間に合わせない。構築開始時は薬剤を診療項目として移行し、後から薬剤マスタへ移す |

やること: 契約の設計（表・列・対応表・変換・検証方法・変更点・実装票の分割案）。
やらないこと: 取込契約・取込処理・移行データの実装、old_db の修正、本番・共有STGへの投入、migration の適用、薬剤の在庫との紐付け（`inventory_id` は空で出す）、薬剤の用量計算パラメータの投入。

## 2. 確認した証拠（接点）

| 項目 | 観測 | 設計への含意 |
| --- | --- | --- |
| 取込契約の表 | `backend/internal/csvimport/cutover_contract.go` の `CutoverTableSpecs` は 21 表で、`medicines` が無い。1 表の定義は、表名・列・ID帯の列・必須の文字列列の 4 つ（`:191-196`、`procedures` は `:245`） | 22 表目として足せる。表の位置は親の後・子の前 |
| 見積明細の薬剤参照 | `estimate_items` の列に `medicine_id` が既にある（`:260`）。参照グラフは `estimate_items.medicine_id → medicines` を `cutoverUnavailableParent`（空であること）と宣言し（`cutover_reference_graph.go:54`）、FK の一覧にも行がある（`cutover_import.go:180`） | 表を足すと、この参照を「CSVの親」に切り替えられる |
| 会計明細 | `billing_items` の列は `id, clinic_id, billing_id, category, name, unit_price, quantity, tax_type, is_insurance_applicable, sort_order`（`cutover_contract.go:256`）。処置・薬剤への参照IDを持たない。`category` は `item_category`（`001_init.sql:108`、`medicine` を含む）で、現行の出力は `procedure` と `other` のみ | 薬剤を処置から外しても会計明細の参照は壊れない |
| `medicines` の DDL | `001_init.sql:740-759`。列: `name, price, is_active, description, parent_id, dosage_form, medicine_unit, inventory_id, default_quantity, tax_type, tax_rate, sort_order, is_non_insurance`。同一医院で名称が重複しない一意索引（`:2452`、削除されていない行のみ）。用量計算の列は後続の追加（`#201`、`:3446` 以降） | 名称の重複は処置と同じ規則で回避する。用量計算の列は旧カルテに元が無いので空のまま |
| 見積明細の DDL | `001_init.sql:1825`。「参照のうち 1 つだけ」という制約は無く、数量が正であることだけを検査する | 参照を `procedure_id` から `medicine_id` へ付け替えても DB 側の制約に触れない |
| producer の分類 | `old_db/sql/migration/030_stage.sql`: `SalTri_Kbn` が唯一の振り分けで、`1` は処置、`2` は商品（`:239-268`、`:288-314`）。旧分類名は処置の `description` に「【旧分類】」、剤形は「【剤形】」の文字列として入る | 薬剤は区分 `1` の行。旧分類の文字列で見分ける |
| 見積明細の producer | `030_stage.sql:1376-1440`: 区分 `1` の行は `procedure_id = sale_item_key + offset`、`medicine_id` は常に NULL | 付け替えは producer で行う |
| 隣接する契約設計 | MIG-15（`docs/work/campaigns/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md`）D1: `medicine_id` は「21 表外のため現状 FK を埋めない」。理由コード `HOLD_REF_NOT_YET_MIGRATED_MEDICINE` | 本設計を採用した時点で、D1 の `medicine_id` の行と理由コードを更新する |

## 3. 設計

### 3.1 どの行を薬剤にするか

規則: 区分 `1` の行のうち、旧分類名の先頭コードが次の 9 つのものを `medicines` に出し、`procedures` からは出さない。それ以外は今までどおり `procedures` に出す。

`07`（フィラリア予防薬）、`18`（注射（抗生剤））、`19`（注射（特殊薬））、`36`（内用）、`37`（内用（抗生剤））、`38`（内用（特殊薬））、`39`（外用）、`40`（外用（眼科薬））、`96`（フィラリア予防注射）

- 旧分類名の括弧に全角と半角が混在する（例: `18.注射(抗生剤)` と `19.注射（特殊薬）`）ため、文字列の全体一致ではなく先頭コードで比較する。
- 規則の入力は producer の canonical `sale_items` の旧分類名の列とする。stage の `description` の文字列を解析しない（文字列の組み立ては変わりうる）。
- 許可リストは、オーナー承認の対応表として old_db の判定台帳（`data-migration-column-target-checklist`）に記録する。

旧分類 61 種の対応表（件数は各院のリハーサル用データの `procedures` の行数。「城東・敷島・箱」は 3 院が同じマスタ由来で同数）:

| 旧分類 | 八王子 | 城東・敷島・箱 | 本票の扱い |
|:---|---:|---:|:---|
| 01.初診料 | 5 | 6 | 処置 |
| 02.再診料 | 28 | 29 | 処置 |
| 03.往診料 | 5 | 10 | 処置 |
| 04.時間外料金 | 8 | 6 | 処置 |
| 05.入院料 | 9 | 16 | 処置 |
| 06.ペットホテル | 21 | 12 | 処置 |
| 07.フィラリア予防薬 | 83 | 78 | **薬剤** |
| 08.ワクチン | 50 | 42 | ワクチン（診療項目のまま） |
| 09.検査 | 40 | 61 | 処置 |
| 10.尿検査 | 10 | 14 | 処置 |
| 11.便検査 | 5 | 5 | 処置 |
| 12.検査（血液） | 137 | 149 | 処置 |
| 13.検査（レントゲン） | 22 | 39 | 処置 |
| 14.検査（眼科） | 5 | 6 | 処置 |
| 15.皮膚検査 | 24 | 19 | 処置 |
| 16.病理関係 | 25 | 28 | 処置 |
| 17.注射料 | 107 | 127 | 処置（手技料） |
| 18.注射(抗生剤) | 47 | 45 | **薬剤** |
| 19.注射（特殊薬） | 128 | 119 | **薬剤** |
| 20.処置 | 265 | 276 | 処置（剤形のある行は判断が割れる） |
| 30.麻酔 | 29 | 28 | 判断が割れる |
| 31.手術 | 88 | 131 | 処置 |
| 32.内外科手術 | 75 | 73 | 処置 |
| 33.骨外科手術 | 22 | 19 | 処置 |
| 35.投薬 | 3 | 3 | 判断が割れる |
| 36.内用 | 671 | 762 | **薬剤** |
| 37.内用（抗生剤） | 78 | 73 | **薬剤** |
| 38.内用（特殊薬） | 75 | 71 | **薬剤** |
| 39.外用 | 272 | 250 | **薬剤** |
| 40.外用（眼科薬） | 95 | 94 | **薬剤** |
| 50.フード:犬（療法食） | 414 | 1,024 | 商品（処置側に入っている） |
| 51.フード（犬）Hills | 101 | 114 | 商品（同上） |
| 52.フード（猫）Hills | 64 | 59 | 商品（同上） |
| 53.療法食サンプル | 21 | 19 | 商品（同上） |
| 54.療法食ｶｰﾄﾞﾌｰﾄﾞｻｰﾋﾞｽ | 1 | 1 | 商品（同上） |
| 55.一般食ｶｰﾄﾞﾌｰﾄﾞｻｰﾋﾞｽ | 1 | 1 | 商品（同上） |
| 56.Nutroフード（犬） | 94 | 94 | 商品（同上） |
| 57.Nutroフード（猫） | 56 | 56 | 商品（同上） |
| 58.フード:猫（療法食） | 313 | 719 | 商品（同上） |
| 59.フード（エキゾ） | 12 | 11 | 商品（同上） |
| 61.フード（犬）一般食 | 167 | 137 | 商品（同上） |
| 62.フード（猫）一般食 | 83 | 85 | 商品（同上） |
| 63.Virbac商品 | 20 | 46 | 判断が割れる |
| 64.ｱｲﾑｽ・ﾕｶﾇﾊﾞﾌｰﾄﾞ（犬） | 13 | 13 | 商品（処置側に入っている） |
| 65.ｱｲﾑｽ・ﾕｶﾇﾊﾞﾌｰﾄﾞ（猫） | 2 | 2 | 商品（同上） |
| 69.処方料 | 1 | 0 | 処置（手技料） |
| 70.文書料 | 18 | 20 | 処置 |
| 80.会計 | 51 | 53 | 処置 |
| 90.その他 | 359 | 392 | 処置 |
| 91.点滴 | 47 | 45 | 判断が割れる |
| 92.Shampoo | 43 | 45 | 商品（処置側に入っている） |
| 93.トリミング | 291 | 417 | 処置 |
| 96.フィラリア予防注射 | 7 | 18 | **薬剤** |
| 97.健康診断 | 2 | 446 | 処置 |
| 98.キャンペーン処理マスタ | 509 | 647 | 処置 |
| 99.しつけ | 34 | 35 | 処置 |
| Ａ.スナック | 3 | 3 | 商品（処置側に入っている） |
| Ｂ.その他 | 2 | 6 | 処置 |
| Ｃ.会計 | 1 | 2 | 処置 |
| Ｄ.トリミング | 3 | 2 | 処置 |
| Ｅ.薬浴 | 2 | 2 | 処置 |
| **合計** | **5,167** | **7,105** | |

グループ別の合計:

| グループ | 八王子 | 城東・敷島・箱 |
|:---|---:|---:|
| 薬剤（本票で `medicines` へ移す） | 1,456 | 1,510 |
| 処置（そのまま） | 2,174 | 3,048 |
| ワクチン（そのまま） | 50 | 42 |
| 商品（処置側に入っているが、本票では移さない） | 1,388 | 2,383 |
| 判断が割れる（診療項目のまま。医院が確認） | 99 + 20.処置の剤形のある行 9 = 108 | 122 + 同 9 = 131 |
| 合計 | 5,167 | 7,105 |

判断が割れる分類の扱い: `medicines` へは移さず、`procedures` に残す。八王子は 108 行、城東・敷島・箱は 131 行を、旧分類名・名称の一覧にして医院へ渡す（EMR-338 の医院向け一覧の更新で案内する）。医院が薬剤と判断した行は、移行後に医院が薬剤マスタへ登録し直すか、後続の移行で `medicines` へ移す。

「商品」の行（八王子 1,388）は、物販の商品が処置側に入っている。本票の範囲外とし、同じ仕組みで `merchandise_items` へ移すかどうかを別に決める（§5）。

### 3.2 `medicines` の列と旧列の対応

| `medicines` の列 | 値の元（producer） | 旧列 | 備考 |
| --- | --- | --- | --- |
| `id` | `sale_item_key + offset`（処置・商品と同じ ID 帯の採番） | `MST_SAL_INFO.PK_Sal_No` | 1 行が `procedures` と `medicines` の両方に出ない（§3.6） |
| `clinic_id` | プレースホルダ `{{CLINIC_ID}}` | — | 契約のプレースホルダ一覧に `medicines.clinic_id` を足す |
| `name` | 主名称。同名が複数あれば「【旧コード:…】」を付けて一意にする | `SalTri_Data` | 同一医院で名称が一意（`001_init.sql:2452`）。処置と同じ規則。名称が空の行は `needs_review` として CSV から外す（現行どおり） |
| `price` | 単価を円に丸めた値 | `SalTanka_Num` | MIG-25 ②-1（処置の料金）の決定に従う。同じ未解決（§5-1） |
| `is_active` | `true` | — | |
| `description` | 備考・別名・別名2・旧分類・剤形を、処置と同じ組み立てで連結 | `SalBiko_Data`、`SalTriNm2_Data`、`SalTri_KbnNm`、剤形 | 剤形の文字列は `dosage_form` に変換できなくても残す（情報を落とさない） |
| `parent_id` | NULL | — | 旧カルテに親子の列が無い |
| `dosage_form` | §3.3 | 剤形の文字列と旧分類 | 変換できない行は NULL |
| `medicine_unit` | NULL | — | 元が無い。`per_tablet` などは旧カルテから導出できない |
| `inventory_id` | NULL（空であることを必須にする） | — | `vaccines.inventory_id` と同じ扱い（`cutoverUnavailableParent`） |
| `default_quantity` | `1.0` | — | DDL の既定値と同じ。CSV の空は NULL になるため、値を明示する |
| `tax_type`、`tax_rate` | MIG-25 ②-1 の決定に従う。それまで処置と同じ（税抜、10%）の固定値 | `SalRv_Kbn` | 処置の料金と税区分は組で決める（§5-1） |
| `sort_order` | 旧カルテの並び順 | — | |
| `is_non_insurance` | `false` | — | 元が無い |

契約の定義（案）: 列は上表の順、ID帯の列は `id` と `parent_id`、必須の文字列列は `name` と `description`。

### 3.3 剤形・単位の変換

`dosage_form` の取りうる値は `tablet`、`liquid`、`injection`、`topical`、`powder`。変換は次の順で適用し、当たらなければ NULL にする。

1. 剤形の文字列から変換する。

   | 剤形の文字列（八王子で確認したもの） | `dosage_form` |
   |:---|:---|
   | 錠剤、錠剤：…、Tab | `tablet` |
   | 注射液、注射、液体（注射液） | `injection` |
   | 粉 | `powder` |
   | 液体 | `liquid` |
   | スポットタイプ | `topical` |
   | フィルム、ジャーキータイプ、液体＋酸素（吸入麻酔 | 変換しない（NULL）。経口の形状や用途が一意に決まらない |

2. 1 で決まらない行は、旧分類から既定値を置く: `39.外用`、`40.外用（眼科薬）` は `topical`、`18.注射(抗生剤)`、`19.注射（特殊薬）`、`96.フィラリア予防注射` は `injection`。`07`、`36`、`37`、`38`（内用・予防薬）は既定値を置かない（錠・液・粉が混在する）。

`medicine_unit` は全件 NULL。

埋まる件数の見込み（リハーサル用データの集計）:

| 院 | 薬剤の行 | 剤形の文字列から | 旧分類の既定値から | 埋まらない行 | 埋まる割合 |
|:---|---:|---:|---:|---:|---:|
| 八王子 | 1,456 | 142 | 458 | 856 | 41% |
| 城東・敷島・箱 | 1,510 | 143 | 435 | 932 | 38% |

埋まらない行は医院が移行後に設定する。変換表と既定値は、医院または獣医師の確認を受けてから確定する。

### 3.4 見積明細の付け替え

producer は、見積明細の参照先が §3.1 の薬剤の行のとき、`procedure_id` を NULL にして `medicine_id` に同じ ID を入れる。`category` は `procedure` から `medicine` に変える（`item_category` に値がある）。それ以外の行は変えない。

付け替えの対象行数（リハーサル用データ）: 八王子 1,296 行、城東 3,669 行、敷島 595 行、箱 442 行。

取込側は、参照グラフの `estimate_items.medicine_id` を `cutoverUnavailableParent` から `cutoverCSVParent`（`medicines` の CSV に存在すること）に切り替える。

### 3.5 会計明細

変更しない。会計明細は処置・薬剤への参照IDを持たず、過去の会計の再計算も行わない。薬剤の行の `category` は現行どおり `procedure` のまま出る。`medicine` に変えるかは別に決める（§5-2）。

### 3.6 処置の表からの除外と ID

- `procedures.csv` には許可リストの行を出さない。`medicines.csv` にだけ出す。
- `procedures`、`medicines`、`merchandise_items` の ID は、同じ採番（`sale_item_key + offset`）の中で互いに重ならない。1 つの旧マスタ行は、3 つの表のうち 1 つにだけ現れる。

### 3.7 取込契約（AnimalEkarte）の変更点

| ファイル | 変更 |
| --- | --- |
| `cutover_contract.go` | `CutoverTableSpecs` に `medicines` を足す（`procedures` の後、`merchandise_items` の前）。プレースホルダ一覧に `medicines.clinic_id` を足す。契約のハッシュの定数を更新する。表の数は 21 から 22 になる |
| `cutover_import.go` | FK の一覧に `medicines.clinic_id → clinics`、`medicines.parent_id → medicines` を足す（`estimate_items.medicine_id → medicines` は既にある） |
| `cutover_reference_graph.go` | `medicines` の参照（`clinic_id` は医院の種、`parent_id` は CSV の親、`inventory_id` は空必須）を足し、`estimate_items.medicine_id` を `cutoverCSVParent` に切り替える |
| `cutover_import_copy.go`、`cutover_contract_validate.go` | 列の型（`dosage_form` の列挙型、`default_quantity` の数値）と必須列の検査に対応する |
| テスト | `cutover_import_test.go`、`cutover_window_zero_test.go`、`cmd/stg-uat-skeleton` の表の数・順序の期待値を更新する。負例を足す（§4） |
| ドキュメント | `docs/ops/deploy/seeds/CLINIC_CSV_IMPORT.md` の「21 表」の表と記述を 22 表に更新する |

契約のハッシュが変わるため、旧契約（21 表）の bundle は取込前の検査で拒否される。現在の 4 院の bundle はすべて `REHEARSAL_ONLY` で、正式取込の対象ではない。

### 3.8 producer（old_db）の変更点

| 対象 | 変更 |
| --- | --- |
| `sql/migration/030_stage.sql` | `medicines` の stage 表と INSERT（許可リストの行）、`procedures` の INSERT から許可リストの行を除く、見積明細の付け替え（§3.4） |
| `scripts/lib/stage-csv-columns.mjs` ほか | CSV の出力列・表の順序・プレースホルダ・契約のハッシュの鏡写し |
| 判定台帳 | 許可リストと剤形の変換表を、オーナー承認の判断として記録する |
| 検証 | 件数の突合（§4）を、リハーサル bundle の生成後に実行する |

## 4. 検証方法

| # | 検査 | 期待値（八王子のリハーサル用データ） |
| --- | --- | --- |
| V1 | 件数の突合: `procedures.csv` + `medicines.csv` = 変更前の `procedures.csv` | 3,711 + 1,456 = 5,167 |
| V2 | `medicines.csv` の全行の旧分類の先頭コードが許可リストに含まれる。`procedures.csv` には許可リストの行が 0 | 0 件の違反 |
| V3 | 3 表（`procedures`、`medicines`、`merchandise_items`）の ID が互いに重ならない | 0 件の重複 |
| V4 | 見積明細: `medicine_id` が空でない行 = 付け替え対象行数。`procedure_id` が空でない行 = 変更前 − 付け替え対象行数。`medicine_id` が `procedures` を、`procedure_id` が `medicines` を指す行は 0 | 1,296 行 |
| V5 | `dosage_form` が埋まる割合 | 41%（600 / 1,456） |
| V6 | 契約: 表の数が 22。旧契約（21 表）の bundle は拒否される。`medicines.csv` が無いのに `medicine_id` が空でない bundle は拒否される。`inventory_id` が空でない `medicines.csv` は拒否される | 全て拒否される |
| V7 | 名称の一意: 同一医院で `medicines.name` が重複しない | 0 件の重複 |

V1〜V5、V7 は、再生成した bundle の集計（件数のみ。値は出さない）で確認する。V6 は取込側の単体テストで確認する。

## 5. 未決・依存

1. 薬剤の `price` と `tax_type`: 処置と同じ未解決。MIG-25 ②-1 の決定（処置の料金を移行に含める）に従う。区分 `1`（処置）の `SalRv_Kbn` の分布は未確認（既存の判定記録の分布は区分 `1`・`2` を合算している）。料金だけ出すと、内税の金額が税抜扱いになる恐れがあるため、組で決める。
2. 会計明細の `category`: 薬剤の行を `medicine` にするか。変えると、過去の会計の分類別の集計に影響する。本票では変えない。
3. 「商品」の行（八王子 1,388）: 同じ仕組みで `merchandise_items` へ移すかは別の決定。
4. 用量計算（`#201`）: `medicines` の計算パラメータと `medicine_dose_params` は旧カルテに元が無く、取り込んだ薬剤では用量計算が使えない。医院が薬剤ごとに設定する。
5. 判断が割れる分類（八王子 108 行）の最終的な扱い: 医院の回答を待つ。
6. MIG-15 の D1（`medicine_id`）と理由コード `HOLD_REF_NOT_YET_MIGRATED_MEDICINE`: 本設計の採用時に更新する。
7. 実施時期: Q2 = 案A のため、10/20 の構築開始には間に合わせない。MIG-25 ②-1 の出力側の修正が先で、薬剤の分離はその上に載る。

## 6. 実装票の分割案

起票は本設計のレビュー後（EMR-342 の実装方針3）。

1. **出力側（MIG プロジェクト）**: `030_stage.sql` の `medicines` の stage 表と INSERT、`procedures` からの除外、見積明細の付け替え、CSV の出力列・順序・ハッシュ、判定台帳の更新、リハーサル bundle の再生成と V1〜V5・V7 の確認。依存: MIG-25 ②-1。
2. **取込側（EMR プロジェクト）**: §3.7 の変更と V6 の単体テスト。出力側と同じ契約のハッシュを使う（両側を同時にそろえる）。
3. **医院確認**: 判断が割れる分類の一覧（名称・旧分類）を医院へ渡し、回答を反映する。EMR-338 の医院向け一覧の更新と連動。

## 7. 確認した範囲と、していないこと

確認した: `cutover_contract.go`、`cutover_import.go`、`cutover_reference_graph.go` の該当箇所、`001_init.sql` の `medicines`・`estimate_items`・`item_category`、old_db の `030_stage.sql` の該当箇所、4 院のリハーサル用データ（`procedures.csv`、`estimate_items.csv`）の集計、MIG-15 の契約設計。

していない: 行データの閲覧、旧DBへの接続、取込処理・契約・CSV の変更、医院への確認、剤形の変換表の獣医師による確認、`cutover_import_copy.go` と `cutover_contract_validate.go` の列の型の扱いの詳細確認（実装票で確認する）。
