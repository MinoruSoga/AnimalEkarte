# TODO-V-LINEAR — 既存 Linear Issue の読取照合票

照会時刻: 2026-09-18 23:22–23:23 JST。使用操作: Linear MCP `get_issue`、`list_issues`、`list_comments`（読取専用）。Issue 作成・更新・コメント・状態変更は未実施。

## 判定の境界

照合元は original main worktree `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte` の 2026-09-18 未コミット版 `todo-verification.md:162-191`、`todo.md:41-46`、`todo-issue.md:1-5`。`todo-verification.md:177-179` は Team Baritech / Project ノア動物病院電子カルテ / BRT-4 配下の既存 Issue 本文とコメントを読み、ローカル ID との直接一致を確認するよう求める。`todo.md:44` と `todo-issue.md:5` は新規 Issue を作らず、外部投稿・状態変更を明示承認後に限る。候補 worktree の root TODO は 2026-09-17 版であり、その相対リンクを 2026-09-18 版の根拠には使わない。

main の TODO にある 2026-09-18 `USER_NOT_LOGGED_IN` は、その時点の履歴である。今回の受領セッションでは上記 MCP 読取が成功した。ただし今回読めた Issue の状態だけが今回の観測値であり、未読 Issue やローカル ID の状態は UNKNOWN。[2026-09-11 の F1–F6 照合](../linear-f1-f6-mapping.md)も履歴であり、F1–F6 の直接対応を確定しない。

## 今回読めた既存 Issue

| Issue / URL | 2026-09-18 23:22–23:23 JST の状態 | 本文・コメントで確認した範囲 | ローカル ID 対応 |
|---|---|---|---|
| [BRT-4](https://linear.app/baritechllc/issue/BRT-4/case-ノア動物病院電子カルテ開発-案件ハブ) | Backlog、`archivedAt=2026-09-15T22:17:25.660Z` | 案件ハブ。本文に BRT-41/42/45/68 へのリンク。 | 親子関係のみ。直接対応なし |
| [BRT-41](https://linear.app/baritechllc/issue/BRT-41/調査-gh249-検査rangeワクチン適合の出典付き承認回収) | Needs Human | BRT-4 の子。コメントは検査 range・ワクチン適合の出典付き承認待ち。 | 今回のローカル ID との直接対応なし |
| [BRT-42](https://linear.app/baritechllc/issue/BRT-42/調査-gh250-formal-producer-bundle-受領cutover前提確認) | Needs Human | BRT-4 の子。コメントは formal producer bundle 受領待ち。 | 今回のローカル ID との直接対応なし |
| [BRT-45](https://linear.app/baritechllc/issue/BRT-45/調査-gh254-納品uat-close条件実施は-brt-68-local-passで閉じない) | Needs Human | BRT-4 の子。本文に S09、V04、clinical E2E、LSTEP、idToken。コメントは #254 close ゲートで、実施本体は BRT-68 と区別。 | 類似主題のみ。full local ID の直接対応なし |
| [BRT-68](https://linear.app/baritechllc/issue/BRT-68/調査-uat人間レーン統合-p1h1h7254-前提) | Needs Human | BRT-4 の子。コメントは #254 close 判定を BRT-45 と区別し、BRT-58–65 の吸収を記載。 | 類似主題のみ。full local ID の直接対応なし |

`list_issues(parentId=BRT-4, includeArchived=true)` で上記子 Issue の Team / Project / parent を確認した。BRT-41/42/45/68 の `relations.duplicateOf` は null。BRT-58–65 は Duplicate と表示され、BRT-68 コメントにも吸収とあるため、細分項目への重複投稿はしない。BRT-45 と BRT-68 は close 判定と実施本体で役割が異なる。親子関係・同じ語・検索ヒットはローカル ID の直接一致根拠にしない。

## ローカル ID と更新下書き

| ローカル ID | 既存 Issue URL | 読取日時 / 状態 | 直接一致根拠 | 未解決の差分 | exact 本文・コメント更新案 |
|---|---|---|---|---|---|
| `TODO-V-LINEAR / META-LINEAR-APPLY` | UNKNOWN | 対応 Issue は UNKNOWN | BRT-4/41/42/45/68 の本文・コメントに full ID なし | 9月13–17日の残件との対応照合 | 直接対応未確定のためなし |
| `PERF-V-LINEAR` / `PERF-STG-LOGIN` | UNKNOWN | 対応 Issue は UNKNOWN | 対象5 Issue に full ID または性能残件としての直接対応なし | `/login`、`/me`、OPTIONS、Container の対象 Issue 本文・コメント照合 | 直接対応未確定のためなし |
| `AUTH-V-LINEAR-READ` / `AUTH-V-LINEAR-WRITE` | UNKNOWN | 対応 Issue は UNKNOWN | 対象5 Issue に full ID なし | D1 対象環境・既存スタッフ・通常 login・メールの受入条件と Issue の照合 | 直接対応未確定のためなし |
| `QA-UAT-S09-FIXTURE` / `QA-UAT-V04-RETEST` / `QA-FULL-CLINICAL-E2E` | UNKNOWN | 対応 Issue は UNKNOWN | BRT-45 に S09/V04/clinical E2E の語はあるが full ID なし | 同一作業・受入条件かの確認 | 直接対応未確定のためなし |
| `QA-UAT-LSTEP-REAL` / `QA-UAT-LINE-IDTOKEN` | UNKNOWN | 対応 Issue は UNKNOWN | BRT-45/68 に LSTEP/idToken/実 token LIFF の語はあるが full ID なし | E1/E2 の実施対象と close 条件の照合 | 直接対応未確定のためなし |
| `UAT-R2-MASTER-PATH` / `UAT-R2-EXCLUSIVE-LOCK` / `UAT-R2-CHART-FIT` | UNKNOWN | 対応 Issue は UNKNOWN | 対象5 Issue に full ID なし | 医院・PO 入力と受入条件 | 直接対応未確定のためなし |
| `UAT-Q3-GENDER-MAP` / `UAT-Q2-VACCINE-SPECIES` / `UAT-Q4-UNPAID-TRIAGE` / `UAT-Q2-TREATMENTS-IMPORT` | UNKNOWN | 対応 Issue は UNKNOWN | 対象5 Issue に full ID なし | 元報告と既存 Issue の直接照合 | 直接対応未確定のためなし |
| `PO-PET-DECEASED-DATA-BACKFILL` / `TASK-444-ADDENDUM-CODEGEN` | UNKNOWN | 対応 Issue は UNKNOWN | 対象5 Issue に full ID なし | PO 判断・既存 Issue 本文の直接照合 | 直接対応未確定のためなし |

今回の読取範囲では、直接対応が確定した行は **0件**。従って投稿可能な exact 更新本文・コメントは **0件**。`list_issues(query=TODO-V-LINEAR)` の関連度検索ヒットも literal 一致として扱わない。未確認の既存 Issue に対応がないとは断定しない。

次の読取作業は、Team / Project / BRT-4 配下の残る既存 Issue の本文・コメントを、上表の full ID・元報告・実装範囲・受入条件で照合すること。直接一致する Issue URL を特定した行だけ、現状との差分と exact 更新文面をここに追記する。反映には対象 URL と文面の明示承認、更新直前の再読取、更新後の再読取が必要。現在の Issue 状態・実装・受入・STG/本番の完了を、この照合票から推定しない。
