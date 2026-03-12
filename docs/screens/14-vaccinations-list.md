# 予防接種一覧 仕様書

## 概要

- **画面の目的**: 予防接種記録の一覧表示・検索・カルテ遷移
- **URLパターン**: `/vaccinations`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ 予防接種管理 [Syringe]    [新規登録ボタン]              │
├──────────────────────────────────────────────────────┤
│ 🔍 [飼主名、ペット名、予防接種名...]     件数表示       │
├──────────────────────────────────────────────────────┤
│ テーブル（ソート可能ヘッダー付き）                     │
│ 実施日 | 飼主名 | ペット名 | 予防接種名 |               │
│ 担当医 | 次回予定 | 操作                               │
├──────────────────────────────────────────────────────┤
│ ページネーション（20件/ページ）                        │
└──────────────────────────────────────────────────────┘
```

## 表示項目

| 列 | className | 表示内容 | 備考 |
|---|---|---|---|
| 実施日 | `w-[120px]` | `r.date`（等幅フォント） | ソート可能 |
| 飼主名 | - | `r.ownerName` | ソート可能 |
| ペット名 | - | `r.petName` | ソート可能 |
| 予防接種名 | - | `r.vaccineName`（太字） | ソート可能 |
| 担当医 | `w-[100px]` | `r.doctor`（無効スタッフ時は赤文字 + AlertTriangle、未設定時「-」） | ソート可能 |
| 次回予定 | `w-[140px]` | `r.nextDate`（等幅フォント） | ソート可能 |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（「カルテを開く」のみ） | - |

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 新規登録 | 「新規登録」ボタン（Plus アイコン） | ペット選択画面へ | `/medical-records/select-pet`（state: `{ activeTab: "予防接種", from: "/vaccinations" }`） |
| カルテを開く | 行クリック or RowActionDropdown「カルテを開く」 | カルテの予防接種タブへ遷移 | `/medical-records/{medicalRecordId}`（state: `{ activeTab: "予防接種", from: "/vaccinations" }`） |
| 列ソート | ヘッダークリック | 3状態サイクル（昇順→降順→なし）、初期値は実施日降順 | 同画面 |

## 重要仕様

- **予防接種の新規登録はカルテ内の予防接種タブから行う**。一覧画面は参照とカルテ遷移専用
- 行クリックでも「カルテを開く」と同じ遷移先（カルテの予防接種タブ）へ移動する
- `useStaffValidation` フックでスタッフ有効性チェックを行い、無効スタッフは赤文字 + AlertTriangle アイコンで警告表示

## コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `VaccinationList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `SortableHeader` | `[S]` | ソート可能ヘッダー（全列対応） |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション |
| `useVaccinations` | `[H]` | 検索・フィルタロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `useTableSort` | `[H]` | ソートロジック |
| `usePagination` | `[H]` | ページネーション（pageSize: 20） |

## 状態管理

| 状態 | 型 | 初期値 | 説明 |
|---|---|---|---|
| `searchTerm` | `string` | `""` | 検索キーワード |
| sortKey | `SortKey` | `"date"` | ソート対象列 |
| sortDirection | `ascending` \| `descending` | `descending` | ソート方向 |

## データ型

```typescript
interface VaccinationRecord {
  id: string;
  medicalRecordId: string;  // カルテ遷移に使用
  date: string;             // 実施日（等幅フォント表示）
  ownerName: string;
  petName: string;
  vaccineName: string;
  doctor: string;
  nextDate: string;
}

type SortKey = "date" | "ownerName" | "petName" | "vaccineName" | "doctor" | "nextDate";
```

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/vaccinations` | ワクチン記録一覧取得 | 未実装 |

## 実装状況

- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
