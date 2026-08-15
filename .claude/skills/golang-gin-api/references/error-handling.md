# Gin error boundary

- handler は error を無視しない。
- known application error は一貫した status/code に mapping する。
- unknown error は generic 500 とし、raw message、SQL、stack、path、secret、個人情報を返さない。
- `c.Error(err)` と error middleware による集中処理を利用できる。
- response を書いた後に別の response を重ねない。
- recovery は panic から server を守るために使い、validation/DB error の通常経路にしない。
- 同じ error を複数 package で重複ログしない。

特定の `AppError` 型や helper 名は Gin公式要件ではない。

Source: https://gin-gonic.com/en/docs/middleware/error-handling-middleware/
