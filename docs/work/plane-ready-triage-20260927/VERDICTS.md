# PLANE-READY-TRIAGE-20260927 — Verdicts

- Campaign `plane-ready-triage-20260927` / Unit `PLANE-READY-TRIAGE-20260927` / Attempt `ATT-PLANE-READY-1`
- Baseline `dd10ca777` (clean tree). Snapshot: `.planning/agent-fast-campaign/plane-ready-triage-20260927/sources/plane-targets.json`, sha256 `850fc532…cd7f`, 65 tickets (Needs Human 46 / Blocked 19).
- Machine-readable twin: `verdicts.json` (schema `plane-ready-verdicts/v1`, embeds `write_plan`). No Plane mutation performed by this receiver.

## Decision rule (binding)

`verdict=ready` のみ: 次に必要なstepがrepo/local scope内でagent実行可能で、未解決の外部入力(STG/PRODアクセス・実機・未到着artifact/roster/bundle・臨床/PO判断・人間承認・UAT/go-live判断・他ticket依存・credential/環境操作・外部受領)が皆無の場合。それ以外は `keep_needs_human`(ゲート=特定人間の判断/行動/供給) または `keep_blocked`(ゲート=依存/外部前提)。`ambiguous: true` = keep判定で不足入力を `missing_input` に明記。**keep_* の suffix はゲート種別の分類であって現stateの維持確認ではない** — 種別と現stateが一致しない15件には write_plan で差分コメントを付ける(本planが許可する遷移は Ready のみのため実state変更はしない)。

## Summary

| Verdict | Count |
|---|---|
| ready | 0 |
| keep_needs_human | 45 |
| keep_blocked | 20 |
| ambiguous (subset) | 12 |

ready=0: 全65件の 2026-09-26 着手プラン(またはLINEAR-FULL繰越verdict)が外部ゲートを宣言しており、repo/local scope内で完結する次stepを持つticketは存在しなかった。

## write_plan (inside verdicts.json)

- action=`set_state_ready`: **0 件** — ready 判定なし
- action=`comment`: **18 件** — verdict が最新commentに情報を追加するticketのみ:
  - Needs Human→Blocked相当の差分: EMR-46, EMR-44, EMR-43, EMR-40, EMR-38, EMR-131, EMR-130, EMR-129
  - Blocked→Needs Human相当の差分: EMR-149, EMR-147, EMR-144, EMR-139, EMR-133, EMR-132, EMR-108
  - ambiguous gate 指摘(差分なし): EMR-200, EMR-60, EMR-127

## Verdict table (grouped ready-first; zero ready, so keep_needs_human then keep_blocked)

| identifier | verdict | gate_kind | reason (one line) |
|---|---|---|---|
| EMR-202 | keep_needs_human | operator_approval | 運用者が再抽選実施を承認・実行(wrangler.jsonc:141はdefault固定で乖離は既解消) |
| EMR-169 | keep_needs_human | clinical_input | 尿検査項目・皮膚/耳詳細・基準値は医師(PO)提示待ち、STG投入はユーザー実行 |
| EMR-166 | keep_needs_human | po_decision | 未対応4項目の採否はユーザーのみ決定可能 |
| EMR-164 | keep_needs_human | po_decision | アニコム連携の今期スコープ採否+医院回答確定は人間判断 |
| EMR-161 | keep_needs_human | user_approval | 訂正文承認とスプシ反映はユーザー(実装実測: login 5/分・PW reset 30分) |
| EMR-151 | keep_needs_human | scheduling | 医院との研修日程・形式・対象職種の合意と実施receipt待ち |
| EMR-150 | keep_needs_human | external_facts | U1–U12全行未記入(DELIVERY_PACKAGE.md実測)、契約事実は契約責任者入力 |
| EMR-149 | keep_needs_human | human_decision | remote apply方式の採用決定(ユーザー)+名簿/個人メール方針供給 (ambiguous) |
| EMR-147 | keep_needs_human | ops_execution | 4系統credential rotationの担当者決定・承認・実行は人間専権 |
| EMR-144 | keep_needs_human | external_input | repo外roster/secrets供給+staff/clinic/role対応表承認 (ambiguous) |
| EMR-139 | keep_needs_human | human_decision | 対象env(STG/本番)と操作者の確定は人間判断(照合pre-stepのみagent可) (ambiguous) |
| EMR-133 | keep_needs_human | external_input | 実LINEテストアカウント供給+実行承認 (ambiguous) |
| EMR-132 | keep_needs_human | user_approval | write有効LSTEP対象・対象者・復旧範囲のユーザー承認 (ambiguous) |
| EMR-117 | keep_needs_human | human_decision | 窓口スプシ一本化で本照合完了とするかのユーザー判断 |
| EMR-116 | keep_needs_human | physical_device | 発生iPad/PCの機種・OS・ブラウザ版と同条件PC情報の供給待ち |
| EMR-114 | keep_needs_human | human_action | 医院に元操作をSTGで再実行してもらうことが前提 |
| EMR-113 | keep_needs_human | external_input | 医院×機器の接続情報・担当確定+実機受信試験許可 |
| EMR-112 | keep_needs_human | scheduling | 医院側UAT受入者・実施日程の決定 |
| EMR-111 | keep_needs_human | external_input | 現行名簿(repo外受領)・個人メール方針・権限方針の供給 (ambiguous) |
| EMR-110 | keep_needs_human | human_uat | 医院ごとの受入担当・実施枠(EMR-112連動)+合成fixture承認 |
| EMR-109 | keep_needs_human | human_uat | 受入担当・実施日・実プリンタ確定+会計/締め運用承認 |
| EMR-108 | keep_needs_human | ops_execution | 対象範囲・backup/監査/復旧の確定+限定訂正承認(ops責任者) |
| EMR-107 | keep_needs_human | read_approval | 対象医院・期間・読取範囲の決定+STG DB読取承認 |
| EMR-106 | keep_needs_human | read_approval | 範囲決定+STG読取承認(species完全一致はvaccine_repository.go:42で確認済) |
| EMR-105 | keep_needs_human | external_input | QA/医院による発生画面・対象版・エラー文言の採取(またはSTG再現許可) |
| EMR-104 | keep_needs_human | external_input | QAが実端末/ブラウザ/回線/行数帯/IME/対象buildを確定 |
| EMR-99 | keep_needs_human | human_uat | EMR-177の医院UAT結果待ち(死亡ガード実装は存在) |
| EMR-97 | keep_needs_human | clinical_input | 証明書の用途・項目・発行者/日付・用紙見本は臨床PO提示 |
| EMR-88 | keep_needs_human | human_decision | 山梨依頼者へ閲覧のみで足りるか受付機能まで必要かの確認 |
| EMR-86 | keep_needs_human | clinical_input | 各院尿検査項目・単位・基準値表の臨床PO提示+origin表示裁定 |
| EMR-51 | keep_needs_human | ops_execution | 4系統credential rotation+revocationはUSER/セキュリティ責任者専権 |
| EMR-49 | keep_needs_human | clinical_input | E-1 bundle全列の臨床記入待ち(NEEDS_CLINICAL) |
| EMR-48 | keep_needs_human | clinical_input | 健診seed項目・単位・rangeの臨床確認+PO決裁(NEEDS_CLINICAL) |
| EMR-47 | keep_needs_human | clinical_input | 項目×種×測定系rangeとワクチン適合mappingの出典付き臨床承認 |
| EMR-45 | keep_needs_human | ops_execution | 全院締め時間config投入はUSER実行(#299はmerge済) |
| EMR-42 | keep_needs_human | external_input | rosterの安全経路確認+email/院/役割→permission_group対応表の明示(PO/運用) (ambiguous) |
| EMR-41 | keep_needs_human | human_decision | U13完了/未完の1語回答はユーザー待ち(U13_status=未完) |
| EMR-39 | keep_needs_human | external_facts | U1–U8/U10/U11の所有・契約・窓口・保持方針は契約責任者記入 |
| EMR-37 | keep_needs_human | po_decision | GH#261節A各SDの削除/一本化/修正/見送り裁定はPO/臨床(B/CはEMR-49/48依存) |
| EMR-36 | keep_needs_human | physical_device | iPhone/Android/iPad 3実機+QA環境担当の試験URL/認証受領手順待ち |
| EMR-33 | keep_needs_human | read_approval | STG/PROD legacy件数のread-only観測は承認窓+DB運用担当実行(secret値非読取) |
| EMR-23 | keep_needs_human | ops_execution | named env非破壊migrateはUSER専権・agent禁止明示 |
| EMR-22 | keep_needs_human | human_uat | H1実LINE・H2実token LIFF・P1 migrate・H3/H5/H7は全てUSER実行 |
| EMR-14 | keep_needs_human | physical_device | NX600/AU10V実機UAT+token供給経路/署名/notarization配布gate |
| EMR-10 | keep_needs_human | value_measurement | 業務責任者のDnD価値実測4点+go/no-go判断(実装着手禁止) |
| EMR-200 | keep_blocked | env_precondition | sweep実行は稼働スタック+合成fixture write前提、スタック起動/停止・db直接操作はagent禁止で自己調達不可 (ambiguous) |
| EMR-148 | keep_blocked | external_precondition | Cloudflare/DB/R2/DNS本番契約+GH Environment reviewer未到(backend-deploy.ymlにproduction経路自体なし) |
| EMR-146 | keep_blocked | dependency | Lane3 verify(MIG-22)+H3-11(EMR-145)未完了+運用開始日/担当未定 |
| EMR-145 | keep_blocked | dependency | EMR-144(H3-9)の対象確定・receipt受領が前提 |
| EMR-141 | keep_blocked | dependency | EMR-140完了+テスト宛先・回数・受信担当承認が前提 |
| EMR-140 | keep_blocked | dependency | EMR-139(PREFLIGHT)完了+付与承認が前提 |
| EMR-134 | keep_blocked | dependency | P1–P4(EMR-147/148/MIG-23/EMR-131)全完了+切替window/判断者確定が前提 |
| EMR-131 | keep_blocked | dependency | E1/E2(EMR-132/133)+医院UAT(EMR-109/110)未完了(証拠mappingはagent可pre-step) (ambiguous) |
| EMR-130 | keep_blocked | dependency | P1/P2/P3/P5/P6/P7+P4/P8/E1/E2 receipt未揃い(コメント0件、全上流依存) |
| EMR-129 | keep_blocked | dependency | 医院別manifest・Lane3 verify・H3-11・5営業日証拠未揃い(コメント0件) |
| EMR-128 | keep_blocked | dependency | EMR-60未完了+--clinical再実行のスタック操作はユーザー専権(前回7FAIL原因解消証跡なし) |
| EMR-127 | keep_blocked | env_precondition | 使い捨てclinic+権限別account fixture未整備+スタック操作はユーザー実行 (ambiguous) |
| EMR-115 | keep_blocked | unarrived_artifact | No.61本文・期待値・原票PDF未到着(資料なし実装禁止) |
| EMR-60 | keep_blocked | env_precondition | 対象8件の使い捨てenv再実行証跡待ち(page-object/見出し整合はbaselineで完了済) (ambiguous) |
| EMR-46 | keep_blocked | unarrived_artifact | 正式COMPLETE producer bundle(HAC-INPUT-2/八王子)未到着(WAIT_EXTERNAL) |
| EMR-44 | keep_blocked | external_precondition | PROD環境未構築(HOLD、backend production deploy経路自体が不存在) |
| EMR-43 | keep_blocked | dependency | UAT人間レーン(EMR-22/BRT-68)実施+close checklist完了の依存(HOLD) |
| EMR-40 | keep_blocked | dependency | 複数gate receipt(#89/#97/#250/#252-255)未揃い+window/判断者未定 (ambiguous) |
| EMR-38 | keep_blocked | external_precondition | Lステップ先方API enable未到(WAIT_EXTERNAL、dual gateでfail-closed維持) |
| EMR-24 | keep_blocked | superseded | M1–M5へ分離済みの廃止メタticket、残存子(EMR-49/41/39等)が個別未解決 |

## Evidence verification notes

判定には3系統の並行read-only検証(backend code / frontend code / docs+workflows)と独立verdict草案の相互照合を実施。主な検証結果:

- `backend/wrangler.jsonc:141` = `scheduling_policy:"default"`(EMR-213経緯でregional非対応) — EMR-202の「config=regional」記述はstale。
- `frontend/e2e/pages/*` page objectsは現行UI見出しと一致(カルテ登録 - ペット選択 等) — EMR-60の「見出しstale」はbaselineで解消済み。
- `backend/checkup-packages/` の尿検査項目・皮膚/耳詳細・基準値は医師確認待ちの空殻 — EMR-169/48の臨床ゲートを裏付け。
- `docs/work/remaining-campaign-20260920/SLACK-INTAKE.md` は不存在、実在は `docs/work/todo-campaign-20260919-ready17/SLACK-INTAKE.md`。`bug-2.md`/`todo-operations.md` 名称は不存在で実体は `bug.md`/`todo.md`(#verification-ledger)。
- `.github/workflows/backend-deploy.yml` には production trigger も `environment:` gate も存在しない。`.github/workflows/e2e.yml` は `e2e/auth-flows.spec.ts` のみ実行。
- `backend/internal/auth/http_types.go:128` = login rate limit 5回/分、`password_reset_service.go:24` = PW reset 30分、`Makefile:370/391` = stg-uat-staff-attach系、`backend/internal/medicalrecord/lab_device_receive_service.go:54` 等、引証した repo path/symbol は全てbaselineで存在確認済み。
