# Persistence Packages

このディレクトリは現行 code の database access を収めているが、repository pattern は Go/Gin公式要件ではない。persistence code の配置や abstraction は、利用者、transaction boundary、query の凝集性で判断する。

## Review points

- DB call に request Context を渡す。
- query は parameterized し、入力を連結して SQL を作らない。
- transaction の開始、commit、rollback と、tx handle の ownership を明確にする。
- not-found、conflict、constraint violation と未知 error を安定して区別する。
- query、join、preload、count、bulk operation、raw SQL のすべてで tenant/ownership scope を維持する。
- soft-delete や history semantics は schema/ADR に合わせ、暗黙条件に依存しない。
- update/delete の affected rows を確認し、存在しない対象や scope 外対象を成功扱いしない。
- N+1、unbounded query、missing index は実測と query plan に基づいて改善する。
- unit test だけで SQL semantics を偽装せず、risk のある query/transaction/isolation は実 DB の integration test で確認する。

repository interface、CRUD method 名、GORM helper、file layout は Go/Gin公式未規定である。

clinic isolation の詳細は [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md) と [ADR-002](../../../docs/architecture/adr/002-multitenancy-clinic-id-isolation.md) を正本とする。
