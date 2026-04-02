---
name: browser-test
description: Chrome DevTools MCPを使ったブラウザ機能テスト。FUNCTIONAL_TEST_REPORT.mdのセクションを実行し結果を更新する。Haikuモデルでコスト効率よく実行。
---

# ブラウザ機能テスト スキル

## 使い方

```
/browser-test <セクション番号 or 機能名>
例:
  /browser-test 1          # Section 1: 飼主・ペット管理
  /browser-test 14         # Section 14: マスタ設定
  /browser-test owners     # 飼主管理
  /browser-test 1-3        # Section 1〜3 まとめて実行
```

---

## ⚠️ 必須: Haiku Agent で実行せよ

このスキルが呼ばれたら、**以下の手順を必ず守れ**：

1. `docs/FUNCTIONAL_TEST_REPORT.md` から対象セクションのテスト項目を読み込む
2. **`Agent` ツールを `model: "haiku"` で起動**し、ブラウザテストを委譲する
3. Haiku Agent の結果を受け取り、`docs/FUNCTIONAL_TEST_REPORT.md` を更新する

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
- テストアカウント: admin@example.com
- パスワード: password
- ブラウザ: Chrome（Chrome DevTools MCP 経由）

## テスト対象
{SECTION_TITLE}
{TEST_ITEMS_TABLE}

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
各テスト項目について以下を実行：

1. **ナビゲーション**: テスト対象ページに移動
2. **操作実行**: テスト項目の操作を実行
3. **結果確認**: 期待動作と実際の動作を比較
4. **ネットワーク確認**: 必要に応じて mcp__chrome-devtools__get_network_request で API レスポンスを確認
5. **コンソール確認**: mcp__chrome-devtools__list_console_messages でエラーがないか確認
6. **スクリーンショット**: NG の場合は mcp__chrome-devtools__take_screenshot で証拠を取得

### Step 4: 結果レポート

以下の形式で結果を報告してください：

```
## テスト結果: {SECTION_TITLE}
実行日時: {DATETIME}

| テスト項目 | 結果 | 備考 |
|-----------|------|------|
| {item1} | OK/NG/Partial/N/A | {観察内容} |
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
- 「--」（未テスト）項目はスキップしてよい
- N/A は実装・データが存在しない場合のみ使用
```

---

## メインセッション（Sonnet）の役割

Haiku Agent の結果を受け取った後：

1. **テストレポートを更新**する
   - `docs/FUNCTIONAL_TEST_REPORT.md` の該当セクションの結果列を更新
   - NG 項目は `docs/tasks/open/crash/` にバグチケットを作成（BUG-XXX 形式）
   - レポート冒頭の「最終更新」日付を更新

2. **サマリを表示**する
   ```
   ## テスト完了: Section {N} - {セクション名}
   - OK: X件 / NG: Y件 / Partial: Z件
   - 新規バグ: BUG-XXX（あれば）
   ```

---

## セクション対応表

| 引数 | セクション | パス |
|------|-----------|------|
| 1, owners | 飼主・ペット管理 | /owners |
| 2, reservations | 予約管理 | /reservations |
| 3, dashboard | ダッシュボード | / |
| 4, medical-records | カルテ管理 | /medical-records |
| 5, examinations | 検査管理 | /examinations |
| 6, accounting | 会計管理 | /accounting |
| 7, hospitalization | 入院・ホテル管理 | /hospitalization |
| 8, vaccinations | 予防接種管理 | /vaccinations |
| 9, trimming | トリミング管理 | /trimming |
| 10, checkups | 定期健診 | /checkups |
| 11, inventory | 在庫管理 | /inventory |
| 12, estimates | 見積管理 | /estimates |
| 13, shifts | シフト管理 | /shifts |
| 14, settings | マスタ設定 | /settings |
| 15, rbac | アカウント・権限管理 | /accounts |
| 16, auth | 認証 | /login |
