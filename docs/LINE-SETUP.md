# LINE予約システム — 手動セットアップ & クレデンシャル管理

> **目的**: LINE公式アカウント・LINE Developers Console での手動セットアップ手順と、取得したクレデンシャルを一元管理する。
> **作成日**: 2026-04-09
> **ステータス**: 進行中

---

## 前提

- テスト用に2つの公式LINEアカウント（城東・八王子）を作成し、本番前に全フロー検証する
- 本番用アカウントは結合テスト完了後に別途作成する

---

## 1. LINE公式アカウントの作成 — 完了

**場所**: https://manager.line.biz

| # | アカウント名 | LINE ID | ステータス |
|---|-------------|---------|-----------|
| 1 | テスト-城東 | @151lnsqa | 完了 |
| 2 | テスト-八王子 | @642hdxoh | 完了 |

---

## 2. LINE Developers — チャネル構成

**場所**: https://developers.line.biz

### プロバイダー

1つのプロバイダー配下に全チャネルを配置。

```
プロバイダー: ノア動物病院
  ├── テスト-城東
  │   ├── Messaging API チャネル（Push通知用）
  │   └── LINEログイン チャネル（LIFF用）
  └── テスト-八王子
      ├── Messaging API チャネル（Push通知用）
      └── LINEログイン チャネル（LIFF用）
```

> **注意**: Messaging APIチャネルにはLIFFアプリを追加できない（2019年11月の仕様変更）。
> LIFF用には別途「LINEログイン」チャネルを作成する必要がある。

### 2a. Messaging API チャネル — 完了

Push通知送信に使用。

| # | チャネル名 | 用途 | ステータス |
|---|-----------|------|-----------|
| 1 | テスト-城東 | Push通知（LINE Messaging API） | 完了 |
| 2 | テスト-八王子 | Push通知（LINE Messaging API） | 完了 |

### 2b. LINEログイン チャネル — 完了

LIFFアプリ登録用。

| 設定項目 | 値 |
|----------|-----|
| チャネルの種類 | **LINEログイン** |
| チャネル名 | `テスト-城東 予約` / `テスト-八王子 予約` |
| チャネル説明 | `LINE予約システム LIFF用` |
| アプリタイプ | **ウェブアプリ** |

| # | チャネル名 | 用途 | ステータス |
|---|-----------|------|-----------|
| 1 | テスト-城東 予約 | LIFF（予約UI） | 完了 |
| 2 | テスト-八王子 予約 | LIFF（予約UI） | 完了 |

---

## 3. LIFFアプリ登録 — 完了

**場所**: LINE Developers > LINEログインチャネル > 「LIFF」タブ > 「追加」

| 設定項目 | 値 |
|----------|-----|
| LIFFアプリ名 | `予約` |
| サイズ | **Full**（全画面） |
| エンドポイントURL | `https://localhost:3001`（開発時。本番切替時に変更） |
| Scope | **profile**, **openid** |
| ボットリンク機能 | **On (Aggressive)** |

---

## 4. Messaging API 設定 — 完了（トークン発行済み）

**場所**: LINE Developers > Messaging API チャネル > 「Messaging API」タブ

| 手順 | 操作 |
|------|------|
| 1 | 「チャネルアクセストークン（長期）」→ 発行 |
| 2 | Webhook URL: 後で設定（staging デプロイ後） |
| 3 | Webhookの利用: 後で ON にする |

---

## 5. クレデンシャル一覧

### テスト-城東（@151lnsqa）

| キー | 値 | 取得元 |
|------|-----|--------|
| Messaging API Channel ID | `2009755545` | Messaging APIチャネル > チャネル基本設定 |
| Messaging API Channel Secret | `25e4661a8de553953a4b34c6ac7a91cb` | Messaging APIチャネル > チャネル基本設定 |
| Messaging API Access Token（長期） | `CUAtYMry8doD9ALCF/6Y0JocVqRrxC8IbOCMyRyxwDw5EJhyujJ4lQ8mVGrt7WawTi+voAxZ79mKAg+4qlsUPBVU6VMZdk7wEA42NZXQAl8gBr2tSYmZpbRzAiLfuGhxuba+koBHVk8yTuaCCjLBzAdB04t89/1O/w1cDnyilFU=` | Messaging APIチャネル > Messaging API タブ |
| LINEログイン Channel ID | `2009755586` | LINEログインチャネル > チャネル基本設定 |
| LINEログイン Channel Secret | `4575c381f7909825d03a928fa4ce61d0` | LINEログインチャネル > チャネル基本設定 |
| LIFF ID | `2009755586-nvKfG3Cp` | LINEログインチャネル > LIFF タブ |
| LIFF URL | `https://liff.line.me/2009755586-nvKfG3Cp` | — |

### テスト-八王子（@642hdxoh）

| キー | 値 | 取得元 |
|------|-----|--------|
| Messaging API Channel ID | `2009755544` | Messaging APIチャネル > チャネル基本設定 |
| Messaging API Channel Secret | `5344ef84eb7072b5894f7e087db28827` | Messaging APIチャネル > チャネル基本設定 |
| Messaging API Access Token（長期） | `pwMi3yP6jhRa0xbmnR0IPEcE5l+OIp21a7ia3hmoiuFSCvqkR5Tmmfm6fLoSTB1Bt7uQjAe9NN7fZ+LBDtNKLGnrqBrjDmhTnws9PVxQKLyinomNzUAb61KADX7NJmFBfEsLQQ9VmlU+tMJcWh+zswdB04t89/1O/w1cDnyilFU=` | Messaging APIチャネル > Messaging API タブ |
| LINEログイン Channel ID | `2009755581` | LINEログインチャネル > チャネル基本設定 |
| LINEログイン Channel Secret | `f5b318734dc824f5c6880c0623dd917b` | LINEログインチャネル > チャネル基本設定 |
| LIFF ID | `2009755581-w5NOA3EW` | LINEログインチャネル > LIFF タブ |
| LIFF URL | `https://liff.line.me/2009755581-w5NOA3EW` | — |

### 本番（未作成）

| キー | 値 | 取得元 |
|------|-----|--------|
| Messaging API Channel ID | — | — |
| Messaging API Channel Secret | — | — |
| Messaging API Access Token（長期） | — | — |
| LINEログイン Channel ID | — | — |
| LINEログイン Channel Secret | — | — |
| LIFF ID | — | — |

---

## 6. リッチメニュー作成 — 未実施（Phase 7 で実施）

**場所**: LINE Official Account Manager > 「リッチメニュー」> 「作成」

| 設定項目 | 値 |
|----------|-----|
| テンプレート | 2列×1行（2ボタン）または 3ボタン |
| ボタンA | 「予約する」→ リンク → `https://liff.line.me/{LIFF_ID}` |
| ボタンB | 「予約確認」→ リンク → `https://liff.line.me/{LIFF_ID}/my-reservations` |
| ボタンC（任意） | 「電話する」→ 電話 → 病院電話番号 |
| 表示期間 | テスト期間中 / 常時（本番） |

---

## 7. 管理画面での設定入力 — 未実施（Phase 4 完成後）

**場所**: AnimalEkarte 管理画面 > 「LINE予約 基本設定」

セクション5のクレデンシャルを管理画面に登録する。

| フィールド | 対応するクレデンシャル |
|-----------|---------------------|
| LINE Channel ID | Messaging API Channel ID |
| LINE Channel Secret | Messaging API Channel Secret |
| LIFF ID | LIFF ID |
| LINE Access Token | Messaging API Access Token（長期） |

---

## 8. 本番切替 — 未実施（全テスト完了後）

| 手順 | 操作 | ステータス |
|------|------|-----------|
| 1 | 本番用 LINE公式アカウント + チャネル作成（手順1〜4 を繰り返す） | 未実施 |
| 2 | LIFF エンドポイントURLを `https://reserve.noah-karte.com` に変更 | 未実施 |
| 3 | `reservation_settings` を本番アカウントの値に更新 | 未実施 |
| 4 | 本番リッチメニューを作成・公開 | 未実施 |
| 5 | 旧「予約 on ライン」のリッチメニューを非公開に切替 | 未実施 |
| 6 | 動作確認（予約フロー全8ステップ + キャンセル + Push通知） | 未実施 |

---

## タイムライン

| 作業 | 実施タイミング | 前提条件 |
|------|--------------|---------|
| 手順1〜4 | **今（Phase 1 着手前）** | なし |
| 手順6 | Phase 7（結合テスト時） | LIFF ID 確定済み |
| 手順7 | Phase 4 完了後 | 管理画面の基本設定画面が実装済み |
| 手順8 | 全テスト完了後 | staging で全フロー検証済み |
