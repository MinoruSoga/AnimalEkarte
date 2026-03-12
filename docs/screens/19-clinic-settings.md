# 病院情報設定 仕様書

## 概要

- **画面の目的**: 動物病院の基本情報（病院名・住所・電話番号・院長名等）の設定・管理
- **URLパターン**: `/settings/clinic`
- **アクセス権限**: 管理者権限（将来的なRBAC対応）
- **位置付け**: マスタ設定トップ（`/settings`）の「基本設定」カテゴリから遷移

## 画面レイアウト

```
┌──────────────────────────────────────────────────────┐
│ ← 病院情報設定 [Building2]                            │
├──────────────────────────────────────────────────────┤
│ Notion風ページレイアウト（px-16 スクロール可能）       │
│                                                      │
│ [Building2 ページアイコン]                            │
│ 病院情報設定  (大タイトル)                             │
│ ─────────────────────────────────────────────────   │
│ 病院名 *  │ (入力)                                   │
│ 支店名    │ (入力)                                   │
│                                                      │
│ 所在地                                               │
│ ─────────────────────────────────────────────────   │
│ 郵便番号  │ (入力)                                   │
│ 住所      │ (入力)                                   │
│                                                      │
│ 連絡先                                               │
│ ─────────────────────────────────────────────────   │
│ 電話番号      │ (入力)                               │
│ FAX番号       │ (入力)                               │
│ メールアドレス │ (入力)                               │
│ WebサイトURL  │ (入力)                               │
│                                                      │
│ その他                                               │
│ ─────────────────────────────────────────────────   │
│ 登録番号  │ (入力)                                   │
│ 院長名    │ (入力)                                   │
│                                                      │
│ 領収書や明細書、処方箋などに印字される病院情報です      │
├──────────────────────────────────────────────────────┤
│ [キャンセル]  [保存]                                  │
└──────────────────────────────────────────────────────┘
```

## フォームセクション構成

### 基本情報（セクション区切りなし、ページ最上部）

| フィールド | 入力部品 | 必須 | バリデーション |
|-----------|---------|------|--------------|
| 病院名 | `NotionPropertyRow` + `propertyInput` | 必須 | 最大50文字 |
| 支店名 | `NotionPropertyRow` + `propertyInput` | - | 最大50文字 |

### 所在地セクション

| フィールド | 入力部品 | 必須 | バリデーション |
|-----------|---------|------|--------------|
| 郵便番号 | `NotionPropertyRow` + `propertyInput` | - | `^\d{3}-\d{4}$` |
| 住所 | `NotionPropertyRow` + `propertyInput` | - | 最大200文字 |

### 連絡先セクション

| フィールド | 入力部品 | 必須 | バリデーション |
|-----------|---------|------|--------------|
| 電話番号 | `NotionPropertyRow` + `propertyInput` | - | 日本形式 |
| FAX番号 | `NotionPropertyRow` + `propertyInput` | - | 日本形式 |
| メールアドレス | `NotionPropertyRow` + `propertyInput`（type=email） | - | RFC 5322準拠 |
| WebサイトURL | `NotionPropertyRow` + `propertyInput` | - | URL形式 |

### その他セクション

| フィールド | 入力部品 | 必須 | バリデーション |
|-----------|---------|------|--------------|
| 登録番号 | `NotionPropertyRow` + `propertyInput` | - | 最大100文字（ヘルプテキスト付き） |
| 院長名 | `NotionPropertyRow` + `propertyInput` | - | 最大50文字 |

## フォームライブラリ

`react-hook-form@7.55.0` を使用。`register` でフィールドを紐付け。

## アクション

| ボタン | 動作 | 備考 |
|-------|------|------|
| キャンセル | `/settings` へ遷移 | `STYLE.sidePeekCancelBtn` |
| 保存 | `handleSubmit(onSubmit)` → `updateClinicInfo` + `reset` | `isDirty` が false の場合は disabled |

## コンポーネント構成

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ClinicSettings` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行（label / required / align） |
| `NotionSectionLabel` | `[S]` | Notion風セクションラベル（所在地・連絡先・その他） |
| `NotionSectionDivider` | `[S]` | Notion風薄罫線ディバイダー |
| `LoadingSkeleton` | `[S]` | ローディング（variant="form", rows=8、ローカルスケルトン） |
| `useClinicInfo` | `[H]` | 病院情報CRUD |

## 状態管理

| 状態 | 説明 |
|---|---|
| `clinicInfo` | `useClinicInfo` から取得した現在の病院情報 |
| `loading` | 病院情報取得中フラグ |
| `isDirty` | `react-hook-form` の変更検知フラグ（保存ボタン disabled 制御） |

## データ型

```typescript
interface ClinicInfo {
  name: string;
  branchName?: string;
  postalCode?: string;
  address?: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
}
```

## ユーザー操作フロー

1. `/settings` の「病院情報設定」カードをクリック
2. ローカルに保存されている病院情報がフォームに読み込まれる
3. 各フィールドを編集
4. 「保存」ボタンでバリデーション → `updateClinicInfo` → フォームリセット
5. 「キャンセル」または保存後に `/settings` へ戻る

## 備考

- ロゴ画像アップロード機能は型定義（`logoUrl`）のみ存在し、UI は未実装
- 領収書・明細書・処方箋等への印字データとして使用される
- ページレイアウトは Notion 風（大タイトル、プロパティ行、セクション区切り線）

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/clinics` | 病院情報取得 | 実装済 |
| PUT | `/api/v1/clinics/:id` | 病院情報更新 | 実装済 |

## 実装状況

- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 実装済
