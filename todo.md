# タスク台帳 — Linear が正本

更新日: 2026-09-06

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **直下整理** | **[BRT-105](https://linear.app/baritechllc/issue/BRT-105)**（Done） |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **会社側ログ** | CorpVault `50_Projects/ノア動物病院電子カルテ/` |
| **完了履歴** | Git 履歴と Linear の完了 Issue。完了項目は本ファイルへ残さない |
| **本ファイルの範囲** | repo と強く結び付く **全未完了作業の入口**（実装・リファクタ・CI/UAT・STG・納品/本番切替USER gate・指示監査） |

## 使い方

- 状態・担当・次の一手は Linear を正本とする。
- 確認済みの製品 FAIL は [`bug.md`](bug.md) に記録し、その後 Linear Issue 化する。
- 行値、患者情報、飼主情報、パスワード、接続資格情報は書かない。
- 完了した項目は削除する。経緯が必要な場合は Git 履歴を参照する。
- 現行ローカル handoff 表の正本は本ファイルの「観測」節。手順は [`OLD_DB_HANDOFF_LOCAL.md`](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)。
- live 共有 STG への再接続はしていない。判定は owner-only report、gitignored 配置、GitHub Actions、zsh 履歴、sibling `old_db` 台帳に限る。

### 実行者

| 実行者 | 意味 |
|--------|------|
| **old_db** | sibling producer。この repo からは CSV を生成できない |
| **USER** | 共有 STG、PlanetScale、医院調整、本番 cutover |
| **agent** | この repo のローカル専用。共有 STG には触れない |

エージェントは PlanetScale、共有 STG への apply、schema 再作成、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しません。

本台帳の更新も claim protocol の対象とする。今回の更新 claim は `claim/TODO-NEXT-ACTIONS`。エージェントは削除せず、main 統合または明示的な中止後に USER が解放する。

---

## 直近の実行順（2026-09-06）

実行状態の正本は Linear。以下は、現在の `main@9ed814fc0cd624fc19e30596708eb4646d24289c` と GitHub Actions の実行証拠から整理した、repo 内で次に進める順序。外部操作、共有 STG、release readiness は個別の承認・証拠なしに完了扱いしない。

| 優先 | ID | 実行者 | 状態 | 次の一手 | 完了条件 / 証拠 |
|------|----|--------|------|----------|-----------------|
| 1 | **CI-BE-DBORTX-INVENTORY** | agent | PR #369 blocker | `clinicRepository.FindByIDs` と `FindActiveIDs` の ambient transaction 参加をテストで固定し、DBOrTx inventoryへ登録する | scoped clinic/lintscanテストが成功し、PR #369の `Backend Test (remaining)` と集約 `Backend` が実行成功 |
| 2 | **CI-K6-SUMMARY-SCHEMA** | agent | 原因確定・未修正 | Performance workflowのaggregate validatorを、実artifactのフラット形式 `metric.count` / `metric.passes` / `metric.fails` に対応させ、旧 `metric.values.*` 形式も必要なら後方互換で扱う。workflow contract testを先にREDへする | contract test、構文・format・scope gateが成功。threshold、認証、負荷条件は変更しない |
| 3 | **CI-K6-RUNTIME-CLOSEOUT** | agent / USER | 優先2待ち | 修正をmainへ統合・push後、新しい明示承認のもとPerformance Testsを1回dispatchする | endpoint k6、spike k6、aggregate validation、always-run cleanupが同一runでsuccess。run IDとhead SHAを記録 |
| 4 | **LEDGER-CI-EVIDENCE-SYNC** | agent | 優先1・3待ち | 本節と関連台帳の古い「performance/k6未実行」を、成功・失敗・未確認を分離した実行証拠へ更新する | 対象SHA、run ID、job/step結論、未証明のSTG/PROD・release境界が一致。完了項目は本ファイルから除去 |
| 5 | **FE-CLINICAL-PLAN-SELECT-LABELS** | agent | 未起票・未着手 | 「診断カテゴリ」「診断病名」labelと `SearchableSelect` triggerの未接続を現HEADで再確認し、別ID・claim・allowlistを確定する | visible labelからtriggerへ到達・focusできる回帰テストとscoped accessibility reviewが成功 |

### 現在の証拠境界

- PR #369 / run `34020749135`: `Backend Test (remaining)` の `TestDBOrTxInventory_MatchesAllowlist` が、上記2メソッドのinventory未登録で失敗。push runのsuccessだけでPRをgreen扱いしない。
- Performance Tests / run `34020760108`: endpoint k6、spike k6、artifact upload、cleanupはsuccess。aggregate validationのみfailure。
- 同Performance artifact: endpointsは `http_reqs=5755`、`iterations=1918`、`successful_logins=1`、spikeは `http_reqs=825`、`iterations=824`、`successful_logins=1`。validatorが `metric.values.*` のみを読んだため0と誤判定した。値は活動量のみで、credential・body・cookie・tokenは記録しない。
- 上記はCI/負荷経路の証拠であり、共有STG、PROD、full clinical/data-dependent E2E、release readinessの証明ではない。

### 既存作業との順序

1. まず優先1でPR #369の必須checkを回復する。
2. 次に優先2・3でPerformance aggregateを修正し、runtime closeoutする。
3. repo CIがgreenになった後、下記 **STG 残作業索引** のHAC-CSV-1以降をUSER/old_db境界で進める。
4. スキル・AGENTS.md監査は既存の優先1〜3を維持するが、release/staging blockerを先に処理する。

---

## 分散台帳から集約した残作業（正本移行中）

未完了タスクのrepo内入口を本ファイルへ集約する。実行状態・担当・Doneは引き続きLinearが正本。元ファイルは監査根拠・完了履歴・詳細runbookとして参照し、同じタスクの状態を複数ファイルで独立更新しない。ただし `todo-now.md` は別所有の `claim/TODO-NOW` が存在し、現物にactive checkboxが残るため、同claimの所有者が入口ポインタ／履歴へ変更するまでは正本移行未完了とする。

| 統合元 | 本ファイルへ移したもの | 統合しないもの / 今後の役割 |
|--------|------------------------|-------------------------------|
| [`todo-now.md`](todo-now.md) | F1〜F6のLinear照合、F3 runtime残差、ClinicalPlan Select label残差 | F1〜F6の実装・統合履歴。`claim/TODO-NOW` があるため本更新では編集せず、pointer化まで移行待ち |
| [`todo-refactor.md`](todo-refactor.md) | TASK-444、トリミング権限・死亡guard | 完了済みFE規約リファクタの履歴。完了カテゴリ・検証表はbacklog化しない |
| [`be-refactor.md`](be-refactor.md) | openのMEDIUM 3件・LOW 8件 | 2026-09-04監査の詳細根拠と982 production Go fileのcoverage表。閉鎖・却下項目は再登録しない |
| [`fe-refactor.md`](fe-refactor.md) | なし（残件0） | 完了監査・ファイルstatusの履歴 |
| [`todo-po.md`](todo-po.md) | なし | Linearと本ファイルへの入口ポインタを維持 |
| [`docs/delivery/`](docs/delivery/README.md) | U1〜U13、本番構築・go-live USER gate | 契約入力、操作説明会、当日runbookの詳細正本。秘密値・個人情報は本ファイルへ転記しない |
| [`docs/ops/deploy/runbooks/`](docs/ops/deploy/runbooks/) | credential rotationと非機密receipt | 承認後にUSERが実施する詳細手順。credential値は転記しない |
| [`docs/ops/testing/UAT-DOMAIN-STATUS.md`](docs/ops/testing/UAT-DOMAIN-STATUS.md) | 製品FAILではないBLOCKED/PARTIALの検証残差 | 過去UATのシナリオ別履歴。`BUG-20260905-001 open` は現行 [`bug.md`](bug.md) の「未対応なし」と競合するため移さない |

### 調整・検証残件

| 優先 | ID | 実行者 | 状態 | 次の一手 / 完了条件 |
|------|----|--------|------|----------------------|
| A1 | **META-LINEAR-F1-F6** | USER / agent | 未照会 | LinearでF1〜F6と下記新規IDの重複・進行中Issueを確認し、必要なIDだけ紐付ける。repo記録だけでDoneへ変更しない |
| A2 | **LEDGER-TODO-NOW-POINTER** | `claim/TODO-NOW` owner | claim待ち | 本ファイルとの重複を解消し、`todo-now.md` をF1〜F6の完了履歴＋本ファイルへの入口ポインタへ変更する。現在のowner以外は編集しない |
| A3 | **QA-FULL-CLINICAL-E2E** | agent / USER | 未準備・未証明 | auth smokeとは別に、clinical/data-dependent E2Eのfixture・allowlist・cleanupを設計する。実行は別承認。full E2E成功と秘密非出力を証拠化 |
| A4 | **QA-UAT-S09-FIXTURE** | agent | BLOCKED | `completed_at` を指定できる承認済みfixture APIまたはscoped UAT helperを用意し、既存会計・DB・システム時計を直接変更せずS09 #2〜#6を再実行できるようにする |
| A5 | **QA-UAT-LSTEP-REAL** | USER | 外部環境待ち | write有効・到達可能なLSTEP環境でS01同期とV05-17 removeを確認する。行値・credentialを残さず、結果enumと制限付きreceiptのみ記録 |
| A6 | **QA-UAT-LINE-IDTOKEN** | USER | mock外・未証明 | 実LINE idTokenでS12のbackend link、409再連携、期限切れ400を確認する。token・LINE user ID・個人情報を証跡へ残さない |
| A7 | **DOC-MANUAL-SOURCE-SYNC** | agent | 未着手 | 埋め込みmanualの初回強制password変更、自動生成・一時password再発行の旧記述を現契約へ同期し、`.claude/CLAUDE.md` の絶対表現も現行例外契約へ合わせる |
| A8 | **CLINICAL-IRREVERSIBLE-GUARD** | agent | 要再確認 | `MedicalRecordFormActions.tsx` の不可逆操作が確認dialogだけに依存していないか現HEADで確認し、必要ならserver側状態検証・lock・補償経路を別タスク化する |

### Frontendリファクタ残件

| ID | 優先度 / 状態 | 方針 / 完了条件 |
|----|---------------|-----------------|
| **TASK-444** | deferred・未実施 | generated/modelsのdomain import 267件は凍結中。一括置換を始めず、公開契約・codegen・consumer移行計画と回帰gateを別単位で確定する |
| **FE-TRIMMING-GUARDS** | 要再確認 | トリミングフォームの権限ref・死亡guardの既存欠落を現HEADで再現する。別ID・claimでTDDし、他フォームの一括変更へ広げない |
| **FE-CLINICAL-PLAN-SELECT-LABELS** | 直近優先5 | 上の直近実行順を正本とし、ここでは重複定義しない |

`todo-refactor.md` に残るローカル`pnpm build`未実行は、現在のCI Frontend Build成功と区別する証拠境界であり、独立タスクとしては登録しない。`readiness-report.md` の300行超fileもsoft 500 / hard 800違反ではないため、行数だけを理由に一括分割しない。

### Backendリファクタ残件

2026-09-04監査時点のopen所見。着手前に現HEAD・Linear重複・該当diffを再確認し、横断一括キャンペーンではなく触る機能単位で小さく閉じる。

| ID | Severity | 状態 / 再開条件 | 方針 |
|----|----------|-----------------|------|
| **BE-RC-005** | MEDIUM | residual | serviceとHTTP境界の5xx二重ログを、新規・変更serviceから解消する。既知4xxはserviceでログせず、5xxはmiddlewareへ一本化。運用・batchの意図的ログは消さない |
| **BE-RC-009** | MEDIUM | residual | 実装側の広いRepository interfaceを、新規consumerまたは対象機能変更時に利用側最小portへ分割する。一括分割しない |
| **BE-RC-014** | MEDIUM | upstream条件待ち | pgx encode判定の`err.Error()`文字列依存。typed errorが利用可能になった時点で`errors.As`へ移行し、現例外を増やさない |
| **BE-RC-015** | LOW | incremental | package.Type stutterは新規・変更面から避ける。一括renameしない |
| **BE-RC-017** | LOW | incremental | exported `Update(..., map[string]any)` は対象repository変更時にunexported update + typed commandへ寄せる。一括unexportしない |
| **BE-RC-019** | LOW | deferred | `medicalrecord`の凝集圧はlab / hospitalization等の業務能力境界が成立する変更時だけ分割を検討する。layer subpackageへ戻さない |
| **BE-RC-021** | LOW | incremental | 新規exportへGoDocを追加する。既存全exportの一括修正はしない |
| **BE-RC-023** | LOW | conditional | clinic validatorのglobal `init()`登録でテスト順副作用を再現した場合、constructor登録へ移す |
| **BE-RC-034** | LOW | incremental | `cmd`に複製された業務ルールを、対象command変更時にauth / clinic / csvimport owner packageのhelperへ寄せる。汎用helper packageを新設しない |
| **BE-RC-035** | LOW | touch-time | auth testの生`TRUNCATE`を当該test変更時に`testdb.Truncate`へ移す。他packageを一掃しない |
| **BE-RC-036** | LOW | actionable | staff repository 4ファイルのstale package commentを`package staff`へ合わせ、comment-only差分として確認する |

### セキュリティ・納品・本番切替のUSERゲート

外部状態はrepo文書の最終記録から変わり得るため、実行直前にprovider・Linear・権限・対象環境を再確認する。エージェントは秘密値の作成・表示・投入、共有STG/PROD apply、production構築、go-liveを自動実行しない。

| 順 | ID | 実行者 | repo内の状態 | 必要な入力 / 完了条件 |
|----|----|--------|----------------|-----------------------|
| P1 | **SEC-SECRETS-5 / #89 / #97** | USER | 4系統rotation receipt未記入 | PlanetScale、Cloudflare、LINE、JWT/暗号鍵を「新発行→target明示で投入→再deploy→health→旧値revoke→旧値拒否」の順にrotation。値を記録せず、#97本文maskとnames-only登録確認まで行う |
| P2 | **#253 / U12 PROD-SETUP** | USER / 開発 | repo最終記録ではProduction未構築 | Cloudflare本番基盤、Required reviewers、production workflow、CF-only rollback、backup/restore rehearsal、実行時URL/CI receiptを揃える |
| P3 | **#250 PROD-DATA-MIGRATION** | USER / 開発 | 事前準備・当日証拠待ち | rehearsal、最終import、入力停止、backup/rollback、件数・clinic_id・金額突合を承認し、day-of gateでPASSを記録 |
| P4 | **#254 AUTHENTICATED-UAT** | USER / agent | full UAT未証明 | 全業務scenarioを認証付きで完走するか、残FAILを納品後対応の合意済み一覧へ隔離する。A3〜A6の残差を混同しない |
| P5 | **#255 STAFF-PROVISION** | USER | 入力未記入 | roster、email方針、clinic対応、role→permission group、休職等方針、actor、環境承認を供給し、PII-free receiptで個人account・所属院scope・権限を確認 |
| P6 | **#258 / U1〜U12 DELIVERY** | USER | 入力反映・最終承認待ち | P1・P2および契約責任者が供給した非機密事実とreceiptを`DELIVERY_PACKAGE.md`へ反映し、#258を最終承認する。U12のProduction証拠は#253が生成し、#258側で重複実行しない。秘密値はrepoへ書かない |
| P7 | **#256 / U13 TRAINING** | USER | 操作説明会未完 | 日程・形式・参加者範囲・実施結果・opaque receipt・別USER close承認を記録。D-254 FAQ/スクショ突合は#254後に行う |
| P8 | **#257 GOLIVE** | USER | HOLD・新window未記入 | P1〜P7、CSP本番API origin、通知、backup acquisition、Go/No-Go authority・support・rollback ownerがgreen後に切替日を設定。当日最終importと突合が未達ならNo-Go |

### 統合時に移さなかった項目

- `todo-refactor.md` / `fe-refactor.md` の完了済みカテゴリ、claim解放履歴、過去のtest件数。
- `be-refactor.md` の再検証済み合格、却下済み一括campaign、982ファイルcoverage表。
- `docs/ops/testing/UAT-DOMAIN-STATUS.md` の解消済み `BUG-20260905-001 open` 表記。現行 `bug.md` とGit履歴を優先する。
- `docs/ops/deploy/CI-CD-PIPELINE.md` の古いActions billing BLOCKED記述。現在はworkflow dispatch可能であり、現行課金状態を再確認せずbacklogへ入れない。
- runbook内の定常運用・当日手順チェックリスト。実行可能になる前提タスクだけを上表へ移した。

---

## スキル・AGENTS.md 監査対応事項（2026-09-06）

ユーザー指定により本節へ記録。Linear への登録・対応実装は未実施。状態・担当の正本は引き続き Linear とし、登録後は対応 Issue を各項目へ紐付ける。以下の未チェックは「監査で改善が必要と判断した指示」を示し、実アプリの脆弱性、秘密情報流出、DB 消失が発生したという意味ではない。

監査基準: [Rethinking skills and prompts for GPT-6 Astra](https://x.com/pvncher/status/2095991462416490862)。対象は HEAD `27630d286bd426b24e67074ceda8e13bf72e0bc0` 時点のスキル65ファイル・AGENTS.md 2ファイル（依存パッケージを含み、symlink の重複を除く）。着手時に現在の内容・既存 claim・他者 WIP を再確認する。

本節に列挙した監査対応は現在未claim。各実装単位のID・allowlistを確定してから個別にclaimを取得し、複数項目を一括所有した扱いにしない。

### 優先1 — データ・秘密情報・未コミット変更を守る

- [ ] **Docker 復旧例の破壊的操作を修正する。** 対象: [docker-patterns](.claude/skills/docker-patterns/SKILL.md) の node_modules 復旧例・キャッシュクリア例。[Makefile](Makefile) の `clean` はDBを含むボリューム削除を伴う。完了条件: 通常復旧から全ボリューム削除を外し、対象限定の手順と、データ削除が必要な場合の既存バックアップ・リセット契約への案内に置き換える。
- [ ] **フロントエンドの秘密管理例とログ例を修正する。** 対象: [security-checklist](.claude/skills/security-checklist/SKILL.md)。`VITE_API_KEY` を秘密保持の正例にしない。Email を非機密ログの正例にしない。完了条件: 秘密はサーバー側、フロント変数は公開可能な設定だけと明記し、秘密検出も値を出さない方式にする。
- [ ] **DB 診断で秘密値を表示しない。** 対象: [postgres-patterns](.claude/skills/postgres-patterns/SKILL.md) の環境変数一覧表示。完了条件: 非秘密キーの許可リストと秘密キーの設定有無だけを確認する例にし、DB_PASSWORD 等がログへ出る診断を残さない。
- [ ] **一時改変・失敗復旧の基準を HEAD から開始時内容へ変更する。** 対象: [golang-testing](.claude/skills/golang-testing/SKILL.md)、[refactor コマンド正本](.claude/commands/refactor.md)。完了条件: 他者 WIP を含むファイル全体の巻き戻しを避け、所有差分だけを復元する。RED 実証の一時改変は隔離環境または開始時内容を確実に保持した手順にする。
- [ ] **fresh-DB 検証を独立した使い捨て環境へ寄せる。** 対象: [migration-seed-safety](.claude/skills/migration-seed-safety/SKILL.md)、[golang-testing](.claude/skills/golang-testing/SKILL.md)、[scoped-verification-gates](.claude/skills/scoped-verification-gates/SKILL.md)、[stg-release-readiness](.claude/skills/stg-release-readiness/SKILL.md)。完了条件: 共有し得るDBのDROP・全ボリューム削除を通常手順にしない。migration/seed等に適用条件を絞り、docs-onlyへfresh-DB検証を要求しない。共有DB・migrationの実行権限は維持する。

### 優先2 — 生成物・参照先・検証権限の矛盾を直す

- [ ] **生成スキルの YAML 引用符処理を修正する。** 対象: [sync-agents-skills.sh](.claude/scripts/sync-agents-skills.sh)、[implement コマンド正本](.claude/commands/implement.md)。生成された `source-command-implement` の description が二重引用で構文エラーになる。完了条件: 引用符を含むdescriptionの回帰検証と生成frontmatter全件の解析が通る。生成物だけの直接修正で済ませない。
- [ ] **Codex ミラーの同期漏れと古い lint 手順を解消する。** 対象: 上記同期スクリプト、[commit前hook](.claude/hooks/pre-bash-commit-quality.js)、[ci-cd-automation](.claude/skills/ci-cd-automation/SKILL.md)、`scoped-verification-gates` の正本・ミラー。完了条件: 正本を修正後に同期し、生成物との差分を読み取りで検出できる。backendに含まれないgolangci-lintや削除済みpackageを実行先にしない。
- [ ] **旧台帳への書込み指示を現行の記録先に合わせる。** 対象: [task-create](.claude/skills/task-create/SKILL.md)、[browser-test](.claude/skills/browser-test/SKILL.md)、[implement コマンド正本](.claude/commands/implement.md)。完了条件: 廃止済みSTATUS.md・旧二台帳を復活させない。実行状態はLinear、確認済み製品FAILはbug.md→Linearとし、本節のようなユーザー指定の文書追記も扱う。外部投稿の許可がなければレビュー可能な下書きを完成させる。
- [ ] **モデル固定と限定検証の手動差戻しを解消する。** 対象: [scoped-verification-gates](.claude/skills/scoped-verification-gates/SKILL.md)、[browser-test](.claude/skills/browser-test/SKILL.md)、[implement-issue](.claude/skills/implement-issue/SKILL.md)、[harness](.claude/commands/harness.md)、[tdd-workflow](.claude/commands/tdd-workflow.md)、[CLAUDE.md](.claude/CLAUDE.md)。完了条件: Haiku/Sonnet等の名前ではなく実行環境・副作用・範囲で可否を決め、許可済みの限定検証を自律的に完了できる。コンテナが対象worktreeを検証することも確認する。
- [ ] **旧Go構造・BE9キャンペーン・委譲先の参照を更新する。** 対象: [golang-testing](.claude/skills/golang-testing/SKILL.md)、[golang-refactoring](.claude/skills/golang-refactoring/SKILL.md)、[test-generation](.claude/skills/test-generation/SKILL.md)、[clinic-isolation-auditor](.claude/agents/clinic-isolation-auditor.md)、`ci-cd-automation`、`tdd-workflow`。完了条件: 現行domain package・実在mockへ案内し、履歴化済みSession A/Bを現行必須手順にしない。存在しないgo-linting参照も更新する。
- [ ] **API・認証・性能計測の例を現行実装に合わせる。** 対象: [go-security](.claude/skills/go-security/SKILL.md)、[react-security](.claude/skills/react-security/SKILL.md)、[api-documentation](.claude/skills/api-documentation/SKILL.md)、[test-generation](.claude/skills/test-generation/SKILL.md)、[performance-profiling](.claude/skills/performance-profiling/SKILL.md)、[postgres-patterns](.claude/skills/postgres-patterns/SKILL.md)。完了条件: Wrap引数、Gin handler、CookieAuth/CSRF、API応答、コンテナ内ポートを正本と一致させる。React.memoをvalidationと扱わず、tenant分類・Preloadの境界も正確にする。
- [ ] **レビュー範囲と完了表示を実証に合わせる。** 対象: [review](.claude/commands/review.md)、[harness](.claude/commands/harness.md)、[implement-issue](.claude/skills/implement-issue/SKILL.md)、`go-security`。完了条件: staged/unstagedと所有範囲を明記し、未測定coverageや未実行テストをPASS/APPROVEDと表示しない。固定3回での停止は、残る安全な作業・失敗原因・完了条件を踏まえる方式へ見直す。

上記の対応に含める個別確認:

- `refactor.md` のexport機械削除は、Feature Indexingの公開契約と利用先を確認する手順へ変更する。
- `browser-test` の固定資格情報例は、指定環境の認証契約への案内に置き換える。
- `deployment` のpush/dispatch例は、その場で外部操作の承認境界を明記する。

### 優先3 — 必要な指示だけ読み込む構成にする

- [ ] **AGENTS.md を作業別の案内にする。** 対象: [AGENTS.md](AGENTS.md)、[CLAUDE.md](.claude/CLAUDE.md)、[claude-code-usage](.claude/rules/claude-code-usage.md)。完了条件: frontend/docs作業にGo/Gin詳細を一律要求せず、path・タスクに合う資料だけ読む。途中で判明した重要な仕様不明点は質問でき、独立した許可済み作業は続行できる。ハーネス固有のthink/compact手順は適用先を明確にする。
- [ ] **長いdescriptionと一般教材を短い入口・選択式referenceに整理する。** 対象: `golang-refactoring`、`go-gin-backend`、`golang-testing`、`test-generation`、`api-documentation`、`docker-patterns`、`performance-profiling`、`implement-issue`、`migration-seed-safety`。完了条件: 発火条件を狭くし、手法列挙やサンプル集は必要時だけ読む。Goテスト教材はgolang-testingへ統合し、FE固有観点を失わない。
- [ ] **重複・互換入口・ハーネス依存を整理する。** 対象: `react-security`と`security-checklist`、`deployment`と`ci-cd-automation`、`implement`と`implement-issue`、`clinic-id-isolation`、`generating-commits`。完了条件: 正本へ案内し二重管理を減らす。委譲は利用可能な機能で同じ監査を行えるようにし、共有treeの記述を推奨運用と誤読させない。analyzing-schema/database-indexing等の互換入口は呼出元を確認してから廃止を判断する。
- [ ] **代表タスクで改修後の指示を再評価する。** 完了条件: docs誤字修正・局所Go不具合・frontendテスト・migration設計・STG準備について、必要資料・許可/停止判断・完了証拠が整合する。臨床安全、tenant分離、transaction/監査の原子性、WIP/claim保護、local/CI/STGの証拠区別を保持する。モデル別効果・速度改善は未測定のまま断定しない。

依存パッケージ内のRedux/PlaywrightスキルとRechartsのAGENTS.mdは、今回の直接修正対象にしない。将来有効化する場合だけ、前提資料の大量読込み、依頼外の移行、global install/kill-all等をprojectの範囲・権限に合わせて評価する。node_modulesの直接編集や全スキル一括ロードは行わない。

---

## STG 残作業索引

推測で「未」にしていた項目は、下の観測で見直した。ここに残すのは **今も未完了と証明できるもの** だけ。

| 優先 | ID | 実行者 | 状態 | blocked-by | 根拠 |
|------|----|--------|------|------------|------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | 未 | なし | AE `hachioji/` 空。old_db の csv-export に八王子 run なし。old_db `todo.md` の HAC-CSV-1 は未チェック |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 | 八王子 bundle が無い |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | 待ち | H0-2、H0-3b / H1-2 | 八王子の STG apply report なし |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 証跡なし | なし | roster / secrets は 0600 で存在。apply の stdout report も zsh 履歴もなし |
| 5 | **H3-11 画面確認** | USER | 証跡なし | H3-9 が未ならログイン不能の可能性 | 飼主検索の確認ログなし。行値は残さない |
| 6 | **Lane 4** | 医院スタッフ / USER | 待ち | 両院の Lane 3。八王子が未 | 2026-09-05 の local UAT は Lane 4 ではない |

STG UAT 対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。`shikishima` / `hakobuneco` は対象外だが、STG には載っている（下表）。

索引から外したもの:

- **AE-OLD-DB-MR-UNIQ（現行 rehearsal 3院）** — ローカル CSV は非 NULL `medical_record_id` 重複 0。STG apply が unique index を通っている。producer に uniquify 実装あり。八王子分は HAC-CSV-1 に含める
- **AE-STG-UAT-LANE3-JOU の 21表投入** — owner-only apply report `PASS`
- **H3-7 の敷島 / Hako bu neco** — 同じく STG apply `PASS`（UAT 対象外）

---

## 観測（2026-09-06）

行値は書かない。live STG の残存は未照会。

### レーン

| レーン | 状態 | 根拠 |
|--------|------|------|
| Lane 0 城東 | 配置済み | `_old_db_handoff/jouto/` 21 CSV + manifest。layout check PASS。old_db `local-ae-v2` と manifest 契約ハッシュが一致 |
| Lane 0 八王子 | **未** | AE `hachioji/` 空。old_db csv-export に八王子ディレクトリなし |
| Lane 1 城東ほか | ローカル apply 済み | `*-apply.json` が `targetHost=db` / `ekarte_db` で PASS（2026-09-05）。八王子の H1-2 ではない |
| Lane 1 八王子 | H0-2待ち | bundle が無い |
| Lane 2 コード | **完了** | STG UAT importer / Make 経路は main 済み |
| Lane 3 城東 CSV | **投入済み** | `stg-uat-apply` `PASS`、lane `stg-uat-rehearsal`、2026-09-05 00:19 JST、PlanetScale STG |
| Lane 3 敷島 / Hako | **投入済み（対象外）** | 同日の STG apply `PASS` |
| Lane 3 八王子 | **未** | apply report なし |
| Lane 4 | **未** | 5営業日の STG 現場証跡なし。local UAT（`uat/20260905`）は別レーン。[`bug.md`](bug.md) の未対応 FAIL はなし |

2026-07-22 の八王子 F6 disposable apply `PASS` は **別物**（disposable DB、`clinicOrdinal=2`、現行 H0-2 ではない）。

### 現行ローカル handoff（gitignored）

| clinic | 配置 | layout check | STG apply | ローカル apply | manifest |
|--------|------|--------------|-----------|----------------|----------|
| jouto | 21 CSV + manifest | PASS | PASS（2026-09-05 00:19 JST） | PASS（`db` / `ekarte_db`） | `REHEARSAL_ONLY` / `jouto-intake-20260822-01` |
| hachioji | なし | FAIL | なし | なし | — |
| shikishima | 21 CSV + manifest | PASS | PASS | PASS | 同上 |
| hakobuneco | 21 CSV + manifest | PASS | PASS | PASS | 同上 |

`make stg-uat-handoff` / `make stg-uat-handoff-preflight` が zsh 履歴にある。wrapper は全医院。八王子は manifest なしで skip。

### Lane 3 手順の証跡

| 手順 | 判定 | 根拠 | 確度 |
|------|------|------|------|
| H3-1 他エンジニア確認・医院連絡 | 作業票なし | repo に記録なし | 不明（未実施とは断定しない） |
| H3-2 backup / 復元担当 | 作業票なし | 同上 | 不明 |
| H3-3 TTL 資格情報 | 入力あり | gitignored `scripts/stg-uat-old-db-handoff.local.env` が 0600 で存在。TTL そのものは未証明 | 中 |
| H3-4 staging deploy / migration | 充足とみなす | CSV apply が PlanetScale STG で PASS。`backend-deploy.yml` の staging 成功は 2026-09-05 09:18 JST（投入後）にもある | 中 |
| H3-5 skeleton | 充足とみなす | apply の preflight を通っている | 中 |
| H3-6 空 band | 充足とみなす | 9/4 失敗のあと 9/5 に PASS。一致しない非空 band は上書きしない契約 | 中 |
| H3-7 城東 / 敷島 / Hako | **PASS** | owner-only `stg-uat-apply` | 高 |
| H3-7 八王子 | **未** | bundle なし | 高 |
| H3-8 失敗側だけ修正 | 城東は失敗→成功 | 9/4 `FAILED_*` をリネームし、9/5 `PASS` | 高 |
| H3-9 staff attach | 入力あり・apply なし | roster `stg-uat-staff-attach-v1` と secrets が 0600。clinic 1–4 を含む。コマンド履歴なし。cmd は stdout のみで report ファイルを書かない | 高 |
| H3-10 未来日 shift | 任意・証跡なし | — | — |
| H3-11 飼主検索 | 証跡なし | — | — |

### AE-OLD-DB-MR-UNIQ

現行 rehearsal 3院は充足。残作業索引には置かない。

- AE 配置 CSV は非 NULL `medical_record_id` 重複 0
- その bundle の STG apply が unique index を通過
- old_db に uniquify（`scripts/lib/local-ae-billing-mr-uniq.mjs`）とテストがある
- AE `jouto` manifest の契約ハッシュは old_db `...-local-ae-v2` と一致（未加工 `rehearsal-current` とは不一致）
- 八王子は未出力のため、今後の HAC-CSV-1 で同じ一意制約を満たす

### Lane 1 任意

| 項目 | 判定 |
|------|------|
| H1-3 A4 UI rehearsal | 任意。`sensitive-local/a4-rehearsal-reports/` なし |
| H1-4 F8 G4 | 任意。`sensitive-local/f8-g4-rehearsal/` なし |
| H1-5 | 規則。失敗 bundle を STG へ送らない |

---

## タスク詳細

### 1. H0-2 / HAC-CSV-1

| 項目 | 内容 |
|------|------|
| 実行者 | old_db / USER |
| 目的 | 八王子の医院 identity 付き 21 CSV と manifest を出す |
| 受け入れ | `_old_db_handoff/hachioji/` に 21 CSV + manifest。各 CSV は 512MiB 未満。旧 7 CSV は使わない。非 NULL `medical_record_id` 重複なし |
| 観測 | old_db の HAC-CSV-1 は未。KNJO 破損により formal TRUSTED は別 blocker。現行 rehearsal 経路でも八王子 run が無い |

### 2. H0-3b / H1-2

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 八王子 bundle の隔離確認とローカル rehearsal |
| 受け入れ | `CLINIC_CODE=hachioji` で `make old-db-handoff-check` が PASS。必要な場合だけ preflight / apply / verify |
| blocked-by | H0-2 |

### 3. AE-STG-UAT-LANE3-HAC

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | maintenance window で八王子 bundle を共有 STG へ投入する |
| blocked-by | H0-2、H0-3b / H1-2 |

### 4. H3-9 staff attach apply

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 移行 staffs.id へ account を後付けする |
| 受け入れ | `make stg-uat-staff-attach-preflight` → `make stg-uat-staff-attach` が成功する |
| 観測 | 入力ファイルはある。apply 成功の証跡はない |

### 5. H3-11 画面確認

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 対象医院へ切り替え、飼主検索が実データで表示されることを確認する |
| 注意 | 行値はログへ残さない |

### 6. Lane 4

| 項目 | 内容 |
|------|------|
| 実行者 | 医院スタッフ / USER |
| 目的 | 必須4業務（検索、受付、カルテ、会計）を両院の STG で連続5営業日再現する |
| 受け入れ | 現場が切替可と判断する。上限8週。local UAT の PASS を Lane 4 完了としない |
| blocked-by | Lane 3 両院（八王子が未） |

---

## STG実データ運用テスト

### 目的

共有 STG に旧システムの医院別データを投入し、医院スタッフが現行業務と並行して新システムを検証します。

- 現行業務の正本は旧システム。
- STG への入力は検証用であり、本番へ移しません。
- 対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。
- 必須業務は検索、受付、カルテ、会計。
- 両院投入後、必須4業務を連続5営業日再現し、現場が切替可と判断したら第2段階へ進みます。上限は8週です。
- 本番 cutover は別工程です。入力は移行日の旧システム出力とし、`PASS` / `TRUSTED_CANDIDATE` 契約を緩めません。

### データ境界

- `002_master`: 医院骨格と参照マスタのみ。accounts や臨床行を含めない。
- `003_demo` / `004_staging`: 退役済み。STG 実データ運用には使わない。
- 業務データ: old_db の医院別 21 CSV と manifest を正本とする。
- `_old_db_handoff/`: ローカル隔離専用。Git、CI、イメージ、通常 migration へ載せない。
- 同一 STG DB で医院ごとに 10M ID band を分ける。
- STG の並行登録データを本番へコピー、昇格、差分追加しない。

### 禁止事項

- デモ臨床と実データを混在させない。
- `pscale connect` で remote DB を localhost に見せかけない。
- 本番用 F6 へ `--allow-local-rehearsal` を流用しない。
- 医院間で band、staff、account、患者、飼主、カルテ、会計を越境させない。
- 共有 STG への投入中に同じ STG で業務入力を続けない。
- PHI や資格情報を標準ログ、Git、Issue、チャットへ出さない。

---

## Lane 0 — 八王子入力

- [ ] **H0-2 / HAC-CSV-1**: 医院 identity 付き 21 CSV と manifest を old_db から出す。各 CSV が 512MiB 未満。非 NULL `medical_record_id` 重複なし。
- [ ] **H0-3b**: `old-db-handoff-check` を通す。manifest なしの旧 `hachioji/` や旧 7 CSV は使わない。

```text
backend/migrations/seeds/_old_db_handoff/hachioji/
```

---

## Lane 1 — 八王子のローカル証明

- [ ] **H1-2**: 八王子 bundle で preflight、apply、verify をローカル rehearsal する。
- [ ] **H1-5**: ローカルで失敗する bundle を共有 STG へ送らない。

必要な場合だけ:

- [ ] **H1-3**: [A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md)
- [ ] **H1-4**: [F8_G4_FAILURE_REHEARSAL.md](docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md)。本番 CSV は渡さない。

---

## Lane 3 — 共有 STG（USER のみ）

[STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊境界に従う。城東 CSV は投入済み。八王子は bundle 到着後の maintenance window。

残チェック:

- [ ] **H3-9**: staff attach の preflight → apply。名簿と secrets は gitignored・mode 0600。
- [ ] **H3-11**: 対象医院へ切り替え、飼主検索が実データで表示される。行値はログへ残さない。
- [ ] **H3-7 八王子**: `make stg-uat-import` または `make stg-uat-handoff`。失敗時は成功側を残す。

作業票が repo に無いもの（未実施とは限らない）:

- H3-1 他エンジニア確認と医院への開始範囲連絡
- H3-2 検証済み full backup と復元担当

予約も検証する場合だけ:

- [ ] **H3-10 / AE-STG-UAT-SHIFT**: 未来日の shift。古い絶対日付をコピーしない。

### 実行契約

- `make stg-uat-import` / `make stg-uat-handoff`
- 手動 fallback: `make stg-uat-csv-import-preflight` / `make stg-uat-csv-import` / `make stg-uat-csv-import-verify`
- `make stg-uat-staff-attach-preflight` / `make stg-uat-staff-attach`

具体的な接続値は repo へ記録しない。

---

## Lane 4 — 第1段階の並行運用

- [ ] 期間中は STG DB を reset しない。
- [ ] backend デプロイ前に医院へ日時を知らせる。適用済み migration 編集や seed 差替えを行わない。
- [ ] STG で新規作成した予約、カルテ、会計を本番へ移さない。
- [ ] STG 障害時も旧システムで現行業務を継続する。
- [ ] ログへ PHI が出た場合は検証を止めて修正する。業務監査の正本は `audit_logs` とする。
- [ ] 製品不具合は `bug.md` へ記録し、Linear Issue 化する。
- [ ] 必須4業務を両院で連続5営業日再現し、現場の切替判断を記録する。

---

## 第2段階へ進む条件

- 両院の bundle 投入と verify が完了している。
- 対象スタッフが自医院へログインできる。
- 必須4業務を連続5営業日再現している。
- ブロッキングな製品バグが残っていない。
- STG 入力を本番へ移さないことを医院が理解している。
- 本番 F6 の `PASS` / `TRUSTED_CANDIDATE` 契約が維持されている。
- 移行日、当日の旧データ出力、本番 run sheet が決まっている。

---

## 次の一手

1. old_db で **HAC-CSV-1** を出す。両院 Lane 4 のブロッカー。
2. 城東で STG ログインするなら **H3-9 apply** の成否を残し、**H3-11** で飼主検索だけ確認する（行値は残さない）。
3. 八王子 bundle 到着後に Lane 0/1/3。
4. 両院で Lane 4。local UAT を Lane 4 完了としない。

---

## 参照

| 文書 | 役割 |
|------|------|
| [docs/ops/deploy/CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) | F6 21表 cutover |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | `_old_db_handoff` のローカル隔離 |
| [docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md](docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) | seed と 21表の境界 |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 再作成、直結、破壊境界 |
| [docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md](docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md) | STG データのライフサイクル |
| [docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) | staff account 運用 |
| [docs/ops/deploy/A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) | 隔離画面確認 |
| [docs/ops/deploy/LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) | ローカル reset |
| [docs/ops/infra/staging/runbook.md](docs/ops/infra/staging/runbook.md) | STG 障害初動 |
