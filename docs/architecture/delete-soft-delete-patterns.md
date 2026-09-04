# Hard Delete / Soft Delete policy

> **目的**: 削除方式を fact ごとに決め、FK、tenant、transaction、audit の境界を安全に保つ。
> **範囲**: 現行実装の policy と STG-001 の clinic deletion outcome。汎用 repository API や全 domain 共通の parent/child recipe は定義しない。

## 1. Fact-level policy

削除方式は「STG/test/master だから hard delete」のような環境名だけで決めない。model、migration、owner domain、保持要件、参照関係、監査 contract を確認する。

| Fact / model shape | 現行の基本動作 | 注意 |
|---|---|---|
| `gorm.DeletedAt` を持つ model | 通常の GORM `Delete` は soft delete | row は残る。親を後で hard delete すると、残存 FK が親削除を妨げることがある |
| `gorm.DeletedAt` を持たない model | GORM `Delete` は hard delete | clinical/financial fact を「soft-delete 非対応だから」だけで消してよいとは限らない |
| `gorm.DeletedAt` model の hard delete | `Unscoped().Delete` が必要 | cleanup が明示された owner-domain transaction 以外では使わない |
| append-only fact（例: `medical_record_addenda`） | delete API を作らない | 訂正は domain が定める追記・revision contract を使う |
| clinic deletion | clinic は hard delete。active dependency は block。eligible soft-deleted permission groups は同一 transaction で cleanup | concrete implementation は `clinicService.DeleteClinic` |

Soft delete 自体は FK 違反ではない。物理 row を残すため、その row の FK が将来の parent hard delete を block し得る、という関係である。

DDL と model が正本である。`ON DELETE CASCADE` を追加して一括削除へ逃げない。migration policy の CASCADE DELETE 禁止と tenant / audit boundary を守る。

## 2. Concrete outcome: clinic deletion (STG-001)

> **Historical incident label**: STG-001 で、soft-deleted `permission_groups` が clinic hard delete の FK blocker になる問題を修正した。以下はその concrete outcome であり、他の parent/child graph へ一般化しない。

`backend/internal/clinic/clinic_service.go` の `clinicService.DeleteClinic` は、1つの transaction 内で次を行う。

1. `LockByIDForUpdate` で対象 clinic を lock する。
2. owner、staff、その他の active blocking reference を再確認する。
3. active blocker があれば conflict として削除しない。
4. `PermissionGroupRepository.DeleteSoftDeletedByClinicID` で eligible な soft-deleted permission groups を hard delete する。
5. clinic を hard delete する。
6. 途中の失敗は cleanup と clinic delete をまとめて rollback する。

この経路の API semantics:

| 条件 | 結果 |
|---|---|
| route permission / system-admin check failure | 403 |
| 対象 clinic が存在しない | 404 |
| active owner / staff / domain dependency がある | 409 |
| soft-deleted permission group だけが残る | 同一 transaction で cleanup 後に clinic delete |
| cleanup / delete の内部失敗 | 5xx。成功扱いにしない |

`DeleteHardInTx` や `DeleteParentWithSoftDeletedChildren` という canonical method は存在しない。新しい generic helper 名としてコピーしない。

## 3. Transaction, tenant, and audit checklist

削除 path の owner domain は、次を同じ transaction / serialization boundary に置く。事前 check と write を別 transaction に分けない。

- 対象 parent/resource の lock、または同等の serialization
- active / soft-deleted dependency の再確認
- request 由来 FK と `clinic_id` ownership の最終確認
- 必要と決めた cleanup
- domain contract が必須とする audit
- 最終 delete

全 query に fact に合った clinic predicate を付ける。`Unscoped()` は soft-delete filter だけを外す。tenant scope や authorization を外すものとして使わない。

HTTP status は endpoint contract を確認する。

- route-level permission failure は 403。
- clinic-scoped lookup で row が存在しない、または別 clinic にしか存在しない場合は通常 404 として情報を隠す。`persistence.DeleteScopedByID` もこの形である。
- active dependency や state conflict は 409。
- すべての cross-clinic case を一律 403 としない。

Audit は path-dependent である。すべての CUD が自動的に `audit_logs` へ入るとは扱わない。clinical/financial/security integrity のため必須と定めた path は business write と同じ transaction で fail closed にする。clinic deletion と permission-group cleanup は現行コード上 audit dependency を持たないため、監査済みと記述しない。監査 coverage を追加する場合は code gap として owner domain で設計する。

## 4. Prohibited operations and review checks

### Prohibited

- migration に `ON DELETE CASCADE` を追加して tenant graph を無条件に消す
- 「staging」「test data」「master」という分類だけで hard delete を許可する
- dependency pre-check を transaction 外で行い、その後の delete まで競合可能な状態にする
- owner-domain API を迂回し、汎用 `Unscoped()` cleanup を呼ぶ
- `clinic_id` predicate なしの `Unscoped()` query/delete
- 実装にない repository method、audit action、audit payload を canonical example として書く
- audit write が必須の path で、business delete と別 transaction にする
- soft-deleted child は常に cleanup、active child は常に 409、child delete は常に parent existence check、という universal rule を作る

### Review checklist

- [ ] model / DDL 上の soft-delete support と FK action を確認した
- [ ] fact の write owner と保持要件を確認した
- [ ] lock、dependency recheck、cleanup、audit、delete が必要な transaction boundary に入る
- [ ] active blocker と eligible cleanup を endpoint ごとに定義した
- [ ] scoped absence / cross-clinic absence を通常 404 とし、route permission failure と区別した
- [ ] `Unscoped()` を使う場合も clinic predicate と明示理由がある
- [ ] audit は実在する action / resource と code path だけを記述した
- [ ] concurrency test、cross-clinic test、rollback test を risk に応じて追加した

## References

- `backend/internal/clinic/clinic_service.go` (`clinicService.DeleteClinic`, `ensureClinicCanBeDeleted`)
- `backend/internal/auth/permission_group_repository.go` (`DeleteSoftDeletedByClinicID`)
- `backend/internal/persistence/scope.go` (`DeleteScopedByID`)
- [`data-flow.md`](data-flow.md)（path-dependent audit / tenant semantics）
- [`backend/docs/api.yaml`](../../backend/docs/api.yaml)（API contract）
- [`ADR-002`](adr/002-multitenancy-clinic-id-isolation.md)（tenant isolation）
