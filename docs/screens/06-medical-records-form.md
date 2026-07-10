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
| 4 | **予防接種** | ワクチン接種履歴の記録、次回予定日の自動算出。 |
| 5 | **定期健診** | 健康診断の結果記録。 |
| 6 | **検査** | 数値検査結果の入力、基準値との比較（H/Lハイライト）。 |
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
- **メイン保存**: 「問診」「診察/治療プラン」等のカルテ本体属性（主訴、診断名等）は、ヘッダーの保存操作で `PATCH /medical-records/:id` にまとめて送信される。
- **アクティブタブの追加保存**: 保存成功直後、その時点で開いているタブが「診察/治療プラン」または「見積書」の場合のみ、`useMedicalRecordPostSave` が対応する登録済みコールバックを追加実行する（他タブ在中時は発火しない。両タブを並行実行することもない）。
- **治療・検査等のサブリソース**: 「治療」タブの明細は行単位の追加/編集/削除操作ごとに `/medical-records/:id/treatments...` へ個別・即時送信される（メイン保存とは独立しており、「バックグラウンド並行保存」ではない）。

### 2.3 臨床安全ガード
- **確定ロック**: `finalized`（確定済）ステータスのカルテはバックエンドが更新を拒否し（409）、訂正は追記（addendum）のみ許可することで真正性を担保。確定への遷移は `PATCH /medical-records/:id` の `status` 指定によるもので、会計完了時の自動確定や画面上の確定ボタンは現状存在しない。
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
| GET | `/api/v1/medical-records/:id` | カルテ本体と全サブリソースの取得 | `medical-records` | `view` |
| POST | `/api/v1/medical-records` | 新規カルテレコードの作成 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id` | カルテ属性の更新 | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/treatments` | 処置・処方明細の追加 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の更新 | `medical-records` | `edit` |
| DELETE | `/api/v1/medical-records/:id/treatments/:treatmentId` | 処置・処方明細の削除 | `medical-records` | `delete` |
| PUT | `/api/v1/medical-records/:id/treatments` | 明細の並び替え（`sort_order` の一括更新。内容の一括更新ではない） | `medical-records` | `edit` |
| GET | `/api/v1/estimates?medical_record_id=` | 「見積書」タブの見積データ取得 | `estimates` | `view` |
| POST | `/api/v1/estimates` | 見積の新規作成 | `estimates` | `create` |
| PATCH | `/api/v1/estimates/:id` | 見積の更新 | `estimates` | `edit` |

---
