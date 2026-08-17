# Claude Fable — PO 確認回答（2026-08-14）

全必読ファイルと参考資料の確認、local 実測が完了した。§5 の見出し順で回答する。

---
A. 読んだもの / 読めなかったもの

読んだもの

- 依頼書 reports/fable-po-confirm-request-2026-08-14.md
- todo-po.md 全文（PO 正本）
- todo.md 全文（§1〜3・§7 含む）
- reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md 全文（82 行 + 完成物 14 本）
- docs/product-philosophy.md 全文
- docs/work/decisions/fable-po-recommendation.md 全文（自分の採択済み pack）
- reports/uat-2026-08-14/FINAL.md 全文（PASS 96 · PARTIAL 4 · BLOCKED 5 · FAIL 0）
- docs/delivery/DELIVERY_PACKAGE.md 全文（U1〜U12 表含む）
- docs/ops/infra/staging/runbook.md 全文（29 行。squash 禁止理由・PlanetScale 109 オブジェクト所有権 blocker を確認）
- phase2.html — 見出し構造 + #284 / #249 の grep 実測（全文通読はしていない。#284 の記載なし・#249 関連は select 異常ハイライト 1 件のみを確認）
- local git 実測（fetch なし・read-only）: main tip 1386e1db0（todo.md §1 と一致）· origin/staging...main = 4 / 1346（Sol E-7 の参考値と一致）

読めなかった / 読まなかったもの

- GitHub live state — 操作禁止のため未取得。todo-po.md / todo.md §7.1 を正とした
- docs/work/residual-closeout-ledger.md — 存在確認のみ（証跡置き場であり判定に内容不要）
- STG / PROD runtime・DB・secret・外部サービス — 接続していない

B. 再審

再審なし。DEC-40〜68 と Fable pack（2026-08-06 採択）を維持。2026-08-14 の local UAT FAIL 0 も local ref 実測も、既裁定の境界を覆す新証拠ではない。

C. Executive

今日 Yes すべきトップ 3

1. PO-11 / #201 — E-1 を臨床責任者へ送付。臨床応答が全 residual 中の最長リードタイムであり、TASK-033・#261 の共通解除条件だ。
2. staging ← main preflight — E-7 実施。staging は main から 1346 commit 遅れで、migration 統合（001 単一化）と PlanetScale 所有権 blocker を跨ぐ。実 LINE・OPS-3/4/5/14・#252 preview のすべてがこの後ろに並んでいる。
3. #256 U13 disposition — 完了/未完の明示。dual sign-off 済みで、残る判断セルは U13 だけだ。

絶対に今 Yes すべきでないトップ 3

1. TASK-033 着手（#201 bundle 承認前の骨格先行）— DEC-48 の一体所有に正面から反する。
2. TASK-021 C/D・LINE-R05 の削除実行（registry / presence inventory ゼロ確認前）— fail-closed guard の除去と不可逆 DROP。gate 前禁止は動かない前提でもある。
3. #257 新 window 設定、および #253 / #254 の完了扱い close — PROD 未構築・staging に current main 未搭載の状態で、local FAIL 0 は閉証拠にならない。DELIVERY_PACKAGE の「STG 稼働中」表記を「検証済み」と読み替えるな。

Sol r2 との関係: RATIFY 79 件 / TIGHTEN 4 件 / OVERTURN 0 件（全 83 行 = Sol 82 行 + PO-008 の CSV / last-visit を依頼どおり分割）

D. 全件確認表

TIGHTEN は No.18（#284）· No.20（TASK-021 C/D）· No.28（#258）· No.68（#249 外部） の 4 行。他は全行 RATIFY。

ID: 1 · PO-11/#201
Sol: DO_NOW
Fable: DO_NOW
関係: RATIFY
POの答え（1文）: bundle 1 通を臨床責任者へ送り、現行 20%
は「現行継続を推奨」に留めて医学的正本化しない（DEC-48/65）
次の人: 臨床責任者
次の一手: §E-1 送付
空欄に残すセル: 対象・policy・単位・出典・approver role・発効日・opaque ref
────────────────────────────────────────
ID: 2 · staging preflight
Sol: DO_NOW
Fable: DO_NOW
関係: RATIFY
POの答え（1文）: local ref 実測 4/1346（本日一致確認）を起点に staging-only 差分・migration
checksum・PlanetScale 所有権 blocker・rollback を merge 前に確定する
次の人: リリース責任者
次の一手: §E-7 実施
空欄に残すセル: 実施者 role・時刻・CI run・backup/rollback ref
────────────────────────────────────────
ID: 3 · staging merge
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: preflight 全 green 後に merge-commit PR のみで取り込む — squash は祖先関係を切ると
runbook 自身が警告している
次の人: リリース責任者
次の一手: preflight 後に PR
空欄に残すセル: PR・merge SHA・CI・health ref
────────────────────────────────────────
ID: 4 · #256 U13
Sol: DO_NOW
Fable: DO_NOW
関係: RATIFY
POの答え（1文）: U13 完了/未完を明示し、未完なら §E-5 を送らず open 維持
次の人: 納品責任者
次の一手: U13 状態確定
空欄に残すセル: U13 status・発効日・close approval・opaque ref
────────────────────────────────────────
ID: 5 · 実LINE UAT
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: current main の STG deploy・health 後にのみ実通知・LIFF・audit を観測し、local mock
を代替にしない
次の人: UAT 責任者
次の一手: §E-8 実施
空欄に残すセル: clinic role・時刻・各結果・opaque ref
────────────────────────────────────────
ID: 6 · PO-10/LINE-R05
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: clinic 別 presence 件数のみ取得し、ゼロ確認前の guard 除去・DROP を禁止する
次の人: DB 運用責任者
次の一手: §E-14 実施
空欄に残すセル: 環境・件数・operator role・opaque ref
────────────────────────────────────────
ID: 7 · PO-12/#249
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 未承認 range は「未判定」を維持し、承認依頼のみ送る
次の人: 臨床責任者
次の一手: §E-2 送付
空欄に残すセル: range 全欄
────────────────────────────────────────
ID: 8 · PO-13/#211
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 臨床行と OPS 行を分離し、両行完成まで実 row 投入は local 含め禁止
次の人: 臨床責任者
次の一手: §E-3 送付
空欄に残すセル: clinical 全欄・env・operator・dry-run/apply/rollback
────────────────────────────────────────
ID: 9 · PO-16/#261
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: #201 opaque ref + runtime 5 項目が揃うまで open 維持、#201 値は複製しない
次の人: PO
次の一手: #201 後に §E-4
空欄に残すセル: bundle ref・5 結果 enum・各 ref
────────────────────────────────────────
ID: 10 · #254 close
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: FAIL 0 でも PARTIAL 4・BLOCKED 5 が残るため §E-6 全 green まで閉じない
次の人: QA 責任者
次の一手: §E-6 実測
空欄に残すセル: window・operator・各結果・sign-off・ref
────────────────────────────────────────
ID: 11 · #256 close
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: U13 + 発効日 + opaque ref + 別 USER 承認が揃った時点でのみ close（no-rewrite）
次の人: PO
次の一手: §E-5 使用
空欄に残すセル: U13・発効日・approval・ref
────────────────────────────────────────
ID: 12 · PO-17
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: named env のみ非破壊 migrate し、STG/PROD reset はしない
次の人: 環境運用責任者
次の一手: 対象 env で手動実行
空欄に残すセル: env・結果・operator role・ref
────────────────────────────────────────
ID: 13 · PO-18/#89·#97
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 4 系統 rotate/revoke/session/health/scan 完了まで close しない
次の人: セキュリティ責任者
次の一手: §E-12 実施
空欄に残すセル: 区分・window・各結果・owner role・ref
────────────────────────────────────────
ID: 14 · PO-19/#253
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: PROD 未構築のため CI/CD・backup・restore を完了扱いしない
次の人: 本番運用責任者
次の一手: U1〜U12 後に構築
空欄に残すセル: 各結果・authority・ref
────────────────────────────────────────
ID: 15 · PO-20/#257
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 全 gate green 前に Go-live 日付を置かず、旧 window は No-Go のまま
次の人: リリース責任者
次の一手: gate 後に window
空欄に残すセル: window・各 role・ref
────────────────────────────────────────
ID: 16 · #250 催促
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: complete bundle 受領まで partial apply 禁止、催促のみ送る
次の人: 移行元責任者
次の一手: §E-9 送付
空欄に残すセル: bundle ID・complete 判定・authority
────────────────────────────────────────
ID: 17 · #259 催促
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 先方 enable まで両 gate default-off 維持、催促のみ送る
次の人: 外部連携責任者
次の一手: §E-10 送付
空欄に残すセル: 各 gate・結果・ref
────────────────────────────────────────
ID: 18 · #284
Sol: DEFER
Fable: DEFER
関係: TIGHTEN
POの答え（1文）: DEFER_PHASE2 は維持するが、今期外の正本 phase2.html に #284 が未記載（本日 grep 実測）—
台帳へ 1 行追記して drift を閉じよ
次の人: PO
次の一手: phase2.html へ 1 行追記（手作業 1 分・unit 不要）
空欄に残すセル: environment・devices・結果・ref
────────────────────────────────────────
ID: 19 · TASK-009 他env
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: STG health 後に対象 env・承認 bundle が確定した場合のみ検討する
次の人: DB 運用責任者
次の一手: bundle/checksum 確認後に USER 判断
空欄に残すセル: env・bundle・結果・approval ref
────────────────────────────────────────
ID: 20 · TASK-021 C/D
Sol: DEFER
Fable: DEFER
関係: TIGHTEN
POの答え（1文）: B→C→D 順序と gate は維持し、加えて採択済み F-021-X の 90 日時計（PO-09:
inventory_start=2026-08-09）により 2026-11-07 までに inventory 無応答なら ACCEPT_RESIDUAL_RISK
再裁定へ上げる
次の人: PO
次の一手: 期限管理し B→C→D 順に解除
空欄に残すセル: 各 gate 結果・ref
────────────────────────────────────────
ID: 21 · TASK-033
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: #201 bundle と一体でのみ着手し、骨格先行を認めない（DEC-48）
次の人: 臨床責任者
次の一手: §E-1 完成
空欄に残すセル: #201 全臨床欄・承認 ref
────────────────────────────────────────
ID: 22 · POST-PULL
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: migration を含む commit の pull 後は利用前に USER が make migrate を実行する
次の人: 開発環境利用者
次の一手: 該当 pull 後に手動実行
空欄に残すセル: commit・env・結果
────────────────────────────────────────
ID: 23 · DR-CLINICAL #201
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 上限/warning は現行継続を推奨するが、出典付き承認までは未承認として扱う
次の人: 臨床責任者
次の一手: bundle 1 行記入
空欄に残すセル: 全臨床値・policy・出典・承認欄
────────────────────────────────────────
ID: 24 · DR-CLINICAL #261→#201
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: opaque ref 参照のみとし、独立値が要る場合は同列別行の追加であって複製ではない
次の人: 臨床責任者
次の一手: #201 後に参照承認
空欄に残すセル: ref・承認・独立行理由
────────────────────────────────────────
ID: 25 · DR-CLINICAL #211
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 未承認のまま「未判定」を維持する
次の人: 臨床責任者
次の一手: §E-3 臨床行記入
空欄に残すセル: row key・type/range・単位・出典・承認
────────────────────────────────────────
ID: 26 · DR-CLINICAL #249
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 未承認のまま「未判定」を維持し、一般値・別測定系を流用しない
次の人: 臨床責任者
次の一手: §E-2 記入
空欄に残すセル: range 全欄
────────────────────────────────────────
ID: 27 · VACCINE-SPECIES
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: master 明示属性のみを権威とし、未承認 mapping は「未判定」のまま silent 非表示にしない
次の人: 臨床責任者
次の一手: 適合性行記入
空欄に残すセル: master row ref・species・alias・出典・承認者・発効日
────────────────────────────────────────
ID: 28 · #258 A/B
Sol: DO_NEXT
Fable: DO_NEXT
関係: TIGHTEN
POの答え（1文）: クライアント所有（A）を推奨するが、送付文 E-11 は削除済み q&a.html
を参照しており修正版（§E）を使え
次の人: 契約責任者
次の一手: REVISE 版 E-11 送付
空欄に残すセル: A/B・契約責任者・client approver・発効日・ref
────────────────────────────────────────
ID: 29 · U1 Cloudflare
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有（A）を推奨
次の人: 契約責任者
次の一手: U1 記入
空欄に残すセル: 契約名義・請求先・移管有無
────────────────────────────────────────
ID: 30 · U2 PlanetScale
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有（A）を推奨し、plan・保持は契約責任者が決め発明しない
次の人: 契約責任者
次の一手: U2 記入
空欄に残すセル: plan・backup 頻度/保持・契約名義
────────────────────────────────────────
ID: 31 · U3 Vercel
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、registrar 権限をクライアント管理下に置く
次の人: 契約責任者
次の一手: U3 記入
空欄に残すセル: plan・契約名義・registrar 権限
────────────────────────────────────────
ID: 32 · U4 GitHub
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、client org・最小権限・退任時 revoke を方針にする
次の人: Repository 責任者
次の一手: U4 記入
空欄に残すセル: organization・role・Collaborator 方針
────────────────────────────────────────
ID: 33 · U5 LINE
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、値は secret 管理へのみ投入する
次の人: 各医院 LINE 管理者
次の一手: U5 記入
空欄に残すセル: owner role・結果・ref（値は書かない）
────────────────────────────────────────
ID: 34 · U6 Lステップ
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、API key は secret 管理のみで扱う
次の人: 各医院連携管理者
次の一手: U6 記入
空欄に残すセル: owner role・結果・ref（key は書かない）
────────────────────────────────────────
ID: 35 · U7 support
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、一次窓口とエスカレーション条件を契約で明確化する
次の人: サポート責任者
次の一手: U7 記入
空欄に残すセル: 連絡手段・受付時間・一次対応 role
────────────────────────────────────────
ID: 36 · U8 監視通知
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、実 address は repo/chat に書かない
次の人: 監視責任者
次の一手: U8 記入 + CF 検証
空欄に残すセル: address（restricted）・検証結果・ref
────────────────────────────────────────
ID: 37 · U9 本番backup
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: A 推奨だが PROD 構築後の restore 実測まで未完として維持する
次の人: 本番運用責任者
次の一手: 構築後に実測
空欄に残すセル: 頻度・保持・手順・所要時間・ref
────────────────────────────────────────
ID: 38 · U10 R2
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、採否・保持・復旧責任を確定する
次の人: ストレージ責任者
次の一手: U10 記入
空欄に残すセル: 採否・保持・復旧方針・承認 ref
────────────────────────────────────────
ID: 39 · U11 audit保持
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: クライアント所有(A)を推奨し、保持年数・廃棄方針は先方が承認する
次の人: データガバナンス責任者
次の一手: U11 記入
空欄に残すセル: 保持年数・廃棄方針・承認 ref
────────────────────────────────────────
ID: 40 · U12 PROD証跡
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 実構築・URL health・rollback 証跡が揃うまで空欄=未構築を維持し、偽の証跡を書かない
次の人: 本番運用責任者
次の一手: 構築後に記入
空欄に残すセル: setup 結果・URL health・rollback・ref
────────────────────────────────────────
ID: 41 · annual_visit_count
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 現行継続を推奨（365 日 rolling）し、client 承認まで他指標と統一しない
次の人: クライアント仕様責任者
次の一手: 承認/修正回答
空欄に残すセル: decision・修正値・承認者・発効日
────────────────────────────────────────
ID: 42 · annual_amount
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 現行継続を推奨（From/To→Year→preset→全期間）し、365 日に自動統一しない
次の人: クライアント仕様責任者
次の一手: 承認/修正回答
空欄に残すセル: decision・修正値・承認者・発効日
────────────────────────────────────────
ID: 43 · CSV
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: 顧客集計 CSV は default 追加せず、必要なら責任者名 + 業務目的付きの新要件として出させる
次の人: クライアント仕様責任者
次の一手: 必要時のみ新要件起票
空欄に残すセル: override 理由・責任者・指標
────────────────────────────────────────
ID: 43b · last-visit/dormant
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: last_visit bucket（90/180/365）と Lステップ
dormant（180/210/240/365）は別目的のまま統一しない
次の人: クライアント仕様責任者
次の一手: 現状維持を確認
空欄に残すセル: override 理由・責任者
────────────────────────────────────────
ID: 44 · L-step 通常同期
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 両 gate default-off・実 2xx のみ成功を維持し、write 失敗は処理失敗として通知する
次の人: 外部連携責任者
次の一手: §E-10 の enable 条件成立後
空欄に残すセル: 承認・通知先・rollback ref
────────────────────────────────────────
ID: 45 · L-step cleanup
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: 本体削除を止めない best-effort を維持しつつ、失敗通知必須で silent success を禁止する
次の人: 外部連携責任者
次の一手: path 別に承認
空欄に残すセル: 承認・通知・rollback ref
────────────────────────────────────────
ID: 46 · OPS-1
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: 4 系統ローテを USER が実施し、旧値失効まで open 維持
次の人: セキュリティ責任者
次の一手: §E-12 実施
空欄に残すセル: 各結果・window・owner role・ref
────────────────────────────────────────
ID: 47 · OPS-2
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: checksum mismatch は local のみ fresh、STG/PROD は非破壊 migrate + 起動証跡
次の人: DB 運用責任者
次の一手: env ごと手動実行
空欄に残すセル: env・結果・ref
────────────────────────────────────────
ID: 48 · OPS-3
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: STG health 後に 0-rule グループ件数のみ取得し、0 行なら close・hit なら backfill 別起票
次の人: DB 運用責任者
次の一手: 承認済み env で件数取得
空欄に残すセル: env・件数・disposition・ref
────────────────────────────────────────
ID: 49 · OPS-4
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: current main の STG health 後に実 LINE link E2E を実測する
次の人: UAT 責任者
次の一手: §E-8 実施
空欄に残すセル: 結果・ref
────────────────────────────────────────
ID: 50 · OPS-5
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: local PASS を反復せず、残る DB/audit・実 LINE・fixture・disposition のみ補う
次の人: UAT 責任者
次の一手: OPS-2/4 後に実施
空欄に残すセル: 各結果・ref
────────────────────────────────────────
ID: 51 · OPS-6
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: PROD 未構築のため設定済み扱いせず、構築時に VITE_SHOW_DEMO_ACCOUNTS=false を確認する
次の人: 本番運用責任者
次の一手: 構築後に確認
空欄に残すセル: 結果・ref
────────────────────────────────────────
ID: 52 · OPS-7
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: AWS IaC 退役済み close を維持し、旧 Terraform を apply しない
次の人: PO
次の一手: closed 維持
空欄に残すセル: なし
────────────────────────────────────────
ID: 53 · OPS-8
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: R2 公開 domain を推測せず、owner 確定後のみ投入する
次の人: ストレージ責任者
次の一手: domain 方針承認後
空欄に残すセル: domain・各結果・ref
────────────────────────────────────────
ID: 54 · OPS-9
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: 非 blocking 目視とし、NG のときだけ bug 化する
次の人: UI QA 責任者
次の一手: release 候補時に 1 回
空欄に残すセル: 結果・bug ID
────────────────────────────────────────
ID: 55 · OPS-10
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: 任意 full type-check を残件化しない
次の人: PO
次の一手: 台帳から外す
空欄に残すセル: なし
────────────────────────────────────────
ID: 56 · OPS-11
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: repo 外 Notion 目視を製品 blocker にしない
次の人: 文書責任者
次の一手: 正常なら close
空欄に残すセル: 結果
────────────────────────────────────────
ID: 57 · OPS-12
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: full PHI seed を default にせず small demo を維持する
次の人: PO
次の一手: 現状確定
空欄に残すセル: 例外承認 ref のみ
────────────────────────────────────────
ID: 58 · OPS-13
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: OPS-2 と同一 window 可だが証跡は独立に残す
次の人: DB 運用責任者
次の一手: 対象 env で実施
空欄に残すセル: env・結果・ref
────────────────────────────────────────
ID: 59 · OPS-14
Sol: DO_NEXT
Fable: DO_NEXT
関係: RATIFY
POの答え（1文）: staging PR 後の remote CI green + coverage artifact を release evidence にする
次の人: CI 責任者
次の一手: USER が CI 起動・確認
空欄に残すセル: run/ref・結果
────────────────────────────────────────
ID: 60 · OPS-15
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: PROD 未構築のため実値設定・deploy をしない
次の人: 本番運用責任者
次の一手: 構築 gate 成立後
空欄に残すセル: 各結果・ref
────────────────────────────────────────
ID: 61 · OPS-16
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: production 相当環境ができるまで rehearsal を完了扱いしない
次の人: 信頼性責任者
次の一手: 環境後に実測
空欄に残すセル: env・時刻・各結果・ref
────────────────────────────────────────
ID: 62 · OPS-17
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: test channel 準備後のみ redelivery/error statistics を有効化する
次の人: LINE 運用責任者
次の一手: rehearsal 実施
空欄に残すセル: channel role・時刻・結果・ref
────────────────────────────────────────
ID: 63a · OPS-18
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: Sentry は free-only・自動 paid upgrade 禁止を維持する
次の人: 監視責任者
次の一手: 責任者確定後に検討
空欄に残すセル: project・owner role・ref（DSN は書かない）
────────────────────────────────────────
ID: 63 · #89/#97 close
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: rotation・revoke・session・health・scan 全完了まで open 維持
次の人: セキュリティ責任者
次の一手: §E-12 後に close 承認
空欄に残すセル: 各結果・close approval・ref
────────────────────────────────────────
ID: 64 · #98
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: 台帳上 closed（§7.1）であり、live open の場合のみ §E-13 で close する
次の人: PO
次の一手: state 差異時のみ確認
空欄に残すセル: live state・close ref
────────────────────────────────────────
ID: 65 · #99
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: 同上 — 再作業せず、live open の場合のみ §E-13
次の人: PO
次の一手: 同上
空欄に残すセル: live state・close ref
────────────────────────────────────────
ID: 66 · #201
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: bundle と TASK-033 の安全な代替記録経路が green になるまで open 維持
次の人: 臨床責任者
次の一手: §E-1 完成
空欄に残すセル: bundle 全欄
────────────────────────────────────────
ID: 67 · #211 apply
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: 両行 + dry-run + rollback が揃うまで実 apply を local 含め HOLD
次の人: データ移行責任者
次の一手: §E-3 後に dry-run
空欄に残すセル: authorization・env・DB history・operator・各結果
────────────────────────────────────────
ID: 68 · #249 外部import
Sol: DEFER
Fable: DEFER
関係: TIGHTEN
POの答え（1文）: DEFER_PHASE2 は維持するが phase2.html に本件（#249-EXT）が未記載 — 採択済み pack の「Yes
→ phase2.html 記載」を履行せよ
次の人: 臨床プロダクト責任者
次の一手: phase2.html へ 1 行追記後、再開条件成立時に新 Issue
空欄に残すセル: 全欄
────────────────────────────────────────
ID: 69 · #250
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: formal producer bundle 受領まで Go-live gate として open 維持
次の人: 移行元責任者
次の一手: §E-9 回答待ち
空欄に残すセル: bundle ID・complete 判定・authority
────────────────────────────────────────
ID: 70 · #252
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: 批准済み値との差分のみ preview し、gate green 後に USER が投入する
次の人: 会計運用責任者
次の一手: staging 後に差分確定
空欄に残すセル: clinic ref・diff・preview・operator・window・ref
────────────────────────────────────────
ID: 71 · #253
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: PROD 未構築のため CI・deploy・backup・restore・rollback 実測まで open 維持
次の人: 本番運用責任者
次の一手: U1〜U12 後に構築・復旧試験
空欄に残すセル: 各結果・authority・ref
────────────────────────────────────────
ID: 72 · #254
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: PASS 96・FAIL 0 でも PARTIAL 4・BLOCKED 5 がある限り open 維持
次の人: QA 責任者
次の一手: §E-6 実測
空欄に残すセル: 各観測・sign-off・ref
────────────────────────────────────────
ID: 73 · #255
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: identity・clinic・permission を推測せず、repo 外 manifest の不明行のみ HOLD
次の人: identity 運用責任者
次の一手: restricted manifest を preflight → USER apply
空欄に残すセル: 結果 enum・role・opaque ref
────────────────────────────────────────
ID: 74 · #256
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: no-rewrite を維持し、U13 完了 + 別 close 承認が揃った時点で close
次の人: PO
次の一手: §E-5 使用
空欄に残すセル: U13・発効日・approval・ref
────────────────────────────────────────
ID: 75 · #257
Sol: HOLD
Fable: HOLD
関係: RATIFY
POの答え（1文）: #252 含む全 gate green と role 確定後のみ新 window を設定
次の人: リリース責任者
次の一手: gate evidence 収集
空欄に残すセル: role・window・ref
────────────────────────────────────────
ID: 76 · #258
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: A 推奨・U1〜U12 を DELIVERY_PACKAGE へ一度だけ記入するまで open 維持
次の人: 契約責任者
次の一手: REVISE 版 E-11 送付
空欄に残すセル: A/B・U1〜U12・承認・発効日・ref
────────────────────────────────────────
ID: 77 · #259
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: 先方 enable と STG live send・cron・stop/rollback 証跡が揃うまで open 維持
次の人: 外部連携責任者
次の一手: §E-10 回答待ち
空欄に残すセル: 各 gate・結果・ref
────────────────────────────────────────
ID: 78 · #260
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: dated plan hub は historical close とし、復活させない
次の人: PO
次の一手: live open なら USER が close
空欄に残すセル: owner role・close ref
────────────────────────────────────────
ID: 79 · #261
Sol: KEEP_OPEN
Fable: KEEP_OPEN
関係: RATIFY
POの答え（1文）: #201 参照と 5 項目の enum + opaque ref が揃うまで open 維持
次の人: PO
次の一手: §E-4 完成
空欄に残すセル: #201 ref・5 結果・close approval
────────────────────────────────────────
ID: 80 · #284（Issue 行）
Sol: DEFER
Fable: DEFER
関係: RATIFY
POの答え（1文）: 裁定は行 18 と同一（phase2.html 追記の TIGHTEN は行 18 側に付した）
次の人: 実機 QA 責任者
次の一手: trigger 成立時に新 unit
空欄に残すセル: 結果・ref
────────────────────────────────────────
ID: 81 · #212/#235
Sol: CLOSE_RECOMMEND
Fable: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: DEC-66/67 どおり terminal close を維持し、再開は phase2 条件成立時のみ
次の人: PO
次の一手: live open なら各 close
空欄に残すセル: 各 close ref

E. 完成物

KEEP 13 本 / REVISE 1 本（E-11）/ DROP 0 本。

┌────────────┬────────┬─────────────────────────────────────────────────────────────────────────────┐
│     本     │  判定  │                                理由（1 行）                                 │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-1 #201   │        │ 列構成が DEC-48/65 と Fable pack F-033                                      │
│ bundle     │ KEEP   │ の列リストに完全一致し、値を一切発明していない                              │
│ 依頼       │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-2 #249   │ KEEP   │ DR-CLINICAL #249 行と同一列で、「未判定」維持の原則を明記している           │
│ range 依頼 │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-3 #211   │ KEEP   │ DEC-58/59 の clinical/OPS 分離構造どおりで、local 含む実 row                │
│ 分離依頼   │        │ 禁止を明記している                                                          │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-4 #261   │ KEEP   │ DEC-64 の値複製禁止 rule を本文に含む                                       │
│ テンプレ   │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-5 #256   │ KEEP   │ DEC-61 no-rewrite + dual sign-off 前提が §7.1 記録と一致する                │
│ close 一行 │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-6 #254   │ KEEP   │ 「local PASS 96 は閉証拠でない」を明文化しており UAT FINAL と整合する       │
│ close 条件 │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-7        │        │ merge-commit 指定・PlanetScale 所有権 blocker 確認が staging runbook        │
│ staging    │ KEEP   │ の警告（squash 祖先切断・109 オブジェクト所有権）と一致。参考値 4/1346      │
│ preflight  │        │ は本日 local ref で一致確認済み                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-8 実     │ KEEP   │ workers.dev 直行 + 実 URL の二面 health が runbook 障害初動と一致する       │
│ LINE STG   │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-9 #250   │ KEEP   │ PHI・CSV 本文・manifest の添付禁止を明記し partial 拒否が明確               │
│ 催促       │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-10 #259  │ KEEP   │ default-off 維持と非機密 enum のみの回答要求で、token 値を要求していない    │
│ 催促       │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-11 #258  │        │ 削除済み q&a.html を複製禁止先として参照している（現正本は                  │
│ U1〜U12    │ REVISE │ todo-po.md）。修正全文は下記                                                │
│ 依頼       │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-12       │        │                                                                             │
│ #89/#97    │ KEEP   │ 4 区分・no-rewrite・値非記載が #98/#99 との役割分離と一致する               │
│ rotation   │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-13       │        │                                                                             │
│ #98/#99    │ KEEP   │ §7.1 の裁定（受容 close / #253 一本化）と一致する                           │
│ close 一文 │        │                                                                             │
├────────────┼────────┼─────────────────────────────────────────────────────────────────────────────┤
│ E-14 PO-10 │        │ SQL は                                                                      │
│  presence  │ KEEP   │ NULL・空白を不在側に正しく分類し（NULLIF(BTRIM(...),'')）、値・clinic ID    │
│ 手順       │        │ を保存しない                                                                │
└────────────┴────────┴─────────────────────────────────────────────────────────────────────────────┘

E-11 修正全文（変更点: ①複製禁止先を q&a.html → todo-po.md 等の現行台帳へ修正 ②U9/U12 が開発側記入であることを 1 行明示し、先方に発明させない）:

件名: #258 DELIVERY_PACKAGE U1〜U12 一括記入依頼

▎ 推奨はA「クライアントがservice account・契約・請求・backup・support・monitoringを所有」です。B「開発者保有」は、料金・期間・責任・終了時移管の明示契約が揃うまで選びません。値の正本はDELIVERY_PACKAGE.md U1〜U12だけとし、他の台帳（todo-po.md等）へ複製しません。
▎
▎ - Ownership選択: [A / B]
▎ - U1 Cloudflare: 契約名義 [ ]、請求先 [ ]、移管有無 [ ]
▎ - U2 PlanetScale: plan [ ]、backup頻度 [ ]、保持 [ ]、契約名義 [ ]
▎ - U3 Vercel: plan [ ]、契約名義 [ ]、domain registrar権限 [ ]
▎ - U4 GitHub: organization [ ]、権限方針 [ ]、Collaborator方針 [ ]
▎ - U5 LINE: owner role [ ]、本番設定結果 [ ]、opaque ref [ ]。値は書かない
▎ - U6 Lステップ: owner role [ ]、本番設定結果 [ ]、opaque ref [ ]。API keyは書かない
▎ - U7 support: 連絡手段 [ ]、受付時間 [ ]、一次対応role [ ]
▎ - U8 monitoring: 送信先 [restricted]、Cloudflare検証結果 [ ]、opaque ref [ ]
▎ - U9 production backup: 頻度 [ ]、保持 [ ]、restore手順 [ ]、実測時間 [ ]
▎ - U10 R2: backup/versioning採否 [ ]、復旧方針 [ ]
▎ - U11 audit log: 保持年数 [ ]、廃棄方針 [ ]
▎ - U12 Production: setup結果 [ ]、URL health [ ]、rollback確認 [ ]
▎ - Contract owner role: [ ]
▎ - Client approver role: [ ]
▎ - 発効日: [ ]
▎ - opaque restricted-evidence reference: [ ]
▎
▎ U9・U12はProduction構築後（#253）に開発側が記入します。今回は空欄のままで構いません。
▎
▎ 金額、secret、実email、実identity、Go-live日付、receipt本文はこの依頼への返信やrepoへ記載しないでください。U13は#256所有であり、#258へ含めません。

F. 今日の最小セット（最大 3）

Sol の今日 3 件（#201 / staging preflight / #256 U13）を維持する。差し替えなし。 3 件とも USER が禁止操作なしで今日完了でき、残件グラフの最長パス（臨床・STG・納品）の先頭をそれぞれ解錠する。

1. PO-11 / #201 — DO_NOW
  - なぜ今か: TASK-033・#261・fail-closed cutover の共通解除条件であり、臨床責任者の応答待ちが全 residual 中の最長リードタイムだからだ。
  - 手順: §E-1 を臨床責任者へ 1 通送る。値は全列空欄のまま。
  - 完了条件: bundle 全列 + approver role + 発効日 + opaque restricted approval reference が返る。
  - やらないこと: 薬用量・単位・対象の推測。20% の医学的正本化。TASK-033 の先行着手。
2. staging ← main preflight — DO_NOW
  - なぜ今か: staging は main から 1346 commit 遅れ（local ref 実測・本日一致確認）で、実 LINE(OPS-4/5)・OPS-3・OPS-14・#252 preview がすべてこの後ろに並ぶ。migration 統合（001 単一化）を跨ぐため、checksum と PlanetScale 109 オブジェクト所有権 blocker の確定が merge 可否そのものを決める。
  - 手順: §E-7 を USER/release owner が実施する（remote refs 取得を含め USER 操作）。
  - 完了条件: E-7 完了一行（staging_preflight=PASS … merge_method=MERGE_COMMIT）が全欄埋まる。
  - やらないこと: merge・push・squash・直接 merge・STG reset・migration 適用。
3. #256 U13 disposition — DO_NOW
  - なぜ今か: visual/privacy dual sign-off は §7.1 で済んでおり、残る判断セルは U13 だけ。完了/未完の明示は今日できる。
  - 手順: U13 を完了または未完として明示し、完了時のみ §E-5 を別 USER 承認へ回す。
  - 完了条件: U13_status・発効日・opaque ref・close 承認が揃う（未完なら「open 維持」を明示して完了）。
  - やらないこと: history rewrite。未完のままの close。画像・PII・参加者名の記録。

G. agent unit

NONE。 phase2.html への 2 行追記（#284・#249-EXT）は 1 分の手作業であり、これを unit 化するのはスコープ肥大だ（F-scope の判断を維持）。USER が次の docs 更新で追記すればよい。

H. 発明しなかったもの

薬用量・warning 閾値・reference range・ワクチン適合性、契約プラン・金額・保持期間、credential・token・DSN、実氏名・email・clinic/staff/patient ID、Go-live 日付、STG/PROD の未取得証跡は一切発明していない。値セルへの答えは「現行継続を推奨」「クライアント所有（A）を推奨」「未承認のまま未判定を維持」「入力先は依頼文・値は臨床/契約責任者が埋める」の 4 形のみを使った。行 20 の 2026-11-07 は新規日付の発明ではなく、§7.1 に記録済みの inventory_start=2026-08-09（PO-09）に採択済み F-021-X の 90 日条項を適用した計算値である。本セッションではコード変更・GitHub 操作（fetch 含む）・migrate・seed 適用・secret 設定・外部送信を実行しておらず、local git の read-only 参照と repo 内ファイルの読み取りのみを行った。

I. カバレッジ

- [x] YES — todo-po の今日・対応・OPS・フォーム・Issue を個別に判定した（§D 全 83 行・欠番なし・実行行と close 行を分離）
- [x] YES — U1〜U12 を個別に判定した（§D No.29〜40）
- [x] YES — PO-008 の 6 行を個別に判定した（§D No.41 / 42 / 43 / 43b / 44 / 45 — Sol が統合していた CSV と last-visit を分割）
- [x] YES — Sol §E-1〜14 を KEEP/REVISE/DROP した（KEEP 13 / REVISE 1 / DROP 0・REVISE は全文提示）
- [x] YES — 臨床数値・契約金額・secret・実 identity・Go-live 日付を発明していない
- [x] YES — 製品 unit を無断で増やしていない（NONE 維持）

---
総括: Sol r2 は正確だ。RATIFY 79 / TIGHTEN 4 / OVERTURN 0。TIGHTEN は verdict を一切変えない — ①E-11 の削除済み q&a.html 参照の修正（送付前に必須）、②TASK-021 の 90 日再裁定期限 2026-11-07 の明示、③④#284・#249-EXT の phase2.html 未記載 drift の解消、の 4 点である。USER は §F の 3 件を今日実行し、#258 には本回答 §E の REVISE 版 E-11 を使えばよい。
