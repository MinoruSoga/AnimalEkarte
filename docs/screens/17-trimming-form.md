# トリミング登録/編集 仕様書

## 概要

- **画面の目的**: トリミング施術記録の新規登録・編集
- **URLパターン**:
  - ペット選択: `/trimming/select-pet`（`[R] TrimmingPetSelection`）
  - 新規: `/trimming/new?petId=xxx`
  - 編集: `/trimming/:id`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

```
┌──────────────────────────────────────────────────────────┐
│ トリミング登録/編集 [Scissors]  [削除(編集時)]  [保存]     │
├──────────────────────────────────────────────────────────┤
│ PatientInfoCard（患者情報・担当スタッフ・サービス区分）    │
├──────────────────────────────────────────────────────────┤
│ 3カラムレイアウト（lg:grid-cols-3 gap-3）                 │
│                                                          │
│ [左カラム]         [中カラム]         [右カラム]           │
│  コース選択         BW / BT           トリミング履歴        │
│  スタイルの希望     USED SHAMPOO      HistoryFilterPanel   │
│  メモ              USED RIBBON        履歴カード一覧        │
│  オプション         TREATMENT                              │
│  希望スタイル画像   備考                                   │
│                   完成画像                                │
└──────────────────────────────────────────────────────────┘
```

## ペット選択（`/trimming/select-pet`）

- `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネントを使用
- ペット未選択かつ新規モードの場合はペット選択画面へリダイレクト

## 左カラム フォーム項目

| フィールド | 入力部品 | 必須 | 備考 |
|-----------|---------|------|------|
| コース選択 | `MasterSelectTrigger` → `MasterSelectModal`（trimming_course マスタ連動） | 必須 | `MasterLink` 付き、選択時に `charge`（料金）自動反映 |
| スタイルの希望 | `Textarea` | - | `min-h-[80px]` |
| メモ | `Textarea` | - | `min-h-[80px]` |
| オプション | `NotionCheckbox` グリッド（2cols） | - | trimming_option マスタ連動、`MasterLink` 付き、各項目に `+¥{price}` 表示 |
| 希望スタイル画像 | `file input` + プレビュー | - | ドラッグ&ドロップUI、Upload アイコン、×ボタンで削除、`h-[180px]` |

## 中カラム フォーム項目

| フィールド | 入力部品 | 備考 |
|-----------|---------|------|
| BW（体重） | `Input` + `radio`（Kg / g） | `BODY_WEIGHT_UNIT_VALUES`、2カラムグリッド |
| BT（体温） | `Input` | placeholder: 体温 |
| USED SHAMPOO | `Input` | placeholder: 使用したシャンプーを入力... |
| USED RIBBON | `Input` | placeholder: 使用したリボンを入力... |
| TREATMENT | `Input` | placeholder: 処置内容を入力... |
| 備考 | `Input` | placeholder: 備考を入力... |
| 完成画像 | `file input` + プレビュー | Upload アイコン、×ボタンで削除、`h-[180px]` |

## 右カラム（トリミング履歴）

- タイトル: 「トリミング履歴」
- `[S] HistoryFilterPanel`: 日付範囲、キーワード検索（「コース名・メモで検索...」）、ソート順（昇順/降順）、クリアボタン
- 履歴カード表示項目:
  - 診療日、コース名バッジ
  - 作成者/更新者・日時
  - スタイルの希望、メモ、画像セクション（`#`記法）
- 空状態: 「該当するトリミング履歴がありません」

## 担当スタッフ選択（PatientInfoCard 経由）

- `PatientInfoCard` の担当スタッフをクリックで `MasterSelectModal`（staff マスタ連動、active のみ）を表示
- タイトル: 「担当スタッフを選択」

## バリデーション

| フィールド | ルール | エラー表示 |
|-----------|--------|-----------|
| 担当スタッフ | 必須 | `toast.warning` で保存ブロック + `FormFieldError`（`role="alert"`、`aria-describedby` 接続） |
| コース | 必須 | `FormFieldError`（`role="alert"`、`aria-describedby` 接続） |

## ヘッダーアクション

| ボタン | 表示条件 | 動作 |
|-------|---------|------|
| 削除 | 編集時のみ | `ConfirmDialog` → 一覧へ遷移 |
| 保存 | 常時 | `handleSave()` → バリデーション → `markClean()` |

## コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `TrimmingForm` | `[R]` | メインページ |
| `TrimmingPetSelection` | `[R]` | ペット選択ページ |
| `PatientInfoCard` | `[S]` | 患者情報カード（担当スタッフ・サービス区分表示） |
| `MasterSelectModal` | `[S][M]` | コース・スタッフ選択モーダル |
| `MasterSelectTrigger` | `[S]` | コース選択トリガーボタン |
| `HistoryFilterPanel` | `[S]` | 履歴フィルタパネル |
| `MasterLink` | `[S]` | マスタ設定へのリンク |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `FormFieldError` | `[S]` | フィールドエラー表示（`role="alert"`） |
| `NotionCheckbox` | `[S]` | チェックボックス |
| `useTrimmingForm` | `[H]` | フォーム状態管理 |
| `useUnsavedChanges` | `[H]` | 未保存変更検知 |
| `useMasterItems` | `[H]` | マスタデータ取得（trimming_course / trimming_option / staff） |

## 状態管理

| 状態 | 型 | 説明 |
|---|---|---|
| `formData` | `TrimmingFormData` | フォーム全体の状態 |
| `styleImagePreview` | `string \| null` | 希望スタイル画像プレビューURL |
| `completedImagePreview` | `string \| null` | 完成画像プレビューURL |
| `courseModalOpen` | `boolean` | コース選択モーダル表示状態 |
| `staffModalOpen` | `boolean` | スタッフ選択モーダル表示状態 |
| `deleteConfirmOpen` | `boolean` | 削除確認ダイアログ表示状態 |
| `isDirty` | `boolean` | 未保存変更フラグ（`useUnsavedChanges`） |
| 履歴フィルタ各種 | `string` / `SortOrder` | historyFilterStartDate, End, searchTerm, sortOrder |

## データ型

```typescript
interface TrimmingFormData {
  styleRequest: string;
  memo: string;
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  treatment: string;
  remarks: string;
  charge: string;
  courseId: string;
  optionIds: string[];
  staffName: string;
  styleImage: File | null;
  completedImage: File | null;
}

interface TrimmingHistoryItem {
  id: string;
  date: string;
  createdBy: string;
  createdAt: string;
  updatedBy: string;
  updatedAt: string;
  course: string;
  styleRequest: string;
  memo: string;
}

type BodyWeightUnit = "Kg" | "g";
type SortOrder = "asc" | "desc";
```

## アクセシビリティ

- 担当医エラー: `PatientInfoCard.staffAriaDescribedBy` → `FormFieldError`（`role="alert"`）と `aria-describedby` 接続
- コース選択エラー: `MasterSelectTrigger.ariaDescribedBy` → `FormFieldError` と `aria-describedby` 接続
- `MasterSelectTrigger`: 選択済み・未選択状態ともに `<button>` 要素（キーボード操作対応）

## ユーザー操作

- コース選択（マスタモーダル）→ 料金自動反映
- オプション複数選択（チェックボックス）→ 各 `+¥{price}` 表示
- 体重単位切替（Kg/g ラジオ）
- スタイル画像・完成画像のアップロード/削除
- 担当スタッフ選択（PatientInfoCard クリック）
- トリミング履歴の検索・日付範囲フィルタ・ソート
- 保存（バリデーション → トースト → 一覧へ遷移）
- 削除（確認ダイアログ → 一覧へ遷移、編集時のみ）
- 未保存離脱時の保護ダイアログ（`NavigationBlocker`）

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/trimmings/:id` | トリミング詳細取得 | 未実装 |
| POST | `/api/v1/trimmings` | トリミング作成 | 未実装 |
| PUT | `/api/v1/trimmings/:id` | トリミング更新 | 未実装 |
| DELETE | `/api/v1/trimmings/:id` | トリミング削除 | 未実装 |
| GET | `/api/v1/master-items?category=trimming_course` | コースマスタ取得 | 実装済 |
| GET | `/api/v1/master-items?category=trimming_option` | オプションマスタ取得 | 実装済 |
| GET | `/api/v1/master-items?category=staff` | スタッフマスタ取得 | 実装済 |

## 実装状況

- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
