# Backend Coding Rules

この文書は backend 作業の入口である。Go/Gin の設計・実装規約の正本は [Go/Gin Backend Guidelines](../.claude/rules/go-gin-backend-guidelines.md)、レビュー手順は [Go/Gin Backend Review](../.claude/refs/go-gin-backend-review.md) とする。

## Authority

1. Go language/toolchain の仕様
2. Go 公式・Gin 公式ドキュメントに基づく正本
3. [Backend Application Invariants](../.claude/refs/backend-application-invariants.md) と ADR
4. API/OpenAPI、database schema、migration などの project contract
5. package 内の局所的な説明

Go/Gin公式が規定しない package layout や layer pattern を、公式要件として追加してはならない。

## Package design

- package は凝集した責務で分ける。現在のディレクトリ名を新規設計へ機械的に複製しない。
- package 名は短く明確な小文字の1単語を優先し、`util`、`common`、`misc` を避ける。
- package API は小さく保ち、exported name の stutter を避ける。
- interface は一般に利用側で、実際に必要な最小メソッドだけを定義する。
- implementation は concrete type を返すことを基本とし、mock のためだけの interface を作らない。
- `internal/` は外部 module から import させない code に使う。`cmd/` は複数 command の entry point を整理する場合に使う。

Handler → Service → Repository、Clean Architecture、layer-first/domain-first は Go/Gin 公式規約ではない。採用・変更する場合は、依存関係と変更単位を根拠に ADR で決める。

## HTTP with Gin

- route group で prefix と middleware scope を表現する。
- public route、authentication、authorization の境界を route 登録時に明示する。
- handler の dependency は closure または struct で注入し、package global に固定しない。
- application が error response を制御する通常の endpoint では `ShouldBind*` を優先し、binding error を必ず処理する。
- body/query/URI/header を型付き input に bind し、型・形式・範囲・長さを境界で検証する。
- response は公開 contract に必要な field だけを含める。
- OpenAPI contract と route、request、response、status code を同期する。
- error mapping は一貫した境界または middleware に集約し、未知の error は内部情報を含まない 500 にする。

## Context and dependencies

- request-scoped 処理には `c.Request.Context()` を渡し、DB と外部 API まで cancellation/deadline を伝播する。
- Context を struct に保存しない。
- `WithCancel` / `WithTimeout` / `WithDeadline` の cancel 関数を必ず呼ぶ。
- dependency は constructor、closure、struct field など型安全な方法で明示する。
- `gin.Context` への dependency 格納は型 assertion が必要なため、必要な request-scoped 値に限定する。

## Errors and logging

- error を無視しない。必要な文脈は `%w` で追加し、`errors.Is` / `errors.As` を使う。
- 同じ error を複数箇所で重複ログしない。十分な request 文脈を持つ境界で1回記録する。
- log は構造化し、secret、token、credential、owner/pet/staff/medical data を含めない。
- DB error、SQL、stack trace、内部 path を response に出さない。
- panic/recover は通常の error 処理の代替にしない。

特定の error helper、logger package、wrap/log の配置は公式未規定である。既存 helper を変更する場合は互換性と error chain を確認する。

## Persistence and integrity

- query は parameterized し、DB call へ Context を渡す。
- transaction の開始、commit、rollback、resource ownership を明確にする。
- schema constraint と application validation の両方を使う。
- migration は versioned SQL で管理し、application startup の暗黙 migration に依存しない。
- clinic-scoped data は、ORM、raw SQL、join、preload、count、bulk operation、background job のすべてで認証済み clinic に制約する。
- client が送信した clinic/owner/pet/staff ID を認可根拠にせず、関連 resource の ownership を server-side で確認する。

ORM や repository pattern の採否は Go/Gin公式未規定である。GORM 固有の規則を Go/Gin 公式規約と呼ばない。

## Security and production

- HTTPS、明示的 CORS allowlist、安全な cookie、認証方式に合う CSRF 対策を行う。
- trusted proxy を明示し、proxy を使わない場合は信頼 proxy を無効化する。
- authentication と authorization を別々に確認する。
- rate limit、request/body/upload size、content type を制限する。
- `http.Server` に workload に合う timeout/limit を設定する。
- SIGINT/SIGTERM を受け、timeout 付き `Shutdown` と resource close を行う。
- goroutine から元の `*gin.Context` を使わない。必要なら `c.Copy()`、通常は標準 Context と必要値だけを渡す。

## Tests and verification

- HTTP handler/middleware は `net/http/httptest` と最小 router で検証する。
- binding、validation、authentication、authorization、not-found、conflict、internal-error を含める。
- tenant/ownership boundary の変更には unauthorized/cross-tenant test を含める。
- cancellation、concurrency、transaction、shutdown は変更 risk に応じて検証する。
- `gofmt`、`go test`、`go vet`、project lint を適用する。
- この repository では Go command を host で直接実行せず、Docker の scoped command を使う。
- full-project command の自動実行禁止は [`.claude/CLAUDE.md`](../.claude/CLAUDE.md) に従う。

coverage threshold や TDD workflow は project quality policy であり、Go/Gin公式のアーキテクチャ要件ではない。

## Before review

- [ ] package boundary と名前が利用者・凝集性を反映している
- [ ] Context、error chain、resource cleanup が維持されている
- [ ] input validation、authentication、authorization、ownership が独立している
- [ ] response/log に内部情報や個人情報がない
- [ ] clinic isolation がすべての data path で維持されている
- [ ] OpenAPI と実装が一致している
- [ ] scoped test と static checks が通っている
- [ ] 独自設計を「Go/Gin公式」と表現していない

## Primary sources

- [Go module layout](https://go.dev/doc/modules/layout)
- [Go package names](https://go.dev/blog/package-names)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Gin API design](https://gin-gonic.com/en/docs/routing/api-design/)
- [Gin dependency injection](https://gin-gonic.com/en/docs/middleware/dependency-injection/)
- [Gin binding](https://gin-gonic.com/en/docs/binding/)
- [Gin security guide](https://gin-gonic.com/en/docs/middleware/security-guide/)
- [Gin graceful shutdown](https://gin-gonic.com/en/docs/server-config/graceful-restart-or-stop/)
