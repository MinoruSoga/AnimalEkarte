# BE: BUG-067 NULL バイトを含む入力で 500 → 400 に修正

## 概要

フロントエンドの NULL バイトサニタイズは修正済み（axios インターセプター）。
しかし API 直接呼び出しで NULL バイトを含む文字列を送ると 500 が返る。
PostgreSQL は NULL バイト含む文字列を拒否するため。

## フロントエンド対応（修正済み）

`src/lib/axios.ts` のリクエストインターセプターで POST/PATCH/PUT の body から `\u0000` を自動除去。

## 残存問題（バックエンド）

```
POST /api/v1/owners
{"name": "テスト\u0000ヌル", ...}
→ HTTP 500（期待: 400 Bad Request）
```

## 期待する動作

- NULL バイトを含む文字列の場合は 400 Bad Request を返す
- または service/handler でサニタイズして処理を続行する

## 実装場所

- `backend/internal/handler/` または `service/` の共通バリデーション
- 全 string フィールドの NULL バイトチェックを追加

## 優先度

High（意図しないクラッシュ）

## 関連

- `docs/tasks/open/crash/BUG-067_null_byte_500.md`
- FUNCTIONAL_TEST_REPORT.md BUG-067
