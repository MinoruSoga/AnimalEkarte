# TASK-011: フォーム要素のNotion UI整合性改善

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー（全ページChromeブラウザ確認後）

---

## 概要

全フォーム画面をChromeで確認した結果、Notion風UIがほぼ実装されているが、**色のhardcoding**と**Tailwind gray色の混在**が整合性を低下させている。design-tokens.ts のトークン体系は完璧だが、使用ルールが徹底されていない箇所がある。

## 依頼内容（原文）

> 各ページのフォーム要素にて、NotionライクなUIになっているか確認し、改善箇所をタスクにしてください。

## Chrome確認結果サマリー

### Notion準拠済み（問題なし）
- OwnerForm.tsx（参照実装）: 完全準拠
- PetSelectionSearchForm.tsx: ほぼ準拠
- Dashboard.tsx: Kanbanカード含め準拠
- マスタ設定サイドピーク: Notion風プロパティ表示
- 一覧画面のNotionFilter: 準拠

### 改善が必要な画面
- **VaccinationForm.tsx**: 黒ボタン(`bg-black`)、Tailwind gray hover
- **ExaminationForm.tsx**: 赤ボタン(`text-red-600`)のhardcoding
- **HospitalizationForm.tsx**: 赤ボタンのhardcoding
- **全フォーム共通**: border/text60/ボタン色の直書き → デザイントークン未使用

## 問題分類

### Critical（視覚的に不整合）
1. **黒色ボタン**: `bg-black text-white` — Notion UIに存在しない色。VaccinationForm のクリア・複製ボタン
2. **Tailwind gray hover**: `hover:bg-gray-50` — Notion暖色（`#F7F6F3`）ではなくcool gray
3. **赤色ボタン直書き**: `text-red-600 hover:text-red-700 hover:bg-red-50` — デザイントークン未使用

### Medium（一貫性不足）
4. **主要ボタン色 inline**: `bg-[#2383E2]` を毎回直書き → `STYLE.confirmPrimary` 未使用
5. **border 直書き**: `border-[rgba(55,53,47,0.16)]` を毎回直書き → `C.borderMedium` 未使用
6. **text60 直書き**: `text-[#37352F]/60` を毎回直書き → `C.text60` 未使用

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | フォーム要素のNotion UI整合性改善（色hardcoding→デザイントークン統一） | FE | FE-039 | - |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx` — 黒ボタン、gray hover
- `frontend/src/features/examinations/routes/ExaminationForm.tsx` — 赤ボタン、主要ボタン直書き
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx` — 赤ボタン、主要ボタン直書き
- `frontend/src/features/medical-records/` 配下 — 同様のhardcoding箇所
- `frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx` — 検索ボタン直書き
- `frontend/src/lib/design-tokens.ts` — 不足するトークンの追加

## 実装順序

1. FE-039（全フォーム画面の色hardcoding → デザイントークン統一）

## 関連イシュー

- FE-039: [フォーム要素Notion UI整合性改善](../../frontend/issues/closed/FE-039-notion-ui-form-consistency.md)
