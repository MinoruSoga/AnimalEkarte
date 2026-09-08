# META-LINEAR-F1-F6 — 読み取り照合と対応案

更新日: 2026-09-09

照会手段: Linear MCP（BRT-226 の読み取り成功）

Linear 上の現行状態: BRT-226 は **Review**。F1〜F6 の対応 ID と状態は **UNKNOWN**（未照合。repo 記録から Done を推定しない）

repo 実装履歴の証跡 SHA: `d4c870f9e`。監査 ID の定義は `75aa2b64c:todo-now.md`（現在の検証結果ではない）

## 照会結果（2026-09-09、BRT-226 のみ再確認）

| 確認対象 | 結果 |
|----------|------|
| Linear Team Baritech / Project ノア動物病院電子カルテ / hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) | URL は repo に存在する。本文・子 Issue・状態は未取得 |
| [BRT-105](https://linear.app/baritechllc/issue/BRT-105) | repo 表記は Done。Linear 再確認は未実施 |
| [BRT-226](https://linear.app/baritechllc/issue/BRT-226) | MCP で Review を確認。所属 Team / Project と親 BRT-4 も確認。Done は未実施 |
| Astra F1〜F6 の Linear ID | **UNKNOWN**。タイトル検索・親子リンク照合は未実施 |

エージェントはこの項目で Linear 書き込みと Done をしない。

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

1. [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下とタイトル検索で `F1`〜`F6` / `Astra` / `CI-K6` / `ClinicalPlan` / `S09` / `clinical E2E` / `V04` を列挙する。
2. 実装済みで後続なしの Issue は、本ファイルと SHA `d4c870f9e`（未 push なら push 後の SHA）をコメントしてから Done にする。
3. k6 を未完了で残している Issue があれば、run `34025435577` と validator コミットへリンクし、別承認で閉じる。
4. 残作業と 1:1 で無い Issue は新規を増やさず、既存 ID にコメントで対応付ける。上表の ledger ID を使う。
5. [BRT-226](https://linear.app/baritechllc/issue/BRT-226) の Done は別判断。F1〜F6 とまとめて閉じない。
