---
name: browser-test
description: Chrome DevTools MCPを使ったブラウザ機能テスト。docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.mdのシナリオを実行し、結果をテスト結果レポートとして出力する。Haikuモデルでコスト効率よく実行。
---

# ブラウザ機能テスト スキル

## 使い方

```
/browser-test <ガイド章番号 or ドメイン名>
例:
  /browser-test 2.1          # 2.1 外来フロー（予約〜受付〜診察）
  /browser-test accounting   # 2.2 会計・経営管理
  /browser-test crm          # 2.3 CRM・Lステップ連携
  /browser-test 3            # 3. 品質ガード・セキュリティ
```

---

## ⚠️ 必須: Haiku Agent で実行せよ

このスキルが呼ばれたら、**以下の手順を必ず守れ**：

1. `docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md`（正本）から該当ドメインのシナリオ（番号付き手順）を読み込む
2. **`Agent` ツールを `model: "haiku"` で起動**し、ブラウザテストを委譲する
3. Haiku Agent の結果を受け取り、テスト結果レポートとして出力する（ガイドに項目表・結果列は存在しないため、ガイド自体は更新しない）

> 旧 FULL_DOMAIN_SCENARIO_TEST_GUIDE.md は廃止済み。手動シナリオは SECTION_14_MANUAL_TEST_GUIDE.md、E2Eは E2E_TESTING_GUIDE.md を参照

メインセッション（Sonnet）は直接 Chrome DevTools MCP ツールを呼ばないこと。
すべてのブラウザ操作は Haiku Agent に委譲する。

---

## Haiku Agent へ渡すプロンプトテンプレート

以下を Haiku Agent のプロンプトとして使用せよ（対象セクション情報を埋めてから渡す）：

```
あなたは Animal Ekarte（動物病院電子カルテシステム）のブラウザ機能テスト担当エージェントです。
Chrome DevTools MCP を使って指定されたテスト項目を実行し、結果を報告してください。

## テスト環境
- URL: http://localhost:3003
- テストアカウント: 下記（正本: docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md 4章）
  | 役割 | ログイン ID | 用途 |
  |------|------------|------|
  | 管理者 | admin@example.com | マスタ設定、権限、会計レポート |
  | 獣医師 | doctor@example.com | カルテ、検査、処方の臨床フロー |
  | 受付/看護 | staff@example.com | 受付、会計、入院ケアの日常業務 |
- パスワード: password
- ブラウザ: Chrome（Chrome DevTools MCP 経由）

## テスト対象
{SECTION_TITLE}
{SCENARIO_STEPS}  ← ガイドの該当ドメインの番号付き手順シナリオを転記

## 実行手順

### Step 1: ブラウザ準備
1. mcp__chrome-devtools__list_pages でページ一覧を確認
2. アプリが開いていなければ mcp__chrome-devtools__new_page で http://localhost:3003 を開く
3. ログイン状態を確認（/login ページなら Step 2 へ、そうでなければ Step 3 へ）

### Step 2: ログイン（未ログイン時のみ）
1. mcp__chrome-devtools__navigate_page で http://localhost:3003/login に移動
2. mcp__chrome-devtools__fill でメールアドレス入力: admin@example.com
3. mcp__chrome-devtools__fill でパスワード入力: password
4. mcp__chrome-devtools__click でログインボタンをクリック
5. mcp__chrome-devtools__wait_for でダッシュボード表示を待機

### Step 3: テスト実行
各シナリオ手順について以下を実行：

1. **ナビゲーション**: テスト対象ページに移動
2. **操作実行**: シナリオ手順の操作を実行
3. **結果確認**: 期待動作と実際の動作を比較
4. **ネットワーク確認**: 必要に応じて mcp__chrome-devtools__get_network_request で API レスポンスを確認
5. **コンソール確認**: mcp__chrome-devtools__list_console_messages でエラーがないか確認
6. **スクリーンショット**: NG の場合は mcp__chrome-devtools__take_screenshot で証拠を取得

### Step 4: 結果レポート

以下の形式で結果を報告してください：

```
## テスト結果: {SECTION_TITLE}
実行日時: {DATETIME}

| シナリオ手順 | 結果 | 備考 |
|-------------|------|------|
| {step1} | OK/NG/Partial/N/A | {観察内容} |
...

### 発見したバグ（NG 項目）
- {バグ説明}

### 総括
- 合計: {total}件
- OK: {ok}件 / NG: {ng}件 / Partial: {partial}件 / N/A: {na}件 / 未確認: {unknown}件
```

## 重要なルール
- 操作の間は必ず mcp__chrome-devtools__wait_for で応答を待つ（タイムアウト: 5000ms）
- API 呼び出しが含まれるテストは mcp__chrome-devtools__get_network_request でステータスコードを確認
- エラーが出た場合はスクリーンショットを取得してから次のテストへ進む
- N/A は実装・データが存在しない場合のみ使用
```

---

## メインセッション（Sonnet）の役割

Haiku Agent の結果を受け取った後：

1. **テスト結果レポートを出力**する
   - 結果はガイド（`docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md`）の更新ではなく、テスト結果レポートとして出力する（ガイドに項目表・結果列は存在しない。ガイド冒頭の日付表記は「最新更新」であり、テスト結果では書き換えない）
   - NG 項目はリポジトリ直下 `STATUS.md`（受入テストバグ台帳）の末尾へ `## BUG-XXX:` 節として起票（ローカル連番 最大+1。症状・再現手順・調査の起点を記載）。バグではない改善事項は `STATUS.md` の `## 個別タスク詳細` へ `### TASK-XXX:` 節として起票（旧 `3-session-agent.html#ledger` は 2026-07-31 廃止）

2. **サマリを表示**する
   ```
   ## テスト完了: {ガイド章} - {ドメイン名}
   - OK: X件 / NG: Y件 / Partial: Z件
   - 新規バグ: BUG-XXX（あれば）
   ```

---

## ドメイン対応表（正本: docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md の章構成）

| 引数 | ガイド章 | 内容 | 主なパス |
|------|---------|------|---------|
| 2.1, outpatient | 2.1 外来フロー（予約〜受付〜診察） | 予約作成・当日受付・カルテ入力・次回来院設定 | /reservations, / |
| 2.2, accounting | 2.2 会計・経営管理 | 会計精算・レジ締め・月次レポート | /accounting/close, /accounting/reports |
| 2.3, crm | 2.3 CRM・Lステップ連携 | タグ管理・対象者抽出・個別送信 | /settings/lstep/tags, /lstep/checkup-sync |
| 3, security | 3. 品質ガード・セキュリティ | RBAC 権限ガード・削除保護(FK)・離脱防止 | 各画面横断 |
