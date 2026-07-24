---
description: AnimalEkarte backendのセキュリティ・データ分離に関するアプリケーション固有不変条件
---

# Backend Application Invariants

この文書は Go/Gin 公式ベストプラクティスではない。AnimalEkarte が医療関連データを安全に扱うための、architecture の形に依存しない不変条件を定義する。

## Tenant and ownership isolation

- clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。
- client が送信した `clinic_id` を認可根拠にしない。認証済み identity から決定する。
- owner、pet、staff、medical record、billing 等の関連 ID は、同じ clinic に属することを server-side で確認する。
- not-found と cross-tenant resource の差で、他 clinic の存在を推測できる応答を返さない。
- bulk query、join、preload、count、export、background job にも同じ scope を適用する。
- transaction の内外で scope を失わない。raw SQL も例外にしない。

具体的な isolation 方式と監査証跡は [ADR-002](../../docs/architecture/adr/002-multitenancy-clinic-id-isolation.md) を正本とする。

## Authorization and sensitive data

- authentication、role/permission authorization、resource ownership をそれぞれ検証する。
- patient/owner/staff 情報、診療内容、credential、token、secret を error や log に出さない。
- response、export、event、audit log は最小限の field だけを含める。
- security boundary の変更は、unauthorized/cross-tenant test を必ず含める。

## Data integrity

- schema migration は project の migration 規約と ADR に従い、application 起動時の暗黙 `AutoMigrate` に依存しない。
- foreign key、unique constraint、transaction を application check の代替ではなく、追加の防御として使う。
- destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。
- business factごとにsource of truthとwrite ownerを1つにする。`appointments`とそのlifecycleは`reservation`、`staffs`と`shift_entries`は`staff`がwrite ownerであり、BE9-2E-0で収束済みの境界を[`appointment_write_owner_lint_test.go`](../../backend/internal/reservation/appointment_write_owner_lint_test.go)の自動gateで維持する。
- write owner以外のdomainはbusiness intentを表すconsumer-side interfaceまたは明示的orchestrationを通し、任意fieldを変更できるgeneric update APIや独立したpersistence writeを公開しない。
- appointmentに紐づく通常カルテは一般診療予約だけを対象とし、appointmentごとにactive recordを最大1件とする。カルテ日付は予約日時のJST日付から導出し、紐づいている間は独立変更させない。削除は対象カルテをlockしたtransaction内で見積依存を再確認してから`clinic_id + id + status=draft`の原子的条件付きsoft deleteを行う。見積Createも同じ親行を先にlockし、見積が先なら削除をConflict、削除が先なら後続見積を拒否する。確定との競合でも確定済みカルテを削除しない。予約検証、重複確認、transaction依存の欠落や失敗を成功扱いにしない。
- cross-domain writeはtransaction owner、全参加write、rollback範囲を明示し、部分成功でbusiness factを不整合にしない。意図的なsaga/best-effort処理は、補償、再試行、監査、部分失敗contractを持たせる。
- appointment、trimming detail、option等で1つのbusiness graphを構成するwriteは同じtransactionで全体を成功またはrollbackさせる。既存trimming appointmentのowner欠損はdetail-only writeでも補完する。参照master・担当者・LINE顧客を検証する必須依存が欠ける場合はwrite前にfail-closedとし、LIFFで明示指定されたstaffはclinic所属・対応可能種別・`is_active=true`・`reservation_visible=true`を満たすことをwrite transaction内で再検証する。write後の再取得が失敗し得る場合はcommit前に行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。
- fail-closedと定めたclinical/financial監査はbusiness writeと同じtransactionへ参加させ、監査dependency欠落または監査write失敗時はbusiness writeもrollbackする。締め後の会計編集はこの対象とする。
- `FOR UPDATE`、`FOR SHARE`、transaction-scoped advisory lockに依存するowner operationはambient transaction不在をfail-closedにする。request由来のclinic-scoped FKは同じtransactionで最終検証し、並行するmaster変更がinvariantを壊す場合は参照行をcommitまで固定する。
- nested `Preload`のpredicateは末尾associationだけに適用される。clinic-ownedの中間associationも独立したclinic predicateでscopeし、破損FKから他院の詳細・個人情報を復元しない。

package/data ownershipの正本は[ADR-006](../../docs/architecture/adr/006-backend-domain-package-boundaries.md)とする。

## Separation from official guidance

これらの不変条件を特定の `handler`、`service`、`repository` package の存在に結び付けない。package 構成が変わっても、すべての入口・use case・persistence path で検証できる形を維持する。
