# Claude Fable — Needs Human 29 件 代行回答パック（2026-08-14）

| 項目 | 値 |
|------|-----|
| **依頼** | [`fable-needs-human-request-2026-08-14.md`](fable-needs-human-request-2026-08-14.md) |
| **OVERTURN** | **0** |
| **Done 遷移指示** | **0** |
| **発明した値** | **0** |
| **対象** | 電カル Needs Human BRT-37〜52 · 55〜67（谷口除外） |

---

## A. 読了・環境

読んだもの: 依頼書全文 · fable-po-recommendation.md（採択済み pack）全文 · Sol r2 全文 · Fable 確認回答 3 本（confirm / uat-human / exec-session）全文 · product-philosophy.md（§0・§5 該当行を実測確認、全文は pack 起草時読了）· UAT FINAL.md 全文 · DELIVERY_PACKAGE.md（前提・U1〜U13 表・U13=#256 所有 DEC-62）· handoff BRT-37〜52 / 55〜67 全 29 本 · walk 一式（linear-all-walk / todo-docs-linear-map / github-linear-map / staging-preflight-status / po10-local-presence）· drafts（E-1/E-5/E-11 全文、E-2/E-3/E-9/E-10 は同一内容を Sol r2 §E 原文で読了）· root todo.md / todo-po.md（ポインタ確認）。

読めなかったもの: STG/PROD runtime・DB・secret（接続禁止）· Linear 実 state（依頼書 §0 と walk 報告を採用）。GH live の再取得は PR #299 のみ（read-only 1 回）。

### 前提の訂正（3 点）

1. **P1 の「pull」は消化済み** — worktree tip は `45c8c8155` で `697d5c597`（001_init.sql +36 行）を既に含む。P1 残は `make migrate`（mismatch なら `make reset`）から。migrate/reset 実施済みかは未確認（DB 非接続）。
2. **PR #299 の CI FAILURE を本日 read-only で再実測し一致確認** — OPEN・draft・MERGEABLE・Backend/Frontend/Codegen Sync FAILURE・Detect/Gitleaks/AgentShield/Worker Tests SUCCESS。旧 M4 文面は現況に合わず、§E M4 を FAILURE 解消前置き版に更新した。
3. **local audit COUNT=373 は H3 の PASS 証拠にしない** — `make reset` で変わる数字であり、稼働の傍証以上に使わない（BRT-60 行で明文化）。

---

## B. トップ 3（今日 USER が Yes すべき）

1. **M1〜M5 の 5 通を今日、M2→M1→M4→M3→M5 の順で送る**（§E 完成文・値は全列空欄のまま）— BRT-66（M1=BRT-39 · M2=BRT-47 · M3=BRT-49 · M4=BRT-55 · M5=BRT-56）。
2. **#299 の CI FAILURE 3 本**（Backend/Frontend/Codegen Sync）の原因解消をリリース責任者に指示し、preflight 残 3 セル（checksum/ownership・backup+rollback owner・CI green）を埋め始める — BRT-55（green 前 merge は禁止のまま）。
3. **P1 残**（`make migrate`→mismatch なら `make reset`→postflight→実施 SHA 記録）を実行し、local の H3/H5/H7 を今日消化する — BRT-65 / BRT-60 / BRT-62 / BRT-64。

---

## C. 絶対 No トップ 3

1. **CI FAILURE 中の #299 merge・squash・staging reset** — required CI 全 green + 残セル記入 + draft 解除より前の merge は一切不可（BRT-55）。
2. **#254/#253/#256 の完了扱い close**、および「local FAIL 0=納品完了」と読める一切の対外表明 — H1〜H3 は disposition で代替できず、U13 未確定・PROD 未構築のまま閉じない（BRT-45/BRT-44/BRT-47）。
3. **値の代理記入と破壊操作** — 臨床 bundle/range/実 row の値発明（BRT-39/41/40）、TASK-033 骨格先行、presence ゼロ確認前の LINE-R05 DROP/guard 除去（BRT-57）、TASK-021 C/D の削除実行。

---

## D. チケット別回答表（29 行）

| Linear | Verdict | 1 文の答え（USER 代行） | 次の一手（誰が） | 送付/コメント | 発明 |
|--------|---------|------------------------|------------------|---------------|------|
| BRT-39 | NEEDS_CLINICAL | bundle 列は再掲どおり（対象・上限/warning policy・救急記録 policy 7 項・単位・出典・approver·opaque ref·発効日）。全列揃うまで TASK-033 禁止。値は臨床専権。催促は今日 M1 の 1 通のみ。 | USER→M1→臨床 | M1 §E-1 | No |
| BRT-41 | NEEDS_CLINICAL | 未判定維持・承認前 unit 起票禁止維持。行テンプレ E-2。一般値・別測定系流用しない。 | 臨床→E-2 | E-2 | No |
| BRT-40 | NEEDS_CLINICAL | 分離テンプレ E-3（Clinical/OPS）。両行完成まで実 row は local 含め禁止（F-211）。 | 臨床→OPS | E-3 | No |
| BRT-51 | RATIFY | #201 opaque ref 参照のみで close。値複製禁止。残 5 項目は E-4 enum。 | PO（#201 後） | E-4 | No |
| BRT-49 | RATIFY | 全行クライアント所有（A）推奨。U9·U12 は #253 後。U13 は #256。M3 一括。 | 契約←M3 | M3 | No |
| BRT-47 | NEEDS_USER_OPS | COMPLETED/未完は代行不能。USER が 1 語。完了なら E-5 1 行で close、未完なら open 維持。 | 納品←M2 | M2 | No |
| BRT-56 | RATIFY | 集計 6 点すべて現行継続推奨。M5 で承認 or 修正 1 通。 | クライアント←M5 | M5 | No |
| BRT-37 | NEEDS_USER_OPS | agent 禁止。順序 DB→CF→LINE→JWT。証跡 1 行 E-12。 | セキュリティ | E-12 | No |
| BRT-38 | NEEDS_USER_OPS | #89 後のみ。マスク・session 無効化・scanner・close の 6 段。 | セキュリティ | - | No |
| BRT-57 | NEEDS_USER_OPS | DROP/guard 除去はゼロ前禁止。STG 件数のみ（PROD は #253 後）。local 0 は代替にしない。 | DB 運用 | E-14 | No |
| BRT-44 | HOLD | PROD 未構築。Environment production 必須。U1〜U12 後に構築チェックリスト。 | 本番運用 | - | No |
| BRT-55 | RATIFY | CI FAILURE 中 merge 禁止（本日実測一致）。green 後: 残セル→draft 解除→merge-commit のみ。 | リリース←M4 | M4 | No |
| BRT-67 | NEEDS_USER_OPS | named env 非破壊 migrate 手順 1 枚確定。agent 禁止。STG のみ（PROD は #253 後）。 | 環境運用 | - | No |
| BRT-42 | WAIT_EXTERNAL | COMPLETE bundle=E-9 全項。partial 禁止。本日再催促なし。翌日以降未応答なら 1 通のみ。 | 移行元待ち | E-9 | No |
| BRT-50 | WAIT_EXTERNAL | E-10 非機密質問。gate OFF 維持。再催促規則は #250 と同一。 | 外部待ち | E-10 | No |
| BRT-46 | WAIT_EXTERNAL | roster 未着 HOLD。推測発行禁止。STG main 確認後〜#257 前。 | identity 運用 | - | No |
| BRT-43 | NEEDS_USER_OPS | #257 gate 入り RATIFY。値は会計/クライアント承認 1 行が正本。agent は値確定しない。 | 会計運用 | - | No |
| BRT-65 | TIGHTEN | pull 消化済。P1 残=migrate→reset→postflight→SHA。 | USER | - | No |
| BRT-66 | RATIFY | 5 通確定。順 M2→M1→M4→M3→M5。#250/#259 催促含めない。 | USER 送付 | §E | No |
| BRT-58 | HOLD | #299→STG health 後。実 LINE 1 往復。mock 不可。値非記録。 | UAT | E-8 | No |
| BRT-59 | HOLD | H1 同一 window。実 token LIFF。値非記録。 | UAT | E-8 | No |
| BRT-60 | TIGHTEN | COUNT=373 は PASS 証拠にしない。操作対応行で判定。disp 不可。 | USER/QA | - | No |
| BRT-61 | RATIFY | 締め 1 回 or ACCEPT_DISPOSITION。90 分持越し可。 | 会計 USER | §G | No |
| BRT-62 | RATIFY | 新規保存+再オープン永続で PASS。 | USER admin | - | No |
| BRT-63 | RATIFY | 持越し可。silent 省略禁止。fixture 不能時のみ disp。 | USER | - | No |
| BRT-64 | RATIFY | spot 4 件 ID 確定（S11·V02 inv·V03 owner·V03 perm-group F1）。H5 同席。 | USER | - | No |
| BRT-45 | HOLD | 今 close 不可。local FAIL0 は 1/8。H1–H3 代替不可。 | QA→別 USER | E-6 | No |
| BRT-48 | RATIFY | 新 window 今決めない。gate=DEC-60 の 8+#252（実残 7·CLOSED 除く）。 | リリース | - | No |
| BRT-52 | RATIFY | phase2 維持。実機 3 台 matrix で close。ソース完了≠close。 | 実機 QA | - | No |

### D 詳細（長文・要約で足りない行）

**BRT-39:** 対象（薬剤・動物種・master row ref・救急/既実施ケース）／上限 policy［現行継続/修正］／warning policy［現行継続/修正］／救急記録 policy 7 項（medicine identity·route·dose/strength/concentration 単位と必須性·weight/species snapshot·reason taxonomy·訂正 rationale·create grant role）／単位・出典・approver role·opaque ref·発効日。全列揃うまで TASK-033 禁止（DEC-48）。

**BRT-56 6 点:** ① annual_visit_count=365d rolling ② annual_amount=From/To→Year→preset→全期間 ③ CSV 標準で付けない ④ last_visit と休眠は別 ⑤ L-step write default-off·実 2xx のみ成功 ⑥ cleanup 失敗は通知必須。

**BRT-37 証跡 1 行:** `#89/#97 rotation: db=[ ] cloudflare=[ ] line=[ ] app_keys=[ ] old_values_revoked=[ ] sessions_invalidated=[ ] health=[ ] scan=[ ] history=NO_REWRITE owner_role=[ ] opaque_ref=[ ]`

**BRT-55 green 後:** ① CI 3 本解消 ② checksum/PS109 ownership disposition（STG reset 不可）③ backup+rollback owner ④ draft 解除 ⑤ merge-commit のみ（人間）。

**BRT-64 4 件:** ① S11 trimming 新規 ② V02 inventory F4 ③ V03 owner F4 ④ V03 permission-group F1（空名拒否+正名成功）。

**BRT-48 gate:** #89/#97/#98/#99/#250/#253/#254/#255+#252。うち #98/#99 CLOSED 済み → 実残 7。

---

## E. 送付文パック（コピー用）

送付順: **M2 → M1 → M4 → M3 → M5**。

### M1 — #201 臨床（BRT-39 / BRT-66）· 宛先: 臨床責任者

**件名:** 【#201】救急投薬 臨床承認 bundle 記入のお願い（期限 [ ]）

#201 に記入用の bundle 表を投稿済みです（Issue #201 のコメント参照）。全列が空欄のまま止まっており、この値はあなたにしか書けません。[ ] までに全列の記入をお願いします。

- 対象薬剤・動物種・master row reference: [ ] ／ 救急・既実施投薬の対象ケース・対象薬剤: [ ]
- 上限 policy: [現行継続 / 修正] ／ warning policy: [現行継続 / 修正]
- 救急記録 policy（medicine identity / route vocabulary / dose・strength・concentration の単位と必須性 / weight・species snapshot / reason taxonomy または bounded free-text / 訂正対象・rationale / create grant 対象 role）: [ ]
- 数値ごとの単位（非数値は「該当なし」）: [ ] ／ 出典: [ ] ／ approver role: [ ] ／ opaque ref: [ ] ／ 発効日: [ ]

上限・warning は「現行継続」と書くだけでも構いません（再計算不要）。実氏名・患者情報・credential・承認資料本文は返信に書かないでください。全列が揃うまで実装（TASK-033）は着手しません。

**禁止:** 「こちらで仮の値を入れておきます」「だいたいで OK」。

### M2 — #256 U13（BRT-47）· 宛先: 納品責任者

**件名:** 【#256】U13 操作説明会 — 完了 / 未完 の 1 語だけください

#256 の close に残る判断セルは U13（操作説明会）だけです。「完了」または「未完」の 1 語でご回答ください。完了なら実施日・発効日 [ ]・opaque ref [ ] を添えてください。未完ならそれだけで結構です（Issue は open 維持。日程はまだ決めなくてよい）。

**禁止:** 「未完でも close で進めます」。

### M3 — #258 U1〜U12（BRT-49）· 宛先: 契約責任者

**件名:** 【#258】DELIVERY_PACKAGE U1〜U12 一括記入のお願い

添付の記入表（`reports/todo-walk-2026-08-14/drafts/E-11-258.md`）に沿って、A/B 選択と U1〜U8 / U10 / U11 をご記入ください。推奨は A（クライアント所有）です。B は料金・期間・責任・終了時移管の明示契約が揃うまで選びません。値の正本は DELIVERY_PACKAGE.md U1〜U12 だけとし、他の台帳へ複製しません。**U9・U12 は Production 構築後（#253）に開発側が記入するため空欄のままで構いません。**U13 は #256 側で扱うため本依頼に含みません。金額・secret・実 email・実 identity・Go-live 日付・receipt 本文は記入しないでください。

**禁止:** 「納品完了に伴い」（PROD 未構築）。

### M4 — #299 preflight / CI（BRT-55）· 宛先: リリース責任者 + STG DB 運用

**件名:** 【#299】CI FAILURE 3 本の解消 → 残り 3 セル → merge 判断（merge-commit のみ）

draft PR #299（main→staging）は現在、required CI のうち Backend / Frontend / Codegen Sync が FAILURE、Detect / Gitleaks / AgentShield / Worker Tests は SUCCESS です。**green になるまで merge しません。**次の順でお願いします。

1. CI FAILURE 3 本の run log 確認と原因解消（修正が必要な場合は別タスク化）
2. migration checksum / PlanetScale 109 オブジェクト所有権の disposition 1 行確定: [ ]（STG_PLANETSCALE_SEED_RUNBOOK 準拠・STG reset は不可）
3. deploy 前 backup 実施 + rollback owner role の記入: [ ]
4. 全 green 後に draft 解除 → merge-commit のみ（squash・直接 merge 禁止・実行は人間）

あわせて PlanetScale サポートへ 109 オブジェクトの REASSIGN 依頼を先行送付してください（リードタイム長・今日送ってよい）。

**禁止:** 「とりあえず merge して STG で確認」。

### M5 — PO-008（BRT-56）· 宛先: クライアント仕様責任者

**件名:** 【顧客集計】集計仕様 6 点の承認または修正のお願い

現行仕様 6 点について、承認か修正指示を 1 通でご回答ください（推奨は現行継続です）。

1. annual_visit_count = 直近 365 日 rolling
2. annual_amount = From/To → Year → preset → 全期間の優先順
3. CSV 出力は標準では付けない
4. last_visit と休眠判定は別ロジックのまま統一しない
5. L ステップ書込は default-off・実送信成功のみ成功扱い
6. cleanup 失敗は通知必須（本体削除は止めない）

修正の場合は、その項目の業務上の目的と決裁者名 [ ] を添えてください。

**禁止:** 「CSV も付けておきます」。

### 補助 — PO-10 presence window（BRT-57）· 宛先: DB 運用責任者

STG の `line_reservation_settings` に read-only の承認 window を 1 回設定し、E-14 の SQL（値非表示・clinic 件数のみ）で集計してください。raw row・clinic ID·secret 値·digest は保存せず、結果は `environment=[ ] clinics_with_legacy_presence=[count] operator_role=[ ] opaque_ref=[ ]` の 1 行のみ。present>0 が 1 件でもあれば LINE-R05 は HOLD 継続です。PROD 分は #253 後。local 0 件は STG の代替になりません。

---

## F. #299 / #254 / #256 close 判定カード

| ID | 今 close/merge してよいか | Yes の最小条件 | No のとき残す 1 行 |
|----|---------------------------|----------------|-------------------|
| **#299 merge** | **No**（CI 3 本 FAILURE） | required CI 全 green + checksum/ownership + backup + rollback owner + draft 解除 → merge-commit のみ・人間 | #299 は CI 3 本 FAILURE 中のため merge 禁止 — green + 残 3 セル後に draft 解除し merge-commit のみ（squash 禁止） |
| **#254 close** | **No**（local FAIL0 は 1/8） | H1/H2 + H3 + H4〜H7（実施 or disp）+ E-6 全欄 + 受入 build SHA + residual_disposition=APPROVED + 別 USER final sign-off | #254 は local FAIL 0 では閉じない — H1〜H3 実施 + H4〜H7 disposition + E-6 全欄 + 別 USER 承認まで open |
| **#256 close** | **No**（U13 未記入） | U13_status=COMPLETED + 発効日 + opaque ref + 別 USER close 承認（E-5） | #256 は U13 の 1 語（完了/未完）待ち — 未完なら open 維持・close 条件は E-5 のまま |

---

## G. 90 分セッション（ケース B）更新

順序 **P1（残）→ M1〜M5 → H3 → H5 → H7** を最終確定。変更点のみ:

1. P1 の「WIP commit→pull」は消化済み — P1 残 = `make migrate` →（mismatch なら）`make reset`（snapshot 自動·USER のみ）→ postflight green → 実施 HEAD SHA 記録、から開始。済みなら postflight 確認のみで H3 へ。
2. M 送付は reset 並走中に実施、内部順序 **M2→M1→M4→M3→M5** — M4 は §E の CI FAILURE 現況版。
3. H3 は COUNT=373 を証拠にしない — 操作対応行（作成/更新/確定+actor_id+確定行の old/new_value）で判定。
4. H4·H6 は明示持ち越し（SESSION.md·silent 省略禁止）。disposition は H4〜H7 のみ可·H1〜H3 不可。
5. 証跡 `reports/uat-human-2026-08-14/` · 各ファイル冒頭に build SHA · merge/close/値記入/再催促ゼロは変更なし。

---

## H. Linear コメント用 1 行（29 件）

```
BRT-37: NEEDS_USER_OPS — 4系統ローテUSER専権・証跡E-12
BRT-38: NEEDS_USER_OPS — #89後にマスク・session無効化
BRT-39: NEEDS_CLINICAL — bundle全列待ち・催促M1送付
BRT-40: NEEDS_CLINICAL — 実row禁止維持・E-3分離記入
BRT-41: NEEDS_CLINICAL — range未判定維持・起票禁止
BRT-42: WAIT_EXTERNAL — 完全bundle待ち・本日再催促なし
BRT-43: NEEDS_USER_OPS — 投入はUSER・#257gate入り追認
BRT-44: HOLD — PROD未構築・U1〜U12後に構築
BRT-45: HOLD — local FAIL0では閉じない・H残
BRT-46: WAIT_EXTERNAL — roster未着・推測発行禁止
BRT-47: NEEDS_USER_OPS — U13完了/未完の1語待ち（M2）
BRT-48: RATIFY — 新window今決めない・gate計9本
BRT-49: RATIFY — A推奨・M3送付・U9/U12後埋め
BRT-50: WAIT_EXTERNAL — 先方enable待ち・gateOFF維持
BRT-51: RATIFY — #201参照のみclose・複製禁止
BRT-52: RATIFY — phase2維持・実機3台でclose
BRT-55: RATIFY — CI赤中merge禁止・green後手順
BRT-56: RATIFY — 6点現行継続推奨・M5で承認依頼
BRT-57: NEEDS_USER_OPS — STG件数のみ・DROP禁止
BRT-58: HOLD — #299→STG後に実LINE通知
BRT-59: HOLD — H1同一window・token非記録
BRT-60: TIGHTEN — 373は証拠外・操作対応行で判定
BRT-61: RATIFY — 締め1回or disp・持越し可
BRT-62: RATIFY — 新規保存+再オープン永続でPASS
BRT-63: RATIFY — 持越し可・silent省略禁止
BRT-64: RATIFY — 4件ID確定・H5同席実施
BRT-65: TIGHTEN — pull済・migrateから開始
BRT-66: RATIFY — 5通確定・M2→M1→M4→M3→M5
BRT-67: NEEDS_USER_OPS — named env非破壊migrate手順
```

---

## 品質ゲート（§5）

| 項目 | 結果 |
|------|------|
| 29 行全記入・欠番なし | ✅ |
| 薬用量·range·credential·Go-live 日付の発明なし | ✅ |
| #299 green 前 merge Yes なし | ✅ |
| #254 の local FAIL0 単独 close なし | ✅ |
| TASK-033 骨格先行 Yes なし | ✅ |
| Done 遷移指示なし | ✅ |
| 谷口 BRT-53/54 非対象 | ✅ |
| 日本語·転記可能文 | ✅ |

Fable セッションの実行範囲: repo 内読み取り · `gh pr view 299`（read-only 1 回）· 本回答ファイル作成のみ。コード変更·merge·migrate·secret·外部送信·Linear/GitHub 状態変更は Fable 側では行っていない。
