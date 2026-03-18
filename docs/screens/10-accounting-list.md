# 会計一覧 仕様書

## 概要
- **画面の目的**: 会計レコードの一覧管理。ステータス別フィルター・保険フィルター・検索・ソートが可能。
- **URLパターン**: `/accounting`
- **コンポーネント**: `[R] Accounting`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ 会計管理  [新規会計登録ボタン]  (CreditCardアイコン)  │
├──────────────────────────────────────────────────────┤
│ 🔍 [飼主名、ペット名...検索]  [全て|保険あり|保険なし] │
├──────────────────────────────────────────────────────┤
│ テーブル（ソート可能ヘッダー）                       │
│ 日時 | 飼主名 | ペット名 | 請求金額 | 支払方法 |     │
│ 保険 | ソース | ステータス | 操作                    │
├──────────────────────────────────────────────────────┤
│ ページネーション（20件/ページ）                      │
└──────────────────────────────────────────────────────┘
```

## 表示項目（テーブル）

| フィールド名 | 型 | 説明 | 備考 |
|------------|-----|------|------|
| 日時 | date | 会計予定日（等幅フォント） | `r.scheduledDate` |
| 飼主名 | string | 飼い主氏名 | `r.ownerName` |
| ペット名 | string | ペット名 | `r.petName` |
| 請求金額 | decimal | 税込請求金額（等幅・太字・右寄せ） | `calculateTotal(r)` |
| 支払方法 | enum | 現金/クレジットカード/電子マネー | `r.payment.method` |
| ステータス | string | 会計状態（StatusBadge） | `r.status` |
| 操作 | - | 詳細ボタン | `RowActionButton` |

## フィルター

| 項目 | 選択肢 | 説明 |
|------|-------|------|
| ステータス | すべて / 会計待ち / 会計済 / キャンセル | `statusFilter` |
| フリーワード | - | 飼主名、ペット名で検索 |

## ソート

テーブルヘッダー（`SortableHeader`）で以下のキーによるソートが可能:
- `scheduledDate`（会計予定日）
- `ownerName`（飼主名）
- `petName`（ペット名）
- `status`（ステータス）

## 行アクション（`RowActionDropdown`）

| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/accounting/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → `deleteRecord`、構造化トースト |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Accounting` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | フリーワード検索（「飼主名、ペット名...」） |
| `ToggleGroup` / `ToggleGroupItem` | UI | 保険フィルター |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `SortableHeader` | `[S]` | ソート可能ヘッダー |
| `PrimaryButton` | `[S]` | 「新規会計登録」ボタン（Plusアイコン） |
| `StatusBadge` | `[S]` | ステータスバッジ（`getAccountingStatusColor` / `getAccountingStatusLabel`） |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `Pagination` | `[S]` | ページネーション（20件/ページ） |

## 使用フック

| フック | 説明 |
|---|---|
| `useAccountingRecords` | 検索・保険フィルタ・削除ロジック |
| `useTableSort` | テーブルソート（`resetKey`なし、初期ソートなし） |
| `usePagination` | ページネーション（`resetKey`: `searchTerm + insuranceFilter`） |

## データ型

`Accounting`, `AccountingStatus`, `PaymentMethod`, `InsuranceFilter`, `DataTableColumn`

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 新規会計登録 | 「新規会計登録」ボタン | ペット選択画面へ | `/accounting/select-pet` |
| 詳細/精算 | 行クリック or 操作「編集」 | 会計精算画面へ | `/accounting/:id` |
| 削除 | 操作「削除」 | ConfirmDialog → 削除 | 同画面 |
| 保険フィルター | ToggleGroup | 保険フィルター変更 | 同画面 |
| 検索 | テキスト入力 | 一覧フィルタリング | 同画面 |
| ソート | ヘッダークリック | 昇順/降順切替 | 同画面 |
| カルテリンク | ペット名列の「カルテ連携」バッジ（クリック） | カルテ詳細へ | `/medical-records/:id` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| サイドバー | `/accounting` | 常時 |
| 「新規会計登録」ボタン | `/accounting/select-pet` | ボタンクリック |
| ペット選択完了 | `/accounting/new?petId=xxx` | ペット選択後 |
| 行クリック / 編集 | `/accounting/:id` | クリック |
| カルテ「会計へ進む」 | `/accounting/new?medicalRecordId=xxx` | カルテタブ「会計(医師確認)」から |
| 入院「会計へ進む」 | `/accounting/new?petId=xxx` | 入院フォーム/詳細から（state経由） |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/accountings` | 会計一覧取得 | 実装済 |
| DELETE | `/api/v1/accountings/:id` | 会計データ削除 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/accounting/routes/AccountingList.tsx`）
- バックエンドAPI: 実装済（`handler/accounting_handler.go`）

