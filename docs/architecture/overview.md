# Architecture Overview

> **目的**: backend の設計原則と request lifecycle を説明する。
> Go/Gin公式は特定のapplication layerやfolder treeを規定しない。任意のdirectory例をarchitecture contractにはしないが、本projectが実測に基づいて採用したpackage/data ownership境界は[ADR-006](adr/006-backend-domain-package-boundaries.md)をcontractとする。

## System context

```text
Browser / external client
        |
        v
React frontend / API client
        |
     HTTPS JSON
        |
        v
Go net/http + Gin
        |
        +--> PostgreSQL
        +--> external services
        +--> background work
```

API contract は [`backend/docs/api.yaml`](../../backend/docs/api.yaml)、data isolation は [ADR-002](adr/002-multitenancy-clinic-id-isolation.md) を正本とする。

## Backend design baseline

backend は [Go/Gin Backend Guidelines](../../.claude/rules/go-gin-backend-guidelines.md) に従う。

- server 内部 code は必要に応じて Go の `internal` mechanism で保護する。
- executable が複数ある場合は `cmd/<command>` で entry point を整理できる。
- application package は凝集性、利用者、依存方向、変更単位で分ける。
- package 名や folder 名ではなく、package API と import dependency を境界として扱う。
- interface は一般に利用側で最小に定義し、implementation は concrete type を返す。
- dependency は closure または struct で型安全に注入し、global state を避ける。

Handler → Service → Repository、Clean Architecture、repository pattern、layer-first/domain-first は Go/Gin公式が定める architecture ではない。必要な設計判断は ADR に記録し、公式由来の規約と区別する。

## Product-driven package architecture

[Product Philosophy](../product-philosophy.md)はfolder treeではなく、要件を疑う、不要なものを削除する、単純化する、cycle timeを短くする、最後に自動化する、という判断順序を定める。このprojectはその順序をbackendへ適用するため、[ADR-006](adr/006-backend-domain-package-boundaries.md)で**domain/capability-firstのmodular monolith**を採用する。

- top-level packageは`internal/<domain>`を基本とし、route、use case、transaction、persistence、testを業務能力ごとのvertical sliceとして扱う。
- domain内の責務は別file/型に分けられるが、`handler`、`service`、`repository` subpackageを機械的には作らない。実際のconsumer、依存方向、変更周期が分かれた場合だけpackageを分離する。
- business factにはsource of truthとwrite ownerを1つだけ置く。`appointments`とそのlifecycleは`reservation`がwrite ownerであり、medicalrecord、trimming、billing、lstep等は独立したappointment write実装を持たない。この境界はBE9-2E-0で収束済みで、[`appointment_write_owner_lint_test.go`](../../backend/internal/reservation/appointment_write_owner_lint_test.go)のAST gateで回帰を検出する。
- cross-domain操作はbusiness intentを表すconsumer側の最小interfaceまたは明示的orchestrationを通す。owner外へ任意field更新APIを公開せず、複数domainにまたがるwriteはtransaction ownerとatomicityを明示する。
- migration facadeは薄いdelegate/type aliasに限定し、旧実装と新実装を二重のwrite pathとして残さない。
- 自動化は安全な手動pathを置き換えるのではなく同じuse caseを再利用し、停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを備える。

BE9の構造移行後、production実装は`internal/<domain>`へ収束した。旧`internal/handler`、`internal/service`、`internal/repository` directoryは**完全削除済み**（test-only residualも含め残存しない）。production codeから旧3 packageへのimport edgeはない。live mechanical lint gateは`backend/internal/lintscan/`に置く。

`cmd/api`は、domainごとのconstructorとroute registrationを直接合成するcomposition rootである。共通機能は責務に応じて`audit`、`persistence`、`scheduler`、`sharedkernel`、`infra/smtp`（package `smtp`）、`testdb`、`textsearch`等の凝集packageへ置き、巨大なlayer aggregateを復活させない。移行の最終証跡はgit履歴と[ADR-006](adr/006-backend-domain-package-boundaries.md)、release gate の作業入口は Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) と [`todo.md`](../../todo.md)（旧 OPS-13〜17 節は死リンク）。

この構成は「Clean Architectureのfolderを再現する」ことではない。ただし、依存方向、consumer-side interface、明示的DI、境界をまたぐtransactionといった原則は必要な箇所で選択的に使う。効率化よりclinical safetyとclinic isolationを優先する。

## Request lifecycle

1. `net/http` / Gin が request を受ける。
2. route group と middleware が recovery、observability、authentication、authorization、rate limit 等を適用する。
3. HTTP boundary が body/query/URI/header を型付き input に bind し、形式を検証する。
4. 認証済み identity から clinic scope を決め、resource ownership を検証する。
5. request Context を database と external service へ伝播する。
6. domain/application logic と persistence が、必要な transaction/invariant を維持して処理する。
7. error boundary が既知 error を stable HTTP contract に mapping し、未知 error を一般化する。
8. 公開 contract に必要な field だけを response として返す。

この sequence は責務を示すが、それぞれを別 package や別 layer にすることを要求しない。小規模 resource は1つの凝集 package にまとまり得る。分離は実際の複雑性と利用者が生じてから行う。

## Security boundaries

- authentication、authorization、resource ownership は独立した check とする。
- clinic-scoped data は read/write/delete、join/preload/count、bulk/background job の全 path で制約する。
- client supplied な clinic/owner/pet/staff ID を認可根拠にしない。
- secret、credential、個人情報、内部 error を response/log に出さない。
- HTTPS、CORS allowlist、CSRF、secure cookie、trusted proxy、rate limit を deployment に合わせる。

詳細は [Backend Application Invariants](../../.claude/refs/backend-application-invariants.md) と [auth.md](auth.md) を参照する。

## Error and observability

- error chain を保持し、処理できる境界まで返す。
- 同じ failure を複数箇所で重複ログしない。
- request ID/trace 等の correlation を Context と structured log で伝播する。
- unknown error は汎用 500 とし、診断情報は server-side に限定する。
- panic recovery は process crash 回避用であり、通常の error flow の代替にしない。

## Server lifecycle

- production server は `http.Server` で timeout/limit を明示する。
- SIGINT/SIGTERM から timeout 付き graceful shutdown を行う。
- 新規requestを止め、HTTP shutdown、background work drain、DB closeの順序を明示する。
- background workは受付gateとbounded concurrencyを持ち、shutdown開始後の追加受付を拒否する。
- goroutine は元の `*gin.Context` を保持せず、終了条件とcancel/error経路を持つ。
- Cloudflare scheduled eventは、通常APIと同じdomain use caseを呼ぶdurable coordinatorを経由し、重複実行防止、pause/resume、catch-up、履歴、失敗通知を備える。運用手順は[Scheduler Operations](../ops/deploy/runbooks/SCHEDULER_OPERATIONS.md)を参照する。

## Testing strategy

- domain HTTP boundaryとmiddlewareは`net/http/httptest`と最小routerでtestする。
- binding、validation、authentication、authorization、not-found、conflict、internal-error を確認する。
- query、transaction、tenant isolation は risk に応じて実 DB integration test を行う。
- cancellation、concurrency、shutdown は変更箇所に応じて検証する。
- package layout そのものではなく、observable behavior と security boundary を test する。
- route composition smoke、OpenAPI drift、package非依存のclinic/audit/transaction lintで、domain間の統合境界を検証する。

## Decision ownership

| Concern | Source of truth |
|:---|:---|
| Go/Gin general guidance | [go-gin-backend-guidelines.md](../../.claude/rules/go-gin-backend-guidelines.md) |
| Product and workflow principles | [product-philosophy.md](../product-philosophy.md) |
| Backend domain/package and write ownership | [ADR-006](adr/006-backend-domain-package-boundaries.md) |
| Appointment workflow and source of truth | [reservation-to-record-flow.md](../spec/reservation-to-record-flow.md) |
| API contract | [`backend/docs/api.yaml`](../../backend/docs/api.yaml) |
| Tenant isolation | [ADR-002](adr/002-multitenancy-clinic-id-isolation.md) |
| Authentication/authorization | [auth.md](auth.md) |
| Database schema | [erd.md](erd.md) and migrations |
| Architecture decisions | [adr/](adr/README.md) |
| 城東検査機器 受信/commit | [ADR-007](adr/007-lab-device-receive-and-commit.md) |
