# TASK-002: 各一覧ページでのフィルタ追加

**作成日**: 2026-03-17
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

会計一覧に未払いフィルタ、健康診断・予防接種一覧に日付範囲フィルタを追加する。

## 依頼内容（原文）

> 各一覧ページでのフィルタ
> - 会計一覧
>     - 未払い者のフィルタを追加
> - 健康診断、予防接種
>     - 日付の範囲指定

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | 健康診断 API に日付範囲フィルタ追加 | BE | BE-041 | - |
| 2 | 予防接種 API に日付範囲フィルタ追加 | BE | BE-042 | - |
| 3 | 会計一覧に未払いフィルタUI追加 | FE | FE-014 | - |
| 4 | 健康診断一覧に日付範囲フィルタUI追加 | FE | FE-015 | #1 |
| 5 | 予防接種一覧に日付範囲フィルタUI追加 | FE | FE-016 | #2 |

## 影響範囲

### DB
- 変更なし（既存の `date` カラムを使用）

### Backend
- `backend/internal/handler/examination_handler.go` — start_date/end_date パラメータ追加
- `backend/internal/service/examination_service.go` — List シグネチャ変更
- `backend/internal/repository/examination_repository.go` — WHERE 日付範囲追加
- `backend/internal/handler/vaccination_handler.go` — 同上
- `backend/internal/service/vaccination_service.go` — 同上
- `backend/internal/repository/vaccination_repository.go` — 同上

### Frontend
- `frontend/src/features/accounting/routes/Accounting.tsx` — ステータスフィルタUI追加
- `frontend/src/features/accounting/api/get-accountings.ts` — status パラメータ追加
- `frontend/src/features/examinations/routes/Examinations.tsx` — 日付範囲フィルタUI追加
- `frontend/src/features/examinations/api/get-examinations.ts` — 日付パラメータ追加
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx` — 日付範囲フィルタUI追加
- `frontend/src/features/vaccinations/api/get-vaccinations.ts` — 日付パラメータ追加

## 実装順序

1. BE-041 / BE-042（API フィルタ追加 — 並行可）
2. FE-014（会計 未払いフィルタ — 独立、即着手可）
3. FE-015 / FE-016（日付範囲UI — BE完了後）

## 関連イシュー

- [BE-041: 健康診断 API 日付範囲フィルタ](../backend/issues/open/BE-041-examinations-date-range-filter.md)
- [BE-042: 予防接種 API 日付範囲フィルタ](../backend/issues/open/BE-042-vaccinations-date-range-filter.md)
- [FE-014: 会計一覧 未払いフィルタUI](../frontend/issues/open/FE-014-accounting-list-unpaid-filter.md)
- [FE-015: 健康診断一覧 日付範囲フィルタUI](../frontend/issues/open/FE-015-examinations-date-range-filter-ui.md)
- [FE-016: 予防接種一覧 日付範囲フィルタUI](../frontend/issues/open/FE-016-vaccinations-date-range-filter-ui.md)
