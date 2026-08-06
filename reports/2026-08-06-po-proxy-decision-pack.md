# PO 代理 Decision Pack — residual HOLD ゲート裁定

> 日付: 2026-08-06 / 起草: PO 決裁代理（Opus）
> 入力: `STATUS.md` §1/§2 · `q&a.html`（DEC-40〜67 · DR-CLINICAL · DISPOSITION-20260803）· `reports/2026-07-31-task-021-phase2-slice2.md` · `reports/2026-07-31-r05-single-sot-phase-b.md`
> 実測: open Issue = 18（`gh issue list` 実行）· claim/* = 0（local/remote とも）
> **本 pack は代理裁定である。USER（実 PO）は全件を後から覆せる。覆す場合は新カード再起票 + 実装巻き戻し。**

---

## 0. Executive summary

- **今日決められること**: TASK-021 の破壊削除は **段階承認**（DTO/OpenAPI までは external 証拠なしでも安全に進められるが、route DROP と migrate DROP は external inventory 待ち）。LINE-R05 の production DROP は **3 条件未達で HOLD**、DROP 対象は `line_reservation_setting.line_channel_secret` と特定済み。#261 は #201 bundle 参照のみで閉じる方針を維持。#256 は DEC-61 no-rewrite 維持。#98 は残余リスク受容で close 可、#99 は #253 一本化で close 可。
- **臨床待ちで決められないこと**: TASK-033（救急投薬 cutover）は #201 bundle 1 行が空欄のため **実装着手そのものを禁止**（例外なし）。#249 の追加 unit 起票も臨床 range 未承認のため不可。#211 の実 row も同様。上限値・warning 帯・動物種 default は本 pack では一切決めない。
- **重要な事実訂正**: `DISPOSITION-20260803` は TASK-377 を `READY_AGENT` としているが、**TASK-377 は既に landed**（commit `e6796be62`「docs: record TASK-377 completion」/ `treatment_dose_deviation_reason_test.go` 他 10 ファイル実在）。D-033 の「TASK-377 と TASK-033 の順序」は **争点として消滅**している。
- **同様に**: `e6796be62` は TASK-032 を「claim 解放待ちの READY_AGENT」としていたが、実測で `claim/*` = 0。この blocker は解消済み。
- **結論として agent が次に触ってよい unit**: **1 件のみ**（D-021-A）。それ以外は全て USER/ops または臨床待ち。

---

## 1. Decision table

| Decision ID | 対象 | Verdict | 発効条件 | 次の実装 unit（agent） | USER/ops 実行 | リスク |
|---|---|---|---|---|---|---|
| **D-021-A** | TASK-021 request DTO / OpenAPI property 削除 | `APPROVE_WITH_CONSTRAINTS` | 即時。write は既に 400 reject 済（slice2）で意味的に dead | **UNIT-021-A**（唯一の READY_AGENT） | なし（merge のみ） | 低 — reject 済み入口の除去。挙動不変 |
| **D-021-B** | response `excluded_courses` 削除 | `HOLD` | external access log / client registry で consumer ゼロが報告されること | なし | external inventory 取得 | 中 — 外部 reader を無言破壊し得る |
| **D-021-C** | master exclusion route DROP (`GET\|PUT .../excluded-reservation-types`) | `HOLD` | D-021-B と同条件 **＋** capable-only 入口への一本化設計完了 | なし | 同上 | 高 — 現 live な inverse write 入口 |
| **D-021-D** | DB migrate DROP（exclusion table/column） | `HOLD` | D-021-B/C 完了 + STG 適用 green + backup + rollback owner | なし | `make migrate`（USER 専権） | 高 — 不可逆 |
| **D-033** | TASK-033 救急投薬 cutover 実装着手 | `NEEDS_CLINICAL` | #201 bundle 1 行の全列記入 | なし | 臨床責任者の bundle 記入 | 臨床安全 — 例外なし |
| **D-033-P** | 上限 / warning / 救急記録 policy | `NEEDS_CLINICAL` | 同上（値は代理裁定しない） | なし | 同上 | 同上 |
| **D-033-O** | TASK-377 と TASK-033 の順序 | `APPROVE`（争点消滅） | TASK-377 は landed 済み。順序問題は発生しない | なし | なし | なし |
| **D-LINE-R05** | production rollout + column DROP | `HOLD` | ①clinic 別 legacy presence inventory ゼロ ②composition test を Phase B 契約へ更新 ③STG 適用 green + backup + rollback owner | なし（②は解除後に unit 化可） | inventory 取得 · migrate | 高 — 不可逆 + 認証系 |
| **D-211** | 健診 package clinic import | `NEEDS_CLINICAL` + `NEEDS_USER_OPS` | DR-CLINICAL 実 row 承認 **かつ** DR-OPS の対象 clinic/環境確定 | なし（synthetic は TASK-374 で実装済） | 実 manifest 確定 · apply | 中 — clinic 隔離・FK |
| **D-249** | 検査機能 追加 unit 起票 | `HOLD` | 臨床 range 値の承認（DR-CLINICAL #249 行） | なし | 臨床責任者記入 | 中 — 臨床判定値 |
| **D-249-EXT** | 外部機器 / 自動化連携 | `DEFER_PHASE2` | 即時（live residual から外す） | なし | phase2.html へ記載 | 低 |
| **D-261** | 臨床安全・画面仕様ギャップ PO | `APPROVE`（方針再確認） | 即時。#201 bundle 参照のみで close。値の複製禁止 | なし | 結果 enum + opaque ref 一行記入 | 低 |
| **D-256** | 操作マニュアル / 研修 | `APPROVE`（DEC-61 維持） | 即時 | なし | 非機密一行記録 | 低 |
| **D-256-024** | TASK-024 screenshot / FAQ sign-off | `APPROVE_WITH_CONSTRAINTS`（必須残） | 即時。#256 close の必須 gate として維持 | なし | Privacy/Repository owner の sign-off | 低 |
| **D-257** | Go-live window | `HOLD`（DEC-60 維持） | 全 gate green（#89/#97/#98/#99/#250/#253/#254/#255） | なし | 新 window 設定（USER 専権） | 高 |
| **D-098** | 旧 RDS credential 残余リスク | `ACCEPT_RESIDUAL_RISK` | provider 側 rotation 完了の非機密確認 + close 理由一行 | なし | provider 確認 · close | 中 |
| **D-099** | 旧 ECS deploy 経路 | `APPROVE`（#253 一本化） | provider に実行可能経路なしの確認一行 | なし | provider 確認 · close | 低 |
| **D-250** | 旧 Access 移行 | `HOLD live` | producer bundle 受領 | なし | 先方待ち | 中 |
| **D-259** | L ステップ Write API | `HOLD live` | 先方 enable | なし | 先方待ち | 低 |
| **D-284** | line-reserve 実機フォント | `DEFER_PHASE2` | 試験環境 + 3 実機の到着が今フェーズ内に見込めない | なし | 実機受領後に再開 | 低 |
| **D-scope** | residual live 範囲 | `APPROVE` | D-249-EXT / D-284 を live から除外。他は維持 | — | STATUS.md 更新 | 低 |

---

## 2. Per-decision cards

### D-021-A — request DTO / OpenAPI property 削除

1. **Question**: external consumer 未確認のまま、既に 400 reject 済みの `excluded_type_ids` request property を削除してよいか。
2. **Options considered**: (A) external 確認まで全面 HOLD / (B) 4 ステップ一括 CLEAN-GO / (C) 段階承認：write 側だけ先行。
3. **Verdict — `APPROVE_WITH_CONSTRAINTS`**: slice2 で Create は非空で 400、Update は存在するだけで 400 になっている。つまりこの property を送る外部 consumer は **既に全員壊れている**。property 削除で新たに壊れる利用者は存在しない（「送れば通る」状態が既に無い）。したがって external access log は **この 1 ステップに限り前提ではない**。read 側（response / route）とは risk 構造が根本的に異なる。
4. **Rejected alternatives**: (A) は既に dead な入口の除去まで止めており、削除目標（原則②）に対し過剰保守。(B) は response/route/DROP を巻き込み、未確認の read consumer を破壊するため却下。
5. **Constraints / non-goals**: response `excluded_courses` は触らない。master exclusion route は触らない。migrate は行わない。repo inverse facade `UpdateExcludedReservationTypes` は Create empty seed 用に維持。
6. **Evidence / refs**: `reports/2026-07-31-task-021-phase2-slice2.md` §1.3・§2.1（Create/Update reject 実装）· 同 §6-2（次 slice の順序）· 実測 `backend/internal/reservation/reservation_staff_request.go`。
7. **STATUS.md update line**:
   `| TASK-021 | exclusion 破壊削除 | USER+PO | 部分 APPROVE（DEC-68: request DTO/OpenAPI のみ）· response/route/DROP は HOLD |`
8. **Issue comment draft**: 「TASK-021: request `excluded_type_ids` は slice2 で既に 400 reject 済みのため、property 削除は挙動不変として承認。response `excluded_courses` と master exclusion route の削除、および migrate DROP は external inventory 未報告のため引き続き HOLD。」
9. **Unblock checklist**: なし（即 READY）。

### D-021-B / C / D — response・route・migrate DROP

1. **Question**: read 側 surface（response field / route / DB column）を external 証拠なしで削除してよいか。
2. **Options considered**: (A) 一括 DROP / (B) external inventory 取得後に段階 DROP / (C) 恒久的に deprecation 維持。
3. **Verdict — `HOLD`**: response field と route は「呼べば今も 200 が返る」live surface である。slice2 の inventory は **in-repo のみ ZERO_IN_REPO** を証明したにすぎず、外部 client については報告書自身が `UNREPORTED (USER)` と明記し「CLEAN-GO の前提に使わない」と宣言している。ここを推測で埋めると院外予約導線を無言破壊し得る。安全側 = HOLD。
4. **Rejected alternatives**: (A) は不可逆かつ検知が遅れる（外部は次回利用時まで壊れたと気づかない）。(C) は削除目標に反するため、条件付き解除を前提とする HOLD を採る。
5. **Constraints / non-goals**: route DROP 時は capable-only 入口へ一本化する設計が先。migrate DROP は agent が実行しない。
6. **Evidence / refs**: slice2 §1.4「Disposition: UNREPORTED (USER)」· §5「意図的に触っていないもの」· §6-3（route DROP 時の capable-only 一本化要件）。
7. **STATUS.md update line**:
   `| TASK-021-B/C/D | response/route/migrate DROP | USER+PO | HOLD — external access log inventory 待ち |`
8. **Issue comment draft**: 「response `excluded_courses`・master exclusion route・DB DROP は、外部 access log / client registry による実利用ゼロの報告を受領するまで HOLD。」
9. **Unblock checklist**: USER が STG/production の access log で当該 path/field の呼び出しゼロを確認 → 報告 → D-021-B 承認 → C 設計 → D は backup + rollback owner 確定後。

**適用窓・rollback（D-021-D 用に事前明示）**: 適用は STG → 24h 観測 → production。rollback 条件は「exclusion 関連の 4xx/5xx 増加」または「予約可能スタッフ数の不整合」。DROP は不可逆のため backup 取得を前提とする。

### D-033 — 救急投薬 cutover と薬量 policy

1. **Question**: 臨床 bundle 未記入のまま TASK-033 のコード骨格に着手してよいか。
2. **Options considered**: (A) 骨格だけ先行実装 / (B) 全面 HOLD / (C) fail-closed 部分のみ先行。
3. **Verdict — `NEEDS_CLINICAL`、例外を認めない**: DEC-48 は「TASK-033 は構造化救急投薬記録と欠落時 fail-closed cutover を**一体で所有し、前者だけを先行させない**」と明示している。骨格先行は (C) を含め DEC-48 に正面から反する。さらに DEC-48 は「安全な代替記録経路が green になるまで current missing-data runtime を変更しない」としており、中途半端な着手は「通常保存が止まるが救急記録経路が未完成」という**最悪の臨床状態**を作る。これは臨床安全に直結するため、極小の例外も認めない。
4. **Rejected alternatives**: (A)/(C) は上記のとおり DEC-48 違反かつ臨床リスク。bundle を代理で埋める案は委任外であり禁止。
5. **Constraints / non-goals**: 上限 mg、warning %、動物種 default を本 pack では**一切決めない**。現行 20% threshold は DEC-65 のとおり変更しない（臨床正本への昇格でもない）。
6. **Evidence / refs**: DEC-48 · DEC-65 · DR-CLINICAL #201 bundle · `STATUS.md:129-131`（BLOCKED: 実装開始禁止）· Issue #201。
7. **STATUS.md update line**:
   `| TASK-033 | #201 救急投薬 cutover | 臨床+USER | HOLD 継続（DEC-48 一体所有）· 骨格先行も禁止 |`
8. **Issue comment draft**: 「TASK-033 は DEC-48 により構造化記録と fail-closed cutover を一体で所有するため、臨床 bundle 記入前の部分着手を行わない。TASK-377（逸脱理由契約）は既に landed 済みで、本 HOLD の対象外。」
9. **Unblock checklist**: 臨床責任者が #201 bundle 1 行へ次の**列**を記入 → HOLD 解除:
   - 対象（上限/warning の対象薬剤・動物種・master row ID、救急/既実施投薬の対象ケース・対象薬剤）
   - 値/範囲（上限 policy＝現行継続 or 修正値、warning policy＝同、救急記録 policy＝medicine identity / route vocabulary / dose・strength・concentration の unit と requiredness / weight・species snapshot / reason taxonomy or bounded free-text / 訂正対象・rationale / create grant 対象 role）
   - 単位 · 出典 · 承認者（role + opaque ref）· 発効日
   （**値は代理裁定しない。空欄は未承認のまま。**）

**cutover 発効環境**: local を前提とする。STG/production への展開は bundle 記入後の別判断（本 pack では承認しない）。

### D-LINE-R05 — production rollout + column DROP

1. **Question**: production で `line_channel_secret` 列を DROP してよいか。
2. **Options considered**: (A) 即 DROP / (B) 3 条件充足後に DROP / (C) 列を恒久保持。
3. **Verdict — `HOLD`**: DROP 対象は **`line_reservation_setting.line_channel_secret`** と特定済み（Phase B 報告の residual matrix）。ただし当該列は現在も `FindWebhookRouteByLineBotUserID` の presence SELECT `(line_channel_secret <> '')` に**使われており KEEP 判定**である。列を先に DROP すると webhook routing が壊れる。加えて composition test が旧契約のまま失敗している。よって未達 3 条件が揃うまで DROP 不可。
4. **Rejected alternatives**: (A) は認証・webhook 経路を破壊する実害が具体的に特定できているため却下。(C) は legacy credential 列の残置であり security 上望ましくないため、条件付き解除を前提とする HOLD。
5. **Constraints / non-goals**: 値・digest を inventory に保存しない。verifier へ reservation secret を再配線しない。agent は migrate/production 操作をしない。
6. **Evidence / refs**: `reports/2026-07-31-r05-single-sot-phase-b.md` residual matrix（model/DB column = HOLD、DROP migration = HOLD、presence SELECT = KEEP）· 同「次の packet（HOLD 解除条件）」1〜3 · `STATUS.md:138-140`。
7. **STATUS.md update line**:
   `| LINE-R05 | production rollout + column DROP | USER/PO | HOLD — 対象列 = line_reservation_setting.line_channel_secret · presence SELECT 依存の解消が先 |`
8. **Issue comment draft**: 「LINE-R05 の DROP 対象は `line_reservation_setting.line_channel_secret`。当該列は webhook routing の presence 判定に使用中のため、①clinic 別 legacy presence inventory ゼロ ②composition wiring test の Phase B 契約更新 ③presence SELECT 依存の解消、が揃うまで DROP しない。」
9. **Unblock checklist**: USER が clinic 別 presence inventory を取得（値は保存しない）→ ゼロ確認 → composition test 更新 unit を起票 → presence SELECT 依存を除去 → STG DROP → 観測 → production。

**リリース順序（承認する構造）**: app デプロイ（列参照の完全除去）→ 読み取り停止の観測期間 → backup → DROP。この順序自体は `APPROVE`、実行時期が HOLD。

### D-211 — 健診 package clinic import

1. **Question**: どの環境への import を承認するか。agent は何をしてよいか。
2. **Options considered**: (A) local へ実 row import / (B) synthetic のみで構造実装、実 row は USER / (C) 全面停止。
3. **Verdict — `NEEDS_CLINICAL` + `NEEDS_USER_OPS`**: DEC-59 は「agent は synthetic fixture だけで import/preflight を実装する」と確定済みで、TASK-374 として既に実装されている（local DDL 存在）。**環境承認は本 pack では出さない** — DEC-58 が要求する「target bundle/environment、seed-export 経路、checksum/reset 境界の migration-seed-safety review」が未完了であり、対象 clinic も未確定のため。実 manifest は repo 外。
4. **Rejected alternatives**: (A) は実 row を臨床承認なしに入れるため却下。CSV 手編集・003_demo の shared 環境 load は DEC-58/59 で明示禁止。
5. **Constraints / non-goals**: 実 manifest・apply receipt を repo に転記しない。opaque restricted reference のみ。
6. **Evidence / refs**: DEC-58 · DEC-59 · Issue #211 · `STATUS.md:166`（DR-OPS / DR-CLINICAL 分離）· `STATUS.md:61`（TASK-374-apply local DDL 済）。
7. **STATUS.md update line**:
   `| TASK-374-apply | checkup package import | USER | HOLD — DR-CLINICAL 実 row 承認 + DR-OPS 対象 clinic/環境確定の両方待ち |`
8. **Issue comment draft**: 「#211: import 構造（synthetic）は TASK-374 で実装済み。実 row の投入環境は、DR-CLINICAL の row 承認と DR-OPS の対象 clinic・環境確定が揃うまで承認しない。」
9. **Unblock checklist**: 臨床責任者が DR-CLINICAL #211 行（stable row key/field · type/choice/range · 出典 · 承認者）を記入 → USER が DR-OPS で対象 clinic・環境・operator・rollback を確定 → dry-run preview → apply。
   **agent がやってよい範囲**: 何も触らない（synthetic 実装は完了済み）。preview script の実行も USER。
   **rollback 条件**: clinic-scoped rollback（対象 clinic の import 分のみ）を事前に用意できない限り apply しない。

### D-249 — 検査機能（院内結果管理）

1. **Question**: 現行 demo seed 方針のまま追加 unit を起票してよいか。
2. **Options considered**: (A) 起票する / (B) 臨床 range 承認まで起票しない / (C) phase2 送り。
3. **Verdict — `HOLD`（追加 unit 起票せず）**: STATUS.md §2 が「承認前に新 unit を起票しない」と明示。`exam_reference_ranges` = 20 は **demo seed** であり臨床正本ではない。この状態で判定ロジックの unit を積むと、未承認値を前提にした実装が増える。DR-CLINICAL #249 行（検査項目・動物種・測定系・下限/上限）が空欄のため、値を要する unit は起票不可。
4. **Rejected alternatives**: (A) は未承認値の上に実装を積む（原則②③違反）。(C) は #249 全体の phase2 送りを意味し、BUG-003 系の臨床安全項目を含むため過剰。
5. **Constraints / non-goals**: reference range の具体値を代理決定しない。
6. **Evidence / refs**: `STATUS.md:167`（#249 次のアクション）· DR-CLINICAL #249 行 · `STATUS.md:38`（ranges=20 は 003_demo 由来）。
7. **STATUS.md update line**:
   `| #249 | 検査機能 | 判断待ち | HOLD — 臨床 range 承認まで新 unit 起票なし · 外部自動化は DEFER_PHASE2 |`
8. **Issue comment draft**: 「#249: 現行 `exam_reference_ranges`=20 は demo seed であり臨床正本ではない。臨床 range 承認まで追加 unit を起票しない。外部機器・自動化連携は phase2 へ移管。」
9. **Unblock checklist**: 臨床責任者が DR-CLINICAL #249 行を記入 → 承認後に **1 unit** を起票。
   **承認後の次 unit（タイトル案・値非依存）**: 「exam reference range の clinic-scoped 承認値 import 契約（値非依存・synthetic fixture のみ）」。**非目標**: 判定閾値の決定、外部機器連携、demo seed の書き換え。

### D-249-EXT — 外部機器 / 自動化連携

- **Verdict — `DEFER_PHASE2`**: 対象機器・接続仕様・責任者・価値実測（削除できる工程・所要時間）が一つも供給されていない。DEC-67 と同じ判断構造（未実測の追加手段は phase2 へ一意移管）。
- **Evidence**: DEC-67（未実測の入力手段追加は close 推奨 + phase2）· `phase2.html`。
- **STATUS.md update line**: `| #249-EXT | 検査 外部自動化連携 | — | DEFER_PHASE2（phase2.html へ移管） |`
- **Unblock**: 業務責任者 role + 対象機器 + 月次頻度 + 現行所要時間 + 削除される工程が揃った時に**新 Issue** で再起票。

### D-261 — 臨床安全・画面仕様ギャップ PO

1. **Question**: #261 を #201 bundle 参照のみで閉じる方針を再確認するか。
2. **Options considered**: (A) #201 参照のみ / (B) #261 独立の値表を新設 / (C) phase2 送り。
3. **Verdict — `APPROVE`（DEC-41/47/64 維持）**: (B) は同一上限/warning 表の二重正本を作り、承認値が drift する。DEC-47 が値の複製を明示禁止し、DEC-64 が「#261 は新しい臨床値や重複 evidence packet を作らない」と確定済み。独立値が真に必要な場合は、**同じ列を持つ別行を追加**するだけとし、値は本 pack では作らない。
4. **Rejected alternatives**: (B) 二重正本のため却下。(C) は runtime residual（権限監査・real LINE/LIFF）が live 必要のため却下。
5. **Constraints / non-goals**: #201 の値を #261 へ複製しない。実 identity を repo に書かない。
6. **Evidence / refs**: DEC-41 · DEC-47 · DEC-64 · DR-CLINICAL #261 行 · `STATUS.md:177`。
7. **STATUS.md update line**:
   `| #261 | 臨床安全・画面仕様ギャップ | USER 専権 | #201 bundle 参照のみで close（DEC-64 維持）· 値の複製禁止 |`
8. **Issue comment draft**: 「#261 は #201 bundle の opaque approval reference を参照して閉じる。独立値が必要な場合のみ同じ列を持つ別行を追加する（値の複製は行わない）。」
9. **Unblock checklist / close 条件の最小セット**: ①DB 方針 ②権限監査 ③real LINE/LIFF ④対象環境 runtime ⑤PO close — 各項目に**結果 enum + opaque restricted ref** の一行のみ。residual live に維持（phase2 へは移さない）。

### D-256 — 操作マニュアル / 研修

1. **Question**: DEC-61 default no-rewrite を維持するか。TASK-024 は必須残か DEFER か。
2. **Options considered**: (A) no-rewrite 維持 + TASK-024 必須 / (B) history rewrite 実施 / (C) TASK-024 を DEFER。
3. **Verdict — `APPROVE`（DEC-61 維持）+ TASK-024 は `APPROVE_WITH_CONSTRAINTS` で必須残**: history rewrite は不可逆で clone/fork/共同作業に影響するため既定は no-rewrite。TASK-024（screenshot / FAQ の visual sign-off）は **PII 再露出の最終防波堤**であり、DEFER すると privacy closure を偽ることになる（DEC-61 が明示的に却下した選択肢）。よって #256 close の必須 gate として維持。
4. **Rejected alternatives**: (B) は両 owner 承認・影響範囲・rollback plan が未供給。(C) は privacy 検証の省略に等しく却下。
5. **Constraints / non-goals**: 氏名・email・staff/clinic/group ID・roster・画像 path/hash・receipt 本文を台帳に書かない。agent は rewrite を実行しない。U13 の日程・形式・参加者は決めない。
6. **Evidence / refs**: DEC-61 · `STATUS.md:173` · `STATUS.md:58`（TASK-024 open）。
7. **STATUS.md update line**:
   `| TASK-024 | #256 screenshot / FAQ sign-off | USER | open · 必須残（DEC-61 維持 · #256 close gate） |`
8. **Issue comment draft**: 「#256 は DEC-61 の default no-rewrite を維持。TASK-024 の visual sign-off は PII 再露出防止の必須 gate として残す。台帳には結果 enum・role・opaque ref のみを記録する。」
9. **Unblock checklist**: Privacy owner と Repository owner が restricted evidence で disposition → clean-demo 再撮影 → visual/content sign-off → 非機密一行記録。
   **非機密一行記録の owner role**: Privacy owner（PII 判定）と Repository owner（history risk 判定）の連名。実名は台帳に書かない。

### D-257 — Go-live

1. **Question**: 旧 window の No-Go を維持するか。新 window を今決められるか。
2. **Options considered**: (A) 旧 window 続行 / (B) No-Go 維持 + gate-driven 再設定 / (C) 全面中止。
3. **Verdict — `HOLD`（DEC-60 維持、構造 D）**: 2026-08-03 window は既に失効しており、本日 2026-08-06 時点で物理的に実行不能。新 window は**今決められない** — #89/#97/#98/#99・#250・#253・#254・#255 の gate が未 green（本 pack で #98/#99 に close 経路を与えたが、実行は USER）。具体日付・担当者・契約条件は委任外。
4. **Rejected alternatives**: (A) 期限切れかつ open gate 多数。(C) delivery 目的の取消根拠がない。
5. **Constraints / non-goals**: 日付を発明しない。旧 AWS 系を rollback 先として復活させない。
6. **Evidence / refs**: DEC-60 · Issue #257 · `STATUS.md:174`。
7. **STATUS.md update line**:
   `| #257 | Go-live 手順・support | USER 専権 | HOLD — 旧 window No-Go 維持 · 新 window は全 gate green 後（DEC-60） |`
8. **Issue comment draft**: 「#257: 2026-08-03 window は失効済み No-Go を維持。新 window は前提 gate（credential residual・移行検証・production/CI/backup・UAT・staff provisioning）が非機密 evidence で green になった後にのみ設定する。」
9. **Unblock checklist / 必要な入力**: 上記 8 Issue の green evidence + **role ベース**の指名 — Go/No-Go authority role、support owner role、rollback owner role（実名は台帳外）。runbook は T-relative・gate-driven に維持（TASK-375）。

### D-098 — 旧 RDS credential 残余リスク

1. **Question**: provider 旧値無効化を必須とするか、残余リスク受容で close するか。
2. **Options considered**: (A) 無効化必須 / (B) ACCEPT_RESIDUAL_RISK で close / (C) 恒久 open。
3. **Verdict — `ACCEPT_RESIDUAL_RISK`**: 旧 RDS は既に到達経路が撤去されている前提の Issue であり、「履歴に残る旧値」は**技術的に無効化不能**（git 履歴の値そのものは rewrite しない = DEC-61）。したがって「無効化必須」を close 条件にすると恒久 open になる。provider 側で当該 credential が既に失効/ローテ済みであることの非機密確認をもって、残余リスクを受容し close する。**実行は USER 専権**、値は一切書かない。
4. **Rejected alternatives**: (A) は達成不能条件で Issue を恒久 open 化。(C) は判断の先送り。
5. **Constraints / non-goals**: 実 secret 値・接続文字列を出力しない。rotation 手順の代行はしない（`NEEDS_USER_OPS`、playbook 参照に留める）。
6. **Evidence / refs**: Issue #98 · `STATUS.md:163` · DEC-61（no-rewrite）。
7. **STATUS.md update line**:
   `| #98 | 旧 RDS credential 残余リスク | USER 専権 | ACCEPT_RESIDUAL_RISK で close 可 — provider 失効の非機密確認 + 受容理由一行 |`
8. **Issue comment draft**: 「#98: 履歴上の旧値は no-rewrite 方針下で技術的に消去しないため、provider 側の失効確認をもって残余リスクを受容し close する。#89 は credential class ごとの無効化・再発行の実行、#97 は公開面由来の露出と session 無効化を所有し、#98 はその後に残る履歴リスクの受容判断のみを所有する。」
9. **Unblock checklist**: USER が provider console で当該 credential の失効を確認 → 結果 enum + opaque ref を一行記録 → close。
   **#89/#97 との役割分離（一文）**: #89 = credential class ごとの無効化・再発行の**実行**、#97 = git/公開面由来の露出と session 無効化の**実行**、#98 = それらの後に残る**履歴残余リスクの受容判断**。

### D-099 — 旧 ECS deploy 経路

1. **Question**: 「実行可能経路なし」確認を誰の責任で完了とするか。rollback SoT を #253 一本化でよいか。
2. **Options considered**: (A) #99 に独自 rollback 手順を残す / (B) #253 一本化 / (C) 恒久 open。
3. **Verdict — `APPROVE`（#253 一本化）**: rollback 手順を #99 と #253 の二箇所で持つと二重管理となり必ず drift する（原則②）。#253 が production 整備・CI/CD・backup gate の正本であるため、rollback SoT はそこへ一本化する。#99 は「旧経路に実行可能な deploy 手段が存在しないこと」の確認のみを所有して close する。
4. **Rejected alternatives**: (A) 二重管理。(C) 確認は provider console で完結可能であり先送り不要。
5. **Constraints / non-goals**: provider 操作は agent が行わない。credential 値を書かない。
6. **Evidence / refs**: Issue #99 · Issue #253 · `STATUS.md:164`（rollback SoT は #253 と一本化）· `STATUS.md:170`。
7. **STATUS.md update line**:
   `| #99 | 旧 ECS deploy 経路の撤去確認 | USER 専権 | close 条件確定 — 実行可能経路ゼロの確認一行 · rollback SoT は #253 一本化 |`
8. **Issue comment draft**: 「#99: rollback 手順の正本は #253 に一本化する。#99 の close 条件は、provider 上に旧 ECS deploy の実行可能経路が存在しないことの非機密確認一行のみとする。」
9. **Unblock checklist**: **責任者 role = production/infra owner**。provider console で旧 deploy pipeline/task definition の実行不能を確認 → 結果 enum + opaque ref を一行記録 → close。

### D-250 / D-259 / D-284 — 依存待ち Issue

| Issue | Verdict | 再開トリガー |
|---|---|---|
| **#250** 旧 Access 移行 | `HOLD live` | producer bundle（completed_at · payment graph · crosswalk）受領 → 受領時点で READY 判定。**#257 の gate に含まれるため live 維持**（phase2 送り不可）。 |
| **#259** L ステップ Write API | `HOLD live` | 先方の Write API enable 通知。enable 後に USER が live send / cron 発火 / stop・rollback を実測。**先方作業のみが blocker で構造は完成済みのため live 維持**。 |
| **#284** line-reserve 実機フォント | `DEFER_PHASE2` | 試験環境 + 対象 3 実機の受領。**Go-live gate に含まれず、実機到着見込みが今フェーズ内に立たないため phase2 へ移管**。到着後に新 unit で再開。 |

- **Evidence**: `STATUS.md:168, 176, 178` · DEC-60（#250 は Go-live gate）。
- **STATUS.md update line**:
  `| #250 | 旧 Access 移行 | 依存待ち | HOLD live（#257 gate）· producer bundle 受領で再開 |`
  `| #259 | Lステップ Write API | 依存待ち | HOLD live · 先方 enable で再開 |`
  `| #284 | line-reserve 実機フォント | 依存待ち | DEFER_PHASE2 · 実機 3 台受領で再開 |`

### D-scope — residual に残す / 外す

1. **Question**: ブラウザ除外に加え、residual live から外してよい TASK/Issue はあるか。agent が次に触ってよい最大 1 unit は。
2. **Verdict — `APPROVE`**: live から外すのは **#284（D-284）と #249 の外部自動化部分（D-249-EXT）の 2 件のみ**。他は全て live 維持。特に #250/#259 は Go-live gate または完成済み構造の待機であり、phase2 送りは実態を歪めるため却下。
3. **agent が次に触ってよい最大 1 unit**: **`UNIT-021-A`**（D-021-A）。これ以外は NONE。
4. **Constraints**: 1 unit = 1 graph。UNIT-021-A 完了・land までは次の unit を起こさない。
5. **Evidence / refs**: `STATUS.md:53`（TASK-010 除外済）· `STATUS.md:149`（1 unit = 1 graph）· DEC-67（phase2 移管の判断構造）。
6. **STATUS.md update line**:
   `| residual scope | 2026-08-06 PO 代理 | PO | #284 と #249-EXT を phase2 へ移管 · 他 live 維持 · 次 agent unit = UNIT-021-A のみ |`

---

## 3. Agent backlog after decisions

**READY_AGENT = 1 unit のみ。**

### UNIT-021-A — request DTO / OpenAPI から `excluded_type_ids` を削除

- **依存**: なし（即着手可）
- **前提**: slice2 の Create/Update reject が landed 済み（実測確認済み）
- **scope**:
  1. `backend/internal/reservation/reservation_staff_request.go` から `excluded_type_ids` フィールドを削除
  2. `backend/docs/api.yaml` の `CreateReservationStaffRequest` / `UpdateReservationStaffRequest` から当該 property を削除
  3. reject 専用テスト（`Create_RejectsNonEmptyExcludedTypeIDs` / `Update_RejectsExcludedTypeIDsPresent`）を「未知フィールドは無視される」契約テストへ置換
  4. `openapi_reservation_staff_capability_contract_test.go` の deprecation assertion を更新
- **非目標（触らない）**: response `excluded_courses` / master exclusion route / handler / repository inverse facade / migrate / available-staffs
- **完了条件**: `./internal/reservation/` と `./internal/apicontract/` が green。挙動変化ゼロ（送信しても以前から 400 だった経路が、未知フィールド無視になるのみ）。

> **注**: これ以降の unit（response 削除 · route DROP · migrate DROP · TASK-033 · #249 追加 unit · LINE-R05 composition test 更新）は全て HOLD/NEEDS_* のため **列挙しない**。ブラウザ検証 unit も対象外。

---

## 4. Explicit non-decisions

本 pack が**裁定せず、触れないもの**:

- credential の実作業（#89/#97 の無効化・再発行、値・token・接続文字列）— `NEEDS_USER_OPS`、playbook 参照のみ
- ブラウザ IU 検証（TASK-010 / IU 32 / `BROWSER_VERIFICATION_*`）— residual 対象外
- `make migrate` / seed apply / DB_RESET / force-push / 本番 DROP の**実行** — 承認しても実行は USER/ops
- local reset の再実行（2026-08-06 完了済み）
- 臨床の具体数値（mg 上限・warning 帯・reference range・動物種 default）— `NEEDS_CLINICAL`
- 価格・契約・課金（#258 U1〜U12、Sentry paid plan、PO-008）— 契約責任者
- 実 identity（approver 実名、staff roster、参加者名簿）— repo 外・opaque ref のみ
- Go-live の具体日付・実行者・連絡先 — USER 専権
- GitHub Issue の close 操作そのもの — USER
- `VERIFIED_FIXED` の付与 — agent は行わない
- E2E_LOGIN_* の注入（TASK-020 / TASK-023）— USER 専権

---

## 5. Conflicts & assumptions

### Conflicts（ソース間矛盾）

1. **【重要】TASK-377 の状態矛盾** — `q&a.html` の `DISPOSITION-20260803` は #201 行で TASK-377 を `READY_AGENT` と記載。しかし commit `e6796be62`（2026-08-04）が「TASK-377 completion」を記録し、`treatment_dose_deviation_reason_test.go` 等 10 ファイルが実在（73 箇所）。
   → **より新しい実装事実を優先**し、TASK-377 = **landed** と判定。`q&a.html` の当該記述は stale。D-033 の順序論点は消滅。**要 SoT 修正。**
2. **TASK-032 の claim blocker 矛盾** — 同 commit は TASK-032 を「`claim/TASK-032` 解放待ち」としたが、実測で local/origin とも `claim/*` = 0（STATUS.md:41 とも一致）。
   → blocker は**解消済み**。ただし #249 の臨床 range gate（D-249）は別途生きているため、TASK-032 の再開可否は D-249 の解除に従属する。
3. **`q&a.html` の参照先ファイル不在** — `q&a.html` は `todo.md` と `3-session-agent.html` を正本として参照するが、両者は STATUS.md へ統合済み（STATUS.md:4）。
   → 現在の正本は `STATUS.md`。`q&a.html` 内リンクは historical。
4. **`STATUS.md:165` の #201 行が `todo.md` を参照** — 統合により無効リンク。→ 実害は小さいが要更新。

### Assumptions（本 pack が置いた仮定 — 最小・検証可能）

| # | 仮定 | 検証方法 | 外れた場合の影響 |
|---|---|---|---|
| A1 | slice2 の Create/Update reject が現在の main に landed している | `reservation_staff_service.go` の reject 分岐を確認（実測済み・4 箇所 hit） | D-021-A が APPROVE → HOLD へ後退 |
| A2 | `excluded_type_ids` を送る外部 consumer は、既に 400 で壊れているため property 削除で新規破壊は起きない | 同上（reject が Create 非空 / Update 存在で発火することの確認） | A1 と同じ。reject に抜け道があれば D-021-A は HOLD |
| A3 | `line_channel_secret` は presence SELECT で現在も参照されている | `FindWebhookRouteByLineBotUserID` を確認（Phase B 報告に KEEP 記載） | 参照が既に消えていれば D-LINE-R05 の解除条件が 1 つ減る（依然 HOLD） |
| A4 | open Issue 18 件が現況 | `gh issue list --state open`（実測済み = 18） | 件数差があれば §2 表の再照合が必要 |
| A5 | #250 は #257 の Go-live gate に含まれる（DEC-60 記載）ため live 維持が妥当 | DEC-60 本文（記載確認済み） | 含まれないなら #250 も DEFER_PHASE2 候補 |

### 新規 DEC として起票すべきもの（既存 DEC を覆さないが追加する）

- **DEC-68（新規）**: TASK-021 の破壊削除を **4 段階に分割**し、request DTO/OpenAPI（段階 A）のみを先行承認する。
  **Supersedes: なし**（slice2 報告の「CLEAN-GO 一括」前提を段階化するのみ。DEC-40〜67 のいずれも覆さない）。
  巻き戻し範囲: UNIT-021-A の 4 ファイル（request DTO · api.yaml · 2 テスト）。revert 1 commit で復旧可能。

---

## 6. USER への確認事項（覆せることの明示）

本 pack は**代理裁定**である。実 PO は全件を後から覆せる。覆す場合は:

1. 新 DEC カードとして再起票（`Supersedes: DEC-xx` を明記）
2. 該当 unit の実装を巻き戻す（UNIT-021-A は 1 commit revert で完了）
3. STATUS.md の該当行を新カードの verdict へ同期

特に **D-021-A**（唯一の APPROVE 実装 unit）と **D-098**（ACCEPT_RESIDUAL_RISK）は判断の余地があるため、異論があれば着手前に差し戻されたい。

---

PO_PROXY_DECISION_PACK complete
ready_agent_units: 1
clinical_blockers: 3
user_ops_blockers: 11
