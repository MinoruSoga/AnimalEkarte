# E2E・システムテスト実行ガイド (End-to-End Testing)

> **目的**: E2Eテストの実行・追加手順を定義する。
> **読者**: 実装者・QA。
> **タイミング**: E2Eテストの実行・追加時。

> **Animal Ekarte**: Playwright を活用した、主要業務フローの自動検証
> **最新更新**: 2026-06-12

---

## 1. テストの目的

本ガイドは、フロントエンド、バックエンド、データベース、および外部サービス（Mock）が統合された状態で、ユーザーが実際に行う操作（予約から会計まで）が正しく機能することを保証するための手順を定義します。

---

## 2. 重点検証シナリオ (Core Scenarios)

### シナリオ A: 外来ワンストップ・フロー
1.  **ログイン**: 特定のクリニックスタッフとしてログイン。
2.  **受付**: 予約なし患者の当日受付（新規飼主・ペット登録）。
3.  **カルテ**: 身体検査結果（バイタル）および処置・薬品の入力と保存。
4.  **精算**: 会計画面での金額一致確認と支払確定。
5.  **確認**: 飼主の来院履歴および売上合計への反映を確認。

### シナリオ B: 入院・退院管理フロー
1.  **入院登録**: ペット選択からケージ割り当て。
2.  **ケア実施**: デイリーカルテへのバイタル記録（異常値を含む）。
3.  **退院**: 退院処理実行後の、入院費用の会計自動連携。

### シナリオ C: LINE 予約・CRM 連携
1.  **外部予約**: LINE LIFF モックからの診察予約。
2.  **内部確認**: 院内カレンダーへの「source=line」での出現確認。
3.  **タグ発火**: 診察完了後の Lステップへの「リマインドタグ」付与確認。

---

## 3. テスト実行手順

### 3.1 準備
```bash
# 環境変数のセット（playwright.config.ts が参照する変数名。未設定時は http://localhost:3003 が既定）
export PLAYWRIGHT_TEST_BASE_URL=http://localhost:3003
export TEST_ADMIN_USER=admin@example.com
```

### 3.2 実行
```bash
# 全ての E2E シナリオを実行 (Headless)
docker compose exec frontend pnpm test:e2e

# 特定の機能（例：会計）に絞って実行
docker compose exec frontend pnpm test:e2e e2e/accounting-flow.spec.ts

# UI モードで動作を確認しながら実行
docker compose exec frontend pnpm test:e2e --ui
```

---

## 4. 品質基準 (Pass Criteria)

- **ハッピーパス**: 100% の成功率。
- **異常系**: バリデーションエラーが仕様書通りのメッセージで表示されること。
- **テナント隔離**: テスト実行後、他クリニックのデータが汚染されていないこと。

---

## 5. LINE 予約（line-reserve）実機 LIFF 確認手順（手動・FE-refactor.md R-F4）

Playwright は実際の LINE アプリ内 LIFF 起動（`liff.init()` / `liff.isInClient()` / `liff.sendMessages()`
等の実 SDK 挙動）を再現できないため、`line-reserve` の予約導線と `liff`（診察券連携・ペット健康カード）は
**実機での手動確認**をもって E2E の代替とする。自動化（Playwright 経由の LIFF モック起動等）は
本計画（FE-refactor.md）の対象外・別トラックとする。

### 5.1 対象と前提

- 対象アプリ: `frontend/line-reserve/`（LINE予約）・`frontend/liff/`（診察券連携・健康カード）
- 前提: 実機（iOS/Android）の LINE アプリ、対象クリニックの LIFF ID が STG/本番環境に設定済みであること
- ローカル/CI では `VITE_LIFF_MOCK=true` によるモック起動のみ検証可能（`use-liff.ts` / `use-liff-link.ts` の
  単体テストでカバー — `line-reserve/src/pages/*.test.tsx`, `liff/src/hooks/use-liff-link.test.ts`）

### 5.2 line-reserve: 予約作成フロー

1. LINE トークルームの公式アカウントメニューから予約導線を開き、LIFF が正しく起動することを確認する。
2. お客様情報 → コース選択 → （トリミングの場合はコース種別・オプション） → スタッフ選択 → 日付選択 →
   時間選択 → ご要望 → 確認画面 → 確定、の一連の画面遷移が実機で崩れないことを確認する。
3. 確定後、LINE トーク画面に予約内容の自動メッセージ（`ConfirmPage.tsx` の `sendLiffMessage`）が
   送信されることを確認する。
4. 選択した時間枠が確定直前に別経路で埋まった場合の 409 エラー（枠が既に埋まっている旨のアラート→
   時間選択への差し戻し）を、可能であれば2端末での競合予約で確認する。

### 5.3 liff: 診察券連携・健康カード

1. スタッフが発行したリンク（`?token=...&clinic_id=...`）を実機の LINE で開き、`LiffLinkPage` の
   連携フロー（連携中→成功/連携済み/期限切れ/エラー）が実機で正しく表示されることを確認する。
2. 連携済みの LINE アカウントから健康カード（`PetHealthPage`）を開き、ペット一覧・ワクチン記録が
   正しく表示されることを確認する。

### 5.4 確認記録

実施結果は [`docs/FUNCTIONAL_TEST_REPORT.md`](../FUNCTIONAL_TEST_REPORT.md) に記録する
（実施日・端末/OS・LINEアプリバージョン・結果を明記）。

---
