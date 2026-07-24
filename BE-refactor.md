# BE-refactor — Go/Gin公式baseline上のdomain/capability移行

> **CURRENT (2026-07-24): BE9 CODE COMPLETE / RELEASE PENDING**。13 target package（owner / pet / staff / auth / reservation / trimming / medicalrecord / billing / inventory / lstep / clinic / manualarticle / httpapi）は全て現行domain packageへ移行済み。旧3layer（handler/service/repository）のproduction implementationとproduction import edgeは0件。
> **完了済みの計画・実行履歴は本docから削除済み**（BE9-0〜2F/3/4の各batch記録・Session A/B並行実行計画・BE8全節）。参照が必要な場合はgit履歴を見る — 削除前の最終版は `git show 327f4b332:BE-refactor.md`。
> **設計・分類の正本**: [ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md)（境界・許可依存グラフ・論点の解決記録）／[be9-2a-boundary-map.md](docs/architecture/be9-2a-boundary-map.md)／[be9-2a-classification-manifest.csv](docs/architecture/be9-2a-classification-manifest.csv)。workflow/SOT: [reservation-to-record-flow.md](docs/spec/reservation-to-record-flow.md)。安全不変条件: [backend application invariants](.claude/refs/backend-application-invariants.md) + ADR-002。

<a id="be9-parallel-sessions"></a>

> Session A/B並行実行計画はclosed（全domain移行完了・2026-07-24）。新規作業割当・開始gateとして使用しない。当時の契約・wave記録はgit履歴（上記）参照。

## 現行有効なbackend制約・知見

BE9移行で確立し、今後のbackend実装でも拘束力を持つもののみを残す。

- **domain packageのroute登録は単一エントリポイント必須**: `openapi_route_drift_test.go`の`buildFuncsFromDir`はbare名でfunc mapを構築するため、同名メソッドが複数struct上にあるとroute setがdrift検知から**静かに脱落**する。per-entity複数`RegisterRoutes`は禁止、`<domain>.Handler.RegisterRoutes` 1本に集約する。
- **新domain package追加時のlint追随**: `serviceWriteRolePackagePrefixes`への新domain prefix追加は初回のみ必要。`knownSafeParamQualifiers`への`"gin"`は追加済みで以後不要。qualifier包括allowlist化（package全型のsafe化）は恒久弱体化なので禁じ手。
- **sharedkernelへの新規追加基準**: literalな3コピー目を待たず、複数consumer・恒久的なdomain境界跨ぎ・acyclicな依存面を根拠に判断する。`internal/sharedkernel`は単一実装の正本。
- **cross-domain依存の解消パターン**: consumer-side interface（例: `AuditLogger`）+ `cmd/api`のadapter、またはmiddleware関数型注入。共有`common/util`逃げ・双方向importは禁止。
- **共有テストDBフレーク（現在も有効）**: `go test -p 1 ./internal/...`で対象外ファイルの赤は退行でない（pre-run再現を確認して続行）。**`-p 1`を外すと無関係packageで`deadlock detected`**（2026-07-21実測・並列DB競合の環境要因）。恒久対処=該当テストの`setupIsolatedTestDB`化は現在も未実施のopen follow-up。
- **response file移動時はscan dir追随を確認**: drift entryはbasename keyingのため、移動でscan対象から外れると「drift解消」の誤シグナルになる。`responseScanDirs`／`migratedDomainRoutePackages`へ移行先dirを追加する。
- **並行性証明テストはproduction実機構をpinする**: tx機構を変えたら、その機構を使う実DB並行性テストも同じbatchで追随させる。旧機構のまま残すと「productionが使っていない機構の証明」に劣化する。
- **安全許可集合マップの複製禁止**: 共有シンボルが「安全側の許可集合」（例: `eligibleMedicineUnitsForPerWeight`）を定義する場合、複製はドリフト＝安全性劣化源。定義側kernelごと帰属domainへ置き、消費者は修飾importで参照する。

<a id="be9-current-state"></a>

## 現在地とrelease gate（2026-07-24 最終reconciliation）

**BE9 implementation = CODE COMPLETE**:

- 13 target packageは全て移行済み。classification manifestのtarget 601 source rowは旧path現存0件で、現在のdomain packageにproduction code/testが存在する。
- 旧3layerのproduction implementationは0件。`internal/handler`はdirectoryごと削除され、`internal/service` 14 fileと`internal/repository` 50 fileは全て`_test.go`。production codeから旧3layerへのGo importは0件。
- `cmd/api`は22 production Go fileへ分割した明示composition rootであり、18 fileがtarget domainを直接importする。巨大な旧`Handler` / `Services` / `Repositories` aggregatorとcompatibility facadeは残らない。
- BUG-421〜428、TEST-ROUTES-01、FMT-BE-01はsource/testへ反映済みで、active task台帳から退役した。trimmingのclinic predicate/audit/transaction、lstep mapping replace、terminal status、checkup cap/abort、cancel cleanup、全domain route compositionを回帰testで固定した。
- migration manifestは761 rowを維持するimmutableなBE9-2A source-path snapshotとして扱う。内訳はtarget 601 row全て旧path削除、keep 160 rowのうち136 present / 24 consolidated or removed。移行後の物理file数との1:1一致は要求しない。

**2026-07-24 follow-up hardening（現行sourceへ同期）**:

- LINE webhookは、受信前にclinic identityを判別するため全`LineReservationSetting`のchannel secretを読む限定例外だけをunscopedとする。署名が一意に一致したclinic IDを以後のowner lookup/updateへ必須scopeとして渡し、異なるclinicで同じsecretが一致する曖昧系はfail closedにする。owner未登録のtyped NotFoundだけをno-opとし、真のlookup/update errorは成功へ畳み込まずnon-2xx retryへ伝播する。follow/unfollow更新は`clinic_id + owner id + expected line_user_id`のCASとし、LINE event timestamp（正数かつ受信時刻+5分以内）を保存値と比較する。followは既存follow/unfollowの両時刻より新しい場合だけ、unfollowは既存followと同時刻以上かつ既存unfollowより新しい場合だけ適用するため、stale・duplicate・out-of-order・再連携前IDのeventは`RowsAffected == 0`の安全なno-op、同時刻はunfollow優先となる。DB errorはnon-2xxへ伝播する。
- 公開LIFF account link成功時はowner情報を返さず`204 No Content`とし、PIIをresponseへ再露出しない。LINE ID token検証のoutbound HTTPはcallerのredirect policyを上書きしてredirectを追従せず、credentialを別originへ再送しない。
- billing confirmation/returnは`Content-Type: application/json`（charset parameter可）だけを受理し、欠落または他media typeはserviceを呼ばず415とする。bodyは8 KiB上限の単一strict JSON objectに限定し、exact lowercase property名とstring値だけを許可してcase variant・null・非string・unknown field・trailing JSON・oversizeを拒否する。trim後の`return_reason`はnon-blankかつ500文字、`memo`は1,000文字を上限とし、actor IDはrequest bodyではなく認証済みstaff contextから導出する。
- scheduler opsのCloudflare Access JWKSは固定team domainとtransport単位で10分cacheし、同時fetchを1本へ集約する。unknown `kid`またはupstream failure後のrefreshは60秒cooldownし、cooldown中はfail closedとする。鍵rotation直後は最大60秒の一時拒否を許容し、未検証JWTによるJWKS fetch増幅を防ぐ。これはWorker isolate内のdefense-in-depthであり、production Access policy・edge rate limit・運用rehearsalの代替にはしない。

**同一source local qualification（2026-07-24）**: global DB lease取得時のhost/container `backend/{cmd,internal}/**/*.go` SHA-256 manifestは`9a646e1a0180a8d77bb38b95ccc4cbaf79aa6c4032ab7c5ecfae16ec664690da`で一致した。leaseを`2026-07-24T10:01:06Z`から`10:03:58Z`まで単独取得し、LINE owner lookup/CAS/order/ambient-tx rollbackとlink token transaction、`internal/billing`全testおよび全race、旧`internal/repository`全testおよび全race、auth active-clinic authority、medicalrecord parent/child lock concurrencyを`-p 1`でPASSした。lease解放後はproduction codeを変更せず、secret scannerのtest-fixture誤検知を避けるため`staff/credential_audit_transaction_test.go`の架空文字列だけを変更し、当該testのraceを再実行した。この時点のhost/container全Go manifestは`516ed1fccf901829ff7e647540d38529d8d50924447c5621c17aa7594d583b10`で一致する。route composition、OpenAPI route/date/billing/LIFF contract、scoped vet、scheduler Node test 48件とworker typecheck、LIFF/billing frontend test 6件、docs symbol drift/self-test/diff-checkもPASSした。

**最終clinic相関follow-up（2026-07-24）**: 独立clinic-isolation監査で、複数医院`{A,B}`を認可された利用者に対し、clinic Aのmedical record / reservationへ破損FKで結ばれたclinic B関連をIN-scope Preloadできる問題を検出した。medical recordはOwner、Pet/Pet.Owner、Doctor、EnteredByStaffを親recordの`clinic_id`へ相関させ、Vitalsも親recordのclinicとpetへ相関させる。reservationはOwner、Pet/Pet.Owner、ReservationType/Group、Doctor、CreatedByStaff、LineCustomerを親appointmentの`clinic_id`へ相関させる。cross-clinic FKとOwner/Pet不一致を持つ破損親は一覧・単件ともNotFound相当でfail-closedにする一方、soft-delete済みの同一clinic関連と過去のstaff assignmentは履歴親行/countを消さず、現在の関連を表示するかはPreload側で判定する。最終worktree/container全Go manifestは`8be69a66dfe697602bc3aefb7a9799af2005f84afc6074a1b7bdd97df3a0333d`で一致し、commit candidate indexの全Go manifestは`8fe7fcaafcc8bcf2c5a959f1e646a56e16b95d823f0717363a486e503865eb0c`。両者の差分は別ownerの保護対象`cmd/csv-import/main_test.go`と未追跡`cmd/migrate/sql_migrations_integration_test.go`だけで、今回再検証した`internal/medicalrecord`・`internal/reservation`・旧`internal/repository`はworktreeとindexが一致する。global DB leaseを`2026-07-24T10:34:24Z`から`10:36:38Z`まで単独取得し、両domainの複数医院汚染・履歴保持・Vitals相関回帰testを`-race`でPASS、3 packageの全通常testをPASS、`internal/reservation`・旧`internal/repository`の全raceをPASSした。Go・clinic isolation・PostgreSQL/GORMの独立再reviewはいずれもApprove（CRITICAL/HIGH/MEDIUM 0）。再配属後も過去カルテにstaff nameを表示する既存の明示的LOWは継承する。これはlocal code gateの証跡であり、fresh DB migration、remote CI/full coverage、外部設定、production rehearsalを完了したことは意味しない。

**Known LOW residual（LINE redelivery）**: ownerを同じLINE User IDへ再紐付けした直後に`line_followed_at` / `line_blocked_at`のwatermarkが両方nilの場合、そのIDに対する非常に古い正規署名済みeventのredeliveryはtimestamp CASの初期比較を通り得る。expected LINE user ID CASにより別ID・別clinicへの波及はなく、現時点でこのfollow/block watermarkを業務判断に使うruntime decision consumerも存在しないためseverityはLOWとする。release rehearsalでは同一ID再紐付け直後の古いredeliveryを観測対象に含め、consumer追加や実害観測時にlink時watermark初期化またはevent ID/last-event persistenceを再評価する。

**RELEASE PENDING（code gapではない）**:

1. explicit approval下でfresh DBへ`002_lstep_snapshot_import_clinic_fk.sql`を実適用し、runner記録、checksum、rollback方針を確認する。
2. remote CIでfull coverage artifactとratchetを確定し、local scoped verificationをrelease evidenceへ昇格する。
3. production環境の実値（DB TLS、CORS/frontend URL、S3、SMTP、scheduler ops secret、alert webhook等）をsecret manager経由で設定し、production deployを実施する。
4. schedulerのpause/resume/status、manual run/catchup、Workers/Container logs、alert、失敗回復、rollbackをproduction相当環境でrehearsalする。
5. LINE webhook code deploy後に、LINE Developers Consoleで既定OFFのWebhook redeliveryとError statistics aggregationを明示的に有効化する。test channel/accountでcontrolled non-2xxからredeliveryを確認し、duplicate・out-of-order・同時刻unfollow優先・上記LOW residualをrehearsalする。Webhook errorsのnon-2xx、connection failure、timeoutを監視し、実施時刻・channel・結果・rollbackを記録する。

この更新ではfresh DB migration、remote push/CI、production deploy、LINE Developers Consoleを含む外部設定変更、ops rehearsalを実行していない。したがって最終表現は **BE9 code complete / release pending** であり、release readyではない。
