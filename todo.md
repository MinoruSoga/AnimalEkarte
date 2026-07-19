# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-17 (3)（台帳整理: 消化済み記述を圧縮、USER アクションを優先順の単一表 U-1〜U-13 へ統合、各行に完了条件と完了後のエージェント作業を明記。更新履歴は git log -- todo.md を正とし本ヘッダに積まない）
> **push・外部書き込み・credential 変更はユーザー所有アクション。**（PR マージはユーザーが手動で行う。本台帳には載せない）
> **別台帳**: 今フェーズでやらないもの = `phase2.html` / BE 保留詳細 = `BE-pending.md` / PO 判断キュー・決裁記録の正本 = `q&a.html`
> **本書の役割**: 今フェーズに着手可能・着手すべきタスクのみの正本台帳。読者 = 次に着手するエージェント（前提知識ゼロで本書だけから作業に入れる粒度で書く）。

---

## 運用規約

### Docker 検証規約（BE・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

### 台帳スコープ規則

- 本書には**今フェーズで着手可能なタスクのみ**を記載する。対応済みは削除する（記録は git 履歴）。
- **「対応済み」判定は gh の state を実測してから行う**（`gh issue view <n> --json state`）。ローカルの消化記録だけを根拠に除去しない。除去後に再オープンされても本書は自力で気付けないため、**除去は gh CLOSED の確認とセット**とする（2026-07-16、#201 を「B2 消化済み」で除去した後に [SAFETY] として再オープンされ、臨床安全の最優先ゲートが 1 日間どの台帳からも不可視になった実例）。
- 今フェーズでやらないもの（次期監査引き継ぎ・再開条件付き見送り・長期目標・PO 決裁済み「やらない」）は `phase2.html` を正本とする。決裁済み「やらない」は実装着手禁止のまま。着手判断が出たら実装単位として本書へ戻し、phase2.html から削除する（二重管理禁止）。
- **open Issue の正本は gh**（`gh issue list --state open`）。本書に重複掲出しない。列挙は必ずドリフトするため番号リストを本書に持たない（2026-07-17: 旧列挙が #250〜#262 を落としていたため撤去）。3セッション並行開発計画（#260）は別対応・本台帳スコープ外。
- 例外 = 「P1 — Open Issues（台帳掲載分）」。**臨床安全・credential 露出など、gh に埋もれると人命・情報漏洩に直結するものだけ**を本書へ掲出する。掲出判断は内容の重大性で行い、件数を増やさない。
- PR マージ判断・マージ状態・マージ用チェックリストは本台帳に載せない（ユーザー手動）。
- PO 決裁の正本は `q&a.html`。本書には「決裁済みで着手可能な作業」だけを USER アクション表・P1 表の形で持つ。
- 着手保留・任意検証の BE 詳細は `BE-pending.md`。再検討トリガが立ったら実装単位として本書へ戻す。

---

## P1 — Open Issues（台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #201 | **[SAFETY] 薬量上限超過の物理ブロックと例外統制** | **再オープン（2026-07-16）・USER BLOCKED（U-1）**。自動計算・snapshot・master UI・監査の基盤は実装済みだが、上限超過を FE `ConfirmDialog`「この数量で保存」で通過でき BE も拒否しない（audit のみで保存）。確認ダイアログを安全統制に使う禁止則（product-philosophy ③）違反・fail-closed 不成立。ブロッカー = U-1 個人責任者ゲート。承認後の実装内容は #201 本文「必須仕様 1〜6」が正本（BE 物理 reject／権限付き例外フロー＝専用 permission・必須理由・同一 tx audit・欠落時 rollback／情報欠落時の挙動確定）。doc 側は `docs/spec/screens/06-medical-records-form.md` §2 に既知の是正対象として明記済み。**本表から除去禁止**（除去→再オープン不可視の前歴あり） |
| #211 | 検査・健診パッケージ化 | **コード・seed・migration は全て完了。残 = DB 適用（U-3）と臨床確認のみ**。完了分: A1+A2（アドプリット削除＋尿比重 min/max 空化・commit 90553a51）／A6 複合 FK（`001_init.sql` の 013 セクションへ統合済み・独立 migration ファイルは存在しない）。既存 DB は 001 の checksum が変わったため `DB_RESET=true` 再適用が必要（fresh DB は 001 適用だけで有効）。やらない分（CRUD UI／四季分割・腎臓ドック／select 異常ハイライト／ライブ E2E）は `phase2.html`。provisional seed の臨床確認はクライアント臨床責任者の回答待ち（回答後 seed 手動更新） |
| #89/#97/#98/#99/#109 | シークレット移行・ローテーション | **USER BLOCKED（U-2）**。リポジトリ側 Phase A は完了済み。残りは credential 実操作: 4系統ローテーション／P5-2 `gh secret set`／#97 Issue 本文マスク／#109 は `STG_DEMO_*` 登録後にフォールバック撤去（撤去作業はエージェント可）。#98/#99 は Phase 8 まで PENDING。Issue クローズはローテ完了後 |

---

## USER アクション一覧（優先順）

> 上から順に消化を推奨。「完了後のエージェント作業」列は、完了の合図をもらったエージェントが**次に何をすべきか**の指示。

| ID | アクション（USER 実施） | 完了条件 | 完了後のエージェント作業 |
|----|------------------------|---------|------------------------|
| U-1 | **#201 [SAFETY] 個人責任者ゲート**: 使用医院の臨床責任者（個人名）を確定し、①絶対上限 ②warning 範囲 ③体重/species/パラメータ欠落時の手動入力可否 ④緊急時例外フローの要否、を承認させる | 承認内容（4点）が #201 に記録される | #201 本文「必須仕様 1〜6」に従い BE 物理 reject＋（④が要なら）権限付き例外フローを実装。`computeDoseGate`（FE）と `backend/internal/service/` の dose 系が対象。実装後 06 doc 更新 |
| U-2 | **SEC-SECRETS-5**: 4系統ローテーション＋ P5-2 `gh secret set`＋ #97 本文マスク。手順 = runbook §0.5 / `infra/cloudflare/README.md` | 4系統の新 credential が有効・旧値無効化・GitHub Secrets 登録済み | ① gitleaks baseline 方針の実装（task.html P1-1）② `STG_DEMO_*` 登録済みなら #109 Phase C: performance-tests のフォールバック撤去 |
| U-3 | **`DB_RESET=true` 再適用（ローカル→STG）**: 001 checksum 変更（#211 A6 統合）＋ seed 変更 2 件（A1+A2 = 90553a51／GAP-2 閲覧専用ロール = 13c6a93a）をまとめて反映。DB reset はエージェント実行禁止 | 両環境で reset 完了・起動 green | ① S02 権限差・S05 の再実行（U-6 の前提解除）。**BUG-403 は 2026-07-17 に stale DB のまま実測・修正・TDD 検証済みでクローズ済み（環境起因ではなく実装欠落と確定）— 本項目の対象から除外**。BUG-404 は commit 58c653df で修正済みだが「次回シナリオ再実行で最終確認」待ちのため bug.md の直近クローズ節に残置中 |
| U-4 | **SD-9 被害判定 SQL を STG/本番で実行**（SQL 本文は下の「個別タスク詳細」）。`9b6a01ed` の修正は新規作成院にしか効かないため既存データの手動確認が必要 | 0 行 → クローズ。ヒット → 結果をエージェントへ共有 | ヒット時: 対象グループへのルールバックフィル（`defaultPermissionRuleTable` 準拠の UPDATE 文起草＋適用手順書）を別タスク起票。`assigned_staff > 0` の行を最優先（該当スタッフが全機能ロックアウト中） |
| U-5 | **SD-14 STG 実機検証**: LINE 紐付け E2E。飼主フォーム（04 画面）で紐付け URL 発行 → LINE で開く → LIFF 遷移 → 紐付け完了まで | 紐付け成功が確認できる | 失敗時: 症状を聞いて `line_link_service.go`／`frontend/liff` の LiffLinkPage を調査・修正（2e4808b5 が直近の修正） |
| U-6 | **受け入れシナリオ実行（続き）**: 第1バッチ S01〜S06 は 2026-07-17 実施済み（30 PASS / 4 FAIL / 13 BLOCKED — レポート = `docs/ops/testing/scenarios/reports/2026-07-17-local.md`）。FAIL 由来の BUG-401〜407 のうち BUG-402/403/406 は修正・検証済み、BUG-401 は BUG-408 へ統合（機能設計待ち）、BUG-405 は仕様未定の機能欠落と確定（FEAT-CHECKIN へ移設）、BUG-407 は対応中。**残 = ①S07〜S12＋V01〜V05 ②U-3 後の再測分（S02 権限差・S05 の残 FAIL）③S02 境界値の再確認** | 全シナリオ実施・要実測 87 件（残 78）に実測値が入る | 実測結果の【要実測】昇格とシナリオ乖離修正（第1バッチ分 4 件は昇格済み） |
| U-7 | Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` の確認/設定 | Vercel ダッシュボードで確認済み | なし（FE は `__VERCEL_ENV__ !== "production"` 一次ガード実装済みのため二重防御） |
| U-8 | `terraform apply`（P2 internal ALB + VPC Origin）。`infra/terraform/terraform.tfvars` はローカル準備済み（gitignore 対象） | apply 成功・疎通確認 | なし |
| U-9 | **ADR-003（PO-006）独立 Issue 起票**（起票操作は USER 専権） | Issue 番号が確定 | 案 1B（TRIGGER）＋支払方法の二重保持解消を実装 |
| U-10 | **Sentry ベンダ確定・課金契約（PO-002）** | ベンダ・DSN が確定 | Phase 1 実装: FE 未捕捉例外＋ErrorBoundary 通知のみ。送信は例外スタック＋リリース版数に限定・PII off・security-review 必須（決裁全文 = q&a.html PO-002） |
| U-11 | **FEAT-searchable-select 目視確認**: 検索・スクロール・選択・カスケード・per-option disabled（対象 15 箇所の一覧は下の「個別タスク詳細」） | 全対象で挙動 OK | NG 箇所があれば個別修正。全 OK なら本項目を台帳から削除 |
| U-12 | [任意] `docker compose exec frontend pnpm type-check` を実行し `use-reception-kanban.ts` の型エラー再現有無を確認（フル type-check は USER 実行コマンド） | エラー再現なし → クローズ | 再現時: エラー全文を受けて修正 |
| U-13 | [任意] Notion EkarteSprint 文字化け 3 語（します／共有済み／事前提供）の目視確認。対象ページ: クレジット訂正フロー／検査④機器データ取込／検査⑥自動連携調査 | 3 ページの該当文が正常 | なし（クローズのみ） |

---

## バグは別台帳

> 未修正バグの正本 = リポジトリ直下 [`bug.md`](bug.md)（BUG-401〜407 は 2026-07-17 に移設。同日中に 401→408 統合・402/403/406 修正・405→本書 FEAT-CHECKIN へ移設・407 対応中・新規 408/409/410 起票）。本書はバグ以外の対応事項のみを扱う。

---

## 個別タスク詳細

### FEAT-CHECKIN: 入院チェックイン（reserved→admitted）導線の新規追加

- **背景**: bug.md BUG-405 として起票されたが、`docs/spec/screens/07/08/09` を全文突合した結果、reserved→admitted への遷移導線は仕様上も実装上も一切存在しない（discharged への遷移＝退院処理のみ定義）ため「バグ」ではなく未着手の機能と判定し本書へ移設（2026-07-17）。
- **現状**: 入院登録は常に status=reserved で作成される（create form が status を送信せず backend default 頼み）。一覧の 入院中/予約 タブは status enum を直接 1:1 マップしているだけ（date 判定なし）なので、チェックイン導線を追加しない限り開始日当日でも予約タブに残り続ける。
- **要件責任者（個人名）の確認が必要**（product-philosophy①）。
- **検討事項（要決裁）**: (a) 開始日到来で自動的に reserved→admitted へ遷移（要: 停止手段・失敗通知・監査ログ — product-philosophy⑤）か、(b) 詳細画面/一覧に明示的な「チェックイン」ボタンを追加し手動で PATCH status=admitted を呼ぶか。まず①要件を疑う（自動でも良いのでは）から検討する。
- **副作用**: `HospitalizationDetailActions.tsx` の「退院処理」ボタンは `status !== 退院済` で表示されるため reserved でも表示され「実質アクティブに見える」混乱を招く。実装時にボタン表示条件も見直すこと。

### U-4: SD-9 被害判定 SQL

```sql
SELECT pg.id, pg.clinic_id, pg.name, COUNT(DISTINCT sp.staff_id) AS assigned_staff
FROM permission_groups pg
LEFT JOIN permission_group_rules r ON r.group_id = pg.id AND r.deleted_at IS NULL
LEFT JOIN staff_permission_groups sp ON sp.group_id = pg.id
WHERE pg.deleted_at IS NULL
GROUP BY pg.id, pg.clinic_id, pg.name
HAVING COUNT(r.id) = 0
ORDER BY pg.clinic_id, pg.id;
```

背景: 新規院開設時のデフォルト権限グループが従来ルール 0 件で作成されており（SD-9・修正 `9b6a01ed`）、ルール 0 件のグループ所属スタッフは全機能アクセス不能になる。修正は新規作成にしか効かないため、既存レコードの確認が必要。

### U-11: FEAT-searchable-select 確認対象（実装完了・目視のみ残）

SearchableSelect 本体 = `frontend/src/components/ui/searchable-select.tsx`。適用済み 15 箇所:
予約区分・担当者(`ReservationFormFields.tsx:334,416`)、診断名1/2+カテゴリ(`DiagnosisHeaderDiagnosis.tsx:52,58,64`)、診療計画病名(`ClinicalPlanSection.tsx:47`)、主訴(`InterviewChiefComplaint.tsx:45`)、ワクチン(`VaccinationForm.tsx:72`)、検査種別・担当医(`ExaminationFormFields.tsx:56,63`)、健診種別・担当医(`CheckupForm.tsx:111,143`)、入院ケージ(`HospitalizationBasicInfo.tsx:106`)、薬剤親カテゴリ(`MedicineSidePanelSections.tsx:67`)、指名フィルタ(`ReceptionFilterPanel.tsx:59`)、医師フィルタ(`ReservationManagementCalendar.tsx:85`)、動物種(`NewOwnerInlineForm.tsx:83`/`PetIdentitySection.tsx:95` in `frontend/src/features/owners/components/` — 2026-07-11 FE-R14 分割 commit `52bd37816` で `PetEditModalFieldSections.tsx` から移設・2026-07-19 再監査で参照先を訂正)、スタッフフィルタ(`ShiftCalendar.tsx:107`・per-option `disabled`)。

意図的スキップ（対応不要）: `ShiftFormDialog` テンプレ選択（非制御アクショントリガー）／`ReservationTypeSidePanel` グループ選択（カラードット custom JSX・実件数<15）。保留候補: Lステップ TriggerType（`LstepDeliveryMonitorPageParts.tsx:71`）。

### CSV import（フル seed 運用・USER）

> 方針（2026-07-15 確定）: フル 003_demo（~529MB・PHI 含みうる）は **Git に載せない**。正本バックアップ = `old_db/sensitive-local/animalekarte-003-demo-full/`。リポジトリの `003_demo` は小さいデモのまま。

- [ ] **USER:** ローカルでフル seed を使う場合: `rsync -a ../old_db/sensitive-local/animalekarte-003-demo-full/ backend/migrations/seeds/003_demo/` のあと `make reset`（エージェントは reset しない）。誤 `git add` 防止のため該当 CSV に `skip-worktree` 推奨。
- [ ] **USER:** STG へのフル seed 適用は別途承認・手動実行（通常は小さいデモのまま）。

---

## 完了済みトピックの参照先（本書には残さない）

- **画面仕様書全数突合 SD-1〜19＋GAP-1/2**: 2026-07-17 に全 21 件消化（実装コミット 142f5ebe〜6d10f4c0）。決裁と乖離裁定の正本 = `q&a.html` 各カード、GitHub 入口 = #261。副産物（BE finalized ガード欠落 5 件封殺・liff.state 復元前読取修正・dbOrTx allowlist 回帰）も同コミット群で対処済み。次期送り 6 件 = `phase2.html`「SD/GAP fix all の副産物」節。
- **docs/spec/screens**: 全 65 doc が実装と同期済み（drift-gate green・画面数 40）。以後は実装コミットに doc 更新を同梱する運用。
- PR #186 レビュー残 = `task.html`（Codex 未 Resolve 4 スレッド＋値投入 1 件）。

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| `bug.md` | 未修正バグの正本台帳 |
| `phase2.html` | 今フェーズでやらないものの正本（次期監査引き継ぎ・見送り・長期目標・やらない決裁） |
| `BE-pending.md` | 着手保留・次期送り・任意検証の BE 詳細 |
| `q&a.html` | 内部 PO 判断キュー（決裁記録の正本。PO-001〜008・SD-1〜19・GAP-1/2 回答済み） |
| `task.html` | PR #186 Codex レビュー残タスク（Open スレッド＋値投入待ち） |
| `BE-refactor.md` | **BE9 active plan**: Go/Gin公式原則を適用し、巨大なhandler/service/repositoryをdomain/resource packageへ大規模移行する。package非依存lint・HTTP/security review・旧layer撤去までを含む。旧BE8固定layer計画はsuperseded historyとして実行禁止 |

> 旧 `BE_todo.md` は本ファイルへ吸収済み（削除）。`FE-refactor.md`（FE 第7期計画・ESLint 境界ガード新設 FE7-0〜2 + FE7-3/4 判断）は 2026-07-18 に対応完了・削除済み。恒久規約は `frontend/CLAUDE.md` へ同梱済み。BE8の固定layer計画は2026-07-19にADR-005でsupersededとなり、backend正本を `.claude/rules/go-gin-backend-guidelines.md` へ置換。`BE-refactor.md` は同日、実装・lint・reviewを新規約へ移すBE9 active planへ更新した。旧 `docs/tasks/`・`docs/archive/` は2026-07-16に廃止（詳細はgit履歴）。
