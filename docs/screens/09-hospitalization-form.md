# 入院登録/編集 仕様書

## 概要
- **画面の目的**: 入院・ペットホテルの新規登録および既存データの編集。治療プランの管理と会計への引き継ぎ。
- **URLパターン**:
  - 新規登録: `/hospitalization/new`（`?petId=xxx` クエリパラメータ使用可）
  - 編集: `/hospitalization/:id/edit`
- **コンポーネント**: `[R] HospitalizationForm`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ ← 戻る  入院登録/編集  [削除(編集時)] [デイリーカルテ(編集時)] [登録/更新] │
├──────────────────────────────────────────────────────┤
│ ■ PatientInfoCard                                    │
│   担当医/スタッフ | 入院タイプ（入院/ホテル）         │
│   ペット名/飼主名/種別等                              │
├──────────────────────────────────────────────────────┤
│ ■ HospitalizationBasicInfo（基本情報）               │
│   入院タイプ(RadioGroup) | ケージ・個室(Select)      │
│   期間: 開始日(NotionDatePicker) ～ 終了日           │
│   メモ(Textarea)                                     │
├──────────────────────────────────────────────────────┤
│ ■ HospitalizationTreatmentTable（治療プラン）         │
│   種別 | 治療内容 | 保険 | 単価 | 数量 | 割引 | 小計 │
│   [マスタ検索] [+ 行を追加]                          │
├──────────────────────────────────────────────────────┤
│ ■ HospitalizationNoteCard × 2（2カラム）             │
│   飼主からのリクエスト | スタッフへの連絡事項         │
├──────────────────────────────────────────────────────┤
│ ■ HospitalizationCostSummary（コスト集計）           │
│   保険ON/OFF | 負担割合 | 割引率/値引額 | 合計       │
├──────────────────────────────────────────────────────┤
│ ■ [会計へ進む]（編集時かつ治療プランあり）            │
└──────────────────────────────────────────────────────┘
```

## PatientInfoCard 表示項目

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 担当医/スタッフ | クリック → `MasterSelectModal`（staff マスタ、active のみ） | 未選択時はバリデーションエラー表示（`FormFieldError`） |
| 入院タイプ | クリックトグル（入院 ↔ ホテル） | `serviceTypeLabel="入院タイプ"` |

## HospitalizationBasicInfo フォーム項目

| フィールド | 入力部品 | 必須 | 備考 |
|---|---|---|---|
| 入院タイプ | `RadioGroup`（入院 / ホテル） | ✅ | `HOSPITALIZATION_TYPE_VALUES` |
| 期間（開始日） | `NotionDatePicker` | ✅ | Calendar アイコン付き |
| 期間（終了日） | `NotionDatePicker` | | 変更時、既存治療プランの数量自動調整提案あり |
| ケージ・個室 | `Select`（cage マスタ連動） | ✅ | `MasterLink` 付き |
| メモ | `Textarea` | | |

### 入院日数変更ダイアログ（`StayDaysAdjustDialog`）
終了日変更時に、変更前の日数と数量が一致する治療プランが存在する場合に表示。
新しい日数への一括更新を提案する。

## HospitalizationTreatmentTable（治療プラン）

診察/治療プランタブのTreatmentTableと同パターン。

| フィールド | 説明 |
|---|---|
| 種別 | マスタ選択時に自動設定（読み取り専用テキスト） |
| 治療内容 | 項目名（マスタ選択または手動入力） |
| 保険 | 保険適用チェックボックス |
| 単価(税込) | 数値入力 |
| 数量 | 数値入力（入院日数連動で自動設定可） |
| 割引(%) | 数値入力 |
| 値引(¥) | 数値入力 |
| 小計 | 自動計算（読み取り専用） |

マスタ検索（`TreatmentSearchDialog`）から処置・検査・薬を検索して自動入力可能。

## HospitalizationNoteCard（メモカード）

| カード | アイコン | プレースホルダー |
|---|---|---|
| 飼主からのリクエスト | MessageSquare | 「リクエストを入力...」 |
| スタッフへの連絡事項 | AlertCircle | 「連絡事項を入力...」 |

## HospitalizationCostSummary（コスト集計）

| フィールド | 説明 |
|---|---|
| 保険ON/OFF | `Switch` トグル |
| 負担割合 | `Select`（50%/70%/90%/100%） |
| 保険負担額 | 自動計算（読み取り専用、緑背景カード） |
| 全体割引率(%) | `Input`（0-100範囲） |
| 全体値引額(¥) | `Input` |
| 税込合計 | 自動計算（読み取り専用） |
| 消費税(内税10%) | 自動計算（読み取り専用） |

## 「会計へ進む」ボタン（編集時のみ）

- 表示条件: 編集モード（`isEdit === true`）かつ治療プランが1件以上
- 治療プランを`AccountingItem[]`に変換し、`location.state`経由で会計新規作成画面へ遷移
- 遷移先: `/accounting/new?petId=xxx`（state に `hospitalizationId`, `accountingItems`, `globalDiscountRate` 等を含む）

## ヘッダーアクションボタン

| ボタン | 条件 | 動作 |
|--------|------|------|
| デイリーカルテ | 編集時のみ（FileTextアイコン） | `/hospitalization/:id`へ遷移 |
| 削除 | 編集時のみ（Trash2アイコン、赤字） | `ConfirmDialog`表示 → 削除後一覧へ |
| 登録/更新 | 常時（`PrimaryButton`） | バリデーション → 保存 |

## バリデーション

| フィールド | ルール |
|---|---|
| 担当医 | 必須（未選択時: FormFieldError + スタッフボタンにエラー状態表示 + トースト警告） |
| 入院タイプ | 必須 |
| ケージ | 必須 |
| 入院開始日 | 必須 |

## フォーム離脱保護

`NavigationBlocker`（`[S]`）により、未保存変更がある場合に画面離脱時に警告を表示。

## 使用フック

| フック | 説明 |
|---|---|
| `useHospitalizationForm` | フォーム状態管理（186行） |
| `useTreatmentPlans` | 治療プラン管理（112行） |
| `useUnsavedChanges` | 未保存変更の追跡 |

## データ型

`HospitalizationFormData`, `TreatmentPlan`, `CreateHospitalizationDTO`, `UpdateHospitalizationDTO`

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| ペット選択 | `/hospitalization/new?petId=xxx` | ペット選択後 |
| 入院詳細「編集」 | `/hospitalization/:id/edit` | ボタンクリック |
| 保存成功（新規） | `/hospitalization` | 登録完了後 |
| 保存成功（編集） | `/hospitalization` | 更新完了後 |
| 「デイリーカルテ」ボタン | `/hospitalization/:id` | 編集時のみ |
| 「会計へ進む」ボタン | `/accounting/new?petId=xxx` | 編集時かつプランあり |
| ペット未選択（新規） | `/hospitalization/select-pet` | 自動リダイレクト |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| POST | `/api/v1/hospitalizations` | 入院作成 | 未実装 |
| PUT | `/api/v1/hospitalizations/:id` | 入院更新 | 未実装 |
| GET | `/api/v1/hospitalizations/:id` | 入院詳細取得（編集時） | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（LocalStorageによるデータ永続化）
- バックエンドAPI: 未実装
