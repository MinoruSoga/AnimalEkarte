# カルテ詳細・入力 仕様書 (Medical Record Form)

## 概要
- **画面の目的**: 臨床現場における電子診療録（SOAPS形式）の作成、検査、処置、処方、および会計連携の統合管理。
- **URLパターン**: 
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`
- **アクセス権限**: 新規作成 (`/medical-records/new`) は `medical-records:create` を要求する `RequirePermission` ルートガードあり。編集 (`/medical-records/:id`) は親ルート `/medical-records` の `medical-records:view` ガードのみを継承し、保存可否はコンポーネント内の `usePermission`（`ResourceMedicalRecords`）で制御。

---

## 1. 画面構成 (9 タブ構成)

膨大な診療情報を整理するため、以下の 9 つのタブで構成されています。

| タブ | 名称 | 内容 |
|:---:|:---|:---|
| 1 | **問診** | 飼主からの主訴、現在の症状、生活環境のヒアリング記録。 |
| 2 | **診察/治療プラン** | 身体検査所見（SOAPのS/O/A）、診断名（第1・第2）、今後の治療方針（Plan）。 |
| 3 | **治療** | 実施した処置、処置項目、および処方薬の明細。 |
| 4 | **予防接種** | ワクチン接種履歴の記録、次回予定日の設定（手入力。「次回予防接種予定設定」ラジオは日付自動算出に未配線）。 |
| 5 | **定期健診** | 健康診断の結果記録。 |
| 6 | **検査** | 検査結果の一覧表示と既存検査の取込（カルテへの紐付け）、基準値との比較（HIGH/LOW判定ハイライト）。 |
| 7 | **画像** | 患部写真、レントゲン、エコー画像、PDF資料のアップロード・管理。 |
| 8 | **見積書** | 提示した概算費用の管理。 |
| 9 | **会計(医師確認)** | 診療費の最終確認と会計ステータスの送信。 |

---

## 2. 主要な臨床ロジック

### 2.1 SOAPS 形式の採用
- **Subjective (主観的)**: 飼主の話。
- **Objective (客観的)**: 視診・触診・聴診結果。
- **Assessment (評価)**: 診断、病名特定。
- **Plan (計画)**: 今後の検査・治療予定。
- **Supplement (補足)**: 注意事項等。

### 2.2 保存プロセス
- **メイン保存**: 画面右下のフローティング「保存」ボタン（`MedicalRecordFloatingActions`）は、アクティブタブに応じて送信先を切り替える。「問診」タブは `PATCH /medical-records/:id/inquiries`（主訴・主訴区分・治療方針）、「診察/治療プラン」タブは `PATCH /medical-records/:id/clinical-plan`（治療方針・診断詳細・診断名）と `PATCH /medical-records/:id`（次回来院推奨日）を送信する。他タブではカルテ本体の送信は行われない（`PATCH /medical-records/:id` へまとめて送信する方式ではない）。
- **アクティブタブの追加保存**: 保存成功直後、その時点で開いているタブが「診察/治療プラン」または「見積書」の場合のみ、`useMedicalRecordPostSave` が対応する登録済みコールバックを追加実行する（他タブ在中時は発火しない。両タブを並行実行することもない）。
- **治療・検査等のサブリソース**: 「治療」タブの明細は行単位の追加/編集/削除操作ごとに `/medical-records/:id/treatments...` へ個別・即時送信される（メイン保存とは独立しており、「バックグラウンド並行保存」ではない）。

### 2.3 臨床安全ガード
- **確定ロック**: `finalized`（確定済）ステータスのカルテはバックエンドが更新を拒否し（409）、訂正は追記（addendum）のみ許可することで真正性を担保。確定への遷移は `PATCH /medical-records/:id` の `status` 指定によるもの。画面右下のフローティングアクション（`MedicalRecordFloatingActions`）に「確定する」ボタンが表示され（編集権限あり・保存済み・未確定の場合のみ）、`MedicalRecordFinalizeDialog` で不可逆であることを確認した上で確定する。確定取り消し（unfinalize）API は存在しないため、確定後の修正経路は訂正追記（addendum）のみ。会計完了時の自動確定は現状存在しない。確定済みカルテはサイドヘッダーに「確定済」バッジ（`StatusBadge`）を常時表示する。
- **薬量自動計算と上限ゲート（#201）**: 「治療」タブの処方明細（`TreatmentRow`）は、対象ペットの species と当日 vital 体重から数量を自動プリフィルする（`calculateDose`。体重未記録・パラメータ未設定時はプリフィルせず手動入力）。保存時は `computeDoseGate` が上限（体重連動上限 weight×max_mg/kg と体重非依存の絶対上限 absolute_max_dose の小さい方）との乖離を判定し、著しい逸脱・上限超過時は `ConfirmDialog`（「この数量で保存する」）を提示する。**既知の是正対象**: 現行はこのダイアログで超過を通過でき、バックエンドも拒否せず audit 記録のみで保存する。確認ダイアログを安全統制に使わない原則（product-philosophy ③）および fail-closed 不成立経路の是正が [#201 [SAFETY]](https://github.com/MinoruSoga/AnimalEkarte/issues/201) で再オープン済み（物理ブロックまたは権限付き例外フローへ変更予定・Lock UI 等の判定は #261）。
- **未保存警告**: 変更がある状態でページを離れようとすると `NavigationBlocker` が警告を表示。

---

## 3. 技術仕様

### 使用コンポーネント
- **`MedicalRecordForm`**: 統合フォーム。
- **`useMedicalRecordForm`**: 複雑な状態管理（9タブ ＋ 履歴引用）を統括するカスタムフック。
- **`historyItems`**: 「問診」タブ内の右カラム（`InterviewHistory`）に過去の問診履歴を表示。「予防接種」タブには別途 `VaccinationHistory` が同様の履歴表示を持つ（フォーム全体で共有される単一サイドパネルではなく、タブごとに独立）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/medical-records/:id` | カルテ本体の取得（Treatments・Vitals・Owner・Pet 等を同梱。clinical-plan・画像・検査等は個別エンドポイント） | `medical-records` | `view` |
| POST | `/api/v1/medical-records` | 新規カルテレコードの作成 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id` | カルテ属性の更新 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/inquiries` | 「問診」タブの保存 | `medical-records` | `edit` |
| PATCH | `/api/v1/medical-records/:id/clinical-plan` | 「診察/治療プラン」タブ（身体検査所見・診断・治療方針）の保存 | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/treatments` | 処置・処方明細の追加 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の更新 | `medical-records` | `edit` |
| DELETE | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の削除 | `medical-records` | `delete` |
| PUT | `/api/v1/medical-records/:id/treatments` | 明細の並び替え（`sort_order` の一括更新。内容の一括更新ではない） | `medical-records` | `edit` |
| GET | `/api/v1/estimates?medical_record_id=` | 「見積書」タブの見積データ取得 | `estimates` | `view` |
| POST | `/api/v1/estimates` | 見積の新規作成 | `estimates` | `create` |
| PATCH | `/api/v1/estimates/:id` | 見積の更新 | `estimates` | `edit` |

---
