# フロントエンド リファクタリング計画

- **作成日**: 2026-07-07
- **対象**: `frontend/` — `src/`（メイン）＋ `liff/` ＋ `line-reserve/`
- **スタック**: React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query

対応済みの完了記録は git 履歴（commit `2a1ef3ad` ほか FE-R* コミット群、`48053b7e`）を正本とする。本書には未消化の別チケットのみを残す。

---

## 別チケット（本計画の範囲外）

| 項目 | 理由 |
|------|------|
| `dangerouslySetInnerHTML` / PrintPortal XSS 監査 | セキュリティレビュー別途 |
| クリニック切替時の React Query `clinic_id` キャッシュ境界 | 未検証（要調査） |
| FE zod と Backend バリデーションの二重管理 | architect 判断が必要 |
| `models.ts`（tygo 生成）・`design-tokens.ts` の分割 | 自動生成/定数カタログのため対象外 |
