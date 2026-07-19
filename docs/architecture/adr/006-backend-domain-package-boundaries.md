# ADR-006: backend domain package 境界と許可依存グラフ

**Status**: Accepted（2026-07-19、BE9-2B pilot完遂を受けてMinoruSogaが昇格判断。6論点のうち5件は未解決のままだが（1件は本ADR起票時に検証済みで決裁事項から除外済み）、各論点に決裁が必要な後続フェーズを紐付けたうえでのAccepted — 詳細は「未解決論点」節参照）
**Date**: 2026-07-19
**Relates to**: ADR-005（Go/Gin公式ベースライン採用）、ADR-002（clinic_id完全隔離）
**Deciders**: MinoruSoga（Accepted への昇格判断者）

## Context

ADR-005は「Handler → Service → Repository、Clean Architecture、layer-first/domain-firstをmandatory architectureにしない」ことと、「packageは凝集性・利用者・依存方向・変更単位で設計する」ことを決定した。しかし具体的にどのdomain package境界を採用するかは未確定のまま残されていた。

BE-refactor.md の BE9-2A（本ADRの起票元タスク）は、全 761 production Go file（`backend/internal/**/*.go` + `backend/cmd/**/*.go`、`_test.go`・`cmd/_archive` 除外）を codegraph（callers/callees/explore）+ grep/git log による再実測（8並列エージェント調査）を通じて、実際のRBAC resource名・route registration構造・GORM型のClinicIDフィールド・Goのreceiver型制約という一次証拠に基づき、target package境界の候補を導出した。

旧 BE-refactor.md §9（BE8-2、`internal/service`のみのgo/ast識別子参照実測、69ドメイン）と旧見積表（filename-prefixのみのカウント）は**正本として継承していない**——本ADRの判断は再実測（[boundary map doc](../be9-2a-boundary-map.md)）に基づく。詳細な再実測データ、per-domainの9列boundary map、§9との差分は同docを参照。**本ADRは決定のみを記録し、詳細data（全761 fileの分類・per-domain data・fan-in/out数値等）はboundary map docへのリンクに留める（二重管理禁止）**。

## Decision

### (a) domain-firstはproject decisionであり、Go/Gin公式の必須構成ではない

Go公式（[Organizing a Go module](https://go.dev/doc/modules/layout)、[Package names](https://go.dev/blog/package-names)）とGin公式（[API design patterns](https://gin-gonic.com/en/docs/routing/api-design/)）は、domain-firstかlayer-firstか、あるいはどちらでもない構成かを規定しない。本ADRが採用するdomain-first境界（下記13 target package）は、AnimalEkarteプロジェクト固有の設計判断であり、[go-gin-backend-guidelines.md](../../../.claude/rules/go-gin-backend-guidelines.md) §2「公式未規定」の範囲内で、ADR-005が示す4基準（凝集性・利用者・依存方向・変更単位）を根拠に確定した。

**採用するtarget package構成（13 target package + 既存cross-cutting packageの現状維持）**:

```text
backend/internal/
  owner/ pet/ staff/ auth/ reservation/ trimming/
  medicalrecord/ billing/ inventory/ lstep/
  clinic/ manualarticle/ httpapi/
  # 現状維持（既存の凝集cross-cutting package）:
  config/ dbconn/ middleware/ infra/ model/
  timeutil/ seedbundle/ logger/ csvimport/
  authjwt/ apperrors/ apicontract/
```

全761 fileの割り当て（`target package` / `現状維持` / `削除`）は [be9-2a-classification-manifest.csv](../be9-2a-classification-manifest.csv) に記録（未分類0件、削除0件——実測で明確な孤児fileが見つからなかったため削除を捏造していない）。per-domainの9列data（owned types/routes/queries、consumers、change frequency、fan-in/out、transaction有無、route有無、tenant boundary）は [boundary map doc §3](../be9-2a-boundary-map.md#3-per-domain-boundary-map9列) を参照。

### (b) 代替案

1. **layer-first維持（ADR-005以前の状態、現状）**: `internal/handler`(269 file)、`internal/service`(202 file)、`internal/repository`(154 file、42 subpackage含む)の3巨大flat packageを維持。却下理由: ADR-005が既にこの構成を"固定layer/層優先subpackage"としてsupersede済み。凝集性が低く（handler.go/service.go/repositories.goが108/95 fieldのcomposition root）、変更単位（1機能=3 layer横断の同時変更）とpackage境界が一致しない。
2. **repository層のみ現行42 subpackage粒度をhandler/serviceへも機械的に踏襲**: repository/checkup, repository/account等の42分割を単純にhandler/serviceへも同数複製。却下理由: 42は「エンティティ単位」の粒度でありRBAC resource・route構造から見た「業務ドメイン」単位（例: checkup/diagnosis/examination/vaccine/prescription等18サブトークンが全てmedicalrecordという1業務ドメインに属する）とは一致しない。42 packageに分割すると、本ADRが検出した10組の生cycleがそのまま42×42の組み合わせで再現され、収拾がつかなくなる。
3. **§9の69ドメイン粒度をそのまま採用**: 却下理由。§9は`internal/service`のみの実測であり、handler/repository/modelを含まない。また69ドメインは「抽出時のコンパイル安全性」の単位であって「業務境界」の単位ではない（§9自身が"抽出安全性の主軸はoutgoing"と明記、業務的凝集は別軸）。boundary map §6の差分表が示す通り、69ドメインの多くは実際には1つの業務ドメイン（medicalrecord等）に集約される。
4. **line/liff/lstepをドメイン境界で最初から3分割**: 却下（保留）理由。lstep調査エージェントは`internal/lstep`/`internal/line`/(reservationへ統合の)liffという3系統への分割を提案したが、これはBE9-2A時点の実測不足（sub-batch粒度の検証未了）を理由に、本ADRでは**単一target package(`internal/lstep`)として採用し、内部sub-batch構造はboundary map §8の参考情報に留める**——3分割は次ADRまたはBE9-2Dの実装知見を踏まえて再検討する。

### (c) cycle・security risk

実測で10組の生の双方向依存（Go import cycle相当）を検出した（当初7組として起票したが、santa dual-review round 1でarchitectレビュアーが追加3組を検出——lstepが所有する`LstepTagSyncService`interfaceがmedicalrecordだけでなくowner/pet/billingの計4ドメインからfieldとして直接保持されていたため。詳細は[boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式)の追記参照）。うち9組は本プロジェクト既存の規約（[go-gin-backend-guidelines.md §3](../../../.claude/rules/go-gin-backend-guidelines.md)「interfaceは利用側で定義」）に基づくconsumer-side interfaceで解消可能——利用側packageが必要最小限のinterfaceを自分のpackage内で宣言し、所有側packageの具象型がそれを構造的に満たし、DIルート（現`service.go`の後継）で配線する。この方式では利用側packageは所有側packageをGo importしない（コンパイル時cycle解消）。

| # | cycle | 解消方式 | 詳細 |
|---|---|---|---|
| 1 | staff↔auth | staffがAccountCreator相当のinterfaceを宣言 | [boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式) |
| 2 | owner↔billing | ownerがInsuranceExistenceChecker相当のinterfaceを宣言 | 同上 |
| 3 | medicalrecord↔billing | billingがMedicalRecordOwnershipChecker/DraftLocker相当のinterfaceを宣言 | 同上 |
| 4 | reservation↔trimming | reservationがmasterFKOwnershipChecker相当のinterfaceを宣言 | 同上 |
| 5 | clinic↔auth | clinicがPermissionGroupWriter相当のinterfaceを宣言 | 同上 |
| 6 | medicalrecord↔lstep | medicalrecordがTagSyncNotifier相当のinterfaceを宣言 | 同上 |
| 8 | owner↔lstep | ownerがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |
| 9 | pet↔lstep | petがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |
| 10 | billing↔lstep | billingがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |

**7組目（reservation↔staff、model.Staff/model.ShiftEntryへの共有テーブル二重書き込み）は interface逆転で解決しない真の依存cycleであり、本ADRでは意図的に未解決のまま記録する**（[boundary map §7.1](../be9-2a-boundary-map.md#71-reservationstaff-共有テーブル書き込み唯一の真のcycleinterface逆転で解決不可)）。これはtenant boundaryの変更ではなくpublic contract（staffの新規exportedメソッド経由に書き込みを統一する、またはreservation固有カラムを別テーブルへ切り出す）の変更が必要な設計判断であり、推測で解決策を選ばずユーザー決裁へ回す——本タスクのSafety boundary（2）に該当する。

**#8-10追加検出の影響評価（round1 architectレビュー）**: この見落としはDAG構造自体（13 package taxonomy、トポロジカル順序）を無効化しない——3組とも#6と同一の解消方式で解消し、lstepはトポロジカル順序で最上位のまま変わらない。Goはimport cycleを物理的に拒否するため、この見落としが実害化する経路は「BE9-2Cでowner/pet/billingへ着手した際のコンパイルエラー」であり、「サイレントなtenant isolation不具合」ではない——census網羅性の是正であり、構造的な誤りの訂正ではない。

**tenant boundary上のセキュリティリスク**として以下を明記する（ADR-002との整合確認）:
- `POST /api/line/webhook`（lstepドメイン）は`clinic_id`なしで受信し、署名検証段階で全クリニックのLINEチャネルシークレットを走査する意図的なcross-clinic-identity読み取りを行う。書き込みは各マッチ行自身の`ClinicID`を使用しclient入力を使わないため安全——ADR-002「no unscoped read」不変条件に対する**証拠に基づく例外**として本ADRで承認する（santa dual-review round 1でclinic-isolation-auditorが`line_link_service.go`の書き込みパスを直接読み、client入力を使わないことを再検証済み）。
- `BillingConfirmation`/`BillingItem`/`Payment`（billing）、`Inquiry`/`CarePlanItem`/`Treatment`/`ClinicalPlan`/`MedicalRecordImage`（medicalrecord）は自前`clinic_id`を持たず、親レコード（billing/medical_record/hospitalization）経由の間接isolation——親側のownership検証が壊れると連鎖的にisolationが崩れる構造的リスクとして明記する。BE9-2B以降のリファクタでこの間接isolationパターンを壊さないこと（当初`Inquiry`のみと記載していたが、santa dual-review round 1でclinic-isolation-auditorが同型の5モデルを、round 2で`BillingItem`/`Payment`を追加検出——詳細は[boundary map §7.3](../be9-2a-boundary-map.md#73-間接isolation-resourceの列挙が未網羅round1-santa-dual-reviewclinic-isolation-auditorで検出)）。`StaffReservationExclusion`も同型（自前clinic_idなし）だが、こちらは新規発見ではなく既に`preload_clinic_scope_lint_test.go`でsite-exception追跡済みの既知の低severity読み取りギャップ。`ExamResult`/`ExamTypeField`/`StaffNote`も検証済み——生きた漏洩なし（詳細はboundary map §7.3）。
- **`BillingItem`は間接isolationの中で唯一、repository層の防御が実質的にno-opであることが判明した**（[boundary map §7.4](../be9-2a-boundary-map.md#74-billing_item_repositorygo-のupdatedelete防御ギャップround2-santa-dual-reviewclinic-isolation-auditorで検出未修正)）——`billing_item_repository.go`のUpdate/DeleteがGORMの`Joins()`をUPDATE/DELETE SQLへ伝播しない罠に該当（Treatment等は正しく回避済み）。現状はservice層の事前check（`FindByID`）でgateされており**生きた漏洩ではない**が、defense-in-depthとして機能していない。**本ADR/BE9-2Aでは修正しない**（measurement/document onlyのscope外）——BE9-2Bでこのファイルへ触れる際にsubquery形式への是正とテスト追加を必須の前提条件とする。
- `AnimalSpecies`（pet）、`Company`（clinic、シングルトン）、`LstepAutoManagedPrefix`/`LstepConditionTagMapping`/`LstepSendPurposeTagPrefix`（lstep）はglobal-master——clinic-id-isolation lintがこれらを誤検知しないよう、BE9-1のlint再設計時に許可リストへ明記が必要。

master-FK-write lint（`backend/internal/service/master_fk_write_inventory_lint_test.go`）はservice write layerのreview-coverageゲートであり、上記global-master allowlist（clinic-id-isolation/preload lint向け）とは対象・目的が異なる別ゲートである。`internal/model`（GORMカラム定義）と`internal/handler`（HTTP request/response DTO）は別layer・別ゲートとして意図的にscope外とし、本lintの監査対象に含めない。role filterはBE9-2で確定するservice write logicを持つdomain packageに対して、単一の拡張ポイントのみを持つ。

### (d) 段階移行方法

`BE9-0(完了) → BE9-2A(本ADR) → BE9-1(lintのpackage非依存化) → BE9-2B(pilot 1件) → {BE9-2C↔BE9-2Dをlargest-ready方式で反復} → BE9-2E → BE9-2F → BE9-3 → BE9-4` の順序で固定する（BE-refactor.mdより）。

移行はstrangler pattern（一斉移動禁止）とし、各batchは以下を満たす：
1. target依存graphがacyclic（本ADRの許可依存グラフに従う、10組のcycleのうち9組は(c)の方式で当該domainへ着手する前に解消する。owner/pet/billingへ着手する場合は#8-10のlstep notifier interfaceも忘れずに解消すること——見落とすとコンパイル時import cycleで即座に失敗する）
2. tenant/認可/transaction/clinical safetyのbaseline testが存在
3. route/API/SQL互換性とrollback単位が定義済み

上記3条件を満たす**ready frontier**からproduction行数・file数・旧aggregator/call-site削減量が最大のdomainを選ぶ（largest-ready方式）。boundary map §8のlstep内部sub-batch案が示す通り、大規模domain内部もsink-first（out-dom=0のsub-batchから）で段階抽出する。

### (e) migration batchごとのbaseline test・移動対象・call-site・facade削除条件

以下はBE9-2C/2D着手時に各batchが固定すべき項目のテンプレート（本ADRの時点では値を確定しない——boundary map §5の許可依存グラフとトポロジカル順序が入力）：

- **トポロジカル順序**（本ADRの許可依存グラフをacyclicity検証した結果、[boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式)のedge解消後）: `httpapi → clinic → inventory → manualarticle → owner → pet → staff → auth → reservation → trimming → billing → medicalrecord → lstep`（13 node、45 edge、cycle 0件。round2 santa dual-reviewの指摘を受け、実際の検証コマンド`python3 be92a_toposort.py`とその出力をboundary map §5.0に明示——「機械検証済み」の根拠を可視化した）。
- **baseline test**: 移行対象domainの既存test（`*_test.go`）をそのまま維持し、移動後も同一test名で green を維持する。tenant isolation runtime test（clinic-scoped/global-master/cross-clinic-identity、boundary map §3の分類）は移行前後で挙動が変わらないことを追加確認する。
- **移動対象**: [classification manifest](../be9-2a-classification-manifest.csv)の`target:<domain>`行。ただし10組のcycleのうち解消方式(c)を適用するファイル（例: `reservation_validators.go`, `owner_service.go`, `clinicService.CreateClinic`等）は、interfaceの宣言・配線変更を伴うため通常のfile移動より慎重なレビューが必要。
- **call-site**: `internal/service/service.go`（108 field）、`internal/repository/repositories.go`（95 field）、`internal/handler/handler.go`（44 route registration呼び出し + 95 fileの共有receiver）が全call-siteの一覧——BE9-2Fで最終的に解体する。
- **compatibility facade / 削除条件**: 移行中は型alias/薄いdelegateとして許可し、call-site移行完了後に削除期限を設ける（BE-refactor.md BE9-2Cの既定方針）。`internal/repository/base.go`/`helpers.go`/`transactor.go`の非exportヘルパーは`repository/repohelpers`の同等export版が既に存在するため、対応するflat root packageの分割完了時点で削除する（boundary map §4.2）。

## 未解決論点（Accepted後も残る決裁事項 — 各項目は着手前ゲートとして後続フェーズへ紐付け済み）

BE9-2B（pilot=manualarticle、低結合domain）は以下6論点のいずれにも抵触しなかったため、これらを未解決のまま本ADRをAcceptedへ昇格した。各論点は「Open」のまま、該当フェーズ着手**前**の決裁ゲートとして残す——次のPRDや実装計画がこの一覧を読み飛ばさないよう、着手判定はこの表に対する具体的な合否で行う。

1. **Open (deferred to: BE9-2C着手直前、reservation/staffへの着手前)** — **reservation↔staffの共有テーブル二重書き込み**（[boundary map §7.1](../be9-2a-boundary-map.md#71-reservationstaff-共有テーブル書き込み唯一の真のcycleinterface逆転で解決不可)）— 2案（staffを唯一の書き込み者にする／reservation固有カラムを別テーブルへ切り出す）のいずれかを選択する必要がある。BE9-2Cでreservationまたはstaffへ着手する前に確定必須。
2. **Open (deferred to: BE9-2D着手直前、lstep内部分割の実装直前)** — **lstepのline/liff内部分割**（(b)代替案4）— 単一`internal/lstep`として当面採用するか、BE9-2A時点で3分割を先行決定するか。本ADRは前者（単一package、内部sub-batchは参考情報）を採用したが、実装知見が浅い段階の判断であり、BE9-2D着手前に再確認を推奨。
3. **Open (deferred to: BE9-2C、reservation/trimmingへの着手時)** — **clinic↔reservation/trimmingの想定依存が実測で確認できなかった**（[boundary map §7.2](../be9-2a-boundary-map.md#72-clinic--reservationtrimming-の消費関係が実測で確認できなかった)）— ClinicHoliday/ClosingSettingsの営業時間制約が予約可否判定へどう反映されるか（未実装の可能性）をreservation/trimmingドメインオーナーへ確認する必要がある。
4. **Open (deferred to: 低影響・随時 — inventory/medicalrecord/lstep/reservationいずれかへの着手時に併せて確定)** — **model/medicine.go・vaccine.goのinventory vs medicalrecord帰属**、**line_reservation_setting.goのlstep vs reservation帰属**（boundary map §4.4）— ファイル移動を伴わない現状維持のため実害は小さい。BE9-2B（manualarticle）はこれらのファイルに触れないため本pilotの前提条件ではなかった。
5. ~~ExamResult/ExamTypeField/StaffNoteの間接isolation保護パターン未検証~~ — **本タスク内で検証完了**（[boundary map §7.3](../be9-2a-boundary-map.md#73-間接isolation-resourceの列挙が未網羅round1-santa-dual-reviewclinic-isolation-auditorで検出)）。3件とも生きた漏洩なし（ExamResultはpre-check+tx方式、ExamTypeFieldは読み取り専用master、StaffNoteは追記専用でupdate/delete method自体が存在しない）。決裁事項からは除外。
6. **Open (deferred to: `billing_item_repository.go`へ最初に触れる時点 — BE9-2C/2Dのbilling domain着手時)** — **`billing_item_repository.go`のUpdate/Delete防御ギャップ**（[boundary map §7.4](../be9-2a-boundary-map.md#74-billing_item_repositorygo-のupdatedelete防御ギャップround2-santa-dual-reviewclinic-isolation-auditorで検出未修正)、santa dual-review round 2で検出）— 現状は生きた漏洩ではない（service層の事前checkでgate済み）が、repository層の防御が実質no-op。BE9-2Aのscope外（コード修正禁止）のため本ADRでは修正しない。BE9-2Bのpilot（manualarticle）はbilling_itemに一切触れないため本pilotの前提条件ではなかった。**次にこのファイルへ触れる者（BE9-2C/2Dでbilling domainへ着手する際）は、subquery形式への是正+クロステナント分離テスト追加を必須の前提条件とする**——このADRを読むエンジニアへの申し送り事項。

## Consequences

### Positive

- 全761 fileの帰属が実測（RBAC resource名・route構造・GORM ClinicIDフィールド・Goのreceiver型制約という一次証拠）に基づき機械的に検証可能な形で確定した。旧BE8/§9のfilename-prefixのみの分類の限界（lstep_tag_sync_care*.goのreceiver-vs-filename不一致等）を7ファイル単位で是正した。
- 10組の生cycleのうち9組は、新規abstractionを発明せず、プロジェクト既存のconsumer-side interface規約（ADR-005/go-gin-backend-guidelines.md §3）だけで解消可能であることを実証した——BE9-2B以降で「解決策が見つからず立ち往生する」リスクを大幅に下げた。
- medicalrecordドメインの実際の規模（185 file、旧見積96 fileの約2倍）が判明したことで、BE9-2C/2Dのbatchサイジング判断がより正確になった。

### Trade-offs

- reservation↔staffの共有テーブル問題は本ADRでは解決していない——未解決のまま次ユニットへ持ち越すため、reservation/staffドメインの実際のmigration着手（BE9-2C/2D）はこの決裁が下りるまで開始できない。
- lstepの内部3分割（lstep/line/liff）を保留したため、lstepドメイン（コードベース最大のfile数、106 file、かつ全コミットの21.4%を占める最活発domain）の実際のmigration設計は本ADRだけでは完結しない——追加のADRまたはBE9-2D設計フェーズでの再検討が必要。
- consumer-side interfaceによるcycle解消は、実装時に新規interfaceの命名・配置・DI配線という設計判断を都度必要とし、単純なfile移動より工数がかかる。

## References

- [BE9-2A boundary map（本ADRの実測データ正本）](../be9-2a-boundary-map.md)
- [BE9-2A classification manifest（全761 fileの分類）](../be9-2a-classification-manifest.csv)
- [ADR-005: Go/Gin公式ベースラインとpackage architecture](005-go-gin-backend-guidelines.md)
- [ADR-002: マルチテナント設計 — clinic_id完全隔離](002-multitenancy-clinic-id-isolation.md)
- [go-gin-backend-guidelines.md](../../../.claude/rules/go-gin-backend-guidelines.md)
- [BE-refactor.md BE9-2A](../../../BE-refactor.md#be9-2a-boundary-mapとadrを確定する)
