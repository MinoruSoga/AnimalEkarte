# 入院管理一覧 仕様書

## 概要
- **画面の目的**: 入院・ホテル患者の一覧管理。ボードビューとリストビューを切り替えて使用する。
- **URLパターン**: `/hospitalization`
- **コンポーネント**: `[R] HospitalizationList`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

### ボードビュー（入院中タブのみ表示）
```
┌──────────────────────────────────────────────────────┐
│ 入院・ホテル管理  [新規入院登録ボタン]               │
├──────────────────────────────────────────────────────┤
│ タブ: [入院中 N] [予約 N] [退院済 N] [すべて N]      │
├──────────────────────────────────────────────────────┤
│ 🔍 [飼主名、ペット名、入院No...] [件数]  [ボード|リスト] │
├──────────────────────────────────────────────────────┤
│ ケージボード                                         │
│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐               │
│ │ケージ1│ │ケージ2│ │ケージ3│ │空き  │               │
│ │ペット │ │ペット │ │      │ │      │               │
│ │カード │ │カード │ │      │ │      │               │
│ └──────┘ └──────┘ └──────┘ └──────┘               │
└──────────────────────────────────────────────────────┘
```

### リストビュー（全タブ）
```
テーブル: 入院No | ペット名 | 飼主名 | 入院種別 |
          入院日 | 退院予定日 | ケージ | ステータス | 操作
ページネーション（20件/ページ）
```

## 表示項目

| フィールド名 | 型 | 説明 | 備考 |
|------------|-----|------|------|
| 入院No | string | 入院番号 | `hospitalizations.hospitalization_no` |
| ペット名 | string | ペット名 | `pets.name` |
| 種/品種 | string | 動物種と品種 | `pets.species`, `pets.breed` |
| 飼主名 | string | 飼い主氏名 | `owners.name` |
| 入院種別 | enum | 入院/ホテル | `hospitalizations.type` |
| 入院開始日 | date | 入院開始日 | `hospitalizations.start_date` |
| 退院予定日 | date | 退院予定日 | `hospitalizations.end_date` |
| ケージ | string | 割り当てケージ名 | `cages.name` |
| ステータス | enum | 入院中/退院済/予約 | `hospitalizations.status` |

## フィルタータブ

| タブ名 | フィルター条件 | カウント表示 |
|--------|--------------|------|
| 入院中 | status = `入院中` | 件数（等幅フォント） |
| 予約 | status = `予約` | 件数（等幅フォント） |
| 退院済 | status = `退院済` | 件数（等幅フォント） |
| すべて | フィルターなし | 件数（等幅フォント） |

## ビュー切替の制約

- ボードビュー（`HospitalizationBoard`）は「入院中」タブ選択時のみ表示可能
- 他タブ（予約/退院済/すべて）では常にリストビュー表示
- ビュー切替トグル（`ToggleGroup`）も入院中タブ選択時のみ表示

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|--------------|------|------|
| `HospitalizationList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページコンテナ |
| `Tabs` / `TabsList` / `TabsTrigger` | `[S]` | ステータスフィルタータブ |
| `SearchFilterBar` | `[S]` | 飼主名・ペット名・入院Noの検索 |
| `ToggleGroup` / `ToggleGroupItem` | UI | ボードビュー/リストビュー切替（LayoutGrid/Listアイコン） |
| `HospitalizationBoard` | `[C]` | ケージボードビュー（react-dndによるD&D） |
| `CageDragPreview` | `[C]` | `useDragLayer`によるカスタムドラッグプレビューオーバーレイ |
| `HospitalizationListView` | `[C]` | リストビュー（テーブル形式） |
| `Pagination` | `[S]` | ページネーション（リストビュー時、20件/ページ） |
| `PrimaryButton` | `[S]` | 「新規入院登録」ボタン（Plusアイコン） |
| `ConfirmDialog` | `[S][M]` | 退院確認ダイアログ |

## 使用フック

| フック | 説明 |
|--------|------|
| `useHospitalizationList` | 検索・フィルタ・ビュー切替・ケージ移動・削除・退院ロジック |
| `usePagination` | ページネーション（resetKey: searchTerm + statusFilter） |

## データ型

`Hospitalization`, `HospitalizationFilterStatus`, `HospitalizationViewMode`

### ステータス値

| 表示名 | 値 | 定数 |
|--------|-----|------|
| 入院中 | `active` | `HOSPITALIZATION_FILTER_STATUS.ACTIVE` |
| 予約 | `reserved` | `HOSPITALIZATION_FILTER_STATUS.RESERVED` |
| 退院済 | `discharged` | `HOSPITALIZATION_FILTER_STATUS.DISCHARGED` |
| すべて | `all` | `HOSPITALIZATION_FILTER_STATUS.ALL` |

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 新規入院登録 | 「新規入院登録」ボタン | ペット選択画面へ | `/hospitalization/select-pet` |
| 詳細表示（詳細カルテ） | ケージカードorリスト行クリック | 入院詳細画面へ | `/hospitalization/:id` |
| ケージ移動 | ボードビューD&D | ケージ変更 | 同画面（状態更新） |
| 退院処理（一覧から） | ボードカードの退院ボタン | ConfirmDialog → 退院ステータスへ変更 | 同画面 |
| 削除 | リスト行の操作 | 削除確認後削除 | 同画面 |
| タブ切替 | タブクリック | ステータスフィルター変更 | 同画面 |
| ビュー切替 | ToggleGroup | ボード/リスト切替（入院中のみ） | 同画面 |
| メモ追加 | ボードカードのメモアクション | 詳細画面へ遷移してトーストを表示 | `/hospitalization/:id` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| サイドバー | `/hospitalization` | 常時 |
| 新規入院登録ボタン | `/hospitalization/select-pet` | ボタンクリック |
| ペット選択完了 | `/hospitalization/new?petId=xxx` | ペット選択後 |
| ケージカード/行クリック | `/hospitalization/:id` | クリック |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/hospitalizations` | 入院一覧取得 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（LocalStorageによるデータ永続化）
- バックエンドAPI: 未実装
