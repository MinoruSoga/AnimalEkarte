---
description: Go/Gin公式ガイドラインとアプリケーション安全不変条件に基づくbackendレビュー手順
---

# Go/Gin Backend Review

Go/Gin一般規約の根拠は[go-gin-backend-guidelines.md](../rules/go-gin-backend-guidelines.md)、AnimalEkarte固有のpackage/data ownershipは[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md)を正本とする。この文書は両者を混同せず実行するreview checklistである。

## 1. Package boundary

- package は凝集しており、名前が短く明確か。
- `util`、`common`、`misc` のような曖昧な package を増やしていないか。
- export は必要最小限か。package 名と export 名が stutter していないか。
- interface は利用側の必要なメソッドだけか。mock のためだけに先行作成されていないか。
- import cycle を避けるための不自然な global state や巨大 package を作っていないか。

新規production codeはADR-006のtarget domain packageへ置く。ただしdomain内へ固定3層subpackageを機械的に作らず、既存codeもmigration planなしに一括移動しない。旧`internal/handler|service|repository`を変更する場合は、未移行実装の保守・安全修正または期限付きcompatibility変更であることを確認する。

## 2. Product/domain ownership（AnimalEkarte project decision）

- 変更は実在する業務能力・利用者・workflowに対応し、folderやinterfaceを先に作ることが目的化していないか。
- 旧`internal/handler`、`internal/service`、`internal/repository`へ新規production実装を追加していないか。残すfacade/adapterには実consumerと削除phaseがあるか。
- route、use case、transaction、persistence、testを同じdomainのvertical sliceとして追跡できるか。
- business factのsource of truthとwrite ownerが一意か。特に`appointments`とそのlifecycleのwriteが`reservation`へ収束しているか。
- write owner以外のdomainが対象tableへ独立して直接writeしていないか。cross-domain writeはbusiness intentを表すconsumer側の最小interface、明示的orchestration、transaction境界を通るか。owner外へ任意fieldを変更できるgeneric update APIを露出していないか。
- row/advisory lockを前提にするoperationがambient transaction不在を拒否するか。request由来master FKの最終検証と必要な共有ロックがwriteと同じtransactionに入り、検証後のmaster変更を許していないか。
- nested `Preload`でclinic-ownedの中間associationにもclinic predicateがあり、末尾masterだけをscopeして他院detailを復元していないか。
- compatibility facadeが薄いdelegate/type aliasに留まり、business ruleやpersistenceを重複実装していないか。削除条件があるか。
- 自動化に停止手段、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyがあるか。
- 効率化がclinical safety、clinic isolation、authorization、auditabilityを弱めていないか。

この節は[Product Philosophy](../../docs/product-philosophy.md)と[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md)に基づくproject固有checkであり、Go/Gin公式要件ではない。

## 3. Request lifecycle

- request-scoped 処理は `c.Request.Context()` を下流へ伝播しているか。
- Context を struct に保存していないか。
- timeout/cancel 関数を確実に呼んでいるか。
- goroutine が元の `*gin.Context`、request body、response writer を保持していないか。
- goroutine に終了条件と error/panic の処理経路があるか。

## 4. HTTP boundary

- body/query/URI/header に合う binder を使い、binding error を処理しているか。
- 型、形式、長さ、範囲、列挙値を境界で検証しているか。
- authentication、authorization、resource ownership を別々に確認しているか。
- response を一度だけ書き、API contract と OpenAPI が一致しているか。
- internal model や秘密情報をそのまま返していないか。

## 5. Error and logging

- error を無視せず、必要なら `%w` で chain を保持しているか。
- 既知 error が安定した status/code に変換されるか。
- 未知 error は一般化した 500 となり、内部詳細を漏らさないか。
- 同じ error を複数箇所で重複ログしていないか。
- log に request correlation 情報はあるか。secret、token、個人情報はないか。
- panic recovery を通常の error 処理に使っていないか。

## 6. Database and security

- query は parameterized されているか。
- DB/外部 API は Context のキャンセルと deadline を受け取るか。
- transaction の commit/rollback と error path が明確か。
- CORS、cookie、CSRF、trusted proxy、rate limit が deployment/auth 方式に合っているか。
- [backend-application-invariants.md](backend-application-invariants.md) の clinic/owner/pet/staff 境界をすべての読み書きで維持しているか。
- update/deleteのaffected rowsを確認し、存在しない対象やscope外対象を成功扱いしていないか。
- N+1、unbounded query、missing indexを実測とquery planに基づいて改善しているか（推測のみで最適化・放置していないか）。
- soft-deleteやhistory semanticsをschema/ADRに合わせ、暗黙条件に依存していないか。
- unit testだけでSQL semanticsを偽装せず、riskのあるquery/transaction/isolationを実DBのintegration testで確認しているか。

## 7. Server lifecycle

- production server に workload に合う timeout/limit があるか。
- SIGINT/SIGTERM から timeout 付き graceful shutdown が行われるか。
- `http.ErrServerClosed` を異常終了として扱っていないか。
- DB、worker、queue などの resource が安全な順序で閉じられるか。

## 8. Tests

- handler/middleware を `httptest` と最小 router で検証しているか。
- binding/validation/authentication/authorization/not-found/internal-error を含むか。
- dependency を差し替え可能で、global state に test が依存していないか。
- write ownerまたは状態遷移を変更した場合、owner外の直接write、許可されないtransition、cross-domain rollbackを検出するtestがあるか。
- 自動処理を変更した場合、停止、失敗通知、監査、手動fallback、重複実行またはretryのtestが変更riskを覆っているか。
- cancellation、concurrency、shutdown は変更 risk に応じて検証されているか。

## 判定時の禁止事項

- P1–P18 の旧番号を判定基準にしない。
- Handler → Service → Repository や Clean Architecture を Go/Gin 公式要件として強制しない。
- fixed directory tree、file size、interface 数、DI 場所を公式要件として指摘しない。
- GORM helper や project 独自 error helper を「Gin公式」と呼ばない。

指摘には、該当コード、実際の不具合または保守性リスク、根拠となる正本の節、最小の改善案を含める。
