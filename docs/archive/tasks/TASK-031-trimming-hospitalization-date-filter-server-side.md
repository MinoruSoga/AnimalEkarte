# TASK-031: トリミング・入院ホテル一覧 日付フィルタのサーバーサイド対応

**作成日**: 2026-03-26
**ステータス**: Open
**依頼元**: 一覧ページの日付範囲フィルタ抜け漏れ調査で発見

---

## 概要

トリミング一覧・入院ホテル一覧は日付範囲フィルタ UI が存在するが、全件ロード後クライアントサイドで絞り込む実装になっている。他の主要5ページ（カルテ・検査管理・会計管理・予防接種・定期健診）と同様に、サーバーサイドフィルタへ移行する。

## 依頼内容（原文）

> 一覧ページの日付範囲フィルタ抜け漏れ調査で発見。トリミング・入院ホテルのみクライアントサイドフィルタのまま残存。

## 仕様確認ログ

確認事項なし（既存の vaccination/examination と同パターンに準拠）

## 現状分析

| ページ | DB カラム | UI フィルタ | FE→BE 連携 | BE サポート |
|--------|----------|-----------|-----------|-----------|
| トリミング | `trimmings.date` | ✅ date-range | ❌ パラメータなし | ❌ なし |
| 入院・ホテル | `hospitalizations.start_date` | ✅ date-range | ❌ パラメータなし | ❌ なし |

### 現状の問題点

**トリミング** (`get-trimmings.ts`):
```typescript
export const getTrimmings = async (): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings");
  // ↑ クエリパラメータなし。全件取得後 use-trimming-records.ts でクライアントサイドフィルタ
```

**入院・ホテル** (`get-hospitalizations.ts`):
```typescript
export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations");
  // ↑ クエリパラメータなし。全件取得後 HospitalizationList.tsx でクライアントサイドフィルタ
```

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | トリミング API に日付範囲フィルタ追加 | BE | BE-063 | - | [ ] |
| 2 | 入院・ホテル API に日付範囲フィルタ追加 | BE | BE-064 | - | [ ] |
| 3 | トリミング一覧 日付フィルタのサーバーサイド移行 | FE | FE-126 | BE-063 | [ ] |
| 4 | 入院・ホテル一覧 日付フィルタのサーバーサイド移行 | FE | FE-127 | BE-064 | [ ] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: トリミング一覧で日付フィルタを設定すると `GET /v1/trimmings?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` が送信される
- [ ] AC-2: 入院・ホテル一覧で日付フィルタを設定すると `GET /v1/hospitalizations?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` が送信される
- [ ] AC-3: 日付フィルタなし（デフォルト）では全件が表示される
- [ ] AC-4: 既存のステータス・種フィルタ・テキスト検索と共存して動作する

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| フィルタリング方式 | サーバーサイド（APIクエリパラメータ） | 他5ページと一致。大量データ時のパフォーマンス確保 | クライアントサイド維持 |
| 参照パターン | `examination_handler.go` / `useFilterExaminationRecords` | 既存の最も近いパターン | - |

## 影響範囲

### Backend
- `backend/internal/handler/trimming_handler.go` — `ListTrimmings`: `start_date`/`end_date` クエリパラメータ追加
- `backend/internal/service/trimming_service.go` — `List()` シグネチャに `startDate, endDate *string` 追加
- `backend/internal/repository/trimming_repository.go` — `FindAll()` に日付 WHERE 条件追加
- `backend/internal/handler/hospitalization_handler.go` — `ListHospitalizations`: `start_date`/`end_date` 追加
- `backend/internal/service/hospitalization_service.go` — `List()` シグネチャ拡張
- `backend/internal/repository/hospitalization_repository.go` — `FindAll()` に日付 WHERE 条件追加

### Frontend
- `frontend/src/features/trimming/api/get-trimmings.ts` — `TrimmingFilters` 型追加・クエリパラメータ送信
- `frontend/src/features/trimming/hooks/use-trimming-records.ts` — クライアントサイド日付フィルタ削除・サーバーサイドへ移行
- `frontend/src/features/trimming/routes/TrimmingList.tsx` — `filters` を `useFilterTrimmingRecords` に渡すよう修正
- `frontend/src/features/hospitalization/api/get-hospitalizations.ts` — `HospitalizationFilters` 型追加
- `frontend/src/features/hospitalization/routes/HospitalizationList.tsx` — クライアントサイド日付フィルタ削除・サーバーサイドへ移行

## 参照実装

- `features/vaccinations/api/get-vaccinations.ts` + `hooks/use-vaccinations.ts` — FE サーバーサイドフィルタパターン
- `backend/internal/handler/examination_handler.go` — `parseDateQuery` パターン
- `backend/internal/repository/examination_repository.go` — GORM 日付 WHERE 条件パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| `useHospitalizationList` フックが `activeFilters` を受け取らない設計のため改修範囲がやや広い | 低 | `HospitalizationList.tsx` に直接 `useGetHospitalizations(filters)` を追加して対応 |

## 実装順序

1. BE-063（トリミング API 日付フィルタ）
2. BE-064（入院ホテル API 日付フィルタ）
3. FE-126（トリミング FE 移行 — BE-063 完了後）
4. FE-127（入院ホテル FE 移行 — BE-064 完了後）

## 関連イシュー

- BE-063: [トリミング API 日付範囲フィルタ追加](../../backend/issues/open/BE-063-trimming-date-range-filter.md)
- BE-064: [入院・ホテル API 日付範囲フィルタ追加](../../backend/issues/open/BE-064-hospitalization-date-range-filter.md)
- FE-126: [トリミング一覧 日付フィルタ サーバーサイド移行](../../frontend/issues/open/FE-126-trimming-date-filter-server-side.md)
- FE-127: [入院・ホテル一覧 日付フィルタ サーバーサイド移行](../../frontend/issues/open/FE-127-hospitalization-date-filter-server-side.md)
