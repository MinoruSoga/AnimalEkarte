---
name: browser-test
description: 利用可能なブラウザツールを使った機能テスト。docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.mdのシナリオを実行し、結果をテスト結果レポートとして出力する。
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
FAIL 起票先: 確認済み製品 FAIL は root `bug.md` に記録し、その後 Linear Issue 化する。旧 STATUS.md は復活させない。外部投稿の許可がなければレビュー可能な下書きまでで止める。

---

## 実行契約

モデル名で可否を決めない。ユーザースコープのブラウザ接続と現セッションの利用可能toolを確認する。Chrome DevToolsはユーザー設定の固定バージョン・isolatedプロファイル・telemetry無効を既定とし、プロジェクトMCP一覧は空に保つ。既存ブラウザのloopback `http://127.0.0.1:9222` への接続は、その対象をユーザーが明示承認した場合だけ行う。認証は指定環境の契約（`E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` 等）へ案内し、固定資格情報例や値をレポートに書かない。

1. 対象が SECTION_14 なら同ガイド、**scenarios ID（Sxx/Vxx）なら `docs/ops/testing/scenarios/`** から手順を読む。V シリーズは `FIELD-LEVEL-PROTOCOL.md` + `FORM-FIELD-INVENTORY.md` も読む
2. 専用プロファイル・対象URL・合成テストデータ・許可された書き込み範囲を確認する。患者/飼主の実データや個人ブラウザを流用しない。既に許可されたテスト範囲は再承認不要。削除や外部通知など未承認の副作用が必要なら、その手順を止めて独立ケースを続ける
3. 結果をテスト結果レポートとして出力する（シナリオ md / SECTION_14 本体は編集しない）

---

## 実行プロンプトテンプレート

対象セクション情報を埋めてから使う：

```
あなたは Animal Ekarte（動物病院電子カルテシステム）のブラウザ機能テスト担当エージェントです。
Chrome DevTools MCP を使って指定されたテスト項目を実行し、結果を報告してください。

## テスト環境
- URL: http://localhost:3003
- 認証: 環境変数 E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD（値をレポートに書かない）
- ロール: 管理者 / 獣医師 / 受付 は seed の役割名で指定（SECTION_14 §4・UAT-ENV-SETUP）
- ブラウザ: ユーザー設定のChrome DevTools MCP（isolatedプロファイル。既存 :9222 接続には対象の明示承認が必要）
- フォーム項目単位: FIELD-LEVEL-PROTOCOL F0–F6 を inventory 全 fieldKey に適用

## テスト対象
{SECTION_TITLE}
{SCENARIO_STEPS}  ← scenarios または SECTION_14 の番号付き手順を転記

## 実行手順

### Step 1: ブラウザ準備
1. 現セッションのtool定義を確認し、以下のDevTools例を利用可能なAPIへ対応させる。Playwrightは再現テスト、DevToolsは診断、Computer Useは必要な操作に使う。専用プロファイルのページ一覧を確認
2. アプリが開いていなければ mcp__chrome-devtools__new_page で http://localhost:3003 を開く
3. ログイン状態を確認（/login ページなら Step 2 へ、そうでなければ Step 3 へ）

### Step 2: ログイン（未ログイン時のみ）
1. mcp__chrome-devtools__navigate_page で http://localhost:3003/login に移動
2. 利用可能な入力toolで指定環境のテスト用メールアドレスを入力（値は記録しない）
3. 指定環境のテスト用パスワードを入力（値は記録しない）
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
- N/A は仕様上対象外の場合だけ使用。データ不足・ツール不足・未実装・未実施はBLOCKED/未確認として記録する
- スクリーンショットとAPI証跡から患者/飼主情報・資格情報を除外する
- local/candidateの成功をSTG/UAT/PRODやrelease readinessへ昇格しない
```

---

## 結果の記録

1. **テスト結果レポートを出力**する
   - 結果は scenarios / SECTION_14 本体ではなく `reports/uat-YYYY-MM-DD/` またはセッション報告
   - 確認済み製品 FAIL は `bug.md` に記録し、その後 Linear Issue 化する。STATUS.md や旧二台帳は復活させない

2. **サマリを表示**する
   ```
   ## テスト完了: {対象}
   - OK: X件 / NG: Y件 / Partial: Z件
   - 新規バグ: Linear / bug.md（あれば）
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
