# BUG-370: 月末未納者一覧（売掛金画面表示）

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: HIGH（月次会計処理に必要）
**Affects**: `features/accounting`, `billings.status`, `billings.scheduled_date`

## 実装結果（2026-04-14）

### 完了
- BE: `AccountingRepository` に `FindUnpaidByBilling` / `FindUnpaidByOwner` 追加
  - 飼主単位は GROUP BY owner_id で 1 クエリ集約（N+1 回避）
  - サマリー (`UnpaidSummary`) + 詳細 (`UnpaidOwnerAggregate`) を返却
  - `billings.status=waiting` かつ `scheduled_date < base_date` かつ `deleted_at IS NULL`
- BE: `AccountingService.ListUnpaidByBilling` / `ListUnpaidByOwner` 追加
- BE: `ListUnpaidBillings` handler（GET `/v1/accountings/unpaid`）
  - `group_by=owner|billing`, `base_date=YYYY-MM-DD` (default: 今日), `page`, `limit`
- BE: ルート登録 `accountings.GET("/unpaid", ...)`
- BE: `unpaidByOwnerResponse` / `unpaidSummaryResponse` response 型追加
- BE: mock repository に 2 メソッド追加（既存テスト維持）
- BE: build / vet 成功
- FE: `get-unpaid-billings.ts` — `useGetUnpaidByOwner` / `useGetUnpaidByBilling` 新規
- FE: `routes/UnpaidCustomerList.tsx` 新規ページ
  - 基準日入力（default=今日）
  - タブ切替: 飼主単位 / 会計単位（URL `?group_by=` 同期）
  - サマリーカード（売掛金総額 / 件数 / 飼主数）
  - 飼主単位テーブル（飼主名・件数・未納額・最古／最新日・経過日数）
  - 行クリックで飼主詳細画面遷移
  - ページネーション
  - 空状態「未納者はいません」
- FE: `config/paths.ts` に `paths.accounting.unpaid` 追加
- FE: `router.tsx` にルート `/accounting/unpaid` 追加
- FE: build / lint 成功（エラー 0）

### 追加対応（2026-04-14）
- FE: 会計単位タブ完成（飼主名・ペット名・診療日・未納額・経過日数）
  - `useGetUnpaidByBilling` を `transformToAccounting` 経由でドメイン型に変換
  - `b.items.subtotal + taxAmount` で総額算出（waiting status のため payment 未確定を考慮）
  - 行クリックで AccountingDetail へ遷移
  - ページネーション両タブ共通対応

### 未完了（別イシュー化候補）
- サイドバー/メニューから「未納者一覧」への導線追加
- `scheduled_date` の浮動挙動（今日日付を含むかのタイムゾーン考慮）

**依頼元（原文）**:

> 月末に未納者一覧を出せるようにしてほしい　会計上、未納者の計上が必要

---

## 概要

月末締めで売掛金額を把握するため、基準日時点で未納（=`status=waiting` かつ `scheduled_date < 基準日`）の billings を一覧表示する。集計粒度は **飼主単位 / 会計単位の2タブ**、未納額合計（売掛金総額）を画面に表示する。**スナップショット保存・印刷・督促管理は対象外**（必要になった時点で別 issue 起票）。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| Q1 | 未納の定義 | **(A) `status=waiting` かつ `scheduled_date < 基準日`** の全件 |
| Q2 | 集計粒度 | **(C) 両方**（飼主単位 / 会計単位のタブ切替） |
| Q3 | 月末スナップショット | **(A) 動的クエリのみ**（保存しない） |
| Q4 | 「会計上の計上」 | **(A) 画面に未納額合計を表示するだけ** |
| Q5 | 印刷帳票 | **不要** |
| Q6 | 督促管理 | **不要** |
| Q7 | 権限 | デフォルト（`accounting` 権限） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | 未納一覧 API（飼主単位 / 会計単位） | BE | BE-110 | - |
| 2 | 未納一覧画面（タブ切替・売掛金総額表示） | FE | FE-248 | #1 |

## 受入条件（Acceptance Criteria）

- [ ] **AC-1**: 会計メニューから「未納者一覧」画面を開ける
- [ ] **AC-2**: 画面上部に「基準日」入力欄があり、デフォルトは**今日**。任意の日付を選択可能
- [ ] **AC-3**: 「飼主単位」「会計単位」のタブが切替可能で、URL クエリ（`?group_by=owner|billing`）と同期する
- [ ] **AC-4**: **飼主単位タブ**: 同じ飼主の複数未納 billing を 1 行に集約。表示項目: 飼主名 / 件数 / 未納額合計 / 最古未納日 / 最新未納日 / 経過日数（最古日からの日数）
- [ ] **AC-5**: **会計単位タブ**: 既存 `AccountingList` と同じ粒度で、未納のもののみを抽出。表示項目: 飼主名 / ペット名 / 診療日 / 未納額 / 経過日数 / 詳細リンク
- [ ] **AC-6**: 画面上部に「**売掛金総額**: ¥X,XXX,XXX（N件 / N名）」のサマリーカードを表示
- [ ] **AC-7**: 抽出条件は **`billings.status=waiting` かつ `billings.scheduled_date < 基準日` かつ `deleted_at IS NULL`**
- [ ] **AC-8**: `cancelled` / `completed` の billing は除外
- [ ] **AC-9**: デフォルトソートは **経過日数の降順**（古い未納が上）
- [ ] **AC-10**: ページネーション対応（既存 `usePagination` パターン踏襲）
- [ ] **AC-11**: 未納が 0 件の場合「未納者はいません」の空状態を表示
- [ ] **AC-12**: 飼主単位タブから飼主名クリックで `OwnerDetail` 画面に遷移
- [ ] **AC-13**: 会計単位タブから行クリックで `AccountingDetail` 画面に遷移

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| DB スキーマ変更 | **なし** | 既存 `billings.status` + `scheduled_date` で抽出可能 | 新規フラグ追加 — Q1(A) のため不要 |
| スナップショット保存 | **なし** | Q3(A) で動的クエリのみ | スナップショットテーブル新設 — 監査要件無し |
| 印刷機能 | **なし** | Q5 で不要 | A4 帳票 |
| API エンドポイント | **新規 1 本**（既存 ListAccountings には統合せず分離） | 飼主単位集約は `GROUP BY owner_id` が必要で SQL ロジックが大きく異なる | 既存 ListAccountings に `group_by` パラメータ追加 |
| 集計粒度の切替 | URL クエリ同期 | 画面リロード・ブックマーク対応 | state のみ |
| キャッシュ戦略 | `useQuery` 標準（`staleTime: MEDIUM` 5分） | 売掛金額は数分の遅延許容範囲 | リアルタイム |

## 影響範囲

### Backend
- `backend/internal/repository/accounting_repository.go` — `FindUnpaidByBilling` / `FindUnpaidByOwner` メソッド追加
- `backend/internal/service/accounting_service.go` — `ListUnpaid` メソッド追加
- `backend/internal/handler/accounting_handler.go` — `ListUnpaidBillings` ハンドラ追加
- `backend/internal/handler/accounting_response.go` — `unpaidByOwnerResponse` 型追加
- `backend/cmd/api/main.go` — ルート追加

### Frontend
- `frontend/src/features/accounting/api/get-unpaid-billings.ts` — 新規 API hook
- `frontend/src/features/accounting/routes/UnpaidCustomerList.tsx` — 新規ページ
- `frontend/src/features/accounting/index.ts` — エクスポート追加
- `frontend/src/app/router.tsx` — ルート追加 `/accounting/unpaid`
- `frontend/src/components/shared/Layout/` のメニューに「未納者一覧」項目追加
- `frontend/src/config/paths.ts` — `paths.accounting.unpaid` 追加

### DB
- **変更なし**

## 参照実装

- `frontend/src/features/accounting/routes/AccountingList.tsx` — フィルタ・ソート・ページネーションのパターン
- `backend/internal/repository/accounting_repository.go:31` — `FindAll` の clinicScope と JOIN 構造

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| `status=waiting` に「当日の会計待ち」が混入 | 低 | `scheduled_date < 基準日` 条件で当日分は自動除外。仕様通り |
| 飼主単位集約時の N+1 クエリ | 中 | 集計は SQL の `GROUP BY owner_id` で 1 クエリで実行。アプリ層ループ禁止 |
| 削除済 billings の混入 | 高 | クエリで `deleted_at IS NULL` を必ず付与（既存 GORM SoftDelete で自動だが明示も追加） |
| 返金（billing_refunds）がある場合の未納額 | 中 | スコープ外（Q4=A 単純表示のため）。`billings.total_amount` をそのまま使用 |

## 未解決事項

- なし

## 実装順序

1. BE-110: API 実装（repository → service → handler → ルート）
2. FE-248: 画面実装（API hook → コンポーネント → ルート登録）

## 関連イシュー

- BE-110: 未納 billings 一覧 API
- FE-248: 未納者一覧画面
