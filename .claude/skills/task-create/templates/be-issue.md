# BE-XXX: イシュータイトル

**Status**: Open
**Priority**: High / Medium / Low
**Affects**: 影響するcontract・package・data・client
**Date Created**: YYYY-MM-DD
**Related**: TASK-XXX, FE-XXX

## Summary

1〜2行で問題と期待するobservable outcomeを説明する。

## Evidence

実際のcode、OpenAPI、schema、log、testを読み、推測ではなく`path:line`付きで現状を記載する。

```text
path/to/current_file.go:123 — 現在のbehaviorと問題
```

## Contract and safety boundaries

- Actor / authentication:
- Required permission / authorization:
- Resource ownership / clinic isolation:
- Request / response / status / stable error code:
- Cancellation / timeout / idempotency / concurrency:
- Compatibility impact:

該当しない項目は理由を記載する。OpenAPI変更がある場合は`backend/docs/api.yaml`を正本にする。

## Required changes

固定のRepository → Service → Handler順や固定pathを前提にしない。凝集性、利用者、依存方向、transaction boundaryに基づき、必要な変更だけを表へ記載する。

| Concern | File/package (evidence-based) | Change | Why this boundary |
|:---|:---|:---|:---|
| HTTP contract | `backend/docs/api.yaml` | request/response/status | client contract |
| Go package API | `backend/internal/<cohesive-package>/...` | type/function/route | consumer and cohesion |
| Persistence | existing query package or new cohesive package | query/transaction | data ownership |
| Schema | `backend/migrations/0NN_*.sql` | additive migration | integrity |
| Client | frontend path | generated/manual type usage | compatibility |

存在しないconcernのためだけにpackage、interface、pass-through layerを作らない。interfaceはconsumerが必要とする最小methodだけを定義する。

## Implementation notes

- Gin route group / middleware scope:
- Binding and validation:
- Context propagation and resource cleanup:
- Error mapping and logging:
- Transaction and tenant predicate:
- Production lifecycle impact:

特定helper名、DTO配置、layer名をGo/Gin公式要件として指定しない。

## Test plan

- HTTP: `httptest`で正常系、binding/validation、authn/authz、主要error
- Integration: query、transaction、constraint
- Security: unauthorized / cross-tenant / ownership
- Concurrency/cancellation/shutdown: 影響する場合
- Verification command: Dockerのscoped commandを具体的に記載

## Completion criteria

- [ ] observable contractとOpenAPIが一致する
- [ ] package/APIは凝集性・利用者・依存方向に基づく必要最小限
- [ ] Context・error chain・resource cleanupを維持する
- [ ] authentication・authorization・ownershipを分離する
- [ ] clinic/owner/pet/staff isolationをruntime testで確認する
- [ ] unknown error、secret、個人情報をresponse/logに出さない
- [ ] migrationは新規番号で追加し、適用済みfileを編集しない（該当時）
- [ ] scoped test/static checkが通る
- [ ] client compatibilityまたはmigration/rollout手順が明記される

## Out of scope

今回行わない変更と、その理由を記載する。folder再編や抽象化は、このissueのobservable outcomeに必要な場合だけ含める。
