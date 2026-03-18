# 検査一覧 仕様書

## 概要
- **画面の目的**: 検査オーダー・結果の一覧管理。検索・ソート・ページネーションが可能。
  検査の新規作成はカルテ内の検査タブから行う。一覧画面は参照とカルテへの遷移のみ。
- **URLパターン**: `/examinations`
- **コンポーネント**: `[R] Examinations`
- **アクセス権限**: 認証済ユーザー全員

## 重要な設計方針

**検査の新規登録はこの画面からは行わない。** カルテ入力画面の「検査」タブからのみ作成可能。
一覧画面の行クリックおよび「カルテを開く」アクションはカルテの検査タブへ遷移する。

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ 検査管理  [検査データ取込ボタン]  (TestTubeアイコン)   │
├──────────────────────────────────────────────────────┤
│ 🔍 [飼主名、ペット名、検査種別...] [件数]             │
├──────────────────────────────────────────────────────┤
│ テーブル（ソート可能ヘッダー）                       │
│ 日時 | 飼主名 | ペット名 | 検査種別 | 結果概要 |      │
│ 担当医 | ステータス | 操作                           │
├──────────────────────────────────────────────────────┤
│ ページネーション（20件/ページ）                      │
└──────────────────────────────────────────────────────┘
```

## 表示項目（テーブル）

| フィールド名 | 型 | 説明 | ソート | 備考 |
|------------|-----|------|--------|------|
| 日時 | date | 検査実施日時（等幅フォント） | ○ | `r.date` |
| 飼主名 | string | 飼い主氏名 | ○ | `r.ownerName` |
| ペット名 | string | ペット名 | ○ | `r.petName` |
| 検査種別 | string | 検査の種類 | ○ | `r.testType` |
| 結果概要 | string | 検査結果の概要（truncate 表示）| - | `r.resultSummary` |
| 担当医 | string | 依頼した獣医師名 | ○ | `r.doctor` |
| ステータス | enum | StatusBadge 表示（依頼中/検査中/完了） | ○ | `r.status` |
| 操作 | - | 詳細ボタン | - | `RowActionButton` |

## ステータスバッジ色

| ステータス | 色 |
|-----------|-----|
| 依頼中 | 青（blue） |
| 検査中 | 黄（amber） |
| 完了 | 緑（green） |

## ソート

テーブルヘッダー（`SortableHeader`）で以下のキーによるソートが可能（初期: `date` 降順）:
- `date`（日時）
- `ownerName`（飼主名）
- `petName`（ペット名）
- `testType`（検査種別）
- `doctor`（担当医）
- `status`（ステータス）

## 行アクション（`RowActionDropdown`）

| アクション | アイコン | 動作 |
|---|---|---|
| カルテを開く | FileText | `/medical-records/{medicalRecordId}` へ遷移（state: `{ activeTab: "検査", from: "/examinations" }`） |

行クリックでも同様にカルテの検査タブへ遷移する。

## ヘッダーボタン

| ボタン | アイコン | 動作 |
|--------|---------|------|
| 検査データ取込 | FileSpreadsheet | 未実装（`outline`バリアント） |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Examinations` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー（「飼主名、ペット名、検査種別...」） |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `SortableHeader` | `[S]` | ソート可能ヘッダー |
| `StatusBadge` | `[S]` | ステータスバッジ（`getExaminationStatusColor`） |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション（20件/ページ） |

## 使用フック

| フック | 説明 |
|---|---|
| `useExaminationRecords` | 検索・フィルタロジック |
| `useStaffValidation` | スタッフ有効性チェック（無効スタッフ警告表示） |
| `useTableSort` | テーブルソート（初期: `date` 降順） |
| `usePagination` | ページネーション（`resetKey`: `searchTerm`） |

## データ型

`ExaminationRecord`, `ExaminationRecordItem`, `DataTableColumn`

## ExaminationRecord 主要フィールド

| フィールド名 | 型 | DB列 |
|------------|-----|------|
| id | string | `exams.id` |
| date | string | `exams.examination_date` |
| ownerName | string | `owners.name` |
| petName | string | `pets.name` |
| testType | string | `exams.test_type` |
| resultSummary | string | `exams.result_summary` |
| doctor | string | `staffs.name` |
| status | enum | `exams.status` |
| medicalRecordId | string | `medical_records.id` |

## 機能詳細

### 1. 異常値のハイライト判定
- **自動判定**: 検査項目の測定値が、マスタで定義された「基準値」の範囲（上限・下限）を外れている場合、一覧の概要や詳細行において赤色（High）または青色（Low）で強調表示する。
- **サマリー生成**: `resultSummary` には、主要な項目や異常が見つかった項目が優先的にテキストとして抽出される。

### 2. カルテおよび会計との動的連携
- **ステータス同期**: カルテ側で検査が実施済みにされると、本一覧のステータスも自動的に「完了」へ更新される。
- **請求連動**: 完了した検査データは、会計画面の明細に自動的に計上される。

### 3. 種別によるUIの切り替え
- **表示切替**: 血液検査、レントゲン、エコーなど、検査種別（`testType`）に応じて表示されるアイコンや、詳細時のテーブルカラム構成が最適化される。

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| カルテを開く | 「カルテを開く」操作 or 行クリック | カルテの検査タブへ遷移 | `/medical-records/:id`（state: `{ activeTab: "検査" }`） |
| 検索 | テキスト入力 | 一覧フィルタリング | 同画面 |
| ソート | ヘッダークリック | 昇順/降順切替 | 同画面 |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| サイドバー | `/examinations` | 常時 |
| 行クリック / 「カルテを開く」 | `/medical-records/:medicalRecordId` | クリック（state: `{ activeTab: "検査", from: "/examinations" }`） |
| カルテ「検査」タブ（from登録） | `/examinations` | 戻るボタン（state.from利用） |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/examinations` | 検査一覧取得 | 実装済 |
| DELETE | `/api/v1/examinations/:id` | 検査削除 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/examinations/routes/Examinations.tsx`）
- バックエンドAPI: 実装済（`handler/examination_handler.go`）

## 備考

- 検査の新規登録・編集はカルテ画面（`/medical-records/:id`）の「検査」タブ（Tab 5: `MedicalRecordExamination`）から行う
- カルテ検査タブ: `MasterSelectModal`（examination マスタ連動）で検査種別を選択し、検査項目テーブル（項目名・測定値・単位・基準値）を管理
- 右カラム `ExaminationHistory` で検査結果履歴を参照可能
- 検査結果セクション: 日付別画像グループ（`ImageGalleryGroup`）でファイル管理
