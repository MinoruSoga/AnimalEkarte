# Persistence Packages

このdirectoryは未移行persistence実装と期限付きcompatibility facadeのmigration surfaceであり、新規production実装の追加先ではない。新規実装は[ADR-006](../../../docs/architecture/adr/006-backend-domain-package-boundaries.md)のtarget domain packageへ置く。ここを変更するのは、未移行実装の保守・安全修正、domain移動、または実consumerと削除phaseを持つcompatibility変更に限る。repository pattern自体はGo/Gin公式要件ではなく、詳細は[`BE-refactor.md`](../../../BE-refactor.md)に従う。

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
