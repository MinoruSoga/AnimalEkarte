# TASK-009: NotionFilter 未移行ページ完了 + フィルタUI要素サイズ拡大

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー（ブラウザテスト結果から）

---

## 概要

TASK-005（NotionFilter残移行）で対応予定だった見積一覧・マスタ設定ページのNotionFilter移行が未完了。加えて、NotionFilterのアイコン・テキストが小さすぎるため、現在の2倍サイズに拡大する。

## 問題点

### 1. 見積一覧（/estimates）— NotionFilter未適用
- 現状: タブフィルタ（すべて/下書き/送付済み/承認済み/却下）+ 検索アイコンのみ
- 期待: NotionFilterツールバー（フィルタを追加 + 並べ替え + 検索トグル）

### 2. マスタ設定ページ（13ページ）— NotionFilter未適用
- 現状: 件数表示 + 検索アイコンのみ（「+ フィルタを追加」「並べ替え」ボタンなし）
- 確認済みページ: スタッフマスタ、動物種類マスタ（いずれもNotionFilter未適用）
- 対象ページ: CageSettings, TrimmingSettings, MedicineSettings, DiagnosisSettings, StaffSettings, HospitalizationSettings, ServiceTypeSettings, InterviewTemplateSettings, ChiefComplaintSettings, AnimalSpeciesSettings, TreatmentPlanMaster, InsuranceSettings, JobTitleSettings

### 3. NotionFilter UI要素が小さすぎる
- 現状サイズ:
  - ツールバーアイコン: `h-3.5 w-3.5`（14px）
  - インラインアイコン: `h-3 w-3`（12px）
  - 全テキスト: `text-xs`（12px）
  - ボタンパディング: `px-2 py-1`
- 要望: **現在の2倍のサイズ**に拡大
  - ツールバーアイコン → `h-7 w-7`（28px）
  - インラインアイコン → `h-6 w-6`（24px）
  - テキスト → `text-base`（16px）以上
  - ボタンパディング → `px-4 py-2`

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | NotionFilter UI要素サイズ2倍化 | FE | FE-035 | - |
| 2 | 見積一覧の NotionFilter 移行 | FE | FE-036 | - |
| 3 | マスタ設定13ページの NotionFilter 移行 | FE | FE-037 | - |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend

**サイズ修正（FE-035）:**
- `frontend/src/components/shared/NotionFilter/NotionFilter.tsx` — ツールバーアイコン・テキストサイズ拡大
- `frontend/src/components/shared/NotionFilter/FilterAddPopover.tsx` — ポップオーバー内アイコン・テキスト拡大
- `frontend/src/components/shared/NotionFilter/FilterRuleRow.tsx` — フィルタルール行のサイズ拡大
- `frontend/src/components/shared/NotionFilter/SortPopover.tsx` — ソートポップオーバーのサイズ拡大
- `frontend/src/components/shared/NotionFilter/SortPill.tsx` — ソートピルのサイズ拡大

**見積一覧（FE-036）:**
- `frontend/src/features/estimates/routes/EstimateList.tsx` — NotionFilter統合

**マスタ設定（FE-037）:**
- `frontend/src/features/master/routes/*.tsx` — 13ページにNotionFilterツールバー追加

## 実装順序

1. FE-035（サイズ修正 — 他タスクの前にやることで全ページに効果が波及）
2. FE-036 + FE-037（並行可）

## 前提タスク

- TASK-003（NotionFilterコンポーネント作成）: 完了済み
- TASK-006（NotionFilter完全準拠）: 完了済み
