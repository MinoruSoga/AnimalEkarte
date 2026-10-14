# EMR-342: 取込契約に薬剤マスタの表を足す設計案

作成: 2026-10-11 / 状態: レビュー用提案、実装未着手。

対象は旧カルテ移行の producer（old_db）→ AnimalEkarte の取込契約に、薬剤マスタ（`medicines`）の表を足す設計。
現行の取込処理、CSV、manifest、契約のハッシュ、DB は変更しない。
[Plane EMR-342](https://app.plane.so/baritechllc/projects/6a009324-8a8a-455c-b16a-41b3f6207b0e/work-items/bcb78166-ae47-4ba6-8291-087a9eca9753/) を進捗の正本とする。

## 2026-10-13 切替方式の改訂（ユーザー採用）

PR #655の統合保留を受け、今回のユーザー回答で「21表維持＋22表候補分離」を採用した。既定 `legacy-21` は21表formalの経路・version/pin・medicine_id空欄制約を保持する。薬剤22表は明示 `medicines-22-candidate` profileで、明示rehearsal modeかつREHEARSAL_ONLYだけ検証可能。正式preflight/apply/resume/verifyは証跡が揃うまで拒否する。以下の「旧21表拒否」は候補22表profile内だけの互換条件に改訂し、既定21表には適用しない。初版の設計経緯は以下に保持する。

> **2026-10-14 ユーザー決定 A（fresh-only）:** `medicines-22-candidate` の target は clinic band 全体が空である fresh な検証専用targetに限る。resume preflight と表別commit apply は target schema 検査の直後に band 全空を要求し、正当な candidate 途中 prefix であっても受理しない（`CUTOVER_REF_BAND_OCCUPIED` で拒否、Begin/COPY 前に deny）。legacy21 formal の既存 prefix resume は変更しない。自動的な delete/rollback/reset/recovery は行わない。候補の再実行には別途用意した空の隔離 target と承認済み手順が必要である。candidate formal activation gate と producer receipt/pin の BLOCKED は変更しない。

この採用はproducer正式artifact/full-file SHA・MIG-34/37/41の4院receipt・consumer互換確認・本番切替/回復承認を代替しない。producer pinは更新しない。数量は§3.2のproducer明示1.0を維持し、consumerは空欄を拒否する（model/APIをnullableへ変更しない）。targetのnumeric(10,2)とvalidated inventory FKを検査する。実CSV投入・migration・deploy・mergeは今回実施しない。

実装・実行手順の正本: [CLINIC_CSV_IMPORT.md](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md)。

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

## 8. 【PROPOSED / UNAPPROVED】会計明細の履歴税率 `tax_rate`（EMR-364 連携案）

**状態: 提案のみ・未承認・未実装。** 以下はすべて producer / consumer 両担当が同一 artifact の receipt で合意するまで採用されない草案。現行の取込契約・CSV・manifest・契約ハッシュ・DB を一切変更していない。EMR-364 が consumer 側 owner、producer 側は MIG-44、生成は MIG-41、公開台帳/pin は MIG-37 が担う。

### 8.1 現行動作（証拠は [CURRENT-PATHS.md §1](../emr-364-365-contract-design/CURRENT-PATHS.md)）

現行の `billing_items` 契約は 10 列で `tax_rate` を持たない（`backend/internal/csvimport/cutover_contract.go:256`、producer `stage-csv-columns.mjs:415-417`・`:553-559`）。COPY は契約列だけを渡すため `tax_rate` は供給されず、nullable の DDL 既定 `0.10`（`backend/migrations/001_init.sql:1921`）が全移行行に入る。精算日由来の 5%/8% は輸送上区別できず、以後この列を読む経路は保存値 0.10 を見る。兄弟の金銭表（`procedures`、`merchandise_items`、`estimate_items`）は既に `tax_type` 直後に `tax_rate` を持つため、`billing_items` だけが欠けている。

### 8.2 提案する 11 列 ordered header

`tax_rate` を `tax_type` の直後に挿入する（他の金銭表と同じ位置関係）:

```
id, clinic_id, billing_id, category, name, unit_price, quantity,
tax_type, tax_rate, is_insurance_applicable, sort_order
```

### 8.3 提案する値域と日付写像

| 条件（canonical 精算日） | `tax_rate` |
| --- | --- |
| 精算日 < 2014-04-01 | `0.05` |
| 2014-04-01 〜 2019-09-30 | `0.08` |
| 2019-10-01 以降 | `0.10` |
| 精算日が無い（NULL） | `0.10` |
| 精算日が非 NULL だが不正・変換不能 | **reject / needs_review**（黙って既定へ落とさない） |

- 列挙は `{0.05, 0.08, 0.10}` の小数割合、required / non-null。CSV セルは 2 桁小数の決定的テキスト（`0.05` / `0.08` / `0.10`）とし、producer は missing 日付でも `0.10` を**明示出力**する（consumer 側既定への依存をしない）。
- 日付判定は canonical 精算日の**日本ローカル暦日**（Asia/Tokyo の暦日）。UTC 変換した日付で境界をずらさない。境界ケース 2014-03-31 / 2014-04-01、2019-09-30 / 2019-10-01 を合成 fixture で固定する。
- `tax_type = 'exempt'` を理由に `tax_rate` を 0 にしない。非課税行にも日付由来の率値をそのまま出す。`tax_type` の復号（`TZei_Flg` 0・空白→excluded、1→exempt、その他→needs_review）は MIG-38/MIG-44 の決定で、本列とは別の契約面。

### 8.4 提案する version・pin・互換性

- 新しい version 名・compiled pin は**採用時に producer / consumer が同一正本 artifact へ記録する**まで仮置きにとどめる（候補名例: `animalekarte-cutover-v2` — bundle 全体の profile 改訂名であり、billing `tax_rate` 単独の version ではない。名称自体も未承認）。この文書は version の採用を主張しない。
- pin = **更新後 producer source ファイル全体の SHA-256**（`scripts/lib/stage-csv-columns.mjs` 全バイト。ヘッダ部のみ・推定値は不可）。現行の compiled pin `cutoverCSVContractSHA256 = 9e02897b…`（`cutover_contract.go:19`）は baseline であり、現 producer source 全体の観測 SHA `8a466c63…`（MIG-40 記録）とは異なる。**順序付き列が不変でも、値写像の意味変更（税復号の向き・負単価の NULL 化等）で全ファイル digest は変わる**ため、pin は列形状ではなく source 全体を縛る。
- 旧 10 列 bundle は**別の legacy profile** として扱い、新 profile では拒否する（欠列・未知列・旧 pin を新 profile として黙って受理しない。現行の header 検査 `cutover_contract_validate.go:609-632` は未知列・欠列を fail-closed で拒否する既存機構と整合する）。legacy 互換モードを設けるかは別途承認が要る receipt 項目。
- 現行の rehearsal 容認（`rehearsalContractDrift`、`cutover_contract.go:56-58`）は REHEARSAL_ONLY artifact かつ rehearsal provenance のときだけ manifest digest 差異・tables[] 順・CSV header 順の bounded drift を記録付きで許す**既存機構**であり、新 profile の採用や互換宣言ではない。10→11 列は列集合が変わるため permutation tolerance の範囲外で、rehearsal でも受理されない。
- §3.7 の `medicines` 追加は表の追加、`tax_rate` は既存表の列追加という別の変更面。両方を含む cumulative profile（22 表 + 11 列）にするか、medicines-only profile を先にして後続 revision とするかは**採用順序の owner 決定**であり、この文書は選択しない。どちらでも単一正本の version/pin/artifact receipt で lockstep にする（EMR-364 scope「EMR-343 と別の独立 contract を作らない」に従う）。
- MIG-15 最新コメントの「`treatments` を 22 表目にする」提案は**ここでは採用しない**。本票の構造提案は medicines-only の 22 表であり、treatments を含む別の 22 表構成や、さらに大きな cumulative profile は将来の owner receipt が必要な別プロファイルとして保持する（MIG-15 自体の設計は別票・未解決 Q1〜Q16）。

### 8.5 提案する検証（採用後の合成 contract test。**全件 planned / NOT RUN**。現時点は未実装）

セル値は**厳密な lexical 一致**のみ受理し、consumer は既定値への coercion・丸め・数値正規化を行わない（許容集合はバイト列 `0.05` / `0.08` / `0.10` の 3 通りのみ）。

| # | 検査 | 期待 |
| --- | --- | --- |
| T1 | 境界日 2014-03-31 / 2014-04-01 / 2019-09-30 / 2019-10-01（日本暦日） | `0.05` / `0.08` / `0.08` / `0.10` |
| T2 | 精算日 NULL | `0.10`（producer が明示出力） |
| T3 | 非 NULL の不正・変換不能な日付（producer 側 source） | producer 側で行 reject または needs_review（黙って `0.10` にしない） |
| T4 | `tax_type=exempt` 行 | 日付由来の率値を保持、`0` にならない |
| T5 | 旧 10 列 bundle / 欠 `tax_rate` / 未知列 / 旧 pin | 新 profile として reject |
| T6 | producer source 全体 SHA-256 | manifest/contract pin と一致（ヘッダのみ hash 不可） |
| T7 | UTC 境界をまたぐ時刻の精算日時 | 日本暦日で判定（UTC 換算日でずらさない） |
| T8 | rate セルの reject 集合: `''`（空）・未知値・`0.09`・`0.101`・`NaN`・`1`・`0.1`・`0.080` 等の非 canonical 表記 | 一律 reject。blank は missing-date の producer 既定 `0.10` とも無関係に必ず拒否（consumer は既定値を合成しない） |
| T9 | 許容 3 値の bytes → preflight → apply → verify 往復 | `0.05`/`0.08`/`0.10` が各境界・exempt 行ともに変換・失われず保持される |
| T10 | manifest の `csvContractSha256` / `stageMappingSha256` が旧 pin のまま | reject（rehearsal drift の対象でもならない） |

### 8.6 保留の owner 決定・receipt

- version 名・ordered 11 列・nullability・列挙 `{0.05,0.08,0.10}`・pin（全ファイル SHA-256）の同一 artifact 合意（provider MIG-44 + consumer EMR-349 系）。
- 旧 10 列 bundle の扱い（拒否のみか、legacy profile の併存許容か）。
- medicines-only profile と `tax_rate` を含む profile の採用順序（§8.4）。
- MIG-15 treatments 22 表案の可否（本票では未承認・別 receipt）。
- producer source の採用基点（MIG-37 統合候補 → MIG-41 単一生成の同一 candidate/profile/run）。

## 9. 【PROPOSED / UNAPPROVED】ワクチン単価の小数精度（EMR-365 連携案）

**状態: 提案のみ・未承認・未実装。** 現行の `vaccines.price bigint`・Go `*int64`・JSON number・CSV `bigint` 経路を一切変更していない。小数対応の採用自体が未承認であり、採用時は §8 と同じ単一正本 artifact の version/pin で producer/consumer lockstep に合意する（bundle 全体の cumulative profile、候補名例 `animalekarte-cutover-v2` — 未承認）。

### 9.1 現行経路（証拠は [CURRENT-PATHS.md §2](../emr-364-365-contract-design/CURRENT-PATHS.md)）

stage は `NULL::bigint AS price` を出し、CSV 空セルは `NULL ''` で SQL NULL、`vaccines.price bigint`（nullable）、Go/JSON は `*int64`、FE は `?? 0`・`parseInt`・number input、未請求 candidate は `UnitPrice int64`・Quantity 1・0.10 固定、合計は int64 円。**現行のどの段でも小数は保持できない**。現行の FE/BE 丸め差分の記録は CURRENT-PATHS §2.9（本票では是正しない）。

### 9.2 MIG-31 の既存決定（保持・変更しない）

legacy source の価格写像ルールは MIG-31 で決定済みであり、この案は変更を提案しない:

- blank/1（external-tax）→ 旧単価そのまま（**÷1.1 は行わない**、旧提案撤回済み）。
- **0・負・NULL・空・リンク無し・未知税・2/3（exempt）→ bare 空 CSV セル → 新 DB では NULL**。
- 正の小数は正のまま保持（0.04 等は値として消さない）。
- 製品 API による新規保存の明示 `0`（マスタに「0 円」を意図入力）は、legacy source の `0`（→ NULL 行）とは別の事象であり、混同しない。

### 9.3 alternatives（すべて PROPOSED / UNAPPROVED）

| 案 | 保存型 | 輸送 | 精度/範囲 | 欠点 |
| --- | --- | --- | --- | --- |
| A（推薦） | `numeric(19,4)` | CSV: canonical decimal 文字列。JSON: **decimal 文字列**（number 不可）。Go: 正確な decimal 型 | 整数部 15 桁・小数部 4 桁。非負の価格として `[0, 999999999999999.9999]` | DDL/Go/API/FE/税計算の全面更新が要る。`bigint` 既存行の移行が要る |
| B | `bigint` 最小単位（円×10⁴ 等の scaled integer） | CSV/JSON とも整数 | int64 範囲内。scale-4 に正規化すると `999999999999999.9999` 相当は `9.9×10¹⁸` で **int64 上限（約 9.22×10¹⁸）を超えうる** — 境界・乗算オーバーフローの check が必須。非 scale の整数部（最大 ≈1.0×10¹⁵）は JS 安全整数 `2⁵³−1`（≈9.007×10¹⁵）未満だが、**×10⁴ の scaled 値は安全整数を越える**ため JS Number 経由は scaled 形でも不可 | scale の意味を全実装が正しく解釈する保証が弱い。円列との混同で倍率バグ |
| C | `numeric(19,2)` | A と同じ | 小数部 2 桁まで。**`0.04` は正確に表現できる**（scale 2 で足りる）。表現不能は `0.004` のような小数第 3 位以降 | legacy raw が小数第 3 位以降を持つ場合に保持不能 — 削る理由は「その桁が来ない」ことの確認のみ |
| D | 現行 `bigint` のまま表示だけ小数第 1 位 | 変更なし | 整数円のみ | 0.4/0.04 を保存できない。表示丸めで「正 → 0/無料」に見せる危険（MIG-31 の正値保持に反する）。棄却 |
| E | `bigint` 円 + 別列 `unit_price_fraction` | 2 列 | 組合せ依存 | 意味が複雑で結合が要る。A に対する明確な利点が無い。棄却 |

legacy `unit_price_raw` の numeric は**任意 scale であり、numeric(19,4) への変換は自動的には無損失ではない**。範囲外・超過桁の行は producer/consumer で reject / needs_review とし、**丸め・切捨てで収めない**。

### 9.4 推薦と輸送・計算の境界（A 案を owner レビューへ出す。採用しない）

- **保存**: `vaccines.price numeric(19,4)`。理由: legacy numeric の観測小数（0.4/0.04/0.5 等）を余裕をもって保持し、float64 の binary 誤差・途中の int64 truncate を全段で避けられる。乗算 `0.4×3` が decimal 算術で厳密に `1.2` になる。
- **CSV**: plain canonical decimal 文字列（例 `0.4000`）。COPY は `''`→NULL を維持し、数値文字列は lexical 検査のうえ型変換する。
- **JSON**: **decimal 文字列**（`"0.4000"`）。JSON number は binary64 由来で 0.4 系を厳密に表せず、大きな整数部では安全整数を失うため不可。さらに大きな magnitude では**小数部も保持できない** — `90071992547409.1234`（整数部 ≈9.0×10¹³、2⁵³ 未満・範囲内の受理値）では binary64 の ULP が約 0.0156 で 10⁻⁴ を超えるため、JS Number では小数第 4 位までが既に失われる。文字列なら 4 桁全てを保持する。
- **Go**: 正確な decimal 型を採用（`*int64`/`*float64` ではない）。検証は非負・範囲・scale を型レベルで持つ。
- **FE**: 表示・入力とも **JS Number への parse を経ない** decimal 表現で扱い、数値演算は承認された円丸め境界までは行わない。`parseInt`/`Number()`/`int64` 経由の入力は新契約では無い。
- **計算**: 単価の丸めは**乗算の前に行わない**。小計・税・最終円の丸め順・丸めモード・丸め段階は owner が規則を確定してから固定する（本票は具体値を捏造しない）。
- **影響範囲**: 共有の請求経路（`billing_items.unit_price bigint`、`quantity numeric(10,1)`、`model.BillingItem.CalculateTaxAmount` の float64、`CalculateBillingTotals` の int64 合計、payments の整数円）をまたぐ変更が要ることを影響として記録する。**他マスタの DB/API 型はこの案の scope 外**（全マスタ decimal 化を無断に拡張しない。別承認の scope が要る）。

### 9.5 表示（owner 決定・未承認。UI 変更なし）

- 表示候補は小数第 1 位まで（user 要望）。`0.4` は `0.4` と表示する（`0.0` に潰さない）。
- `0.04` のような正の小値は、1 桁表示が `0.0` になる場合は**表示を `0.04` に落とす**か「正・1 円未満」の明示 marker を出す。`0.0`/`-`/無料への推論は禁止（MIG-31 の正値保持と整合）。
- 保存 scale（4）と表示桁（1 または marker）は別の決定であり、表示丸めで保存値を書き換えない。

### 9.6 匿名ケースの固定期待値（**全件 planned / NOT RUN**、採用後の contract test 前提）

| # | 入力（匿名） | 期待（推薦 A） | 補足 |
| --- | --- | --- | --- |
| 1 | `0.4` | 保存 `0.4000`、表示 `0.4` | 0 化禁止 |
| 2 | `0.04` | 保存 `0.0400`、表示 `0.04` または marker | `0.0` → 無料推論禁止 |
| 3 | `0.5` | 保存 `0.5000`、表示 `0.5` | |
| 4 | 整数 `500` | 保存 `500.0000` | 後方互換の整数 |
| 5 | source `0` | 空 CSV → `NULL`（MIG-31 既存決定） | 製品 API の明示 `0` は別ケース（5b） |
| 5b | 製品 API 明示 `0` | `0.0000`（明示ゼロ、legacy source 0 の NULL とは別） | owner が受理規則を決める |
| 6 | source 負値 | 空 CSV → `NULL`（MIG-31 既存決定） | MIG-40 の別件負値議論と独立 |
| 7 | NULL / 空セル / リンク無し / 未知税 / 2・3（exempt） | 空 CSV → `NULL`（MIG-31 既存決定） | unbilled の master_unbillable warning に該当 |
| 8 | `0.4` × quantity `3` | 小計 `1.2`（decimal 厳密）→ 税・円丸め規則は owner | 単価は乗算前に丸めない |
| 9 | 最大 `999999999999999.9999` | 受理（範囲内・整数部 15 桁） | numeric(19,4) の上限 |
| 10 | `999999999999999.9999 + 0.0001` 超過 | reject / needs_review | 範囲外は丸めて収めない |
| 11 | `0.00001`（小数第 5 位） | reject / needs_review | scale 超過は切捨てない |
| 12 | `90071992547409.1234`（範囲内受理値） | decimal 文字列で往復、4 桁全保持 | JS Number では binary64 ULP > 10⁻⁴ となり小数第 4 位が失われる根拠ケース（整数部は 2⁵³ 未満） |
| 13 | CSV 往復 `0.4` | `0.4000` で保存・読み戻し（現行では不可） | 現行 COPY 失敗パスの置換対象 |

「税額・最終合計の期待値」は owner が丸め順・丸めモードを確定するまで固定値を作らない（EMR-365 受入の条件）。

### 9.7 保留の owner 決定・receipt

- 小数対応の採否自体（本票では推薦のみ・未承認）。
- 保存型・scale・表示桁・JSON 表現（decimal 文字列）・数量の扱い・税/合計の丸め順・旧整数 bundle 互換の同一 artifact 合意。
- shared billing `unit_price`/`quantity`/合計経路への波及をどこまで同一 revision に含めるか（scope 境界）。
- 他マスタ・他会計明細への decimal 化は別承認の scope。
