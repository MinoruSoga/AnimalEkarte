---
name: go-gin-backend
description: "Go/Gin backend の設計・実装・レビュー。発火は backend の新規実装やレビュー時。固定 layer は強制しない。詳細は本文と go-gin-backend-guidelines。"
license: MIT
metadata:
  version: 2.0.0-animal-ekarte
---

# Go/Gin Backend

## When to use

- Go/Gin backend の package/API を設計するとき
- endpoint、middleware、database/external dependency を実装・reviewするとき
- package 分割、interface、DI、Context、error handling を判断するとき
- production lifecycle、security、HTTP test を確認するとき

## Required sources

最初に次を読む。

1. `.claude/rules/go-gin-backend-guidelines.md`
2. `.claude/refs/go-gin-backend-review.md`
3. tenant/ownership/security boundary を扱う場合は `.claude/refs/backend-application-invariants.md`

## Workflow

### 1. Define the observable contract

- route、method、request、response、status、error、authorization を確認する。
- OpenAPI と既存 client への互換性を確認する。
- cancellation、timeout、transaction、security boundary を列挙する。

### 2. Choose the smallest coherent package shape

- 既存 folder tree を template として複製しない。
- code の利用者、凝集性、dependency direction、変更単位を確認する。
- 小さい機能は同じ package でよい。独立した利用者や責務が生じたら分割する。
- interface は consumer が必要とする最小 method だけを定義する。
- implementation は concrete type を返すことを基本とする。

### 3. Compose Gin explicitly

- dependency が少なければ closure、多ければ struct handler を使う。
- `RouterGroup` で prefix と middleware scope を表現する。
- dependency を package global や untyped context value に固定しない。
- public/authenticated/authorized route の境界を route registration で確認する。

### 4. Protect the request boundary

- body/query/URI/header に合う `ShouldBind*` を選び、error を処理する。
- 型、形式、長さ、範囲、列挙値を検証する。
- authentication、authorization、resource ownership を別々に検証する。
- `c.Request.Context()` を DB/external API へ伝播する。
- public contract に必要な field だけを response に含める。

### 5. Handle failures and resources

- error chain を `%w` で保持する。
- known error を stable HTTP contract に mapping する。
- unknown error は内部情報を含まない 500 にする。
- 同じ error を重複ログしない。
- transaction、rows/body/file、cancel function をすべて cleanup する。

### 6. Verify risk

- handler/middleware は `httptest` と最小 router で test する。
- binding、validation、authn/authz、not-found、conflict、500 を含める。
- DB semantics、transaction、tenant isolation は integration test で確認する。
- concurrency、cancellation、goroutine、shutdown は影響時に test する。
- AnimalEkarte では Docker の scoped command を使う。

## Never claim as official

- Handler → Service → Repository / Clean Architecture
- repository/service interface の一律必須化
- DI は `main.go` のみ
- logging は service のみ
- DTO/model/validator の固定配置
- GORM helper や特定 error helper
- fixed directory depth、file count、coverage threshold

これらを採用する場合は project ADR/invariant として根拠を示す。

## Related skills

- server/middleware/lifecycle の詳細: `golang-gin-api`
- REST contract の詳細: `gin-api-design`
- Go test: `golang-testing`
- Go security: `go-security`
