# EMR-293 — スタッフ無効化監査の残件

日付: 2026-10-08
対象: `backend/internal/staff/credential_audit.go` の `writeAccountDeactivationAuditTx` と `backend/internal/staff/staff_repository.go` の `Delete`
結論: 監査イベント `staff.account.deactivate` は既にある。このノートは新しい監査書き込みを追加しない。ギャップは未修正である。

## 現行

`writeAccountDeactivationAuditTx`（`credential_audit.go` 115–140 行）の引数は `tx`, `clinicID`, `staffID`, `accountID`, `actorStaffID` だけである。IP も UserAgent も取らない。

成功時に書く `model.AuditLog` は次だけを設定する。

- `ClinicID`
- `ActorType`（初期値は `model.AuditActorTypeSystem`）
- `Action` = `staff.account.deactivate`（定数 `auditActionStaffAccountDeactivate`）
- `Resource` = `model.AuditResourceAccount`（値 `"account"`）
- `ResourceID` = 無効化した account ID
- `NewValue` = `staff_id` と `is_active: false`

`actorStaffID != 0` のときだけ `ActorID` をその値にし、`ActorType` を `model.AuditActorTypeStaff` にする。`actorStaffID == 0` のときは `ActorID` を付けず、`ActorType` は system のままである。`IPAddress` と `UserAgent` は代入しない。

呼び出し:

- `staffRepository.Delete`（`staff_repository.go` 527 行）は `deactivateLinkedAccountTx(tx, clinicID, id, 0)` を呼ぶ。actor は 0 である。
- `staffRepository.Update`（275 行）と `UpdateForReservation`（729 行）は `deactivateLinkedAccountTx` に `cmd.ActorStaffID` を渡す。
- `deactivateLinkedAccountTx`（356 行）は同じ `actorStaffID` を `writeAccountDeactivationAuditTx` に渡す。

## 判断待ち

次の 3 点は waiting decision（判断待ち）である。ここでは決めない。

- IP: helper は IP を受け取らず、`AuditLog.IPAddress` を設定しない。判断待ち。
- UserAgent: helper は UserAgent を受け取らず、`AuditLog.UserAgent` を設定しない。判断待ち。
- Delete actor: `Delete` は actor `0` を渡す。その行は `ActorID` なしの system actor になる。Delete に実 actor を渡すかどうかは判断待ち。
