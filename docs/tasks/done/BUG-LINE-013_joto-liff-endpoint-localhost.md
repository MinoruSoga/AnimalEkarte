# BUG-LINE-013: 城東 LIFF エンドポイント URL が localhost のまま

## 概要

城東 LIFF アプリ (`2009755586-nvKfG3Cp`) のエンドポイント URL が `https://localhost:3001` のままで、LINE アプリから開くと ERR_CONNECTION_REFUSED で失敗する。

## 再現

1. ブラウザで `https://liff.line.me/2009755586-nvKfG3Cp` にアクセス
2. LINE OAuth を経由して `https://localhost:3001` にリダイレクトされる
3. `ERR_CONNECTION_REFUSED` エラー

## 原因

LINE Developers Console の LIFF 設定で、エンドポイント URL が開発時の値から更新されていない。

参照: `docs/line/setup.md` 手順3、8「本番切替 — 未実施」

## 比較

| クリニック | LIFF ID | 現在のエンドポイント | 状態 |
|---|---|---|---|
| 八王子 (clinic_id=3) | 2009755581-w5NOA3EW | `https://stg.noah-karte.com/line-reserve/3` | ✅ 動作 |
| **城東** (clinic_id=4) | 2009755586-nvKfG3Cp | `https://localhost:3001` | ❌ 接続拒否 |

## 修正手順

LINE Developers Console で設定変更（コード変更不要）:

1. https://developers.line.biz にアクセス
2. プロバイダー「ノア動物病院」→ チャネル「テスト-城東 予約」(LINEログイン)
3. 「LIFF」タブ → LIFF アプリ「予約」を編集
4. エンドポイント URL を `https://stg.noah-karte.com/line-reserve/4` に変更
5. 保存

## 優先度

**MEDIUM** — 城東用 LIFF アプリ自体は動くが、LINE アプリからのアクセスが不可。結合テスト前に修正必要。
