# 飼主・ペット登録/編集 仕様書

## 概要
- **画面の目的**: 飼い主情報とペット情報の登録・編集
- **URLパターン**:
  - 新規登録: `/owners/new`
  - 編集: `/owners/:id`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────────────┐
│ ← 戻る  飼主・ペット登録/編集  [登録/更新ボタン]             │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│ ■ 飼主情報（4カラムグリッド、max-w-[1400px]）               │
│  飼主No | 郵便番号 | 会社名 | 会員区分（ボタングループ）     │
│  飼主名 * | 住所1 | 郵便番号(自宅) | [危険人物 Switch]      │
│  飼主名(カナ)* | 住所2 | 住所1(自宅) | [備考 Textarea]      │
│  飼主生年月日 | メールアドレス | 住所2(自宅)               │
│  電話番号 * | 会社電話番号（colspan=2） | 値引率(%)          │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│ ■ ペット情報 [ペット追加ボタン]                              │
│  テーブル: ペット番号 | ペット名 | 生 | 種別 | 性別 |       │
│           生年月日 | 毛色 | 体重 | 環境 | 備考 | 操作       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## 表示項目（飼主情報フォーム）

| フィールド名 | 項目ID | 入力部品 | 必須 | 説明 | 備考 |
|------------|--------|---------|------|------|------|
| 飼主No | `ownerId` | `Input` | - | 飼主ID番号 | 編集時無効化 |
| 郵便番号 | `postalCode` | `Input` | - | 郵便番号（会社） | |
| 会社名 | `company` | `Input` | - | 会社名 | |
| 会員区分 | `membershipType` | `Button` | - | 会員区分ボタン | 非会員/会員/退亡者/他診/準 |
| 飼主名 | `ownerName` | `Input` | ✅ | 飼い主の氏名 | |
| 住所1 | `address1` | `Input` | - | 会社住所1 | |
| 郵便番号(自宅)| `homePostalCode`| `Input` | - | 自宅郵便番号 | |
| 危険人物 | `isDangerous` | `Switch` | - | 危険人物フラグ | |
| 備考・特記事項| `remarks` | `Textarea` | - | メモ | |
| 飼主名(カナ) | `ownerNameKana` | `Input` | ✅ | 氏名カナ | |
| 住所2 | `address2` | `Input` | - | 会社住所2 | |
| 住所1(自宅) | `homeAddress1` | `Input` | - | 自宅住所1 | |
| 飼主生年月日 | `birthDate` | `NotionDatePicker` | - | 飼主生年月日 | |
| メールアドレス| `email` | `Input` | - | メールアドレス | |
| 住所2(自宅) | `homeAddress2` | `Input` | - | 自宅住所2 | |
| 電話番号 | `phone` | `Input` | ✅ | 電話番号 | |
| 会社電話番号 | `companyPhone` | `Input` | - | 会社電話 | |
| 値引率 (%) | `discountRate` | `Input` | - | 飼主別値引率 | |

## PetEditModal フォーム項目（3カラムグリッド: md:2, lg:3）

| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|----------|---------|---------|------|------|
| ペットNo | `petNumber` | `Input` | - | |
| ペット名 | `petName` | `Input` | ✅ | |
| ペット名(カナ)| `petNameKana` | `Input` | - | |
| 動物種 | `animalSpeciesId`| `Select` | ✅ | マスタ連動 |
| 性別 | `gender` | `Select` | ✅ | |
| 生年月日 | `birthDate` | `NotionDatePicker` | - | |
| 品種 | `breed` | `Input` | - | |
| 毛色 | `color` | `Input` | - | |
| 体重(kg) | `weight` | `Input` | - | |
| 去勢・避妊手術日| `neuteredDate`| `NotionDatePicker`| - | |
| 入手区分 | `acquisitionType`| `Select` | - | 購入/譲渡/保護/その他 |
| ペットの危険度| `dangerLevel` | `Select` | - | 低/中/高 |
| 食べ物 | `food` | `Input` | - | |
| 飼育環境 | `environment` | `Input` | - | |
| 保険 | `insuranceId` | `Select` | - | マスタ連動 |
| 備考・特記事項| `remarks` | `Textarea` | - | |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `OwnerForm` | `[R]` | メインページ |
| `PetEditModal` | `[C][M]` | ペット追加/編集モーダル |
| `PageLayout` | `[S]` | ページレイアウト（戻るボタン付き） |
| `NavigationBlocker` | `[S]` | フォーム離脱保護（未保存変更あり時に確認ダイアログ） |
| `NotionDatePicker` | `[S]` | 日付ピッカー |
| `FormFieldError` | `[S]` | フィールドエラーメッセージ（`role="alert"`） |
| `ConfirmDialog` | `[S][M]` | ペット削除確認 |
| `PrimaryButton` | `[S]` | 登録/更新ボタン |
| `useOwnerForm` | `[H]` | フォーム状態管理・保存・バリデーション |
| `useUnsavedChanges` | `[H]` | 未保存変更検知（`markDirty` / `markClean` / `isDirty`） |

## 機能詳細

### 1. フォーム離脱防止ガード
- **未保存検知**: `useUnsavedChanges` フックにより、フォームの変更（`isDirty`）を監視。
- **ナビゲーションブロック**: 変更がある状態で画面を離れようとすると、`NavigationBlocker` により確認ダイアログが表示される。保存成功時に `markClean()` を呼び出すことでブロックを解除する。

### 2. ペット情報のライフサイクル管理
- **ローカル管理**: `PetEditModal` での追加・編集・削除の結果は、まず `OwnerForm` の React 状態（`pets` 配列）に一時的に反映される。
- **一括更新**: 画面最下部の「登録/更新」ボタン押下時に、飼主情報とペット情報の全リストがバックエンドへ送信される。

### 3. 入力支援ロジック
- **自動ID生成**: 新規登録時、飼主Noはバックエンド側で自動採番される。
- **日付入力**: `NotionDatePicker` を使用し、カレンダー選択だけでなく、キーボードによるテキスト入力（"2020/01/01"等）にも対応。

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 保存 | 「登録/更新」ボタン | バリデーション → 保存 → `markClean()` | 成功後: `/owners` |
| キャンセル/戻る | 「戻る」ボタン | 一覧画面へ戻る | `/owners` |
| ペット追加 | 「ペット追加」ボタン | `PetEditModal` を開く（新規） | モーダル表示 |
| ペット編集 | ペット行クリック or ドロップダウン「詳細・編集」 | `PetEditModal` を開く（編集） | モーダル表示 |
| ペット削除 | ドロップダウン「削除」 | `ConfirmDialog` 表示後に削除 | 同画面 |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| 飼主一覧 | `/owners/new` | 新規登録ボタン |
| 飼主一覧 | `/owners/:id` | 行クリック |
| 保存成功 | `/owners` | 保存完了後 |

## バリデーション
- **飼主名**: 必須（未入力時にフィールドエラー表示）
- **飼主名(カナ)**: 必須（未入力時にフィールドエラー表示）
- **電話番号**: 必須（未入力時にフィールドエラー表示）
- **ペット名**: 必須（PetEditModal内でバリデーション）
- **種**: 必須（PetEditModal内）
- **性別**: 必須（PetEditModal内）
- **生年月日**: 必須（PetEditModal内）
- バリデーション失敗時: トースト通知（PetEditModal）またはフィールドエラー表示（飼主フォーム）
- アクセシビリティ: 必須フィールドに `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `ownerData` | `OwnerData` | 飼主フォームデータ |
| `pets` | `PetInfo[]` | ペット一覧 |
| `petModalOpen` | `boolean` | PetEditModal 表示フラグ |
| `editingPet` | `PetInfo \| null` | 編集中のペットデータ |
| `fieldErrors` | `Record<string, string>` | フィールドエラーメッセージ |
| `isDirty` | `boolean` | 未保存変更フラグ（`useUnsavedChanges`） |
| `deletePetTarget` | `{ id: string; name: string } \| null` | 削除対象ペット（ConfirmDialog 用） |

## データ型
`OwnerData`, `PetInfo`, `PetFormData`, `PetEditModalData`, `MembershipType`

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| POST | `/api/v1/owners` | 飼主作成 | 実装済 |
| PATCH | `/api/v1/owners/:id` | 飼主更新 | 実装済 |
| GET | `/api/v1/owners/:id` | 飼主詳細取得 | 実装済 |
| POST | `/api/v1/pets` | ペット作成 | 実装済 |
| PATCH | `/api/v1/pets/:id` | ペット更新 | 実装済 |
| GET | `/api/v1/pets/:id` | ペット詳細取得 | 実装済 |
| DELETE | `/api/v1/pets/:id` | ペット削除 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/owners/routes/OwnerForm.tsx`）
- バックエンドAPI: 実装済（`handler/owner_handler.go`, `handler/pet_handler.go`）

## 備考
- 旧仕様と比較して飼主フォームのフィールドが大幅に増加（郵便番号・会社名・会員区分・自宅住所・危険人物フラグ・値引率 等）
- ペット情報は別テーブルとしてモーダル（`PetEditModal`）で管理。旧仕様のインラインフォームから変更
- 旧仕様にあったペットフィールド（マイクロチップID・保険情報）は PetEditModal 内に統合
- `NavigationBlocker` により未保存状態での離脱時に確認ダイアログが表示される
