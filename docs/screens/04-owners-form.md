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

| フィールド名 | 入力部品 | 必須 | 説明 | 備考 |
|------------|--------|------|------|------|
| 飼主No | `Input` | - | 飼主ID番号 | `owners.id` |
| 郵便番号 | `Input` | - | placeholder: 123-4567 | `owners.postal_code` |
| 会社名 | `Input` | - | 会社名 | `owners.company` |
| 会員区分 | `Button` グループ | - | `MEMBERSHIP_TYPE_VALUES`（非会員/会員/退亡者/他診/準） | `owners.membership_type` |
| 飼主名 | `Input` | ✅ | 飼い主の氏名 | `owners.owner_name`、`aria-invalid`+`aria-describedby` 接続 |
| 住所1 | `Input` | - | 会社住所1 | `owners.address1` |
| 郵便番号(自宅) | `Input` | - | 自宅郵便番号 | `owners.home_postal_code` |
| 危険人物 | `Switch` | - | 「該当する」ラベル | `owners.is_dangerous` |
| 飼主名(カナ) | `Input` | ✅ | 氏名のフリガナ（カタカナ推奨）、placeholder: ハヤシ フミアキ | `owners.owner_name_kana`、`aria-invalid`+`aria-describedby` 接続 |
| 住所2 | `Input` | - | 会社住所2 | `owners.address2` |
| 住所1(自宅) | `Input` | - | 自宅住所1 | `owners.home_address1` |
| 備考・特記事項 | `Textarea` | - | min-h-[140px]、rowspan | `owners.remarks` |
| 飼主生年月日 | `NotionDatePicker` | - | 飼主の誕生日 | `owners.birth_date` |
| メールアドレス | `Input`（type=email） | - | メールアドレス | `owners.email` |
| 住所2(自宅) | `Input` | - | 自宅住所2 | `owners.home_address2` |
| 電話番号 | `Input` | ✅ | placeholder: 090-1234-5678 | `owners.phone`、`aria-invalid`+`aria-describedby` 接続 |
| 会社電話番号 | `Input` | - | colspan=2 | `owners.company_phone` |
| 値引率 (%) | `Input`（type=number） | - | 飼主専用割引率 | `owners.discount_rate` |

## 表示項目（ペット一覧テーブル）

| フィールド名 | 説明 | 備考 |
|------------|------|------|
| ペット番号 | ペットの患者番号 | `pets.pet_number` |
| ペット名 | ペットの名前 | `pets.name` |
| 生 | 生死ステータス | `pets.status` |
| 種別 | 犬/猫等 | `pets.animal_species_id` → `animal_species` |
| 性別 | 雄/雌/不明 | `pets.gender` |
| 生年月日 | ペットの誕生日 | `pets.birth_date` |
| 毛色 | 毛の色 | `pets.color` |
| 体重 | 体重(kg) | `pets.weight` |
| 環境 | 飼育環境 | `pets.environment` |
| 備考 | メモ（truncate 表示） | `pets.remarks` |
| 操作 | ドロップダウンメニュー | `MoreHorizontal size-5` アイコン |

## ペット行ドロップダウンアクション

| アクション | 遷移先 |
|-----------|--------|
| 詳細・編集 | `PetEditModal` を開く |
| 予約作成 | `/reservations?petId=xxx` |
| カルテ作成 | `/medical-records/new?petId=xxx` |
| トリミング | `/trimming/new?petId=xxx` |
| 入院登録 | `/hospitalization/new?petId=xxx` |
| 会計登録 | `/accounting/new?petId=xxx` |
| 削除 | `ConfirmDialog` 表示後に削除（destructive スタイル） |

## PetEditModal フォーム項目（3カラムグリッド: md:2, lg:3）

| フィールド | 入力部品 | 必須 | 備考 |
|----------|---------|------|------|
| ペット名 | `Input` | ✅ | `aria-invalid`+`aria-describedby` 接続 |
| ペット名カナ | `Input` | - | |
| 種 | `Select`（`animal_species` マスタ連動） | ✅ | `pets.animal_species_id` FK |
| 性別 | `Select`（雄/雌/不明） | ✅ | `PET_GENDER_VALUES` |
| 生年月日 | `NotionDatePicker` | ✅ | |
| 品種 | `Input` | - | |
| 毛色 | `Input` | - | |
| 避妊去勢日 | `NotionDatePicker` | - | |
| 入手種別 | `Select`（購入/譲渡/保護/その他） | - | `ACQUISITION_TYPE_VALUES` |
| 危険度 | `Select`（低/中/高） | - | `DANGER_LEVEL_VALUES` |
| フード | `Input` | - | |
| 保険名 | `Select`（`insurances` マスタ連動） | - | `pets.insurance_id` FK |
| 保険詳細(負担割合) | `Select`（50%/70%/90%/100%/その他） | - | `PET_INSURANCE_RATIO_VALUES` |
| 備考・特記事項 | `Textarea` | - | |

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
| POST | `/api/v1/owners` | 飼い主作成 | 未実装 |
| PUT | `/api/v1/owners/:id` | 飼い主更新 | 未実装 |
| GET | `/api/v1/owners/:id` | 飼い主詳細取得 | 未実装 |
| POST | `/api/v1/pets` | ペット作成 | 実装済 |
| PUT | `/api/v1/pets/:id` | ペット更新 | 実装済 |
| GET | `/api/v1/pets/:id` | ペット詳細取得 | 実装済 |

## 実装状況
- フロントエンド(ui-sample): 実装済
- バックエンドAPI: ペットCRUD実装済、飼い主API未実装

## 備考
- 旧仕様と比較して飼主フォームのフィールドが大幅に増加（郵便番号・会社名・会員区分・自宅住所・危険人物フラグ・値引率 等）
- ペット情報は別テーブルとしてモーダル（`PetEditModal`）で管理。旧仕様のインラインフォームから変更
- 旧仕様にあったペットフィールド（マイクロチップID・保険情報）は PetEditModal 内に統合
- `NavigationBlocker` により未保存状態での離脱時に確認ダイアログが表示される
