# 共通operation基盤の継続静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-operation-core-20261023 |
| 種別 | 489限定静的監査の受領記録 |
| 決定状態 | 既存C06/Q契約を継承。新しい基盤や試験環境の選定なし |
| 実装状態 | PARTIAL。共通成果は実装済み、ContextGate共有拡張・consumer接続・全owner受入は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

受領監査の基準mainは `171b56575e7ee028e54358dad22ae28574853125`（13:28 UTC再取得）。489本文・全24コメントを読了した報告を受領。最新は `3607393e-fc9d-463c-b1c8-c9a8eba8267d`（11:45 UTC）。全ownerの全callerを監査済みとはしない。新規欠陥・新規票を起票せず、[索引の判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)と既存残件へ対応付ける。

<a id="operation-core-audit-adopted-subsets"></a>
## 採用済み成果とscope

| PR | 採用SHA | 限定成果 |
|---|---|---|
| [736](https://github.com/MinoruSoga/AnimalEkarte/pull/736) | `0014244103ce7ee06838879a55329d331687784f` | DB-only共通Runner、binding/result/attempt、現在認可、replay、unknown |
| [778](https://github.com/MinoruSoga/AnimalEkarte/pull/778) | `eb137881c013f9bd7b0335c01c62ede31429cb23` | ClaimMetadataWriter/短claim TX、manualarticle既存edit接続、HMAC config |
| [834](https://github.com/MinoruSoga/AnimalEkarte/pull/834) | `cf928c9d50cd51e1ac4b61b296653176e400d9be` | Q共通Request.VerifiedCommand/constructor/再検証。consumer接続なし |
| [744](https://github.com/MinoruSoga/AnimalEkarte/pull/744) | `57f80744ea25650a2deaff50cfd8a696b78c1a5b` | 最終001を使う試験入口の基盤。489意味論受入PASSの代わりではない |
| [822](https://github.com/MinoruSoga/AnimalEkarte/pull/822) / [824](https://github.com/MinoruSoga/AnimalEkarte/pull/824) | `abbe7bc5c86c39cf42a50577d15f0a4330114ca0` / `7839faa615031db613d86ef4f6971fe4d11008ab` | 医院時間/policy/anchor保存基盤。writer/489接続完成ではない |

<a id="operation-core-audit-ac-evidence"></a>
## 共通基盤ACと実装証拠

行番号は全て固定mainの[contract.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/contract.go)、[runner.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/runner.go)、[store.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/store.go)、[fingerprint.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/fingerprint.go)、[verified_command.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/verified_command.go)への参照。実行状態は全NOT_RUN。

| 接点 | 静的確認したsubset・限界 |
|---|---|
| 永久scope/同key | contract:35–83、store:55–104/114–123。clinic/owner/actor/intent/keyをDB検索で固定。fingerprint:105–155のHMACはowner/intent/Contentをcanonical化しCASは対象外。clinic/actorがHMAC bytesそのものに入るとは記載せず、DB scope束縛との組合せを区別 |
| ClaimMetadataWriter | contract:218–231、runner:112–125、store:43–108。短独立TXのINSERT DO NOTHING、RowsAffected==1のwinnerだけcallback、失敗は同TX rollback、loserは元binding読取。業務writer/成功receiptをclaim callbackへ持ち込まない |
| 業務TX | runner:150–153/321–408。owner Gate→binding NOWAIT→成功replay→unknown停止→Apply前savepoint→Apply→InsertResult→保存receipt再読。store:186–214はambient TX必須、成功はoperation_results行の存在で判定 |
| Lookup/現在認可 | runner:186–226/230–291。binding不在はscope-not-found。result読取不能でも元anchorの現在認可を先に検査し、失権はnot-authorized、認可済みはinternal error。記録済みunknownだけresult-unavailable。NOWAIT 55P03はcallbackからbusy error/rollbackでin-progress。conflictも元receipt/anchorで認可し本文非開示（128–145） |
| unknown | runner:352–390、store:260–309。Apply前savepointへ戻しbinding lock下でunknownを保存/commit。両記録失敗時には永続unknownが残らないことをrunner:438–460が明記。EffectsDatabaseOnly以外はclaim前拒否（103–105）。外部送信の安全な再実行機構ではない。旧rollback窓指摘を新規bugとして起票しない |
| Q | contract:69–84、verified_command:50–111、runner:492–508/352–368。constructorでcopy/projector全key値一致、claim前/Apply直前に再検証。Qはrequest memoryのみで共通durable payloadではない。projectorのinjectivity/purity、ApplyがQだけを読む責務はconsumerにも残る |
| GateとContextGate | contract:193–215にGate(ctx) error、runner:322に実呼出しあり。一方、派生contextを引き継ぐContextGate共有capability/interface/hookは固定mainに存在せず、検索hitはmodel所有注記/採択文書のみ。コメント `a6691204-3797-402b-9136-b7b3e32dad33` も未着手を示す。全gate未実装でも、ContextGate実装済みでDIだけ不足でもない |

<a id="operation-core-audit-production-consumers"></a>
## 実callerの限定確認

- production構築は[composition_runtime.go:751–792](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/api/composition_runtime.go#L751-L792)のmanualEditDeps。受領監査のproduction operation.NewRunner検索ではこれだけ。
- [manualarticle/handler.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/handler.go)の174 Execute/267 LookupがHTTP実呼出し。
- [participant.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/participant.go):160–168→[repository.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/manualarticle/repository.go):164–172がclaim hook。元anchorはparticipant:173–186→repository:174–192でbinding/clinic/owner/intentから解決。
- 元対象/現在grantはparticipant:191–215。Apply側の現在認可/lock/CAS後、286–294でContent commitment再計算、301–360で内容版/CAS/履歴/必須監査、371–381でreceipt返却。成功保存はRunner側の同TX。
- manualarticleは既存Content-only adapterで独自command digest照合あり。Q追加で一律移行済みとはしない。NewVerifiedCommand検索では共通定義/試験/docsだけでproduction consumer hitなし。検索範囲から全ownerの不在証明へ拡張しない。

<a id="operation-core-audit-dependencies"></a>
## 既存残件と受領先

| owner/単位 | 残る必要subset・実行境界 |
|---|---|
| 489 | ContextGate共有拡張、Q最初のconsumer引渡し/受入、001/RLS/deferred/Execute-Lookup並行/NOWAIT/rollback/cancel/unknown保存失敗の意味論検証。全owner完成を先行供給gateにしない |
| 448/492（D） | 共通医院時間保存側等の最初のQ consumer。#834は取得側/領域値/Q共通に限定。family registry/lookup allowlistと同TX audit/receiptは各所有者と分けて受領 |
| 483/493/458/496–497/503–504/516–517等 | 実adapter/認可/CAS/監査/receipt/route接続は各既存票。helper実在を参加済み証拠にしない。本監査は全owner網羅検査ではない |
| 491 D6 | 失権fixture/実RLS受入依存を保持。pet LifeStateReader/死亡gate受入を489基盤採用で代替しない |
| operation固有DB fixture | 下記001適合と実行証跡を分離。基盤をゼロから再実装しない |

<a id="operation-core-audit-db-fixture-boundary"></a>
## DB fixtureのsource観測と実行許可

[store_realdb_test.go:1–48](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/store_realdb_test.go#L1-L48)はoperationrealdb opt-in・未検証の記録を持ち、33–48はSetupTestDB/AutoMigrate double。[testdb/isolated_001.go:33](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/testdb/isolated_001.go#L33)のSetupIsolated001TestDBは実在するが当該store fixtureは使用していない。これは既存ヘルパーと接続先の観測であり、その隔離環境での実行を新たな必須gateにしない。fixture経路の最終001適合と、試験の実行許可・環境指定は別である。

#744の基盤go build/vet PASS作者報告後にも古いtestdb未compile記述が残る。最新#834/PlaneはAutoMigrate失敗と表現しており、今回現在のcompile失敗を再現したとはしない。残件はoperation固有fixtureの001統合・実行証跡未受領。本人の試験延期を維持し、後日の実行場所・方法を本記録で指定し直さない。

<a id="operation-core-audit-summary"></a>
## 受領要約

C06 Runner/ClaimMetadataWriterとQ共通subsetはmain採用済み。実接続を確認した共通Runner consumerはmanualarticle既存editで、Q production consumerは未接続。Gate(ctx)は実装/呼出済みだがContextGate共有interface/hookは未実装。短claimのbinding/anchorと、最外業務TXのwriter/監査/成功receiptを分ける。DB-only unknownの既修正を新規欠陥とせず、必要検証・consumer・各owner参加残を既存票に維持する。全owner調査完了・全AC PASS・Doneではない。

関連根拠：[489](https://app.plane.so/baritechllc/browse/EMR-489/)、[C06参加正本の固定基準](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/shared/operation-participation.md#L124-L157)。今回製品/DB/試験/生成/設定/Plane変更なし。
