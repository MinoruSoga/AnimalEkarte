# 飼主・ペット登録/編集 仕様書

## 概要
- **画面の目的**: 飼い主情報とペット情報の登録・編集
- **URLパターン**:
  - 新規登録: `/owners/new`
  - 編集: `/owners/:id`
- **アクセス権限**: 認証済ユーザー全員

## 画面構成
- **ヘッダー**: タイトル、戻るボタン、保存ボタン（登録/更新）
- **飼主情報セクション**: 4カラムグリッド形式の入力フォーム
- **ペット情報セクション**: ペット一覧テーブル、ペット追加ボタン
- **編集モーダル**: ペット情報の詳細入力（`PetEditModal`）

## 表示項目（飼主情報フォーム）

| フィールド名 | 項目ID | 入力部品 | 必須 | 備考 |
|------------|--------|---------|------|------|
| 飼主No | `ownerId` | `Input` | - | 編集時無効 |
| 郵便番号 | `postalCode` | `Input` + 検索 | - | 会社 |
| 会社名 | `company` | `Input` | - | |
| 会員区分 | `membershipType` | `ButtonGroup` | - | 非会員/会員/退亡者/他診/準 |
| 飼主名 | `ownerName` | `Input` | ✅ | |
| 住所1 | `address1` | `Input` | - | 会社 |
| 郵便番号(自宅)| `homePostalCode`| `Input` + 検索 | - | |
| 危険人物 | `isDangerous` | `Switch` | - | |
| 備考・特記事項| `remarks` | `Textarea` | - | |
| 飼主名(カナ) | `ownerNameKana` | `Input` | ✅ | |
| 住所2 | `address2` | `Input` | - | 会社 |
| 住所1(自宅) | `homeAddress1` | `Input` | - | |
| 飼主生年月日 | `birthDate` | `NotionDatePicker` | - | |
| メールアドレス| `email` | `Input` | - | |
| 住所2(自宅) | `homeAddress2` | `Input` | - | |
| 電話番号 | `phone` | `Input` | ✅ | |
| 会社 電話番号 | `companyPhone` | `Input` | - | |
| 値引率 (%) | `discountRate` | `Input` | - | |

## ペット編集モーダル (`PetEditModal`)

| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|----------|---------|---------|------|------|
| ペットNo | `petNumber` | 表示のみ | - | 自動採番 |
| ペット名 | `petName` | `Input` | ✅ | |
| ペット名(カナ)| `petNameKana` | `Input` | - | |
| CFBE | `animalSpeciesId`| `Select` | ✅ | 種別マスタ連動 |
| 性別 | `gender` | `Select` | ✅ | 雄/雌/不明 |
| 生年月日 | `birthDate` | `NotionDatePicker` | - | |
| BREED | `breed` | `Input` | - | 品種 |
| 毛色 | `color` | `Input` | - | |
| 体重(kg) | `weight` | `Input` | - | |
| 去勢・避妊手術日| `neuteredDate`| `NotionDatePicker`| - | |
| 入手区分 | `acquisitionType`| `Select` | - | 購入/譲渡/保護/その他 |
| ペットの危険度| `dangerLevel` | `Select` | - | 低/中/高 |
| 食べ物 | `food` | `Input` | - | |
| 飼育環境 | `environment` | `Input` | - | |
| 保険 | `insuranceId` | `Select` | - | 保険マスタ連動 |
| 備考・特記事項| `remarks` | `Textarea` | - | |

## 特徴的な機能
- **郵便番号検索**: 会社・自宅それぞれの郵便番号から住所を自動入力。
- **フォーム離脱保護**: 未保存の変更がある場合、`NavigationBlocker` により確認ダイアログを表示。
- **飼主変更**: `PetEditModal` 内からペットの飼主を別の既存飼主に変更可能。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| POST | `/api/v1/owners` | 飼主作成 |
| PATCH | `/api/v1/owners/:id` | 飼主更新 |
| POST | `/api/v1/pets` | ペット作成 |
| PATCH | `/api/v1/pets/:id` | ペット更新 |
| DELETE | `/api/v1/owners/:id` | 飼主削除 |
