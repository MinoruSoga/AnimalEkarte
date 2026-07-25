# AnimalEkarte — TODO

> 更新: 2026-07-24（BE9 implementation完了後reconciliation）

## 運用

- 本書は、エージェントが直ちに着手できる未完了タスクの台帳とする。
- タスクは「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で追加する。
- 対応済みセクションは削除し、完了記録はgit履歴と各実装testを正本とする。
- GitHub Issueと対応するタスクはIssueのstateを実測し、Issue一覧を本書へ重複掲出しない。
- release/運用gateは実装タスクと混在させず、[`q&a.html` OPS-13〜17](q&a.html#ops)と該当runbookで追跡する。

## 正本の境界

| 内容 | 正本 |
|------|------|
| 着手可能な実装タスク | 本書の「個別タスク詳細」 |
| GitHub Issueのstate・一覧 | GitHub Issues |
| BE9構造移行・進捗・release gate | BE9は2026-07-24にcode complete（release pending）。境界は[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) / [boundary map](docs/architecture/be9-2a-boundary-map.md)、release gateは[`q&a.html` OPS-13〜17](q&a.html#ops) |
| FEデザイン準拠・リファクタリング計画 | [`FE-refactor.md`](FE-refactor.md) |
| 今フェーズで着手しない事項 | [`phase2.html`](phase2.html) |
| 着手保留・任意検証のBE技術債 | [`BE-pending.md`](BE-pending.md) |
| PO判断・USER実操作・P0ブロッカー | [`q&a.html`](q&a.html) |
| Issueを3セッションで着手するためのガイドview（正本=各Issue・受け入れ条件を複製しない） | [`3-session-agent.html`](3-session-agent.html)（削除・退役の対象にしない） |

## 個別タスク詳細

### TASK-ADR003: 予約⇔会計の支払方法二重保持解消（ADR-003 案1B TRIGGER）

- PO-006 裁定済み。DEC-9（2026-07-25 q&a.html）で GitHub Issue 起票を待たず本書追跡へ変更。着手時期 = 納品後。
- 内容の正本 = q&a.html PO-006／DEC-9。USER が Issue 起票したら本エントリを Issue へ移設し二重掲出しない。

### TASK-251: 締め集計 category contract 確定実装（#251・8→12分類）

- 業務決裁確定（q&a.html DEC-21・**USER 本人裁定** 2026-07-25）。着手時期 = 納品後（S3）。contract 正本 = DEC-21（本エントリは実装スコープの入口であり決裁の「なぜ」は複製しない）。
- Phase 0 棚卸し（外部エージェント調査・Fable spot-verify）で確定した実装スコープ:
  - ① 正式カテゴリ = 12分類（enum 現状追認）。#251 タイトル「8分類」→「12分類」修正は Issue 本文転記（USER 承認後）に含める。
  - ② hospitalization 退院会計の other 固定を撤廃し CarePlanItem.Type／Procedure.IsSurgery→category resolver（`backend/internal/medicalrecord/hospitalization_service.go:431`）。treatment 経路（`backend/internal/billing/billing_item_service.go:405,462`）と共通化＝category contract 単一ソース化。
  - ③ vaccination を接種記録（`Vaccination.VaccineID`→Vaccine）から会計明細自動生成。`BillingItem` へ VaccineID provenance 列追加の migration が必要。自動化は停止／失敗通知／監査／idempotency（原則⑤）。
  - ④ hotel source=`HospitalizationTypeHotel`（②連動）、training は新規 source 設計。両カテゴリ維持。
  - 含意(a) category authority を BE resolver に一本化し FE/client は保持しない。
  - 含意(b) 締め集計の未知値 fail-closed = 生カラム無制限 GROUP BY（`backend/internal/billing/accounting_repository_reports_close.go:44`・`cash_register_service.go:265`）を12値 allowlist 経由にし typo/legacy を締め表へ黙って通さない（受け入れ条件「unknown/legacy を黙って変換しない」）。
  - 含意(d) 全書込経路（treatment/hospitalization/vaccination/trimming/merchandise/manual）を同一 typed category source に集約。
- #247（月次統合表）は本 TASK の contract 完了後に着手。
- Issue #251 本文への決裁転記（タイトル12分類修正含む）＝外部書込・USER 承認後にエージェント実施。転記後は本エントリを Issue 追跡へ移し二重掲出しない。
- 出典: #251 Phase 0 棚卸し Completion Report（2026-07-25・DEC-21）。

### BUG-429: 慢性疾患Listの親Pet相関漏れ（cross-tenant read防御ギャップ）

- CRITICAL(security)。`backend/internal/pet/chronic_condition_repository.go:33` の `FindByPetID` が子テーブル述語 `clinic_id = ? AND pet_id = ?` のみで、親petsへのclinic相関JOIN/EXISTSが無い。`pet_chronic_conditions.pet_id` は単一FK（clinic複合FKでない）ため、破損FK行（child.clinic_id=A × 他院pet_id）が存在すると他院の疾患名・診断日・notesを読める。過去のクロステナントread IDOR監査（13 repo修正）と同型で、現行app roleではRLSも遮断しない。
- 対応: 親pets相関（JOIN/EXISTS）追加＋破損FK fixtureによるDB-backed isolation test追加。#239実装前のsecurity blocker。
- 出典: #239調査 Completion Report（2026-07-25）。ソース実測で確認済み。

### BUG-430: stage-importの医院非限定DELETE

- CRITICAL。`backend/cmd/stage-import` のdeleteScopeが `owner_id >= 300000`（pets経由の継承含む）でclinic_id非限定。実行すると他院の高番ownerデータを削除し得る。`backend/cmd/stage-import/main_test.go:217-246` がこの挙動をテストで固定化している（cross-clinic保護テストは無い）。
- 対応方針変更（2026-07-25 DEC-20）: **stage-import退役で解消**（deleteScope修正には投資しない）。本番cutoverはrunbook既定の21表csv-import正式経路であり本ツールは本番使用禁止・local限定＋`--confirm-local-destroy`ガード既存。退役実装=cmd/stage-import削除またはビルド除外（#250再基準化転記とセット・USER承認後）。
- 出典: #251調査 Completion Report（2026-07-25）。テスト実測で確認済み。

### BUG-431: Reception危険度バッジのAPI契約不整合（サイレント機能不全）

- HIGH。`frontend/src/features/reception/api/transforms.ts:94` が `reservation.pet?.danger_level` を読むが、backendの `backend/internal/reservation/nested_summary_response.go` petSummaryResponseは当該キーを返さない。fixtureでは表示できるが実APIでは危険バッジが点灯しない。
- 対応: petSummaryResponseへ danger_level を追加（staff-internal限定・owner向け経路へ流さない）か、FE参照を削除するかを確定して是正。#234の前提是正。
- 出典: #234調査 Completion Report（2026-07-25）。両ファイルgrepで確認済み。

### BUG-432: 飼主生年月日がフォームから保存されない＋一覧列がpet値を表示

- HIGH。DB（`owners.birth_date`）・BE DTO・OpenAPI・DatePickerは実装済みだが、`frontend/src/features/owners/hooks/use-owner-form.ts` のcreate/update送信payloadにowner birthDateが含まれず、入力しても保存されない。さらに `OwnersListTable.tsx` の「生年月日」列はownerでなくpetの生年月日を表示している。
- 対応: payloadへbirth_date追加＋既存値を空へ戻す契約（JSON null vs 省略）の確定＋一覧列の正本（飼主DOB/ペットDOB）確定。#262の前提是正。
- 出典: #262調査 Completion Report（2026-07-25）。grepで整合確認済み。

2026-07-23に起票したBUG-421〜428、TEST-ROUTES-01、FMT-BE-01は2026-07-24のBE9実装でsource/testへ反映済みのため、本active listから削除した。release pending項目（fresh DB migration、remote CI/coverage、production deploy/ops rehearsal）は実装taskではないため本書へ再掲しない。
