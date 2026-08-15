---
name: golang-gin-api
description: "Gin公式に基づくHTTP server infrastructure。router/middleware、trusted proxy、security、goroutine safety、http.Server timeout、graceful shutdown、httptestを扱う。application layer構成はgo-gin-backendを参照。"
license: MIT
metadata:
  version: 3.0.0-animal-ekarte
---

# Gin HTTP Infrastructure

application/package 設計は `go-gin-backend`、API contract は `gin-api-design` を併用する。

## Server composition

- middleware を明示的な順序で登録する。
- trusted proxy を deployment topology に合わせて設定する。proxy を使わない場合は `SetTrustedProxies(nil)` を検討する。
- dependency は closure または struct handler で注入する。
- route group で prefix と middleware scope を表現する。
- production lifecycle を制御する場合は `http.Server` を使う。

```go
srv := &http.Server{
    Addr:              addr,
    Handler:           router,
    ReadHeaderTimeout: readHeaderTimeout,
    ReadTimeout:       readTimeout,
    WriteTimeout:      writeTimeout,
    IdleTimeout:       idleTimeout,
}
```

timeout の値は workload/infrastructure に合わせる。sample の秒数を公式固定値として使わない。

## Graceful shutdown

- SIGINT/SIGTERM を受ける。
- timeout 付き Context を作り、cancel を defer する。
- `Server.Shutdown` で in-flight request を待つ。
- `http.ErrServerClosed` を通常終了として扱う。
- DB、worker、queue 等を明示した順序で close する。

## Middleware

- recovery は最後の防御であり、通常の error handling の代替にしない。
- logging は request ID/trace と status/latency を structured field で記録し、秘密情報を除外する。
- CORS は allowlist、authn/authz は route scope、rate limit は信頼できる identity/key に基づける。
- middleware が request を拒否する場合は abort し、後続 handler を実行させない。

## Goroutine safety

- goroutine から元の `*gin.Context` を使わない。
- Gin 固有値が必要なら `c.Copy()` を使う。
- 通常は標準 `context.Context` と必要な immutable value だけを渡す。
- goroutine の終了条件、cancel、error/panic 通知を設計する。

## Verification

- `httptest.NewRecorder` と最小 router で middleware/route を test する。
- middleware order、abort、header、status、body、panic recovery を確認する。
- shutdown/cancellation は server lifecycle を変更した場合に test する。

## References

- [routing.md](references/routing.md)
- [middleware.md](references/middleware.md)
- [error-handling.md](references/error-handling.md)
- [Gin custom HTTP configuration](https://gin-gonic.com/en/docs/server-config/custom-http-config/)
- [Gin official documentation](https://gin-gonic.com/en/docs/)
