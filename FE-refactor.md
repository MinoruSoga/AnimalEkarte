# FE-refactor.md — frontend/ 未消化バックログ

- **対象**: `frontend/`
- **作成日**: 2026-07-11 / **縮約日**: 2026-07-11
- **第 3 期（FE3-1〜14）は CLOSED**。実行手順やコミットハッシュ等の詳細は本書には残さない — git 履歴（`git log --grep='FE3-'`、完了コミット `904da793` ほか）が正本。
- 本書に残るのは以下の「未着手（別トラック）」項目のみ。新規着手前に各行の「実行しない理由」を確認すること。

## 未着手（別トラック）

| 項目 | 内容 | 実行しない理由 |
|---|---|---|
| **CI 赤の解消（最優先・オーナー対応）** | origin/main の Backend Lint 赤（`openapi_route_drift_test.go:672` builtinShadow、`billing_item_trimming_test.go` ×4 ほか）+ Codegen Sync 赤（`make codegen` 未実行の models.ts drift）+ 未 push ローカルコミット群 | backend 起因。FE 実行者が触ると事故る |
| AccountingListTable の支払方法表記 | `credit_card: "クレジットカード"`（AccountingListTable.tsx:25）vs 他画面 "カード" | どちらが正かは PO 判断のユーザー可視変更 |
| `{cond && <JSX>}` の機械強制 | `eslint-plugin-react`（jsx-no-leaked-render）等の導入。現状違反 0 件で導入即 green | devDependency 追加はオーナー判断 |
| `dangerouslySetInnerHTML` / PrintPortal XSS 監査 | セキュリティレビュー | 別途レビュー枠 |
| React Query `clinic_id` キャッシュ境界 | クリニック切替時のキャッシュ分離検証 | 調査タスク（結果次第で挙動変更） |
| FE zod ↔ Backend バリデーション二重管理 | 設計判断 | architect 判断 |
| MedicalRecordForm.tsx（411 行）の再分割 | モーダル JSX 群の抽出には大量の prop 引き回し判断が必要 | 800 行上限内・効果がリスクに見合わない |
| シフト休憩時刻 Input 4 箇所の aria-label | `ShiftFormDialog.tsx:304,311`・`ShiftTemplateSidePanelFields.tsx:138,145`（可視ラベルは「〜」区切りのみ） | ラベル文言の新規決定（例「休憩N 開始時刻」）が必要 = 文言発明は仕様追加 |
| medical-records の `Treatment` UI 型が #201 に未追随 | 生成型の `dose_*` 5 フィールド等を持たない stale twin | #201 の FE UI 実装（OPEN）側で対応すべき機能作業 |
| `TreatmentPlan`（src/types/index.ts）の stale twin 解消 | 生成型と同名・別形状。transform 層に寄せるかリネームか | transform 化は挙動リスク、リネームだけでは本質未解決 — 設計判断 |
| `BackendTrimming`（types/trimming.ts）の codegen 化 | 手書きミラー DTO（ファイル冒頭に手動同期の警告あり） | tygo がハンドラ DTO を出力しない設計に関わる |
| ForgotPasswordPage の anti-enumeration と toast の矛盾 | エラー時 `{status:"sent"}` を返しつつ handleApiError が toast する（列挙防止意図を toast が部分的に破る） | 挙動変更（セキュリティ文脈）— 別判断 |
| タイマー直値 2 件のコメント付与 | `use-medical-record-manual-errors.ts:35`（50ms）・`use-medical-record-form-modals.ts:28`（100ms） | コメントのみの nit。次に当該ファイルを触るコミットに同乗させる |
| `models.ts`（tygo 生成）・`design-tokens.ts` の分割 | 自動生成/定数カタログ | 対象外（確定済み）。なお旧バックログの「手書き models.ts」は**存在しない**ことを実測確認済み（誤記だった） |
