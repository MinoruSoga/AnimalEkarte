# FABLE PO RECOMMENDATION PACK — residual ゲート最終推奨（2026-08-06）

> 起草: PO 推奨裁定官（Fable）。**USER（実 PO）が本 pack の推奨に従って最終決定する。Fable は実行しない。USER は全件を後から覆せる。**
> 入力: `STATUS.md` §1/§2 · `reports/2026-08-06-po-proxy-decision-pack.md`（Opus 代理）· `q&a.html` DEC-40〜67 / DR-CLINICAL · slice2 / Phase B report · 実測（下記）
> 実測（本 pack 起草時）: UNIT-021-A landed（`b917c2992` · request DTO/api.yaml に `excluded_type_ids` なし）· LINE-R05 presence SELECT 使用中（`line_reservation_setting_repository.go:56`）· **composition test は Phase B 契約へ更新済み・green（`fac8c86b2` 2026-07-31 · `go test ./cmd/api/ -run Composition` = ok）** · open Issue 18 · claim/* 0

---
> **採択**: 2026-08-06 USER — Fable 推奨を最終決定として採択（チェックリスト方針 Yes）。実装取り込み: DEC-68 · STATUS 更新（UNIT-DEC-68-DOC）。製品コード追加変更なし。


## 0. Executive recommendation

- **今日 Yes すべきトップ 3**: ① F-021-A の事後追認 + DEC-68 起票（記録を残す。黙認しない） ② F-098 / F-099 の close 経路確定（provider 確認一行 → close） ③ F-scope（#284 と #249-EXT の phase2 移管を確定し residual を締める）。
- **絶対に今 Yes すべきでないトップ 3**: ① TASK-033 の骨格先行の「極小例外」（DEC-48 一体所有に正面から反する） ② in-repo ZERO だけでの 021-B（response 削除）緩和（外部 reader 無言破壊リスク） ③ inventory ゼロ確認前の LINE-R05 presence 参照除去 unit（fail-closed guard の除去 = 移行未完 clinic の無警告受理）。
- Opus pack は **全 20 裁定を RATIFY**（OVERTURN 0）。ただし事実訂正 2 件: LINE-R05 の解除条件②（composition test 更新）は**既に充足済み**、021-A は「承認」ではなく「landed 事実の追認」に変化。

---

## 1. Decision table

| Decision ID | 対象 | Opus 前裁定 | **Fable 推奨 Verdict** | Opus との関係 | 発効条件 |
|---|---|---|---|---|---|
| F-021-A | UNIT-021-A 事後追認 | APPROVE_WITH_CONSTRAINTS | **RATIFY**（REVERT せず・DEC-68 起票で記録） | RATIFY | 即時 |
| F-021-B | response `excluded_courses` 削除 | HOLD | **HOLD** | RATIFY + TIGHTEN（証拠は access log では不足 → client registry 必須） | external consumer ゼロの一行報告 |
| F-021-C | master exclusion route DROP | HOLD | **HOLD** | RATIFY | B 充足 + capable-only 一本化設計 |
| F-021-D | DB migrate DROP | HOLD | **HOLD** | RATIFY | B/C 完了 + STG green + backup + rollback owner |
| F-021-X | inventory 90 日無応答時の扱い | （なし） | **ACCEPT_RESIDUAL_RISK を再裁定条件として予約** | N/A（新議題） | 依頼発行から 90 日 |
| F-033 | 救急投薬 cutover 着手 | NEEDS_CLINICAL（例外なし） | **NEEDS_CLINICAL**（骨格先行の例外を認めない） | RATIFY | #201 bundle 1 行の全列記入 |
| F-033-P | 上限 / warning / 救急記録 policy 値 | NEEDS_CLINICAL | **NEEDS_CLINICAL**（値は一切書かない） | RATIFY | 同上 |
| F-033-O | TASK-377 / 033 順序 | APPROVE（争点消滅） | **RATIFY**（377 landed 実測一致） | RATIFY | — |
| F-LINE-R05 | production rollout + column DROP | HOLD（3 条件） | **HOLD**（残条件は①inventory ゼロ ③presence 参照除去の 2 つ。**②は充足済み**） | RATIFY + TIGHTEN（事実訂正） | ①③充足 + STG green + backup |
| F-LINE-R05-b | presence 参照除去 code unit の先行承認 | 解除後に unit 化 | **HOLD**（inventory ゼロ**後**にのみ READY_AGENT） | RATIFY | clinic 別 presence inventory ゼロ報告 |
| F-LINE-R05-c | production DROP の恒久 REJECT 化 | — | **REJECT**（恒久化しない。条件付き HOLD 維持） | RATIFY | — |
| F-211 | 健診 package clinic import | NEEDS_CLINICAL + NEEDS_USER_OPS | **NEEDS_CLINICAL + NEEDS_USER_OPS**（agent は何もしない・実 row は local も禁止） | RATIFY | DR-CLINICAL 行 + DR-OPS 環境確定 |
| F-249 | 検査 追加 unit 起票 | HOLD | **HOLD**（値非依存 import 契約 unit の先行起票も**しない**） | RATIFY | DR-CLINICAL #249 行の記入 |
| F-249-EXT | 外部機器 / 自動化連携 | DEFER_PHASE2 | **DEFER_PHASE2** | RATIFY | phase2 再開条件成立時に新 Issue |
| F-261 | 臨床安全ギャップ close 方針 | APPROVE（#201 参照のみ） | **RATIFY**（二重正本の例外なし・別行追加方式のみ） | RATIFY | 5 項目の結果 enum + opaque ref |
| F-256 | マニュアル no-rewrite | APPROVE（DEC-61 維持） | **RATIFY** | RATIFY | — |
| F-024 | screenshot / FAQ sign-off | APPROVE_WITH_CONSTRAINTS（必須残） | **RATIFY**（DEFER しない） | RATIFY | 両 owner sign-off |
| F-257 | Go-live window | HOLD（DEC-60） | **HOLD**（新 window は今決めない） + gate 候補に **#252 追加を USER 確認** | RATIFY + TIGHTEN | 全 gate green |
| F-098 | 旧 RDS credential 残余 | ACCEPT_RESIDUAL_RISK | **ACCEPT_RESIDUAL_RISK**（provider 失効確認一行が前提） | RATIFY | 失効確認 enum + opaque ref |
| F-099 | 旧 ECS deploy 経路 | APPROVE（#253 一本化） | **APPROVE**（#253 一本化） | RATIFY | 実行可能経路ゼロ確認一行 |
| F-250 | 旧 Access 移行 | HOLD live | **HOLD live** | RATIFY | producer bundle 受領 |
| F-259 | L ステップ Write API | HOLD live | **HOLD live** | RATIFY | 先方 enable |
| F-284 | line-reserve 実機フォント | DEFER_PHASE2 | **DEFER_PHASE2** | RATIFY | 実機 3 台受領 |
| F-scope | residual 境界 | APPROVE | **RATIFY**（外すのは #284 / #249-EXT のみ） | RATIFY | STATUS.md 更新 |

---

## 2. Per-decision cards

### F-021-A — UNIT-021-A の事後追認

1. **Question**: landed 済みの request `excluded_type_ids` 削除（`b917c2992`）を追認するか、REVERT を推奨するか。記録は残すか。
2. **Options**: (A)★ RATIFY + DEC-68 を q&a.html に正式起票 / (B) RATIFY + 黙認（記録なし） / (C) REVERT_IMPLEMENTATION。
3. **Fable 推奨 Verdict — RATIFY**: slice2 で Create 非空 / Update 存在時とも既に 400 reject 済みであり、property 削除で新規に壊れる external consumer は存在しない（「送れば通る」経路が事前に消滅している）。実測でも request DTO / api.yaml から削除済み・scoped test green・response / route / migrate は未変更で HOLD 境界が守られている。挙動リスクは「未知フィールド無視化」のみで、revert 1 commit で完全復旧可能。REVERT する理由がない。
4. **Counter-argument**: OpenAPI から property が消えると、生成 client を再生成した外部 SDK は送信コード自体が型エラーになり、「400 で気づく」から「ビルドで壊れる」へ障害の顕在化点が変わる。
5. **Why still recommend**: その経路は既に機能していない（常時 400）。ビルド時に壊れる方がむしろ早期検知であり、fail-closed 原則に沿う。
6. **Constraints / non-goals**: response `excluded_courses`・master exclusion route・migrate・inverse facade `UpdateExcludedReservationTypes` は触らない。
7. **Evidence**: `b917c2992` · `reports/2026-08-06-unit-021-a-excluded-type-ids-request-removal.md` · slice2 §1.3/§2.1 · 実測 `backend/internal/reservation/reservation_staff_request.go:10,30`。
8. **USER Yes/No 一文**: 「Yes: UNIT-021-A を追認し、DEC-68（4 段階分割・段階 A のみ先行）を q&a.html に起票して記録を残す。」
9. **Unblock checklist**: USER が Yes → agent が UNIT-DEC-68-DOC（§5）で DEC-68 カード起票 + STATUS stale リンク修正 → USER が目視確認。
10. **STATUS.md 更新案**: `| TASK-021 | exclusion 破壊削除 | USER+PO | A 追認済み（DEC-68 起票）· B/C/D HOLD（external inventory 待ち） |`

### F-021-B/C/D — response / route / migrate DROP

1. **Question**: external 証拠なしで read 側 surface（response field / route / DB column）の削除へ進んでよいか。B→C→D の順序と inventory 手順・rollback 条件は。
2. **Options**: (A)★ HOLD 維持（B→C→D 順序固定・90 日再裁定条項付き） / (B) in-repo ZERO を根拠に B のみ緩和 / (C) 恒久 deprecation。
3. **Fable 推奨 Verdict — HOLD（3 件とも）**: slice2 自身が external 利用を `UNREPORTED (USER)` と明記し「CLEAN-GO の前提に使わない」と宣言している（実測 §1.4/§6 で確認）。response / route は「呼べば今も 200」の live surface であり、推測での削除は院外導線の無言破壊になり得る。順序 B→C→D は維持（read 契約 → 入口 → 不可逆 DROP の順にリスクが増える正しい階段）。**in-repo ZERO だけで B を APPROVE するのは不可** — さらに B は access log でも証明できない（field は既存 GET 応答に同乗しており、log は route 呼び出ししか示さない）。B の証拠は client registry（既知 consumer の列挙宣言）が必須。
4. **Counter-argument**: external inventory が永遠に届かなければ dead surface が恒久残置され、保守コスト・攻撃表面が残る（削除原則②違反）。
5. **Why still recommend**: 不可逆破壊 > 保守コスト。ただし恒久 open 化を防ぐため F-021-X（依頼発行から 90 日無応答なら ACCEPT_RESIDUAL_RISK として**再裁定に上げる**）を新設し、silent HOLD にしない。
6. **Constraints / non-goals**: C は capable-only 入口一本化設計が先。D は agent 実行禁止（USER 専権 migrate）。
7. **Evidence**: slice2 §1.4「Disposition: UNREPORTED (USER)」·§6-3 · Opus D-021-B/C/D。
8. **USER Yes/No 一文**: 「Yes: B/C/D の HOLD と B→C→D 順序を維持し、90 日再裁定条項を追加する。」
9. **Unblock checklist（USER 最小 inventory 手順・秘密を残さない）**: ① STG/production の access log（ALB/nginx 等）で `excluded-reservation-types` route の直近 90 日呼び出し件数を集計（path と件数のみ。token・IP・UA は台帳に書かない）→ C 用証拠 ② API client registry で「in-repo FE / LIFF 以外の API consumer 不在」を一行宣言 → B 用証拠 ③ 結果 enum + opaque ref を STATUS に一行 → B 承認 → C 設計 → D は backup + rollback owner 確定後。**rollback 条件**: B/C 後に exclusion 関連 4xx/5xx 増加 or 予約可能スタッフ数の不整合 → revert commit。D は backup restore のみ（不可逆）。
10. **STATUS.md 更新案**: `| TASK-021-B/C/D | response/route/migrate DROP | USER+PO | HOLD — client registry + access log inventory 待ち · 依頼発行から 90 日無応答で ACCEPT_RESIDUAL_RISK 再裁定 |`

### F-033 / #201 — 救急投薬 cutover

1. **Question**: 骨格先行禁止を維持するか。臨床 bundle の列は何か。cutover 環境の段階は。解除後の最初の unit は。
2. **Options**: (A)★ 骨格先行禁止を維持（例外なし） / (B) 値非依存の fail-closed 部分のみ極小先行 / (C) スキーマだけ先行。
3. **Fable 推奨 Verdict — NEEDS_CLINICAL・例外を認めない**: DEC-48 は「構造化救急投薬記録と欠落時 fail-closed cutover を一体所有し、前者だけを先行させない」「安全な代替記録経路が green になるまで current missing-data runtime を変更しない」と明示する。部分着手は「通常保存が止まるのに救急記録経路が未完成」という最悪の臨床状態を作り得る。(B)(C) を含むあらゆる例外は DEC-48 違反。**cutover 環境**は Opus どおり **local 先行**を追認し、STG/production 展開は bundle 記入後の別裁定とする（本 pack では承認しない）。
4. **Counter-argument**: 一体所有だと bundle 記入後の実装が大きくなり、薄い縦切り（product philosophy ④）に反してリードタイムが延びる。
5. **Why still recommend**: 臨床安全 > cycle time（製品哲学自身が「安全原則は効率原則に優先」と明記）。縦切りの薄さは bundle 記入後の unit 設計で確保すればよい。
6. **Constraints / non-goals**: mg 上限・warning %・動物種 default を一切書かない。現行 20% threshold は DEC-65 のとおり変更しない。
7. **Evidence**: DEC-48 · DEC-65 · DR-CLINICAL #201 bundle（q&a.html:492-501）· TASK-377 landed（`e6796be62` 実測）。
8. **USER Yes/No 一文**: 「Yes: 骨格先行禁止を維持し、臨床責任者へ #201 bundle 記入依頼（下記列リスト）を送付する。」
9. **Unblock checklist — bundle 列リスト（**列名のみ・値は未供給 = NEEDS_CLINICAL**）**: 対象（薬剤・動物種・master row ID・救急/既実施投薬の対象ケース）／値・範囲（上限 policy 継続可否、warning policy、救急記録 policy: medicine identity / route vocabulary / dose・strength・concentration の unit と requiredness / weight・species snapshot / reason taxonomy or bounded free-text / 訂正対象・rationale / create grant 対象 role）／単位／出典／承認者（role + opaque ref）／発効日。
10. **STATUS.md 更新案**: `| TASK-033 | #201 救急投薬 cutover | 臨床+USER | HOLD 継続（DEC-48 一体所有・骨格先行禁止）· 解除後の最初 unit = UNIT-033（下記） |`
    **解除後の最初の 1 unit タイトル案（値非依存）**: 「UNIT-033 — 構造化救急投薬記録 + 欠落時 fail-closed cutover（一体・#201 bundle 準拠・local 限定）」。

### F-LINE-R05 — `line_channel_secret` DROP

1. **Question**: HOLD + 3 条件を追認するか。presence 参照除去 code unit を DROP と分離して先行承認できるか。production DROP を恒久 REJECT にすべきか。composition test 更新 unit を READY_AGENT にするか。
2. **Options**: (A)★ HOLD 維持・残条件を①③の 2 つに訂正・presence 参照除去は inventory ゼロ後に READY_AGENT / (B) presence 参照除去 unit を今すぐ承認 / (C) production DROP 恒久 REJECT（列恒久保持）。
3. **Fable 推奨 Verdict — HOLD（訂正付き RATIFY）**: **事実訂正 — Opus の条件②「composition test を Phase B 契約へ更新」は既に充足済み**（`fac8c86b2` 2026-07-31 で更新、実測 `go test ./cmd/api/ -run Composition` = green。Opus pack の「旧契約のまま失敗している」は起草時点で stale）。よって composition test 更新 unit は**不要（no-op）— READY_AGENT にしない**。残条件は ①clinic 別 legacy presence inventory ゼロ ③presence SELECT 依存の解消、の 2 つ。**presence 参照除去の先行承認は不可**: 実測で `legacyCredentialPresent` は webhook 検証の **fail-closed reject 条件**（`line_link_service.go:416-423` — legacy 列に値が残る clinic は処理を拒否）。inventory ゼロ確認前に除去すると、移行未完（dual-SoT）状態の clinic を無警告で受理する。inventory ゼロ後は真に冗長化するため、そこで初めて code unit 化する。**production DROP の恒久 REJECT はしない**: legacy credential 列の恒久残置は security 上望ましくなく、条件付き HOLD が正しい。
4. **Counter-argument**: 署名検証は既に clinic_integrations 側 credential で行われており、presence guard は冗長 — 今除去してもセキュリティは劣化しないという読み。
5. **Why still recommend**: guard は「検証の冗長性」ではなく「移行完了の検出装置」。除去は fail-closed → fail-open の変更であり、認証系では inventory 証拠なしに行わない（Safety first）。
6. **Constraints / non-goals**: inventory に値・digest を保存しない。verifier へ reservation secret を再配線しない。migrate/production 操作は agent 禁止。
7. **Evidence**: Phase B report residual matrix（presence SELECT = KEEP · column/DROP = HOLD）· 実測 `line_reservation_setting_repository.go:56` · `line_link_service.go:416-423` · `fac8c86b2` · composition test green（本日実測）。
8. **USER Yes/No 一文**: 「Yes: LINE-R05 は残条件①③の 2 条件 HOLD に訂正して維持し、presence 参照除去 unit は inventory ゼロ報告後にのみ READY_AGENT とする。」
9. **Unblock checklist**: USER が clinic 別 presence inventory（`(line_channel_secret <> '')` の clinic 件数のみ・値非保存）を取得 → ゼロ確認一行 → UNIT-LINE-R05-PRESENCE（参照除去 + presence guard 削除 + テスト更新）を READY_AGENT 化 → land → STG DROP → 観測 → backup → production DROP。リリース順序（app デプロイ → 読み取り停止観測 → backup → DROP）は Opus 追認。
10. **STATUS.md 更新案**: `| LINE-R05 | production + column DROP | USER/PO | HOLD — 残条件 = ①presence inventory ゼロ ③presence SELECT 依存除去（②composition test は fac8c86b2 で充足済み） |`

### F-211 — 健診 package import

1. **Question**: NEEDS_CLINICAL + NEEDS_USER_OPS を追認するか。agent の可動範囲は。local への実 row 投入は。close 条件の最小セットは。
2. **Options**: (A)★ 追認・agent は何もしない・実 row は local も禁止 / (B) local に限り実 row 先行投入 / (C) 全面停止。
3. **Fable 推奨 Verdict — NEEDS_CLINICAL + NEEDS_USER_OPS（RATIFY）**: synthetic 実装は TASK-374 で完了済みであり、agent に残る作業はない。**local への実 row 投入も絶対禁止を推奨** — DEC-59 が 003_demo を昇格元にしないと確定しており、臨床未承認 row は local でも「正しく見える誤値」となって以降の検証全体を汚染する。DEC-58 の migration-seed-safety review（target bundle/environment・seed-export 経路・checksum/reset 境界）未完了のため環境承認も出さない。
4. **Counter-argument**: local は隔離環境であり、実 row を早く入れた方が画面検証・operator 訓練が進む。
5. **Why still recommend**: 「早く入れた誤値」は後で必ず二重正本 drift を生む（DEC-47 の禁止構造）。synthetic fixture で検証は既に可能。
6. **Constraints / non-goals**: 実 manifest・apply receipt を repo に転記しない（opaque restricted ref のみ）。
7. **Evidence**: DEC-58 · DEC-59 · Issue #211 · STATUS.md:61（TASK-374-apply local DDL 済）。
8. **USER Yes/No 一文**: 「Yes: #211 は DR-CLINICAL 行 + DR-OPS 環境確定が揃うまで実 row を（local 含め）投入しない。」
9. **Unblock checklist / close 最小セット**: ① DR-CLINICAL #211 行（stable row key/field・type/choice/range・出典・承認者） ② DR-OPS（対象 clinic・環境・operator・clinic-scoped rollback） ③ dry-run preview 結果 enum ④ apply receipt の opaque ref — 各一行。clinic-scoped rollback を用意できない限り apply しない。
10. **STATUS.md 更新案**: `| TASK-374-apply | checkup package import | USER | HOLD — DR-CLINICAL + DR-OPS 両gate 待ち · 実 row は local 含め投入禁止 |`

### F-249 — 検査機能 unit 起票

1. **Question**: 起票禁止を追認するか。値非依存の import 契約 unit だけ先行起票してよいか。外部自動化 DEFER を追認するか。
2. **Options**: (A)★ 起票禁止維持（値非依存 unit も先行しない） / (B) 値非依存 import 契約 unit のみ先行起票 / (C) phase2 送り。
3. **Fable 推奨 Verdict — HOLD（RATIFY・(B) も採らない）**: STATUS §2 の明文「承認前に新 unit を起票しない」が現行 SoT。(B) は TASK-374 と同型で技術的には安全だが、①DR-CLINICAL #249 行（検査項目・動物種・測定系・下限/上限）の確定前に契約形状を固めると手戻りリスクがある ②現在 agent 並行余地を作る必要がない（他 unit も全て人間 gate 待ち）③「追加だけで削除ゼロ」の実装は原則②に反する。先行のリターンが薄くリスクだけ足す。
4. **Counter-argument**: import 契約は値非依存で列構造も既知であり、承認後の立ち上がりを 1 unit 分速められる。
5. **Why still recommend**: 短縮幅は小さく（契約実装は承認後すぐ可能）、未承認値前提の実装面が増えるリスクが上回る。SoT 明文ルールを黙って緩めない。
6. **Constraints / non-goals**: reference range の具体値を決めない。demo seed（ranges=20）を臨床正本と読み替えない。
7. **Evidence**: STATUS.md:174 · DR-CLINICAL #249 行 · DEC-47 Q4 · Opus D-249。
8. **USER Yes/No 一文**: 「Yes: #249 は臨床 range 承認まで新 unit を一切起票せず、外部自動化は phase2 へ移管する。」
9. **Unblock checklist**: 臨床責任者が DR-CLINICAL #249 行を記入 → 承認後に 1 unit（「exam reference range の clinic-scoped 承認値 import 契約（値非依存・synthetic fixture のみ）」= Opus 案を追認）。
10. **STATUS.md 更新案**: `| #249 | 検査機能 | 判断待ち | HOLD — range 承認まで起票なし（値非依存 unit も先行せず）· 外部自動化 DEFER_PHASE2 |`

### F-261 — 臨床安全ギャップ

1. **Question**: #201 bundle 参照のみで close する方針を追認するか。二重正本の例外は。close 一行の最小セットは。residual live 維持か。
2. **Options**: (A)★ #201 参照のみ + 必要時は同列別行追加 / (B) #261 独立値表新設 / (C) phase2 送り。
3. **Fable 推奨 Verdict — RATIFY（APPROVE）**: DEC-47 Q2 が値複製を明示禁止、DEC-64 が「#261 は新しい臨床値・重複 evidence packet を作らない」と確定済み。**二重正本を許す例外はない** — 独立値が真に必要な場合も「同じ列を持つ別行の追加」であり複製ではない。runtime residual（権限監査・real LINE/LIFF）が live 必要のため phase2 送りは不可。residual live 維持。
4. **Counter-argument**: #201 参照のみだと画面仕様側の判断根拠が bundle に埋もれ、後から #261 単体で監査しにくい。
5. **Why still recommend**: 探索性の低下は opaque ref の一行記録で補える。値 drift は臨床事故に直結し、比較にならない。
6. **Constraints / non-goals**: #201 の値を #261 へ複製しない。実 identity を repo に書かない。
7. **Evidence**: DEC-41 · DEC-47 Q2 · DEC-64 · STATUS.md:184。
8. **USER Yes/No 一文**: 「Yes: #261 は #201 bundle の opaque ref 参照のみで close し、二重正本例外を認めない。」
9. **Unblock checklist / close 最小セット**: ①DB 方針 ②権限監査 ③real LINE/LIFF ④対象環境 runtime ⑤PO close — 各「結果 enum + opaque restricted ref」一行、計 5 行。
10. **STATUS.md 更新案**: `| #261 | 臨床安全・画面仕様ギャップ | USER 専権 | #201 参照のみで close（DEC-64）· 5 項目 enum+ref · live 維持 |`

### F-256 / F-024 — マニュアル · screenshot sign-off

1. **Question**: TASK-024 必須残を追認するか DEFER するか。DEC-61 no-rewrite を維持するか。
2. **Options**: (A)★ no-rewrite 維持 + TASK-024 必須残 / (B) TASK-024 DEFER / (C) history rewrite 実施。
3. **Fable 推奨 Verdict — RATIFY（両方）**: TASK-024 の visual sign-off は PII 再露出の最終防波堤であり、DEFER は「privacy closure を偽る」として DEC-61 が明示的に却下した選択肢そのもの。**privacy リスクの明示**: DEFER すると、reachable history に残る差し戻し前画像 / full-seed 画像の PII が納品物・研修資料経由で再露出しても検出工程が存在しない状態になる。no-rewrite は不可逆性（clone/fork/共同作業破壊）から既定維持。
4. **Counter-argument**: TASK-024 は Go-live gate（DEC-60 の 8 Issue）に含まれておらず、DEFER しても納品自体は可能。residual を 1 件減らせる。
5. **Why still recommend**: gate 外であることは「省略可能」を意味しない。#256 close の必須 gate として維持するのが DEC-61 の構造。
6. **Constraints / non-goals**: 氏名・email・ID・roster・画像 path/hash・receipt 本文を台帳に書かない。agent は rewrite を実行しない。U13 日程・形式・参加者は決めない。
7. **Evidence**: DEC-61 · `reports/2026-07-31-task-024-manual-audit.md` · STATUS.md:58。
8. **USER Yes/No 一文**: 「Yes: DEC-61 no-rewrite を維持し、TASK-024 を #256 close の必須 gate として残す。」
9. **Unblock checklist**: Privacy owner + Repository owner が restricted evidence で disposition → clean-demo 再撮影 → visual/content sign-off → 連名の非機密一行記録。
10. **STATUS.md 更新案**: `| TASK-024 | #256 sign-off | USER | open · 必須残（DEC-61 維持・Fable 追認） |`

### F-257 — Go-live

1. **Question**: 旧 window No-Go・新 window HOLD を追認するか。新 window を今決めないことを推奨するか。gate リストの過不足は。
2. **Options**: (A)★ RATIFY + gate 候補 #252 の追加を USER 確認 / (B) 仮 window を先置き / (C) gate リスト現状のまま。
3. **Fable 推奨 Verdict — HOLD（RATIFY + TIGHTEN）**: 2026-08-03 window は失効済みで物理的に実行不能、No-Go 維持は自明。**新 window は今決めない** — gate green 前の日付は再失効し、DEC-60 の失敗構造を反復するだけ。**gate 過不足**: DEC-60 の 8 Issue（#89/#97/#98/#99/#250/#253/#254/#255）に加え、**#252（各院の締め時間設定値投入）を gate 追加候補として USER 確認を推奨** — 締め処理は会計に直結し、未投入のまま go-live すると初回締めで事故る。外してよい Issue はなし。#258（納品ドキュメント）は契約系であり技術 gate に含めない現状を維持。
4. **Counter-argument**: window を仮置きしないと院側・先方の調整が始まらず、納品が漂流する。
5. **Why still recommend**: 調整は T-relative runbook（TASK-375 構造）で日付なしに進められる。失効 window の再生産の方が信頼コストが大きい。
6. **Constraints / non-goals**: 日付を発明しない。旧 AWS 系を rollback 先に復活させない。
7. **Evidence**: DEC-60 · Issue #257 · STATUS.md:181 · #252（STATUS.md:176）。
8. **USER Yes/No 一文**: 「Yes: 新 window は全 gate green 後にのみ設定し、#252 を gate に加えるか否かを USER が一行で確定する。」
9. **Unblock checklist**: 8（+1 候補）Issue の green evidence + role ベース指名（Go/No-Go authority / support owner / rollback owner）→ 新 window 設定。
10. **STATUS.md 更新案**: `| #257 | Go-live | USER 専権 | HOLD — 旧 window No-Go · 新 window は全 gate green 後 · #252 の gate 追加可否は USER 確定待ち |`

### F-098 / F-099 — credential residual · 旧 ECS

1. **Question**: #98 の ACCEPT_RESIDUAL_RISK と #99 の #253 一本化 close を追認するか。close 前の最小 USER 証拠は。#89/#97 との役割分離は。
2. **Options**: #98: (A)★ 受容 close / (B) 無効化必須で恒久 open — #99: (A)★ #253 一本化 / (B) 独自 rollback 手順維持。
3. **Fable 推奨 Verdict — 両方 RATIFY**: #98 — git 履歴上の旧値は no-rewrite（DEC-61）下で技術的に消去不能であり、「無効化必須」条件は達成不能 = 恒久 open 化。provider 側失効の非機密確認を**前提条件**とした受容 close が唯一の合理的経路。#99 — rollback 手順の二箇所持ちは必ず drift する（原則②）。#253 一本化 + 「実行可能経路ゼロ」確認一行で close。
4. **Counter-argument**: provider 失効確認だけでは「第三者が既に旧値を取得済み」だった期間のリスク（横展開・再利用）は消えない。
5. **Why still recommend**: 失効済み credential は取得されていても使用不能。到達経路も撤去済みが前提の Issue であり、残るのは受容判断のみ。受容理由を一行明記するため silent ではない。
6. **Constraints / non-goals**: 実 secret 値・接続文字列を出力しない。rotation 実行は #89/#97 側（NEEDS_USER_OPS）。
7. **Evidence**: Issue #98/#99 · DEC-61 · STATUS.md:170-171 · Issue #253。
8. **USER Yes/No 一文**: 「Yes: #98 は provider 失効確認一行を得て残余リスク受容で close、#99 は経路ゼロ確認一行 + rollback SoT #253 一本化で close する。」
9. **Unblock checklist**: production/infra owner role が provider console で ①当該 credential 失効 ②旧 ECS deploy pipeline/task definition の実行不能 を確認 → 各「結果 enum + opaque ref + 受容/確認理由」一行 → close。
10. **STATUS.md 更新案**: `| #98/#99 | credential 残余 / 旧 ECS | USER 専権 | close 経路確定（Fable 追認）— provider 確認一行のみで close 可 |`
    **#89/#97 との役割分離（一文）**: #89 = credential class ごとの無効化・再発行の実行、#97 = 公開面由来露出と session 無効化の実行、#98 = 実行後に残る履歴残余リスクの受容判断、#99 = 旧 deploy 経路の不存在確認。

### F-250 / F-259 / F-284 — 依存待ち

1. **Question**: 各 Issue の HOLD live vs DEFER_PHASE2 と再開トリガー。
2. **Options**: 各 Issue につき live 維持 / phase2 送り。
3. **Fable 推奨 Verdict**: **#250 HOLD live（RATIFY）** — DEC-60 の Go-live gate に含まれるため phase2 送りは Go-live 前提の歪曲。トリガー = producer bundle（completed_at・payment graph・crosswalk）受領。**#259 HOLD live（RATIFY）** — 構造完成済み・blocker は先方 enable のみ。トリガー = enable 通知 → USER が live send / cron / stop・rollback 実測。**#284 DEFER_PHASE2（RATIFY）** — Go-live gate 外・実機到着見込みなし。トリガー = 試験環境 + 3 実機受領 → 新 unit。
4. **Counter-argument**: #250/#259 も phase2 へ送れば residual 台帳が更に締まる。
5. **Why still recommend**: #250 は gate、#259 は「待つだけで READY になる」完成品。phase2 送りは実態を歪め、再開コストを増やすだけ。
6. **Constraints / non-goals**: 先方への催促方法・契約条件は決めない。
7. **Evidence**: DEC-60（#250 gate 記載）· STATUS.md:175,183,185。
8. **USER Yes/No 一文**: 「Yes: #250/#259 は HOLD live、#284 は phase2 移管とする。」
9. **Unblock checklist**: 各トリガー成立時に USER が STATUS 行を READY へ更新。
10. **STATUS.md 更新案**: Opus 提示の 3 行をそのまま採用（`#250 HOLD live` / `#259 HOLD live` / `#284 DEFER_PHASE2`）。

### F-scope — residual 境界

1. **Question**: live residual から追加で外せるものはあるか。全 Yes 後の次の 1 agent unit は。「今は何もしない」が最適か。
2. **Options**: (A)★ 外すのは #284 / #249-EXT のみ・次 unit は docs 1 件のみ / (B) 追加で外す / (C) 完全 NONE（docs もしない）。
3. **Fable 推奨 Verdict — RATIFY**: 追加で外せる TASK/Issue は**ない**（#250/#259 は上記のとおり live 正当、TASK-020/023/022/024 は納品・privacy の実 gate）。**製品コードについては「今は何もしない」が最適** — 残 blocker は全て人間 gate であり、agent 作業を発明するのはスコープ肥大（原則①②）。唯一の例外は記録系: USER が F-021-A に Yes した場合のみ **UNIT-DEC-68-DOC**（§5）が READY_AGENT。
4. **Counter-argument**: agent を遊ばせると momentum を失い、gate 解除後の立ち上がりが遅れる。
5. **Why still recommend**: 各 gate の解除後 unit は本 pack で既にタイトルまで定義済み（UNIT-033 / UNIT-LINE-R05-PRESENCE / #249 import 契約）。立ち上がりは既に最短化されている。
6. **Constraints / non-goals**: ブラウザ unit（TASK-010）は agenda 外。VERIFIED_FIXED は付けない。
7. **Evidence**: STATUS.md「Agent 規律」（residual agent 作業は尽きた）· Opus D-scope。
8. **USER Yes/No 一文**: 「Yes: residual 境界は Opus 案どおり確定し、次の agent unit は UNIT-DEC-68-DOC のみ（F-021-A Yes 時）とする。」
9. **Unblock checklist**: USER が本チェックリストを返す → 採用行を STATUS へ反映（agent 実行可・docs のみ）。
10. **STATUS.md 更新案**: `| residual scope | 2026-08-06 Fable | PO | Opus 20 裁定を全件追認（OVERTURN 0）· 次 agent unit = UNIT-DEC-68-DOC のみ |`

---

## 3. USER 採否チェックリスト

| # | 推奨アクション | Fable 推奨 | USER 採否 | 採否後の次手 |
|---|----------------|------------|------------------|--------------|
| 1 | UNIT-021-A を追認し DEC-68 を起票（黙認しない） | RATIFY | ☐ Yes / ☐ No | Yes → UNIT-DEC-68-DOC を agent 起票 |
| 2 | 021-B/C/D の HOLD と B→C→D 順序を維持（in-repo ZERO での緩和拒否） | HOLD | ☐ Yes / ☐ No | Yes → inventory 依頼を発行 |
| 3 | 021 inventory 90 日無応答で ACCEPT_RESIDUAL_RISK 再裁定条項を追加 | APPROVE_WITH_CONSTRAINTS | ☐ Yes / ☐ No | Yes → STATUS に期限行を追記 |
| 4 | TASK-033 骨格先行禁止を維持（例外なし）・local 先行 cutover | NEEDS_CLINICAL | ☐ Yes / ☐ No | Yes → 臨床責任者へ bundle 列リスト送付 |
| 5 | LINE-R05 HOLD を残条件①③の 2 条件へ訂正して維持 | HOLD | ☐ Yes / ☐ No | Yes → STATUS の条件行を更新 |
| 6 | presence 参照除去 unit は inventory ゼロ後にのみ READY_AGENT | HOLD | ☐ Yes / ☐ No | Yes → USER が presence inventory 取得 |
| 7 | production DROP を恒久 REJECT にしない（条件付き HOLD 維持） | REJECT（恒久化を） | ☐ Yes / ☐ No | — |
| 8 | #211: 実 row は local 含め投入禁止・agent は何もしない | NEEDS_CLINICAL+NEEDS_USER_OPS | ☐ Yes / ☐ No | Yes → DR-CLINICAL/DR-OPS 記入依頼 |
| 9 | #249: 起票禁止維持（値非依存 unit も先行しない） | HOLD | ☐ Yes / ☐ No | Yes → 臨床 range 記入依頼 |
| 10 | #249-EXT: 外部自動化を phase2 移管 | DEFER_PHASE2 | ☐ Yes / ☐ No | Yes → phase2.html 記載 |
| 11 | #261: #201 参照のみで close・二重正本例外なし | APPROVE | ☐ Yes / ☐ No | Yes → 5 項目 enum+ref 記入 |
| 12 | #256/#024: no-rewrite 維持 + TASK-024 必須残 | APPROVE | ☐ Yes / ☐ No | Yes → 両 owner disposition 開始 |
| 13 | #257: 新 window は今決めない + #252 の gate 追加可否を確定 | HOLD | ☐ Yes / ☐ No | Yes → #252 gate 可否を一行記入 |
| 14 | #98: provider 失効確認一行 → 残余リスク受容 close | ACCEPT_RESIDUAL_RISK | ☐ Yes / ☐ No | Yes → provider console 確認 |
| 15 | #99: 経路ゼロ確認一行 + rollback SoT #253 一本化 → close | APPROVE | ☐ Yes / ☐ No | Yes → provider console 確認 |
| 16 | #250/#259 HOLD live · #284 DEFER_PHASE2 · residual 境界確定 | RATIFY | ☐ Yes / ☐ No | Yes → STATUS 3 行更新 |

---

## 4. Priority order for USER

1. **チェックリスト #1〜#3, #5〜#7, #16 を採否**（判断のみ・作業ゼロ・即日可）
2. **#98/#99 の provider console 確認 → close**（2 件 close で residual が実際に減る）
3. **E2E_LOGIN_* 注入 → TASK-020/023**（`NEEDS_USER_OPS` — 手順は既存 runbook 参照。一行で足りる）
4. **021-B/C/D の inventory**（access log 集計 + client registry 一行宣言・秘密非記載）
5. **LINE-R05 presence inventory**（clinic 件数のみ・値非保存）
6. **臨床記入依頼の送付（値は書かない）**: ① #201 bundle（F-033 カードの列リスト） ② DR-CLINICAL #249 行 ③ DR-CLINICAL #211 行 + DR-OPS
7. **TASK-022 / TASK-024 の human 証跡**
8. **#257 新 window 設定は全 gate green 後**（最後）

---

## 5. Agent backlog after USER adopts Fable recommendations

- **UNIT-DEC-68-DOC**（チェックリスト #1 Yes 時のみ READY_AGENT・docs のみ・1 unit = 1 graph）: q&a.html に DEC-68 カード（TASK-021 4 段階分割・段階 A 先行承認・Supersedes なし）を起票し、STATUS.md:172 の stale な `todo.md` リンクを是正。製品コード変更なし。
- 上記以外は **NONE**（UNIT-033 / UNIT-LINE-R05-PRESENCE / #249 import 契約は各 gate 解除後に READY 化。ブラウザ unit は対象外）。

---

## 6. Explicit non-decisions

- 臨床の具体数値（mg 上限・warning %・reference range・動物種 default・#211 実 row 値）— `NEEDS_CLINICAL`
- credential 実作業（#89/#97）・`make migrate`・seed apply・DB_RESET・force-push・本番 DROP の実行 — `NEEDS_USER_OPS` / USER 専権
- E2E_LOGIN_* の注入（TASK-020/023）— `NEEDS_USER_OPS` 一行のみ
- ブラウザ IU 検証（TASK-010 / `BROWSER_VERIFICATION_*`）— residual 対象外
- Go-live の具体日付・実行者・連絡先 / 価格・契約・課金（#258 U1〜U12）/ 実 identity・approver 実名
- GitHub Issue の close 操作・`VERIFIED_FIXED` 付与

---

## 7. Conflicts & assumptions

### Opus pack との差分

| # | 項目 | Opus | Fable | 種別 |
|---|------|------|-------|------|
| 1 | LINE-R05 条件②（composition test 更新） | 未達・「旧契約のまま失敗」 | **充足済み**（`fac8c86b2` 2026-07-31・実測 green）。Opus 記載は起草時点で stale | 事実訂正（verdict 不変） |
| 2 | D-021-A | APPROVE（着手前） | landed 事実の**事後追認** + DEC-68 起票を明示要求 | 状態変化の反映 |
| 3 | 021-B の証拠 | 「access log / client registry」並記 | B は access log では証明不可（field は既存 GET に同乗）→ **client registry 必須** と明確化 | TIGHTEN |
| 4 | 021 恒久 open 化リスク | 言及なし | 90 日無応答 → ACCEPT_RESIDUAL_RISK 再裁定条項（F-021-X）を新設 | 新議題 |
| 5 | #257 gate リスト | DEC-60 の 8 Issue | **#252 を追加候補**として USER 確認を要求 | TIGHTEN 提案 |
| 6 | READY_AGENT | UNIT-021-A（1 件） | landed 済みのため製品 unit は **NONE**。docs unit（DEC-68）のみ条件付き | 状態変化の反映 |

### 最小の仮定

| # | 仮定 | 検証 | 外れた場合 |
|---|------|------|-----------|
| A1 | presence guard（`legacyCredentialPresent`）は fail-closed reject として機能 | 実測済み（`line_link_service.go:416-423`） | F-LINE-R05-b の論拠再検討 |
| A2 | `excluded_type_ids` の reject（Create 非空 / Update 存在）は 021-A 直前まで live だった | slice2 §1.3 + 021-A report | F-021-A を HOLD へ差し戻し REVERT 検討 |
| A3 | open Issue 18 件・claim/* 0 は現況 | STATUS 実測表（2026-08-06） | §1 表の再照合 |
| A4 | #252 は go-live 前に投入必須の会計系設定である | Issue #252 本文で USER 確認 | gate 追加候補を取り下げ |

---

FABLE_PO_RECOMMENDATION_PACK complete
ratify_opus_count: 20
overturn_opus_count: 0
ready_agent_if_user_yes: 1
clinical_blockers: 3
user_checklist_items: 16
