# カルテ登録・編集フロー詳細仕様 (Medical Record Lifecycle)

## 概要
本システムにおける電子カルテ（Medical Record）は、臨床的な真正性と多種多様なサブリソースの整合性を維持するため、独自のライフサイクルと保存ロジック（ハイブリッド同期）を採用しています。

---

## 1. 登録と ID 確定のプロセス

### 1.1 未保存状態 (Draft in Memory)
新規作成画面 (`/medical-records/new?petId=xxx`) に遷移した直後は、データはクライアントサイドの React ステート上のみに存在します。

### 1.2 ID の発番と URL 昇格
最初の「保存」アクション（React 19 Action）がバックエンドで成功した時点で、サーバー側で `medical_records.id` が発番されます。
- **URL 置換**: 保存成功と同時に、フロントエンドは `replace` を実行し、ブラウザの URL を `/medical-records/:id` へ昇格させます。これにより、以後の「追記」や「再保存」が同一レコードに対して継続可能になります。

---

## 2. ハイブリッド保存ロジック (Hybrid Sync)

カルテは 9 つのタブ（問診 / 診察・治療プラン / 治療 / 予防接種 / 定期健診 / 検査 / 画像 / 見積書 / 会計(医師確認)）に分かれた膨大なデータを扱うため、「カルテ本体」と「タブ別サブデータ」を独立して保存します。

1.  **メイン保存**:
    - 主訴、カテゴリ、担当医、ステータス、来院種別、推奨再診日などの「カルテ本体」の属性を保存 (`PATCH /v1/medical-records/:id`)。
2.  **タブ別サブデータの即時保存**:
    - メイン保存の成否とは独立して、各タブ内の操作（追加・編集・削除）が発生した時点で即座に個別 API へ送信されます（メイン保存完了を待つゲート処理ではありません）。
    - **治療 (Tab 3)**: 項目の追加・更新・削除はそれぞれ `POST`/`PATCH`/`DELETE /v1/medical-records/:id/treatments(/:treatmentId)`。ドラッグ&ドロップでの並び替えのみ `PUT /v1/medical-records/:id/treatments` で一括更新。
    - **バイタル**: タブではなく `VitalsModal`（モーダルダイアログ）から `POST /v1/medical-records/:id/vitals` で記録。
    - **検査 (Tab 6)**: 新規検査は `/examinations/new` 画面で作成し、カルテの検査タブでは既存の検査記録を「取り込み」ダイアログ経由で `PATCH`（`medical_record_id` を設定）し紐付けます。

---

## 3. ステータス管理と保護

カルテには以下の 2 つの主要なフェーズが存在します。

### 3.1 作成中 (Draft)
- スタッフによる自由な追記・修正が可能です。
- 会計ステータスも「未精算」として、連動して更新されます。

### 3.2 確定済 (Finalized)
- 臨床記録としての真正性を担保するため、**編集操作が物理的にロック**されます。
- **ロックの波及**: 確定されたカルテに紐付く「検査結果」「処置明細」「画像」「健診記録」も、自動的に読み取り専用（Disabled）へと移行します。
- **修正**: 修正が必要な場合は、十分な権限を持つスタッフがステータスを一時的に「作成中」へ差し戻す必要があります。

---

## 4. 安全機能

- **離脱ブロック (`NavigationBlocker`)**: 未保存の変更（`isDirty=true`）がある状態でページを離れようとすると、破棄の確認を強制します。
- **二重送信防止**: React 19 の `useActionState` により、保存処理中の重複クリックを無効化します。

---

## API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| PATCH | `/api/v1/medical-records/:id` | カルテ本体の属性を保存・更新 | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/treatments` | 治療明細の個別追加 | `medical-records` | `create` |
| PATCH | `/api/v1/medical-records/:id/treatments/:treatmentId` | 治療明細の個別更新 | `medical-records` | `edit` |
| DELETE | `/api/v1/medical-records/:id/treatments/:treatmentId` | 治療明細の個別削除 | `medical-records` | `delete` |
| PUT | `/api/v1/medical-records/:id/treatments` | 治療明細の並び替え（一括） | `medical-records` | `edit` |
| POST | `/api/v1/medical-records/:id/vitals` | バイタル測定結果の記録（`VitalsModal`） | `medical-records` | `edit` |
| PATCH | `/api/v1/examinations/:id` | 既存検査記録をカルテへ紐付け（取り込み） | `examinations` | `edit` |

---
