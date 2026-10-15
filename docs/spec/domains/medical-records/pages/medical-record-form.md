# カルテ詳細・入力 仕様書 (Medical Record Form)

[17:17合意の診療前日付訂正](../../reservations/pages/reservation-management.md#requirement-fr-reservations-preconsultation-date-record-correction)では、未記入カルテは確認して予約日へ合わせ、記入済みなら通常の日付変更を停止して内容確認の訂正へ渡す。以下の作成後日付固定・直接PATCH拒否は現行実装／通常更新の参照で、この合意の未記入・診療前の確認付き訂正を実装済みとする根拠ではない。確定済みカルテの変更禁止・追記による訂正、保存ルーティング契約は維持する。

17:23 JSTの[LINE予約6点合意](../../../line/reservation-spec.md#line-agreement-1723)へ従う。公開された診療・トリミング・ホテルは予約対象、入院は医院相談。ホテルには医師指名・医師シフト・通常予約の同日制約を適用しない。受付制限・入力保持・受付前取消・必要設定不足の扱いと受入条件はLINE正本を参照する。ホテル受入基準は17:35合意の医院別同時宿泊頭数・全宿泊期間判定で、実装・実機受入は未確認。

17:17 JSTの院内予約8点一括合意は[予約入口の決定一覧](../../reservations/README.md#internal-reservation-batch-1717)を参照する。誤取消／誤終了訂正、未終了予約保護、診療前の日付・カルテ整合、マスタ利用・名称履歴、休診日予約／日跨ぎ勤務、患者照合、ホテル担当時点／入院記録失敗の各正本と受入条件へ従う。LINE受付条件は17:23合意（ホテル受入基準は17:35合意、計算・保存方式は設計）、実装・実機受入は未確認。

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.medical-records.medical-record-form／画面仕様 |
| 決定状態 | 2026-10-14初回5項目は部分合意。最新の業務正本はlifecycle初回合意節。対象外の既存決定状態を保持し、カルテ全体完了／実装受入にしない |
| 実装状態 | 合意と現実装は別。main ec4f7146a の必要箇所を静的照合し、差分は予約正本の残差一覧へ記録。実装変更なし |
| 実機検証状態 | 今回未実施。過去の記録と今回の文書確認を分ける |
| 文書確認日 | 2026-10-14（合意・参照・語彙の確認。実機受入日ではない） |

<a id="purpose"></a>
## 目的・対象範囲

| シナリオ | 業務上の価値 | 優先度・理由 |
|:---|:---|:---|
| 主訴区分を空欄で保存する | 任意項目を無理に選ばず記録する | 未決／要確認（既存文書に優先度の根拠なし） |

### 概要

初診のwelcome／follow-upは[D15の確認付きフォロー](../../../line/lstep-integration.md#crm-followup-confirmation)に従い、カルテ作成だけでは送信しない。スタッフが実施済み・送信対象を確認し、一般のお礼と医療的再検査／受診案内を区別する。後者は医師の要フォロー指定がある場合だけとし、担当医選択・カルテ保存／確定をその指定の代用にしない。現行create副作用の実装説明はこの目標への適合証明ではない。既存保存§2・認可／必要監査と結果不明照合を保持し、通知失敗からカルテを作り直さない。受入はD15のAC-crm-followup-record-only／medical／confirmation-failure／duplicate-unknownへ接続する。

- **画面の目的**: 臨床現場における電子診療録（SOAPS形式）の作成、検査、処置、処方、および会計連携の統合管理。

<a id="access-preconditions"></a>
## 利用者・権限・前提

- **アクセス権限**: 新規作成 (`/medical-records/new`) は `medical-records:create` を要求する `RequirePermission` ルートガードあり。編集 (`/medical-records/:id`) は親ルート `/medical-records` の `medical-records:view` ガードのみを継承し、保存可否はコンポーネント内の `usePermission`（`ResourceMedicalRecords`）で制御。

担当医以外の権限あるスタッフの入力、医院が認めた医師の確認・確定・訂正確認と実編集者ログは[初回合意](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)。実grantと医師判定の対応は要確認。操作前提の詳細は本書の各操作・保存契約と関連正本を参照する。

<a id="navigation"></a>
## 入口・出口

- **URLパターン**: 
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`

| 入口／出口 | 契機 | 引継ぎ情報 | 根拠 |
|:---|:---|:---|:---|
| URLパターン | 画面を開く | 上記URL／既存導線。呼出元の網羅は要確認 | 本節の元記載 |
| 戻り先・他画面 | 本節の既存操作 | 記載が無い引継ぎ・戻り先は要確認 | 本節／関連フロー |

診察予約の通常カルテは[受付時の作成契約](../../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity)へ変更する。既存の新規画面表示・保存操作は実装参照を含み、合意への追従は未確認。予約変更は[予約管理へ集約](../../reservations/pages/reservation-management.md#requirement-fr-reservations-edit-entry)し、カルテから予約を変更しない。通常新規と過去未登録診療は[初回合意の受付接続／専用入力](../lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry)へ従う。ホテル・入院・トリミング専用記録は別ownerの内部仕様。

### 死亡登録後の生前診療記録

[死亡後適用の最新合意](../lifecycle.md#medical-record-deceased-retrospective)に従い、生前診療の後日入力・誤記訂正を死亡状態を戻さず理由と履歴付きの専用経路で扱う。既決の過去入力・医師確認・当日カード／自動課金なしを維持し、死亡後の新しい診療や未来予約／実施は許可しない。通常フォーム／選択画面の死亡拒否は現行差として区別し、専用UI・保存対応を実装済みとしない。退院・精算はD05/D10へ渡す。

<a id="display-inputs"></a>
## 表示情報・入力項目

| 項目 | 意味 | 型／形式 | 必須／任意 | 初期値 | 候補元 | 制約／根拠 |
|:---|:---|:---|:---|:---|:---|:---|
| 主訴詳細・治療方針・診断 | 問診・SOAPSの記録 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 入力・履歴複写 | 下記「SOAPS 形式の採用」、[保存契約](../lifecycle.md#hybrid-sync) |
| 主訴区分 | 主訴の分類 | nullable（*uint64） | 任意・空欄可 | 未記載／要確認 | 主訴区分マスタ | [画面保存操作](#22-保存プロセス) |
| 担当医・来院種別・診療日 | 患者ヘッダーの記録属性 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 既存値、候補の詳細は要確認 | [画面保存操作](#22-保存プロセス) |
| 次回予定 | 日付のみの未予約意図 | intended_date（日付） | 未記載／要確認 | 未記載／要確認 | 次回予定ボタン | [次回予定契約](../../../workflows/reservation-to-record-flow.md#next-visit-intent) |
| 画像・資料 | 診療資料の添付 | JPEG／PNG／GIF／PDF | 未記載／要確認 | 未記載／要確認 | ファイル選択 | 下記「画像・資料の選択制限」 |
| 訂正内容・訂正理由 | 確定後の追記 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 入力 | 下記「臨床安全ガード」（理由500文字以内） |
| 治療明細・投与逸脱理由 | 処置・処方の記録 | 未記載／要確認 | 逸脱理由は条件付き必須 | 未記載／要確認 | 医院マスタ・履歴複写 | 下記「臨床安全ガード」、処置マスタの節 |
| 他タブの入力 | 予防接種・健診・検査・見積・会計確認 | 未記載／要確認 | 未記載／要確認 | 予防接種は種類別設定または手入力へ統一（具体間隔・初期値は追加確定しない）。他は要確認 | 下記タブ表の関連仕様 | 独立画面・API定義を複製しない。詳細は既存関連仕様／要確認 |

### 1. 画面構成 (9 タブ構成)

膨大な診療情報を整理するため、以下の 9 つのタブで構成されています。

| タブ | 名称 | 内容 |
|:---:|:---|:---|
| 1 | **問診** | 飼主からの主訴、現在の症状、生活環境のヒアリング記録。 |
| 2 | **診察/治療プラン** | 身体検査所見（SOAPのS/O/A）、診断名（第1・第2）、今後の治療方針（Plan）。 |
| 3 | **治療** | 実施した処置、処置項目、および処方薬の明細。 |
| 4 | **予防接種** | 左ペインに当該ペットの接種記録一覧（`GET /vaccinations?pet_id=` + `page`/`limit`）。空なら EmptyState、「記録を追加」で入力フォーム。右カラムは `VaccinationHistory`。カルテネストの `/medical-records/:id/vaccinations` は存在せず使わない。次回日は[D08最新合意](../../vaccinations/lifecycle.md#fr-vaccinations-schedule)の種類別設定または手入力へ統一する。現行 `calculateNextDate` 配線とカルテ既定 `4weeks`／独立既定 `1year` は解消対象の実装差であり、目標の固定既定ではない。 |
| 5 | **定期健診** | 健康診断の結果記録。`/checkups` の編集は `?tab=定期健診&checkupId=` で本タブを開き、該当行を編集状態にする。 |
| 6 | **検査** | 検査結果の一覧表示と既存検査の取込（カルテへの紐付け）、基準値との比較（HIGH/LOW判定ハイライト）。[D07最新合意](../../examinations/lifecycle.md#fr-examinations-record)に従い担当医未選択でも保存可、入力者ログは担当医と別に保持する。 |
| 7 | **画像** | 患部写真、レントゲン、エコー画像、PDF資料のファイル選択によるアップロード・管理。ドラッグ&ドロップ追加は[追加3項目合意](#requirement-fr-medical-records-image-drop-selected-print)で業務採用。#235の現実装・価値実測の状態は今回更新／確認していない。 |
| 8 | **見積書** | 提示した概算費用の管理。 |
| 9 | **会計(医師確認)** | 診療費の最終確認と会計ステータスの送信。 |

#### 1.1 スティッキー患者ヘッダー (Sticky Patient Header)

9 タブの上部には `MedicalRecordStickyHeader` が画面上部に sticky 固定で常時表示され、内部の `PatientContextHeader` が患者コンテキストを保持する。タブ切替・スクロールに関係なく以下をクリック不要で確認できる。

- **今回カルテのバイタルチップ**: `useGetVitals` で当該カルテのバイタルを取得し、最新 1 件を `latestVisitVitalChips` が `vitalsSummary` へ変換して `PatientContextHeader` へ渡す。体温 T / 心拍数 HR / 呼吸数 RR / 測定体重（`weightUnit` 付き）を、ペット名行直下の属性行へ**計測時刻を付けずに常時表示**する。当該カルテにバイタル記録が無い、または 4 項目すべて未入力の場合はチップ自体を表示しない。
- **マイクロチップ番号**: ペットの `microchip_number`（props 名 `microchipNumber`）が存在する場合、ペット名の右隣にモノスペースのチップとして表示する（`aria-label` は「マイクロチップ番号 <値>」）。未登録または空白のみの場合はチップを表示しない。
- **次回予定ボタン（`NextVisitButton`、EMR-308）**: ヘッダーの「次回予定」はペットの tentative appointment（`status='tentative'`、日付のみ `intended_date`）を `PUT /v1/reservations/pets/:petId/next-visit-intent` で upsert、`DELETE` 同パス（`reservations:edit`）で取消する。カルテ PATCH ではない。権限は操作別（新規=`reservations:create`・既存 tentative の上書き=`create`+`edit`・確定=`edit`）で、write 契約の正本は [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §2.3、一覧・確定の管理面は [reservation-management.md](../../reservations/pages/reservation-management.md) を参照。

#### 2.1 SOAPS 形式の採用

- **Subjective (主観的)**: 飼主の話。
- **Objective (客観的)**: 視診・触診・聴診結果。
- **Assessment (評価)**: 診断、病名特定。
- **Plan (計画)**: 今後の検査・治療予定。
- **Supplement (補足)**: 注意事項等。

#### 2.4 画像・資料の選択制限

ファイル選択欄の対象はJPEG/PNG/GIF/PDF。一度の選択は10件まで、1ファイル10MiB以下・合計50MiB以下とする。件数・合計サイズの超過、または1件でも個別サイズ超過がある場合、その選択分全体を受け付けずエラーを表示する（`ImageGalleryFilter`）。拡張子・形式の選択補助はファイル内容の安全性検証を保証しない。

#### 2.5 処置マスタのプロビジョニング（皮下点滴）

「治療」タブで選択する処置項目（皮下点滴を含む）は `procedures` テーブルの `clinic_id` 付きレコードであり、**医院ごとのランタイムマスタデータ**として管理される。seed 配布物（backend/migrations/seeds/002_master に procedures の CSV は存在しない）やコード内の静的定義は存在しない。各医院はマスタ設定画面 `/settings/treatment-items` の「処置」タブ（`?tab=procedure`、表示名「処置マスタ」。`ResourceMasterMedical` 権限で保護された `TreatmentPlanMaster`）から `useCreateProcedure` 等で登録・編集する。皮下点滴もこの運用経路で医院ごとに追加するものであり、デプロイ物・seed 変更を伴う導入経路は存在しない。

<a id="operations"></a>
## 操作・業務ルール・状態遷移

| 要件ID | 定義の所在 | 決定根拠 |
|:---|:---|:---|
| FR-medical-records-optional-chief-complaint-type | [既存の要件本文](#requirement-fr-medical-records-optional-chief-complaint-type) | 既存記載を継承。ID付与で合意済みにしない |

| 操作 | 実行条件 | 成功後状態／遷移 | 失敗挙動 | 根拠 |
|:---|:---|:---|:---|:---|
| 保存・確定・訂正 | 権限、保存済み、会計確認状態等 | 保存・finalized・addendum | 未確認状態は確定不可、失敗時挙動は各契約 | 下記「臨床安全ガード」、保存節 |
| 履歴複写 | コピー可能行・編集ロック条件 | 問診ローカル変更／治療明細の個別保存 | 上書き確認の取消で現在値保持 | 下記「履歴複写」 |

### 2. 主要な臨床ロジック



#### 2.3 臨床安全ガード

最新の[入力者・編集ログ](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)、[必須項目追加なし／未完了プラン警告](../lifecycle.md#requirement-fr-medical-records-explicit-finalize-plan-warning)、[削除・訂正](../lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction)、[評価不能時の医師確認・理由付き手入力](../lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual)は初回合意が正本。警告確認後は未完了でも確定可。以下のUI・API説明には現行実装参照を含み、合意へ追従済みではない。

- **確定ロック**: `finalized`（確定済）ステータスのカルテはバックエンドが更新を拒否し（409）、訂正は追記（addendum）のみ許可することで真正性を担保。確定への遷移は `PATCH /medical-records/:id` の `status` 指定によるもの。画面右下のフローティングアクション（`MedicalRecordFloatingActions`）に「確定する」ボタンが表示され（編集権限あり・保存済み・未確定の場合のみ）、会計(医師確認)が `confirmed` の場合だけ有効になる。未確認・差戻し・取得中・取得失敗・確認状態なしでは「会計確認が未完了です」として確定を物理ブロックする。有効時は `MedicalRecordFinalizeDialog` で不可逆であることを確認した上で確定する。確定取り消し（unfinalize）API は存在しないため、確定後の修正経路は訂正追記（addendum）のみ。会計完了時の自動確定は現状存在しない。確定済みカルテはサイドヘッダーに「確定済」バッジ（`StatusBadge`）を常時表示する。
- **ロック中の画像タブ（EMR-216）**: `isLocked`（`recordFinalized || !canSubmit`）は `MedicalRecordServiceTabs` から `MedicalRecordImage` へ伝播し、ロック中は `canUpload`/`canDeleteImage` が false になる。「画像アップロード」「撮影」ボタンと file input を描画せず、ドラッグ&ドロップはドロップガイドを出さず `dropEffect=none` で拒否し、画像の削除ボタンも出さない。UI 非表示に加え callback（`handleFilesSelected`/`handleDrop`/`handleDeleteRequest`/`handleDeleteConfirm`）側でも fail-closed に拒否するため、ロック中の画像 mutation API 呼出はゼロ（外側の disabled fieldset は D&D イベントを遮断できないため、この二重防壁が必要）。画像一覧・画像リンクの閲覧は維持され、検索・期間フィルタ等のフォーム入力はタブ共通の disabled fieldset 配下で他タブと同一の境界規則に従う。ロック解除で全操作が復帰する。
- **訂正追記モーダル**: `AddendumModal` の修正内容・修正理由は controlled input。バリデーション失敗後も入力済みの値を保持する（React 19 `useActionState` の remount で消えない）。修正理由は 500 文字以内。
- **薬量自動計算と絶対上限ゲート（#201）**: 「治療」タブの処方明細（`TreatmentRow`）は、対象ペットの species と当日 vital 体重から数量を自動プリフィルする（`calculateDose`）。保存値がマスタ上限（体重連動上限 weight×max_mg/kg と体重非依存の絶対上限 absolute_max_dose の小さい方）を超える場合、フロントエンドは理由をインライン表示して追加・更新を送信せず、バックエンドも Create/Update の永続化前に 400 で拒否する。`ConfirmDialog` による解除経路は設けない。上限内でも下限未満または推奨値からの著しい乖離は、インラインの逸脱理由（`dose_deviation_reason`、空白のみ不可・500文字以内）が必須。理由を snapshot と同一 transaction の監査へ保存し、actor・監査依存の欠落や監査失敗時は保存しない（`computeDoseGate` / `ensureDoseDeviationAuditReady` / `auditDoseDeviationTx`）。現行実装は体重未記録・species 正規化不能・投与量パラメータ未設定時に評価をスキップして保存を継続する参照で、最新合意の医師確認・理由・編集履歴を満たす証拠ではない。最新契約は評価不能を明示し医師確認・理由付き手入力を認める。パラメータ取得の非 NotFound エラーと species 不一致は既存どおり fail-closed とする。権限付き例外フロー（Design B）は実装しない。
- **臨床承認の境界**: コード上のゲートと、対象薬・値・範囲・単位・出典を含む臨床承認は別。#201 の Closed は #261 への承認作業の集約であり、医師による臨床 bundle の承認完了を意味しない（[承認作業 #261](https://github.com/MinoruSoga/AnimalEkarte/issues/261)）。
- **未保存警告**: 変更がある状態でページを離れようとすると `NavigationBlocker` が警告を表示。
- **特記マーク（EMR-173/231）**: スティッキーヘッダー（`PatientContextHeader`）に共有 `DangerBadge` で表示: ペット名横に `danger_level` 高=赤い八角形アイコン・中=黄い三角形アイコン（文言なし、Popover で補足メモ）、低・未設定は非表示。飼主 `is_dangerous` は飼主名横に文言なしの赤い八角形アイコン（代替名「特記」）。医療記録中もマークを常時視認できるようにするための安全表示。


<a id="medical-record-agreement-input-four"></a>
## 入力まわり4項目の合意（2026-10-14）

質問 Sentinel_374bdacb47a0819180046f40e27a41dc と回答 Sentinel_6287d7cbcf988191aac47fbe504cfad1「ok／4 基本的にはチェックボックスとモーダルで実施済に切り替える」が根拠。1〜3の提案を承認、4は具体的操作指定を優先する。これは部分合意で、[初回5項目](../lifecycle.md#medical-record-agreement-first-five)・保存§2・予約／受付接続・編集ログ・医師確定・評価不能時手入力を保持し、実装／API／実機受入済みとしない。

<a id="requirement-fr-medical-records-selective-history-copy"></a>
### 1. 選択した問診・治療明細の複写

過去カルテから問診と治療明細を選んで複写できる。問診を上書きするときは確認し、取消なら現在入力を保持する。複写元は変更せず、問診は既存どおりローカル入力へ適用し問診保存で永続化する。治療明細は選択した項目だけを既存の個別保存へ渡し、今回の有効マスタの価格を用いる。元カルテの当時価格は元記録に保持する。

無効なマスタや対応しない項目は知らせ、再選択してもらう。旧単価・無型／不一致項目を通知なしで新しい明細へそのまま保存するfallbackは採用しない。対象と現在の医院・患者・grant・確定状態、薬量評価／理由・確認を保存時にも検査する。

現行の治療複写は逐次POSTで、途中失敗時も成功済み行が残る。今回、全件同時成功・全体rollbackへ変更していない。成功済み／失敗／未送信の対象を区別して通知し、再試行で成功済みを重複作成しないよう照合する。結果不明を失敗確定にしない。UIの選択範囲と表示方式・再送方式は設計、全体atomic化は未承認のまま。下の旧履歴複写説明は現行実装参照で、無効／不一致fallback等は本合意へ未追従。

<a id="requirement-fr-medical-records-manual-template-insertion"></a>
### 2. 問診テンプレートの挿入・追記／置換

カルテで登録済みテンプレートを手動選択し、問診欄へ挿入できる。基本は追記、置き換える場合だけ確認する。置換確認の取消は既存入力を保持する。適用した内容は問診の未保存入力として扱い、テンプレート選択だけでサーバー保存しない。問診の既存保存・離脱警告・編集権限・確定ロックを継承する。

自動挿入・動的サジェスト、別の診断／処方欄への自動適用は追加しない。テンプレートマスタCRUDとカルテ内の適用を別操作にし、挿入で元テンプレートを書き換えない。具体的問診欄・カテゴリ対応と連続適用の表示はUI／型設計で詰め、今回未指定の必須入力や固定カテゴリを追加しない。

<a id="requirement-fr-medical-records-master-history-snapshots"></a>
### 3. 診療マスタの権限編集・無効化・当時内容保持

権限のあるスタッフが診療マスタを編集できる。医院scope・対象masterの操作grantを保持し、カルテ編集権限だけでmaster編集を全開放しない。薬量の対象・値・範囲・単位・出典の承認は医師という初回合意を維持し、新たな臨床責任者役割を作らない。

無効化した項目は新規選択できない。既存カルテは対象項目が持つ名称・単位・価格を記録当時の内容で保持し、名称変更／価格改定・無効化で過去記録を置換しない。再診での新しい複写明細は今回価格、複写元／既存明細は当時価格と区別する。全masterへ存在しない単位／価格を新設する意味ではない。

現在master参照しかない既存データの当時内容を最新値から推測補完しない。確認できる資料・保存snapshotの調査／移行は未実施で、取得不能を当時値確定と表示しない。snapshot方式・master版と選択中の無効化／改定競合は技術設計。訂正として内容を変更するときは初回の編集者ログ・医師確認／確定後追記・会計整合へ従う。

<a id="requirement-fr-medical-records-plan-execution-states"></a>
### 4. 治療プランの項目状態と実施済み操作

通常カルテの治療プラン項目ごとに「未実施・実施済み・後日予定・中止」を管理する。確定時は未実施・後日予定を未完了として警告し、確認した医師は未完了でも確定可能。中止は理由を残し、実施済みへの変更では誰がいつ記録したかを残す。中止と実施済みを同義にせず、確定・診療終了・精算で全項目を実施済みにしない。

実施済みへの切替はチェックボックスとモーダルを基本操作とし、項目と実施内容を識別できるよう既存の診察／治療プラン画面へ接続する。両方を必ず順番に経由する二段階業務や、二度の状態writeを必須にしない。チェック解除だけで無条件に戻さず、[追加合意の理由付き取消モーダル](../lifecycle.md#requirement-fr-medical-records-reasoned-plan-execution-undo)へ接続する。実施済みの誤記訂正・確定済みへの追加変更は初回の専用訂正／追記保護へ渡し、逆向きtoggleで解除しない。

記録者は操作した認証済みスタッフ、記録時刻は状態を記録した時刻。実際の実施者・実施時刻と同義にせず、操作しただけで実施者／実績時刻を自動確定しない。今回指定のない実施者選択・実施時刻入力を必須にしない。実施を記録することと、治療明細・課金生成／取消は別の境界で、チェックだけで新しい請求ルールを作らない。入院・ホテルの専用記録／登録時snapshotには本項目状態を自動適用しない。

<a id="履歴複写"></a>
### 履歴複写（従来実装参照）

以下の旧コピー動作には最新合意への未追従を含む。選択複写・無効／不一致項目の通知再選択・今回価格と元記録保持は上記合意を優先し、従来のfallbackを新業務許可にしない。

- **問診履歴の前回複写（EMR-182 / EMR-219）**: `InterviewHistory` の各行は、複写可能な問診値または `recordId`（明細複写の起点）を持つ場合に限り行末尾に「コピー」ボタンを表示する（`InterviewHistoryItem` の `copySource` が存在する行のみ。複写不可の行には表示しない）。複写ペイロードは `useGetPetMedicalHistory` 内の `transformToHistoryItem` が `InquirySummaryResponse` の `inquiry.chief_complaint`（主訴詳細）・`inquiry.notes`（治療方針）・`inquiry.chief_complaint_type_id`（主訴区分）に加え、元カルテ `id` を `recordId` として保持する（feature 側の `transforms.ts` も同形）。問診値を持たなくても `recordId` のみの行は治療明細の複写が可能。「コピー」押下時、`MedicalRecordInterview` は現在の主訴詳細・治療方針・主訴区分が未編集の既定値なら即時適用し、いずれかがユーザー編集済みなら `ConfirmDialog` で上書き確認を求める（キャンセルは現在値を保持）。問診項目の複写はローカル state の setter のみを更新して `NavigationBlocker` の未保存警告を有効化し、API 送信は行わない。ボタンは行リンク（`/medical-records/:id` 詳細遷移）の sibling の `<button>` であり、確定済み・送信不可のとき `isLocked` prop で明示的に disabled にする。`InterviewHistory` パネル自体は編集ロック用の disabled fieldset の外に配置するため、確定済みでも履歴検索・行リンクは有効のまま（`ExaminationForm` が読み取り専用パネルを fieldset 外の兄弟に置くのと同じ境界規則で、`MedicalRecordMountedTab` の `internalLock` 指定タブが内部でこの境界を管理する）。
  - **治療明細の複写（EMR-219）**: 複写適用の単一経路（`applyCopySource`）が `onCopyRecordTreatments(recordId)` を発火し、`useCopyTreatmentDetails` が複写元カルテの明細を `GET /v1/medical-records/{sourceId}/treatments` で適用時に lazy fetch する（履歴一覧レスポンスに明細は含まれないため）。各明細行は `buildCopiedTreatmentPayload` で create ペイロードへ変換し、consultation/procedure/medicine 行の `unit_price` は複写元の当時価格を持ち込まず、現行マスタ（`/v1/masters/consultations`・`/v1/masters/procedures`・`/v1/masters/medicines`、キャッシュは共有 master 参照キー）の `id`・`price` で再解決する。マスタ不一致または無型（`item_type=other`）行は複写元の `unit_price`・参照をそのまま維持する。作成は既存の `POST /v1/medical-records/{currentId}/treatments` へソース `sort_order` 昇順で逐次 POST し、`sort_order` は複写先の現在最大値 +1 からのオフセットを加算する。現在カルテに `id` が無い（自動作成前）場合は問診項目の複写のみ適用され明細複写はスキップされる。fetch/create 失敗時はエラートーストを出し、複写元カルテへの書込みは一切行わない（作成済み行はロールバック API が無いため残る）。

<a id="persistence"></a>
## 保存・取消・再試行・同時編集

<a id="requirement-fr-medical-records-optional-chief-complaint-type"></a>
#### 2.2 保存プロセス

> **追記（2026-10-07 監査）**: 保存ルーティング（アクティブタブ単位のメイン保存・ヘッダー属性の即時保存・タブ別サブデータの即時保存と各エンドポイント）の契約本文の正本は [保存ルーティング契約](../lifecycle.md#hybrid-sync)（ハイブリッド保存ロジック）。本節は 9 タブ構成（§1）のうち本画面固有の操作仕様のみを残し、ルーティングの重複掲載（旧・送信先表と mermaid 図）をやめた。

- **メイン保存の UI**: 画面右下のフローティング「保存」ボタン（`MedicalRecordFloatingActions`）。送信先はアクティブタブで切り替わる（「問診」は `PATCH /medical-records/:id/inquiries`、「診察/治療プラン」は `PATCH /medical-records/:id/clinical-plan`。切替契約の本文は [保存ルーティング契約](../lifecycle.md#hybrid-sync)）。他タブではカルテ本体の送信は行われない（`PATCH /medical-records/:id` へまとめて送信する方式ではない）。予防接種タブでは偽成功を避けるため外側の保存ボタン自体を表示せず、タブ内の接種記録追加を使う。
- **主訴区分は任意（空欄可）**: 「問診」タブの主訴区分は必須項目ではない。`InterviewChiefComplaint` 内のクリア可能な `SearchableSelect`（`clearable`）で選択し、候補は `useGetChiefComplaintTypes` の主訴区分マスタから供給される。クリア操作で空欄に戻せ、空欄時は `PATCH /medical-records/:id/inquiries` へ `chief_complaint_type_id: null` が送られる。バックエンドは nullable（`*uint64`）として受け付け、明示的なクリアとして永続化するため、保存・再読込後も空欄のまま維持される。
- **ヘッダー即時保存の UI 側挙動**: 担当医・来院種別・診療日はヘッダー変更と同時に `PATCH /medical-records/:id` する（保存ボタンを経由しない）。来院種別の成功後は `queryKeys.medicalRecords.detail` を invalidate し、再読込でラベルが戻らないようにする。失敗時はローカル state をロールバックする。appointment 紐付き通常カルテの `date` 固定（予約開始の JST 日付・変更は BE Conflict・UI は未紐付け時のみ成功）の契約本文は [保存ルーティング契約](../lifecycle.md#hybrid-sync) / [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §5。EMR-308: ヘッダーの「次回予定」（`NextVisitButton`）はカルテ PATCH ではなく `PUT /v1/reservations/pets/:petId/next-visit-intent`（取消は `DELETE` 同パス）でペットの tentative appointment（`status='tentative'`、日付のみ `intended_date`、枠・時間・担当未定）を upsert/削除する。write 契約の正本は [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §2.3、画面側の管理は [reservation-management.md](../../reservations/pages/reservation-management.md) を参照。
- **アクティブタブの追加保存**: 保存成功直後、その時点で開いているタブが「診察/治療プラン」または「見積書」の場合のみ、`useMedicalRecordPostSave` が対応する登録済みコールバックを追加実行する（他タブ在中時は発火しない。両タブを並行実行することもない）。見積タブは `items` を create/update 同一 tx で置換永続化する（独立画面 `/estimates` はヘッダ金額のみ。詳細は [23-estimate-form.md](../../../screens/23-estimate-form.md)）。
- **治療・検査等のサブリソース**: 「治療」タブの明細は行単位の追加/編集/削除操作ごとに個別・即時送信される。メイン保存とは独立しており、「バックグラウンド並行保存」ではない（タブ別サブデータの即時保存契約は [保存ルーティング契約](../lifecycle.md#hybrid-sync)）。

<a id="errors-empty"></a>
## エラー・空状態

| 状況 | 検知／表示 | 利用者の次の操作 | 根拠 |
|:---|:---|:---|:---|
| 会計確認が未完了 | 確定ボタンをブロック | 会計確認状態の確認、復旧の詳細は要確認 | 臨床安全ガード |
| 画像件数・容量超過 | 選択分全体を拒否してエラー | 制限内のファイルを選択 | 画像・資料の選択制限 |
| ヘッダー更新失敗 | ローカルstateをロールバック | 未記載／要確認 | 画面保存操作 |
| 予防接種記録0件 | EmptyState | 記録を追加 | 9タブ表 |
| その他エラー・空状態 | 未記載／要確認 | 未記載／要確認 | タブ別条件の不足は要確認 |

<a id="acceptance"></a>
## 受入条件

この表は既存仕様から抽出した代表条件。すべての入力・操作・異常系への要件IDと受入条件の網羅的対応は未記載／要確認。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-optional-chief-complaint-type | 入力可能な記録で主訴区分を空欄にする | 保存して再読込する | 主訴区分は空欄のまま | この問診保存だけを確認 | 任意項目の空欄が保持される | [要件](#requirement-fr-medical-records-optional-chief-complaint-type)・[既存UATの所在](../README.md#受入条件の所在)。実機未確認 |
| AC-medical-records-copy-select | 問診／治療明細がある過去カルテ | 対象を選び複写 | 選択対象だけ今回入力に適用し元記録は不変 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-overwrite-cancel | 問診に既存入力あり | 上書き確認を取り消す | 現在入力を保持し上書きしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-current-price | 過去明細と今回有効masterの価格が異なる | 治療明細を複写 | 新明細は今回価格、元の当時価格は保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-invalid-master | 複写対象が無効／対応masterなし | 複写を適用 | 知らせて再選択し、旧価格fallbackで新規保存しない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-copy-partial-failure | 逐次保存で一部成功・後続失敗／不明 | 結果を確認して再試行 | 成功済みを保持・照合し重複作成しない、全件rollbackへ変えない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-selective-history-copy)。実機未確認 |
| AC-medical-records-template-append | 問診に入力あり | templateを手動選択して追記 | 既存入力を保ち未保存問診に追記、即時API保存なし | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-template-replace-cancel | 問診内容をtemplateで置換しようとする | 確認を取り消す | 既存問診を保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-template-no-auto | template候補がある | 画面を開く／患者を切り替える | 勝手に挿入／自動提案／元template更新をしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-manual-template-insertion)。実機未確認 |
| AC-medical-records-master-inactive-history | 記録済み項目を無効化／名称単位価格変更 | 新規選択と過去閲覧 | 無効化した項目は新規利用不可。有効のまま改定した項目は今回値で新規選択可。どちらも既存当時内容は保持 | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-master-history-snapshots)。実機未確認 |
| AC-medical-records-master-scope-grant | master操作権限なし／別医院 | masterを編集 | カルテ編集可だけでmasterを編集させない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-master-history-snapshots)。実機未確認 |
| AC-medical-records-plan-checkbox-modal | 権限あるスタッフが未完了項目を確認 | チェックボックスとモーダルの導線で実施済みにする | 対象項目を実施済みにし実記録者・記録時刻を残す、二度writeしない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-uncheck-not-undo | 実施済み項目がある | チェックを解除しようとする | 無条件の実施取消・確定ロック解除へ転用しない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-cancel-reason | 中止対象の項目がある | 中止にする | 理由を残し実施済みとは扱わない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-deferred-warning | 未実施／後日予定項目がある | 医師が確定警告を確認 | 警告して確定可、項目を実施済みに自動更新しない | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |
| AC-medical-records-plan-recorder-not-performer | 記録者と実施者が異なり得る | 実施済みを記録 | 記録者／記録時刻を残し実施者へ自動転記せず未指定項目の必須化なし | 本ケースを単独確認 | 合意結果と既存保護を維持 | [FR](#requirement-fr-medical-records-plan-execution-states)。実機未確認 |

<a id="decisions-evidence"></a>
## 未決事項・関連仕様・根拠

未記載の型・必須条件・初期値・エラー・受入結果は要確認。既存の未実装・判断待ち・臨床承認待ちは以下の元記載を維持する。今回の形式整理を合意や実装完了の根拠にしない。

[文書規約](../../../authoring-guide.md)・[ドメイン入口](../README.md)。根拠は整理前の既存仕様本文と、その本文が示す日付付き追記・関連資料。文書確認日は上記の構造確認日であり、過去の承認日を置き換えない。API表・コード名は参照情報で、wire型とDB定義の正本は文書規約に従う。

### 3. 技術仕様



#### 使用コンポーネント

- **`MedicalRecordForm`**: 統合フォーム。
- **`useMedicalRecordForm`**: 複雑な状態管理（9タブ ＋ 履歴引用）を統括するカスタムフック。
- **`historyItems`**: 「問診」タブ内の右カラム（`InterviewHistory`）に過去の問診履歴を表示。「予防接種」タブは左ペイン一覧と右カラム `VaccinationHistory` の両方を `useGetPetVaccinations`（`GET /vaccinations?pet_id=`）で描画する（フォーム全体で共有される単一サイドパネルではなく、タブごとに独立）。

#### API連携

| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/medical-records/:id` | カルテ本体の取得（飼主／患者／担当医・問診要約・体重投影等。治療・全バイタル・clinical-plan・画像・検査等は個別取得） | `medical-records` | `view` |
| POST | `/api/v1/medical-records` | 新規カルテレコードの作成 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id` | カルテ属性の更新 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/inquiries` | 「問診」タブの保存 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/clinical-plan` | 「診察/治療プラン」タブ（身体検査所見・診断・治療方針）の保存 | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/treatments` | 処置・処方明細の追加 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の更新 | `medical-records` | `edit` |
| DELETE | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の削除 | `medical-records` | `delete` |
| PUT | `/api/v1/medical-records/:id/treatments` | 明細の並び替え（`sort_order` の一括更新。内容の一括更新ではない） | `medical-records` | `edit` |
| GET | `/api/v1/estimates?medical_record_id=` | 「見積書」タブの見積データ取得 | `estimates` | `view` |
| POST | `/api/v1/estimates` | 見積の新規作成（タブからは `items` を同一 tx で保存） | `estimates` | `create` |
| PATCH | `/api/v1/estimates/:id` | 見積の更新（タブからは `items` を同一 tx で置換） | `estimates` | `edit` |
| GET | `/api/v1/vaccinations` | 予防接種タブのペット単位一覧（必須クエリ `pet_id`。`page`/`limit` 必須） | `vaccinations` | `view` |
| POST | `/api/v1/vaccinations` | 予防接種タブからの新規接種記録 | `vaccinations` | `create` |


<a id="medical-record-agreement-integration-three"></a>
## 画像・印刷・後日結果・訂正影響の追加3項目

2026-10-14、質問 Sentinel_f99722e55be0819189e5d50080928914 に対する Sentinel_59d45d62afe08191bc0b289c42486962 の「ok」に基づく3FR／12AC。既存の入力・確定・訂正と保存§2を継承し、実装／実機受入済みを表さない。

<a id="requirement-fr-medical-records-image-drop-selected-print"></a>
### 画像の共通入力検証と対象範囲を選ぶ印刷

画像はファイル選択とドラッグ＆ドロップの両方で登録でき、同じ形式・件数・容量・操作権限・医院／患者照合・確定保護を適用する。JPEG／PNG／GIF／PDF、一度10件まで、1ファイル10MiB以下・合計50MiB以下という既存制限を保持する。選択分に超過があれば全体を受け付けず、D&Dだけで制限を迂回しない。形式の選択補助とファイル内容の検証は別で、保存時の検証も維持する。

印刷前に対象のカルテと範囲を選びプレビューする。範囲は既存の患者ヘッダー・各タブ・バイタル・画像／資料から選択する設計とし、選択した内容に関係する訂正・追記は元記載との関係と日付／記録時刻、記録者が分かる表示を含める。訂正済みを元のままの記載として印刷しない。範囲選択は未保存draftの自動保存・全タブ一括保存を発火せず、保存済みの内容を対象にする。未保存内容は印刷に反映されないことを表示し、印刷cancelは入力を保持する。

既存の閲覧／操作認可と対象情報の閲覧範囲を検査し、表示できない他ownerのデータを印刷で取得・公開しない。新しいgrant IDや飼主への自動公開・送付は追加しない。画面・帳票レイアウトや既存の印刷入口へのadapterは技術設計で、印刷成功が診療確認・確定を代行しない。

<a id="requirement-fr-medical-records-late-result-shared-history"></a>
### 後日検査結果の共通参照と履歴

作成中カルテには後日結果を追加できる。確定済みカルテは元の記載を保持して結果を追記する。結果訂正は変更前後・理由・操作主体／記録時刻を履歴として保持し、検査画面とカルテで同じ結果とその訂正履歴を参照する。別々の結果本文を二重管理せず、患者・医院・検査・カルテの対応を照合する。

確定時に参照した結果版と後日の結果／訂正版を区別し、現在の結果だけで確定時の記載を上書きしない。カルテの確定解除、通常子入力PATCHによる確定記録の変更、別患者結果の取込を認めない。検査ownerの確定／訂正・権限手順は検査正本へ渡し、検査側の確定解除をカルテの確定解除の根拠にしない。作成中で通常入力が必要な結果待ちを自動確定しない。結果の未取得・取得失敗・対応不明は「結果なし」「正常」へ置換しない。

<a id="requirement-fr-medical-records-correction-owner-impact"></a>
### 訂正による会計・在庫への影響表示と受渡し

カルテ訂正だけで確定済み請求・返金・在庫を自動変更しない。変更した記録／項目と理由、関連する請求・在庫等への影響を表示し、各機能の訂正操作へ案内する。対応が不明な影響を「影響なし」「訂正済み」と表示しない。手元で差額・数量を推測して確定処理せず、権限あるownerの確認・操作へ渡す。

精算確定歴がある予約の通常診療終了Undo禁止、確定済み原記録・元日付の保持、当日カード／自動課金なしの過去入力、復元だけで開始／課金しない境界を維持する。締め後会計・返金・在庫増減の内部手順はD10／12／13で詰める。カルテの案内表示や訂正履歴を、返金成功・在庫増減成功の証拠にしない。閲覧権限がないownerの内容は公開せず、操作権限不足は案内先の権限不足として区別する。

### 追加受入条件（実機未確認）

| ID | Given | When | Then | 検証 | 状態 | 正本 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-image-drop-same-validation | 同じファイル集合 | 選択／D&Dで追加 | 形式・件数・容量・権限・scope・確定保護が同じ | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-image-drop-reject-batch | 個別または合計サイズ／件数超過 | どちらの入口でも追加 | その選択分全体を拒否し保存済み添付を保持 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-print-range-corrections | 訂正／追記のある保存済み記録 | 範囲を選びプレビュー／印刷 | 選択対象と関係する訂正追記・元記載の関係を識別可能に表示 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-print-draft-cancel-scope | 未保存入力または閲覧不可のowner情報あり | 印刷準備／cancel | draft未反映を表示、無断保存・権限外印刷・入力破棄・飼主自動公開なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-image-drop-selected-print) |
| AC-medical-records-late-result-draft | 作成中で患者／医院を照合済み | 後日結果を追加 | 検査ownerの同じ結果を参照し自動確定しない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-finalized | 確定済みの結果参照あり | 後日結果／訂正版到着 | 元記載・参照版を保持し追記、unfinalizeや原本文上書きなし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-revision-history | 結果が訂正される | 検査画面とカルテを開く | 同じ結果IDと履歴を参照し前後・理由・主体／時刻を識別 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-late-result-unknown-scope | 別患者／医院または必要取得失敗 | 結果取込／再表示 | 取込を止め照合、未取得を正常／結果なしにしない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-late-result-shared-history) |
| AC-medical-records-impact-financial-stock-no-auto | 確定請求または在庫関連あり | カルテ訂正 | 影響とowner訂正入口を表示し自動請求変更・返金・在庫増減なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-unknown | 関連影響の取得失敗／対応未確認 | 訂正後に確認 | 未確認として表示し影響なし／訂正済みへ置換しない | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-handoff-permission | owner操作権限不足 | 訂正入口へ進もうとする | 権限不足を表示し迂回操作・他owner情報公開なし | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |
| AC-medical-records-impact-settled-history | 精算確定歴／締め後または過去入力／復元 | カルテ側の操作 | 既決のUndo禁止・自動課金なしを保持し会計／在庫内部はowner確認 | 個別条件を確認 | 実機未確認 | [FR](#requirement-fr-medical-records-correction-owner-impact) |

<a id="medical-record-shared-coverage-ac"></a>
## 全タブ・共通操作の横断受入条件

これまでの合意と保存§2に対する14ACを補う。新しい業務許可や実機PASSではない。各条件は問診・診察／治療プラン・治療・予防接種・健診・検査・画像・見積・会計医師確認、バイタル・ヘッダー・専用訂正の該当write／readに適用し、各ownerの正本を維持する。

| ID | Given | When | Then | 検証 | 状態 | 正本 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-medical-records-coverage-active-tab-only | 複数タブにdraftあり | 問診／診察でメイン保存 | アクティブ対象の既存送信先だけ保存し別タブを無断保存しない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-header-immediate | ヘッダー変更とタブdraftあり | 担当医／来院種別／日付を許可範囲で変更 | 既存個別保存と予約整合を保持、タブdraftを保存成功扱いにしない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-child-independent | 親save前のdraftと子入力あり | 許可された子操作 | 既存子保存先へ独立送信し親save待ち／全タブ一括保存へ変更しない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-vital-read-state | バイタル履歴取得失敗／不完全または患者切替中 | 表示／入力／薬量評価 | 測定なしへ置換せず対象一致と現在情報を再照合 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual) |
| AC-medical-records-coverage-snapshot-null | 旧master参照・任意null・未取得値あり | 読込／複写／保存 | 当時値を保持し未取得を0／空欄／現masterの当時値へ推測補完しない。APIごとの未送信／null／clear・旧状態を照合 | 全該当経路を照合 | 実機未確認 | [正本](#requirement-fr-medical-records-master-history-snapshots) |
| AC-medical-records-coverage-finalized-all-writers | finalized親の全子／外部owner経路 | 通常の追加更新削除を要求 | 原記載を変更せず許可された専用追記／訂正へ渡す。GET初期化・関連変更・機器取込を含む全writerを照合 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction) |
| AC-medical-records-coverage-finalize-concurrent-deps | 警告／医師確認後に子や依存が変更 | 確定する | 最新依存を再照合し旧確認で確定しない。未完了確認は確定可、取得不明は停止 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-explicit-finalize-plan-warning) |
| AC-medical-records-coverage-scope-generation | 旧医院／患者／actorの遅い取得応答あり | 画面／医院／患者を切替 | 新対象へ旧値を適用せずscopeを再取得・照合 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-child-audit-failure | 本体／子／専用操作の必要ログ保存が失敗 | 更新を保存 | 必要ログなしの成功を返さず同じ操作単位をrollback | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-save-partial-dirty | 問診成功・他draft／子操作失敗あり | 保存結果を表示 | 成功対象だけdirtyを解除し未保存入力／失敗対象を保持 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#hybrid-sync) |
| AC-medical-records-coverage-unknown-receipt | 送信後通信断で結果不明 | 再試行／取消を要求 | 同じ操作を照会／再送し重複転記・複写・復元を作らない。未対応receiptの旧APIは成功と推測せず関連照合し不明ならwrite停止 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-empty-record-hide-restore) |
| AC-medical-records-coverage-versions-no-overwrite | 他者が同対象を変更済み | 古い版の更新 | 上書きせず再取得し自分の入力と比較・再確認 | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| AC-medical-records-coverage-special-record-separation | 入院／ホテル／トリミング記録が併存 | 通常記録の状態／複写／受付を操作 | 専用記録へ新状態や作成契約を自動適用しない | 全該当経路を照合 | 実機未確認 | [正本](../lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| AC-medical-records-coverage-owner-read-and-handoff | 別ownerへの影響・印刷対象あり | 閲覧／引継ぎ後戻る | 認可範囲と現在版を検査、遷移をowner訂正成功とみなさない。転記／過去／訂正で通常createの在庫副作用を誤流用しない | 全該当経路を照合 | 実機未確認 | [正本](#requirement-fr-medical-records-correction-owner-impact) |

## 現APIとの最終照合の入口

[実キー・型・null／clearと互換契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)へ接続する。実APIの既存受理範囲を新業務許可へ拡大しない。患者転記・過去入力・復元・確定後結果追記を、通常PATCH／createの代用で提供しない。取得経路にはGetOrCreate初期化があるため、印刷用read modelは副作用のない保存済みprojectionで構成する。共通14ACは全該当writer／read経路への確認条件で、現コードの適合PASSを示さない。
