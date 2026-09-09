# META-LINEAR-F1-F6 — 読み取り照合と対応案

更新日: 2026-09-09（追加照合: Grok Build / branch `docs/meta-linear-pr-20260909` / base `c0950fbdf`）

照会手段（時系列）:

1. **Codex Linear MCP（2026-09-09）**: プロジェクト55件（archive を含めて要求、次ページなし）と BRT-4 本文、関連語検索を照合。成功した読み取り結果は下記「照会結果」に保持する。
2. **Grok Build 追加照合（同日・本更新）**: Linear issue-read MCP は本セッションで **UNAVAILABLE**（接続 MCP に Linear なし。`tasks` の trigger 列挙は issue 本文読取ではない）。公開 `linear.app` は SPA/ログイン壁で本文未取得 → ライブ状態は **UNKNOWN**（Done ではない）。55件タイトル照合の単純反復はせず、定義再確認と次の本文・コメント照会リストを具体化した。

Linear 上の観測状態（Codex 2026-09-09 を正とし、Grok では再検証せず保持）: BRT-226 は **Review**。F1〜F6 に直接対応する Issue は当時の検索範囲で特定できず、対応 ID と状態は **UNKNOWN**。不存在・Done を推定しない。

repo 実装履歴の証跡 SHA: `d4c870f9e`。監査 ID の定義は `75aa2b64c:todo-now.md`（現在の検証結果ではない）

## F1〜F6 対応表（定義 × Linear）

定義 SoT: `git show 75aa2b64c:todo-now.md` の着手一覧。6件は別 identity。Linear 列は直接対応が確認できるまで **未特定 / UNKNOWN**。repo 実装済みは受入・Linear Done を意味しない。

| 監査 ID | 定義（75aa2b64c） | 分類 | Linear 対応 | 根拠 / 状態 |
|---------|-------------------|------|-------------|-------------|
| F1 | 診察プラン更新・削除とカルテ確定の未直列化 | 製品defect指摘 | **未特定 / UNKNOWN** | Codex: 55件+関連語で直接 ID なし。Grok: Linear MCP なし・本文未再取得 |
| F2 | 編集内容と保存versionの不一致 | 製品defect指摘 | **未特定 / UNKNOWN** | 同上。候補 ID を発明しない |
| F3 | CI E2E・負荷テストの認証fixture不足 | verification gap | **未特定 / UNKNOWN** | repo 側 k6 後続は閉じ済みでも Linear 未照合。full clinical E2E と混同しない |
| F4 | カルテ受入項目表とpayloadの不一致 | docs drift / verification gap | **未特定 / UNKNOWN** | 同上 |
| F5 | 診察所見3欄のlabel/id未接続 | a11y defect指摘 | **未特定 / UNKNOWN** | 診断セレクト後続 `FE-CLINICAL-PLAN-SELECT-LABELS` は別履歴 |
| F6 | 検証skillに廃止packageの例示 | harness drift | **未特定 / UNKNOWN** | **監査 F6 ≠ 八王子 cutover F6**（cutover 運用は `H0-2` / `HAC-CSV-1`） |

関連観測 Issue（F1〜F6 の代替 ID ではない）:

| Issue | 役割 | Codex 観測（2026-09-09） | 混同禁止 |
|-------|------|--------------------------|----------|
| [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | hub / 実行 SoT | Backlog。プロジェクト入口 | F1〜F6 の代替 ID にしない |
| [BRT-226](https://linear.app/baritechllc/issue/BRT-226) | セキュリティ Review | Review。Team/Project・親 BRT-4 確認。Done 未実施 | Astra 六件ではない |
| [BRT-45](https://linear.app/baritechllc/issue/BRT-45) | #254 全体受入 | Needs Human。S09/V04/clinical E2E を含む close 条件 | S09 PASS と混同しない |
| [BRT-68](https://linear.app/baritechllc/issue/BRT-68) | LINE/LIFF 残 | Needs Human。残は実 LINE / 実 token LIFF | S09 PASS に拡張しない |
| [BRT-61](https://linear.app/baritechllc/issue/BRT-61) | 旧 S09 H4 | Duplicate | 過去完了 ≠ 現在 S09 PASS |
| [BRT-105](https://linear.app/baritechllc/issue/BRT-105) | （一覧上 Done） | プロジェクト一覧で Done | F1〜F6 直接対応は未証明 |

## 照会結果（2026-09-09 Codex）

| 確認対象 | 結果 |
|----------|------|
| Linear Team Baritech / Project ノア動物病院電子カルテ / hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | プロジェクト55件の一覧と親 ID、hub 本文を取得。hub は Backlog |
| [BRT-105](https://linear.app/baritechllc/issue/BRT-105) | プロジェクト一覧で Done を確認 |
| [BRT-226](https://linear.app/baritechllc/issue/BRT-226) | MCP で Review を確認。所属 Team / Project と親 BRT-4 も確認。Done は未実施 |
| Astra F1〜F6 の Linear ID | **UNKNOWN**。プロジェクト一覧と Team 内の `Astra` / `ClinicalPlan` / `CI-K6` / `S09` / `scoped-verification` / `payload` 検索で直接対応を特定できず。検索結果の本文は一部省略されるため、全 Issue 本文・コメントの網羅照合ではない |
| [BRT-45](https://linear.app/baritechllc/issue/BRT-45) | 本文取得。Needs Human。S09 / V04 / clinical E2E を含む #254 全体の close 条件を管理 |
| [BRT-68](https://linear.app/baritechllc/issue/BRT-68) / [BRT-61](https://linear.app/baritechllc/issue/BRT-61) | 本文取得。BRT-68 は Needs Human、現在の残は H1 実 LINE / H2 実 token LIFF。旧 S09 H4 の BRT-61 は Duplicate。過去 H4 の完了を現在の S09 PASS に拡張しない |

エージェントはこの項目で Linear 書き込みと Done をしない。

### Grok 追加照合（2026-09-09・本セッション）

- Linear issue-read: **UNAVAILABLE**（セッション限定。プロジェクト全体の恒久障害としては扱わない）。
- 公開 URL 再取得: BRT-4 等は本文未取得（Loading / 認証壁）→ ライブ状態 **UNKNOWN**。
- 定義再確認: `75aa2b64c:todo-now.md` 着手一覧の F1〜F6 を6行のまま再掲（上表）。
- Codex 成功結果は無効化しない。Cursor の当時の読取不可も同様にセッション制約として保持する。

#### 次に必要な照会（Linear 読取可能なセッション向け）

タイトル一覧の再カウントだけでは不十分。Team Baritech / Project ノア動物病院電子カルテ / hub BRT-4 配下で **本文＋コメント全文** を対象にする。

1. `Astra` OR `F1` OR `F2` OR `F3` OR `F4` OR `F5` OR `F6`（issue + comment）
2. `ClinicalPlan` OR `clinical plan` OR `診察プラン` OR `カルテ確定`
3. `CI-K6` OR `k6` OR `SUMMARY-SCHEMA` OR `RUNTIME-CLOSEOUT` OR `auth-flows` OR `fixture`
4. `S09` OR `scoped-verification` OR `verification skill` OR `廃止 package`
5. `payload` AND (`受入` OR `acceptance` OR `カルテ`)
6. (`label` OR `id`) AND (`所見` OR `診断` OR `FE-CLINICAL-PLAN-SELECT-LABELS`)
7. 全文取得の明示 pull: BRT-4 / BRT-45 / BRT-68 / BRT-61 / BRT-105 / BRT-226 と (1)–(6) のヒット
8. コメント検索: `d4c870f9e` / `todo-now.md` / `astra-history` / `34025435577` / `33972458396`

### Cursor 結果の引き継ぎ

Cursor は `db7b6fa24` 起点の `AnimalEkarte-ledger-20260909` で文書2ファイルを変更し、未コミットで引き渡した。そのセッションでは Linear MCP が利用できなかった。これは Cursor の接続制約であり、上記 Codex の読み取り成功を無効にしない。

main の `227f3a6e7` にある監査 ID 訂正・独立作業の継続・S09 要約の同期を維持する（本 candidate は `c0950fbdf` を含み、`227f3a6e7` / `48e89dbe4` を祖先に持つ）。Cursor が報告した F1/F2/F3/F5/F6 のソース確認は補足調査として扱い、新たな runtime 証明にはしない。隔離 Docker の billing / clinicale2e `-short` GREEN は Cursor 報告値で、この引き継ぎでは再実行していない。testdb V04、ブラウザ S09、clinical E2E の未実行境界を維持する。

## repo 側の F1〜F6（実装履歴）

`todo.md#astra-history` と git 履歴からの対応。指摘の定義は `git show 75aa2b64c:todo-now.md` の着手一覧に従う。Linear の同名 Issue と 1:1 であることは未証明。下表は実装履歴であり、現在の受入・release 判定ではない。

| 監査 ID | repo での意味 | 実装状態 | 後続 ledger ID |
|---------|---------------|----------|----------------|
| F1 | 診察プラン更新・削除とカルテ確定の未直列化 | 実装・統合済み（履歴入口は `todo.md#astra-history`） | なし（履歴ポインタのみ） |
| F2 | 編集内容と保存 version の不一致 | 実装・統合済み | なし |
| F3 | CI E2E・負荷テストの認証 fixture 不足 | auth smoke 実装済み。manual run `33972458396`。後続の k6 は `CI-K6-SUMMARY-SCHEMA` / `CI-K6-RUNTIME-CLOSEOUT` で閉じた | 閉じ済み。full clinical E2E は別 ID |
| F4 | カルテ受入項目表と payload の不一致 | 実装・統合済み | なし |
| F5 | 診察所見3欄の label/id 未接続 | 実装・統合済み | なし（診断セレクトの後続は `FE-CLINICAL-PLAN-SELECT-LABELS` で閉じた） |
| F6 | 検証 skill に廃止 package の例示 | 実装・統合済み | なし。八王子 cutover の F6 とは別 ID |

閉じた後続（履歴は git）: `CI-BE-DBORTX-INVENTORY`、`CI-K6-SUMMARY-SCHEMA`、`CI-K6-RUNTIME-CLOSEOUT`、`FE-CLINICAL-PLAN-SELECT-LABELS`。

## 2026-09-06 以降の repo 残（Linear と 1:1 ではない）

| ledger ID | repo 状態 | Linear に書いてよいこと | 書いてはいけないこと |
|-----------|-----------|-------------------------|----------------------|
| `QA-UAT-S09-FIXTURE` | HTTP/CLI/cleanup 実装済み。S09 ブラウザ再実行は未。S09 は BLOCKED | helper + HTTP/CLI 実装。S09 は BLOCKED | S09 PASS / UAT PASS |
| `QA-UAT-V04-RETEST` | testdb DELETE GREEN。live HTTP 403 | 500 回帰は testdb 非再現。V04 は UNKNOWN | V04 PASS |
| `QA-FULL-CLINICAL-E2E` | fixture + allowlist 置換済み | `--clinical` 未実行。auth smoke と別 | full E2E PASS |
| `QA-UAT-EVIDENCE-SYNC` | `UAT-DOMAIN-STATUS.md` が集計正本 | 開いている製品 FAIL は 0 | 未再実行 scenario を PASS |

## 重複・食い違い

1. **F1〜F6 と現行 `todo.md` の未完了 ID は別物**。未完了は S09 / V04 / clinical E2E の受入残、USER ゲート、STG、deferred である。
2. **F3 の k6 失敗を Linear 上で未完了のまま残している可能性**がある。repo では validator + run `34025435577` まで閉じている。
3. **F3 の E2E を full clinical suite と混同しない**。CI が実行するのは `e2e/auth-flows.spec.ts` のみ。
4. **F6（監査）と F6（八王子 cutover）は同名別物**。cutover 運用は `H0-2` / `HAC-CSV-1`。
5. **BRT-226** はセキュリティ修正の Review。F1〜F6 本体ではない。Done は USER。

## USER が Linear で行う手順

見つからない Issue は UNKNOWN のまま。推測で新規 Issue を量産しない。Done は USER だけが遷移する。

1. 今回の検索で特定できなかった F1〜F6 は、既存 Issue URL または本文・コメントの追加照合で対応を確定する。直接対応が確認できるまで推測で ID を割り当てない。読み取りは agent が継続可能。
2. 直接対応が確認できた Issue に限り、本ファイルと実装 SHA、固有の受入条件を照合して反映案を確定する。コメント・Done は USER の承認と受入判断後に行う。実装履歴だけでは閉じない。
3. k6 を未完了で残している Issue があれば、run `34025435577` と validator コミットへリンクし、別承認で閉じる。
4. 残作業と 1:1 で無い Issue は新規を増やさず、既存 ID にコメントで対応付ける。上表の ledger ID を使う。
5. [BRT-226](https://linear.app/baritechllc/issue/BRT-226) の Done は別判断。F1〜F6 とまとめて閉じない。
