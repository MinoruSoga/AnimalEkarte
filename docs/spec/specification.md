# システム全体概要 (SPECIFICATION)

> **目的**: システム全体の機能要件と主要業務フローの概要を提供する。
> **読者**: 全読者(新規参加者のオンボーディング含む)。
> **タイミング**: システム概要を把握したい時。

> **Animal Ekarte**: 動物病院向け統合カルテ・経営管理システム
> **最新更新**: 2026-07-29 | **ステータス**: Code-synced（ADR-006 domain package cutover 完了。本ステータスは production 資格認定を意味しない。release gate は [todo.md#ops](../../todo.md#ops) OPS-13〜17）

---

## 1. プロジェクト概要

「Animal Ekarte（アニマル・カルテ）」は、予約管理から診察、入院、会計、そして LINE/Lステップを活用した CRM までをシームレスに統合した、臨床現場の安全と病院経営の効率化を支援する次世代型プラットフォームです。

---

## 2. 核心となる 3 つの設計原則

### 2.1 臨床の安全 (Clinical Safety)
- **ヒューマンエラーの排除**: 死亡ペットに対する誤操作の物理的ブロック、検査異常値の自動ハイライト、ワクチン次回予定の自動計算機能を標準装備。
- **整合性と監査**: 臨床・会計・権限など業務上重要な変更は `audit_logs` に操作者・時刻付きで記録する（全 123 テーブルの自動全件監査ではない。経路ごとに明示実装し、締め後会計編集など clinical/financial integrity が要求する監査は同一 transaction で fail-closed）。臨床記録の真正性を担保する確定（Lock）フローを実装（カルテ画面の確定ボタン＋確認ダイアログによる明示操作、確定後は編集 UI を無効化。確定の解除は不可・訂正は追記のみ。詳細: [screens/06-medical-records-form.md §2.3](screens/06-medical-records-form.md)）。

### 2.2 高い操作性 (Notion-like UX)
- **コンテキストの維持**: 画面遷移を最小限に抑える「サイドパネル編集」と、入力と同時に結果が変わる「リアクティブ検索」を採用。
- **マルチデバイス最適化**: 診察室でのタブレット（iPad等）利用を想定し、クリックターゲットの拡大と視認性の高いフォント設計を徹底。

### 2.3 経営の可視化 (Business Intelligence)
- **リアルタイム集計**: 医院ごとの営業時間に合わせた正確な日次・月次売上、客単価、来院頻度の自動分析。
- **データ駆動型 CRM**: 臨床ステータスと連動した **15 種類**の自動配信トリガーによる、離脱防止と関係構築の自動化。

---

## 3. 主要な業務ドメイン

### 3.1 外来・入院統合サイクル
- **外来**: カンバン形式の受付から SOAPS 形式のカルテ、検査、画像管理までのシームレスな臨床フロー。
- **入院**: ケージ稼働状況の可視化と、時系列でのデイリーケア計画・実施記録。

### 3.2 会計・売掛管理
- **決済**: インボイス制度、ペット保険窓口精算、複雑な税率混在計算への完全対応。
- **経営**: レジ金の実査と過不足管理、月次売上レポートの CSV 出力。

### 3.3 LINE 予約 & Lステップ CRM
- **予約**: スタッフシフトと連動した 24 時間自動受付。
- **CRM**: 顧客の状態（CPM）に合わせた高度な自動配信ロジック（Lステップ連携）。

---

## 4. 技術スタックと信頼性

- **フロントエンド**: React 19 / Tailwind 4 / shadcn/ui による高速な SPA。
- **バックエンド**: Go 1.25 / Gin / GORM による API（スキーマは **123 テーブル**。[ADR-006](../architecture/adr/006-backend-domain-package-boundaries.md) により domain/capability-first の modular monolith へ cutover 済み。production 実装は `internal/<domain>` および命名済み cross-cutting package に置き、旧 layer-first 集約（`internal/handler` / `internal/service` / `internal/repository`）は削除済み。ADR-006 の Implemented は code/package 境界の完了であり、release ready ではない）。
- **認可・セキュリティ**: **37 種類のリソース**に対する RBAC 制御と、クリニック間の物理データ隔離。
- **品質保証**: クリティカルな domain には unit / integration test と write-owner 等の静的 gate を置く。Playwright E2E は表示・主要導線の任意検証であり、全クリティカルパス網羅や全 handler 結合テスト完備を主張しない（方針: [docs/ops/ci-policy.md](../ops/ci-policy.md)、[docs/ops/testing/](../ops/testing/README.md)）。

---
