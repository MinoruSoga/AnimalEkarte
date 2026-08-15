# 統合テスト・品質保証計画書 (Integration Test Plan)

> **目的**: 統合テスト戦略を定義する。
> **読者**: QA・実装者。
> **タイミング**: 統合テスト戦略の立案時。

> **Animal Ekarte**: システム全体の整合性と信頼性を保証する検証戦略
> **最新更新**: 2026-07-16 | **ステータス**: 推奨検証フェーズ完了

---

## 1. 概要
本計画書は、各機能（Feature）が単体で動作するだけでなく、データベース、外部 API（LINE/Lステップ）、およびフロントエンドの各層が統合された状態で、業務フローが破綻なく完結することを保証するための戦略を定義します。

---

## 2. 検証の三層構造

### 2.1 ユニットテスト (Unit Testing)
- **Backend**: `testify/assert` を用いた table-driven テストによる Service/Repository のロジック検証。
- **Frontend**: `Vitest` + `React Testing Library` によるコンポーネントの挙動検証。
- **合格基準**: 全テストスイートの 100% PASS。カバレッジは ratchet 方式（ベースラインから tolerance 0.5pp を超える低下で CI fail — 正本は [../coverage-policy.md](../coverage-policy.md)）。

### 2.2 API 統合テスト (API Integration)
- **目的**: 実際の DB 接続を伴い、トランザクションや外部キー制約、認可ミドルウェアの連動を確認。CI（`.github/workflows/ci.yml`）の Backend job が PostgreSQL サービスコンテナ + 全マイグレーション適用の上で実行する。
- **重点項目**: クリニック間でのデータリーク防止、一括保存時の不整合回避。
- **補助ゲート**: clinic_id 隔離 lint（Preload clinic-scope）・master-FK write / audit-tx / CASCADE / dbOrTx の各 inventory gate・docs-symbol-drift 等の静的検査は **ローカル `make ci`** で実行する（リモート CI は build/test/gitleaks 等の薄いゲート。詳細: docs/ops/ci-policy.md）。

### 2.3 E2E システムテスト (End-to-End)
- **ツール**: Playwright によるブラウザ自動操作（`frontend/e2e/` 配下の spec 群。設定は `frontend/playwright.config.ts`）。
- **シナリオ**: 「予約の作成」から「会計の完了」までの一連の臨床フローを、実機環境に近い状態で再現。
- **実行**: `frontend/scripts/run-e2e.sh`（CI の `.github/workflows/e2e.yml` も同スクリプトを使用）。手順の正本は [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md)。

### 2.4 手動検証・受け入れシナリオ（三層を補完する層 = L4/L5）

> **層定義の正本**: [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)（L0–L5）。本節は L1–L3 計画との接続用。

- **納品前受け入れ（L4）**: [scenarios/](scenarios/README.md) — 業務フロー S01〜S13 + フォーム V01〜V05。  
  - フォームは **項目単位**まで実施（[scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) + [scenarios/FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md)）。  
  - 環境: [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md)。
- **補完手動（L5）**: [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md)（browser-test スキルが使用するドメイン重点シナリオ）。

---

## 3. 負荷テストとパフォーマンス (k6)

大規模な病院での利用を想定し、以下の基準でシステムの耐久性を検証します。

| 項目 | 目標値 | 検証方法 |
|:---|:---|:---|
| **同時接続数（定常）** | 50 ユーザー | `load-tests/k6-api-endpoints.js` の段階負荷試験。 |
| **応答速度 (p95)** | 500ms 以下 | 同スクリプトの threshold（`p(95)<500`）で機械判定。 |
| **スパイク耐性** | 100 ユーザーへ急増 | `load-tests/k6-spike-test.js`（threshold: `p(95)<2000`）。 |
| **メモリ安定性** | 500MB 以下 | 長時間稼働時のプロファイリング。 |

自動実行は `.github/workflows/performance-tests.yml`（毎日 12:00 JST + `workflow_dispatch`。k6 負荷試験 2 本 + Lighthouse 監査）。STG 向け持続負荷は `load-tests/k6-cf-stg-sustained.js` を手動で使用する。

---

## 4. 実行・レポート手順

1.  **環境構築**: `make up` による検証用クリーン環境の起動。
2.  **テスト実行**: フルスイートは CI（`ci.yml` / `e2e.yml`）が正本。ローカルはスコープ限定で実行する — BE は `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`（フル `go test ./...` は実行禁止 — [`.claude/CLAUDE.md`](../../../.claude/CLAUDE.md) の「Auto-Execution Prohibited Commands / Scoped Verification Exception」）、E2E は `frontend/scripts/run-e2e.sh <spec>`。
3.  **結果集計**: 失敗したテストケースのトリアージと修正、再実行。
4.  **最終報告**: 実施時のテスト結果レポートとして記録（旧 FUNCTIONAL_TEST_REPORT.md は削除済み — git 履歴参照）。

---
