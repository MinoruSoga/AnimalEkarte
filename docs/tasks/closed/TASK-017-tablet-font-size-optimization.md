# TASK-017: 全一覧ページのフォントサイズをタブレット最適化（最低 text-base）

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー

---

## 概要

すべての一覧ページで最低フォントサイズを `text-base`（16px）に統一する。テーブルセル・ヘッダー・ボタン・バッジ等の付随要素も対象。タブレット端末での視認性を改善する。

## 依頼内容（原文）

> すべての一覧ページにて、文字の大きさをタブレットに最適なサイズにしてください。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 具体値: A(text-base統一)/B(特定サイズ)/C(レスポンシブ) | A — 最低フォントサイズを text-base に |
| 2 | スコープ: マスタ設定・ダッシュボード・予約も含むか | すべて含む |
| 3 | テーブルヘッダーも対象か | 対象 |
| 4 | ボタン・バッジ等の付随要素も対象か | 対象 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | design-tokens.ts の STYLE プリセット更新 | FE | FE-067 | - | [x] |
| 2 | 共有コンポーネントの text-xs/text-sm 置換 | FE | FE-068 | #1 | [x] |
| 3 | 一覧ページの text-xs/text-sm 置換 | FE | FE-069 | #1 | [x] |
| 4 | マスタ設定ページの text-xs/text-sm 置換 | FE | FE-070 | #1 | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 全一覧ページ（飼主、ワクチン、トリミング、在庫、見積、会計、入院、予約、ダッシュボード）で `text-xs` / `text-sm` のテキストが存在しない
- [ ] AC-2: 全マスタ設定ページ（12画面）で `text-xs` / `text-sm` のテキストが存在しない
- [ ] AC-3: 共有コンポーネント（StatusBadge、StatusPill、Pagination、NotionFilter）の最低フォントサイズが text-base
- [ ] AC-4: `npm run build` がエラーなく通る
- [ ] AC-5: `npm run lint` がエラーなく通る
- [ ] AC-6: 既存の UI レイアウトが大きく崩れない（文字サイズ拡大による軽微な幅変更は許容）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 変更の起点 | design-tokens.ts の STYLE プリセット更新を先に行う | STYLE 経由のページは自動的に修正される。波及効果が大きい | 各ページを個別修正 |
| ルール | `text-xs` → `text-base`、`text-sm` → `text-base` に一律置換 | 「最低 text-base」のルール。text-base 以上はそのまま維持 | レスポンシブ分岐 |

## 影響範囲

### Frontend — design-tokens.ts（STYLE プリセット）
- `tableHeaderCell`: `text-xs` → `text-base`
- `tableEmpty`: `text-sm` → `text-base`
- `formHeaderTitle`: `text-sm` → `text-base`
- `formHeaderDesc`: `text-xs` → `text-base`
- `searchInput`: `text-sm` → `text-base`
- `paginationBtnActive`: `text-sm` → `text-base`
- `paginationInfo`: `text-xs` → `text-base`
- `propertyLabel`: `text-sm` → `text-base`
- `propertyInput`: `text-sm` → `text-base`
- `btnPrimary`/`btnAccent`/`btnDanger`/`btnOutline`: `text-sm` → `text-base`
- `badge`: `text-sm` → `text-base`
- `sectionLabel`: `text-xs` → `text-base`
- `selectCompact`: `text-sm` → `text-base`
- `sidePeekCancelBtn`/`sidePeekSaveBtn`: `text-sm` → `text-base`
- `formLabel`/`formInput`/`formInputLight`: `text-sm` → `text-base`
- `inlineAddBtn`: `text-sm` → `text-base`

### Frontend — 共有コンポーネント
- `StatusBadge.tsx`: `text-sm` → `text-base`
- `NotionStatusPill.tsx`: `text-xs` → `text-base`
- `Pagination.tsx`: `text-sm` → `text-base`
- `NotionFilter/FilterRuleRow.tsx`: 多数の `text-sm` → `text-base`
- `NotionFilter/SortPopover.tsx`: `text-sm` → `text-base`
- `NotionFilter/FilterAddPopover.tsx`: `text-sm` → `text-base`
- `NotionFilter/SortPill.tsx`: `text-sm`/`text-xs` → `text-base`

### Frontend — 一覧ページ
- `VaccinationList.tsx`: 全 `text-sm` → `text-base`
- `TrimmingList.tsx`: 全 `text-sm` → `text-base`
- `InventoryList.tsx`: 全 `text-sm` → `text-base`
- `EstimateList.tsx`: 全 `text-sm` → `text-base`
- `Accounting.tsx`: 全 `text-sm` → `text-base`
- `ReservationManagement.tsx`: `text-sm`/`text-xs` → `text-base`
- `Dashboard.tsx`: `text-sm`/`text-xs` → `text-base`

### Frontend — マスタ設定ページ（12+画面）
- `InsuranceSettings.tsx`, `StaffSettings.tsx`, `JobTitleSettings.tsx`
- `AnimalSpeciesSettings.tsx`, `ChiefComplaintSettings.tsx`
- `HospitalizationSettings.tsx`, `ServiceTypeSettings.tsx`
- `InterviewTemplateSettings.tsx`, `DiagnosisSettings.tsx`
- `TrimmingSettings.tsx`, `MedicineSettings.tsx`, `CageSettings.tsx`
- `CompanySettings.tsx`, `MasterSettingsIndex.tsx`
- `TreatmentPlanMaster.tsx`

## 参照実装

- `features/owners/routes/OwnersList.tsx` — `STYLE.tableCell`（text-base）を正しく使用している参照実装

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 文字サイズ拡大によるテーブルカラム幅の溢れ | 中 | 変更後に全画面を目視確認。`truncate`/`max-w` で対処 |
| NotionFilter ポップオーバー内のレイアウト崩れ | 低 | ポップオーバーの width 調整が必要になる可能性あり |
| StatusPill の text-xs → text-base でバッジが大きくなりすぎる | 中 | padding 調整を検討 |

## 未解決事項

- なし

## 実装順序

1. FE-067: design-tokens.ts STYLE プリセット更新（波及効果最大）
2. FE-068: 共有コンポーネントのハードコード置換
3. FE-069: 一覧ページのハードコード置換
4. FE-070: マスタ設定ページのハードコード置換

## 関連イシュー

- [FE-067: design-tokens.ts STYLE プリセットのフォントサイズ最低 text-base 化](../../frontend/issues/open/FE-067-design-tokens-font-size-base.md)
- [FE-068: 共有コンポーネントのフォントサイズ最低 text-base 化](../../frontend/issues/open/FE-068-shared-components-font-size-base.md)
- [FE-069: 一覧ページのフォントサイズ最低 text-base 化](../../frontend/issues/open/FE-069-list-pages-font-size-base.md)
- [FE-070: マスタ設定ページのフォントサイズ最低 text-base 化](../../frontend/issues/open/FE-070-master-settings-font-size-base.md)
