# 受入テスト環境セットアップ (UAT Environment Setup)

> **目的**: `scenarios/` 受入（項目単位含む）を実行できる状態にする。  
> **読者**: 受入実施者・AI エージェント。  
> **アーキテクチャ正本**: [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)  
> **最新更新**: 2026-08-14

---

## 1. 前提チェックリスト

| # | 項目 | 合格条件 |
|:--|:--|:--|
| 1 | Docker スタック | `frontend` `:3003` · `backend` `:8080` · `db` healthy |
| 2 | マイグレーション | 起動済み DB が現行 migration 適用済み（エージェントは `make migrate` を自動実行しない） |
| 3 | 認証 env | `.env.local` に `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`（値をログに出さない） |
| 4 | LIFF（local） | `LIFF_MOCK=true`（compose/backend）かつ `VITE_LIFF_MOCK=true`（frontend） |
| 5 | API ログイン | `POST /api/v1/login` が 200 |
| 6 | ブラウザ経路 | (A) Chrome remote debugging `:9222` または (B) Playwright 利用可 |
| 7 | 作業ツリー | 他セッションの WIP を潰さない（`git status`） |

ワンショット確認:

```bash
./docs/ops/testing/scripts/check-uat-env.sh
```

---

## 2. スタック起動（ユーザー操作）

エージェントは `docker compose up` を自動実行しない。未起動時はユーザーが:

```bash
# リポジトリルート
make up          # または docker compose up -d
# migration が必要な pull のあと（ユーザーが実行）
make migrate
```

ヘルス確認:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3003/
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health
```

---

## 3. 認証

```bash
# .env.local（例・値は各自）
E2E_LOGIN_EMAIL=...
E2E_LOGIN_PASSWORD=...
```

- シナリオ・レポート・chat にパスワードを書かない。
- SECTION_14 のロール表は**役割名**で解釈する。メール固定値は seed 依存のため env を正とする。
- ログイン確認（パスワードを echo しない）:

```bash
set -a && source .env.local && set +a
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -H 'X-Requested-With: XMLHttpRequest' \
  -H 'Origin: http://localhost:3003' \
  -d "{\"email\":\"${E2E_LOGIN_EMAIL}\",\"password\":\"${E2E_LOGIN_PASSWORD}\"}"
# 期待: 200
```

---

## 4. ブラウザ経路

### 4A. Chrome DevTools MCP（browser-test 公式）

```bash
# macOS 例 — ユーザーデータは UAT 専用ディレクトリを推奨
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --remote-debugging-port=9222 \
  --user-data-dir=/tmp/chrome-debug-uat \
  --no-first-run \
  about:blank
```

- プロジェクト `.mcp.json` は `http://127.0.0.1:9222` を参照。
- Grok/Claude セッションで chrome-devtools MCP が connected であること（`/mcps`）。

### 4B. Playwright MCP / スクリプト

- `frontend/node_modules/playwright` が存在する（compose frontend ビルド済み想定）。
- フル通しは `reports/uat-YYYY-MM-DD/` 配下にランナーを置き、シナリオ md は編集しない。

### 4C. line-reserve（LIFF mock）

local では BE/FE 双方の mock が必須。compose 既定:

- backend: `LIFF_MOCK=true`
- frontend: `VITE_LIFF_MOCK=true`

STG/本番では mock 禁止。実 LINE は Linear の人間レーン。

---

## 5. レポート出力先

```text
reports/uat-YYYY-MM-DD/
  FINAL.md
  results.json          # formId.fieldKey.Fx を推奨
  bug-candidates.json
  *.png
```

- 製品 FAIL → `bug.md` 必須（env BLOCKED は書かない）。Linear Issue 化は後続
- 環境 BLOCKED → Linear Needs Human または FINAL の BLOCKED 表
- **scenarios/*.md に結果を書かない**

---

## 6. 実施コマンド（エージェント向け要約）

1. `./docs/ops/testing/scripts/check-uat-env.sh` が OK  
2. [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md) の層を確認  
3. S: [scenarios/README.md](scenarios/README.md) の順  
4. V: 各 V + [FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) + [FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md)  
5. `reports/uat-YYYY-MM-DD/FINAL.md` を作成  

---

## 7. トラブルシュート

| 症状 | 確認 |
|:--|:--|
| login 401 | E2E_LOGIN_* · seed アカウント · clinic |
| fe 200 だが API 失敗 | CORS · cookie · `X-Requested-With` |
| LIFF 401 | BE/FE の LIFF_MOCK 不一致 |
| Chrome MCP 死活 | `:9222` listen · `.mcp.json` browserUrl |
| データ汚染 | 名前に Vxx/Sxx プレフィックス · 終了時削除/無効化 |
