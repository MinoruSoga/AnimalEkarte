---
name: gin-api-design
description: "Gin の resource routing / binding / error を設計する。発火は API 契約を決めるとき。folder/layer は規定しない。"
origin: ECC (adapted for AnimalEkarte)
---

# Gin API Design

## When to activate

- endpoint の追加・変更
- route group、versioning、pagination、filtering の設計
- request/response/error contract のreview
- authentication/authorization middleware scope の設計

最初に `.claude/rules/go-gin-backend-guidelines.md` の Gin 節を読む。

## Design sequence

1. resource、actor、authorization、ownership を定義する。
2. method/path、request、response、status、error code を OpenAPI で定義する。
3. route group と middleware scope を決める。
4. body/query/URI/header の型付き input と validation を決める。
5. pagination/filter/sort の limit と stable ordering を決める。
6. compatibility、idempotency、concurrency、rate limit を確認する。
7. `httptest` で contract を検証する。

## Resource routing

Gin公式は API が成長した場合、resource ごとの route registration と `RouterGroup` による整理を示している。

```go
func RegisterUserRoutes(group *gin.RouterGroup, h *UserHandler) {
    users := group.Group("/users")
    users.GET("", h.List)
    users.GET("/:id", h.Get)
    users.POST("", h.Create)
}
```

- URL path versioningではgroup単位で管理できる。header versioningも選択肢であり、採用方式をcontract全体で一貫させる。
- public と protected route を同じ registration の暗黙条件にしない。
- resource route のファイル分割は Gin の拡張例であり、application package 構成の公式要件ではない。

## Binding and validation

- body/query/URI/header に適切な binder/tag を使う。
- custom response を返す API では `ShouldBind*` を優先し、error を必ず処理する。
- required、format、length、range、enum を境界で検証する。
- validation 成功を authorization 成功とみなさない。
- trim/lowercase は normalization であり、SQL injection/XSS 対策ではない。

## Responses and errors

- API 内で response/error format を一貫させる。
- internal model や database error を直接返さない。
- known application error は stable status/code、unknown error は generic 500 にする。
- create/update/delete の status と body は OpenAPI contract に明記する。
- particular envelope、helper、DTO suffix は Gin公式要件ではない。

## Collection endpoints

- limit に上限を設ける。
- page/offset または cursor の semantics を明記する。
- stable ordering と tie-breaker を持たせる。
- filter/sort field は allowlist 化し、入力を SQL fragment として連結しない。
- response に next cursor または pagination metadata を一貫して含める。

## Security review

- authentication、authorization、ownership を別々に確認する。
- credential 付き CORS で wildcard origin を使わない。
- auth 方式に合う CSRF、secure cookie、rate limit を設定する。
- tenant/ownership boundary は `.claude/refs/backend-application-invariants.md` に従う。

## Sources

- [Gin API design](https://gin-gonic.com/en/docs/routing/api-design/)
- [Gin grouping routes](https://gin-gonic.com/en/docs/routing/grouping-routes/)
- [Gin binding](https://gin-gonic.com/en/docs/binding/)
- [Gin error middleware](https://gin-gonic.com/en/docs/middleware/error-handling-middleware/)
- [Gin security guide](https://gin-gonic.com/en/docs/middleware/security-guide/)
