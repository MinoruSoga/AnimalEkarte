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

## 表示・フォーム項目（明細一覧）

| フィールド | 内容 | 備考 |
|-----------|------|------|
| 区分 | `Badge` (診察/検査/処置/手術/処方/フード/物販/その他) | `ItemCategory` 型 |
| 項目名 | 名称 + 「カルテ連携」バッジ | `source === "medical_record"` 時 |
| 単価 | `unitPrice` | 三桁区切り表示 |
| 数量 | `quantity` | |
| 保険 | 緑点バッジ | `isInsuranceApplicable === true` 時 |
| 金額 | 小計 (単価 * 数量) | |
| 削除 | `Trash2` ボタン | `source === "manual"` 時のみ |

## フォーム項目（ペット保険）

| フィールド | 入力部品 | 備考 |
|-----------|----------|------|
| 保険利用 | `Switch` | |
| 負担割合 | `Select` | 50% / 70% / 90% / 100% |
| 保険負担額| 読み取り専用 | 緑背景カード、マイナス表示 |

## フォーム項目（決済情報）

| フィールド | 入力部品 | 備考 |
|-----------|----------|------|
| 支払方法 | ボタン選択 | 現金 / カード / 電子マネー |
| お預かり金額| `Input(number)` | 右寄せ、通貨単位付き |
| クイック入力| `Button` × 3 | 丁度 / 千円単位 / 一万単位 |
| お釣り | 読み取り専用 | 不足時は赤文字 |

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

## 帳票出力仕様

`[C] AccountingDocument` により、以下の2種類の帳票を出力する。

### 1. 領収書 (`receipt`)
- **形式**: 80mm幅レシート形式
- **印字項目**:
  - 宛名（飼主名 + 「様」）
  - ペット名
  - **領収金額**（中央・巨大太字）
  - 診療費等合計、保険負担額（適用時）、請求金額、お預かり、お釣り
  - 病院情報（名称、住所、TEL、登録番号）

### 2. 診療明細書 (`statement`)
- **形式**: A4サイズ形式
- **印字項目**:
  - 病院情報、発行日、会計No
  - 宛名、ペット名、品種
  - **明細テーブル**: 項目名、カテゴリ、単価、数量、金額
  - **集計エリア**: 合計金額、10%対象額・税額、8%対象額・税額、保険適用額、最終請求金額

## 機能詳細

### 1. 混在消費税の自動計算
- **税率別集計**: 各明細に設定された税率（8% または 10%）に基づき、税抜小計および消費税額を個別に算出する。
- **端数処理**: 現状の実装では切り捨て処理を採用し、`taxBreakdown` オブジェクトで内訳を管理する。

### 2. ペット保険の窓口精算ロジック
- **マイナス計上**: 保険適用が有効（`hasInsurance: true`）な場合、選択された負担割合（50%〜100%）に基づき、保険負担額をマイナス項目の明細として自動生成する。
- **請求額連動**: 合計金額から保険負担額を差し引いた額が、最終的な「請求金額」となる。

### 3. 入金と釣銭管理
- **リアルタイム計算**: 「お預かり金額」の入力に対し、請求金額との差額を即座に算出し、不足がある場合は「不足」として赤文字で表示する。
- **確定時の不整合防止**: お釣りがマイナス（不足）の状態では「会計確定」ボタンが disabled となり、未収金の発生をUIレベルで防止する。

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
| GET | `/api/v1/accountings/:id` | 会計詳細取得 | 実装済 |
| PATCH | `/api/v1/accountings/:id` | 会計データ更新 | 実装済 |
| POST | `/api/v1/accountings` | 会計データ作成 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/accounting/routes/AccountingDetail.tsx`）
- バックエンドAPI: 実装済（`handler/accounting_handler.go`）
