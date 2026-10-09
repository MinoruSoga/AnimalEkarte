# 非破壊セキュリティ検証・修復の記録（EMR-332）

Campaign: emr-332-security-defense / Unit: U07-REVERIFY-CONSOLIDATE / Attempt: att-u07-1
Recorded by: receiver-emr332-u07-01 / Recorded at: 2026-10-09
Consolidated ticket revision: **EMR-332 @ 2026-10-08T17:54:31Z** (Plane `baritechllc`/EMR, work item `a6f8d3c9-4b34-4bda-8cdf-82ee9ec359b6` — campaign `sources/emr-332-ticket-body.txt`)
初回Campaign base revision: `de28e9765a29091339a872a85f1875a947e07e5b` (origin/main at campaign start; per-case worktree HEADs は `verdict-matrix.md` の target revision 列参照)

## 現在状態 (Codex follow-up)

[repair-20261009.md](repair-20261009.md)が後続修復と限定検証の記録。C10の実handlerへ内容形式検証を配線し、C05/C09の過剰な影響・重大度主張を訂正した。元campaign COMPLETE/G2/G3は初回監査の履歴であり、この追加diffの検証・受入を証明しない。staging実測は引き続き0。

検証専用コードの現行配置は `backend/tests/security/`。内容別の7 packageへ整理し、比較decoderも `_test.go` に変更した。移動対応表・現行コマンド・同一試験の確認結果は [test-layout-20261009.md](test-layout-20261009.md) を参照。通常の共有ファイルhandler回帰試験は `backend/internal/lstep/shared_file_content_test.go` に置く。

docsも内容別の `baseline`、`identity-access`、`input-output`、`browser`、`egress`、`uploads`、`dependencies`、`workflow-audit`、`assessment` へ整理した。U01–U07のUnit/Attempt/Receiver表記は初回作業の履歴識別子であり、現行ファイル配置には使用しない。過去のコマンドと生ログの本文は保持する。追加の全体配置調査と修復は [test-layout-20261009.md](test-layout-20261009.md) の後続記録を参照。

## Deliverables (U07 acceptance artifacts)

| Doc | 内容 |
|-----|------|
| `docs/work/emr-332-security/verdict-matrix.md` | C01–C18 全 18 行の判定マトリクス (ASVS/Top10/STRIDE, revision, method, normal/refusal, **defense/detection 別列**, side-effect calls, evidence ref, unverified scope) |
| `docs/work/emr-332-security/remediation-plan.md` | AC4 改善計画 — C05要件差分・C09強化候補・C10適用修復 + 記録済み候補 (R1–R4, RC-1, RC-1..4, U04/U06 観察) |
| `docs/work/emr-332-security/assessment/reverification.md` | 初回適用修復ゼロの履歴 + C05比較decoderの初回証跡（現行は `tests/security/inputvalidation/strict_bind_helper_test.go`） |
| `docs/work/emr-332-security/assessment/residual-known-issues.md` | 残件・既知問題の一覧 (record-only) |
| 内容別フォルダ（上記の現行配置） | ケース別の判定記録と証跡。baselineおよび各検証内容のフォルダを参照 |

## Assurance boundary — 4 状態の分離 (AC6)

本キャンペーンの記録は以下の **4 つの状態を混同しない**。いずれのケースも下の状態を上の状態として主張しない (ticket 本文「文書完成、模擬合格、ステージング実測、本番受入を別状態として報告する」)。

| 状態 | 意味 | 本キャンペーンでの到達 |
|------|------|----------------------|
| **文書完成** (documented) | 検証計画・境界・判定レコードが文書として完成 | **達成** — U01 safety baseline + U02–U06 全 18 ケースの判定レコード + 本 U07 統合成果物 |
| **模擬合格** (simulated pass) | メモリ・模擬・httptest・静的レビューの範囲で期待防御/検出を観測 | **部分的に達成** — 防御 PASS 16 (限定) / FAIL 1 (C05要件差分) / UNKNOWN 1 (C09) の記録は static+simulated の範囲のもの (各 case doc 参照) |
| **ステージング実測** (staging-observed) | 所有 staging origin への境界内実送信による観測 | **ゼロ — 実測セットは空 (empty)**。live 送信 **0/20 requests**。理由: (a) staging 稼働リビジョン照合が undetermined (baseline/safety-record §2; repo 内に deploy 記録/Worker version stamp/既知差分証跡なし — `browser/c08-csrf.md` 補足の照合試行記録); (b) 専用テストアカウント未払出 (§7); (c) 許可テナント識別子未プロビジョン (§3) |
| **本番受入** (release acceptance) | 人間受入による本番相当の承認 | **未着手 — 対象外**。Done 遷移は人間受入または明示指示がある場合のみ (AC7/EMR332-X3)。キャンペーンの目標状態は Review まで |

**この index と配下の全 verdict は「模擬合格」までの記録であり、ステージング実測・本番受入の証明ではない。**

## Staging 稼働リビジョンの状態

- **undetermined** — staging が serve する revision と task branch 内容の照合は未達成。照合の 3 経路 (deploy 記録 / Worker バージョン識別 / 既知差分証跡) すべてで repo 内証拠なし (`browser/c08-csrf.md` 「稼働リビジョン照合の試行記録」)。
- チケット記載のローカル参照 HEAD `40384aa3bb92eaa3ea950d56a9a0bf604ff08f2c` と staging 稼働 revision の一致も起票時点で未確認。
- したがって staging-observed legs は全ケース **BLOCKED** (baseline/safety-record §2 規則)。

## Residual items (残件 — verbatim 列挙)

1. **staging-observed legs 全件 BLOCKED** — live 送信 0/20 requests、稼働リビジョン undetermined (全 18 ケースの staging-observed 行)。
2. **real-DB legs 未実行** — `TestRealDB_*` 系 (repository filter 実効果)、CAS/advisory-lock の実 SQL 挙動 (`MarkNoShowAt` FOR UPDATE / guarded check-in CAS / `pg_try_advisory_xact_lock` / `PromoteTentativeOnCheckin`/`AbsorbTentativeOnBooking`)、audit repository 層 (`repository_db_test.go`, `repository_real_ddl_test.go`)、`TestAutoCreateFromReservation_ConcurrentSameClinicPetJSTDayCreatesOneRecord` — いずれも Compose `db` Postgres 前提で本安全条件下では未実行。
3. **C07 render leg BLOCKED** — React render→DOM 実測 (vitest + jsdom/happy-dom) は node_modules 不在・共有 compose 禁止・task-isolated container 未供給のため未実施。
4. **C16 multi-instance 集約制限 UNKNOWN** — `RateLimitStore` は process-local map (`rate_limit.go:29-33`); 複数プロセス/インスタンス間の集約制限は deploy 形態依存で未検証。
5. **C18 通知経路到達性 UNKNOWN** — `ReservationNotifier` 実装の配送確認証跡なし (send 呼出のみ確認)。
6. **C14 脆弱性状態 NOT ASSESSED** — `go vuln`/`pnpm audit`/Dependabot/OSV 照合未実行 (network/data 依存)。pinning/lockfile/provenance の構成判定 PASS と既知 CVE 露出は別事項。
7. **baseline/source-inventory.md:46 stale cite** — `repository_db_test.go:33-36` と記載されているが `baseline/test-classification.md` §1 では `:31-34`。**record-only — 当該ファイルは U01 の owned file であり本ユニット (U07) では編集しない**。
8. **後続修復と候補** — C10 RC-1は適用・メモリ検証済み。C05は対象API/互換性/impact評価待ちで全体strict化は未採用。C09 R1–R4, C14 RC-1..4, U04 edge-preflight/headers observations (edge preflight への最小 security header、edge/Gin 複製 drift contract test、frontend `_headers` `/*` の明示 Cache-Control), U06 `auth_staff_validation_fail_open` mislabel rename 候補。詳細は `remediation-plan.md`。

## Unverified scope (未検証範囲の集約)

- 認証済み live チェック全般 (専用アカウント・許可テナント未プロビジョン) — C01–C04, C13, C17 等の認証前提 live 検証。
- staging/production 実測全般 (上記 staging-observed 空集合)。
- 実 DB/redis への障害注入・ロック/スコープ実効果 (上記 real-DB legs)。
- Worker edge 層の実転送挙動・`/_internal/*` 実 404・TLS handshake 実測・cookie 属性の live 観測 (login 禁止により恒久的 static-only)。
- GORM fluent `Where("col = ?", v)` 数千 site の個別列挙 (C06 は Raw/Exec/Sprintf/LIKE リスククラス 119 site のみ網羅)、pgx `Query`/`QueryRow` 系の grep 非対称 (csvimport 27 site は補助確認)。
- 臨床 API 応答の field-level 過剰露出網羅監査 (C13 は公開面/エラー面の代表サンプルのみ)。
- 他 package の任意 `slog` 呼出における secret/PII 非含有の網羅 (C18 残件 UNKNOWN)。
- panic recovery・network timeout・外部 LINE API 実接続・負荷/閾値探索 (禁止条件により未実施)。

## Required human actions (人間操作が必要な残件)

本キャンペーンの BLOCKED/UNKNOWN を解消し staging-observed・本番受入へ進めるには以下の人間操作が必要:

| # | 必要な人間操作 | 解除する残件 |
|---|----------------|--------------|
| 1 | **専用テストアカウントの払い出し** (検証専用 staff/LIFF アカウント — 共有デモパスワードは不使用) | 認証済み live チェック全般 (baseline/safety-record §7) |
| 2 | **単一の許可テナント識別子のプロビジョン** (検証専用 staging clinic ID の記録) | 認証テナントスコープ live チェック (§3) |
| 3 | **test-DB 環境の供給** (本安全条件で実行可能な DB 経路 — Compose `db` または隔離 test DB の明示許可) | real-DB legs (TestRealDB_*, CAS/advisory-lock, audit repo) |
| 4 | **isolated frontend container の供給** (node_modules を持つ task-isolated な render test 環境) | C07 render leg BLOCKED |
| 5 | **deployment topology の確定** (Worker→Container のインスタンス構成・複数インスタンス有無) | C16 multi-instance 集約制限 UNKNOWN、C15 advisory-lock 到達確認 |
| 6 | **通知配送証跡の提示** (ReservationNotifier 等の既存配送観測証跡) | C18 通知経路到達性 UNKNOWN |
| 7 | **staging 稼働リビジョンの照合手段** (deploy 記録/Worker version の提示または照合の実施) | staging-observed legs 全件 BLOCKED (§2 規則) |
| 8 | **人間受入** — 本記録のレビューと明示指示がない限り Plane `Done` 遷移は行わない (AC7) | 本番受入状態 |

## 安全・逸脱の記録

- 後続修復の検証はネットワーク無効Docker・メモリmockのみ。実サービス/DB/storage/LINEへ接続せず、本番・実ユーザー・PII・実テナントへの試験は行っていない。初回unitの副作用列はstatic/mockの範囲であり、実環境全体の副作用ゼロを証明しない。
- live 送信: **0/20 requests** (上限内で実送信ゼロ)。
- 証跡内に secrets/tokens/PII/実ペイロードなし — マーカーは `SECCHK20261009*` 形の英数字 sentinel のみが各 case doc で引用される。
- 文書訂正の lineage: u03 docs における att-u03-1→att-u03-2 訂正履歴は `input-output/c05-input-validation.md`・`input-output/c06-sql-inventory.md` 末尾の訂正履歴表に記録 (verdict 不変)。
