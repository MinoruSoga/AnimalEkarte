# 飼主・ペット登録/編集 仕様書 (Owner Form)

![飼主登録画面](./images/04-owners-form.png)

## 概要
- **画面の目的**: 飼い主基本情報および、紐付く全てのペット（患者）情報の一元管理。
- **URLパターン**: 
  - 新規登録: `/owners/new`
  - 編集: `/owners/:id`
- **アクセス権限**: 新規登録 (`/owners/new`) は `owners:create` を要求する `RequirePermission` ルートガードあり。編集 (`/owners/:id`) は親ルート `/owners` の `owners:view` ガードのみを継承し、保存可否（`fieldset disabled`）は `usePermission` によるコンポーネント内制御。

---

## 1. 画面構成

### 1.1 飼主情報セクション (Owner Information)
Notion スタイルの 4 カラムグリッドを採用し、臨床現場での迅速な入力をサポートします。
- **基本属性**: 飼主No、氏名（漢字・カナ）、電話番号、メールアドレス、会社名、会社電話、備考。新規かつ複数所属時は登録先医院。
- **危険人物 (`owners.is_dangerous`)**: UI ラベルは「危険人物」。飼主（人）への対応上の注意を表します。動物の取扱注意を表す `pets.danger_level` / `danger_reason` とは別概念であり、両者は混用せず、それぞれ必要な場合は併存します（DEC-14 ⑤）。
- **経済・サービス属性**: 
    - **会員区分**: 非会員 / 会員 / 退亡者 / 他診/準 の 4 区分をボタンで選択。
    - **個別値引率**: 特定の事情に基づくデフォルトの値引設定（会計時に自動適用）。
    - **DM**: 未設定／必要／不要。
- **住所管理**: 会社・自宅それぞれの郵便番号と住所1/2。地図連携は未実装。
- **飼主生年月日とサイグラム**: `OwnerAddressFields` で生年月日 DatePicker の下にサイグラムを縦積みする（`flex-col`）。空・不正時は視覚表示なし（読み上げは「サイグラム分類なし」）。横並びで入力欄幅を潰さない。値は `birth_date` として保存する。
- **保存 form**: `noValidate`。HTML5 email/max が Action 前に黙って止めない。重複メール／電話は BE 409。

### 1.2 ペット管理セクション (Pet Management)
一人の飼い主に紐付く複数のペットをテーブル形式の一覧で管理します。

| 項目 | 説明 |
|:---|:---|
| **基本情報** | 飼主配下一覧の列はペット番号・ペット名・生死・種別・性別・生年月日・毛色・体重・環境・備考・操作。モーダルは識別／身体／ケアの独立見出し。動物種は識別列、品種は身体列。 |
| **臨床ステータス** | **生存 (`alive`) / 死亡 (`deceased`)**。死亡時は日付と理由を記録。 |
| **身体的特徴** | 毛色、体重、去勢・避妊手術日、血液型、マイクロチップ番号。 |
| **行動・安全属性** | **危険度 (`danger_level`)**: 咬癖や攻撃性がある場合、`高 / 中 / 低` で設定。一覧画面で警告が表示されます。 |
| **医療背景** | 常用フード、飼育環境（自由入力）、加入保険（マスタ選択）。 |

- **動物取扱注意 (`pets.danger_level` / `danger_reason`)**: `danger_reason` は動物を安全に取り扱うための理由です。理由の入力は危険度「高」の場合のみ必須で、「中」では任意です。スタッフ内部情報として扱い、Owner Report・LIFF・line-reserve の owner-facing 応答にはフィールド自体を含めません。
- **飼主変更 (BUG-373)**: `PetEditModal` からペットの紐付け先飼主を変更可能（`PATCH /api/v1/pets/:id` の `owner_id`）。変更先飼主の値引率/会員区分が現飼主と異なる場合、会計金額への影響を警告する確認モーダルを挟んでから確定する。
- **副飼主**: 永続済みペットに `PetSubOwnersSection`。409 時は再読込を求める。

### 1.3 LINE/Lステップ連携セクション (編集時のみ、`LineIntegrationCard`)
- **紐付け状況**: LINE User ID の取得状態をリアルタイム表示。
- **連携用URLの発行**（未連携時、`LineLinkTokenSection`、SD-14）: 「連携用URLを発行」ボタンで `POST /api/v1/owners/:id/line/link-token` を呼び、返却された LIFF URL（[38-liff-pet-health.md](./38-liff-pet-health.md) の `LiffLinkPage` 紐付けフロー参照）を読み取り専用入力欄に表示しコピーできる。`owners` の edit 権限でゲート。
- **配信除外**: 「配信除外」スイッチによる、リマインドの一時停止機能（`PATCH /clinics/:clinicId/owners/:id/delivery-exclusion`）。
- **配信注意フラグ**: リマインドを止めずに注意喚起のみ行うフラグ＋理由メモ（`PATCH /clinics/:clinicId/owners/:id/delivery-caution`）。配信除外とは独立した別スイッチ。
- **転院ステータス**: 転院済みフラグの切替（`PATCH /clinics/:clinicId/owners/:id/transfer-status`）。
- **個別メッセージ送信**: 
    - **`LineSendPanel`**: サイドパネルから特定の飼い主へ直接 LINE メッセージを送信。
    - **ファイル共有**: 血液検査結果などの PDF や画像をアップロードし、LINE 経由で共有可能（**`shared_files`** ストレージ連携）。

### 1.4 会計履歴セクション（編集時のみ）
- 該当飼主の会計履歴を一覧表示。`accounting` の閲覧権限を持つユーザーにのみ表示されます（権限がない場合、見出しごと非表示）。

---

## 2. 主要な臨床安全機能

### 2.1 危険個体の視覚的警告
`danger_level` が「高」に設定されたペットは、飼主一覧において、即座に目立つ警告バッジが表示され、スタッフの安全確保を促します。
- **理由の開示**: 高危険度バッジは、クリック / タップ / Enter / Space のいずれか1操作で開くアクセシブルな Popover とし、`danger_reason` を表示します。理由が空の既存データでは `理由未登録` を表示します。

### 2.2 データの真正性保護
- **保存アクション**: React 19 の `useActionState` を活用し、保存中の二重送信防止とエラー箇所への自動フォーカス移動を実現。
- **離脱ブロック**: 編集途中のページ遷移時に、`NavigationBlocker` が変更破棄の確認を求めます。

---

## 3. 技術仕様

### 使用コンポーネント
- **`OwnerForm`**: 統合フォーム。
- **`PetEditModal`**: ペット情報の詳細入力（遅延ロード対応）。
- **`LineIntegrationCard`**: 外部連携管理部品。
- **`LineSendPanel`**: 個別メッセージおよび共有ファイル送信 UI。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/owners/:id` | 飼主・ペット情報の取得 | `owners` | `view` |
| POST | `/api/v1/owners` | 新規登録（ペット一括登録含む） | `owners` | `create` |
| PATCH | `/api/v1/owners/:id` | 飼主基本情報の更新 | `owners` | `edit` |
| PATCH | `/api/v1/pets/:id` | ペット単体の属性変更 | `owners` | `edit` |
| POST | `/api/v1/shared-files` | LINE 共有用ファイルのアップロード | `owners` or `medical-records` | `edit` (or `create`/`edit`) |
| POST | `/api/v1/clinics/:clinicId/owners/:id/line/send` | LINE個別メッセージ送信 | `owners` | `edit` |
| GET | `/api/v1/clinics/:clinicId/owners/:id/line/send-logs` | LINEメッセージ送信履歴取得（`pending` 行がある間は5秒間隔でポーリング） | `owners` | `view` |
| PATCH | `/api/v1/clinics/:clinicId/owners/:id/line-user-id` | LINE User ID の手動設定・解除 | `owners` | `edit` |
| PATCH | `/api/v1/clinics/:clinicId/owners/:id/line-id-confirm` | LINE ID 確認の記録 | `owners` | `edit` |
| POST | `/api/v1/owners/:id/line/link-token` | 連携用トークン + LIFF URL の発行（SD-14） | `owners` | `edit` |
| PATCH | `/api/v1/clinics/:clinicId/owners/:id/delivery-exclusion` | 配信除外フラグの切替 | `owners` | `edit` |
| PATCH | `/api/v1/clinics/:clinicId/owners/:id/delivery-caution` | 配信注意フラグ・理由の切替 | `owners` | `edit` |
| PATCH | `/api/v1/clinics/:clinicId/owners/:id/transfer-status` | 転院ステータスの切替 | `owners` | `edit` |

---
