# フロントエンド リファクタリング計画

- **作成日**: 2026-07-07
- **対象**: `frontend/` — `src/`（メイン）＋ `liff/` ＋ `line-reserve/`
- **スタック**: React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query

本書は未消化バックログのみを記録する。対応済みの完了記録は git 履歴（commit `2a1ef3ad` ほか FE-R* コミット群、`48053b7e`、`14849cdc`（calcAgeAt 統一）、および v2 側 FE-R* コミット群 `c8962b49` ほか）を正本とする。

---

## 未消化バックログ

| 項目 | 理由 |
|------|------|
| AccountingListTable の支払方法表記 | `credit_card: "クレジットカード"`（AccountingListTable.tsx:25）vs 他画面 "カード" の表記分岐。どちらが正かは PO 判断のユーザー可視変更。統一する場合は共有定数（daily-accounting-utils.ts）に寄せる |
| `{cond && <JSX>}` の機械強制 | `eslint-plugin-react`（jsx-no-leaked-render）等の導入。devDependency 追加はオーナー判断。現状違反0件のため導入即 green |
| `dangerouslySetInnerHTML` / PrintPortal XSS 監査 | セキュリティレビュー別途 |
| クリニック切替時の React Query `clinic_id` キャッシュ境界 | 未検証（要調査）。クリニック切替時のキャッシュ分離検証、結果次第で挙動変更 |
| FE zod と Backend バリデーションの二重管理 | architect 判断が必要 |
| OwnerInfoFieldSections.tsx（441行）/ MedicalRecordForm.tsx（411行）の再分割 | R-F19 分割後も400行超の残差。800行ハード上限内で優先度 LOW。JSX 移動リスクが効果に見合わない |
| shadow rgba 1件（ShiftTemplateSettingsParts.tsx:225） | design-audit の C1/C3/C5 対象外の cosmetic nit。既存 STYLE 定数と rounded 指定が異なり単純置換不可 |
| `src/types/` 手書き型 vs `types/generated/` 重複棚卸し | FE-R6 で個別シンボルは処理済みだが全量監査は未実施。tygo 生成側の設計に関わるため別途 |
| `models.ts`（tygo 生成）・`design-tokens.ts` の分割 | 自動生成/定数カタログのため対象外 |
