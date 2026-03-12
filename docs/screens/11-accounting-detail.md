# 会計精算 仕様書

## 概要
- **画面の目的**: 診療費の計算、保険適用処理、支払い処理、領収書/診療明細書の発行
- **URLパターン**:
  - 新規: `/accounting/new`（`?petId=xxx` or `?medicalRecordId=xxx`、`location.state`からの引き継ぎあり）
  - 既存: `/accounting/:id`
- **コンポーネント**: `[R] AccountingDetail`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト（2カラム構成）

```
┌──────────────────────────────────────────────────────┐
│ ← 戻る  会計精算  [診療明細書] [領収書発行]（精算完了後）│
├──────────────────────────────────────────────────────┤
│ PatientInfoCard（受付情報バナー）                     │
│ 受付No: xxx | 飼主名様 - ペット名ちゃん               │
│ [カルテ確認リンク] or [入院連携リンク]               │
├──────────────────────────┬───────────────────────────┤
│ 左カラム：明細テーブル   │ 右カラム：入金パネル        │
│ AccountingItemTable      │ AccountingPaymentPanel     │
│                          │                           │
│ [物販・その他追加] ボタン │ ■ ペット保険              │
│                          │  保険ON/OFF Switch         │
│ 区分 | 項目名 | 単価(税込)│  負担割合: 50%/70%/90%/100%│
│ 数量 | 保険 | 金額       │  保険負担額（緑背景カード）│
│                          │                           │
│ 税抜小計: xxx            │ ■ 決済情報                │
│ 消費税: xxx              │  請求金額（4xl太字・中央）  │
│ 合計: ¥xxxxx             │                           │
│                          │  支払方法: [現金][カード]  │
│                          │            [電子マネー]    │
│                          │  お預かり金額: [____] 円   │
│                          │  [丁度] [千円単位] [一万単位]│
│                          │                           │
│                          │  お釣り: ¥xxxxx           │
│                          │                           │
│                          │  [会計確定] ボタン         │
└──────────────────────────┴───────────────────────────┘
```

## ヘッダーアクション（精算完了後のみ表示）

| ボタン | アイコン | 動作 |
|---|---|---|
| 診療明細書 | FileText | `handlePrint("statement")` → `AccountingDocumentPreview`表示 |
| 領収書発行 | Printer | `handlePrint("receipt")` → `AccountingDocumentPreview`表示 |

## 受付情報バナー

`受付No: {id} | {ownerName}様 - {petName}ちゃん` を表示。
`AccountingRelatedLinks`コンポーネントによる関連リンク:
- `medicalRecordId` 存在時: カルテへのリンク（FileTextアイコン）
- `hospitalizationId` 存在時: 入院詳細へのリンク（BedDoubleアイコン）

## 左カラム: 明細テーブル（`AccountingItemTable`）

### テーブル列

| 列 | className / align | 表示内容 |
|---|---|---|
| 区分 | `w-[100px]` | `Badge`（`getItemCategoryLabel`） |
| 項目名 | - | `item.name` + カルテ連携バッジ（`source === "medical_record"` 時） |
| 単価(税込) | align:right, `w-[100px]` | `¥{unitPrice}` |
| 数量 | align:center, `w-[80px]` | `item.quantity` |
| 保険 | align:center, `w-[80px]` | 適用時: 緑●、非適用時: グレー「-」 |
| 金額 | align:right, `w-[120px]` | `¥{unitPrice × quantity}` |
| 削除 | `w-[50px]` | 手動追加項目のみ Trash2 ボタン |

### 区分バッジ

| 区分値 | 表示名 |
|--------|--------|
| `examination` | 診察 |
| `test` | 検査 |
| `procedure` | 処置 |
| `surgery` | 手術 |
| `medicine` | 処方 |
| `food` | フード |
| `goods` | 物販 |
| `other` | その他 |

### 明細フッター

税抜小計 / 消費税 / 合計（太字・大文字）

### 手動追加ダイアログ（物販・その他追加）

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 区分 | `Select` | `MANUAL_ITEM_CATEGORY_VALUES`: 療法食・フード / 物販・ケア用品 / その他 |
| 品目名 | `Input` | placeholder: 例: ロイヤルカナン 3kg |
| 単価 (税込) | `Input`（type=number） | placeholder: 0 |

カルテ連携アイテム（`source === "medical_record"`）および入院連携アイテム（`source === "hospitalization"`）は削除不可。手動追加アイテム（`source === "manual"`）のみ削除可能。

## 右カラム: 入金パネル（`AccountingPaymentPanel`）

### ペット保険カード

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 保険ON/OFF | `Switch` | CardHeader内トグル |
| 負担割合 | `Select` | `INSURANCE_RATIO_VALUES`: 50%/70%/90%/100%（保険ON時のみ表示） |
| 保険負担額 | 読み取り専用 | 緑背景カード、マイナス表示 |

### 決済情報カード

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 請求金額 | 読み取り専用 | 中央配置、4xl太字 |
| 支払方法 | `Button` グリッド（3cols） | `PAYMENT_METHOD_VALUES`: 現金/カード/電子マネー、選択時は `bg-[#37352F]` |
| お預かり金額 | `Input`（type=number） | 右寄せ大文字、「円」ラベル付き |
| クイック入力 | `Button` × 3 | 「丁度」「千円単位」「一万単位」 |
| お釣り | 読み取り専用 | `bg-[#F7F6F3]` カード、不足時は赤文字 |
| 会計確定 | `Button`（full-width） | Save アイコン、disabled条件: `changeAmount < 0` or 預り金未入力 or 精算完了済 |

### 支払方法値

| 表示名 | 値 |
|--------|-----|
| 現金 | `cash` |
| クレジットカード | `credit_card` |
| 電子マネー | `electronic_money` |

## 金額計算ロジック

```
税抜小計 = Σ(単価(税込) × 数量 / 1.1) ※内税計算
消費税(内税10%) = Σ 各行消費税

合計金額(税込) = Σ(単価(税込) × 数量)

保険負担額（保険ON時）= Σ(保険対象アイテム金額) × 負担割合 × (-1)

請求金額 = 合計金額 + 保険負担額

お釣り = 預かり金額 - 請求金額（負値の場合は不足表示）
```

## 書類プレビューダイアログ（`AccountingDocumentPreview`）

- 領収書（`receipt`） / 診療明細書（`statement`）の切替表示
- プレビュー本体: `AccountingDocument`コンポーネント
- 印刷ボタン（`window.print()`）
- **保険負担内訳**: 保険適用額・自己負担額・保険者負担額の3行セット（保険設定時のみ）
- **シミュレーション注釈**: 保険負担割合別の参考金額を注記表示
- **入院連携バッジ**: `source === "hospitalization"` の明細行に「入院連携」バッジ表示
- 印刷エリア（`hidden print:block`）に`AccountingDocument`を配置

## データ初期化（`useAccountingDetail`）

| ケース | 初期化方法 |
|--------|-----------|
| 新規（カルテ連携） | `location.state.accountingItems`から明細を初期化 |
| 新規（入院連携） | `location.state.hospitalizationId` + `accountingItems`から初期化 |
| 新規（ペット選択） | `petId`クエリパラメータから空の会計を生成 |
| 既存編集 | `findAccountingById`でロード、`payment`存在時は保険・支払情報を復元 |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `AccountingDetail` | `[R]` | メインページ |
| `AccountingItemTable` | `[C]` | 明細テーブル |
| `AccountingPaymentPanel` | `[C]` | 入金パネル |
| `AccountingDocumentPreview` | `[C][M]` | 書類プレビューダイアログ |
| `AccountingDocument` | `[C]` | 印刷用書類本体（領収書/診療明細書） |
| `AccountingRelatedLinks` | `[C]` | 関連リンク（カルテ・入院詳細へのリンク） |
| `PatientInfoCard` | `[S]` | 患者情報ヘッダー |
| `PageLayout` | `[S]` | ページレイアウト（戻るボタン付き） |
| `ConfirmDialog` | `[S][M]` | 会計確定確認 |

## 使用フック

| フック | 説明 |
|---|---|
| `useAccountingDetail` | 会計CRUD・計算・書類ロジック |
| `usePrint<DocumentType>` | 印刷状態管理（初期値: `"statement"`） |

## データ型

`Accounting`, `AccountingItem`, `PaymentInfo`, `AccountingCalculation`, `DocumentType`, `ItemCategory`, `ManualItemCategory`, `PaymentMethod`, `InsuranceRatio`, `ItemSource`, `TaxRate`

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 明細追加 | 「物販・その他追加」ボタン | 追加ダイアログ開く | ダイアログ |
| 明細削除 | 行の削除ボタン（手動追加のみ） | 明細削除 | 同画面 |
| 保険適用切替 | Switch ON/OFF | 保険計算の有効/無効切替 | 同画面 |
| 負担割合変更 | Select選択 | 保険負担額の再計算 | 同画面 |
| 支払方法選択 | 支払方法ボタン3択 | 支払方法の選択 | 同画面 |
| 預かり金入力 | テキスト入力 | お釣り自動計算 | 同画面 |
| 丁度/千円/万円 | クイック入力ボタン | 請求額に応じた金額自動入力 | 同画面 |
| 会計確定 | 「会計確定」ボタン | ConfirmDialog → ステータスを`completed`へ | 同画面（確定済表示） |
| 領収書発行 | 「領収書発行」ボタン（確定後） | `AccountingDocumentPreview`モーダル表示 | モーダル |
| 診療明細書 | 「診療明細書」ボタン（確定後） | `AccountingDocumentPreview`モーダル表示 | モーダル |
| 印刷 | モーダル「印刷する」ボタン | `window.print()` | - |
| カルテへ | 受付バナー「カルテ確認」リンク | カルテ詳細へ遷移 | `/medical-records/:id` |
| 入院詳細へ | 受付バナー「入院連携」リンク | 入院詳細へ遷移 | `/hospitalization/:id` |
| 戻る | 「戻る」ボタン | `location.state.from`または会計一覧へ | `/accounting` |

## バリデーション

| 条件 | バリデーション |
|------|--------------|
| 会計確定ボタン有効化 | `receivedAmount > 0` かつ `changeAmount >= 0` かつ 未確定 |
| 会計確定後 | ボタン無効化（再確定不可） |
| 預り金不足時 | `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続 |

## アクセシビリティ

- 預り金入力: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続（不足額エラー時）
- カルテへのリンク遷移（受付バナー内）

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| 会計一覧 | `/accounting/:id` | 行クリック |
| ペット選択 | `/accounting/new?petId=xxx` | ペット選択後 |
| カルテ（会計タブ） | `/accounting/new?medicalRecordId=xxx` | 「会計へ進む」ボタン |
| 入院フォーム/詳細 | `/accounting/new?petId=xxx` | 「会計へ進む」ボタン（state経由） |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/accountings/:id` | 会計詳細取得 | 未実装 |
| POST | `/api/v1/accountings` | 会計作成 | 未実装 |
| PUT | `/api/v1/accountings/:id` | 会計更新 | 未実装 |
| POST | `/api/v1/accountings/:id/complete` | 会計確定 | 未実装 |
| GET | `/api/v1/accountings/:id/receipt` | 領収書取得 | 未実装 |
| GET | `/api/v1/accountings/:id/statement` | 診療明細書取得 | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（LocalStorageによるデータ永続化）
- バックエンドAPI: 未実装
