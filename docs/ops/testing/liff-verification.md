# LIFF 検証経路（mock ローカル / 実トークン STG）

> **目的**: BUG-010 の公式方針を固定する。ローカルと CI が保証するのは **mock-token 経路のみ**。実 LINE トークン検証は **STG 手順**であり、秘密は環境変数のみ。実トークン検証をローカル必須ゲートにしない。  
> **読者**: 開発者・QA・AI エージェント・STG 実施者。  
> **関連**: [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) · [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) §5 · [S04](scenarios/S04-liff-reservation-journey.md) · [S12](scenarios/S12-liff-pet-health.md) · [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)

---

## 1. 方針（approach A）

| 経路 | 何を保証するか | 必須ゲートか |
|:---|:---|:---|
| **ローカル mock** | UI/API の mock-token 導線（`VITE_LIFF_MOCK` / `LIFF_MOCK`）。実 LINE サーバ・実 idToken 検証は対象外 | **公式ローカル受入** |
| **CI（mock only）** | compose 既定の mock スタック上の単体/E2E。実 LINE トークン job は置かない | リモートは薄いゲート。実トークン CI は **非必須** |
| **STG 実トークン** | 実機 LINE アプリ + 実 LIFF ID / チャネル秘密での `liff.init`・idToken・連携成立 | 人間 STG レーン。ローカル必須 **ではない** |

**明示**: 実 LINE トークン検証はローカルの mandatory gate ではない。local mock PASS だけで STG/本番の実トークン受け入れ完了とみなさない（[UAT-254-CLOSE-CHECKLIST](scenarios/UAT-254-CLOSE-CHECKLIST.md) と同趣旨）。

---

## 2. ローカル mock が保証すること

- FE: `VITE_LIFF_MOCK=true` のとき共有フックが `mock-token` 等で即 ready（`frontend/src/shared-liff/use-liff.ts`）。設定元は各アプリ `liff-config.ts`（`frontend/liff` / `frontend/line-reserve`）。
- BE: `LIFF_MOCK=true` のとき LIFF 認証ミドルウェアが mock バイパス（`backend/internal/middleware/liff_auth.go`）。
- compose 既定: `docker-compose.yml` が backend に `LIFF_MOCK=true`、frontend に `VITE_LIFF_MOCK=true` を注入（`.env.example` 注記と一致）。
- **保証しないもの**: 実 LINE SDK の `liff.init` / `isInClient` / `sendMessages`、実 idToken 署名検証、STG/本番のチャネル設定ミス検出。
- **モック限界（S12）**: FE mock の連携ページは API を呼ばず success 表示のみになり得る。連携成立の実証明は BE mock API または STG 実 LINE。

release モードでは BE が `LIFF_MOCK=true` を拒否して起動しない（`backend/internal/config/config.go`: `LIFF_MOCK must not be set in release mode`）。STG/本番で mock を足さない。

---

## 3. ローカル受入手順（実トークン不要）

### 3.1 環境フラグ

| 層 | 変数名 | ローカル期待 |
|:---|:---|:---|
| Backend | `LIFF_MOCK` | `true`（compose 既定） |
| Frontend | `VITE_LIFF_MOCK` | `true`（compose frontend service） |

値そのものをログ・ドキュメント・git に書かない。名前のみ参照する。

### 3.2 自動（リポジトリに存在するコマンド）

依存インストール済みのとき（`frontend/node_modules` あり）:

```bash
# 共有 LIFF mock フック
pnpm --dir frontend exec vitest run src/shared-liff/use-liff.test.ts

# 連携フック（mock 分岐含む）
pnpm --dir frontend exec vitest run liff/src/hooks/use-liff-link.test.ts
```

BE mock 認証の回帰（Go テスト・LIFF_MOCK 分岐）:

```bash
cd backend && go test ./internal/middleware/ -count=1 -run 'LiffAuth_Mock'
```

環境ヒント（スタック未起動でもスクリプト自体は実行可）:

```bash
./docs/ops/testing/scripts/check-uat-env.sh
```

Playwright 全体（`make e2e` / `pnpm --dir frontend test:e2e`）は院内 SPA 中心。`frontend/liff` と `frontend/line-reserve` は独立 SPA で Playwright の主対象外（[E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) §5）。ローカル LIFF 受入の公式は **上記 vitest + mock フラグ整合 + scenarios 手動**。

### 3.3 手動 mock 手順（ファイル実在に紐付け）

1. ユーザーが `make up`（エージェントは compose を自動起動しない — [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md)）。
2. `./docs/ops/testing/scripts/check-uat-env.sh` で LIFF mock ヒントを確認。
3. line-reserve: ブラウザで `/line-reserve/{clinicId}/`（末尾スラッシュ）。手順正本は [S04](scenarios/S04-liff-reservation-journey.md)。BE/FE の mock 不一致時は 401。
4. liff ヘルス/連携: `/liff/{clinicId}/`。手順正本は [S12](scenarios/S12-liff-pet-health.md)。mock の success-only 限界に注意。
5. 結果は `reports/uat-YYYY-MM-DD/` へ。`scenarios/*.md` に結果を書かない。

---

## 4. CI は mock only

- リモート必須ゲート（`.github/workflows/ci.yml`）に **実 LINE トークン必須 job は無い**（[ci-policy.md](../ci-policy.md)）。
- 手動 E2E（`.github/workflows/e2e.yml` / `workflow_dispatch`）は `docker compose` スタック前提。compose 既定の `LIFF_MOCK` / `VITE_LIFF_MOCK` により **LIFF トークン観点は mock のみ**。
- approach B（STG 秘密付き実トークン必須 CI job）は本方針の対象外。追加しない。
- CI/ローカルで実トークンを mandatory にしない。

---

## 5. STG 実トークン手順（環境変数のみ・fail-closed）

### 5.1 必要な秘密・設定（名前のみ）

リポジトリ証拠（`.env.example` 等）上の **名前**:

| 用途 | 環境変数名（値は書かない・コミットしない） |
|:---|:---|
| FE LIFF アプリ ID | `VITE_LIFF_ID` |
| LINE チャネル | `LINE_CHANNEL_ACCESS_TOKEN` · `LINE_CHANNEL_SECRET` |
| ローカル専用 mock（STG では **設定しない / false**） | `LIFF_MOCK` · `VITE_LIFF_MOCK` |

クリニック設定上の LIFF ID / チャネルは STG の病院設定 UI またはデプロイ秘密管理に置く。ドキュメント・PR・チャットにトークン値を貼らない。

### 5.2 実施概要

1. STG フロント/バックが **mock なし**（release guard により `LIFF_MOCK=true` は起動不可）。
2. 実機 LINE で [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) §5 および S04/S12 の STG 注記に従う。
3. 予約確定メッセージ・アカウント連携・ヘルスカードを実 idToken で確認する。
4. 証跡は実施日・端末/OS・LINE 版・結果のみ。トークン・secret 値は残さない。

### 5.3 秘密が無いとき（fail-closed）

- STG 実トークン手順に必要な env / クリニック LIFF 設定が揃わない場合:
  - **スキップして偽の PASS にしない**
  - 記録は **BLOCKED** または Linear **Needs Human**
  - ローカル mock 成功を実トークン成功の代替にしない
- エージェントは STG 秘密の取得・本番/STG の破壊的操作・Discord 通知を行わない（人間権限）。

---

## 6. やってはいけないこと

- 実トークン・チャネル secret をリポジトリ・ドキュメント・CI ログに書く
- ローカル必須ゲートとして実 LINE ログインを要求する
- release/STG で `LIFF_MOCK=true` を有効化する
- mock の UI success だけを連携成立の受け入れ完了とする

---

## 7. クイック参照

| 質問 | 答え |
|:---|:---|
| ローカルで何を通す？ | mock フラグ整合 + vitest（use-liff / use-liff-link）+ S04/S12 手動 mock |
| CI で実 LINE は？ | 不要・未実装が正（mock only） |
| 実トークンはどこ？ | STG 実機 + env 名のみの秘密。欠落時は Needs Human / BLOCKED |
| BUG-010 の完了条件は？ | 本ページの経路分割が運用に使われ、実トークンがローカル mandatory でないこと（Done 判定は人間） |
