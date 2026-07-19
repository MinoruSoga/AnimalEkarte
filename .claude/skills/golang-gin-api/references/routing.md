# Gin routing review

- `RouterGroup` で共通 prefix と middleware scope をまとめる。
- API が成長したら resource ごとに `Register...Routes(*gin.RouterGroup)` を分ける。
- registration function は受け取った group に route を登録し、hidden global router に依存しない。
- versioning を採用するなら group 単位で管理する。
- public/protected route の境界を registration で明示する。
- duplicate route、wildcard conflict、middleware omission を route test で検出する。

resource route のファイル分割は Gin公式の拡張 pattern だが、application 全体の folder architecture を規定しない。

Sources:

- https://gin-gonic.com/en/docs/routing/api-design/
- https://gin-gonic.com/en/docs/routing/grouping-routes/
