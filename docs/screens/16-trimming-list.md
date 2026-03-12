# トリミング一覧 仕様書

## 概要

- **画面の目的**: トリミング施術記録の検索・一覧管理
- **URLパターン**: `/trimming`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ トリミング管理 [Scissors]    [新規登録ボタン]           │
├──────────────────────────────────────────────────────┤
│ [開始日] 〜 [終了日]  [クリア]                        │
│ 🔍 [飼主名、ペット名...]     件数表示                  │
├──────────────────────────────────────────────────────┤
│ テーブル（ソート可能ヘッダー付き）                     │
│ 診療日 | 飼主名 | ペット名 | 種 | 体重 |               │
│ スタイル希望 | 担当 | ステータス | 操作               │
├──────────────────────────────────────────────────────┤
│ ページネーション（20件/ページ）                        │
└──────────────────────────────────────────────────────┘
```

## 表示項目

| 列 | className | 表示内容 | 備考 |
|---|---|---|---|
| 診療日 | `w-[120px]` | `record.date`（等幅フォント） | ソート可能、初期降順 |
| 飼主名 | - | `record.ownerName` | ソート可能 |
| ペット名 | - | `record.petName`（上段）+ `record.petNumber`（下段、薄文字） | ソート可能 |
| 種 | `w-[80px]` | `record.species` | ソート可能 |
| 体重 | `w-[80px]` | `record.weight` | - |
| スタイル希望 | `max-w-[200px]` truncate | `record.styleRequest` | - |
| 担当 | `w-[100px]` | `record.staff`（無効スタッフ時は赤文字 + AlertTriangle） | ソート可能 |
| ステータス | `w-[100px]` | `StatusBadge`（`getTrimmingStatusColor`） | ソート可能 |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（編集 / 削除） | - |

## フィルタ項目

| 項目 | 入力部品 | 備考 |
|---|---|---|
| 開始日 | `NotionDatePicker` | `lg:w-[160px]` |
| 終了日 | `NotionDatePicker` | `lg:w-[160px]`、「〜」で接続 |
| キーワード | `SearchFilterBar` | 飼主名・ペット名で検索 |
| クリア | `Button`（outline） | 全フィルタリセット |

## ステータスバッジ色

| ステータス | 色 |
|-----------|-----|
| 予約 | 青（blue） |
| 進行中 | 黄（amber） |
| 完了 | 緑（green） |

## 行アクション

| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/trimming/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → `deleteRecord`、構造化トースト |

## コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `TrimmingList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `NotionDatePicker` | `[S]` | 日付範囲ピッカー（×2） |
| `SortableHeader` | `[S]` | ソート可能ヘッダー |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `Pagination` | `[S]` | ページネーション |
| `useTrimmingRecords` | `[H]` | 検索・フィルタ・削除ロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `useTableSort` | `[H]` | ソートロジック |
| `usePagination` | `[H]` | ページネーション（resetKey でリセット） |

## 状態管理

| 状態 | 型 | 初期値 | 説明 |
|---|---|---|---|
| `searchDate` | `{ from: string; to: string }` | `{ from: "", to: "" }` | 日付範囲フィルタ |
| `searchKeyword` | `string` | `""` | キーワード検索 |
| `deleteTarget` | `{ id: string; label: string } \| null` | `null` | 削除確認対象 |
| sortKey | `SortKey` | `"date"` | ソート対象列 |
| sortDirection | - | `descending` | ソート方向 |

## データ型

```typescript
interface TrimmingRecord {
  id: string;
  date: string;
  ownerName: string;
  petName: string;
  petNumber: string;
  species: string;
  weight: string;
  styleRequest: string;
  staff: string;
  status: "予約" | "進行中" | "完了";
}

type SortKey = "date" | "ownerName" | "petName" | "species" | "staff" | "status";
```

## ユーザー操作

- テキスト検索（飼主名・ペット名）
- 日付範囲フィルタ（開始日・終了日）
- 列ヘッダークリックでソート（3状態サイクル: 昇順→降順→なし）
- 新規登録ボタン → ペット選択 → フォーム画面
- 行クリック or 「編集」 → トリミングフォーム（編集）
- 「削除」 → 確認ダイアログ → 削除処理・構造化トースト

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/trimmings` | トリミング一覧取得 | 未実装 |
| DELETE | `/api/v1/trimmings/:id` | トリミング削除 | 未実装 |

## 実装状況

- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
