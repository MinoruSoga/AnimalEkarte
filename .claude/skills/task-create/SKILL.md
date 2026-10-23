---
name: task-create
description: "抽象的なタスク依頼を実装粒度に分解し、Plane (EMR) へ起票する。実行 SoT は Plane。起票フォーマットは docs/ops/plane-ticket-conventions.md。repo 入口は todo.md（出典・仕様履歴）。STATUS.md と旧二台帳は復活させない。"
---

# Task Decompose — タスク分解・Plane 起票

抽象的なタスク依頼を受け取り、コードベースを調査した上で:

1. **起票先は Plane**（workspace `baritechllc` / project `EMR`）。タイトル・ラベル・本文テンプレート・UUID 対応表の正本は [`docs/ops/plane-ticket-conventions.md`](../../../docs/ops/plane-ticket-conventions.md)
2. **エージェントの新規起票は承認不要**（2026-10-06 決定）。作成した `EMR-N` は報告に列挙する
3. repo 入口・FAIL・人間レーンは [`todo.md`](../../../todo.md)。Slack/スプシ/GitHub 等の出典がある場合は todo.md 該当節に `→ EMR-N` の対応を残す
4. 複数の独立した成功単位がある場合は **親票 + 成果ごとの子票**（`parent`）で表現する。BE/FE各1票を固定で作らない

> **パス正本の注意**: `STATUS.md`、旧 `3-session-agent.html#ledger`、`backend/issues/` / `frontend/issues/`・docs/tasks 体系、Linear 新規 Issue は廃止済み。復活させない。`bug.md` は空であり、製品 FAIL 台帳は Plane の `bug` ラベルが正本。

## 起動トリガー

ユーザーが以下のようなタスク依頼を行った場合に起動:
- `/task-create <タスク依頼文>`
- 「タスク分解して」「イシュー作って」「チケット切って」

---

## Phase 0: 起票根拠の批判的評価（Phase 1 の前に必須）

タスクファイル・依頼が存在する＝実装前提ではない。分解前に以下を確認し、根拠が弱ければ**廃止候補（WONTFIX）として提示**する:

1. PO 明示要求の記録があるか（PO-QA / SPEC / commit history）
2. 仕様書本体に定義があるか
3. 既存実装との重複がないか（grep で確認。「来院動機」が実装済 recommendation_reason と重複していた実例）
4. 起票根拠が網羅的テストガイド等の「想定リスト派生」なら必要性の二次評価必須
5. SPEC-XXX QY 参照は該当 Q の中身まで突合する（参照ミスの実例あり）

（出典: memory feedback_task_origin_audit）

---

## Phase 1: 識別子と構造（自動）

### 1.1 タスク ID

Plane が採番する `EMR-N` を使う。旧来の `TASK-XXX` 連番採番は廃止。起票後に返る identifier を確定 ID とし、claim ブランチ（`claim/EMR-N`・AGENTS.md packet claim protocol）も同じ ID を使う。

### 1.2 成功単位ごとの親子構造

旧 `TASK-XXX-BE` / `TASK-XXX-FE` のテキスト分割は廃止。Plane の親子構造を使う:

1. 調査後に[正本の成功単位](../../../docs/ops/plane-ticket-conventions.md#scope-and-readiness-contract)で境界付き成果を列挙し、既存の適切な実行票/子票を照合する。層・endpoint・componentだけで分割せず、同じ層にも複数の独立成果があり得る。
2. 集約が必要なら親票を用意し、各成果の範囲/対象AC・完了条件・owner・後続/必要subsetを対応付ける。Phase 4で不足する実行票だけを作成し、返却された各IDを確定する。
3. 依存は[必要subset4状態](../../../docs/ops/plane-ticket-conventions.md#adopted-baseline-and-dependency-subsets)・供給owner/実行票・受領SHAで示し、両ID確定後にPhase 5でPlane relationも同期する。必要API/型/runtime受領待ちと、許可済みmock/契約先行の独立成果を分け、BE親票全体のDoneを一律gateにしない。

---

## Phase 2: コードベース調査（必須）

タスク依頼文から影響範囲を特定するため、以下を調査する。**推測で書かない。必ずコードを読む。**

### 2.1 DB スキーマ調査

```bash
# 関連テーブル・カラムの確認
rg -n "関連キーワード" backend/migrations --glob '*.sql'
```

### 2.2 Backend モデル・API 調査

```bash
# package配置を仮定せずbackendの全Go codeを調査
rg -n "関連キーワード" backend --glob '*.go'
# route/contract
rg -n "関連パス" backend --glob '*.go'
rg -n "関連パス|関連キーワード" backend/docs/api.yaml
```

現行の`model`/`handler`/`service`/`repository` directoryは追加調査の絞り込みに使ってよいが、必須layerや検索範囲の上限にしない。

### 2.3 Frontend コンポーネント・API 調査

```bash
# 関連 feature
ls frontend/src/features/関連feature/
# API hooks
grep -rn "関連キーワード" frontend/src/features/*/api/
# コンポーネント
grep -rn "関連キーワード" frontend/src/features/*/components/
grep -rn "関連キーワード" frontend/src/features/*/routes/
# 型定義
grep -n "関連キーワード" frontend/src/types/generated/models.ts
```

### 2.4 調査結果の整理

調査結果を以下のカテゴリに分類:
- **DB変更が必要**: テーブル追加・カラム追加・マイグレーション
- **Backendモデル変更**: model/*.go の修正 → 必要入力の同一固定SHAを基準に本人が`make codegen`を実行し、agentは正規成果を受領する。agent自動実行/生成物手編集は禁止。必要成果未受領は前提未受領/未完成であり、試験NOT_RUNとは別
- **Backend API変更**: 新規エンドポイント or 既存修正
- **Frontend API変更**: api/ の追加・修正
- **Frontend UI変更**: components/ or routes/ の修正
- **変更不要**: 既に実装済みの部分

### 2.5 類似機能の特定

タスクと似た既存実装を特定し、参照実装として記録する。
- 同じパターンの feature はどれか（例: CRUD なら `owners/` を参照）
- 同じ UI パターンの画面はどれか
- 同じ DB パターンのテーブルはどれか

---

## Phase 2.5: 仕様確認（推測禁止・質問必須）

**このフェーズは Phase 2 の調査完了後、起票（Phase 3）の前に必ず実行する。**

### 目的

タスク依頼文とコードベース調査から判断できない曖昧な点を洗い出し、**推測せずにユーザーに質問する**。

### 2.5.1 曖昧さチェックリスト

以下の観点で、タスク依頼文に不足・曖昧さがないか検証する:

| カテゴリ | チェック項目 |
|---------|------------|
| **UI/UX 仕様** | レイアウト・配置が明確か？ Figma があるか？ 既存画面を参考にするか？ |
| **ビジネスルール** | 権限・条件分岐・バリデーションルールが明確か？ |
| **エッジケース** | 空データ・上限値・エラー時の挙動が定義されているか？ |
| **データフロー** | どこからデータを取得し、どの形式で保存するか明確か？ |
| **既存機能への影響** | 変更が他の画面・API に副作用を与えないか？ |
| **優先度・スコープ** | must-have と nice-to-have の線引きが明確か？ |
| **命名・用語** | UI ラベル・API パラメータ名が確定しているか？ |

### 2.5.2 質問の提示

曖昧な点が **1つでもあれば**、以下のフォーマットでユーザーに質問する:

```
## 仕様確認（回答をお願いします）

調査の結果、以下の点が不明です。起票前に確認させてください。

### 必須（これがないと実装方針が決まらない）
1. [質問内容]
2. [質問内容]

### 任意（デフォルト値で進めてもよい場合はスキップ可）
3. [質問内容] — デフォルト: [推定値]
4. [質問内容] — デフォルト: [推定値]
```

**ルール:**
- 「必須」の質問がある場合は **ユーザーの回答を待ってから Phase 3 に進む**
- 「任意」の質問のみの場合は、デフォルト値を提示し、ユーザーが「それでOK」と言えば進める
- **曖昧な点が一切なければ、このフェーズはスキップして Phase 3 に進む**（無理に質問を作らない）
- 回答が得られない・リポジトリ調査だけでは決められない場合、実装票の起票を無期限に保留せず、論点を【人間への質問】として本文先頭に載せた `needs-human` 票、または【調査問い】を持つ調査票（state=Research）として起票する分岐を取れる（どちらにするかは論点の性質で判断し、迷ったらユーザーに確認する）

### 2.5.3 回答の記録

ユーザーの回答はチケット本文先頭の `【人間への質問】` 節（未解決の場合。形式は規約文書の「人間への質問・決定待ち」節）または受入条件の根拠として記録する。

---

## Phase 3: Plane チケット作成

### 出力先: Plane `workitem create`（`project_id` は規約文書の附録参照）

[`docs/ops/plane-ticket-conventions.md`](../../../docs/ops/plane-ticket-conventions.md) の規約に従う:

- `name`: `出典アンカー + 種別マーカー + 要旨`
- `labels`: `source:*` 必須1つ + `lane:*` 1つ + 該当するもの（`bug` / `needs-human` / `risk:high` 等）
- `state`: 既定は Backlog。[着手契約](../../../docs/ops/plane-ticket-conventions.md#scope-and-readiness-contract)を満たした限定scopeはReady、調査票はResearch、人間待ちはNeeds Human。Readyを環境準備/実行許可/実試験済みとは扱わない
- `description_stripped`: 種別テンプレート（バグ票 / 機能票 / 調査票 / 人間作業票）に従う。1票15〜30行を目安。起票前に規約文書の「3問テスト」（現状・あるべき姿・取り組み方の3点が本文だけで答えられるか）を満たすこと

**推測で書いた節・「調査が必要」の記載は禁止**（調査は Phase 2 で完了させる）。ただし調査票・needs-human 票は未知事項の明示自体が目的のため対象外 — その場合も分かっている事実と調査手順は具体的に書く。

観点の抜け確認には `templates/be-issue.md` / `templates/fe-issue.md` を参照してよい（出力形式は Plane 規約に従う）。

---

## Phase 4: 成果ごとの実行票を確定

Phase 2の調査と仕様確認を反映した独立成果の一覧を一巡する。各成果に適切な既存票があれば再利用し、重複確認後に不足する子票だけを`parent`付きで作成する。単一成果なら不要な親/子票を増やさない。BE1票/FE1票という固定ペアではなく、宣言した成功単位ごとに実行IDを確定し、親AC→各実行IDと残件ownerを網羅対応する。

各票は[着手契約](../../../docs/ops/plane-ticket-conventions.md#scope-and-readiness-contract)に従い、実callerを含む機能成果又は明示した契約/基盤/mock成果のどちらかと、その受入/activation条件を記す。未受領の機能scopeをReadyにしない。

---

## Phase 5: 両ID確定後の依存relation同期

必要subsetの受領を待つconsumerごとに、供給する既存の限定実行票/子票とconsumerの両IDが実在・確定した後、consumerの`blocked_by`を供給票へ作成/同期する。既存relationは読取照合し重複を作らない。供給範囲が親票全体なら限定成功単位の対応を先に確定し、未確定IDを捏造してrelationを作らない。

[正本の両側整合確認](../../../docs/ops/plane-ticket-conventions.md#existing-ticket-scope-reassessment)に従い、両票本文/最新決定コメント/依存記述とrelationが同じsubset・受領条件を指すことを再読確認する。本文だけの依存や不要な全親Done待ちを残さず、関係/状態を根拠なしに一括削除しない。受領SHA等が未確定なら待ち条件として残し、relationの存在だけで受領済みとはしない。

---

## Phase 6: ユーザーへの報告

起票後、以下のサマリーを出力:

```
## タスク分解完了

### 起票済み（N件）
- EMR-NNN: タイトル — state / labels

### 実装順序
1. EMR-NNN — 宣言した契約/基盤又は機能の成功単位・共有owner
2. EMR-NNN — 必要subset/file/symbol・供給票/採用SHAを受領後に結線（許可済みmock/契約先行scopeは別記）
未受領/未選定・停止範囲・activation gate・検証statusを別に報告

状態の正本は Plane。出典があるものは todo.md 該当節に対応行を追記済み。
```

---

## 粒度の基準: 「AIが実装可能」

各チケットは以下の条件を満たす粒度で切る:

1. **単一責務**: 1チケット = 1つの明確な変更目的
2. **自己完結**: 本文と固定版参照で必要な採択契約・前提・停止条件が分かる（着手時の最新main差分再確認は必要）
3. **コード参照付き**: 変更対象のファイルパス・行番号・現在のコードを記載
4. **Before/After明確**: 何を何に変えるかが具体的
5. **完了条件が検証可能**: 「〜が動く」ではなく「〜のレスポンスに〜が含まれる」
6. **宣言した成果が完結**: [規約の成功単位/1PR完了条件](../../../docs/ops/plane-ticket-conventions.md#scope-and-readiness-contract)を正とする。時間/票数/層の小ささだけで判断せず、必要な安全性・共有編集・後続受領まで対応を明記

### 分割の判断基準

[Plane規約の粒度・必要subset・先行成果/機能PRの区別](../../../docs/ops/plane-ticket-conventions.md#scope-and-readiness-contract)に従う。DBとAPI、endpoint単位、component単位という層/ファイル形状だけで機械分割しない。共有file・依存・原子的な成功単位から、宣言した成果が検証可能な形で完結するscopeを決める。

DB+Model/共有API/library/mock FEの先行成果は明示可能だが、供給file/symbol/契約版/固定SHA・後続owner/票・受入・activation gateを必須とする。機能PRは必要なcaller/DI/reader/writer/adapter/routeまで閉じ、helperだけで全機能完了にしない。未選定は[限定停止/質問](../../../docs/ops/plane-ticket-conventions.md#decision-and-shared-edit-boundaries)へ分離する。

`make codegen`は本人実行。必要model入力を採用した同一固定SHAと必要出力を渡し、agentは正規成果を受領する。未生成/未受領の必要型はコード未完成/前提未受領であり、試験を実行していないNOT_RUNとは別に記録する。個人Skill、生成物、権限を本ガイドから横変更しない。

---

## 禁止事項

- **推測でコードを書かない**: 必ずコードベースを読んでから記載
- **仕様を推測で埋めない**: 不明な点は Phase 2.5 でユーザーに質問する。回答を得てから起票に進む（実装票の場合。未回答の論点自体を残す needs-human 票・調査票は Phase 2.5.2 の分岐に従い起票可）
- **存在しないファイルを参照しない**: Glob/Grep で実在を確認
- **チケットに「調査が必要」と書かない**: 調査は Phase 2 で完了させる（実装票の場合。調査票は【調査問い】が本文の本体でありこの限りではない — 既知の事実と調査手順は具体的に書くこと）
- **1チケットに複数の独立した変更を詰め込まない**
- **UI の見た目を推測で記載しない**: Figma デザインがない場合は「UI仕様は別途確認」と明記
- **受入条件を曖昧にしない**: 「〜が動く」ではなく「〜で〜を入力し〜が表示される」レベルで記述する
- **質問を省略しない**: 曖昧な点があるのに質問せず進めることは禁止。ただし曖昧な点がない場合に無理に質問を作ることも禁止
- **秘密・患者/飼主の実データを本文に書かない**（規約文書の禁止事項参照）
