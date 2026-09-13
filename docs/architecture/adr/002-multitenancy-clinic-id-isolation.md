# ADR-002: マルチテナント設計 — clinic_id 完全隔離

**Status**: Accepted
**Date**: 2026-04-01
**Deciders**: MinoruSoga

## Context

複数クリニックのデータを単一 PostgreSQL インスタンスで管理する。
テナント間のデータ漏洩は医療情報保護の観点から絶対に許容されない。

## Decision

すべてのテナント分割データテーブルに `clinic_id` を持たせ、認証済みidentityから決定したclinic scopeを、すべてのread/write/delete pathへ適用する。packageやlayerの名称には依存しない。

以下は採用時の歴史的な擬似コードで、コピー可能な Go 実装ではない。現行の lookup は `backend/internal/owner/repository.go` を参照する。

```go
// clinicScope で WHERE clinic_id = ? を付与するという設計意図の省略例
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, ownerID uint) (*model.Owner, error) {
    return r.db.WithContext(ctx).Scopes(r.clinicScope(clinicID)).First(&model.Owner{}, ownerID)
}
```

GORM helperの使用有無だけで安全と判定しない。raw SQL、join、preload、count、bulk処理、background job、request由来FKのownershipを含め、全data pathをruntime isolation testで検証する。cross-tenant accessを可能にする変更はCRITICALとして拒否する。

## Consequences

**ポジティブ:**
- schema、query predicate、application ownership check、runtime testの多層防御で漏洩リスクを下げられる
- `internal/lintscan` の AST gate と runtime isolation test、重点 review の `clinic-isolation-auditor` で漏洩パターンを検出する。広範な healthcare review は補助であり、golangci-lint 単独を隔離の証明にしない

**ネガティブ:**
- clinic-scoped operationは認証済みclinic identityを明示的に受け渡す必要がある
- system_admin等の横断queryは通常scopeと別の明示的なauthorization・audit・testが必要になる

## References

- [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md)
- [Go/Gin Backend Review](../../../.claude/refs/go-gin-backend-review.md)

## 現行実装の補足（2026-09-06）

採用時の「すべてに `clinic_id`」はテナント隔離の設計原則であり、すべての GORM 型に直接 `ClinicID` field があるという inventory ではない。global master、account/session、親経由の子レコード、DDL trigger で clinic を複製する列を区別する。現在の DDL 分類は [erd.md](../erd.md)、request-time authority は [auth.md](../auth.md) を参照する。

現在の `ownerRepository.FindByID` は `findOwnerByID` / `persistence.DBOrTx` と clinic-scoped lookup を使う。

### 記録者と現在の担当者の区別（2026-09-13）

予約の `created_by` とカルテの `entered_by` は、登録操作を行ったスタッフの履歴上の識別子である。スタッフは複数医院に所属でき、システム管理者は医院所属行なしでも操作できるため、記録者の主所属医院は記録の所有医院を決めない。記録者を単独staff FKで保持する方式は、カルテの002 migrationと同じ扱いとする。

- **書き込み**: HTTPでは認証されたスタッフIDだけを採用し、JSONによる記録者指定を認めない。予約の通常・一括・管理画面登録では、保存と同じtransactionで有効なスタッフと操作医院への有効な所属、または有効なシステム管理者accountを検証し、認可に使った行をcommitまで共有ロックする。取込・修復等の直接書き込みも、この記録者の正当性を保持する。
- **読み取りの限定例外**: 親の予約・カルテを認証されたclinic scopeで取得できる場合、その記録者のID・氏名を現在の所属や管理者権限と独立した履歴情報として参照できる。予約の記録者preloadはID・氏名だけを取得する。現在所属を失ったスタッフも履歴として扱い、任意のstaffプロフィール、account、免許情報等への横断アクセスには拡張しない。削除済みスタッフは予約の登録者IDを残し、氏名preloadを返さない。
- **例外に含まれない関連**: 現在の担当医、飼主、ペット、予約区分、LINE顧客等のclinic所有関係には、引き続き親clinicとの相関と不整合行の拒否を要求する。

この区別は既存のPreload lintのStaff例外を無制限に広げるものではない。[予約の実DDL回帰テスト](../../../backend/internal/reservation/reservation_created_by_fk_test.go)で、記録者の偽装拒否・未所属者の書き込み拒否・最小投影・権限喪失後の履歴・別医院からの親取得拒否を検証する。
