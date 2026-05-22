# LINE・Lステップ 連携セットアップガイド (External Integration Setup)

> **Animal Ekarte**: 初期導入時の外部サービス設定手順
> **最新更新**: 2026-05-21

---

## 1. 準備するもの

セットアップには以下の管理権限が必要です。
- LINE Developers アカウント。
- LINE Official Account Manager (LINE公式アカウント) 管理者権限。
- Lステップ 管理画面へのアクセス権限。
- Animal Ekarte クリニック管理者権限。

---

## 2. LINE Developers Console での設定

### 2.1 Messaging API チャネルの作成
1.  **プロバイダー**を選択または新規作成。
2.  **Messaging API** チャネルを新規作成。
3.  **Channel ID** と **Channel Secret** を控える。

### 2.2 LIFF アプリの作成
1.  チャネル内で LIFF を新規作成。
2.  URL: `https://reserve.noah-karte.com` (環境に応じて設定) を登録。
3.  Scope: `profile`, `openid` を許可。
4.  **LIFF ID** を控える。

---

## 3. Lステップ管理画面での設定

### 3.1 API キーの取得
1.  Lステップの「設定」>「API設定」へ移動。
2.  **Lステップ API キー**を生成・発行し控える。

### 3.2 タグの事前定義
自動マーケティングを機能させるため、以下のタグを Lステップ側で作成しておきます。
- `CPM_01_出会い` 〜 `CPM_05_ノア`
- `VISIT_120日超` 〜 `VISIT_240日超`
- `HLTH_健診未受診`, `HLTH_ワクチン期限間近`

---

## 4. Animal Ekarte 管理画面での登録

**場所**: `サイドバー「設定」 > 「外部連携」 > 「Lステップ連携設定」`

| 入力項目 | 内容 |
|:---|:---|
| **Lステップ APIキー** | 手順 3.1 で取得したキー。 |
| **LINE Channel ID** | 手順 2.1 で取得した ID。 |
| **LINE Channel Secret** | 手順 2.1 で取得したシークレット。 |
| **LIFF ID** | 手順 2.2 で取得した LIFF ID。 |
| **Messaging API Access Token** | LINE Developers 画面で発行した「長期トークン」。 |

---

## 5. 疎通テストの実行

設定完了後、画面上の **「接続テストを実行」** ボタンを押下してください。
- **GREEN**: 正常に連携されています。
- **RED**: 認証情報（特にシークレットやトークン）の有効性を再確認してください。

---
