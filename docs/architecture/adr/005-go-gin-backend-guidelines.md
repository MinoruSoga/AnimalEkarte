# ADR-005: Go/Gin公式ベースラインとpackage architecture

**Status**: Accepted
**Date**: 2026-07-19
**Supersedes**: ADR-001 の backend package architecture 部分

## Context

従来の規約は Handler → Service → Repository、P1–P18、固定 directory、DI/log/error helper の配置を「Go/Ginベストプラクティス」として扱っていた。しかし Go 公式と Gin 公式は、これらの application architecture を規定していない。

公式由来の指針と project 固有判断が混在すると、不要な abstraction、package 増加、誤った compliance 指摘、実装構成への過度な依存が生じる。

## Decision

- Go/Gin 共通規約は [Go/Gin Backend Guidelines](../../../.claude/rules/go-gin-backend-guidelines.md) を正本とする。
- package は凝集性、利用者、依存方向、変更単位で設計する。
- `internal` の import 制約、package naming、consumer-side interface、Context 伝播等は Go 公式指針に従う。
- route group、closure/struct DI、binding/error middleware、security、graceful shutdown、`httptest` 等は Gin 公式指針に従う。
- Handler → Service → Repository、Clean Architecture、repository pattern、layer-first/domain-first を mandatory architecture にしない。
- P1–P18 を廃止し、公式ガイドと application safety invariant に基づく review へ置き換える。
- clinic/owner/pet/staff isolation、認可、監査、医療データ完全性は framework 非依存の [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md) と ADR-002 で維持する。

採用時、この ADR 自体は directory の一括移動を許可しなかった。その後 ADR-006 が domain migration を完了した。移行期間中だけ旧 directory と新しい境界が一時的に共存した。

## Consequences

### Positive

- 公式一次資料に根拠を追跡できる。
- folder/layer の形ではなく、package API、request lifecycle、security、observable behavior をレビューできる。
- 不要な interface や pass-through layer の増加を避けられる。
- architecture を変更しても tenant isolation 等の安全条件を維持できる。

### Trade-offs

- resource ごとに最適な package shape が異なり、固定 template より設計判断が必要になる。
- 移行中は旧 directory と新しい規約が一時的に共存した。ADR-006 完了後、旧3 layer directory は残っていない。
- project 固有 pattern が必要な場合は、公式要件と混ぜず ADR/invariant として根拠を記録する必要がある。

## References

- [Go: Organizing a Go module](https://go.dev/doc/modules/layout)
- [Go: Package names](https://go.dev/blog/package-names)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Gin API design](https://gin-gonic.com/en/docs/routing/api-design/)
- [Gin dependency injection](https://gin-gonic.com/en/docs/middleware/dependency-injection/)
- [Gin binding](https://gin-gonic.com/en/docs/binding/)
- [Gin security guide](https://gin-gonic.com/en/docs/middleware/security-guide/)
- [Gin graceful shutdown](https://gin-gonic.com/en/docs/server-config/graceful-restart-or-stop/)
