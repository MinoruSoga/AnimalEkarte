---
name: browser-test
description: Chrome DevTools MCPを使ったブラウザ機能テスト。docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.mdのシナリオを実行し、結果をテスト結果レポートとして出力する。Haikuモデルでコスト効率よく実行。
---

# ブラウザ機能テスト スキル

## 使い方

```
/browser-test <ガイド章番号 or ドメイン名 or scenarios ID>
例:
  /browser-test 2.1          # SECTION_14 2.1 外来
  /browser-test accounting   # 2.2 会計
  /browser-test V03          # scenarios V03（項目単位 F 含む）
  /browser-test S01          # scenarios S01
```

**受入正本**: `docs/ops/testing/scenarios/` · アーキテクチャ: `docs/ops/testing/TEST_ARCHITECTURE.md`  
フォーム V は `FIELD-LEVEL-PROTOCOL.md` を全 fieldKey に適用。環境: `UAT-ENV-SETUP.md`。  
FAIL 起票先: root `todo.md` 受入バグ節（旧 STATUS.md 記述は廃止）。

---

## ⚠️ 必須: Haiku Agent で実行せよ

このスキルが呼ばれたら、**以下の手順を必ず守れ**：

1. 対象が SECTION_14 なら同ガイド、**scenarios ID（Sxx/Vxx）なら `docs/ops/testing/scenarios/`** から手順を読む。V シリーズは `FIELD-LEVEL-PROTOCOL.md` + `FORM-FIELD-INVENTORY.md` も読む
2. **`Agent` ツールを `model: "haiku"` で起動**し、ブラウザテストを委譲する
3. Haiku Agent の結果を受け取り、テスト結果レポートとして出力する（シナリオ md / SECTION_14 本体は編集しない）

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
- 認証: 環境変数 E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD（値をレポートに書かない）
- ロール: 管理者 / 獣医師 / 受付 は seed の役割名で指定（SECTION_14 §4・UAT-ENV-SETUP）
- ブラウザ: Chrome（Chrome DevTools MCP · remote debugging :9222）
- フォーム項目単位: FIELD-LEVEL-PROTOCOL F0–F6 を inventory 全 fieldKey に適用

## テスト対象
{SECTION_TITLE}
{SCENARIO_STEPS}  ← scenarios または SECTION_14 の番号付き手順を転記

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
   - 結果は scenarios / SECTION_14 本体ではなく `reports/uat-YYYY-MM-DD/` またはセッション報告
   - NG 項目は root `todo.md` 受入バグ節へ `### BUG-XXX`（ローカル連番 最大+1）。環境・決裁は `todo-po.md`

2. **サマリを表示**する
   ```
   ## テスト完了: {対象}
   - OK: X件 / NG: Y件 / Partial: Z件
   - 新規バグ: BUG-XXX（あれば）
   ```

---

## ドメイン対応表

| 引数 | 正本 | 内容 |
|------|------|------|
| 2.1, outpatient | SECTION_14 §2.1 | 外来フロー |
| 2.2, accounting | SECTION_14 §2.2 | 会計・経営 |
| 2.3, crm | SECTION_14 §2.3 | CRM・Lステップ |
| 3, security | SECTION_14 §3 | 品質ガード |
| S01–S13 | scenarios/S*.md | 業務受入 |
| V01–V05 | scenarios/V*.md + FIELD-LEVEL-PROTOCOL | フォーム項目単位受入 |
