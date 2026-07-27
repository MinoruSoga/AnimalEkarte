# AnimalEkarte — TODO

> 更新: 2026-07-27(6)（TASK-251 = U3 完了 `65a0dd08d`・U2b-full 第1段完了 `7a64d9e63` まで履歴反映済み。残 = U2b-full 第2段〔進行中・別 session〕と U4。統合済み001を `DB_RESET=true` で再構築適用 → `make codegen` → full type-check → runtime確認が USER/CI 手順〔TASK-251 / TASK-ADR003 entry 参照〕。⚠commit は必ず `git commit -- <paths>` で path 制限）

## 運用

- 本書は、エージェントが直ちに着手できる未完了タスクの台帳とする。
- タスクは「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で追加する。
- 対応済みセクションは削除し、完了記録はgit履歴と各実装testを正本とする。
- GitHub Issueと対応するタスクはIssueのstateを実測し、Issue一覧を本書へ重複掲出しない。
- release/運用gateは実装タスクと混在させず、[`q&a.html` OPS-13〜17](q&a.html#ops)と該当runbookで追跡する。

## 正本の境界

| 内容 | 正本 |
|------|------|
| 着手可能な実装タスク | 本書の「個別タスク詳細」 |
| GitHub Issueのstate・一覧 | GitHub Issues |
| BE9構造移行・進捗・release gate | BE9は2026-07-24にcode complete（release pending）。境界は[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) / [boundary map](docs/architecture/be9-2a-boundary-map.md)、release gateは[`q&a.html` OPS-13〜17](q&a.html#ops) |
| FEデザイン準拠・リファクタリング計画 | [`FE-refactor.md`](FE-refactor.md) |
| 今フェーズで着手しない事項 | [`phase2.html`](phase2.html) |
| 着手保留・任意検証のBE技術債 | [`BE-pending.md`](BE-pending.md) |
| PO判断・USER実操作・P0ブロッカー | [`q&a.html`](q&a.html) |
| Issueを3セッションで着手するためのガイドview（正本=各Issue・受け入れ条件を複製しない） | [`3-session-agent.html`](3-session-agent.html)（削除・退役の対象にしない） |

## 個別タスク詳細

### TASK-ADR003: 予約⇔会計の支払方法二重保持解消（ADR-003 案1B）

- PO-006／DEC-9 裁定済み。**実装は完了**（旧migration 006の複合FK〔現行所在は統合済み001末尾の旧006アーカイブブロック〕とmodel inventory登録 = `c434c4e66`）。完了証跡・詳細な検証記録の正本 = git 履歴（本 entry の旧全文は `b2897a409` 以前）。USER が Issue 起票したら本エントリを Issue へ移設し二重掲出しない。
- 残作業:
  - **USER/CI 手順（未実施）**: 既存データを事前照合 → 統合済み `001_init.sql` を `DB_RESET=true` 相当の承認済み再構築で適用（統合前001はchecksum mismatchになるため必須）→ `make codegen` → full type-check → payment splitのclinic複合FKをruntime確認（[q&a.html OPS-2](q&a.html) で追跡。DB操作は実装taskではない）。
  - **未防御の `payments`**: `clinic_id` 列が無く複合 FK を張れない。follow-up 2案（billings join 照合 / `payments.clinic_id` 追加+backfill）は未裁定・未実施。裁定は SEC-DUR-01（tenant/provenance graph の横断方針）と併せて行う。

### TASK-251: 締め集計 category contract 確定実装（#251・12分類）

- contract 正本 = #251 本文（DEC-21・**USER 本人裁定** 2026-07-25・原文 = q&a.html git 履歴）。完了 unit の証跡正本 = git 履歴と各実装 test（本 entry には残作業のみを置く。旧全文 = `b2897a409` 以前の本 entry 履歴）。
- **完了済み unit（hash のみ）**: U1 category resolver 単一ソース化 `2154dc9de` ／ U2a 物販 master 由来導出 `33690944e`+`24d4ae2fd`（BUG-436 解消） ／ U2b-min 手入力12分類必須 Select `b8a29d3e7` ／ U2b-full 第1段 other 理由必須化+旧migration 009（現行001内）`7a64d9e63` ／ U5 締め集計12値 fail-closed+`AllItemCategories` `581d6b87c` ／ U3 ワクチン接種 pull 型導出+旧migration 008（現行001内）`65a0dd08d`（第2走の trigger 案は coordinator 裁定で宣言的構成へ簡素化。派生 = SEC-DUR-01・BUG-440）。
- **残 unit**:
  - **U2b-full 第2段** = 締め表「未分類・要確認」件数表示（DEC-16③残・別 session が Codex 発行済み・進行中）。
  - **U4** = ④ training の新規 source 設計（DEC-21・現状 source 不在のため設計から。hotel 側は U1 の `HospitalizationTypeHotel` 経路で到達済み）。
- **USER/CI 手順（未実施）**: 旧008/009を末尾へ統合済みの `001_init.sql` を `DB_RESET=true` 相当の承認済み再構築で適用（統合前001はchecksum mismatchになるため必須）→ `make codegen` → full type-check → ワクチン候補→取込→削除→再候補化の runtime 確認。
- #247（月次統合表）は本 TASK の contract 完了後に着手。

### SEC-DUR-01: provenance FK graphのdurable enforcement方針（全FK共通・横断）

- 第2走reviewer所見: provenance link後に`pets.owner_id`/`pets.clinic_id`、`owners.clinic_id`、`medical_records.owner_id`/`medical_records.pet_id`/`medical_records.clinic_id`、各source master（`vaccines.clinic_id`等）が変更されると、Vaccination/Treatment/Merchandise等の既存provenance graphが後から不整合になり得る。
- 個別sourceだけへtriggerを追加するのではなく、全provenance FKに共通するarchitecture/PO論点として、service依存チェック・宣言的FK・許可された親mutationの再相関/移送をどの層で保証するかを決める。飼主変更（`pets.owner_id`）等の実機能を恒久ブロックしないことを必須条件とする。
- **着手 = 納品後・architecture/PO裁定待ち**。U3はlink時のapp検証・transaction lock・宣言的clinic FKをdelivery boundaryとして完了し、本entryの実装は開始しない。

### BUG-440: vaccination claim解放のimmutable actor監査

- HIGH（独立security review）。U3の要求どおり明細soft-delete時に`vaccination_id`/内部`clinic_id`をNULL化して再候補化するため、削除済み`billing_items`行だけでは「どの接種claimを誰が解放したか」を復元できない。通常ログはdurable actor auditの代替にならない。
- 対応はclaim作成・解放を同一transaction内のimmutable audit/historyへactor・対象vaccination・理由付きで保存し、audit失敗時に本体writeもrollbackする設計とする。completed/cancelled会計の削除防御はU3 repair round 1で実装済み。
- **着手 = 納品後・audit schema/actor contractのarchitecture裁定後**。U3のexact delete→再候補化contractを維持し、本unitでは監査schema/APIをdrive-by追加しない。

### SEC-SWEEP-02: grandchild FKの親相関掃引 + 同型欠陥のstatic lint新設

- SEC-SWEEP-01（掃引 `3321c801f`、修正 `c16c011f2`＋`0736cd6f9` で **完了**）が明示的に繰り越した2件。掃引はpet直下の子テーブルreadのみを対象としており、孫テーブルは別クラスとして未着手。
- **① grandchild相関の掃引（read-only調査）**: `daily_records` / `care_logs` / `exam_results` / `billing_items` / `medical_record_images` / `medical_record_addenda` 等、petの孫にあたる表のreadが中間表を経由した親clinic相関を持つかをschema-firstで全数確認する。SEC-SWEEP-01と同じ手順（schema universe → model mapping → consumer discovery → state分類）を適用する。
- **② static lintの新設**: 同型欠陥（`pet_id`等の単一FKに対し親clinic相関を欠くread）をraw SQLとGORMの双方で検出するlintを既存6 lintの隣へ追加する。BUG-429の1件もSEC-SWEEP-01の9件も人手監査でしか見つかっていない。機械で止めない限り同じクラスが再発する。
- 修正の参照実装 = `backend/internal/pet/chronic_condition_repository.go:37,52`。相関にpets側の `deleted_at` / `deceased_at` を含めない制約は本掃引にも適用する（含めるとsoft-delete済み・死亡ペットの履歴が黙って消える挙動回帰になる）。
- **第1走完了（7/27・`864a886b7`・Mode 3 照合済み）**: ①census 28表を独立reconcile付きで全数分類 ②lint `grandchild_parent_clinic_correlation_lint_test.go` をRED→GREENで新設（9対象+raw SQL） ③修復 = checkup_field_results / prescriptions / treatment_plans（直読+hospitalization側count・ambiguous column修正込み）。並行sessionが daily_records / medical_record_addenda 修復+別lint `pet_grandchild_...` を同時作成（当該session側で清算・**lint 2本の統合が必要**）。
- **残（CRITICAL・修復予算超過/legacy契約裁定要で第1走が正しくBLOCKED返し）**: `appointment_trimming_details`（trimming_repository.go:30）／`billings`（accounting_repository.go:192,280）／`estimates`（estimate_repository.go:43）／`medical_records` appointment edge（medical_record_repository.go:375）／`vital_records`（vital_repository.go:40）。HIGH: staff_repository.go:329 の addenda/vitals 依存count。vital→accounting→estimate は既存legacy testが「親をsanitizeしつつ壊れた子は可視のまま」を要求し3-strikeでBLOCKED — 修復には既存契約の裁定が先。lintのraw SQL検出は `*`/`id`/`alias.id` 形のみで一般化も残。
- 出典: SEC-SWEEP-01実行結果のcalibration / follow-up節（2026-07-25・掃引完了に伴い本エントリへ移設）。

### BUG-441: lintscan inventory gate 2本がmainで赤（未登録8件・3 unit由来）

- HIGH（gate red）。`docker compose exec backend go test ./internal/lintscan/ -count=1` が `TestDBOrTxInventory_MatchesAllowlist`（6件）と `TestMasterFKWriteInventory_AllowlistMatchesRealSource`（2件）でFAILする（2026-07-27実測・migration統合とは無関係のbaseline-red）。
- 内訳と由来: ① `auth/permission_group_repository.go` の `CreateWithRules`/`UpdateWithRules`/`replaceRules`（FE12-02 P-A `99bac632e`） ② `billing/billing_item_repository.go` の `ValidateVaccinationCreateReference`（U3 `65a0dd08d` — ambientTxParticipationExpectations未登録も併発） ③ `medicalrecord/exam_type_repository.go` の `FindByID` と `examinationService.Create/Update` の `ExamTypeFieldID` DTO追加（#249系）。
- 対応: **ゲートを黙らせるだけのallowlist追記は禁止**。各methodのambient-tx参加をtx atomicity/isolation testで実証した上で登録する（gateのerror文言が要求する手順どおり）。②はU3由来なので優先。①③は各変更ownerのreview必須。
- 出典: migration統合unitのMode 3照合（2026-07-27・coordinator実測）。統合executorはbaseline-redとして正しく切り分け済み。

### BUG-430: stage-importの医院非限定DELETE

- CRITICAL。`backend/cmd/stage-import` のdeleteScopeが `owner_id >= 300000`（pets経由の継承含む）でclinic_id非限定。実行すると他院の高番ownerデータを削除し得る。`backend/cmd/stage-import/main_test.go:217-246` がこの挙動をテストで固定化している（cross-clinic保護テストは無い）。
- 対応方針変更（2026-07-25 DEC-20）: **stage-import退役で解消**（deleteScope修正には投資しない）。本番cutoverはrunbook既定の21表csv-import正式経路であり本ツールは本番使用禁止・local限定＋`--confirm-local-destroy`ガード既存。退役実装=cmd/stage-import削除またはビルド除外（#250再基準化転記とセット・USER承認後）。
- 出典: #251調査 Completion Report（2026-07-25）。テスト実測で確認済み。

### BUG-433: 生成FE型がGoドメインモデル由来のため、応答DTOに無いフィールドが型上は存在扱いになる

- HIGH（サイレント機能不全の生成器）。**S3/S2いずれにも属さない横断課題**。`frontend/src/types/generated/models.ts` は tygo が `backend/internal/model/` から生成しており（同ファイル冒頭コメント）、OpenAPI／応答DTOからではない。このため FE の型は *Goドメインモデル* を写し、HTTP が実際に返す *応答DTO* とは一致しない。DTOに無いフィールドは実行時 `undefined` なのに型検査は通る。
- 実害の実例: BUG-431（受付の危険度バッジが実APIで一度も点灯しなかった・`463e07424` で修正）は本ドリフトの1インスタンスに過ぎない。fixtureは型どおり作られるためテストでも検出されない。
- 実測された残存ギャップ: 生成 `Pet` は31プロパティ、修正後の予約pet DTOは9。残22フィールド（`clinic_id` `owner_id` `animal_species_id` `name_kana` `gender` `birth_date` `color` `blood_type` `microchip_number` `neutered_date` `acquisition_type` `food` `environment` `phone` `last_visit` `insurance_id` `remarks` `deceased_at` `deceased_reason` `created_at` `updated_at` `insurance`）は型上は利用可能だがワイヤに存在しない。他モデル（Owner/Reservation等）も同構造。
- 対応方針（未確定・要判断）: ①応答DTOからFE型を生成する経路へ切り替える ②生成型を「ドメインモデル」と明示リネームし、画面が使う型は応答DTO由来へ分離する ③現状維持で個別に埋める（BUG-431と同じ対症）。①②は生成基盤の変更を伴うため納品後が妥当。納品前は、新規に生成型のフィールドへ依存する実装を書くときに**そのフィールドが応答DTOに実在するかを都度確認する**運用で凌ぐ。
- 出典: BUG-431 修正時に判明したNew Work（2026-07-25・executorが残22フィールドを実測列挙）。

### BUG-435: 生成FE型が陳腐化したままmainに乗っていた（codegen-check未励行）

- MEDIUM。FE12-07 で `make codegen` を回したところ、意図した型mapping修正（`any` 17→0）とは無関係な追随差分が同時に出た: audit定数7件（`AuditActionAuthPasswordChange` / `Reset` / `AdminReplace`、`AuditActionTrimmingCreate` / `Update` / `Delete`、`AuditResourceAccount` / `AuditResourceTrimming`）と `TokenBlacklist` のdoc comment。つまり **Go model を変更した際に `make codegen` が回されず、`frontend/src/types/generated/models.ts` が陳腐化した状態でmainに乗っていた**。
- `Makefile:349` に `codegen-check: codegen` + `git diff --exit-code frontend/src/types/generated/` が存在し、本来これがCIで検出する。検出されなかった理由（CIに未配線／配線済みだが未実行／実行され無視された）は未確認。
- 実害: FEが存在しない定数を参照しても型検査が通らないだけで安全側だが、逆に**BEが追加した定数をFEが使えない**状態が黙って続く。BUG-433（生成型がドメインモデル由来で応答DTOと不一致）とは別問題であり、そちらは構造の誤り、本件は同期の欠落。
- **条件修正完了（7/27・`213e210b4`・Mode 3 照合済み）**: 旧 `if:` が main 向け PR で job をスキップしていたのを paths-filter 単独条件へ簡素化。main 直接更新・main 向け PR の双方で fail するようになった。
- **残余（運用論点・実装taskではない）**: 歴史的経緯の実測で前提が反転 — 7/25 のドリフト（`dad69bc6a`）は**直接更新の CI run 30087212418 が検出し fail していた**。真の欠陥は「失敗した main CI が到達を防げず、誰にも届かなかった」こと（main 直 push 運用では CI は事後検知であり、branch protection は PR にのみ効く）。main CI 失敗の人間向け通知経路（サイレント障害の封鎖）を OPS として裁定要。
- 出典: FE12-07 実行時に判明したNew Work（2026-07-25・`git diff frontend/src/types/generated/models.ts` 実測）。FE-refactor.md 範囲外のためこちらへ記載。

### BUG-437: `ExamTypeField` に `ClinicID` を追加したが read 側 clinic-scope registry へ未登録（cross-tenant read IDOR の再発リスク）

- HIGH（clinic 分離）。`b4d10e083`（#249 Phase 2 U1）で main に入っている（検出当時はゲート FAIL・現在は照合ゲート green だが機械的再発防止が未完のまま — 下記参照）。同 commit が `backend/internal/model/examination_type.go:32` の `ExamTypeField` に `ClinicID uint64` を追加した。write 側 registry には既に登録済み（`backend/internal/lintscan/master_fk_write_inventory_lint_test.go:129` の `"ExamTypeFieldID": "ExamTypeField (sub-master of ExaminationType, #124)"`）だが、**read 側の `clinicScopedMasterAssoc`（`backend/internal/lintscan/preload_clinic_scope_lint_test.go`）に entry が無く、`masterModelReadWriteExemptions` の例外にも入っていない**。
- 実害: clinic-scoped master を FK 値で Preload する際に `clinic_id` 述語が無いと、汚染された FK（#124/#125 の write 側検証ギャップ、または是正前データ）に対して**他院の master 名・値が返る**。b3638d5e で手作業修正した cross-tenant read IDOR と同型である。
- 検出経路: BE10-2 B5（`b28c4a105`）が `TestMasterModelReconciliation_RealSourceIsConsistent` を復旧した直後に検出した。**B4（`c430072d8`）以降このゲートは `no such file or directory` で死んでおり、B5 が直すまでこの追加は無検出で通過する状態だった。**
- 対応（3択・当該変更の owner が選ぶ）: ① Preload が既に存在するなら `clinicScopedMasterAssoc` へ association 名を登録する ② `internal/model` に association field が無いなら `masterModelReadWriteExemptions` へ理由付きで追加する ③ `ClinicID` の追加を撤回する。**ゲートを黙らせる目的だけの ② は禁止**。
- 再現: `docker compose exec backend go test ./internal/lintscan/ -run TestMasterModelReconciliation_RealSourceIsConsistent -count=1`（2026-07-25 の検出時実測。R-4 = `9ca93f249`・BUG-438 = `296ea7bb7` の解消後、`internal/lintscan` は green — 下記の部分対応も参照）。
- **部分対応・BLOCKED（2026-07-25・working tree）**: ①登録を選択し、read 側 registry に一意な association 名 `"ExamTypeField": "ExamTypeField"` を追加した。`Items` は `ExamResult` / `BillingItem` / `EstimateItem` / `LabExamResultItem` にも使われる汎用名のため登録していない。`ExamTypeField` へ解決する既存の `Preload("Items")` 2件（`backend/internal/medicalrecord/exam_type_repository.go:37,46`）は既に `"clinic_id = ?"` を持ち、DB-backed cross-tenant test も green なので現行 runtime read は安全。`masterModelReadWriteExemptions` と write 側 `clinicScopedMasterFKField` は無変更。
- 変更ファイル: `backend/internal/lintscan/preload_clinic_scope_lint_test.go`, `todo.md`。検証: 対象照合ゲート = `ok   github.com/animal-ekarte/backend/internal/lintscan 0.008s`、read lint（登録直後・最終）= `ok   github.com/animal-ekarte/backend/internal/lintscan 0.352s` / `0.355s`、DB-backed cross-tenant test = `ok   github.com/animal-ekarte/backend/internal/medicalrecord 0.260s`、`gofmt -l` = stdout 0 行、`go vet ./internal/lintscan/` = exit 0。
- 独立 clinic-isolation / code review で HIGH 1件が残った。read lint は Preload path の末尾 association 名だけで registry を引き、未登録名を無視する（`backend/internal/lintscan/preload_clinic_scope_lint_test.go:245-256`）。live read は `Items` なので、この2件から `clinic_id` を削除しても lint は検出しない。`Items` の一律登録は上記の同名別モデルを誤検出するため不可。解消には親 model / site を識別する context-aware rule または association 構造変更が必要で、いずれも本 unit の registry 登録予算を超える。したがって照合ゲートは green だが、機械的な再発防止が未完成のため BUG-437 は未解消。
- `internal/lintscan` package は 2026-07-25 時点で **green**（R-4 = `9ca93f249` の stale sentinel 是正、BUG-438 = `296ea7bb7` の CASCADE allowlist 登録で残 FAIL 2件が解消済み）。したがって本件は「ゲートが赤い」問題ではなく、**live read 経路 `Preload("Items")` を機械的に守る手段が無い**という再発防止の穴として残っている。
- 出典: BE10-2 B5 の Mode 3 照合（2026-07-25・生成側が独立実測）。

2026-07-23に起票したBUG-421〜428、TEST-ROUTES-01、FMT-BE-01は2026-07-24のBE9実装でsource/testへ反映済みのため、本active listから削除した。release pending項目（fresh DB migration、remote CI/coverage、production deploy/ops rehearsal）は実装taskではないため本書へ再掲しない。
