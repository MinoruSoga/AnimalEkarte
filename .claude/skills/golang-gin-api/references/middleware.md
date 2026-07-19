# Gin middleware review

## Scope and order

- global middleware は全 route に必要なものだけにする。
- route group middleware で public/protected、version、resource scope を表現する。
- middleware order による値・header・recovery・logging の依存を test する。
- request を拒否したら `Abort*` と return を行う。

## Security

- CORS origin は allowlist。credential と wildcard origin を組み合わせない。
- authentication と authorization を分離する。
- CSRF は cookie/bearer 等の認証方式に合わせる。
- rate limiting key は spoofable な header を無条件に信用しない。
- trusted proxy を明示する。
- security header と body/upload limit を deployment に合わせる。

## Dependency injection

- middleware dependency は closure で capture できる。
- application dependency を `gin.Context` に格納すると型 assertion が必要になるため、handler closure/struct を優先する。
- request-scoped identity/trace 等だけを context value として扱う。

Sources:

- https://gin-gonic.com/en/docs/middleware/using-middleware/
- https://gin-gonic.com/en/docs/middleware/dependency-injection/
- https://gin-gonic.com/en/docs/middleware/security-guide/
