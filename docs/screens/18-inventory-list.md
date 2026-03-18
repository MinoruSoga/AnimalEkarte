# 在庫管理 仕様書

## 概要

- **画面の目的**: 在庫品目の一覧表示・検索・フィルタリング、および登録・編集・削除
- **URLパターン**:
  - 一覧: `/inventory`
  - 新規登録: `/inventory/new`
  - 編集: `/inventory/:id`
- **アクセス権限**: 認証済ユーザー全員

---

## 12.1 在庫一覧（`/inventory`）

### 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ 在庫管理 [Package]   [更新ボタン]  [在庫登録ボタン]    │
├──────────────────────────────────────────────────────┤
│ 🔍 [品名、保管場所、仕入先で検索...]  件数             │
│    [カテゴリー Select]  [状態 Select]                  │
├──────────────────────────────────────────────────────┤
│ テーブル（ソート可能ヘッダー付き）                     │
│ 品名 | カテゴリー | 在庫数 | 単位 | 発注点 |           │
│ 状態 | 保管場所 | 期限 | 仕入先 | 操作               │
├──────────────────────────────────────────────────────┤
│ ページネーション（20件/ページ）                        │
└──────────────────────────────────────────────────────┘
```

### テーブル列

| 列 | className / align | 表示内容 | 備考 |
|---|---|---|---|
| 品名 | - | `item.name` | ソート可能 |
| カテゴリー | `w-[100px]` | `CATEGORY_LABELS[item.category]` | ソート可能 |
| 在庫数 | `w-[100px]`, align:right | `item.quantity` + `item.unit` | ソート可能 |
| 最低在庫 | `w-[100px]`, align:right | `item.minStockLevel` + `item.unit` | - |
| 保管場所 | `w-[120px]` | `item.location` | ソート可能 |
| 有効期限 | `w-[120px]` | `item.expiryDate` | ソート可能 |
| ステータス | `w-[100px]` | `StatusBadge` (十分/残少/在庫切れ) | - |
| 操作 | `w-[80px]`, align:right | `RowActionButton` | - |

### フィルタ項目

| 項目 | 入力部品 | 選択肢 |
|---|---|---|
| キーワード | `NotionFilter` | 品名・保管場所・仕入先で検索 |
| カテゴリー | `Select` | 全カテゴリー / 医薬品 / 消耗品 / フード / その他 |
| 状態 | `Select` | 全ての状態 / 在庫あり / 残りわずか / 在庫切れ |

### カテゴリ一覧

| 値 | 表示名 |
|----|--------|
| `medicine` | 医薬品 |
| `consumable` | 消耗品 |
| `food` | フード |
| `other` | その他 |

### ステータスバッジ色

| ステータス | 値 | 表示名 | 色 |
|-----------|-----|--------|-----|
| 在庫あり | `sufficient` | 在庫あり | 緑（green） |
| 残りわずか | `low` | 残りわずか | 黄（amber） |
| 在庫切れ | `out_of_stock` | 在庫切れ | 赤（red） |

**ステータス自動判定ロジック:**
- `quantity <= 0`: 在庫切れ
- `quantity <= minStockLevel`: 残りわずか
- それ以外: 在庫あり

### 行アクション

| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/inventory/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → 削除、構造化トースト |

### コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `InventoryList` | `[R]` | メインページ（`Suspense` + `InventoryContent` の2層構成） |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `SortableHeader` | `[S]` | ソート可能ヘッダー（9列中7列） |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `Pagination` | `[S]` | ページネーション |
| `LoadingSkeleton` | `[S]` | ローディング（variant="table", rows=8, columns=7） |
| `useTableSort` | `[H]` | ソートロジック（数値 `comparator` 適用） |
| `usePagination` | `[H]` | ページネーション |

### 状態管理

| 状態 | 型 | 初期値 | 説明 |
|---|---|---|---|
| `searchTerm` | `string` | `""` | キーワード検索 |
| `categoryFilter` | `string` | `"all"` | カテゴリフィルタ |
| `statusFilter` | `string` | `"all"` | 状態フィルタ |
| `deletedIds` | `Set<string>` | 空 | 削除済みID（楽観的UI） |
| `deleteTarget` | `{ id: string; label: string } \| null` | `null` | 削除確認対象 |

### 特記事項

- データ取得に React `Suspense` を使用（`createResource` / `clearResource` でキャッシュ管理）
- 更新ボタン（RefreshCw アイコン）でキャッシュをクリアし再取得（`useTransition` でペンディング中はスピナー表示）
- カルテ保存時の在庫消費連動（`consumeStock`）は `useInventory` フック経由で実装

---

## 12.2 在庫登録/編集（`/inventory/new` / `/inventory/:id`）

### 画面構成

- ヘッダー: タイトル「在庫登録」/「在庫編集」（Package アイコン）+ 戻るボタン → `/inventory`
- フォームカード（`STYLE.formCard`、`max-w-2xl`）
- ローディング: `[S] LoadingSkeleton`（variant="form", rows=5、編集時）
- フォーム離脱保護（`[S] NavigationBlocker`）

### フォーム項目

#### 基本情報
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 品名 | `name` | `Input` | ✅ | |
| カテゴリ | `category`| `Select` | ✅ | 医薬品/消耗品/フード/その他 |
| 単位 | `unit` | `Input` | ✅ | 例: 錠, 本, 袋 |

#### 在庫情報
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| 現在庫数 | `quantity`| `Input(number)`| ✅ | |
| 最低在庫数 | `minStockLevel`| `Input(number)`| ✅ | |
| 保管場所 | `location`| `Input` | | |
| 有効期限 | `expiryDate`| `Input(date)`| | |

#### 仕入先情報
| フィールド | 項目ID | 入力部品 | 備考 |
|-----------|--------|----------|------|
| 仕入先 | `supplier`| `Input` | |
| 最終入荷日 | `lastRestocked`| `Input(date)` | |


### バリデーション

| フィールド | ルール | エラー表示 |
|-----------|--------|-----------|
| 品名 | 必須（空チェック） | `FormFieldError`（`role="alert"`, `aria-describedby`） + 構造化トースト |
| 単位 | 必須（空チェック） | `FormFieldError`（`role="alert"`, `aria-describedby`） + 構造化トースト |

### アクション

| ボタン | 位置 | 動作 | 備考 |
|-------|------|------|------|
| 削除 | 左（編集時のみ） | `ConfirmDialog` → `deleteInventoryItem` → 一覧へ | Trash2 アイコン、赤色 ghost |
| キャンセル | 右 | `/inventory` へ遷移 | outline |
| 保存 / 更新 | 右 | バリデーション → API → `markClean` → 遷移 | Save アイコン、送信中は Loader2 スピナー |

### コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `InventoryForm` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `PrimaryButton` | `[S]` | 保存ボタン |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `LoadingSkeleton` | `[S]` | ローディング表示 |
| `NotionDatePicker` | `[S]` | 日付ピッカー |
| `FormFieldError` | `[S]` | フィールドエラー表示 |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `useUnsavedChanges` | `[H]` | 未保存変更検知 |

---

## データ型

```typescript
interface InventoryItem {
  id: string;
  name: string;
  category: InventoryCategory;
  quantity: number;
  unit: string;
  minStockLevel: number;
  location?: string;
  expiryDate?: string;
  supplier?: string;
  status: InventoryStatus;
  lastRestocked?: string;
}

type InventoryCategory = "medicine" | "consumable" | "food" | "other";
type InventoryStatus = "sufficient" | "low" | "out_of_stock";
```

## 機能詳細

### 1. 在庫アラート発火ロジック
- **動的判定**: `InventoryStatus` は、各品目の「現在庫数」と「最低在庫数（発注点）」をリアルタイムで比較して決定される。
  - **十分**: 現在庫 > 最低在庫
  - **残少**: 0 < 現在庫 ≦ 最低在庫
  - **在庫切れ**: 現在庫 = 0
- **サマリー表示**: 一覧上部のアラートサマリーにより、発注が必要な品目数を一目で把握できる。

### 2. 有効期限の管理
- **期限切れ警告**: `expiryDate` が本日以前の品目、または本日より30日以内に迫っている品目について、一覧上で日付を赤字またはオレンジ色で強調表示する。

### 3. 棚卸し支援
- **CSV出力**: `FileSpreadsheet` アイコンから、現在のフィルタ条件に一致する在庫リストを CSV 形式でダウンロードし、オフラインでの棚卸しに利用できる。

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/inventories` | 在庫一覧取得 | 実装済 |
| GET | `/api/v1/inventories/:id` | 在庫詳細取得 | 実装済 |
| POST | `/api/v1/inventories` | 在庫品目作成 | 実装済 |
| PATCH | `/api/v1/inventories/:id` | 在庫品目更新 | 実装済 |
| DELETE | `/api/v1/inventories/:id` | 在庫品目削除 | 実装済 |

## 実装状況

- フロントエンド: 実装済（`features/inventory/routes/InventoryList.tsx`）
- バックエンドAPI: 実装済（`handler/inventory_handler.go`）
