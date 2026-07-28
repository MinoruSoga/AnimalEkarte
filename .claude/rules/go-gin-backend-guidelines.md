---
paths:
  - "backend/**/*.go"
---
# Go/Gin Backend Guidelines

> **正本**: Go/Gin バックエンドの一般規約はこの文書を正本とする。
> `.agents/rules/` は `.claude/scripts/sync-agents-skills.sh` による生成物であり、直接編集しない。
> 最終確認: 2026-07-19（Go 公式・Gin 公式の一次資料）、2026-07-22（AnimalEkarte project decision）

## 1. 適用範囲と読み方

この文書は、Go 公式ドキュメントと Gin 公式ドキュメントから導ける、HTTP サーバー開発の共通ベースラインを定義する。

規則の根拠を次の3種類に分ける。

- **Go公式**: Go toolchain の仕様、Go 公式ドキュメント、Go team のレビュー指針。
- **Gin公式**: Gin 公式ドキュメントが示す API、推奨または代表的パターン。
- **公式未規定**: チームが必要に応じて ADR で決める設計。Go/Gin公式要件として扱わない。

公式サンプルのファイル名やディレクトリ名は説明用であり、その形自体が必須とは限らない。アプリケーション固有の認可、テナント隔離、監査、データ保護は `.claude/refs/backend-application-invariants.md` を併用する。

## 2. モジュールとパッケージ

### Go公式に基づく方針

- サーバー内部だけで使うコードは `internal/` に置く。Go toolchain が親ツリー外からの import を拒否する。
- 複数の実行コマンドがある場合は `cmd/<command>/` に整理できる。単一コマンドの小規模プロジェクトでは必須ではない。
- 最初から複雑な階層を作らない。コードが増え、凝集性や再利用境界が明確になってからパッケージを分ける。
- 同じディレクトリの複数 `.go` ファイルは1つのパッケージを構成する。ファイル分割は可読性のためであり、アーキテクチャ境界ではない。
- package は「何を提供するか」が明確な、凝集した単位にする。循環 import は作れないため、依存方向が自然になる境界を選ぶ。

最小の例:

```text
go.mod
cmd/
  api-server/
    main.go
internal/
  <cohesive-package>/
```

`<cohesive-package>` は `handler`、`service`、`repository`、ドメイン名のいずれにも固定しない。

### パッケージ名

- 短く、明確な、小文字の1単語を優先する。
- `under_scores`、`mixedCaps`、意味の曖昧な `util`、`common`、`misc` を避ける。
- export 名に package 名を繰り返さない。例: `http.HTTPServer` のような stutter を避ける。
- 頻繁に一緒に import される package 同士で同名を避け、呼び出し側の rename を常態化させない。
- getter に `Get` を機械的に付けない。所有者や境界が読み取れる名前を選ぶ。

### 公式未規定

Go/Gin は次を規定しない。

- layer-first / domain-first の選択
- Handler → Service → Repository、Clean Architecture、repository pattern
- `service/<domain>`、`model/`、`pkg/` などの固定配置
- ディレクトリの深さ、1 package のファイル数、ファイルの行数
- Go ファイル名の単語区切り
- request/response DTO の固定配置

必要なら、実測した依存関係と変更単位を根拠に ADR で決める。

### AnimalEkarte project decision（公式未規定）

[`docs/product-philosophy.md`](../../docs/product-philosophy.md) は product の WHAT / WHY と判断順序を定める上位文書であり、folder tree を直接規定しない。本projectはその判断順序を実装へ落とすため、[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md) で **domain/capability-first の modular monolith** を採用する。

- top-level package は `internal/<domain>` を基本とし、route、use case、transaction、persistence、testを業務能力ごとのvertical sliceとして変更できる境界にする。
- BE9-2B以降、既存未移行codeの保守・安全修正と移動に必要なcompatibility変更を除き、新規production実装を`internal/handler`、`internal/service`、`internal/repository`へ追加しない。新規実装はADR-006のtarget domain packageへ置く。BE9移行は2026-07-24にcode complete（release pending）となり、境界の正本は[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md)と[boundary map](../../docs/architecture/be9-2a-boundary-map.md)とする。
- domain package内に`handler`、`service`、`repository` subpackageを機械的に作らない。同じ利用者・変更単位なら、責務を別file/型に分けた同一packageでよい。実際に独立したconsumer、依存方向、変更周期が生じた場合だけsubpackageへ分離する。
- 1つのbusiness factには1つのsource of truthとwrite ownerを置く。`appointments`とそのlifecycleのwrite ownerは`reservation`とし、他domainはbusiness intentを表すconsumer-sideの最小interfaceまたは明示的なorchestrationを通じて操作する。owner外へ任意fieldを変更できるgeneric update APIを公開しない。`appointments`はBE9-2E-0で収束済みであり、境界は[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md)を正本とし、自動回帰gateは[`appointment_write_owner_lint_test.go`](../../backend/internal/reservation/appointment_write_owner_lint_test.go)で維持する。
- appointmentに紐づく通常カルテは一般診療予約だけを対象とし、appointmentごとにactive recordを最大1件とする。カルテ日付は予約日時のJST日付から導出し、紐づいている間は独立変更させない。削除は対象カルテをlockしたtransaction内で見積依存を再確認してからdraft限定の原子的soft deleteを行い、見積作成・確定との競合を同じ親行lockで直列化する。先行した見積がある場合や非draftならConflictとし、削除が先行した場合は後続見積を拒否する。予約検証、重複確認、transaction依存の欠落や失敗を成功扱いにせずfail-closedにする。
- cross-domain transactionは、write owner、transaction境界、失敗時のatomicityを明示する。別domainのtableへ独立実装から直接writeする経路を増やさない。
- appointment、trimming detail、option等で1つのbusiness graphを構成するwriteは同じtransactionで全体を成功またはrollbackさせる。既存trimming appointmentのowner欠損はappointment fieldを変更しないdetail-only writeでも同じtransaction内で補完する。参照master・担当者・LINE顧客を検証する必須依存が欠ける場合はwrite前にfail-closedとし、LIFFで明示指定されたstaffにはclinic所属・対応可能種別に加えて`is_active=true`かつ`reservation_visible=true`を要求する。best-effortを採用する場合は、部分成功contract、再試行、補償、監査を明示する。
- clinicalまたはfinancial integrityのためfail-closedと定めた監査はbusiness writeと同じtransactionへ参加させ、監査dependency欠落または監査write失敗時はbusiness writeもrollbackする。締め後の会計編集はこの対象とする。
- write結果を返すための再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗応答へ反転させないcontractにする。
- `FOR UPDATE`、`FOR SHARE`、transaction-scoped advisory lockに依存するoperationはambient transaction不在をfail-closedにする。request由来のclinic-scoped FKはwriteと同じtransactionで最終検証し、並行master変更で判定が無効になる場合は対象行をcommitまで固定する。
- nested GORM `Preload`のpredicateは末尾associationだけへ適用される。clinic-ownedの中間associationも独立したclinic predicateでscopeし、破損FKを経由したcross-clinic detail復元を防ぐ。
- migration中のcompatibility facadeは薄いdelegateまたはtype aliasに限定する。同じbusiness ruleやwrite処理を複製せず、consumer移行後の削除条件を持たせる。
- 自動化は手動で安全に完結できる同じuse caseを基礎とし、停止手段、失敗通知、監査、手動fallback、重複実行を防ぐidempotencyまたは明示的retry policyを備える。
- 自動status transitionは候補readだけを根拠にせず、write時のcompare-and-setでstatus・時刻・tenant・遷移を否定するbusiness evidenceを再評価する。resource単位の監査が必須なtransitionは、状態変更と監査を同じtransactionでfail-closedにする。同じevidenceを逆向きに変更する競合workflowがある場合は、両者を同じresource-scoped serialization機構へ参加させ、各writeのcommitまで順序を保持する。
- 効率化よりclinical safety、clinic isolation、authorization、auditabilityを優先する。package配置だけをこれらの成立根拠にしない。

## 3. Package API と interface

- interface は一般に利用側で定義し、利用側が本当に必要とする最小メソッド集合にする。
- 実装 package は具体型を返すことを基本とする。利用側は必要ならそれを小さな interface として受け取る。
- mock を作る目的だけで interface を先に作らない。代替実装や呼び出し側の抽象化が必要になった時点で導入する。
- export は最小限にし、公開識別子には GoDoc で読めるコメントを付ける。
- constructor は不変条件や依存関係の確立に価値がある場合に使う。空の値が有効なら不要な constructor を作らない。
- zero value を有用にできる型は、有用に設計する。

## 4. `context.Context`

- request-scoped のキャンセル、deadline、trace を使う関数は `context.Context` を第1引数として明示的に受け取る。
- Context を struct の field に保存しない。request ごとに引数で渡す。
- Gin の入口から DB・外部 API へは `c.Request.Context()` を伝播する。
- `context.WithCancel`、`WithTimeout`、`WithDeadline` が返す cancel 関数は必ず呼ぶ。
- optional parameter や一般的な依存注入に Context value を使わない。request-scoped で、process/API 境界を越える値に限定する。
- request と無関係な処理へ機械的に Context を追加しない。

## 5. Gin の組み立てと依存注入

Gin公式が示す代表的な選択:

- 依存が少ない handler は closure で依存を capture する。
- 共有依存や複数 handler が増えたら、依存を持つ struct に handler method を定義する。
- `gin.Context` に依存を格納する middleware injection は型 assertion が必要になるため、closure または struct を優先する。
- test double を渡せるよう、依存を package global に固定しない。

```go
type UserHandler struct {
    store UserStore
}

func NewUserHandler(store UserStore) *UserHandler {
    return &UserHandler{store: store}
}

func (h *UserHandler) Get(c *gin.Context) {
    user, err := h.store.Find(c.Request.Context(), c.Param("id"))
    // bind/validation/error mapping/response
}
```

Go/Gin は、DI を `main.go` だけで行うこと、DI container の採否、特定の層間依存を規定しない。

## 6. Routing と middleware

- `RouterGroup` で共通 prefix と middleware scope を表現する。
- public route と authenticated/authorized route の境界を route 登録時に明示する。
- API が増えたら、resource ごとの route 登録関数やファイルに分ける。これは Gin 公式の拡張パターンであり、アプリケーション全体の package 構成を強制しない。
- middleware は必要な scope にだけ適用し、登録順を意識する。
- middleware は request を中断する場合 `Abort*` を使い、後続 handler が実行されないことを保証する。
- URL path versioning を採用する場合は `RouterGroup` で整理できる。header versioning も選択肢であり、選んだ方式と互換性contractをAPI全体で一貫させる。

```go
func RegisterUserRoutes(group *gin.RouterGroup, h *UserHandler) {
    users := group.Group("/users")
    users.GET("/:id", h.Get)
}
```

## 7. Binding、validation、出力

- body、query、URI、header に合った binder と tag を使う。
- application が response を制御する通常の API では `ShouldBind*` を優先し、返った error を必ず処理する。`Bind*` は禁止 API ではない。
- 外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。
- trim や lowercase は正規化であり、SQL injection や XSS 対策の代替ではない。
- validation 後も authorization と resource ownership を独立して確認する。
- response は一度だけ書き、return して handler を終了する。
- API 全体で成功・error・pagination の contract を一貫させる。
- 内部 model を無条件に JSON 化せず、公開 contract に必要な field だけを返す。

特定の envelope、DTO 名、変換関数名は Gin 公式要件ではない。

## 8. Error handling

- error は、それを処理できる境界まで返す。無視しない。
- `%w` と `errors.Is` / `errors.As` で error chain を保持する。
- 同じ error を複数レイヤーで重複ログしない。request 境界など、十分な文脈を持つ場所で1回記録する。
- 例外として、未知 pg コードはサーバー側欠陥の診断に SQLSTATE が不可欠なため、`c.Error(err)` による request 境界での記録を必須とし、domain service 側に同じ error のログが残る重複を許容する。
- Gin では `c.Error(err)` と error-handling middleware による集中 mapping を利用できる。
- 既知の application error を安定した HTTP status/code に変換し、未知の error は汎用的な 500 response にする。
- DB error、stack trace、SQL、file path、secret、個人情報を response に含めない。
- panic recovery は process 全体の crash 回避用であり、通常の error handling の代わりにしない。

独自 error type や helper の採用は可能だが、特定の型名、helper 名、どの package で wrap/log するかは公式未規定である。

## 9. Security

- production traffic は HTTPS を使用し、必要に応じて HSTS と security header を設定する。
- CORS は明示的な allowlist を使う。credential を許可する場合、wildcard origin を使わない。
- cookie は用途に応じて `Secure`、`HttpOnly`、`SameSite`、適切な expiry を設定する。
- 認証方式に応じた CSRF 対策を行う。
- authentication と authorization を分け、route/resource ごとに両方を検証する。
- rate limit、request/body/upload size、content type、file path を制限する。
- SQL は parameterized query または安全な ORM binding を使い、入力で SQL を組み立てない。
- proxy 配下では信頼する proxy を明示する。proxy を使わない場合は `SetTrustedProxies(nil)` を検討する。
- client IP、forwarded header、host header を無条件に信用しない。
- secret、token、credential、個人情報をコード、error、log に出さない。

アプリケーション固有の clinic/owner/pet/staff 分離は `.claude/refs/backend-application-invariants.md` に従う。

## 10. Goroutine

- goroutine から元の `*gin.Context` を直接使わない。Gin の値が必要なら `c.Copy()` を使う。
- Gin 固有情報が不要なら、標準 `context.Context` と immutable な必要値だけを渡す。
- goroutine の終了条件、cancel、panic、error の通知経路を設計し、request 終了後に無制限に残さない。
- request body や response writer を request 終了後に参照しない。

## 11. Production server lifecycle

- timeout や shutdown を制御する production server は `http.Server` で構成する。
- `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout`、`IdleTimeout`、header/body size は workload と infrastructure に合わせる。固定値を公式値として扱わない。
- SIGINT/SIGTERM を受け、timeout 付き `Server.Shutdown` で graceful shutdown する。
- `http.ErrServerClosed` は通常終了として扱う。
- DB、queue、worker などの resource も終了順序を決めて close する。
- 起動時に必須設定を検証し、不完全な状態で serving を開始しない。

## 12. Testing と toolchain

- `gofmt` を適用し、`go test`、`go vet` と project の lint を CI で実行する。
- HTTP handler と middleware は `net/http/httptest` で検証する。
- Gin test では必要に応じて `gin.TestMode` と最小 router を使う。
- 正常系だけでなく、binding error、validation、authentication、authorization、not found、internal error を検証する。
- cancellation/deadline、graceful shutdown、concurrent access は risk に応じて test する。
- test から global state を漏らさず、subtest を並列化する場合は共有 state の安全性を確認する。

AnimalEkarteではproject decisionとして、write ownerを変更するsliceにowner外の直接writeがないこと、許可された状態遷移、cross-domain transactionのrollbackを確認するtestを含める。write-ownerのAST gateは`FirstOrCreate`を含むmutation、typed parameter、free function/receiver method戻り値とcross-file factory、query変数、local/package constant、table alias、直接または変数代入した`TableName()`、静的string helper、generic appointment map mutatorを検出対象にする。自動処理を変更する場合は、停止、失敗通知、監査、手動fallback、重複実行またはretry時の安全性を変更riskに応じて検証する。

coverage threshold、TDD 手順、unit/integration/E2E の必須構成は Go/Gin 公式では規定されない。project quality gate として採用する場合は別規約として扱う。

## 13. 公式ベストプラクティスとして扱わないもの

次は採用してもよいが、Go/Gin公式の要件ではない。

- Handler → Service → Repository / Clean Architecture
- repository interface、service interface の一律必須化
- DI は `main.go` のみ
- logging は service のみ
- validator を単一ファイルへ集約
- GORM 固有 helper、transaction pattern、CRUD method 名
- DTO/model/request/response の固定配置
- package/file/directory の固定サイズ
- P1–P18 のような project 固有 compliance 番号
- tenant scope の具体的実装
- coverage 80% などの固定閾値
- immutability の一律強制

これらが必要なら、公式由来の規約と混ぜず、ADR または application invariant として根拠・適用範囲・検証方法を記録する。

## 14. 一次資料

### Go

- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Package names](https://go.dev/blog/package-names)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Canceling in-progress database operations](https://go.dev/doc/database/cancel-operations)
- [`context` package documentation](https://pkg.go.dev/context)

### Gin

- [API design patterns](https://gin-gonic.com/en/docs/routing/api-design/)
- [Dependency injection](https://gin-gonic.com/en/docs/middleware/dependency-injection/)
- [Grouping routes](https://gin-gonic.com/en/docs/routing/grouping-routes/)
- [Using middleware](https://gin-gonic.com/en/docs/middleware/using-middleware/)
- [Model binding and validation](https://gin-gonic.com/en/docs/binding/)
- [Error handling middleware](https://gin-gonic.com/en/docs/middleware/error-handling-middleware/)
- [Security guide](https://gin-gonic.com/en/docs/middleware/security-guide/)
- [Trusted proxies](https://gin-gonic.com/en/docs/server-config/trusted-proxies/)
- [Custom HTTP configuration](https://gin-gonic.com/en/docs/server-config/custom-http-config/)
- [Goroutines inside middleware](https://gin-gonic.com/en/docs/middleware/goroutines-inside-a-middleware/)
- [Graceful restart or stop](https://gin-gonic.com/en/docs/server-config/graceful-restart-or-stop/)
- [Testing](https://gin-gonic.com/en/docs/testing/)
