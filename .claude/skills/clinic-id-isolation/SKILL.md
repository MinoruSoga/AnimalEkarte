---
name: clinic-id-isolation
description: package配置に依存せず、database read/write・resource ownership・tenant scope変更をclinic_id越境の観点で監査するルータ。実チェックはclinic-isolation-auditorへ委譲する。
---

# Clinic ID Isolation Router

このskillはチェックロジックを重複実装しない。`.claude/refs/backend-application-invariants.md` を基準に、詳細監査を `clinic-isolation-auditor` へ委譲する。

## Trigger

backend の任意のpackage、migration、background jobで、次のいずれかを変更したときに発動する。

- SELECT / preload / join / count / exists / export
- create / update / upsert / delete / bulk operation
- request由来のclinic-scoped foreign keyやresource IDの永続化
- authentication identityからclinic scopeを決める処理
- transaction、audit、raw SQL、ORM scope

`internal/repository` / `internal/service` というdirectory名や、特定method名だけで発火を限定しない。

## Invariants summary

1. **Read**: base query、join、preload、countを含む全read pathで、認証済みclinicまたは明示的に認可された横断scopeを保証する。
2. **Write**: target rowとrequest由来のparent/master FKが同じ認証済みclinicに属することを、永続化前またはatomic predicateで保証する。
3. **Delete/bulk/background**: interactive requestと同じtenant/ownership条件を維持し、scope外resourceの存在を漏らさない。
4. **Verification**: helper名やcodeの見た目だけで合格にせず、cross-tenant runtime testで実際の拒否を証明する。

`clinicScope`、`FindByID`、GORM `Scopes` は現在利用できる実装手段であり、唯一の正解ではない。同等以上のtenant predicate、ownership check、schema constraint、testがあればpackage形状に依存せず評価する。

## Delegation

`clinic-isolation-auditor` を起動し、変更された全data pathを監査する。

```text
Task(subagent_type: clinic-isolation-auditor)
```

## Completion

- 全read/write/delete/background pathのtenant保証が確認されている。
- request由来FKのnested fieldまでownershipを確認している。
- 新規または変更したboundaryにcross-tenant testがある。
- auditorのApprove/Warning/Block結果が添付されている。
