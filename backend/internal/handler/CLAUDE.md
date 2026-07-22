# Gin HTTP Boundary

このdirectoryは未移行HTTP実装と期限付きcompatibility facadeのmigration surfaceであり、新規production実装の追加先ではない。新規実装は[ADR-006](../../../docs/architecture/adr/006-backend-domain-package-boundaries.md)のtarget domain packageへ置く。ここを変更するのは、未移行実装の保守・安全修正、domain移動、または実consumerと削除phaseを持つcompatibility変更に限る。詳細は[Go/Gin Backend Guidelines](../../../.claude/rules/go-gin-backend-guidelines.md)と[`BE-refactor.md`](../../../BE-refactor.md)に従う。

## Review points

- route group で prefix と middleware scope を明示する。
- public/authenticated/authorized route の境界を登録時に確認する。
- body、query、URI、header に合う `ShouldBind*` を使い、error を必ず処理する。
- 入力の型、形式、長さ、範囲を検証した後、resource ownership を別途確認する。
- `c.Request.Context()` を request-scoped な下流処理へ渡す。
- response は公開 contract に必要な field だけを含め、OpenAPI と status code を一致させる。
- application error は一貫した HTTP contract に mapping し、未知 error の内部情報を出さない。
- dependency は closure または struct で型安全に注入し、global state を避ける。
- route/handler は `httptest` と最小 router で正常系・4xx・5xxを検証する。

特定の response helper、DTO ファイル名、変換関数名、下流 package の種類は Gin 公式要件ではない。

clinic/owner/pet/staff の認可とデータ分離は [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md) を必ず維持する。
