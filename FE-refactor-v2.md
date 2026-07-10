# FE-refactor-v2.md — frontend/ リファクタリング未消化記録

- **作成日**: 2026-07-10
- **対象**: `frontend/`（main / liff / line-reserve）

対応済みの完了記録（旧 FE-refactor.md 分・本書 FE-R1〜FE-R17 分）は git 履歴（commit `c8962b49` ほか FE-R* コミット群）を正本とする。本書には未消化の別トラックのみを残す。

---

## 別トラック（本計画では実行しなかった・記録のみ）

| 項目 | 内容 | 実行しない理由 |
|---|---|---|
| useMasterSave validationError fix | `toast.error(error)` 追加（WIP 3 ファイル: use-master-save.ts / use-master-save.test.ts / StaffSettings.test.tsx） | 別セッションが作業中の挙動変更。本計画では触っていない |
| AccountingListTable の支払方法表記 | `credit_card: "クレジットカード"`（AccountingListTable.tsx:25）vs 他画面 "カード" の表記分岐 | どちらが正かは PO 判断のユーザー可視変更。統一する場合は共有定数（daily-accounting-utils.ts）に寄せる |
| `{cond && <JSX>}` の機械強制 | `eslint-plugin-react`（jsx-no-leaked-render）等の導入 | devDependency 追加はオーナー判断。現状違反0件のため導入即 green |
| R-F17 統合フェーズ（calcAgeAt 共通化） | pets の年齢計算統合 | 挙動変更を伴う（旧 FE-refactor.md で `fix:` 分離確定済み） |
| dangerouslySetInnerHTML / PrintPortal XSS 監査 | セキュリティレビュー | 別途レビュー枠 |
| React Query の clinic_id キャッシュ境界 | クリニック切替時のキャッシュ分離検証 | 調査タスク（結果次第で挙動変更） |
| FE zod ↔ Backend バリデーション二重管理 | 設計判断 | architect 判断が必要 |
| OwnerInfoFieldSections.tsx (441行) / MedicalRecordForm.tsx (411行) の再分割 | R-F19 分割後も400行超の残差 | 800行ハード上限内で優先度 LOW。JSX 移動リスクが効果に見合わない |
| shadow rgba 1件（ShiftTemplateSettingsParts.tsx:225） | design-audit の C1/C3/C5 対象外の cosmetic nit | 既存 STYLE 定数と rounded 指定が異なり単純置換不可 |
| `src/types/` 手書き型 vs `types/generated/` 重複棚卸し | FE-R6 で個別シンボルは処理済みだが全量監査は未実施 | tygo 生成側の設計に関わるため別途 |
