# BE-refactor フォローアップ status（2026-07-02 第2次）

> BE-refactor.md 本体は並行セッションが編集中のため直接更新せず、本ファイルで
> フォローアップの結論と BE-refactor.md の陳腐化訂正を集約する（Truth Source Priority:
> executable checks / current code > BE-refactor.md の記述）。

## フォローアップ 5 件の結論

| # | フォローアップ | 結論 | 証拠 |
|---|---|---|---|
| 1 | R1-3 break_hours 既存データ監査 | ✅ 完了 | `docs/runbooks/break_hours_shape_audit.md`（read-only 監査 SQL + 手順）。保存ガード `validateBreakHoursShape` と予約 fail-closed は既にテスト済み |
| 2 | date-only drift 22件 contract 判断 | 🟡 修正方針確定・autonomous 実装 BLOCKED（FE/PO 確認要） | 下記「date-only drift」参照。R3-3 gate が interim 追跡 |
| 3 | RLS 実効化 | 🟡 ローカル実効性確認=✅ / full 実効化=BLOCKED（別タスク） | `rls_effectiveness_test.go`（6 subtest GREEN）。下記「RLS」参照 |
| 4 | docs 陳腐化 cleanup | ✅ 本ファイルで訂正集約 | 下記「BE-refactor.md 陳腐化訂正」 |
| 5 | dbOrTx inventory lint | ✅ 完了 | `dbortx_inventory_lint_test.go`（GREEN）+ CI job `dbortx-inventory` |

## dbOrTx inventory lint（#5）

`backend/internal/repository/dbortx_inventory_lint_test.go`。ambient transaction に参加する
（`dbOrTx(ctx, r.db)` を使う）repository メソッド 80 個を allowlist に固定し、go/ast で双方向突合する:

- **regression 検出**: 固定メソッドが `dbOrTx` をやめて `r.db.WithContext` へ revert すると fail
  （R1-1/R1-2 の部分コミット/TOCTOU 修正の退行）。
- **新規レビュー強制**: 未登録メソッドが `dbOrTx` を使い始めると fail → allowlist 追加時に
  atomicity/isolation テストの添付を促す。

**意図的な非対象（taint 限界）**: 「WithTx 内で呼ばれるのに `r.db.WithContext` を使う（参加漏れ）」の
検出は service→repository 跨ぎのデータフロー解析が必須で go/ast 単体では不可（master_fk_write と同じ断念）。
参加漏れの正本ガードは各 tx フローの atomicity テスト（`*_tx_atomicity_test.go`）。

## RLS（#3）

**現状（実測・database-reviewer 確認）**: 001_init.sql の RLS は `ENABLE` のみで `FORCE` なし。
migration に `CREATE ROLE`/`OWNER TO` が無く、アプリは POSTGRES_USER=テーブル owner（かつ superuser）で
接続し `app.current_clinic_ids` GUC を SET しない。よって **RLS は runtime で dormant**（owner/superuser は
非 FORCE RLS を bypass）。これは 001_init.sql:2895-2905 に明記された意図的 baseline。**実効的なテナント
境界はアプリ層**（repository の clinic_id 述語 + service の所有権/#124 検証）が担う。

**ローカルで安全に実装した範囲（✅・6 subtest）**: `rls_effectiveness_test.go` が probe テーブル + 非 superuser
ロール（`SET LOCAL ROLE`）+ `SET LOCAL app.current_clinic_ids` で、実ポリシー式 `app_private.has_clinic_access`
が (a) 越境 SELECT を隠し(USING・clinic 1/2 差分)、(b) **同一 clinic INSERT は成功**し越境 INSERT は拒否され(WITH
CHECK の positive/negative 両control)、(c) bypass_rls で batch 全件アクセスを許し、(d) GUC 未設定で全行不可視
(fail-closed) になり、(e) **非 FORCE では table owner が RLS を bypass（＝現行 dormant の根本原因）し、`FORCE ROW
LEVEL SECURITY` 適用後は owner も RLS に従う**ことを、すべて実 SQL で実証。**RLS ポリシーの correctness と実効化
readiness を固定**（適用済みスキーマ/アプリは不変更・probe テーブル/ロールはテスト内で作成/破棄）。
本番実効化では app の DB ロールが clinic テーブルの owner のため (e) の FORCE 経路が load-bearing。

**full 実効化=BLOCKED（別タスク・要 architect/PO）**: 非 owner アプリロール + 全 clinic テーブルへ
`FORCE ROW LEVEL SECURITY` + 全 DB tx で `SET LOCAL app.current_clinic_ids`（接続層/transactor/clinic-context
middleware 配線）+ batch bypass が必要。**all-or-nothing**（FORCE を配線なしで適用するとアプリ全クエリが
遮断され機能停止）で、接続層を跨ぐ高リスク改修のためフォローアップのスコープ外。

**app 層境界担保の既知の例外（`system_admin`）**: 「実効的なテナント境界はアプリ層」と述べたが、
`system_admin`（`is_system_admin`）は `permission_middleware.go`/`context_helpers.go` で clinic 所属
チェックを意図的にスキップする**審査済み**の app 層バイパス（AUDIT-H2, 2026-05 でレビュー済み。
`resolveAllClinicIDs()` が `mainClinicID` 単体に制限し全医院スキャンは防止）。RLS 実効化設計では
system_admin セッションの DB ロール/GUC 割当を必須の検討事項に含めること（下記「batch/bypass 経路監査」節）。

## date-only drift（#2）

**実測**: openapi `format: date` 宣言の date フィールド（birth_date/last_visit/neutered_date/date/
scheduled_date/valid_until/expiry_date/last_restocked 等）の response が handler で `*time.Time`（RFC3339
datetime wire）配信されており、response 側で 22 箇所の drift（`internal/apicontract` allowlist）。

**契約の非対称性（重要）**:
- **request 側**: `jsonDate`（`handler/date.go`・UnmarshalJSON のみ）が YYYY-MM-DD と RFC3339 の両方を受理。
  → openapi `format: date` は request 入力形式として妥当。
- **response 側**: 生の `*time.Time` → RFC3339 datetime。date-only 出力用の marshaler は不在。
  → openapi `format: date` と乖離（＝drift）。加えて 6/30 c0e32421 が tz-consistency のため response を
  意図的に `localTimePtr`（datetime）化した経緯がある。

**修正方針（確定）**: 本筋は **response を date-only 出力へ統一**（request の jsonDate 思想 + openapi
`format: date` と整合・semantic に date・tz 表示バグも解消）。openapi→date-time 化は request が date-only を
受理する事実と衝突するため不適。

**autonomous 実装 = BLOCKED（FE/PO 確認要）**: 22 の FE-consumed response フィールドの wire 変更
（datetime→date-only）は FE 表示への影響確認が要り、かつ 6/30 の意図的 datetime 化の revert になるため、
一方的変更はしない。**R3-3 gate（`openapi-date-format-drift` CI）が現状 22 drift を allowlist 固定 + 新規
drift を防止**して interim 追跡する（＝本領域の「対応」は drift の可視化・regression 防止で達成）。

## BE-refactor.md 陳腐化訂正（現コード/検証と不一致の箇所）

| BE-refactor.md 箇所 | 陳腐化した記述 | 現実（executable check） |
|---|---|---|
| L87 / L132（R2-1/R3-3） | 「openapi format↔実装は 6/30 で 0/76 整合」 | **不成立**。response 側に 22 drift（`internal/apicontract` allowlist）。R3-3 gate で可視化済み |
| L29 / L91（D4/R2-2） | medication-history 孤児 API を削除する | **既に削除済み**（`grep -rn medication-history backend/` 空） |
| L30 / L97（D5/R2-3） | permissionGroup.UpdateRules の二重化解消 | **二重化は現存せず**。72e8887c の対象は UpdateStaffGroups（clinic 貫通）で対応済み |
| L31–32 / L139–141（D6/R3-4） | repo テスト pool 枯渇・auth ライフサイクルテスト 0件 | **既に解消/存在**。`getTestDatabaseConnection` が mainDB close + MaxOpenConns(10) + t.Cleanup。`password_reset_token_repository_test.go`/`token_blacklist_repository_test.go` が期限/単回使用/失効/クリーンアップを網羅 |
| L38 / L159（D13/R3-7） | 臨床結果テーブルに DB 複合 FK 不在 | checkup_field_results=**migration 012 で複合 FK 追加済み**。exam_results=`clinic_id` 列不在で**構造的に複合 FK 不可**（RLS+app 層で境界担保・別タスク） |

## Phase A 追加調査（2026-07-03・RLS full 実効化の前提調査）

> 対象範囲は下記5項目。**ローカル作業IDで管理する（GitHub Issueではない）**:
>
> | ローカル作業ID | 内容 | 旧ラベル（非実在・参考のみ） |
> |---|---|---|
> | Phase A-1 | date-only FE 影響インベントリ | 旧 #216 相当 |
> | Phase A-2a/2b/2c | RLS Phase0（role privilege実測 / tx entrypoint inventory / 接続プーラー構成） | 旧 #219 相当 |
> | Phase A-3 | 非owner ロール RLS ローカル実証 | 旧 #221 相当 |
> | Phase A-4 | batch/bypass 経路監査 | 旧 #223 相当 |
> | Phase A-5 | BE-refactor.md 陳腐化記述の状態確認 | 旧 #225a 相当 |
>
> **注記**: 依頼元プロンプトはこれらを `#216`/`#219`/`#221`/`#223`/`#225a` という GitHub Issue
> 番号で参照していたが、`gh issue list --state all` および `gh issue view <N>` で確認した結果
> これらの番号は**実在しない**（2026-07-03 再確認時点の直近 Issue 最大値は #215）。本セクションでは
> 番号を実在の Issue 追跡として扱わず、上記ローカル作業IDで管理する。Issue化が必要な場合は
> 別途 PO 判断で起票する（起票ドラフトは `docs/be-refactor-issue-drafts.md` 参照）。

### Phase A-2a. RLS role privilege 実測（旧ラベル: #219 相当・非実在）

**実測方法**: `backend/internal/repository/rls_role_privilege_test.go`（新規）が
`SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user` を
ローカルテスト DB に対して実行し `t.Logf` で記録する（pass/fail 条件にはしない＝measurement gate）。

**実測結果（ローカル docker db・2026-07-03）**:

```
current_user=ekarte_user rolsuper=true rolbypassrls=true
```

これは本ファイル冒頭「RLS」節の質的記述（「POSTGRES_USER=テーブル owner（かつ superuser）で接続」）を
**実測で裏付ける**。superuser 接続は FORCE ROW LEVEL SECURITY すら bypass するため、現状のロール構成では
RLS を全 clinic テーブルへ `FORCE` してもアプリの挙動は変わらない（既知の dormant 理由の再確認）。
**STG/prod のロール構成は未確認**（本タスクは STG/prod DB へ接続していない・別スコープ）。

### Phase A-2b. DB トランザクション開始点の inventory（旧ラベル: #219 相当・非実在）

**⚠️ 訂正（2026-07-03 再検証・当初記述の誤り）**: 当初「`grep -rn '\.Begin(\|BeginTx('` の結果は 0 件」
「本コードベースは GORM の `.Transaction(...)` のみを tx 開始経路として使う」と記述したが、これは
**`internal/` 配下（GORM クエリ層・`cmd/api` のリクエストパスが使う層）に限れば正しいが、リポジトリ全体
としては誤り**。実際には非 GORM の生 `Begin()`/`BeginTx()` 呼び出しが 4 箇所存在する:
`cmd/migrate/main.go:177,316,351`（`database/sql` の `db.Begin()`、migration 適用ツール）、
`cmd/stage-import/apply.go:155`（pgx の `targetPool.Begin(ctx)`、STG データ import ツール）。
いずれも `cmd/api` が使わない独立バイナリ（migration/STG import 専用）であり、RLS `SET LOCAL` を
リクエストパスの tx 開始点へ配線する設計上の結論（下記）には影響しないが、「完全列挙」を謳う以上
0 件という書き方は不正確だった。正しい記述: **`internal/`（`cmd/api` 経由のリクエストパス）に限れば
GORM `.Transaction(...)` のみが tx 開始経路で `Begin()`/`BeginTx()` は 0 件。リポジトリ全体では
migration/stage-import ツールに 4 件の非 GORM `Begin()` が別途存在する（RLS 設計のスコープ外）**。

**`.Transaction(` 呼び出し site 全数（非テスト、2026-07-03 実測・再検証で 1 件の誤分類を訂正）**:

- **ambient tx 起点（`Transactor.WithTx`）**: `repository/transactor.go:29`
  （`context.WithValue(ctx, txKey{}, tx)` で ctx に tx を格納。`dbOrTx(ctx, r.db)` を使う
  repository メソッドはこの ambient tx に自動参加する。参加メソッド一覧は
  `dbortx_inventory_lint_test.go` の allowlist（80 個）が正本）。
- **`dbOrTx(ctx, r.db).Transaction(...)` で自身も ambient tx 参加可能な起点（4 箇所。当初 3 箇所と
  記述したが `trimming_repository.go:85` を「新規 tx を開始する起点」側に誤分類していたため訂正）**:
  `accounting_repository.go:298` / `checkup_field_repository.go:110` / `staff_repository.go:143` /
  `trimming_repository.go:85`
- **`r.db.WithContext(ctx).Transaction(...)` で ambient tx を無視し常に新規 tx を開始する起点（21 箇所。
  当初 22 箇所と記述し `trimming_repository.go:85` を誤って含めていたため訂正）**:
  `campaign_repository.go:92`, `clinic_repository.go:103`, `examination_repository.go:168`,
  `helpers.go:15,37`（`reorderByClinicID`/`reorderGlobal`。レシーバではなく引数 `db *gorm.DB` 経由のため
  grep パターンに `r\.db\.` 前提を使うと見落とす — 検証時の実際の落とし穴）, `lstep_tag_cache_repository.go:244`,
  `manual_article_repository.go:60`, `owner_repository.go:173`, `permission_group_repository.go:102,218,243`,
  `repositories.go:243`, `reservation_schedule_repository.go:91`, `reservation_staff_repository.go:68,130,226,295`,
  `reservation_type_liff_repository.go:104`, `shift_entry_repository.go:122`, `shift_template_repository.go:93`,
  `treatment_repository.go:185`
- **service 層の tx 起点（1 箇所）**: `service/lstep_csv_import_service.go:171`（`s.db.WithContext(ctx).Transaction`）

**RLS 実効化への含意（訂正後の数値で再計算）**: 将来 `SET LOCAL app.current_clinic_ids` を tx 開始点で
配線する場合、上記「21 箇所（ignore-ambient）+ service 1 箇所」は ambient tx（`Transactor.WithTx`）の
外側で**独立した新規トランザクション/コネクション**を必ず開始するため、ambient tx 側にのみ GUC を設定
する設計では漏れる。「dbOrTx 参加可能な 4 箇所」も ambient tx 無しで単独呼び出しされた場合は同様に新規
tx を開始するため**条件付きで漏れる**。合計 **tx 開始点は 27 箇所**（ambient 起点 1 + dbOrTx 参加可能
4 + ignore-ambient 21 + service 1）が GUC 配線の対象候補（＝全 tx 開始点への SET LOCAL 配線が必要。
dormant な現状ではこの漏れは実害なし）。

### Phase A-2c. 接続プーラー構成（旧ラベル: #219 相当・非実在・ローカルのみ）

`docker-compose.yml` に PgBouncer 等の外部コネクションプーラーは**定義されていない**。
`backend` サービスは `db`（postgres:18-alpine）へ直結し、アプリ内 `sql.DB` プールを
`repository/db.go`（`SetMaxOpenConns(50)` / `SetMaxIdleConns(25)` / `SetConnMaxLifetime(30*time.Minute)`）
で管理する唯一のプール層。外部プーラーが無いため、`SET LOCAL`（tx スコープ）方式の GUC 配線は
PgBouncer transaction pooling モードで `SET LOCAL` が意図せず別クライアントへ漏れる既知の罠
（プーラーがコネクションをトランザクション間で使い回す際に session-level GUC が残留する問題）を
**踏まない**。**STG/prod 構成は未確認**（別スコープ・本タスクは接続していない）。

### Phase A-3. 非owner ロール RLS ローカル実証（旧ラベル: #221 相当・非実在）

`backend/internal/repository/rls_effectiveness_test.go`（**git 管理外・未コミット**、2026-07-03 時点で
disk 上に存在）を再実行し、6 subtest すべて PASS を再確認した:

```
--- PASS: TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC (0.36s)
    --- PASS: .../USING:_clinic_1_セッションは_clinic_1_の行のみ見える
    --- PASS: .../USING:_clinic_2_セッションは_clinic_2_の行のみ見える
    --- PASS: .../WITH_CHECK:_同一clinic_INSERT_は成功・越境clinic_INSERT_は拒否される
    --- PASS: .../bypass_rls:_batch_セッションは全_clinic_の行を見える
    --- PASS: .../GUC_未設定セッションは何も見えない（fail-closed_の_DB_版）
    --- PASS: .../FORCE:_非FORCEでは_owner_が_RLS_を_bypass、FORCE_適用後は_owner_も従う
```

このテストは `CREATE ROLE ... NOLOGIN`（superuser/owner 権限なし・`GRANT SELECT, INSERT` のみ）で
作成した**真に非owner・非superuser**のロールへ `SET LOCAL ROLE` して USING/WITH CHECK/bypass/fail-closed を
実証しており、「非owner runtime role の RLS 有効性」は**ローカルで実証済み**と判断する。
**未コミットのため次回コミット時にこのファイルを含めること**（現状は working tree にのみ存在）。

### Phase A-4. batch/bypass 経路監査（旧ラベル: #223 相当・非実在）

**⚠️ 訂正（security-reviewer 指摘・当初調査の見落とし）**: 当初 `grep -riE 'bypass|batch.*admin|admin.*batch'`
（ASCII 英字のみ）で「app 層に bypass 機構は 0 件」と結論したが、これは誤り。**`system_admin`
（`is_system_admin`）が実在する app 層の RBAC＋テナント境界バイパス経路**であり、日本語コメント
「バイパス」（カタカナ）表記・識別子 `system_admin`/`IsSystemAdmin` の語順不一致により英字 regex が
false negative になっていた（`backend/internal/handler/permission_middleware.go:10,23`
「system_admin は全権限バイパス」/ `context_helpers.go:104-191` `extractIsSystemAdmin` →
`isAdmin` なら clinic 所属チェックを完全スキップ、ただし `resolveAllClinicIDs()` は
system_admin でも `mainClinicID` 単体に制限し全医院スキャンは防止済み）。この機構自体は
既存監査（AUDIT-H2, 2026-05）でレビュー済みの**既知**の app 層バイパスであり、本タスクが
新規発見したセキュリティホールではない。**訂正後の結論**: 「本番コードパスに bypass 機構は
一切現れない」は誤り→「`system_admin` という審査済みの app 層バイパスが存在する。ただし DB
接続ロール/GUC レベルでの bypass（`SET LOCAL app.bypass_rls` 相当）は存在しない」に訂正する。
RLS full 実効化設計では、**system_admin セッションの DB ロール/GUC 設計を必須の検討事項に含める
こと**（system_admin リクエストで `app.bypass_rls=on` にするか、`resolveAllClinicIDs()` が返す
許可 clinic 集合を `app.current_clinic_ids` に設定するかは architect 判断）。

**DB 接続ロールレベルの bypass**: `bypass_rls` という語は `rls_effectiveness_test.go` 内の
Postgres セッション GUC（`SET LOCAL app.bypass_rls = 'on'`）としてのみ存在し、本番コードパスには
一切現れない（この結論は訂正後も変わらず妥当。app 層 RBAC バイパスと DB 層 RLS バイパスは別階層）。

**batch エントリポイント inventory**: `backend/cmd/` 配下、HTTP リクエストを扱わない独立バイナリ 4 種:

| バイナリ | 用途 | DB 接続 |
|---|---|---|
| `cmd/migrate` | migration 適用 | `database/sql`、`DB_HOST`/`DB_USER` 等（app と同一 env）。**他3種と異なり意図的に local ガード無し**（migration は STG/prod にも適用する正規経路のため） |
| `cmd/lstep-migrate` | Lステップ関連移行 | 同上 |
| `cmd/seed-old-db` | ローカル seed（`DB_HOST` が local host 以外なら fail-fast する安全装置あり） | 同上 |
| `cmd/stage-import` | STG データ import（target/`DB_HOST` が non-local なら fail-fast、**さらに `--confirm-local-destroy` の2要素確認あり**） | 同上（`STAGE_DB_*` は読み取り元） |

（`cmd/api` は HTTP リクエストサービング。`cmd/coverage-ratchet` は DB 非接続の開発ツールで対象外）

**security-reviewer 指摘（MEDIUM・未対応の follow-up）**:
1. `seed-old-db`/`stage-import` の local ガードはホスト名文字列完全一致（`db`/`localhost`/`127.0.0.1`）のみで、
   SSH/kubectl port-forward 経由で STG/本番を localhost へトンネリングする運用では無力化される。typo 防止
   には十分だが、意図的/不注意な誤接続防止としては不十分。
2. `stage-import` には `--confirm-local-destroy` の2要素確認があるが、同じく破壊的な `seed-old-db` には
   ホストチェックのみで対応する確認フラグが無く非対称。

**結論**: 現状、batch バイナリと `cmd/api` は**同一 DB ロール**（`ekarte_user`、実測 rolsuper=true）で接続しており、
app 層で区別された「bypass 用の別ロール/別フラグ」は存在しない。RLS が dormant な現状では bypass 相当の状態が
全経路で常態化しているため、「bypass が request path に漏れる」という意味でのリスクは**現状は成立しない**
（漏れる対象の区別自体が無い）。

**leak 防止テスト/lint の要否 = 現時点では実装しない（BLOCKED・時期尚早）**: 存在しない bypass 機構に対する
漏洩防止ガードは検証対象が無く、実装すれば根拠のない speculative コードになる（本プロジェクト方針
「推測オーバー実装禁止」に反する）。**この lint は RLS full 実効化設計（非owner interactive ロール + 別
BYPASSRLS batch ロールの新設）と同時に作るべき follow-up**として記録する: 実効化時に
「`cmd/api` の接続プールが BYPASSRLS ロールを使っていないこと」を機械的に検出するテストを追加する。

**⚠️ 誤読防止（database-reviewer 指摘・重要）**: 「RLS が dormant」は「テナント境界が無い」ことを意味しない。
現行のテナント分離は **RLS ではなく app 層の clinic_id 述語**（P3.1 `preload_clinic_scope_lint_test.go` の
機械強制・P4 `clinicScope`・各種 `*_clinic_isolation_test.go`）が正本として担っている。RLS 実効化は
defense-in-depth の追加層であり、それ自体が現行防御の代替ではない。本 batch/bypass 節の結論
（「今は leak 対象が無い」）は RLS 文脈限定であり、app 層防御の状態を評価するものではない。

### Phase A-1. date-only FE 影響インベントリ（旧ラベル: #216 相当・非実在）

`knownDateFormatDrifts`（18キー・22箇所）について frontend 3アプリ（`src/`/`liff/src/`/`line-reserve/src/`）を
grep 調査した（read-only・FE コード変更なし）。

**総括**: 18キー中 **14キーは既存の transform 層**（`.split("T")[0]` / `.slice(0,10)` / `formatJSTDate` 経由）が
防御的に日付部分だけを取り出しており、datetime→date-only 移行そのものによる直接破壊は**低リスク**。
残り3パターンは移行前に個別対応が必要:

| 対象 | 箇所 | 問題 |
|---|---|---|
| `examination.date` | `features/medical-records/api/get-record-examinations.ts:29` | カルテ内検査一覧が `.slice(0,16).replace("T"," ")` で**時刻まで意図的に表示**（例: `09:30`）。date-only 化すると時刻表示が失われる**臨床上の実害を伴う回帰**。他の examination.date 消費箇所（`transforms.ts:33`）は日付のみで無害だが本箇所だけ非対称 |
| `inventory.expiry_date` / `last_restocked` | `features/inventory/api/inventory.ts:29,31`、`InventoryList.tsx:228` | 変換なしで生 ISO 文字列をそのままテーブル描画（**date-only化とは独立の既存表示バグ**。移行はむしろ改善方向だが正しいフォーマッタ通過は別修正が必要） |
| `estimate.valid_until` | `EstimateForm.tsx:105` / `EstimateDetail.tsx:108` / `EstimateList.tsx:117,200`（4箇所個別 `.slice(0,10)`） | transforms.ts が無変換で通し、描画層4箇所がそれぞれ防御的 truncate。移行は安全だが正規化ポイントが分散し保守性が低い |

**調査ギャップ（未確定・follow-up 要）**: `owner.last_visit` / treatment CRUD 直接画面の `treatment.date` /
`medical-records/api/checkups.ts:28` の `date` は、決定論的 grep の時間予算内では消費箇所を確定できなかった。

**結論**: date-only 化の実 wire 変更は本タスクのスコープ外（FE/PO 確認要・R3-3 gate が interim 追跡、上記
「date-only drift」節と同じ結論）。本インベントリはその判断時の作業リストとして機能する。

### Phase A-5. BE-refactor.md 陳腐化記述の状態確認（旧ラベル: #225a 相当・非実在）

本ファイル冒頭で D4/R2-2・D5/R2-3・D6/R3-4 の陳腐化訂正は既に集約済み（上表参照）。
`git diff BE-refactor.md` を確認したところ、**BE-refactor.md 本体には R3-3/R3-4 領域に未コミットの
並行セッション編集が現在進行中**（`状態（2026-07-02 検証・完了）` 等の追記）。本ファイル自体の
「並行セッション編集中は直接更新しない」方針に従い、**BE-refactor.md への直接追記は今回も見送り
BLOCKED として記録する**（衝突回避・本ファイルへの追記で代替済み）。

### Docker scoped 検証コマンド（実行済み）

```
docker compose exec backend go test ./internal/repository/... \
  -run 'TestDBConnectionRolePrivileges_LocalMeasurement|TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC' -v
```

結果: 2 テスト（`TestDBConnectionRolePrivileges_LocalMeasurement` 1件 + `TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC` の 6 subtest）すべて PASS。
