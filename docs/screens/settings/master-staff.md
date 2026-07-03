# スタッフ管理 仕様書 (Staff Management)

## 概要
- **画面 of Purpose**: システムを利用する全スタッフの基本情報、職種、および LINE 予約用の公開プロフィールの統合管理。
- **URLパターン**: `/settings/staff`
- **アクセス権限**: スタッフ管理者権限が必要（`ResourceMasterStaff`）

---

## 1. 画面構成

### 1.1 スタッフ一覧テーブル
職種や稼働ステータスでソート可能な、院内人員の全体リスト。
- **表示項目**: 氏名、職種（獣医師/看護師/トリマー/受付）、アカウント連携状況、有効/無効。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 氏名（漢字/カナ）、職種、所持資格、略歴。
- **システム連携**: ログイン用メールアドレスの紐付け、および**権限グループ**の割り当て。
- **マルチクリニック**: 複数拠点に所属する場合、それぞれの院での役割やメイン所属院の設定を個別管理。

---

## 2. LINE 予約公開設定

スタッフが飼い主向けの LINE 予約画面に表示される際の詳細情報を定義します。

| 項目 | 説明 |
|:---|:---|
| **予約ページに表示** | オンにすると LINE アプリ上の「担当医選択」に出現。 |
| **LINE 表示名** | 院内での呼称とは異なる、飼い主向けの親しみやすい名称。 |
| **自己紹介・画像** | 専門分野や得意な治療の説明文、およびプロフィール写真のアップロード。 |
| **対応不可メニュー** | 院内全メニューのうち、当該スタッフが担当できない区分（例：看護師が手術を担当等）を論理的に除外。 |

---

## 3. 技術仕様

### 3.1 認可の波及
ここで割り当てられた「権限グループ」は、スタッフが再ログイン（またはトークン更新）した際に、全ての API エンドポイントに対するアクセス可否を決定します。

### 使用コンポーネント
- **`MasterSelectModal`**: 職種、権限、所属院の高速検索・選択。
- **`ClinicAssignmentGrid`**: 複数拠点への所属状況を可視化する専用部品。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/staffs` | スタッフ一覧の取得 | `master-staff` | `view` |
| GET | `/api/v1/masters/staffs/:id` | 特定のスタッフ詳細の取得 | `master-staff` | `view` |
| POST | `/api/v1/masters/staffs` | 新規スタッフの作成 | `master-staff` | `create` |
| PATCH | `/api/v1/masters/staffs/:id` | プロフィールや権限の更新 | `master-staff` | `edit` |
| DELETE | `/api/v1/masters/staffs/:id` | スタッフの削除 | `master-staff` | `delete` |
| PATCH | `/api/v1/masters/staffs/reorder` | 表示順序の一括保存 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/permission-groups` | スタッフの権限グループ割り当て取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/permission-groups` | スタッフの権限グループ割り当て更新 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/clinics` | スタッフの医院割り当て取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/clinics` | スタッフの医院割り当て更新 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/excluded-reservation-types` | スタッフの対応不可予約区分の取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/excluded-reservation-types` | スタッフの対応不可予約区分の更新 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/capable-reservation-types` | スタッフの対応可能予約区分の取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/capable-reservation-types` | スタッフの対応可能予約区分の更新 | `master-staff` | `edit` |

---

