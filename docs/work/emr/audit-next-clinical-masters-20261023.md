# 診療マスタ・トリミングコースの継続静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-clinical-masters-20261023 |
| 種別 | 限定静的監査の受領記録 |
| 決定状態 | 既存採択仕様を参照。新たな業務・技術選定なし |
| 実装状態 | 以下のAC・接点だけを分類。全票・全ACの完了認定なし |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。基準mainは `171b56575e7ee028e54358dad22ae28574853125`。受領報告では454/455/457の最新本文と全コメント（23/15/15件、全ページ）を読取。以下はその固定範囲の記録であり、本編集で再度全件照会したという意味ではない。

<a id="clinical-master-audit-conclusion"></a>
## 結論と現状態

3票の監査時点の状態はUAT、担当者は未設定。この事実をUATの一般定義に置き換えず、状態・担当は変更しない。

- CRUD/CAS/auditと画像draft安全処理は既採用。未実装へ戻さない。
- 454/455はreceipt、source adoptionとstaff overrideの分離、固定snapshot受領が残る。455のDBOrTx 5件も固定mainで未登録。
- 457はA06旧2属性解除とcontract_version reader引渡しが未接続。480の最新T4削除guard残差は480へ維持する。description clearと画像cleanupは既実装を再利用する。
- 全実行NOT_RUN。UATをDoneへ昇格せず、本人の試験延期をPR作成の新gateへ追加しない。

<a id="clinical-master-audit-adopted-prs"></a>
## 採用PR

| PR | 確認したhead | 採用merge SHA |
|---|---|---|
| [735](https://github.com/MinoruSoga/AnimalEkarte/pull/735) | `db5484561a857e3968fa45c57ec12c4164c6cd68` | `768abc2068f6f303318e21ea4fbf14a5c2c0d3a7` |
| [730](https://github.com/MinoruSoga/AnimalEkarte/pull/730) | `f5304ccb241218f01b7a4d2335cf1e32b1c361b3` | `f66e8e53fa0555d1d9f2b0e8d53828d2a95b0810` |
| [731](https://github.com/MinoruSoga/AnimalEkarte/pull/731) | `f2eecbd0d014a4b80c68ee172e68fa5e19cdda11` | `a3717ccc0bd74f1ef02862a629ad285a521b3ae7` |
| [745](https://github.com/MinoruSoga/AnimalEkarte/pull/745) | `4ea3b5a3aad16efc4d54df331a49db6519eaa079` | `da27989fa812c1319b62a1c3b4684c419b110fe1` |
| [811](https://github.com/MinoruSoga/AnimalEkarte/pull/811) | `3a55ec750645c4aa639372b350621bba913be2d2` | `db07e72d87d9c2bd8f40fb815be9ffb5532e8285` |
| [828](https://github.com/MinoruSoga/AnimalEkarte/pull/828) | `7e8bb1263ea5f48bc03e5fd9faafbaca389816bd` | `05a6287e2cf92b983d421931a1243358a76eaded` |

全てmergedという限定成果の証拠。親票全体の完了や現mainでの実行受入には換算しない。

<a id="clinical-master-audit-insurance-evidence"></a>
## 454のAC別判定

全行の実行状態はNOT_RUN。AC prefixは `AC-accounting-item-insurance-`。根拠は全て上記固定main。

| AC suffix | 静的判定・限定証拠／残差 | 受領先 |
|---|---|---|
| four-masters | **implemented**。診察/処置の設定writerと共通save、455/456専有consumerは採用済み。実HTTP/UI往復の受入とは別。[consultation service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/consultation_service.go#L149)、[vaccine service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_service.go#L140)、[共通save](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/hooks/use-treatment-plan-master-saves.ts#L1) | 454＋455/456 |
| existing-med-exam | **unverified**。選択時bool反転とunknown拒否は接続済み。薬剤/検査の全writer・過去明示値保存の網羅的往復は未確認。[選択model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatments-tab-model.ts#L26)、[unbilled reader](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_item_repository_unbilled.go#L121) | 454 |
| not-guarantee | **missing**。選択payloadはis_insurance boolean。採用根拠版とスタッフ確認修正を分離した永続adapterの完了証拠なし。資格/補償保証へ昇格不可。[選択model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatments-tab-model.ts#L26)、[source reader](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_item_repository_unbilled.go#L22) | 454 adapter→481 typed source/505 snapshot |
| independent-discount | **dependency**。#828で保険と全体割引除外のwriter/監査は別fieldとして同TX採用済み。全区分snapshot・保険算定への受領は残る。[監査](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/master_insurance_audit.go#L37)、[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_request.go#L22) | 447/449 A01＋454→481/505 |
| authorized-save | **implemented**。clinic/actor、同TX audit、編集開始版CASの接続を確認。現在grant失効/他院/RLSの実行証拠は追加しない。[監査](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/master_insurance_audit.go#L37)、[repository](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_repository.go#L110)、[service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/consultation_service.go#L149) | 454/455＋466後段受入 |
| snapshot-stable | **dependency**。現master JOINとbool反転を過去snapshot/採用版永続化の完成に数えない。原値/確認修正/固定snapshot consumer受領が残る。[reader](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_item_repository_unbilled.go#L22)、[選択model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatments-tab-model.ts#L26) | 454→481/505 |
| explicit-reapply | **dependency**。対象行・元版/今回版・スタッフ確認値と履歴の明示再適用はconsumer受領scope。新規選択bool写像だけでは閉じない。[選択model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatments-tab-model.ts#L26)、[reader](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_item_repository_unbilled.go#L22) | 481/505、454 adapter |
| finalized-correction | **dependency**。理由付き訂正・取消再請求は会計ownerの有限成功単位。master保存へ重複実装しない。[業務正本](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/accounting/lifecycle.md#L501) | 505/462 |
| default-unresolved | **implemented**。新規4種true/明示値保存と更新省略保持は採用済み。AC IDを新規初期値判断待ちと読み替えない。[consultation](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/consultation_service.go#L149)、[vaccine](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_service.go#L140)、[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_request.go#L22) | 454/455/456 |
| legacy-missing | **implemented**。旧nullは更新省略で維持。未知master値は選択時に停止しbool補完しない。全経路runtime受入はNOT_RUN。[選択model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatments-tab-model.ts#L26)、[service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_service.go#L140)、[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_request.go#L22) | 454/455/456 |
| read-failure | **unverified**。shared save・canonical mappingとエラー停止は部分証拠。別clinic/ID遅延・dirty保護の全画面fixture実行/完全追跡は未確認。[共通save](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/hooks/use-treatment-plan-master-saves.ts#L1) | 454共通UI |
| atomic-retry | **missing**。CAS/auditは実装済みだがrequest/serviceにoperation_key/receipt照会接続なし。同結果再送/commit後応答断は未完。489 helper存在をconsumer完成にしない。[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_request.go#L22)、[監査](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/master_insurance_audit.go#L37)、[repository](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/vaccine_repository.go#L110) | 454継続（Q2案A）→489必要subset、455専有接続 |

<a id="clinical-master-audit-next-work"></a>
## 有限成功単位と455/457の限定接点

| 単位 | owner・必要subset | 完了条件／残条件 |
|---|---|---|
| M1 master保存receipt参加 | 454共通、455/456専有、489 helper | 同prepared commandからContent/Apply。同key同内容成功再送・CAS前提更新・別内容拒否・応答断lookup・現在認可をfixtureへ対応。原子的audit/CASを再利用し、489の必要契約/保存subsetだけを受領。全backendをgateにしない |
| M2 6種保険採用adapter | 454→481/505 | item kind/id/clinic、source/master版、explicit/legacy_missing/unavailable、品目値と確認修正を保存JSONで区別。I04/I05へ受領SHAを付与。481 typed binding/505固定snapshotの必要subset。483を計算ownerにしない |
| 455継続接点 | 455専有・454共通 | 既採用CRUD/CAS/auditを再利用し、receipt/adoption/override/snapshot接続とDBOrTx未登録5件を既存scopeへ保持。個別5メソッドや全ACの網羅表は今回受領資料にないため推測追加しない |
| 457継続接点 | 457、後続A06、480受領先 | A06旧2属性解除とcontract_version readerを未接続として保持。description clear/画像cleanupは再実装しない。最新T4 detail revision由来のcourse削除guard残差は480に保持し、旧予約履歴全件guardへ戻さない |

後2行は限定接点の受領記録であり、455/457全ACを列挙・検証した意味ではない。A06とT4の基準は[既存選定契約](emr-master-update-selected-contract-20261022.md)・[480契約](emr-480-booked-duration-selected-contract-20261021.md)を参照し、今回再選定しない。
