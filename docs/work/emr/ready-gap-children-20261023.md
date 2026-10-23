# 既存実装の限定後続とReady整理の受領記録

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.ready-gap-children-20261023 |
| 種別 | 535再利用、553/554作成済みの事実と限定着手条件の受領記録 |
| 決定状態 | 公開済みReady整理を記録。候補を製品実装済みへ昇格しない |
| 実装状態 | 後続UI・inventory/fixtureは未着手。採用済みAPI/builder/基盤は再利用 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

基準mainは `77c28fbfbafa257dcca2421af99ff72da6889717`。[チケット規約](../../ops/plane-ticket-conventions.md)の必要subset・共有hunk・実行ID・1PR完了条件を適用する。Readyは計画/契約の成立であり、実装開始は別指示とclaim取得後。fixture記述と実行は別で、NOT_RUNだけをReady阻害にせず、未検証をPASSにも未生成コードをNOT_RUNにもしない。本記録は新しい実行許可・設定・権限を作らない。

<a id="ready-gap-public-facts"></a>
## 公開済みの整理

以下は親セッションから受領した投稿・読戻し済みの事実。535/553/554はいずれも個人未割当・実装未開始。本docs作業はPlane書込をしていない。

| 実行票 | 公開済みの事実 | 親・残scope |
|---|---|---|
| [535](https://app.plane.so/baritechllc/browse/EMR-535/) | **既存Readyを再利用**。今日新規作成していない。最新SHAと対象WIPを確認した限定観測をコメント `2c31d7f1-5d72-4de3-a238-73840b4d28cb` へ追補・読戻し済み | 528の未設定UI、親528 Reviewを維持。528直下に同じ別子を作らない |
| [553](https://app.plane.so/baritechllc/browse/EMR-553/) | Ready作成済み、ID `31eaf077-4491-4708-96b4-c6f9f0fcda95` | parent523。header戻る保護の単一owner/1PR |
| [554](https://app.plane.so/baritechllc/browse/EMR-554/) | Ready作成済み、ID `dae75766-cf11-42bf-9cd2-8f141ca6eb48`。個人未割当、実装未開始 | parent521。521の17件＋455の5件を同じ22件子で扱い、455側に別子を作らない |

親523/521のReview、455のUATは維持。子リンクの追補は523コメント `2744842a-1b76-4edc-82e1-28992c7f4938`、521 `c83c1f3c-ea2d-498e-a087-697c6e1909e9`、455 `7cf83314-b908-4ec7-a20d-2ec694b60205` へ投稿・読戻し済み。**relationは操作していない**。親の状態を一部成果だけでDoneにしない。

<a id="ready-gap-observation"></a>
## 競合観測と実行境界

2026-10-23の受領観測では、main77c28fb、2つのローカルrepositoryの計13worktree、open製品PR844/845を確認した。対象4scopeの専用PR・明示claim・対象ファイル未commit WIPはその観測範囲で見つからなかった。他環境や実行agentの不存在は保証せず、着手直前に再確認する。本docs用claimは製品実装scopeのclaimではない。

実装・試験/build/codegen・DB適用/起動/reset・Docker・ブラウザ・実鍵/deployは今回行わない。生成物を手編集せず、既存必須CIを弱めない。実機FAILを未確認のままbug扱いにしない。過去[846](https://github.com/MinoruSoga/AnimalEkarte/pull/846)は監査docs採用で、製品修正・受入PASSではない。

<a id="ready-gap-hospitalization"></a>
## 535：528の未設定UI

既存535本文・ACとコメント `8627a65b` の範囲を再利用する。必要API/builder/read/権限・dirty保持hookはmain採用済みで、533/#825全体を新しい着手gateにしない。[過去監査](audit-next-master-clears-20261023.md#master-clear-audit-optional-ui-plan)の既知UI後続を既存535の1PRへ引き渡す。

- **成功単位**：体格/料金単位の2 Select解除操作→既存空フォーム→Update null PATCH、Createの省略→per_day既定との区別、失敗時draft保持と操作fixtureまで。
- **owner・変更範囲**：535をclaimした1sessionがside panelと必要な専有model/fixtureを担当。shared hook/APIを再設計しない。新たな同内容子票を作らない。
- **受入**：small→未設定でbody_size:null、per_night→未設定でbilling_unit:null、片方だけ解除で他方維持、保存失敗保持、再読込未設定、新規未選択はキー省略、0価格維持。API三値/備考clearを作り直さない。
- **残件**：製品実装と操作fixture、後段の実機受入は未完。535 Readyは完了・試験PASSではない。着手前のclaim/PR/WIP再確認を要する。

<a id="ready-gap-header-back"></a>
## 553：523 P01 header戻る保護

出典は523コメント2286c5c6/9807d76f。目的は未保存編集をheader戻るで失わないこと。実装責務523/450 P01、親Reviewと採用#775/#820を維持する。

| 項目 | 限定着手条件 |
|---|---|
| 現状証拠 | [MasterPageShell.tsx:65](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/MasterPageShell.tsx#L65)はsettingsへ直接navigate。TreatmentPackageSettings.tsx:233–245は行切替/closeをdirty.runWithDiscardCheckで保護するが、555–561のshellへ戻るguardを渡さない。beforeunloadはSPA戻るの代替でない。実機再現NOT_RUN |
| 存在済みsubset | shell/route、useSidePeekDirtyの破棄確認、#820の失権/競合比較 |
| 同PRのscope | shellの任意onBack受入口＋package caller＋fixture。新API/BE/生成物なし、別owner供給待ちなし。新業務判断なし |
| 所有 | 553 claim者がshellのonBack型/伝搬hunkとpackage header handlerを単一writerで所有。prop未指定は既存navigate、他masterの挙動を変えない。450へP01対応を引き渡し重複実装しない |
| 手順 | dirty/保存中/create-unknown規律を読む→package headerを既存破棄確認へ結線→shell既定互換/route操作fixture→独立レビュー/引渡し |
| 受入 | cleanは従来のsettingsへ遷移。dirtyは確認取消でinput/target/dirty保持、承認時のみ遷移。保存中に処理/入力を失わず、create-unknownの結果不明・再送禁止を無断解除しない。失権/409比較後も無断破棄しない |
| fixture | TreatmentPackageSettings.test.tsxへ実header接続を検証。shell mockでguardを迂回してPASSにしない。必要ならMasterPageShell.test.tsxにprop指定/未指定互換。clean/dirty取消/承認/pending/unknown/失権/409保持を記述、実行NOT_RUN |
| 1PR完了条件 | 実header→package guard→確認→遷移を接続し、取消・異常保持とshell既定互換をfixtureへ対応。限定diff/最終head/独立レビュー/採用SHA/NOT_RUNを523/450へ引渡す |
| 対象外・残件 | 532 S06/S07/S08/S10の再修正、522/524、C06新設、全master一括guard化、親全AC Done。実機受入は523/450へ残す |
| 停止 | 同hunk claim競合は該当編集だけ停止。既存unknown規律との矛盾は具体経路を523へ返し、自動保存・盲目的POST再送・新回復方針を追加しない |

<a id="ready-gap-inventory"></a>
## 554：521の17件＋455の5件

出典は521コメント0321dede/681e3129と455コメントdbd0d13d/3140dd14。目的は既存DBOrTx参加メソッド22件のinventoryと原子性証拠を一つの限定基盤PRで閉じること。非原子的な製品保存を実行観測したとの指摘ではない。単一ownerは554をclaimするsessionで、個人は未割当。

現mainのdbortx_inventory_lint_test.go:259はpackage LockAllPackagesForReorderだけ、489–490はvaccine FindByID/Deleteだけ。既存fixtureの存在は全22件のrollback/read visibility証明ではない。

<a id="ready-gap-inventory-methods"></a>
### 対象22件

| repository | 登録・fixture対応対象 |
|---|---|
| medicalrecord/treatment_package_repository.go（15） | UpdateSortOrderBumpVersion, LockPackagesByIDAsc, IncrementCatalogRevision, SoftDeleteCAS, FindAll, LockClinicRow, MaxSortOrder, ReplaceItems, FindLiveItemsByPackageIDs, Create, GetCatalogRevision, FindAllItemsForPackage, LockCatalogRow, FindByID, UpdateFieldsCAS |
| clinic/clinic_repository.go（2） | CreateTreatmentPackageCatalogVersion, DeleteTreatmentPackageCatalogVersion |
| medicalrecord/vaccine_repository.go（5） | CountUsageByVaccineID, LockWriteRows, CountChildrenByParentID, Create, Update |

<a id="ready-gap-inventory-integration"></a>
### 必要subset・統合順

存在済みは#775/#820のpackage/catalog repository、#730/#828後のvaccine repository、persistence.DBOrTx/Transactorとtestdb。同PRは22件allowlist＋メソッド→同TX/rollback/read visibility/lock fixture対応表＋不足test code。新CAS/receipt業務契約は選定しない。

最終統合では[845](https://github.com/MinoruSoga/AnimalEkarte/pull/845)の固定head `23e944bd0ffad7c29b82b2b04e8771581deda2f4` にあるpet reader登録/説明（旧570付近）、reservation説明（旧1021付近）、ambientTxParticipationExpectations（旧1089付近）の3hunkを保持する。本子はpackage/clinic catalog/vaccineの既存sectionへの22件登録hunkを所有し、845の3hunkを編集しない。

準備・fixture記述は独立に進められる。最終採用前に845の固定採用差分を受領し、LifeStateReader entry/expectationと本22件の両方を静的確認する。845採用前なら最終統合だけ待ち、**845全機能のUAT/Doneはgateにしない**。最新diffが同hunkへ拡大した場合は該当編集だけ止めownerと調整する。親521 Review/455 UATを維持し、455の別子を作らない。532/#820の4修正、544の#828実行検証、549起動修復と別成果。

<a id="ready-gap-inventory-acceptance"></a>
### 変更file・手順・完了条件

再利用するfileはbackend/internal/lintscan/dbortx_inventory_lint_test.go、backend/internal/medicalrecord/treatment_package_repository_test.go、backend/internal/clinic/clinic_service_atomicity_test.go、backend/internal/medicalrecord/vaccine_repository_test.go。不足する実repository証拠の新設は同domainの *_repository_tx_atomicity_test.go に限定する。clinicのfake service testだけを実repositoryのTX参加証明にしない。production repository修正は計画外で、実欠陥が見つかればownerへ返す。

1. 22件それぞれのcaller/TX/lock順を固定する。
2. 既存fixtureの具体assertへ対応付け、不足fixtureを記述する。
3. 3domain hunkを一度だけ登録し、845差分保持を静的レビューする。
4. 521へ17件、455へ5件を分けて受領記録を残す。

受入期待値は以下。fixtureの準備と実行を区別し、今回は実行NOT_RUN。

- write群は変更/版/catalog/子rowが同TXに参加し、意図的失敗時に全変更rollback。
- reader/count群はambient未commit変更を読み、別医院を混ぜない。
- lock群はclinic→catalog→package→component、vaccine専用target UPDATE/parent確認の既存順を維持し、同TX lock参加/待機・解放をfixtureへ対応。
- Create/Update/Replace/Delete/CAS競合のzero-writeと既存NULL/履歴保持を壊さない。全22件に具体test名/期待値が必要で、allowlist追記だけでは完了しない。

1PR完了条件は22件inventory、既存証拠再利用/不足fixture、方法別coverage表、独立レビュー・最終head・採用SHA・NOT_RUN手順の自己完結。実DB PASS・全backend復旧は別で、対象外compile失敗を今回FAIL/PASSと推測しない。

DDL/model/api.yaml/生成、DB起動/適用、全27件修復、package apply/共通plan、vaccine receipt/会計adapter、実行検証は対象外。521実受入と455→454/489 receipt・481/505 adapterは既存ownerへ残す。544実行票へ22件を追加しない。fixture作成に未選定production変更が必要ならそのsubsetだけ止め521/455へ証拠を返す。relation作成は許可と両側記述を確認する別操作で、今回行わない。

<a id="ready-gap-evidence-limits"></a>
## 受領範囲と限界

既存子・重複検索と対象コメントの全件読取報告（528=11、523=8、521=11、455=16、535=1）を基準とし、535以外に同内容の限定子がないことを確認後、553/554を作成した報告を受領した。549は修復集約、544は828検証で対象が異なる。現在の全claim不在・全環境準備・製品受入を保証しない。

本記録は既存Ready再利用と作成済みの限定票への対応を公開するもの。全既存票の分割・状態変更・自動化実装・実試験は行っていない。本文・最新決定コメント・両票依存記述・Plane relationの同期は[規約](../../ops/plane-ticket-conventions.md)に沿う別作業として、未操作のrelationを完了扱いにしない。
