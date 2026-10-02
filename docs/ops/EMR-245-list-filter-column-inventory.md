# EMR-245 — 一覧フィルタと表示列の対応棚卸し

> **対象チケット**: EMR-245「一覧のフィルタを、表示されている列に揃える」/ 作成: 2026-10-02 / 基準コミット: `608a5f55a` 派生 worktree (`/private/tmp/ae-emr-245-r2`)

## 背景

カルテ一覧の従来フィルタは「横断検索 (`search`)」が JOIN した複数テーブルへ OR で
部分一致する構造で、データ規模が大きくなると高コストな全行評価を誘発し得た。
EMR-245 は **表示されている列に対応する独立した条件としてフィルタを提供**し、
各条件を AND 結合の半結合 (EXISTS) で評価する方式へ揃える。
本書は全リスト画面について「表示列 ⇔ フィルタ可能フィールド」の対応表を記録し、
カルテ一覧以外のギャップをフォローアップ対象として残す。

## 用語

- **server**: フィルタ値が API クエリパラメータとして送信され、ページネーション・件数もサーバ側で確定する。
- **hybrid**: 一部条件のみサーバ送信し、残りは取得済みページ内でクライアント側絞り込み（**ページスコープ**: 現在ページ内にしか効かない条件あり）。
- **client**: サーバは範囲取得（日付・上限件数など）のみで、表示条件はすべてクライアント側。
- **–**: 該当 UI を持たない。

## 本ユニットで対応した一覧（カルテ一覧）

`frontend/src/features/medical-records` / `backend/internal/medicalrecord`

| 表示列 | フィルタ | 方式 | クエリパラメータ |
|--------|----------|------|------------------|
| 診療日 | 診療日（期間） | server | `start_date` / `end_date` |
| 飼主名 | 飼主名（含む） | server | `owner_name`（name/name_kana 畳込み ILIKE、255 byte 上限） |
| ペット名 | ペット名（含む） | server | `pet_name`（同上） |
| 種 | 動物種（完全一致） | server | `animal_species_id` |
| 主訴 | 主訴（含む） | server | `chief_complaint`（inquiries.chief_complaint ILIKE） |
| 関連 | – | – | リンク列（カルテ/会計への遷移）。フィルタ対象外 |
| 担当医 | 担当医（完全一致） | server | `doctor_id` |
| ステータス | ステータス（完全一致） | server | `status` |
| 医院 | 医院スコープ | server | `clinic_ids`（ClinicScopeFilter。複数医院運用時のみ列表示） |
| 操作 | – | – | 行アクション列 |
| （列外） | 薬・処置・診察・在庫 | server | `medicine_id` / `procedure_id` / `consultation_id` / `inventory_id`（マスタID完全一致。既存温存） |
| （列外・例外） | 診療内容 横断検索 | server | `search`（カルテNo・飼主名・ペット名・主訴の OR 検索。列フィルタと AND 併用可、UI 上「横断検索」と明示して区別） |

- 新設の text 型フィルタは `PropertyFilter` で `condition: "contains"` 固定（条件選択ステップなし、空白のみ入力は拒否、値ピルから再編集可）。
- いずれの検索・フィルタ変更も `useMedicalRecordsUrlState` の resetKey で page=1 へリセット。
- ページネーション・件数は従来どおりサーバ側（`buildBase` で count/取得が同一クエリ基盤）。

## 一覧の棚卸し（カルテ以外）

凡例の列: **表示列**（主要列）→ **フィルタ可能フィールド** → **方式** → **ギャップ / 状態**。

### 飼主・ペット一覧 `features/owners` — server

- **表示列**: 飼主No / 飼主名 / （医院）/ ペット番号 / ペット名 / 生死 / 種 / 生年月日 / 体重 / 環境 / 前回来院 / 操作
- **フィルタ**: `search`（飼主名・ペット名等の横断）, `species`, `include_deceased`（生死）, `checkup_history`（健診受診履歴）, `clinics`。loader 経由で全てサーバ送信、ページネーションもサーバ側。
- **ギャップ**: 飼主名 / ペット名 / 生年月日 / 体重 / 環境 / 前回来院 の列対応フィルタなし（横断検索で飼主名・ペット名は部分的に代替可）。**follow-up**

### 健診一覧 `features/checkups` — hybrid

- **表示列**: 実施日 / 飼主名 / ペット名 / 健診種別 / 次回予定 / 結果・所見 / 担当医 / 操作
- **server**: 実施日（`start_date`/`end_date`）、期限状態（`next_start_date`/`next_end_date`）、`pet_id`、page/limit。
- **client（ページスコープ）**: 検索バー、動物種（`filterCheckupsBySpecies`）。
- **ギャップ**: 飼主名・ペット名・健診種別・結果・担当医は検索バー頼みのページ内絞り込み。**follow-up**

### 見積一覧 `features/estimates` — client

- **表示列**: 見積番号 / タイトル / 飼主名 / 有効期限 / 合計金額（+ ステータス列）
- **フィルタ**: ステータス、有効期限、検索バー — すべて `filterAndSortEstimates` でクライアント側。`useGetEstimates()` はパラメータ未指定（status/owner_id/medical_record_id は型上存在するが一覧から未使用）で先頭ページのみ取得し、その範囲内で絞り込み・ページングする。
- **ギャップ**: 列対応フィルタなし・取得範囲外の行には一切フィルタが効かない。**follow-up**

### 検査一覧 `features/examinations` — hybrid

- **表示列**: 日時 / 飼主名 / ペット名 / 検査種別 / 結果概要 / 担当医 / ステータス / 操作
- **server**: 日付（`start_date`/`end_date`）、`pet_id`、page/limit（BUG-411 でページネーションはサーバ化済）。
- **client（ページスコープ）**: ステータス、検査種別、担当医、検索バー（`CLIENT_ONLY_FILTER_KEYS`）。
- **ギャップ**: 飼主名・ペット名・検査種別・結果概要・担当医・ステータスの列対応 server フィルタなし。**follow-up**

### 入院一覧 `features/hospitalization` — hybrid

- **表示列**: 入院No / 飼主名 / ペット名 / 種 / タイプ / 担当医 / 入院開始日 / 退院予定日 / ステータス（+ ケージボード view）
- **server**: ステータスタブ（入院中/予約/退院済）、入院日期間（`startDate`/`endDate`）、page/limit。
- **client（ページスコープ）**: 検索バー（飼主名・ペット名・入院No）、入院区分、種（`applyHospitalizationClientFilters`）。
- **ギャップ**: 飼主名・ペット名・タイプ・担当医・退院予定日の列対応 server フィルタなし。**follow-up**

### 在庫一覧 `features/inventory` — hybrid

- **表示列**: 品名 / カテゴリ / 在庫数 / 最低在庫 / 保管場所 / 有効期限 / ステータス / 操作
- **server**: 検索バー（`search`）、カテゴリ `is`、ステータス `is`、page/limit（BUG-412）。
- **client（ページスコープ）**: カテゴリ・ステータスの `is_not` 除外、列ソート（`excludeInventoryItems`/`useSortableData`）。
- **ギャップ**: 保管場所・有効期限の列対応フィルタなし。`is_not` 条件はページ内のみ有効。**follow-up**

### 予防接種一覧 `features/vaccinations` — hybrid

- **表示列**: 実施日 / 飼主名 / ペット名 / 予防接種名 / 次回予定 / 操作
- **server**: 日付（`start_date`/`end_date`）、検索バー（`search`）、上限 `HISTORY_FETCH_LIMIT` の範囲取得（BUG-502）。
- **client（ページスコープ）**: 担当医フィルタ、検索語の重複適用（`useFilterVaccinations`）、件数/ページングは取得済み範囲内。
- **ギャップ**: 飼主名・ペット名・予防接種名の独立フィルタなし（横断検索が代替）。**follow-up**

### トリミング一覧 `features/trimming` — client 主導

- **表示列**: 診療日 / 飼主名 / ペット名 / 種 / 犬種 / 体重 / スタイル希望 / ステータス / 担当 / 操作
- **server**: 日付範囲（`start_date`/`end_date`）、`pet_id`、上限 `HISTORY_FETCH_LIMIT` の範囲取得のみ。
- **client（ページスコープ）**: 検索バー（飼主名・ペット名・犬種）、ステータス、種、担当（`useFilterTrimmingRecords`）、ソート。
- **ギャップ**: 表示列の大半がクライアント側のみ。件数表示も取得範囲内の値。**follow-up**

### 会計 `features/accounting` — server + server タブ

- **会計一覧タブ 表示列**: （医院）/ 日時 / 飼主名 / ペット名 / 請求金額 / 支払方法 / ステータス / カルテ / 操作
- **server**: `search`、`status`（+`status_op`）、`payment_method`（+`payment_method_op`）、`start_date`/`end_date`、`clinic_ids`、page/limit。
- **当日会計タブ**: `daily_date` 指定でサーバ取得。未納タブ: `start_date`/`end_date`/`group_by` + page/limit サーバクエリ。
- **ギャップ**: 飼主名・ペット名の独立列フィルタなし（横断 `search` が代替）。カルテ列はリンク。**follow-up（低）**

### 受付（予約ボード）`features/reception` — client 補助

- **表示**: 日別タイムライン／ボード（テーブルではないが行一覧性あり）。
- **server**: `date` + 上限取得（`DAY_VIEW_FETCH_LIMIT`）。
- **client**: フィルタパネル（来院タイプ複数選択、指名/担当医、トリミングのみ表示）。
- **ギャップ**: ボード形式のため列フィルタという概念はないが、担当医等は当日分内のみ。**follow-up（低）**

### マスタ設定一覧 `features/master` / `features/clinic-settings` — client

- `MasterCRUDPage` 系のマスタテーブル（動物種・薬・処置・診察・商品・職種・支払方法・ケージ・保険・予約種別・権限グループ・スタッフ・入院種別・健診種別・検査種別・トリミング設定・診断名・治療計画・面談テンプレート・ラボ項目・キャンペーン・LINE予約枠等）。
- **フィルタ**: `use-master-crud` が取得済み項目に対して `searchTerm`（既定は名称一致）と `filterProperties` をクライアント側で適用。
- **方式**: マスタは小規模・全件取得前提で client-side。表示列≒名称・ステータスでフィルタは列に概ね対応済み。
- **ギャップ**: 列ごとの個別フィルタはなし（必要になった時点で検討）。**follow-up（対象外寄り）**

### サポート: バグレポート一覧 `features/support/routes/BugReportsPage.tsx` — なし

- **表示列**: 日時 / 報告者 / 件名 / 画面 / スクショ / 状態
- **フィルタ**: なし（詳細選択のみ）。件数も小さい想定。**follow-up（必要時）**

## フォローアップ整理

| 対象 | 優先度 | 理由 |
|------|--------|------|
| 飼主一覧 | 中 | 行数が大きく列とフィルタの乖離あり。`search` 兼用の分離化候補 |
| 健診 / 検査 / 入院 / 在庫 / 予防接種 / トリミング | 中 | hybrid — ページスコープの client 条件が存在し、全件に効かないフィルタが残る。特に見積・トリミングは取得範囲自体が限定 |
| 会計一覧 | 低 | server 化済み。列対応の独立フィルタ追加のみ |
| 受付ボード | 低 | 日別ビューのため列フィルタ概念が薄い |
| マスタ設定 | 低 | 小規模・全件取得前提。現状 client-side で実用上問題なし |
| バグレポート | 低 | フィルタ UI なし・件数小 |

## 共通インフラメモ

- 列テキストフィルタは `PropertyFilter` の新 `type: "text"`（`contains` 固定、空白拒否、Enter/適用ボタン確定）を再利用する。
- BE 側は `textsearch` の folded 式（`FoldedExpr`）+ パラメータ化 `ILIKE ? ESCAPE '\'` を使い、JOIN+OR を避け EXISTS 半結合で組み立てるのが本ユニットの定型パターン。他一覧へ展開する場合も同型を踏襲すること。
