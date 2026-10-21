# EMR-293 — スタッフ無効化監査の残件

初回調査日: 2026-10-08
決定追補日: 2026-10-21（EMR-366）
対象: `backend/internal/staff/credential_audit.go` の `writeAccountDeactivationAuditTx` と `backend/internal/staff/staff_repository.go` の `Delete`
結論: 監査イベント `staff.account.deactivate` は既にある。EMR-366の本人判断により、スタッフHTTP操作は操作した認証済み職員IDを記録し、この監査へのIP/UserAgent追加保存はしない。本書は設計・着手契約の更新であり、コード・設定・権限・実データを変更せず、Delete actorの実装ギャップ解消や試験完了を意味しない。今回の試験は **NOT_RUN**。

## 固定sourceによる現行観察（実装済み目標と区別）

以下は固定main `7a57b7bfebacdf212bc9555ac6622c0df8f3af99` の静的観察。2026-10-08の履歴記述と、2026-10-21の採用判断を分ける。

- [監査helper・password更新/account attachの既存監査](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/staff/credential_audit.go)
- [repository Delete・無効化経路](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/staff/staff_repository.go)
- [HTTP handler](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/staff/staff_handler.go)・[service Delete](https://github.com/MinoruSoga/AnimalEkarte/blob/7a57b7bfebacdf212bc9555ac6622c0df8f3af99/backend/internal/staff/staff_service_core.go)

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

## EMR-366 本人決定（2026-10-21）

初回調査では以下の3点を waiting decision として残した。2026-10-21 01:40 UTC、本人が「操作した職員を記録し、この監査にはIP・ブラウザ情報を追加保存しない」の確定と、rootの文書PR #774/#777/#780を必要な修正・チェック後にmainへマージしてよいかという2点の確認に「両方ok」と回答したため、今回の判断状態を次のとおり更新する。歴史時点の未決と現在の採用判断を混同しない。この承認に基づくEMR-366追記を含め、PR #780の範囲は本書とEMR-484・EMR-511の計3文書。マージはrootが行い、この文書整備の子threadでは行わない。

| 論点 | 固定sourceの観察／初回の判断状態 | 最新の本人決定 | 実装状態 |
|:---|:---|:---|:---|
| IP | helperはIPを受け取らず、IPAddressを設定しない。初回は判断待ち | スタッフ無効化・削除に伴うこの監査へのIP追加保存はしない | 現在の非保存を維持する契約。今回コード変更なし |
| UserAgent | helperはUserAgentを受け取らず、UserAgentを設定しない。初回は判断待ち | この監査へのUserAgent追加保存はしない | 現在の非保存を維持する契約。今回コード変更なし |
| Delete actor | Deleteはactor 0を渡し、ActorIDなしのsystemとなる。初回は判断待ち | スタッフHTTP操作は操作した認証済み職員IDを記録し、Deleteの0/system記録を解消対象にする | handler→service→auditへのtrusted actor伝播は今後の実装対象。未完了 |

既存のpassword更新・account attach等が保存するIP/UA監査は変更しない。本決定は `staff.account.deactivate` の監査境界であり、既存監査全体のIP/UA削除や設定・権限変更を認めるものではない。

## 採用済みの着手契約と残調査

- スタッフ無効化・削除のHTTP入口では、認証コンテキストから取得したtrusted authenticated staff IDをhandler→service→repositoryの無効化処理→auditへ渡し、操作した職員を記録する。clientのbody/query/headerで申告されたactorを信用しない。
- 必要なactorが欠落しているスタッフHTTP操作は、スタッフ・account等の変更前に既存の認証エラーで拒否する。actor 0/systemへ黙って補完せず、新しいエラー型や権限を本書から追加しない。
- 既存Update／UpdateForReservationを含む無効化入口とDeleteの呼出し全体を調査し、trusted actorの伝播・欠落時の拒否境界を確認してから実装する。現存コードにこの契約がすべて成立しているとは主張しない。
- 正当な既存background/system操作の有無と呼出し元は未調査事項として分ける。HTTP actor欠落と正当なsystem操作を同一視せず、調査前にsystem経路を削除・禁止しない。正当な経路がある場合は既存の意図と認可を確認して別扱いを明示する。
- スタッフ/accountの変更と必要監査を同じtransactionで成立させ、監査失敗時に変更を残さない既存境界を維持する。実装時には認証済みactorの帰属、actor欠落時のzero-write、監査失敗rollback、および既存password更新/account attach監査の保持を確認対象とする。

## 検証と完了状態

今回行ったのは固定source・文書差分・リンクの静的確認のみ。コード・API・設定・権限・実データ変更なし。実装、試験、Docker、DB適用、runtime/UATは **NOT_RUN**。旧資料の当時の127PASS等の結果を今回の検証として転用しない。

本人判断の未決3点は解消したが、Deleteのtrusted actor伝播、全無効化入口、正当なbackground/system操作の棚卸し、実装・試験は残依存であり、EMR-366の実装完了を意味しない。
