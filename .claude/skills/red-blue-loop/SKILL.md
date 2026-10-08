---
name: red-blue-loop
description: 自社アプリ(AnimalEkarte)への許可された攻撃シミュレーション→検知→Plane起票→修正→再攻撃検証を1ラウンドとして反復する Red/Blue セキュリティループ。「セキュリティテスト」「ペネトレ」「脆弱性診断」「レッドチーム」「攻撃して弱点を探して」「攻撃→防御のループを回して」と言われた時に使用。対象はローカル Docker スタックまたは STG のみで、本番は絶対に対象にしない。
---

# Red/Blue セキュリティループ

このリポジトリのアプリ(AnimalEkarte — 動物病院 EMR)を対象に、「攻撃 → 検知 → 起票 → 修正 → 再攻撃」のラウンドを反復するための手順。

## 設計上の前提

- **標準モデルで実施する**。自分が所有するアプリへの権限付きテストという文脈では標準モデルでも拒否は生じにくく、攻撃計画の質は推論力に比例する。無検閲モデルの導入は不要(むしろ量子化劣化で計画の質が落ちる)。
- **評価基準は [SECURITY.md](../../../SECURITY.md)**。invariants(TENANT-1 / AUTHZ-1 / PUBLIC-1 / PRIV-1 / OBJECT-1 / PROVIDER-1 / AUDIT-1 / CONFIG-1 / LIMIT-1)、重大度定義(Critical/High/Medium/Low)、受容済みリスク(再報告禁止リスト)はすべて同文書が正本。本 skill はそれらを複製しない。
- 成果物は「実測された脆弱性」と「再攻撃で使える計画」のみ。推測だけの指摘は起票しない(Plane 規約の「推測で書かない」と同型の失敗パターン)。

## 安全ゲート(ラウンド開始前に必ず通す)

1. **対象の明示**。対象は次のいずれかのみ:
   - ローカル Docker スタック(`http://localhost:<port>`)。ワークツリーで `scripts/tools/worktree-env.sh` が払い出したポートをそのまま使う
   - STG(`animalekarte-stg-frontend.baritech-soga.workers.dev` / `animalekarte-stg-api.baritech-soga.workers.dev`)。synthetic データ専用
   - それ以外(本番 URL、PlanetScale 等の共有インフラ単体、LINE/LSTEP 等の外部 API、第三者サイト)は対象外。ユーザーが対象を明示していない場合は開始前に確認する
2. **スタックの起動はユーザー操作**。`make up` は自動実行禁止コマンド。落ちている場合はユーザーに起動を依頼し、待ち時間は偵察(手順1)を先行させる。
3. **攻撃の性質制限**:
   - DoS・高レート負荷は行わない。並列数は小さく保ち、連番 ID スキャンも数十 ID 程度に抑える(脆弱性の有無は少数の実測で判定できる)
   - データを破壊する payload(DROP/TRUNCATE/全件 DELETE/大量書き込み)は使わない。STG は synthetic とはいえ共有環境
   - 外部コラボレータ(requestbin 等への SSRF 確認)は使わない。応答確認が必要なら Docker ネットワーク内に自前のリスナーを立てる
   - 実マルウェアを使わない。アップロード試験は EICAR テスト文字列や無害なダミーファイルで行う
4. **証跡の置き場所**。スクリーンショット・レスポンス実物は Plane チケットに添付する。リポジトリには置かない(ルート直下 PNG を清掃した前例あり)。トークン・認証情報は ticket 投稿前に必ずマスクする。
5. ツールは Docker 経由を原則とする(`docker run --rm --network host ...`)。ホストの curl による単発 API 呼び出しは可。

## ラウンドの手順

### 0. 準備

- テストアカウント: 複数 clinic の staff(権限違い)× clinic A/B を揃える。ローカル seed や STG デモカタログのアカウントを使い、実ユーザーの認証情報は使わない。足りなければユーザーに確認する。
- ルート票(任意): ループを継続的に回すなら `kind:case-hub` の親票を 1 つ起票し、各ラウンドの findings を子票として親に紐付ける(Plane 規約の BE/FE 分割・親子ルールに従う)。

### 1. 偵察(Recon)

- 攻撃面の列挙: `backend/internal/**/routes*.go` のルート一覧、SECURITY.md の Surface 表(公開境界の条件)、frontend の API 呼び出し箇所。
- 実在ルート・パラメータは必ずソースから確認する。記憶や推測でエンドポイントパスを捏造しない。
- 出力: 「エンドポイント × 認可要件 × 試験方法」のリストをラウンド計画としてまとめる。

### 2. 攻撃計画(Plan)

- [references/attack-catalog.md](references/attack-catalog.md) から invariant ごとの観点を選び、各試験の「期待される合格シグナル」を**実行前に**書く。
- 1 ラウンドで全部を狙わない。テーマを絞る(例: 「ラウンド1 = TENANT-1 の pet/medicalrecord 系」)。
- カタログにない観点を自分で追加する場合は、このファイルの形式(目的/手順/合格基準)に従ってラウンド計画に書き、再現性が確認できたらカタログへ還元する。

### 3. 実行(Execute)

- curl + テストアカウントのトークンで API を叩く。UI を介す試験はブラウザ MCP を使う。
- 実測をすべて記録: リクエスト(トークンはマスク)、レスポンス、ステータス。**合格も不合格も記録する** — 「攻めたが破れなかった」リストが次ラウンド以降の回帰資産になる。
- 失敗(=脆弱性の疑い)を検知したら、最小の再現 curl まで削ってから次へ進む。

### 4. トリアージ

- 重大度は SECURITY.md の表で判定する。clinic 分離(TENANT-1)破れは最小でも High 起点扱い。
- 受容済みリスク・owner 承認済み仕様に当たるものは起票しない(カタログ末尾の「再報告禁止リスト」参照)。
- ソース確認(`path:line`)を起票前に済ませ、事実(実測)と推測(原因仮説)を分離する。

### 5. Plane 起票

- [docs/ops/plane-ticket-conventions.md](../../../../docs/ops/plane-ticket-conventions.md) のバグ票テンプレートに従う。タイトル `[BUG:] 要旨`(出典アンカーは `RedBlue-R<N>:` 形式。N はラウンド番号)、labels は `source:manual` + `lane:dev` + `bug`(+ clinic 分離・権限に踏むなら `risk:high`)、state は受入条件が書けていれば Ready。
- 再現手順は「常に/時々/未再現」を明記し、**エージェントが再実行できる形(マスク済み curl コマンド)**で書く。これが再攻撃(手順6)の入力になる。
- 起票した `EMR-N` をセッション報告に列挙する。

### 6. 修正と再攻撃(回帰)

- 修正自体は通常の実装フロー(`implement-issue`、claim protocol、scoped 検証)に従う。本 skill の担当は起票まで。
- **再攻撃**: 修正済みチケットについて、手順5で記録した同一攻撃を再実行し、塞がったことを実測する。塞がっていない場合はチケットにコメントで残す(state 遷移は Plane 規約の権限に従う)。
- 安定して再現できる攻撃は自動テスト化する(Go httptest での認可テスト等 — `golang-testing` skill の配置規則に従う)。回帰テストが蓄積すると、次ラウンドの Red は「まだ誰も攻めていない領域」だけを狙えばよくなる。

## 完了条件(ラウンド報告)

- 実施した試験の全リスト(合格/不合格の両方)、起票した `EMR-N`、再報告禁止リストに当たって除外したもの
- 実行できなかった試験は PASS にせず BLOCKED/SKIP として理由を書く(「ユーザー手動で」への逃げは `scoped-verification-gates` と同型の chronic 失敗)
- ラウンド計画とサマリの保存先: hub 票コメント or `docs/work/security-red-blue/round-<N>.md`(証跡画像は Plane 添付のみ)

## 関連 skill

- `security-checklist`: 静的なコード側チェックリスト(Blue 側の実装・レビュー時)
- `go-security` / `react-security`: 言語固有の防御実装
- `clinic-id-isolation`: TENANT-1 の実装規約(攻撃対象の境界定義としても読む)
- `implement-issue`: 修正フロー(手順6)
- `scoped-verification-gates`: 修正後のスコープ限定検証
