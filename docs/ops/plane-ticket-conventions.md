# Plane チケット起票規約

> **目的**: Plane への起票フォーマットと運用ルールの正本。人間・エージェント・外部流入（スプシ/Slack/チャットボット）のどの経路でも同じ構造で起票し、下流の調査・実装セッションが本文だけで動けるようにする。
> **読者**: このリポジトリで作業する全員・全エージェント。
> **タイミング**: Plane work item を作成・更新するとき。

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
| `source:*`（**必須1つ**） | 手動・エージェント起票 = `source:manual`、Slack転記 = `source:slack`、スプシ = `source:client-sheet`、社内docs移行 = `source:internal-docs`、チャットボット = `source:chatbot`（明示的な reviewable export のみ。自動起票は禁止 — SEC-N3 裁定）。表にない出典（UAT での発見・GitHub Issue 起点・エージェントが調査/実装中に発見した事案）は `source:manual` とし、出典はタイトルアンカーと本文の【出典】節で保持する |
| `lane:*`（1つ） | `lane:dev`（実装）、`lane:ops`（運用）、`lane:perf`（性能）、`lane:finance`、`lane:strategy`。複数に跨る場合は**作業の主対象**で1つ選ぶ（性能改善の実装 → `lane:perf`、会計・料金系の実装 → `lane:finance`、いずれでもない機能実装 → `lane:dev`） |
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
| 受入条件が揃いエージェント/人間がすぐ着手できる | Ready |
| 調査・スパイク票 | Research |
| 人間の判断・操作待ちで滞留 | Needs Human |
| 今期対象外 | Phase2（ラベルではなく state としても存在する。運用は state を優先） |
| 本番切替（2026-10-20）までに完了必須 | 本番切替対応 |

複数に同時に該当する場合は滞留・保留側を優先する: `Phase2` > `Needs Human` > `本番切替対応` > `Ready` / `Research` > `Backlog`（切替期限の情報は priority や本文に併記する。state は workflow 上の実際の位置を表す） |

## 本文テンプレート

プレーンテキストで `【節名】` マーカーを使う（既存票と同じ書式）。1票あたり15〜30行を目安にし、長い根拠はリポジトリ内ファイルへのパス参照にする。種別が複数に見える場合は**実行者優先**で選ぶ — `role:human` の作業（人間が手順で実行するバグ修復・データパッチ等）は SOP 型を使い、`bug` ラベルは性質として併用する。

### 設計思想（採用根拠）

チケットは「**境界付きの設計契約**」として書く — 読んだ実行者（エージェントまたは人）が再調査・再確認なしで着手でき、完了条件が第三者にも検証可能であることを合格基準とする。以下の既知のプラクティスを採用根拠とする:

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

加えて: スコープ外（Non-goals）が明示され、事実と推測が分離されていること。「〜が必要」「〜を検討する」だけで終わる票は禁止。

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
【現象】観測されている挙動（1-2行）
【再現手順】再現頻度（常に/時々/未再現）→ 最短の UI 操作 or API 呼び出し。再現に必要なデータ条件も
【期待挙動】
【実測挙動・根拠】path:line 付きの現状コード or レスポンス実測（起票前にコード確認済みであること）
【影響範囲】clinic_id / owner / pet / staff 分離への影響、隣接画面・LIFF/印刷への波及
【スコープ外】この修正で扱わない範囲（抜本的再設計・別要因の類似不具合は別票へ）
【修正方針】採用案（複数案あるなら案A/案Bと採用理由）。原因が未特定なら「調査票 EMR-N に分離」等と書く
【受入条件】「〜の画面で〜を入力し〜が表示される」形式。エッジケース（空データ・エラー・上限値）を含める
```

事実（実測値・ログ・`path:line`）と推測（原因の仮説）は分けて書く。仮説は【修正方針】の候補案として記し、未検証の断定はしない。

### 機能・タスク票

```text
【人間への質問】（あれば本文先頭。形式は上記「人間への質問・決定待ち」節を参照）
【背景・目的】業務上の目的（依頼原文の要旨。ユーザーの声は要約せず原文引用。責任者が分かるなら明記）
【現状】今の実装/プロセスがどうなっているか。証拠として `path:line`・実測値・現行の振る舞いを書く
【あるべき姿】変更後の振る舞い契約 — UI/API/データとして何が成立するか
【スコープ】やること / やらないこと（Non-goals。本票ではやらない範囲を明記）
【実装方針・手順】取り組み方: 採用案と理由、参照実装（例: features/owners/ の対応パターン）、実装順序（DB → BE → codegen → FE）
【影響範囲】DB / BE / FE 別、migration 要否（要なら「適用はユーザー手動」と明記）
【受入条件】検証可能な条件 + scoped な検証コマンド
```

【現状】を書くためのリポジトリ調査は起票前に行う（`task-create` Phase 2）。未調査のまま【現状】を空欄や推測で埋めてはいけない — 調べ切れない場合はいったん調査票として起票する。

### 調査票（state=Research）

```text
【人間への質問】（あれば本文先頭。形式は上記「人間への質問・決定待ち」節を参照）
【背景】なぜ調べるか
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
【前提条件】必要なアクセス権・アカウント・事前に揃える値・依存チケット（EMR-N）
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

- `external_source` / `external_id` を出典追跡に使う。既存慣行: `source:slack` → `SLACK-*`、`source:client-sheet` → `row:N`、`source:internal-docs` → `animalekarte-markdown-migration-*`、`source:linear-migration` → external_source=`linear` + external_id=旧 Linear UUID、`source:manual` → 空（`source:chatbot` は自動起票禁止のため形式未確立）。表にない新しい出典（GitHub Issue 起点等）は最初の起票時に形式を決め、この行を更新する
- スプシ・Slack 原文は本文に `【クライアント原文】` 節として転記する（後で出典を遡れるようにする）。秘密・実データを含む場合は転記しない（下記禁止事項）

## BE/FE 分割

- 1票に収まらない変更は **親票 + 子票**（`parent` に親の UUID）で表現する。旧 `-BE` / `-FE` テキスト分割は Plane では使わない
- 子票の依存は本文に「`EMR-N`（BE）完了後に着手」と明記し、state は依存側を Backlog に留める
- 粒度の基準は `task-create` スキルの「AIが実装可能」節（単一責務・自己完結・コード参照付き・1-2時間規模）に従う

## claim / 作業開始時

- パケット claim protocol（AGENTS.md）の `<TASK-ID>` には Plane ID をそのまま使う: `claim/EMR-298`
- state 遷移の権限: 作業者は Running → Review → UAT まで自己遷移してよい（UAT 入り＝人間に受入を依頼する意味）。**Done / Cancelled / Duplicate への遷移は人間の受入コメントまたは明示指示がある場合のみ** — 実装者・エージェントは自己 Done しない。遷移時は証跡（コミット・検証結果・受入記録）をコメントに残す
- 逆方向の遷移（UAT → Running の差戻し、Blocked / Needs Human 解除後の戻し、Running → Backlog の棚卸し）も人間判断または人間のコメント指示に基づく。解除後の戻り先は Backlog（再トリアージ）または Ready（引き続き着手可能なら）

## 禁止事項

- **秘密・認証情報・患者/飼主の実データを本文・コメントに書かない**（Plane 移行時に credential パターンを redaction した前例あり。SEC-N3 裁定によりチャットボット経路の自動起票も禁止済み）
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
