# META-LINEAR-F1-F6 — 読み取り照合と対応案

更新日: 2026-09-11（Codex Linear MCP を読み取り専用で再照会。外部書き込みなし）

2026-09-11 の Codex Linear MCP 読み取りでは、Team Baritech / Project ノア動物病院電子カルテ、BRT-4、関連語検索、BRT-45、BRT-68、BRT-226 と各 Issue のコメントを照合した。Project と BRT-4 は Backlog、BRT-45 と BRT-68 は Needs Human、BRT-226 は Review だった。F1〜F6 の直接対応 ID は特定できず、対応 ID と状態は **UNKNOWN** とする。

Linear 上の観測状態（Codex 2026-09-11）: BRT-226 は **Review**。F1〜F6 に直接対応する Issue は今回の検索範囲で特定できず、対応 ID と状態は **UNKNOWN**。不存在・Done を推定しない。

監査 ID の定義は `75aa2b64c:todo-now.md`（現在の検証結果ではない）。

## F1〜F6 対応表（定義 × Linear）

定義 SoT: `git show 75aa2b64c:todo-now.md` の着手一覧。6件は別 identity。Linear 列は直接対応が確認できるまで **未特定 / UNKNOWN**。

| 監査 ID | 定義（75aa2b64c） | 分類 | Linear 対応 | 根拠 / 状態 |
|---------|-------------------|------|-------------|-------------|
| F1 | 診察プラン更新・削除とカルテ確定の未直列化 | 製品defect指摘 | **未特定 / UNKNOWN** | 2026-09-11: 関連語検索と取得可能なコメントに直接対応なし。候補 ID を発明しない |
| F2 | 編集内容と保存versionの不一致 | 製品defect指摘 | **未特定 / UNKNOWN** | 同上。候補 ID を発明しない |
| F3 | CI E2E・負荷テストの認証fixture不足 | verification gap | **未特定 / UNKNOWN** | 2026-09-11: CI-K6 / k6 / fixture 等で直接対応なし。full clinical E2E と混同しない |
| F4 | カルテ受入項目表とpayloadの不一致 | docs drift / verification gap | **未特定 / UNKNOWN** | 2026-09-11: payload / 受入 / カルテ等で直接対応なし |
| F5 | 診察所見3欄のlabel/id未接続 | a11y defect指摘 | **未特定 / UNKNOWN** | 2026-09-11: label / id / 所見等で直接対応なし。候補 ID を発明しない |
| F6 | 検証skillに廃止packageの例示 | harness drift | **未特定 / UNKNOWN** | **監査 F6 ≠ 八王子 cutover F6**（cutover 運用は `H0-2` / `HAC-CSV-1`） |

## 照会結果（2026-09-11 Codex）

| 確認対象 | 結果 |
|----------|------|
| Linear Team Baritech / Project ノア動物病院電子カルテ / hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | Project と hub は Backlog。BRT-4 の親なし・所属 Project/Team を確認 |
| [BRT-226](https://linear.app/baritechllc/issue/BRT-226) | MCP で Review を確認。所属 Team / Project と親 BRT-4 も確認。Done は未実施 |
| Astra F1〜F6 の Linear ID | **UNKNOWN**。Team/Project 内で `Astra` / `ClinicalPlan` / `CI-K6` / `S09` / `scoped-verification` / `payload` / `label` / `id` / 関連日本語を検索し、直接対応を特定できず。検索結果の長い本文は一部省略され、横断コメント検索 API は利用できないため、全 Issue 本文・コメントの網羅照合ではない |
| [BRT-45](https://linear.app/baritechllc/issue/BRT-45) | 本文とコメントを取得。Needs Human。S09 / V04 / clinical E2E を含む #254 全体の close 条件を管理 |
| [BRT-68](https://linear.app/baritechllc/issue/BRT-68) | 本文とコメントを取得。Needs Human。現在の残は H1 実 LINE / H2 実 token LIFF |

エージェントはこの項目で Linear 書き込みと Done をしない。

## 次に必要な照会（Linear 読取可能なセッション向け）

タイトル一覧の再カウントだけでは不十分。Team Baritech / Project ノア動物病院電子カルテ / hub BRT-4 配下で **本文＋コメント全文** を対象にする。

1. `Astra` OR `F1` OR `F2` OR `F3` OR `F4` OR `F5` OR `F6`（issue + comment）
2. `ClinicalPlan` OR `clinical plan` OR `診察プラン` OR `カルテ確定`
3. `CI-K6` OR `k6` OR `SUMMARY-SCHEMA` OR `RUNTIME-CLOSEOUT` OR `auth-flows` OR `fixture`
4. `S09` OR `scoped-verification` OR `verification skill` OR `廃止 package`
5. `payload` AND (`受入` OR `acceptance` OR `カルテ`)
6. (`label` OR `id`) AND (`所見` OR `診断` OR `FE-CLINICAL-PLAN-SELECT-LABELS`)
7. 全文取得の明示 pull: BRT-4 / BRT-45 / BRT-68 / BRT-226 と (1)–(6) のヒット
8. コメント検索: `d4c870f9e` / `todo-now.md` / `astra-history` / `34025435577` / `33972458396`

## 現在の repo 受入残（Linear と 1:1 ではない）

| ledger ID | repo 状態 | Linear に書いてよいこと | 書いてはいけないこと |
|-----------|-----------|-------------------------|----------------------|
| `QA-UAT-S09-FIXTURE` | S09 ブラウザ再実行は未。S09 は BLOCKED | ブラウザ未実行。S09 は BLOCKED | S09 PASS / UAT PASS |
| `QA-UAT-V04-RETEST` | browser runtime 未実行 | V04 は UNKNOWN | V04 PASS |
| `QA-FULL-CLINICAL-E2E` | `--clinical` 未実行 | auth smoke と別 | full E2E PASS |

## 重複・食い違い

1. **F1〜F6 と現行 `todo.md` の未完了 ID は別物**。未完了は S09 / V04 / clinical E2E の受入残、USER ゲート、STG、deferred である。
2. **F3 の k6 対応と Linear 状態の対応付けは UNKNOWN**。ライブ再照会で確認する。
3. **F3 の E2E を full clinical suite と混同しない**。CI が実行するのは `e2e/auth-flows.spec.ts` のみ。
4. **F6（監査）と F6（八王子 cutover）は同名別物**。cutover 運用は `H0-2` / `HAC-CSV-1`。
5. **BRT-226** はセキュリティ修正の Review。F1〜F6 本体ではない。Done は USER。

## USER が Linear で行う手順

見つからない Issue は UNKNOWN のまま。推測で新規 Issue を量産しない（2026-09-11 以降は Linear free issue limit のため **新規作成禁止**。新規本文は [todo-issue.md](../../todo-issue.md)）。Done は USER だけが遷移する。既存チケットのコメント／状態更新のみ許可。

1. 今回の検索で特定できなかった F1〜F6 は、既存 Issue URL または本文・コメントの追加照合で対応を確定する。直接対応が確認できるまで推測で ID を割り当てない。読み取りは agent が継続可能。
2. 直接対応が確認できた Issue に限り、本ファイルと実装 SHA、固有の受入条件を照合して反映案を確定する。コメント・Done は USER の承認と受入判断後に行う。実装履歴だけでは閉じない。
3. k6 を未完了で残している Issue があれば、Git 履歴の検証証跡と照合し、別承認で閉じる。
4. 残作業と 1:1 で無い Issue は新規を増やさず、既存 ID にコメントで対応付ける。上表の ledger ID を使う。
5. [BRT-226](https://linear.app/baritechllc/issue/BRT-226) の Done は別判断。F1〜F6 とまとめて閉じない。
