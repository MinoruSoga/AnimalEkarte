# スタッフ管理 仕様書 (Staff Management)

## 概要
- **画面の目的**: システムを利用する全スタッフの基本情報、職種、および LINE 予約用の公開プロフィールの統合管理。
- **URLパターン**: `/settings/staff`
- **アクセス権限**: スタッフ管理者権限が必要（`ResourceMasterStaff`）

---

## 1. 画面構成

### 1.1 スタッフ一覧テーブル
職種や稼働ステータスで絞り込み可能な、院内人員の全体リスト（氏名・職種でのカナ正規化検索付き）。
- **表示項目**: 氏名、職種、権限グループ（先頭2件+残数表示）、ステータス（有効/無効）。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 氏名、職種（有効な職種マスタからの選択）、資格番号、有効/無効ステータス。
- **システム連携**: ログイン用メールアドレス（新規作成時のみ入力可・以後は表示のみ）とパスワード、および**権限グループ**の割り当て。新規作成はメール・パスワードとも空欄ならログインアカウントなしで登録できる。メールを入力する場合はパスワードが必須（8文字以上・72バイト以下、文字と数字をそれぞれ含む）。既存のパスワードは変更する場合のみ入力する。
- **マルチクリニック**: 所属医院をチェックボックスで割り当て（新規作成時はスタッフ登録後に設定可能）。

---

## 2. LINE 予約公開設定

スタッフが飼い主向けの LINE 予約画面に表示される際の詳細情報を定義します。

| 項目 | 説明 |
|:---|:---|
| **予約ページに表示** | オンにすると LINE アプリ上の「担当医選択」に出現。 |
| **LINE 表示名** | 院内での呼称とは異なる、飼い主向けの親しみやすい名称（空欄なら氏名を使用）。 |
| **スタッフ種別** | 医師 / 看護師 / 設備 の 3 択（`doctor` / `nurse` / `resource`）。 |
| **説明文・画像 URL** | LINE 予約画面に表示する説明文、およびプロフィール画像の URL 指定（アップロード機能ではなく URL テキスト入力）。 |
| **対応可能コース** | 有効な予約区分をカテゴリ（診療 / トリミング / 入院・ホテル / その他）別に一覧し、チェックした区分のみ当該スタッフが担当可能（新規作成時はスタッフ登録後に設定可能）。 |

---

## 3. 技術仕様

### 3.1 認可の波及
ここで割り当てられた「権限グループ」は、API 側では毎リクエスト permission_group_rules から実効権限が評価されます。画面側の `/v1/me` は定期ポーリング・フォーカス時再取得を無効にし、キャッシュの staleTime は5分です（`get-me.ts`）。起動・ログイン・トークン更新の結果、または `refreshPermissions` / `ME_QUERY_KEY` の明示的無効化で更新するため、別端末での権限変更がメニューへ即時反映される保証はありません。

### 使用コンポーネント
- **`StaffSidePanel`**: `MasterSidePanel` ベースの統合編集パネル。
- **`StaffBasicInfoSection`** / **`StaffLineReservationSection`**: 基本属性・LINE 予約公開設定の編集セクション。
- **`StaffClinicsSection`** / **`StaffPermissionGroupsSection`** / **`StaffExcludedReservationTypesSection`**: 所属院・権限グループ・対応可能コースの割り当てセクション（最後の部品名には旧 excluded 名が残るが、画面は capable 側の選択 UI）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/staffs` | スタッフ一覧の取得 | `master-staff` | `view` |
| GET | `/api/v1/masters/staffs/:id` | 特定のスタッフ詳細の取得 | `master-staff` | `view` |
| POST | `/api/v1/masters/staffs` | 新規スタッフの作成 | `master-staff` | `create` |
| PATCH | `/api/v1/masters/staffs/:id` | プロフィールや権限の更新 | `master-staff` | `edit` |
| DELETE | `/api/v1/masters/staffs/:id` | スタッフの削除 | `master-staff` | `delete` |
| PATCH | `/api/v1/masters/staffs/reorder` | 表示順序の一括保存（BE実装済みだが本画面からは未呼出） | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/permission-groups` | スタッフの権限グループ割り当て取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/permission-groups` | スタッフの権限グループ割り当て更新 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/clinics` | スタッフの医院割り当て取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/clinics` | スタッフの医院割り当て更新 | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/excluded-reservation-types` | スタッフの対応不可予約区分の取得（FE は capable 側 API で永続化するため未呼出） | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/excluded-reservation-types` | スタッフの対応不可予約区分の更新（FE は capable 側 API で永続化するため未呼出） | `master-staff` | `edit` |
| GET | `/api/v1/masters/staffs/:id/capable-reservation-types` | スタッフの対応可能予約区分の取得 | `master-staff` | `view` |
| PUT | `/api/v1/masters/staffs/:id/capable-reservation-types` | スタッフの対応可能予約区分の更新 | `master-staff` | `edit` |

---
