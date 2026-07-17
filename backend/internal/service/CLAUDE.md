# Service Layer — P1 / P8 / P10 / P11 / P13 / P17

## P1: FindByID before Delete/Update (MANDATORY)

```go
// ✅
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to find vaccine")
    }
    // ...
}

// ❌ Update/Delete without FindByID first
```

## P8: apperrors.Wrap on all error returns (MANDATORY)

```go
// ✅
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccine")
}

// ❌
if err != nil {
    return nil, err  // 未ラップ
}
```

## P10: FK dependency check before Delete (MANDATORY)

```go
// ✅ — check references, return 409 if in use
count, err := s.repo.CountUsageByVaccineID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to count usage")
}
if count > 0 {
    return apperrors.WrapConflict(fmt.Errorf("vaccine %d is in use", id))
}
```

**注意**: 末端エンティティ（他から FK 参照されない）は依存チェック不要。

## P11: slog.ErrorContext before repo error return (MANDATORY)

```go
// ✅
vaccines, err := s.repo.FindAll(ctx, clinicID)
if err != nil {
    slog.ErrorContext(ctx, "failed to find vaccines", "error", err)
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}

// ❌ ログなし
if err != nil {
    return nil, apperrors.Wrap(err, "failed to find vaccines")
}
```

**除外（ログ不要）**: `WrapInvalidInput` / NotFound 存在確認 / `WrapConflict`

## P13: Definition order in service file (MANDATORY)

```
1. const
2. buildFunc (buildXxxUpdateFields etc.)
3. interface
4. struct
5. constructor (NewXxxService)
6. methods
```

## P17: Input struct naming (MANDATORY)

```go
// ✅
type CreateVaccineInput struct { ... }
type UpdateVaccineInput struct { ... }

// ❌
type VaccineCreateRequest struct { ... }  // 順序逆
type CreateVaccineParams struct { ... }   // Params は違反
```

## カルテ子エンティティ書込（MANDATORY）

カルテ（`medical_record`）の子エンティティ（treatment / examination / vital / prescription /
checkup_field_result）への書込は、確定(finalize)と子エンティティ書込が競合するレースを防ぐため、
以下の不変条件を必ず守る（BE-refactor.md X-11 由来）:

1. tx 内で `medicalRecordRepo.LockByIDForUpdate(ctx, clinicID, medicalRecordID)`（BE-refactor.md
   R31 で `LockDraftByID` からリネーム）を呼び、返却された `record.Status` を確認してから
   finalized チェックを行う。**名前に反して status 不問で行ロックする** — finalized 判定は
   呼び出し側（service）の責務であり、ロック取得自体は draft 限定ではない。
2. 子 repo（treatment/examination/vital/prescription/checkup_field_result）の Create/Update は
   `dbOrTx(ctx, r.db)` で ambient tx に参加させる。参加させないと、`LockByIDForUpdate` の
   `FOR UPDATE` 行ロックと子テーブルの `medical_record_id` FK チェックがデッドロックする。

既存 8 サービス（`treatment_service.go` / `examination_service.go` / `vital_service.go` /
`prescription_service.go` / `checkup_field_result_service.go` / `medical_record_image_service.go` /
`estimate_service.go` / `billing_confirmation_service.go`）が先例。検証は
`medical_record_finalize_lock_concurrency_test.go`（`LockByIDForUpdate` の行ロック自体の並行性）
と、個別 repo の tx atomicity test（`examination_repository_tx_atomicity_test.go` /
`checkup_field_result_tx_atomicity_test.go` 等）が担う。

## 例外: Transactor を注入できないサービスの簡易ガード（`clinical_plan_service.go` / `inquiry_repository.go`）

`clinical_plan_service.go`（所見・診断）と `inquiry_service.go`（問診）も medical_record_id に
紐付く「カルテ子エンティティ書込」だが、上記の `LockByIDForUpdate` + tx + `dbOrTx` パターンは
**採用していない**。理由: 両サービスのコンストラクタは `cross_tenant_master_fk_write_test.go`
（他エージェント作業中ファイルのため変更不可）から `repository.Transactor` 無しの旧シグネチャで
複数箇所呼ばれており、かつそのテストは成功パス（実際に repo.Update を呼ぶ）を含むため、
`Transactor` を新規必須引数として追加できない。

代わりに以下の軽量パターンを使う:
- **clinical_plan**: service 層で `medRec.FindByID`（ロック無し）による事前ステータス確認を行い、
  友好的な Conflict メッセージを返す。レースの最終防衛は `clinical_plan_repository.go` の
  `Update`/`Delete` が `medical_records.status = 'draft'` を WHERE に含めることで原子的に担う
  （0 行 → `existsInClinic` で NotFound/Conflict を再判定）。
- **inquiry**: サービス層は無変更。ガード全体が `inquiry_repository.go` の
  `SaveByMedicalRecordID` 内（事前ステータス確認 + `Updates` 自体への `status='draft'` WHERE）に
  閉じている。

行ロックを取らないため、ロックベースの7サービスと比べてごく短い理論上のレース窓が残る
（サービス層の事前チェックと repo の書込の間に確定が割り込む極小ケース）。ただし書込自体は
repo 層の atomic WHERE で必ず拒否されるため、確定済みカルテへのデータ混入は発生しない。

## パッケージ分割規約（BE8・2026-07-17 — 正本 = /.claude/skills/be8-package-refactor/SKILL.md §3）

- **新規ドメイン service はフラット直下に置かず `service/<domain>/` サブパッケージで作る**。命名は repository と同規約（単数形・全小文字・stutter 禁止）。
- ドメイン間参照は **consumer 側の小文字ローカル interface** で受ける（先例: `reservation_service.go` の `reservationTypeFinder`）。import cycle は interface 抽出で解決する。
- 注意: service を走査する自作 lint は現存しない（安全網は repository 側のみ）。分割バッチは /.claude/skills/be8-package-refactor/SKILL.md BE8-5 の手順に従う。
