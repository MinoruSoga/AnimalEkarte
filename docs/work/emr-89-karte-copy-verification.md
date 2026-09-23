# EMR-89: スプシNo.9 カルテ複製時の金額注意喚起 — 検証記録

状態: **検証済み・実装対象なし**。スプシ No.9 の回答（「前回カルテをコピーして新規カルテを作成する機能は未実装。明細はマスタ選択時点の現在価格で入力されるため過不足は発生しない」）が現行コードと一致することを確認した。

## チケットの前提

「前回来院した時のカルテをコピーする。入力したものがコピーされると、金額に変更があった場合に会計に過不足がでる。複製する際にポップアップなどで金額に関する注意がでるか」（スプシ No.9）。

## 検証結果

### 1. カルテ複製機能は存在しない（警告の付け先なし）

- カルテ一覧の行アクションは 編集/削除 のみ。複製アクションなし（`frontend/src/features/medical-records/routes/MedicalRecordsListPanels.tsx`）。
- `backend/docs/api.yaml` に複製/コピー系エンドポイントなし。`backend/internal/medicalrecord/` に複製サービスなし。
- `use-apply-medical-record` / `use-apply-clinical-plan` は既存カルテのフォーム hydrate と治療計画同期であり、新規複製ではない。
- 新規カルテ自動作成（`use-medical-record-auto-create.ts`）は `pet_id` / `owner_id` / `visit_date` / `visit_type` / `appointment_id` 等のみ送信し、前回カルテの内容は引き継がない。
- 「過去のカルテ」パネル（`InterviewHistory.tsx`）は参照リンクのみで複製操作なし。

### 2. 唯一の「複製」はワクチン履歴のメタ複製（金額を含まない）

- `VaccinationHistory.tsx` L167 の「複製」ボタン → `use-medical-record-vaccination-form.ts` L72 `handleDuplicate`。複製されるのはワクチンID・ロット番号・次回予定日・備考のみ。金額・診療明細は複製しない。

### 3. 明細金額は常に「選択時点の現在マスタ価格」

- 処置明細: `lib/treatments-tab-model.ts` `buildMasterSelectionPayload` が `unit_price: item.unitPrice`（マスタ選択時点の価格）。
- 治療プラン: `MedicalRecordDiagnosisPlan.tsx` も同パターン。
- マスタから選ばない手入力経路でも旧価格の自動引継ぎは存在しない。

## 残リスク（製品 FAIL ではない・参考記録）

過去カルテ画面には保存時点の単価が表示される。担当者が過去カルテを見ながら新規カルテの単価を手入力で転記する運用は理論上残るが、これはコピー機能の欠陥ではなく人的転記であり、チケット回答どおりシステム経路での過不足は発生しない。PO が追加対策を望む場合のみ「過去カルテ表示中は金額が診察日時点」旨の表示追加が選択肢になり得るが、本チケットの要求範囲外。

## 結論

チケット本文の【未対応】回答は正確。複製機能が存在しないため「複製時の警告」の実装対象がなく、コード変更は不要。チケットは検証完了としてクローズ可能。
