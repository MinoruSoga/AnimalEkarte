# Plane チケット起票規約

> **目的**: Plane への起票フォーマットと運用ルールの正本。人間・エージェント・外部流入（スプシ/Slack/チャットボット）のどの経路でも同じ構造で起票し、下流の調査・実装セッションが本文だけで動けるようにする。
> **読者**: このリポジトリで作業する全員・全エージェント。
> **タイミング**: Plane work item の作成・更新、着手プランの受領、session割当、前提merge後の再判定、完了報告時。

## 基本ルール

- **実行 SoT は Plane**。workspace `baritechllc`、プロジェクト `EMR`。タスクの現在状態は Plane を見る。`todo.md` は出典・仕様履歴の記録であり、未完了タスク本文の置き場ではない（`bug.md` は空。製品 FAIL 台帳も Plane）。
- **エージェントの新規起票は承認不要**（2026-10-06 決定。Linear 時代の「新規作成禁止」ルールは Linear の件数制限対策であり、Plane では適用しない）。起票したら作成した `EMR-N` を報告に列挙する。
- Plane への**既存チケットのコメント・状態変更**も同じく可。ただし Done / Cancelled / Duplicate への遷移は受入証拠か人間裁定がある場合のみ。
- 承認不要の範囲は work item の**作成・本文/ラベル/状態更新・コメント**に限る。一括変更・プロジェクト設定や権限の変更・チケット削除などの広範/不可逆操作は承認対象（AGENTS.md の外部投稿・不可逆操作の停止境界に準ずる）。
- 起票ツールは Plane MCP の `workitem`（`create` / `update`）。ラベル・状態は UUID 指定（附録の対応表を使う）。本文は `description_stripped`（プレーンテキスト）に書けば HTML に包まれる。HTML 制御が必要な場合のみ `description_html`。

## 必須フィールド

| フィールド | ルール |
|:---|:---|
| `name`（タイトル） | `出典アンカー + 種別マーカー + 要旨`。下表参照。要旨だけの単票も可だが、出典があるものはアンカーを落とさない |
| `labels` | **`source:*` を必ず1つ** + `lane:*` を1つ + 該当するものは任意個付ける（下表「ラベル」） |
| `priority` | urgent / high / medium / low / none。下表「優先度の目安」 |
| `state` | 下表「作成時 state」 |
| `description` | 種別テンプレートに従う。「調査が必要」と書いて終わらせない（起票前に調べる） |

### タイトル規約

```
[出典アンカー][種別マーカー] 要旨
```

- **出典アンカー**（出典がある場合、先頭）: `【バックログNo.NN】` / `スプシNo.NN｜` / `GitHub #NNN` / `Slack(YYYY-MM-DD)` / `UAT-XXX:`
- **種別マーカー**: バグは `BUG:`、基盤・横断テーマは `【基盤】`、調査・スパイクは `[リサーチ]`。通常の機能・改善はマーカーなし
- 例: `【バックログNo.65】初診カルテ未作成のまま会計完了できないようにする`、`BUG: トリミングの画像アップロード欄が保存されない`、`[リサーチ] Upstash 導入要否の評価`
- 旧来の `BUG:` と `スプシNo.NN｜[BUG]` が混在している。新規は出典アンカー→種別マーカーの順で統一する

### ラベル

| 系統 | 付け方 |
|:---|:---|
| `source:*`（**必須1つ**） | 手動・エージェント起票 = `source:manual`、Slack転記 = `source:slack`、スプシ = `source:client-sheet`、社内docs移行 = `source:internal-docs`、チャットボット = `source:chatbot`（明示的な reviewable export のみ。chat 本文の自動起票は禁止 — SEC-N3 裁定）。※bug-report は別経路: Plane 連携設定時は保存成功後の自動起票と手動再送を許可（2026-10-11 owner 指示、EMR-340。chat 本文は含まない）。表にない出典（UAT での発見・GitHub Issue 起点・エージェントが調査/実装中に発見した事案）は `source:manual` とし、出典はタイトルアンカーと本文の【出典】節で保持する |
| `lane:*`（1つ） | `lane:dev`（実装）、`lane:ops`（運用）、`lane:perf`（性能）、`lane:finance`、`lane:strategy`。複数に跨る場合は**作業の主対象**で1つ選ぶ（性能改善の実装 → `lane:perf`、会計・料金系の実装 → `lane:finance`、いずれでもない機能実装 → `lane:dev`）。ドキュメント・CI・開発ツールの整備は `lane:dev`、外部サービス・運用手順は `lane:ops` |
| `agent:*` / `role:*` | 担当レーンを示すとき。`role:human` = 人間作業、`agent:dev-engineer` = 実装委譲候補、など |
| `bug` | 製品欠陥（UAT・実運用で確認済みの FAIL）。要望・改善には付けない |
| `needs-human` | 人間の判断・操作を含む票。state=Needs Human と併用可（label は検索性、state は workflow 位置） |
| `Phase2` | 今期対象外。再開条件を本文に書く |
| `risk:high` | 臨床安全・データ・権限に踏む変更 |
| `kind:case-hub` / `case-category:*` | 案件ハブ専用。通常の作業票には付けない |

### 優先度の目安

| priority | 目安 |
|:---|:---|
| urgent | 臨床安全・データ破損・本番障害・医院業務の停止 |
| high | 主要機能の不具合、本番切替ブロッカー |
| medium | 通常のバグ・機能要望 |
| low | nice-to-have、改善 |
| none | ハブ・親票・優先度未判定の intake |

### 作成時 state

| 状況 | state |
|:---|:---|
| 既定（トリアージ前） | Backlog |
| 対象scopeの受入条件・計画/採択契約・必要subset受領/共有ownerが[着手契約](#scope-and-readiness-contract)を満たす（環境準備/実行許可/実試験の完了とは別） | Ready |
| 調査・スパイク票 | Research |
| 人間の判断・操作待ちで滞留 | Needs Human |
| 今期対象外 | Phase2（ラベルではなく state としても存在する。運用は state を優先） |
| 本番環境構築の開始（2026-10-20）までに完了必須 | 本番切替対応（state 名は `scripts/tools/plane-merge-state.sh` / `plane-merge-state.test.sh` が文字列参照するため変更しない） |

複数に同時に該当する場合は滞留・保留側を優先する: `Phase2` > `Needs Human` > `本番切替対応` > `Ready` / `Research` > `Backlog`（切替期限の情報は priority や本文に併記する。state は workflow 上の実際の位置を表す）

<a id="scope-and-readiness-contract"></a>
## 粒度・着手前調査・1PRの完了契約（2026-10-23）

規約改善の要求者/判断窓口はMinoru Soga。目的は、着手後に前提の不存在・未決判断・共有編集競合が判明して限定helperだけで止まる事態と、同じ仕様の再質問/再調査を減らすこと。票数や文書量を増やすことは目的にしない。計画上の確認指標は、各実装scopeに固定SHA・必要subsetの状態・owner・一意の完了条件・後続の対応があり、未説明の差分がないこと。実際の手戻り/リードタイム改善は今後測定するもので、本改訂だけで効果測定済み/全票整備済みとはしない。[業務効率の意思決定原則](../product-philosophy.md)を継承する。

### 親票と実装成果の分離

- 「着手プランがある」と「そのscopeを追加判断なく1PRで閉じられる」は別判定。親/ハブは全体責務と網羅対応、実装票は境界付き成果を所有する。親番号の一部PRが採用されても親全体の完成にしない。
- 「3票ずつ」等の票数や所要時間だけでsessionを束ねない。依存関係・共有ファイル・原子的な成功単位から1session/1PRの範囲を決め、必要なら既存親の子票として分割する。分割後も全AC/残件の受領先を対応付け、同じ実装を二重起票しない。
- 実callerを含む機能PRと、明示した契約/基盤成果PRを区別する。DB/API契約/library/mock FEの先行PRは正当な限定成果にできる。ただし成果の型と非対象、後続owner/票、受領条件、activation gateを本文/PRへ明記する。先行成果の採用を機能公開や製品全体完成に換算しない。

<a id="adopted-baseline-and-dependency-subsets"></a>
### 採用基準と必要subsetの実査

起票/割当前と着手直前に、対象範囲の最新採用main SHA、関連PRのbase/head/採用merge SHAとopen/merged状態、source symbolを実際に読む。schema/model/API定義に加え、必要route・composition/DI・reader/writer・adapter・callerの存在と接続を固定リンクで記録する。設計文書の名前をコードの存在証拠にせず、PR候補headや作者PASSをmain全体の実装/受入済みと扱わない。

必要subsetごとに次の表を埋める。全親票Doneを一律に待たず、実装に必要な供給物と受領順を限定する。

| 状態 | 記録する証拠/次の扱い |
|---|---|
| 存在済み | 採用main/merge SHAとfile/symbol/wire、今回callerが接続できる範囲。単なる宣言の存在と実接続を区別 |
| 今回同PRに含む | 作成/変更するsubset、編集owner・file/hunk、順序、同PRの完了条件 |
| 別owner待ち | 供給owner/票/PR、必要file/symbol/契約版、固定受領SHA、未受領部分とその停止範囲。独自stub/二重型で先回りしない |
| 未選定 | 影響する業務/技術論点、質問/判断窓口、回答先、停止するscope。決まっていることと混ぜない |

大型schema統合・前提PRのmerge・共有interface更新後は、変更差分と全必要callerへの影響を再読し、表と1PR完了条件を再判定する。古いReadyや「以前見た」を再読の代用にしない。必要subset未受領でも独立に閉じる限定成果は進められるが、未接続を統合済みとはしない。

<a id="decision-and-shared-edit-boundaries"></a>
### 未決判断・共有編集・安全性の閉包

1. 正本・決定履歴・既採用成果を先に調べ、既決を実装担当/本人へ質問し直さない。本当の矛盾/未決は影響scopeだけ停止し、問い・判断owner・回答先を固定する。着手前に必要な保存/認可/競合の設計判断を「実装受入時に確認」で先送りしない。独立に確定している範囲は続行できる。
2. 共有編集ownerと変更file/hunk、[claim/worktree規約](../../AGENTS.md#packet-claim-protocol-mandatory)、API YAML・composition root・migration番号・生成入力等の統合順を明記する。migration番号は最新mainで調整し、他sessionのclaimを無視して同じ編集scopeを追加割当しない。担当名/claim未決は明記し、既存ownerへ無断でscopeを足さない。
3. DB/concurrencyを伴うscopeは、全writer/caller、lock mode/順序と複数頭等の集合順、CAS/fence、同TXのwrite/receipt/監査、旧経路の切替を調査する。作成/変更/削除/復元/batch/import/内部caller等から関係する入口を列挙し、必要なroute/DI/reader/adapterまで接続を閉じる。helperだけを機能完了にしない。
4. 今回対象外のwriterが安全性に影響するなら「対象外」と記すだけで未確認を消さない。既存参加の証拠、必要な追補owner/受領条件、未受領時のactivation停止を記す。参加閉包が未成立なのにruntimeを有効化しない。親全体/全domainのDone待ちへ広げず、必要な安全subsetに限定する。
5. チケット本文や依存受領は実行権限の追加ではない。DB適用・正規生成・外部操作等の実行者/承認範囲は既存規約と本人指示を維持する。必要な実装scopeと、実行許可・機能activationを分ける。

<a id="implementation-start-plan-fields"></a>
### 着手プランの必須項目

バグ/機能の実装票には以下を本文へ記すか、固定版の計画へリンクして今回scopeを特定する。テンプレートの長さは目安であり、必須条件を行数のために削らない。親/調査/SOPはそれぞれの成果に適用し、未選定を実装可能と見せない。

| 項目 | 必須内容 |
|---|---|
| 問題/目的 | 現状の差分、業務目的、判断窓口。不要な入力/二重管理/再確認を増やさない |
| 基準SHA/現状証拠 | 最新採用main、関連PRの採用状態、確認したfile/symbol/実接続。未確認の範囲 |
| in/out | このsession/1PRの成果型・対象ACと非対象。非対象の受領先 |
| 必要subset | 上表の4状態、供給/受領owner、必要interfaceと採用SHA |
| 採択契約/異常条件 | 正本/決定履歴、入力→出力/保存/副作用の例、権限/空/失敗/競合/unknown等の対象条件。候補は未選定と明記 |
| 変更file/owner | 専有/共有file・hunk、claimと競合確認、共有単一owner/統合順 |
| 実装順 | 最小の供給→受領→caller接続順と旧経路切替。無関係の全親Done待ちにしない |
| 1PR完了条件 | 対象AC→実ファイル/接続/fixtureと限定成果の証拠。機能PRは必要callerまで、基盤PRは明示interfaceと後続引渡しまで |
| 後続/残件 | 実装・未接続・依存待ち・未選定・受入/activationの残りとowner/票 |
| 停止/質問条件 | 何が起きたらどのsubsetを止め、誰へ何を質問し、どこへ回答を残すか |
| 検証status | fixture設計/作成と実行の別、対象コマンド/環境/結果又はNOT_RUN・理由、base/final head/採用merge SHA。独立レビュー対象head |

<a id="verification-and-readiness-status"></a>
### 検証statusと状態運用

- NOT_RUN（未実行）、未実装、未接続、必要生成成果の未生成/未受領（コード未完成/前提未受領）、依存待ち、仕様未決、対象外（理由）を別に記録する。fixtureの存在は実行PASSではなく、PR作成≠main採用≠runtime公開≠受入PASS。対象ACごとの証拠と残件を残す。
- 本人が試験を延期/禁止したscopeに、実行をPR作成の新しい必須gateとして足さない。fixtureを定義/作成する仕事と実行を分け、許可された静的確認とNOT_RUNの限界を明示してPRにできる。ユーザー専有の生成/DB適用をagentへ暗黙委任しない。
- 既存対象外のcompile失敗だけで変更scopeを無限に広げない。固定baseでも生じる根拠/既存ownerと、今回変更の未解決を分ける。実行禁止なら「実行観測なし」とし、推測を既存失敗にも今回PASSにも使わない。
- Readyは対象scopeの計画/契約と実行前提の確認。必要判断が未決、依存を含んだscopeが未受領、共有ownerが未確定ならその停止範囲を分離して状態を判断する。明示した独立の契約/基盤成果だけがReadyでも、後続の機能全体をReadyにしない。
- Running/Review/UAT/Doneとマージ連動は[既存の本人運用](#claim-and-work-start)を維持する。一部分のmergeや自動Review遷移は親全ACの完成/人間受入ではない。新テンプレ未記載だけを理由に旧票を一律降格せず、変更が必要なら個別根拠を残し、既存の状態遷移権限に従う。

<a id="scope-splitting-examples"></a>
### 改訂前後の例（新しい業務/実装割当ではない）

| 改訂前 | 改訂後の着手契約 |
|---|---|
| 「医院時間GET/共通保存/LINE/policyを全部この票で実装」 | 親に全体AC対応を保持し、共有保存/API契約の明示先行成果→共通時間のread/write成功単位→LINE受付設定の旧writer切替→担当者policyの保存/lookupという依存順を記載。各PRの必要model/registry/C06等の供給SHA・owner・未受領を固定する。GET helperだけで全部完了にせず、実際の分割は共有file/原子性を確認して決める（固定4PRという規則ではない） |
| 「共通死亡readerを追加すれば完了」 | 機能PRなら実caller/DIと対象操作の既存認可/関連writerのlock参加まで照合。明示されたlibrary先行成果ならinterface/fixtureと後続caller票・activation gateを引き渡す。死亡判定やlock参加契約が未選定なら当該subsetを保留し、担当者に新業務判断を委任しない |

<a id="existing-ticket-scope-reassessment"></a>
### 既存票へ適用する手順

1. 最新採用main SHAを固定し、Ready/Running/Review/UATを優先する点検対象一覧と観測範囲を作る。状態名やコメント件数を実装数/全AC数へ換算しない。
2. 既存PR/採用成果/child/ownerを照合し、重複・既実装を除外する。Done等の限定成果は再実装せず、現在の受領先と残件を追う。
3. 成功単位・共有file・依存/安全性の閉包からscopeを再分割する。未決/未受領だけを隔離し、可能な独立成果を止めない。
4. 既存の正しい票へ実証差分を追補する。新子票が必要なら重複確認を先に行い、親AC/依存と全子票の網羅対応を残す。数合わせの起票/担当再割当/状態の一律変更をしない。
5. 依存する両票の本文・最新決定コメント・両側の依存記述・Plane relation（blocked_by等）が、同じ必要subsetと受領条件を指すことを確認する。必要なら既存の限定子票へ依存を対応付け、本文だけ更新して親票全体による機械的blocked表示を放置しない。関係や状態を根拠なしに一括削除しない。
6. 個別票に基準SHA・根拠・変更scope・判断/質問/供給先と検証statusを記録し、状態変更は既存権限に従う。他sessionのclaim/進行を尊重する。
7. 点検済み/未点検/未確認を報告する。全票を見ていないのに「全部整備済み」としない。本規約PRはルール整備だけで、既存票の再調査/修正/実装/試験を済ませた記録ではない。

## 本文テンプレート

プレーンテキストで `【節名】` マーカーを使う（既存票と同じ書式）。1票あたり15〜30行を目安にし、長い根拠はリポジトリ内ファイルへのパス参照にする。種別が複数に見える場合は**実行者優先**で選ぶ — `role:human` の作業（人間が手順で実行するバグ修復・データパッチ等）は SOP 型を使い、`bug` ラベルは性質として併用する。

### 設計思想（採用根拠）

チケットは「**境界付きの設計契約**」として書く — 実行者が既知の前提・採用済み契約を再発見せず、未決の設計判断を追加せず着手でき、成功単位・必要依存・完了条件が第三者にも検証可能であること（[着手前の採用SHA再確認](#scope-and-readiness-contract)は省略しない）を合格基準とする。以下の既知のプラクティスを採用根拠とする:

| 根拠 | 採用する考え方 |
|:---|:---|
| [Linear Method](https://linear.app/method/write-issues-not-user-stories) | タイトルは作業をそのまま平文で述べる。ユーザーの声は要約せず原文引用。issue = 明確な成果を持つタスク |
| [Mozilla bug-writing](https://bugzilla.mozilla.org/page.cgi?id=bug-writing.html) | 再現手順が最重要。「常に再現 / 時々再現 / 未再現」を明記する |
| [How to Report Bugs Effectively (Tatham)](https://www.chiark.greenend.org.uk/~sgtatham/bugs.html) | **事実と推測を分離する** — 実測は実測、仮説は仮説と書き分ける |
| [What a Good Engineering Ticket Looks Like (nat.io)](https://nat.io/blog/what-a-good-engineering-ticket-looks-like) | Problem / Impact / Requirements / Non-goals / Test cases の5要素を備える。**solution-first 起票（問題を定義せず解決策から書く）はアンチパターン** |
| SOP（標準作業手順書）構造 | 人間作業票は 目的→前提条件→番号付き手順（1手順1動作+期待結果）→検証→例外/ロールバック |
| INVEST | 粒度は Independent / Small / Testable（`task-create` の「AIが実装可能」基準と同一） |

### 3問テスト（起票の合格基準）

どの種別の票も、本文を読んだ時点で次の3問に答えられなければ起票未完了（調査不足 → `task-create` Phase 2 のリポジトリ調査に戻るか、調査票として起票し直す）:

1. **現状** — 今システム/プロセスはどうなっているか。証拠（`path:line`・実測値・ログ・再現条件）はあるか
2. **あるべき姿** — 完了後どう振る舞うべきか。第三者が検証できる形か（受入条件・検証手順）
3. **取り組み方** — どう進めるか。実装方針・参照実装・手順の順序が具体的に書かれているか

加えて: [着手契約](#scope-and-readiness-contract)に従い、スコープ外（Non-goals）・必要subset・共有編集ownerと基準SHAが明示され、事実と推測が分離されていること。「〜が必要」「〜を検討する」だけで終わる票は禁止。

### 人間への質問・決定待ち（あれば本文の先頭に配置）

人間の判断・回答・操作を必要とする票は、【人間への質問】節を**本文の先頭**（他の全節より前）に置く。一覧から開いた人間が最初に読む位置に置き、**本文を読まなくても回答できる自足形式**にする。質問がなければ節自体を省略する。

```text
【人間への質問】
- Q1. <問いを1文で>（背景: <1行の最小文脈>）
  - 案A: <内容> — <影響>
  - 案B: <内容> — <影響>
  - 推奨: 案A（<理由>）
  - 未回答時: <ブロックされる範囲> / 暫定: <デフォルト動作>
- Q2. <問い>（背景: ...）
  - 自由記述（選択肢を列挙できない場合のみ。何を決めればよいかは明示する）
```

- **1問1論点・番号付き**: コメントに「Q1: A」「Q2: 要相談」と番号だけで回答できる形にする
- **選択式を優先**: 案A/案Bと各案の影響、推奨案と理由を必ず添える（「どうしましょう？」だけの問いは禁止）
- **未回答時の影響を明記**: 何がブロックされるか、暫定デフォルトがあればそれも書く
- **回答後は消さない**: 回答を節内に追記して残す（決定経緯の証跡）
- 未解決の質問を含む票は `needs-human` ラベルを付け、回答待ちで完全に滞留するなら state=Needs Human にする
- 原則として質問は起票前の仕様確認（`task-create` Phase 2.5）で潰す。チケットに残すのは起票時点でも残った論点のみ
- 旧票の `【人間決定待ち】` 節（EMR-298 等）はこの `【人間への質問】` に統合したものと読み替える

### バグ票（label `bug`）

```text
【人間への質問】（あれば本文先頭。形式は上記「人間への質問・決定待ち」節を参照）
【出典】（例: スプシNo.61 / UAT-XXX / 社内報告）
【調査基準】採用main SHA/関連PR採用状態/確認file・symbol、実測と静的読解の区別
【前提・依存】必要subset4状態/供給・共有owner/受領SHA/未選定と停止範囲
【現象】観測されている挙動（1-2行）
【再現手順】再現頻度（常に/時々/未再現）→ 最短の UI 操作 or API 呼び出し。再現に必要なデータ条件も
【期待挙動】
【実測挙動・根拠】path:line 付きの現状コード or レスポンス実測（起票前にコード確認済みであること）
【影響範囲】clinic_id / owner / pet / staff 分離への影響、隣接画面・LIFF/印刷への波及
【スコープ外】この修正で扱わない範囲（抜本的再設計・別要因の類似不具合は別票へ）
【修正方針】採用案（複数案あるなら案A/案Bと採用理由）。原因が未特定なら「調査票 EMR-N に分離」等と書く
【1PR完了条件・後続】成果型/対象AC→接続/fixture/共有owner/残件owner・activation gate（必須項目を含む固定版計画へのリンクでも可）
【受入条件・検証status】「〜の画面で〜を入力し〜が表示される」形式。エッジケース（空データ・エラー・上限値）、対象AC→接続/fixtureの対応、scopedコマンドと結果又はNOT_RUN・理由（実行許可は別）
```

事実（実測値・ログ・`path:line`）と推測（原因の仮説）は分けて書く。仮説は【修正方針】の候補案として記し、未検証の断定はしない。

### 機能・タスク票

```text
【人間への質問】（あれば本文先頭。形式は上記「人間への質問・決定待ち」節を参照）
【背景・目的】業務上の目的（依頼原文の要旨。ユーザーの声は要約せず原文引用。責任者が分かるなら明記）
【調査基準】採用main SHA/関連PRのhead・採用状態、確認file/symbol
【現状】今の実装/プロセスがどうなっているか。証拠として `path:line`・実測値・現行の振る舞いを書く
【前提・依存】必要subset4状態/供給owner・票/受領SHA/共有編集と未選定の停止範囲
【あるべき姿】変更後の振る舞い契約 — UI/API/データとして何が成立するか
【スコープ】やること / やらないこと（Non-goals。本票ではやらない範囲を明記）
【実装方針・手順】取り組み方: 採用案と理由、参照実装（例: features/owners/ の対応パターン）、実装順序（DB → BE → codegen → FE）
【影響範囲】DB / BE / FE 別、migration 要否（要なら「適用はユーザー手動」と明記）
【1PR完了条件・後続】成果型/対象AC→接続/fixture/共有owner/残件・activation gate（必須項目の詳細は固定版参照も可）
【受入条件・検証status】対象AC→接続/fixtureの対応、scopedコマンドと結果又はNOT_RUN・理由（実行許可は別）
```

【現状】を書くためのリポジトリ調査は起票前に行う（`task-create` Phase 2）。未調査のまま【現状】を空欄や推測で埋めてはいけない — 調べ切れない場合はいったん調査票として起票する。

### 調査票（state=Research）

```text
【人間への質問】（あれば本文先頭。形式は上記「人間への質問・決定待ち」節を参照）
【背景】なぜ調べるか
【調査基準】採用main SHA/関連PR/既存調査と採用契約の確認範囲
【前提・依存】調査に必要な資料/権限/未受領と停止範囲
【既知の現状】分かっている事実（path:line・実測値・試したこと）。真っ白な状態での起票は不可
【調査問い】答えるべき問いを列挙
【調査手順】どう調べるか（確認するコード範囲・実験・比較対象・時間BOX）
【対象外】この調査で扱わない範囲
【成果物】レポートの保存先（例: docs/work/<topic>/）と判定文書
【判定基準】調査結果をどう解釈して次アクションを決めるか
```

### ハブ・親票

目的・子票への導線・全体の受入口径のみ。集計対象から外すため `kind:case-hub`（+ `case-category:*`）を付ける。

ハブ票の合格基準は4点のみ: 【目的】【子票一覧】【統合方針】【全体受入】。3問テストの個別要件（実装方針・手順の詳細等）は子票側に課し、ハブ票への過剰記載は避ける。

### 人間作業票（SOP 型、`role:human`）

環境設定・資格情報登録・外部サービス操作・リリース/切替作業など、**人間が手順に従って実行する作業**は SOP（標準作業手順書）構造で書く。「どの画面で・何を操作し・何が見えれば成功か」を実行者が迷わない粒度にする:

```text
【目的】この作業で達成されること・実行のトリガー条件
【調査基準】手順/採用SHA・対象環境の確認時点（未確認は明記）
【前提・依存】必要なアクセス権・アカウント・事前に揃える値・必要subset/依存チケット（EMR-N）
【対象外】この手順で扱わない作業（例: 本番反映は別票）
【手順】
1. <動詞>: <対象画面/ツールでの操作> → 期待結果: <成功の見え方>
2. <動詞>: ... → 期待結果: ...
   ※分岐: <条件> の場合は <別操作>
【検証】完了の確認方法（どの表示/レスポンス/ログを見るか）
【例外・ロールバック】失敗時の対処・元に戻す手順・エスカレーション先
【完了報告】結果の記録先（本票へのコメント、1Password 項目名 等）
```

実例: EMR-270（LINE 資格情報 STG 登録）、EMR-297（本番動作確認）系の票がこの型に相当。

## 外部ソース連携

- `external_source` / `external_id` を出典追跡に使う。既存慣行: `source:slack` → `SLACK-*`、`source:client-sheet` → `row:N`、`source:internal-docs` → `animalekarte-markdown-migration-*`、`source:linear-migration` → external_source=`linear` + external_id=旧 Linear UUID、`source:manual` → 空、bug-report 起票（作成時自動 + 手動再送）→ external_source=`animalekarte-bug-report` + external_id=`<環境名>:<保存済み報告 ID>`（環境名は APP_ENV 由来の `development`/`staging`/`production` allowlist。STG/prod で同じ報告番号でも衝突しない。2026-10-13 以前の既存票は環境名なし数値 ID で backfill しない）+ labels=[`source:manual`, `lane:dev`] 固定（UUID は `PLANE_SOURCE_LABEL_ID`/`PLANE_LANE_LABEL_ID` で注入。2026-10-11 owner 指示で自動起票を許可。既存報告の backfill は行わない）（`source:chatbot` は chat 本文 export 禁止のため形式未確立）。表にない新しい出典（GitHub Issue 起点等）は最初の起票時に形式を決め、この行を更新する
- スプシ・Slack 原文は本文に `【クライアント原文】` 節として転記する（後で出典を遡れるようにする）。秘密・実データを含む場合は転記しない（下記禁止事項）
- bug-report 自動起票は `audit_logs` に action=`support_bug_report.plane_ticket_auto` で結果分類（`claimed`/`creation_failed`/`claim_lost`/`claim_error`）を記録する。監査書込みは報告保存とは別トランザクションの best-effort であり、監査書込み失敗時は ERROR ログで可視化する（DB-atomic ではない点に注意）。監査値は固定分類と safe ID のみで、報告タイトル/本文/URL 等のユーザー入力や Plane 応答本文は含めない

## 成功単位による分割（BE/FEを含む）

- 1票に収まらない変更は **親票 + 子票**（`parent` に親の UUID）で表現する。旧 `-BE` / `-FE` テキスト分割は Plane では使わない。Plane の Epic type・intake（下書き取込）も使わない — 集約は `kind:case-hub` + parent 関係、intake/draft は起票前の仮置きであり正式起票時に本規約を適用する
- 子票は親票の `source:*`・`lane:*` を継承し、本文の【出典】節に親票 `EMR-N` を記す
- 子票の依存は[必要subset4状態](#adopted-baseline-and-dependency-subsets)と供給/受領SHAで記す。待つconsumerと必要subsetを供給する限定実行票/子票の両IDが実在・確定した後、consumerの`blocked_by`を供給票へ作成/同期し、両票本文/決定コメントと同じ受領単位か確認する。既存relationを照合して重複を避け、存在しないIDや機械的な全親待ちで代用しない。受領が必要なscopeは依存待ちとして扱い、独立した限定成果はその完了条件で判断する。BE親票全体のDone待ちを一律条件にしない
- `task-create` の単一責務・自己完結・コード参照という観点を使い、[本規約の成功単位/1PR完了条件](#scope-and-readiness-contract)で粒度を判断する。時間見積りは補助で、時間や票数だけで共有契約/安全性を切り離さない。個人Skillや権限の変更は本改訂に含めない

<a id="claim-and-work-start"></a>
## claim / 作業開始時

- パケット claim protocol（AGENTS.md）の `<TASK-ID>` には Plane ID をそのまま使う: `claim/EMR-298`
- **claim 取得 = state を `Running` に更新する**。`claim/EMR-N` ブランチの取得と Plane の state 遷移は1組で行い、片方だけの状態にしない。`Ready` / `本番切替対応` のいずれもclaim前に対象scopeの[着手契約](#scope-and-readiness-contract)成立を確認・記録してから、着手者が自己遷移してよい。`本番切替対応` は期限によるstateだけで契約成立とみなさず、Readyへの変更は不要で期限情報を保持する。`Backlog` / `Research` は対象scopeの[着手契約](#scope-and-readiness-contract)を満たしReadyへ昇格してからclaimする。調査先行も独立した調査成果のscope/必要subsetを定義し、未確定の実装全体を着手可能とはしない（`本番切替対応` から動かす場合も切替期限の情報は priority・target_date・本文に残る）。`Phase2` / `Needs Human` / `Blocked` の票は claim しない — 滞留・保留理由の解消が先
- state 遷移の権限: 作業者は Running → Review → UAT まで自己遷移してよい（UAT 入り＝人間に受入を依頼する意味）。**Done / Cancelled / Duplicate への遷移は人間の受入コメントまたは明示指示がある場合のみ** — 実装者・エージェントは自己 Done しない。遷移時は証跡（コミット・検証結果・受入記録）をコメントに残す
- **マージ連動**（`.github/workflows/plane-merge-state.yml`）: `main` へマージされた PR の `EMR-N` は Review にする。`staging` へマージされた PR では、その PR の `head.ref` が `main`（production main の staging 取り込み）の場合に限り、差分に含まれる main マージ済み PR の `EMR-N` のうち現在 Review の票だけを UAT にする。feature ブランチなど `main` 以外を head に持つ staging 向けマージでは state を動かさない。`EMR-123`、`[EMR-123]`、`EMR-282/325/326` を同じ番号列として読む。括弧の有無は問わない。マージせずのクローズと `production` は対象外。Done / Cancelled / Duplicate / Needs Human / Blocked / Phase2 は動かさない。UAT 以降の票は main への再マージでも Review に戻さない（退行防止）。state PATCH 成否とは独立に監査コメントを `external_id`（`gh-merge:<PR番号>:<base>:<遷移先>`）で照合し、コメント照合・作成の失敗はその実行を失敗として扱う。再実行時に state が既に遷移済みでも欠けたコメントを1件だけ補完し、既存コメントは二重投稿しない。UAT の票への main マージイベントでも同様に、状態を戻さず監査コメントのみ補完する（補完コメントは実際の現在 state を記し、状態変更がないことを明記する）。シークレット `PLANE_GITHUB_AUTOMATION_KEY` が無いと workflow は失敗する。Plane 公式 GitHub 連携の「マージ成功」への state 割り当ては空のままにする
- 逆方向の遷移（UAT → Running の差戻し、Blocked / Needs Human 解除後の戻し、Running → Backlog の棚卸し）も人間判断または人間のコメント指示に基づく。解除後の戻り先は Backlog（再トリアージ）または Ready（引き続き着手可能なら）。例外はなく、UAT 以降の票は main への再マージでも Review に戻さない
- **Running 不整合チェック**: 「claim ブランチがあるのに unstarted 系 state」「claim が無いのに Running のまま」は運用不整合として発見者が報告する。棚卸し（Running → Backlog 等）自体は上記の通り人間判断
- **委譲・ファンアウト時**: 子セッションのブリーフが Plane 書き込み禁止の場合があるため、Running への遷移はディスパッチする側（オーケストレータ）が行う
- `Backlog → Ready` / `Research → Ready` の昇格（トリアージ）は、対象scopeが[着手契約](#scope-and-readiness-contract)を満たす票なら起票者・着手者が行ってよい（作成時 Ready と同等の判断）。判断に迷うものは人間に振る

## 禁止事項

- **秘密・認証情報・患者/飼主の実データを本文・コメントに書かない**（Plane 移行時に credential パターンを redaction した前例あり。SEC-N3 裁定のうち chat 本文の自動起票禁止は継続 — bug-report の自動起票は 2026-10-11 owner 指示で許可済み）
- **推測で書かない**: 根拠節は起票前にコード・画面・API を確認して `path:line` で書く。確認できていない主張は「未確認」と書く
- **solution-first で起票しない**: 問題（現状とあるべき姿のギャップ）を定義せず解決策から書き始めた票は差し戻し対象
- **受入条件を曖昧にしない**: 「〜が動く」ではなく「〜で〜を入力し〜が表示される」
- **1票に複数の独立した変更を詰め込まない**
- **状態を正本のない憶測で進めない**: Done/Cancelled/Duplicate への遷移は証跡か人間裁定必須

## 附録: UUID 対応表

MCP 経由で `labels` / `state` を指定する際の UUID（`label list` / `state list` が返す現行値。2026-10-08 に Plane MCP で実測確認済み。ID が変わっていたら再取得する）。

### project / workspace

| 対象 | 値 |
|:---|:---|
| workspace | `baritechllc` / `9390d9fe-2f5f-4417-963f-0148872a4d40` |
| project `EMR` | `6a009324-8a8a-455c-b16a-41b3f6207b0e` |

### states

| state | UUID | group |
|:---|:---|:---|
| Backlog | `1e649506-9fea-4f22-ba91-92c05d0811d6` | backlog（default） |
| Ready | `ae54bcb7-360d-44f7-96bc-57f3371eb61b` | unstarted |
| Research | `61539c6a-b8e8-4d68-81da-59026bd9b893` | unstarted |
| 本番切替対応 | `5e1563c8-918a-497f-a934-153fc34909cf` | unstarted |
| Running | `d5c6b02e-dd7d-41c3-851b-b66a0ffd1448` | started |
| Review | `da81bef3-6906-4d2b-8147-ffc2f6652870` | started |
| UAT | `12e5e533-f4bf-48f6-881a-9cf1856841b6` | started |
| Needs Human | `2fb56b99-4162-4ab0-8be4-86c03a399a78` | started |
| Blocked | `5468dd3d-e773-4bde-b77f-01982438d4ac` | started |
| Done | `c781c566-2bf9-44c7-a4c8-f2f32bef19b9` | completed |
| Cancelled | `0dceafc7-2517-4761-9920-edcfec589050` | cancelled |
| Duplicate | `a1023289-83e4-4398-8e19-5fa843c98109` | cancelled |
| Phase2 | `23ec4e15-3731-40be-945e-c827379e61fd` | backlog |

### labels

| label | UUID |
|:---|:---|
| `source:manual` | `31530203-1a6e-445f-9952-01c6eac98fe1` |
| `source:chatbot` | `3357e5e9-f453-4313-9142-4464b400c10f` |
| `source:client-sheet` | `3f2e0827-a986-4069-8320-5ae2f5516163` |
| `source:internal-docs` | `d4bed152-0735-49ca-b783-fae1b2004a00` |
| `source:linear-migration` | `dd6cccaf-cf5f-4dff-9691-8757fd35bc44` |
| `source:slack` | `e61129a1-99f8-497a-bdb3-44451d93e374` |
| `lane:dev` | `7a38738a-d932-4166-a7fa-87524f29b824` |
| `lane:ops` | `2fe9abe7-420c-418f-a8dc-2bbbb5b6b385` |
| `lane:perf` | `f63315e1-26ba-4111-bff5-2dcacfa3bf47` |
| `lane:finance` | `bc1080cf-05be-477e-9434-19d05b6e87aa` |
| `lane:strategy` | `25179b67-0df1-48ea-a5eb-ea6b4bb8c167` |
| `bug` | `b79295b5-0519-4ab2-97cb-a7b08b86e344` |
| `needs-human` | `b364d7f4-c4d1-4a0f-b0b0-f4c97851762e` |
| `Phase2`（label） | `a37d386f-e7fe-4987-adbf-9a0fa16572fc` |
| `risk:high` | `00c06803-7158-464d-acda-3b01a2db4edd` |
| `queue:self-improvement` | `ade23e92-c5b8-4bd6-a17e-b949be3350df` |
| `role:human` | `11262058-7bb5-4d86-b8d1-46db5bdd1fcd` |
| `role:worker` | `ca54278c-08ca-45ef-a419-a550d994ae31` |
| `role:checker` | `7e713989-a60e-46d5-9a8c-92caa3385ca5` |
| `agent:dev-engineer` | `ed5781cf-3daf-435e-a158-455140bae25d` |
| `agent:dev-pm` | `330618b9-6874-41ac-8f92-5d867c289968` |
| `agent:dev-reviewer` | `b60993d3-f597-4a82-9fb5-0b9bdf121028` |
| `agent:ops-orchestrator` | `422939d1-8129-4544-838a-d119b821a9f6` |
| `agent:bookkeeper` | `98a905a4-12d7-418c-aeea-9f074d663c82` |
| `agent:finance-checker` | `de92ceed-a35a-4dc1-9dec-bfda55cf5efa` |
| `agent:mechanism-improver` | `aad6251d-3250-483e-9d34-ef6da5fd2a18` |
| `agent:strategist` | `65e229ea-fcca-4677-9f7b-43e235abb7c2` |
| `kind:case-hub` | `a30d654b-4f7b-473a-a0bd-9cac97e5cbdd` |
| `case-category:client` | `ffb88807-2ef0-421b-b7bf-707585333eaa` |
| `case-category:internal` | `ad034d76-a4f6-4233-8cab-f116239cc2f7` |
| `case-category:service` | `1f746640-2ea5-4638-9fe3-24ce27bd7079` |
