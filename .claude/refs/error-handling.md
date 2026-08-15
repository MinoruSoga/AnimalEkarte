---
description: Go/GinとTypeScriptに共通するerror handlingの原則
---

# Error Handling

## Common principles

- error を無視しない。処理できる境界まで返すか、明示的に回復する。
- user/client 向け message と内部診断情報を分離する。
- secret、credential、個人情報、SQL、stack trace、内部 path を外部へ返さない。
- 同じ failure を複数層で重複ログしない。必要な文脈が揃う境界で1回記録する。
- retry 可能性、client action、observability に必要な stable code を設計する。

## Go

- 文脈を追加する場合は `fmt.Errorf("...: %w", err)` を使う。
- error の種類は `errors.Is` / `errors.As` で判定し、message 文字列比較をしない。
- sentinel/type は caller が programmatically 判定する必要がある場合だけ公開する。
- panic/recover を通常の validation や DB failure に使わない。
- cleanup error と primary error のどちらを返すか明示する。

## Gin HTTP boundary

- binding/validation error は client-correctable な 4xx に変換する。
- authentication、authorization、not-found、conflict、rate-limit を一貫した contract に mapping する。
- unknown error は汎用的な 500 response にし、内部 error は server-side で記録する。
- `c.Error(err)` と error middleware による集中 mapping を利用できる。
- response を書いた後に別の error response を重ねない。

特定の application error 型、helper 名、layer ごとの wrap 責務は Go/Gin 公式未規定である。backend の詳細は [go-gin-backend-guidelines.md](../rules/go-gin-backend-guidelines.md) を参照する。

## TypeScript/UI

- `catch` では `unknown` として narrowing する。
- user が回復できる message と、telemetry/debug 情報を分離する。
- promise rejection を放置せず、UI state を loading/success/error として表現する。
- API error code を型付けし、message 文字列による分岐を避ける。
