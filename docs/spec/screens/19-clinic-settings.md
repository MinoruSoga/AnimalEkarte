# 医院マスタ設定 仕様書 (Clinic Master Settings)

## 概要
- **画面の目的**: システムを利用する各拠点（本院・分院）の基本情報、連絡先、およびインボイス制度に対応した各種パラメータの一元管理。
- **URLパターン**: `/settings/clinic`
- **アクセス権限**: 医院管理者権限が必要（`ResourceHospitalSettings`）

---

## 1. 画面構成

### 1.1 医院一覧テーブル
登録されている全拠点のリスト。
- **表示項目**: 院名、電話番号、メール、ステータス。

### 1.2 詳細編集サイドパネル (`ClinicMasterSidePanel`)
- **基本情報**: 院名、ステータス（有効/無効）、郵便番号、住所、電話/FAX番号、登録番号、院長名、メール、Webサイト。
- **消費税率**: 通常課税・軽減税率の定義。
- **明細兼領収書設定**: ロゴ表示・登録番号警告・項目カテゴリ・各セクション（病院情報ヘッダー／飼主・ペット情報／明細テーブル／お会計サマリー）の表示可否をトグルで切替。セクション表示順の並べ替えとフッター文言の編集も可能（ロゴ画像自体のアップロード機能はこの画面には無い）。

### 1.3 法人情報セクション (`CompanyInvoiceSection`)
- 医院一覧の上部に表示される、法人（シングルトン、`companies` テーブル）のインボイス登録番号のみを編集する小セクション。
- 拠点（`Clinic`）ごとの「登録番号」（1.2）とは別エンティティ。法人単位でインボイス発行事業者登録番号を1つだけ保持する。
- 保存すると全拠点の会計伝票・領収書（`AccountingDetailPage` 経由）へ即座に反映される。

---

## 2. 主要な運営機能

### 2.1 マルチテナントの核
ここで定義された各医院が、スタッフの「所属」の選択肢となり、全ての診療・会計データの論理的なテナント分離境界（`clinic_id`）の源泉となります。

### 2.2 帳票の適格性保証
登録された所在地、代表者名、インボイス番号は、システムから発行される全ての領収書・明細書へ正確に反映され、臨床以外の事務コストを大幅に削減します。

---

## 3. 技術仕様

### 使用コンポーネント
- **`ClinicMasterSettings`**: メインページ。
- **`ClinicMasterList`**: 医院一覧テーブル（`topSection` に `CompanyInvoiceSection` を差し込む）。
- **`ClinicMasterSidePanel`**: コンテキストを維持した詳細編集パネル。
- **`CompanyInvoiceSection`**: 法人（`Company` シングルトン）のインボイス登録番号編集セクション。`useGetCompany` / `useUpdateCompany`（`frontend/src/features/master`）を使用。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics` | 医院一覧取得（この画面は `scope=all` 指定でスタッフ割当外も含む全件を取得。詳細パネルも一覧データから参照） | `hospital-settings` | `view` |
| POST | `/api/v1/clinics` | 新規拠点の開設 | `hospital-settings` | `create` |
| PATCH | `/api/v1/clinics/:clinic_id` | 拠点情報の更新 | `hospital-settings` | `edit` |
| DELETE | `/api/v1/clinics/:clinic_id` | 拠点情報の削除（物理削除。参照中のデータが残る場合は 409 で拒否） | `hospital-settings` | `delete` |
| GET | `/api/v1/company` | 法人情報取得（インボイス登録番号を含む） | `hospital-settings` | `view` |
| PATCH | `/api/v1/company` | 法人情報の部分更新（本画面ではインボイス登録番号のみ編集可能） | `hospital-settings` | `edit` |

---

