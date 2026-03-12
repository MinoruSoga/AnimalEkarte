# カルテ一覧 仕様書

## 概要
- **画面の目的**: 全ペットの診療記録（電子カルテ）を一覧表示・検索・管理する
- **URLパターン**: `/medical-records`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────┐
│ [FileTextアイコン] カルテ管理  [新規カルテ作成ボタン] │
├──────────────────────────────────────────────────────┤
│ 🔍 [飼主名、ペット名、カルテNo、主訴で検索...]  [N件] │
├──────────────────────────────────────────────────────┤
│ テーブル                                             │
│ 診療日↑↓ | 飼主名↑↓ | ペット名↑↓ | 種↑↓ |         │
│ 訴 | 関連 | 担当医↑↓ | ステータス↑↓ | 操作        │
├──────────────────────────────────────────────────────┤
│ [ページネーション: 20件/ページ]                       │
└──────────────────────────────────────────────────────┘
```

## 表示項目

| フィールド名 | 型 | 説明 | ソート | 備考 |
|------------|-----|------|--------|------|
| 診療日 | date | 診察日時（等幅フォント） | ○ | `medical_records.visit_date`（`date`） |
| 飼主名 | string | 飼い主氏名 | ○ | `owners.name`（`ownerName`） |
| ペット名 | string | ペット名 | ○ | `pets.name`（`petName`） |
| 種 | string | 犬/猫等 | ○ | `pets.species` |
| 訴 | string | 主訴（省略表示 truncate、max-w-[200px]、title属性でフルテキスト） | - | `medical_records.chief_complaint`（`chiefComplaint`） |
| 関連 | link | 関連会計へのリンク（`CreditCard` アイコン付きボタン）。未リンク時は「—」 | - | `billings` テーブルへの紐付け |
| 担当医 | string | 担当獣医師名（無効スタッフ時 `AlertTriangle` アイコン + 赤文字警告） | ○ | `staffs.name` |
| ステータス | enum | StatusBadge 表示 | ○ | `medical_records.status` |
| 操作 | - | 編集・削除ドロップダウン | - | `RowActionDropdown` |

## ステータスバッジ色定義

| ステータス | 色 |
|-----------|-----|
| 作成中 | 黄（amber）|
| 確定済 | 緑（green）|

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MedicalRecords` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページコンテナ（FileText アイコン付きタイトル） |
| `SearchFilterBar` | `[S]` | フリーワード検索（飼主名・ペット名・カルテNo・主訴） |
| `DataTable` | `[S]` | テーブルコンテナ |
| `DataTableRow` | `[S]` | クリックでカルテ詳細へ遷移 |
| `SortableHeader` | `[S]` | ソート可能なカラムヘッダー |
| `StatusBadge` | `[S]` | ステータスバッジ（`getMedicalRecordStatusColor` でカラー決定） |
| `RowActionDropdown` | `[S]` | 行ごとの編集・削除ドロップダウン |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `Pagination` | `[S]` | ページネーション（20件/ページ） |
| `PrimaryButton` | `[S]` | 「新規カルテ作成」ボタン |
| `useMedicalRecords` | `[H]` | カルテ一覧データ取得・削除 |
| `useStaffValidation` | `[H]` | 担当医の有効性チェック（マスタで非アクティブ時に警告表示） |
| `useTableSort` | `[H]` | テーブルソート管理 |
| `usePagination` | `[H]` | ページネーション管理 |

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 新規カルテ作成 | 「新規カルテ作成」ボタン | ペット選択画面へ遷移 | `/medical-records/select-pet` |
| カルテ詳細/編集 | 行クリックまたは操作ドロップダウン「編集」 | カルテ詳細/編集フォームへ | `/medical-records/:id` |
| 削除 | 操作ドロップダウン「削除」 | `ConfirmDialog` 表示後に削除 | 同画面 |
| 検索 | テキスト入力 | 一覧をリアルタイムフィルタリング | 同画面 |
| ソート | カラムヘッダークリック | 昇順/降順ソート切替 | 同画面 |
| 関連会計を開く | 「会計」リンクボタンクリック | 会計詳細画面へ遷移（`e.stopPropagation()`） | `/accounting/:accountingId` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| サイドバー | `/medical-records` | 常時 |
| 「新規カルテ作成」ボタン | `/medical-records/select-pet` | ボタンクリック（`state: { from: "/medical-records" }`） |
| 行クリック | `/medical-records/:id` | 行クリック |
| ペット選択完了 | `/medical-records/new?petId=xxx` | ペット選択後 |
| 「会計」ボタン | `/accounting/:accountingId` | 会計リンクボタンクリック |

## バリデーション
- 検索はフリーワード（部分一致）、複数フィールドをOR検索
- 削除時: 「関連する治療・検査データも削除されます。この操作は元に戻せません。」の警告文を表示

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `searchTerm` | `string` | 検索キーワード（`usePagination` の `resetKey` にも利用） |
| `deleteTarget` | `{ id: string; label: string } \| null` | 削除対象（`ConfirmDialog` 用） |
| `sortedData` | `MedicalRecord[]` | ソート済みデータ（`useTableSort`） |
| `paginatedData` | `MedicalRecord[]` | ページネーション済みデータ（`usePagination`） |

## ソートキー
`date` / `ownerName` / `petName` / `species` / `doctor` / `status`

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/medical-records` | カルテ一覧取得 | 未実装 |
| DELETE | `/api/v1/medical-records/:id` | カルテ削除 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装

## 備考
- 旧仕様では「カルテNo」カラムがあったが、現在は列から除外（`SCREENS.md` 記載: 内部ID はUI上非表示）
- 「関連」カラムが追加されており、会計データとの紐付けをインラインで確認できる
- 担当医が無効スタッフ（マスタで非アクティブ）の場合、`useStaffValidation` で判定し `AlertTriangle` アイコン + 赤文字で警告表示
