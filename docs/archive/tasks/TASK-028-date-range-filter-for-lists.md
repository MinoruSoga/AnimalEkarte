# TASK-028: 一覧ページへの日付範囲フィルタ追加

**作成日**: 2026-03-25
**ステータス**: Closed
**依頼元**: 健康診断、カルテ、会計管理、予防接種、定期健診の一覧ページにて、日付の範囲選択のフィルタを追加してください。

---

## 概要

5つの一覧ページに日付範囲フィルタを追加する。うち2ページはすでに実装済みのため、残り3ページが対象。

## 依頼内容（原文）

> 健康診断、カルテ、会計管理、予防接種、定期健診の一覧ページにて、日付の範囲選択のフィルタを追加してください。

## 仕様確認ログ

確認事項なし（既存の NotionFilter `date-range` パターンに準拠して実装する）

## 現状分析

| ページ名 | Feature | 日付フィルタ | 備考 |
|---------|---------|------------|------|
| 健康診断 (検査管理) | examinations | ✅ 実装済み | NotionFilter + date-range + BE対応 |
| 予防接種 | vaccinations | ✅ 実装済み | NotionFilter + date-range + BE対応 |
| 定期健診 | checkups | ❌ 未実装 | BE APIは`start_date`/`end_date`対応済み。FEのみ変更 |
| カルテ | medical-records | ❌ 未実装 | BE/FE 両方変更が必要 |
| 会計管理 | accounting | ❌ 未実装 | BE/FE 両方変更が必要（loaderData → useQuery 移行も必要） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | カルテ一覧API に日付範囲フィルタ追加（handler→service→repository） | BE | BE-056 | - | [x] |
| 2 | 会計管理一覧API に日付範囲フィルタ追加（handler→service→repository） | BE | BE-057 | - | [x] |
| 3 | 定期健診 - CheckupsListをNotionFilterに移行＋日付範囲フィルタ追加 | FE | FE-118 | - | [x] |
| 4 | カルテ一覧 - 日付範囲フィルタ追加 | FE | FE-119 | BE-056 | [x] |
| 5 | 会計管理一覧 - 日付範囲フィルタ追加（loaderData→useQuery移行含む） | FE | FE-120 | BE-057 | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 定期健診一覧で「フィルタ」を追加すると「日付（date-range）」フィルタが表示され、期間を選択すると絞り込まれる
- [ ] AC-2: カルテ一覧で「フィルタ」→「日付」から期間を選択するとAPIに`start_date`/`end_date`が送信され、診療日で絞り込まれる
- [ ] AC-3: 会計管理一覧で「フィルタ」→「日付」から期間を選択するとAPIに`start_date`/`end_date`が送信され、会計日（scheduled_date）で絞り込まれる
- [ ] AC-4: 日付フィルタなし（デフォルト状態）では全件が表示される
- [ ] AC-5: 既存のステータスフィルタ・テキスト検索・ソートと日付フィルタが共存して動作する（会計管理）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 日付フィルタのフィルタリング実装 | サーバーサイドフィルタ（APIクエリパラメータ） | 既存の vaccinations/examinations と同パターン | クライアントサイドフィルタ（全件取得後にメモリ内絞り込み） |
| 会計管理のデータ取得方式 | loaderData → useGetAccountings hook に移行 | 日付フィルタをリアクティブに渡すにはstateが必要。loaderは初回のみ実行のため | loaderData を維持して URL params でフィルタ |
| 定期健診のUI刷新 | CheckupsList を NotionFilter に移行（古いシンプルUIを廃止） | 他の全一覧ページが NotionFilter を使用。一貫性のため | 古いUIに日付入力を追加 |

## 影響範囲

### DB
- 変更なし

### Backend
- `backend/internal/handler/medical_record_handler.go` — `ListMedicalRecords`: `start_date`/`end_date` クエリパラメータ追加
- `backend/internal/service/medical_record_service.go` — `List()` シグネチャに `startDate, endDate *string` 追加
- `backend/internal/repository/medical_record_repository.go` — `FindAll()` シグネチャ・クエリに `startDate, endDate *string` 追加
- `backend/internal/handler/accounting_handler.go` — `ListAccountings`: `start_date`/`end_date` クエリパラメータ追加
- `backend/internal/service/accounting_service.go` — `List()` シグネチャに `startDate, endDate *string` 追加
- `backend/internal/repository/accounting_repository.go` — `FindAll()` シグネチャ・クエリに `startDate, endDate *string` 追加

### Frontend
- `frontend/src/features/checkups/routes/CheckupsList.tsx` — NotionFilter移行 + date-range フィルタ
- `frontend/src/features/medical-records/api/get-medical-records.ts` — `MedicalRecordFilters` 型追加、APIに日付パラメータ追加
- `frontend/src/features/medical-records/hooks/use-medical-records.ts` — `useFilterMedicalRecords` に date フィルタ引数追加
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx` — NotionFilter に date-range プロパティ追加
- `frontend/src/features/accounting/api/get-accountings.ts` — `AccountingFilters` 型追加、APIに日付パラメータ追加
- `frontend/src/features/accounting/routes/Accounting.tsx` — loaderData → useGetAccountings 移行 + date-range フィルタ
- `frontend/src/features/accounting/loaders.ts` — 不要になる可能性（要検討）

## 参照実装

- `features/vaccinations/` — date-range フィルタの完全実装パターン（get-vaccinations.ts + use-vaccinations.ts + VaccinationList.tsx）
- `features/examinations/` — 同パターンの別実装（useFilterExaminationRecords + Examinations.tsx）

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 会計管理のloaderData→useQuery移行で初期ローディング状態の表示が変わる | 低 | `isLoading` 状態をAccountingコンポーネントで処理 |
| カルテのページネーションがAPIの`total`を使っているため、フィルタ後の件数が正しく反映されない | 中 | APIのpagination totalをフィルタ後件数で返すように確認 |

## 未解決事項

なし

## 実装順序

1. BE-056（カルテAPI日付フィルタ）
2. BE-057（会計管理API日付フィルタ）
3. FE-118（定期健診 - BE変更不要なので先行可能）
4. FE-119（カルテFE - BE-056完了後）
5. FE-120（会計管理FE - BE-057完了後）

## 関連イシュー

- BE-056: [カルテ一覧API日付範囲フィルタ追加](../../backend/issues/open/BE-056-medical-records-date-range-filter.md)
- BE-057: [会計管理一覧API日付範囲フィルタ追加](../../backend/issues/open/BE-057-accounting-date-range-filter.md)
- FE-118: [定期健診一覧 NotionFilter移行＋日付範囲フィルタ](../../frontend/issues/open/FE-118-checkups-notionfilter-date-range.md)
- FE-119: [カルテ一覧 日付範囲フィルタ追加](../../frontend/issues/open/FE-119-medical-records-date-range-filter.md)
- FE-120: [会計管理一覧 日付範囲フィルタ追加](../../frontend/issues/open/FE-120-accounting-date-range-filter.md)
