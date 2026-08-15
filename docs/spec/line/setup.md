# LINE・Lステップ 連携セットアップガイド (External Integration Setup)

> **目的**: LINE Developers Console・管理画面の初期セットアップ手順を提供する。
> **読者**: クリニック導入担当・運用者。
> **タイミング**: 新規クリニックのLINE連携初期設定時。

> **Animal Ekarte**: 初期導入時の外部サービス設定手順
> **最新更新**: 2026-07-10

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

**場所**: `サイドバー「Lステップ連携」 > 「連携設定」`

| 入力項目 | 内容 |
|:---|:---|
| **Lステップ APIキー** | 手順 3.1 で取得したキー。 |
| **LINE Channel Access Token** | LINE Developers 画面で発行した「長期トークン」。画面上のラベルは「LINE Channel Access Token」。 |
| **LINE Channel Secret** | 手順 2.1 で取得したシークレット。 |
| **LIFF ID** | 手順 2.2 で取得した LIFF ID。 |

---

## 5. 疎通テストの実行

設定保存後、画面上の **「Lステップ接続テスト」** / **「LINE接続テスト」** ボタンを押下してください。
- **実装注記 (2026-07-31)**: UI は 2 ボタンだが、両方とも同一 API `POST .../lstep-settings/test-connection` を呼び出す（`use-lstep-settings.ts` コメント: single BE endpoint）。サービス側 `TestConnection` が Lステップ側と LINE Messaging 側の双方を検証する。
- **成功**: トースト通知で連携成功が表示されます。
- **失敗**: 認証情報（特にシークレットやトークン）の有効性を再確認してください。

---
