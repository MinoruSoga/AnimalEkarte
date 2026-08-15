# Backend Coding Rules

この文書は backend 作業の入口である。Go/Gin の設計・実装規約の正本は [Go/Gin Backend Guidelines](../.claude/rules/go-gin-backend-guidelines.md)、レビュー手順は [Go/Gin Backend Review](../.claude/refs/go-gin-backend-review.md) とする。

## Authority

1. Go language/toolchain の仕様
2. Go 公式・Gin 公式ドキュメントに基づく正本
3. [Backend Application Invariants](../.claude/refs/backend-application-invariants.md) と ADR
4. API/OpenAPI、database schema、migration などの project contract
5. package 内の局所的な説明

Go/Gin公式が規定しない package layout や layer pattern を、公式要件として追加してはならない。

## Package design

- package は凝集した責務で分ける。現在のディレクトリ名を新規設計へ機械的に複製しない。
- package 名は短く明確な小文字の1単語を優先し、`util`、`common`、`misc` を避ける。
- package API は小さく保ち、exported name の stutter を避ける。
- interface は一般に利用側で、実際に必要な最小メソッドだけを定義する。
- implementation は concrete type を返すことを基本とし、mock のためだけの interface を作らない。
- `internal/` は外部 module から import させない code に使う。`cmd/` は複数 command の entry point を整理する場合に使う。

Handler → Service → Repository、Clean Architecture、layer-first/domain-first は Go/Gin 公式規約ではない。採用・変更する場合は、依存関係と変更単位を根拠に ADR で決める。

## AnimalEkarte domain/capability architecture

これはGo/Gin公式要件ではなく、[Product Philosophy](../docs/product-philosophy.md)を実装へ落とすためのproject decisionである。境界の正本は[ADR-006](../docs/architecture/adr/006-backend-domain-package-boundaries.md)とする。

- `internal/<domain>`を基本とするdomain/capability-firstのmodular monolithとし、route、use case、transaction、persistence、testをvertical sliceで変更する。
- BE9-2B以降、既存未移行codeの保守・安全修正と移動に必要なcompatibility変更を除き、新規production実装を`internal/handler`、`internal/service`、`internal/repository`へ追加しない。新規実装はADR-006のtarget domain packageへ置く。BE9移行は2026-07-24にcode complete（release pending）となり、境界の正本は[ADR-006](../docs/architecture/adr/006-backend-domain-package-boundaries.md)と[boundary map](../docs/architecture/be9-2a-boundary-map.md)とする。
- domain内の`handler`、`service`、`repository` subpackageは必須ではない。実際のconsumer、依存方向、変更周期が分かれる場合だけ分割する。
- business factごとにsource of truthとwrite ownerを1つにする。`appointments`とそのlifecycleは`reservation`がwrite ownerであり、他domainから独立した直接writeを行わない。owner外はbusiness intentを表すconsumer-side interfaceだけを宣言し、generic field-update APIを受け取らない。この境界は[ADR-006](../docs/architecture/adr/006-backend-domain-package-boundaries.md)を正本とし、[`appointment_write_owner_lint_test.go`](internal/reservation/appointment_write_owner_lint_test.go)で回帰を検出する。
- appointmentに紐づく通常カルテは一般診療予約だけを対象とし、appointmentごとにactive recordを最大1件とする。日付は予約日時のJST日付から導出し、紐づいている間は独立変更させない。削除は対象行をlockしたtransaction内で見積依存を再確認してから`clinic_id + id + status=draft`の原子的条件付きsoft deleteを行う。見積Createも同じ親行を先にlockし、見積が先なら削除をConflict、削除が先なら後続見積を拒否する。確定との競合でも確定済みカルテを削除しない。検証・重複確認・transaction依存の欠落や失敗はfail-closedにする。
- cross-domain呼び出しはbusiness intentを表すconsumer側の最小interfaceと型安全なDIを基本とする。owner外へ`map[string]any`等の任意field更新APIを公開せず、複数domainにまたがるwriteはownerとtransaction境界を明示する。
- appointment、trimming detail、option等で1つのbusiness graphを構成するwriteは同じtransactionで全体を成功またはrollbackさせる。既存trimming appointmentのowner欠損はappointment fieldを変更しないdetail-only writeでも補完する。参照master・担当者・LINE顧客を検証する必須依存が欠ける場合はwrite前にfail-closedとし、LIFFで明示指定されたstaffにはclinic所属・対応可能種別に加えて`is_active=true`かつ`reservation_visible=true`を要求する。best-effortを選ぶ場合は部分成功contract、再試行、補償、監査を明示する。
- fail-closedと定めたclinical/financial監査はbusiness writeと同じtransactionへ参加させ、監査dependency欠落または監査write失敗時はbusiness writeもrollbackする。締め後の会計編集はこの対象とする。
- `FOR UPDATE`、`FOR SHARE`、`pg_advisory_xact_lock`を正しさの根拠にするoperationはambient transaction不在を拒否する。request由来のclinic-scoped FKは永続化と同じtransactionで再検証し、並行master変更で判定が無効になる場合は対象行をcommitまで共有ロックする。
- nested GORM `Preload`は末尾associationの条件だけでは中間associationをscopeできない。clinic-ownedの中間関連も明示的に`Preload(..., "clinic_id = ?", clinicID)`し、runtime clinic-isolation testとAST gateを併用する。
- compatibility facadeは薄いdelegate/type aliasだけを許可し、business ruleやpersistence実装を複製しない。consumer移行後の削除条件を持たせる。
- 自動処理には停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを設ける。
- 自動status transitionは、対象条件をwrite時にcompare-and-setで再評価する。臨床記録など遷移を否定するbusiness evidenceも同じ判定へ含め、resource単位の監査が必須なら状態変更と同じtransactionでfail-closedにする。同じevidenceを逆向きに変更する競合workflowがある場合は、両者を同じresource-scoped serialization機構へ参加させ、各writeのcommitまで順序を保持する。
- masterの「使用中は削除不可」は、Find→CountUsage→Deleteの非原子シーケンスを正しさの根拠にしてはならない。正しさの境界は`clinic_id + id`とusage不在（または許可されたstatus等）を同一SQLに束ねた条件付き原子DELETE（または同等のcompare-and-set soft delete）とし、`RowsAffected == 0`をConflict/NotFoundに正規化する（正例: `billing.estimateRepository.DeleteIfNotLocked`）。早期CountはUX用に残してよいが、防御本体にしてはならない。usage attachが並行し判定を無効化し得るpathでは、上のFK再検証条項に従いrequest由来FKを同txで再検証し、必要なら親masterをcommitまで共有ロックする（`FOR UPDATE`/`FOR SHARE`を正しさの根拠にするならambient tx必須）。新規productionと当該Deleteを変更する実装はこの規則に従う。既存の非原子Count→Delete（例: inventory / merchandise 他多数master）は既知のresidual race debtとし、一括retrofitは別作業とする（本規則の文書化だけではproductionを変えない）。
- LSTEP等のscheduled/cronバッチに、上のbest-effort条項（部分成功contract・再試行・補償・監査）と自動処理条項（停止・失敗通知・監査・手動fallback・idempotency/retry）を次のように適用する。(1) バッチ成立条件の欠落（必須dependency未構成、clinic一覧取得失敗、必須設定サービス不在）と、clinical status遷移のようにresource単位の監査が必須なwriteはfail-closedとする（正例: no-showのMark+audit同一transaction）。(2) multi-clinic / multi-owner / multi-triggerで1件失敗後も続行するintentional best-effortを選ぶ場合は、部分成功contract（`(processed, errs)`およびdurable向け`BatchRunResult`で`Processed = Succeeded + Failed`）、失敗のerrorログとFailed計上、`processed_count`/`error_count`を含む監査、再実行時のidempotencyまたは再評価、外部API失敗時の補償または次回runでの収束方針、sync無効化等の停止手段とPartial/Failedを契機とする手動fallbackを明示する。(3) 失敗をログのみで飲み込み成功扱いにするsilent swallow（例: 取得失敗で`return 0, nil`としerror_countに載せない）は新規禁止。既存のsilent pathはknown debtとし、触る変更でintentional best-effortまたはfail-closedへ寄せる。(4) 副次的side-effect（tag cache削除、API失敗カウンタ更新など）をbest-effortにする場合もログ必須とし、本処理の成功契約を反転させないことと失敗時の収束/補償を短く明示する。
- folder移動だけでclinic isolation、authorization、clinical safetyが成立したと判断しない。既存のruntime testとapplication invariantで検証する。

## HTTP with Gin

- route group で prefix と middleware scope を表現する。
- public route、authentication、authorization の境界を route 登録時に明示する。
- handler の dependency は closure または struct で注入し、package global に固定しない。
- application が error response を制御する通常の endpoint では `ShouldBind*` を優先し、binding error を必ず処理する。
- body/query/URI/header を型付き input に bind し、型・形式・範囲・長さを境界で検証する。
- response は公開 contract に必要な field だけを含める。
- OpenAPI contract と route、request、response、status code を同期する。
- error mapping は一貫した境界または middleware に集約し、未知の error は内部情報を含まない 500 にする。

## Context and dependencies

- request-scoped 処理には `c.Request.Context()` を渡し、DB と外部 API まで cancellation/deadline を伝播する。
- Context を struct に保存しない。
- `WithCancel` / `WithTimeout` / `WithDeadline` の cancel 関数を必ず呼ぶ。
- dependency は constructor、closure、struct field など型安全な方法で明示する。
- `gin.Context` への dependency 格納は型 assertion が必要なため、必要な request-scoped 値に限定する。

## Errors and logging

- error を無視しない。必要な文脈は `%w` で追加し、`errors.Is` / `errors.As` を使う。
- 同じ error を複数箇所で重複ログしない。十分な request 文脈を持つ境界で1回記録する。
- log は構造化し、secret、token、credential、owner/pet/staff/medical data を含めない。
- DB error、SQL、stack trace、内部 path を response に出さない。
- panic/recover は通常の error 処理の代替にしない。

特定の error helper、logger package、wrap/log の配置は公式未規定である。既存 helper を変更する場合は互換性と error chain を確認する。

## Persistence and integrity

- query は parameterized し、DB call へ Context を渡す。
- transaction の開始、commit、rollback、resource ownership を明確にする。
- write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。
- schema constraint と application validation の両方を使う。
- migration は versioned SQL で管理し、application startup の暗黙 migration に依存しない。
- clinic-scoped data は、ORM、raw SQL、join、preload、count、bulk operation、background job のすべてで認証済み clinic に制約する。
- client が送信した clinic/owner/pet/staff ID を認可根拠にせず、関連 resource の ownership を server-side で確認する。

ORM や repository pattern の採否は Go/Gin公式未規定である。GORM 固有の規則を Go/Gin 公式規約と呼ばない。

## Security and production

- HTTPS、明示的 CORS allowlist、安全な cookie、認証方式に合う CSRF 対策を行う。
- trusted proxy を明示し、proxy を使わない場合は信頼 proxy を無効化する。
- authentication と authorization を別々に確認する。
- rate limit、request/body/upload size、content type を制限する。
- `http.Server` に workload に合う timeout/limit を設定する。
- SIGINT/SIGTERM を受け、timeout 付き `Shutdown` と resource close を行う。
- goroutine から元の `*gin.Context` を使わない。必要なら `c.Copy()`、通常は標準 Context と必要値だけを渡す。

## Tests and verification

- HTTP handler/middleware は `net/http/httptest` と最小 router で検証する。
- binding、validation、authentication、authorization、not-found、conflict、internal-error を含める。
- tenant/ownership boundary の変更には unauthorized/cross-tenant test を含める。
- write ownerまたは状態遷移の変更には、owner外の直接writeがないこと、intent-specific operation、cross-domain transactionのrollbackを確認するtestを含める。
- write-ownerのAST gateは`FirstOrCreate`を含むmutation、typed parameter、free function/receiver method戻り値とcross-file/package-qualified factory、query変数、local/package constant、table alias・schema-qualified table、直接または変数代入した`TableName()`、静的string helper、schema-qualified raw SQL、generic appointment map mutatorを検出する。
- 自動処理の変更には、停止、失敗通知、監査、手動fallback、重複実行またはretry時の安全性を変更riskに応じて検証する。
- cancellation、concurrency、transaction、shutdown は変更 risk に応じて検証する。
- `gofmt`、`go test`、`go vet`、project lint を適用する。
- この repository では Go command を host で直接実行せず、Docker の scoped command を使う。
- full-project command の自動実行禁止は [`.claude/CLAUDE.md`](../.claude/CLAUDE.md) に従う。

coverage threshold や TDD workflow は project quality policy であり、Go/Gin公式のアーキテクチャ要件ではない。

## Before review

- [ ] package boundary と名前が利用者・凝集性を反映している
- [ ] 旧`internal/handler|service|repository`へ新規production実装を追加しておらず、残すfacade/adapterにconsumerと削除phaseがある
- [ ] business factのsource of truth/write ownerが一意で、owner外の直接writeや重複実装がない
- [ ] vertical slice、cross-domain transaction、自動化の停止/監査/fallbackが変更範囲に応じて検証されている
- [ ] master「使用中は削除不可」のDeleteを新規/変更した場合、条件付き原子DELETE（または明示した親ロック+原子DELETE）になっており、非原子Count→Deleteを正しさの根拠にしていない
- [ ] LSTEP/自動バッチのintentional best-effortに部分成功contract・Failed計上・監査・再実行/補償が明示され、silent swallowを新規に増やしていない
- [ ] Context、error chain、resource cleanup が維持されている
- [ ] input validation、authentication、authorization、ownership が独立している
- [ ] response/log に内部情報や個人情報がない
- [ ] clinic isolation がすべての data path で維持されている
- [ ] OpenAPI と実装が一致している
- [ ] scoped test と static checks が通っている
- [ ] 独自設計を「Go/Gin公式」と表現していない

## Primary sources

- [Go module layout](https://go.dev/doc/modules/layout)
- [Go package names](https://go.dev/blog/package-names)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Gin API design](https://gin-gonic.com/en/docs/routing/api-design/)
- [Gin dependency injection](https://gin-gonic.com/en/docs/middleware/dependency-injection/)
- [Gin binding](https://gin-gonic.com/en/docs/binding/)
- [Gin security guide](https://gin-gonic.com/en/docs/middleware/security-guide/)
- [Gin graceful shutdown](https://gin-gonic.com/en/docs/server-config/graceful-restart-or-stop/)
