# 職員無効化・account連動・監査メタデータ限定静的監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-staff-disable-20261023 |
| 種別 | EMR-289 / EMR-366の限定静的監査受領記録 |
| 決定状態 | 既存の連動無効化・実職員actor契約を継承。追加の認可判断は行わない |
| 実装状態 | source接続と未接続・依存・未確認を以下で区別 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

基準mainは `77c28fbfbafa257dcca2421af99ff72da6889717`。受領監査の14:11 UTCのmain読取も同SHA。以下は監査報告の整形であり、文書編集時に全callerを独立再監査したものではない。実装という分類はsource接続の確認で、受入PASSではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-scope)を参照。

<a id="staff-disable-conclusions"></a>
## 結論

- 289の連動無効化・非対称再有効化・共有account拒否・同TX監査は現mainにある。再実装ではなく経路別受入証跡が残る。監査時の状態はUAT、個人assigneeなし。
- 366はReviewだが機能未完。PR780は採用判断のdocs-onlyで、Review化コメント `e78f59ba-da7f-4a76-a3ef-ad02082d8d3f` はそのmerge由来。本文の旧「Readyを維持」と現在状態を区別し、状態巻戻し・Done判定はしない。
- real actor必須、IP/UA新規保存なしは2026-10-21採用済み。再質問せず、admin Deleteのliteral 0、admin Updateのoptional actor、reservation Deleteのactor未取得を既存366の4 HTTP入口閉鎖へ渡す。
- 一医院の指名停止とglobal staff/account失効は別。493のassignment契約を289へ混入させない。旧reservation statusの認可差は意図済みとも新security欠陥確定とも断定せず、有限なcaller照合を残す。
- 新規重複起票0。旧PR・コメントのPASSを今回のruntime証跡へ転用しない。

<a id="staff-disable-adoption"></a>
## 根拠と履歴

[289](https://app.plane.so/baritechllc/browse/EMR-289/)の本文・全9コメントと[366](https://app.plane.so/baritechllc/browse/EMR-366/)の本文・全2コメントを再取得した報告を受領した（両next_page_results=false）。289の最新STG観測は2026-10-19のrate limit/budget_exceededによるBLOCKED、受入条件UNKNOWN。旧Backlog記載より監査時のUATを優先する。

| PR | 採用SHA | 責務 |
|---|---|---|
| [561](https://github.com/MinoruSoga/AnimalEkarte/pull/561) | `7ff2a7e272305c44438fc5fba3c010e328e934c7` | Update / UpdateForReservation / Delete / provisioning、global account_id、非対称再有効化の実装 |
| [780](https://github.com/MinoruSoga/AnimalEkarte/pull/780) | `455ec2f4d84c5e08da568bf99fc6d825506b499d` | 3文書の監査契約。runtime実装ではない |
| [551](https://github.com/MinoruSoga/AnimalEkarte/pull/551) | `1ed4b947dda0f3b95d20a5784612bc5bc7d91f20` | main→staging release。独立した本機能実装PRとして扱わない |

credential_audit.go:110がPR551のreview4178809086を参照するため、release/review起点と561の実装起点を分離する。551を履歴全体が誤り・無関係とはしない。[採用監査契約](emr-293-deactivation-audit.md)は全HTTP入口のtrusted actorと欠落時拒否を要求し、background/system caller調査を別に残す。[846](https://github.com/MinoruSoga/AnimalEkarte/pull/846)は10文書のdocs-only採用であり、本監査をそのbranchへ追加していない。

<a id="staff-disable-ac"></a>
## AC別の分類

26行は監査上の正規化で、Plane原文のAC数を主張しない。全行のruntimeはNOT_RUN。パス・行は上記固定SHAのもの。主要sourceは[staff repository](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/staff/staff_repository.go)、[staff service core](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/staff/staff_service_core.go)、[staff handler](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/staff/staff_handler.go)、[reservation handler](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/reservation/reservation_staff_handler.go)、[reservation service](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/reservation/reservation_staff_service.go)。

| ID / owner | 条件 | 静的区分 | 固定SHAの根拠・限界 |
|---|---|---|---|
| 289-01 / 289 | 通常PATCHの連動無効化 | 実装 | staff_repository:266–276,295–356。同TXで成功したaccount遷移のみ監査 |
| 289-02 / 289 | 予約status PATCHの連動無効化 | 実装 | reservation_staff_handler:132–161 / service:195–230 / staff_repository:717–730。必須trusted actorあり |
| 289-03 / 289 | 管理・予約Deleteの連動無効化 | 実装 | reservation_staff_service:182–192、staff_service_core:328–408、staff_repository:410–532。actorは366の別残件 |
| 289-04 / 289 | staff再有効化でaccountを復活させない | 実装 | staff_repository:268–270,722–724。account復活は別操作 |
| 289-05 / 289 | 別active staffの共有account拒否・全rollback | 実装 | staff_repository:328–338。別院の別staffも対象。並行競合NOT_RUN |
| 289-06 / 289 | 未紐付・既無効・削除accountのno-op | 実装 | staff_repository:295–325。account実遷移なしならdeactivate監査なし |
| 289-07 / 289 | 成功監査・監査失敗rollback | 実装 | staff_repository:341–356、credential_audit.go:115–141。ValidateLog/Create失敗を同TXへ返す |
| 289-08 / 289 | 自己無効化・共通最終adminガード（限定） | 実装 | staff_service_core:228–273、staff_service_update.go:60–68。管理Update/予約statusはself-deactivate、共通Deleteはlast-admin。管理Deleteのselfはhandler限定、予約Deleteは366-09 |
| 289-09 / 289 | 新規login拒否 | 実装 | backend/internal/auth/auth_service.go:93–136 |
| 289-10 / 289 | 既存session拒否 | 実装 | middleware/auth.go:80–94、auth/current_access_service.go:213–276、cmd/api/composition_auth.go:152–182。default current-access cache OFF。明示TTL時の窓・実配備値は未確認 |
| 289-11 / 289 | refresh拒否 | 実装 | auth/http_session_refresh.go:87–115 |
| 289-12 / 289 | provisioning初期無効account反映 | 実装 | staff_provisioning_apply.go:310–350、staff_provisioning_repository.go:264–290。false省略を明示UPDATEで回避、同ambient TX |
| 289-13 / 289 | 既存不整合データ修復 | 依存 | 289本文。機能の遡及修復外。別運用で環境・対象・承認を固定。DB未読取 |
| 366-01 / 366 | admin DeleteのA→B監査actor=A | 未接続 | staff_handler:248–265 / service_core:328,400 / repository:523–529。optional actorは自己guardのみ、service引数なし・repo literal 0 |
| 366-02 / 366 | 必要actor欠落時zero-write | 未接続 | staff_handler:201–203,257–261、reservation_staff_handler:115–128。admin Update optional、reservation Delete未取得。上流auth存在は必須handler境界の代替でない |
| 366-03 / 366 | body偽actor不採用 | 実装 | staff_handler:248–265、staff_request.go:60–86。Deleteはbody bindなし、Updateはtrusted contextからservice inputへ。欠落問題は別 |
| 366-04 / 366 | Update/statusの既存actor伝播維持 | 実装 | staff_handler:201–203、reservation_staff_handler:146–151、staff_repository:275,729。存在時伝播あり、欠落fail-closedは366-02 |
| 366-05 / 366 | IP/UA追加なし・他監査維持 | 実装 | credential_audit.go:115–141,90–100、staff_handler:204–211。2026-10-21採用済み |
| 366-06 / 366 | 現存route grant・同TX監査の保持（限定） | 実装 | staff/handler.go:212–218、reservation/routes.go:123–128、auth/http_permission.go:34–87,143–178。全認可閉包ではなく、global statusはcross-03 |
| 366-07 / 366 | real/missing/forged actor fixture | 未接続 | staff_handler_http_test.go:375–470、staff_account_deactivation_test.go:508–528はsystem/nil期待。変更・追加必要、実行NOT_RUN |
| 366-08 / 366 | 明示system/background保持・全caller棚卸し | 未確認 | 採用監査契約。HTTP4入口を分類しただけで全background/ops集合の完了は主張しない |
| 366-09 / 366 | 予約Deleteの自己削除拒否・trusted actor伝播 | 未接続 | reservation_staff_handler:115–128 / service:182–192 / staff_service_core:328–400。self guard/actor引数なし、共通last-adminだけではself識別不可 |
| cross-03 / 493・staff認可owner | global statusと全所属認可の適合 | 未確認 | reservation/routes.go:128、reservation_staff_service:195–230、493契約:7,32,63–65。selected clinic grantでglobal is_active更新はsource事実。admin全所属checkとの差を有限照合 |
| cross-01 / 493・289境界 | A院停止でB院/accountを失効させない | 依存 | master-staff.md:148–169、493契約:7,65。院別停止へglobal status APIを流用しない |
| cross-02 / 493・予約保護owner | 未来・現在診療中予約の解消前拒否 | 依存 | master-staff.md:114–120,165–180、staff_service_update.go:60–68。self/admin guardは予約保護でなく493/467接続残 |
| execution / 289・366 | 実試験・受入・配備確認 | NOT_RUN | 289コメントc880eebd-1d10-408e-aea5-7df1457dbf90の2026-10-19 STG BLOCKED。旧PASSを転用しない |

<a id="staff-disable-http-boundaries"></a>
## HTTP入口と認可境界

| 入口 | 現在grant・対象scope | actorの現状 |
|---|---|---|
| master PATCH | master-staff:edit。handler:185–200で医院別grant、service_update:25–37 / core:276–325で全active所属医院の編集権限（system_admin明示例外） | optionalStaffID、存在時伝播。無効化時の必須欠落拒否は366残 |
| master DELETE | master-staff:delete。所属・参照guard、複数所属拒否 | optionalをself guardのみへ使用、serviceへ未伝播、repo 0 |
| reservation status PATCH | routes:123–128のmaster-staff:edit、selected clinic。reservation repository:109–151でstaff＋assignment lock、staff_repository:717–730でprimary clinic scoped global更新 | ExtractStaffID必須:146–151。全所属編集権限との対称性はこのchainにない |
| reservation DELETE | master-staff:delete、selected clinic。service:182–192→staff Delete(false) | handler:115–128で取得なし、self guardなし、下流0 |

上流の通常authを通ることと、必須actor境界は別。欠落contextが本番で必ず到達可能とは主張しない。service署名変更前に正当な明示background/system callerを有限探索し、HTTPをsystemへfallbackさせない。

[現在grant評価](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/auth/http_permission.go#L25-L87)はrequest内memo。RequirePermission:143–178は他院grantだけでselected clinic writeを許さず、取得失敗・identity不足を拒否しsystem_adminは明示例外。これは毎request評価で、同TX commit直前の全grant再読・競合線形化の証明ではない。

[職員仕様](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/spec/screens/settings/master-staff.md#L52-L56)の1staff＋複数assignmentsと、旧移行由来の同account別staff行を区別する。共有accountの別active staffは医院を絞らず拒否し、inactive/deleted他行は拒否しない。一staffの複数所属についてadmin PATCHは全所属編集権限、Deleteは複数所属拒否で、共有account guardとは別。account更新はis_active=true条件付きでRowsAffected=0なら重複監査なし。再link/再有効化との全並行競合は未証明。

[493契約](emr-493-reservation-eligibility-contract-20261021.md)はA院assignmentだけを変更し、B院・staff/accountを保持する。旧capable/excluded切替と、global退職の認可を混同しない。493統合ownerが現FE status caller・表示意図を識別し、院別停止はeligibilityへ、global退職はstaff ownerの全院認可・保護へ対応を固定する。未確認差を新grantで救済しない。366はactorのownerであり院別停止全体を背負わない。

<a id="staff-disable-fixtures"></a>
## fixtureと未実行受入

既存staff_account_deactivation_test.go:74–243はcascade/reactivation/unlinked/ambient rollback、317–421は他院共有active拒否とinactive/deleted共有、423–505はdurable auditと失敗rollback。508–528のDelete system/nil actor期待は366達成証拠ではない。reservation_staff_account_deactivation_test.go:76は共有拒否、106はstatus actor。staff_handler_http_test.go:375–390はactorなしDelete、394–470は通常errors/clinic/selfで、新mandatory-actor fixtureではない。

有限な後続fixtureはすべてNOT_RUN：

1. 4 HTTP入口のA→B、actorなしzero-write、偽body actor不採用。正当system callerは別fixture。
2. edit/delete欠落・他院grantのみ・全所属不足・複数所属Delete拒否・全入口self/last-admin。
3. 他院の別active staff共有accountは409/all rollback。inactive/deleted共有許可と同staff複数assignmentの差。
4. audit失敗時staff/account/audit全rollback、実遷移1件、再有効化の非復活、unlinked/no-op。
5. login/session/refresh拒否、明示TTL時の失効窓。配備TTLと環境versionを固定して観測。
6. 493境界のA院pause/B院保持、global退職別入口、保護予約未解消拒否・手動解消後・過去完了だけの場合。289 cascadeだけで閉じない。

<a id="staff-disable-handoff"></a>
## owner引渡し・限界

- 366（個人未割当）：4入口actor差分、service/repo署名、real/missing/fake fixture。IP/UA再決定不要。background/system caller集合は未確認を保持。
- 289（UAT、個人未割当）：既存実装の経路別受入。今回試験禁止のため実行しない。古い隔離Docker計画は新たな実行許可でなく、全環境の既存不整合修復は別承認。
- 493統合・467予約保護と446/489の必要subset：医院別停止・退職・保護のcaller切替。全親Doneは不要だが未接続をactivation済みとしない。

current grantsのTX内競合、共有account再link、background/ops全caller、実配備TTL・既存データ完全性は未確認。静的未読と実行NOT_RUNを混同しない。2票と明示した交差境界だけの監査で、全職員ライフサイクル・全認可・全AC実行完了ではない。本docs作業からPlane投稿・状態変更は行わない。
