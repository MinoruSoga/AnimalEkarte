# TASK-031: 返金機能の実装

**作成日**: 2026-03-26
**ステータス**: Open
**依頼元**: 返金機能の実装

---

## 概要

会計精算済み（`completed`）の請求に対して返金処理を行う機能を実装する。
Stripe 準拠モデルを採用し、`billing_refunds` テーブルで返金レコードを独立管理する。
部分返金・全額返金・複数回返金をサポートする。

## 依頼内容（原文）

> 返金機能の実装
>
> 会計精算済み（completed）の請求に対して返金処理を行う機能を実装する。
>
> ### 返金モデル（Stripe準拠）
> - `billing_refunds` テーブルで返金レコードを独立管理
> - `billings.status` は変更しない（completedのまま）
> - 部分返金・全額返金の両方をサポート
> - 返金を複数回に分けて実施可能（例：¥5,000のうち¥2,000 + ¥3,000）
> - 返金理由は任意入力
> - 返金可能額 = 元の請求金額 - 既返金合計（¥0になったらUI上でロック）
>
> ### UI
>
> **会計一覧**
> - 返金済み（一部含む）の会計に「返金あり」バッジを表示
>
> **会計精算画面**
> - 「返金する」ボタン（completed の会計のみ表示）
> - 返金フォーム：返金額（数値入力）・返金理由（テキスト任意）
> - 返金履歴リスト（過去の返金一覧）
> - 返金可能残額の表示

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 返金の種別は？ | 全額・一部両方対応 |
| 2 | 返金理由は必須か？ | 任意 |
| 3 | 返金後のステータス管理 | Stripe モデル採用（billing_refunds テーブルで独立管理、billings.status は completed のまま） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | billing_refunds テーブル + BillingRefund モデル + refund API (GET/POST) | BE | BE-062 | - | [ ] |
| 2 | 会計一覧「返金あり」バッジ + 会計精算画面 返金フォーム・履歴・残額表示 | FE | FE-126 | #1 | [ ] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 会計精算済み（completed）の会計詳細画面に「返金する」ボタンが表示される
- [ ] AC-2: 返金フォームに返金額（必須）・返金理由（任意）を入力して送信すると `POST /accountings/:id/refunds` が呼ばれ、返金レコードが保存される
- [ ] AC-3: 返金後、返金履歴に追加された返金が即座に表示される
- [ ] AC-4: 返金可能残額 = total_amount - Σrefunds.amount が正しく表示される
- [ ] AC-5: 返金可能残額が ¥0 になると「返金する」ボタンが無効化される
- [ ] AC-6: 部分返金を複数回実施できる（例: ¥5,000 の請求に ¥2,000 → ¥3,000 の順で返金可能）
- [ ] AC-7: 会計一覧で返金が1件以上ある会計行に「返金あり」バッジが表示される
- [ ] AC-8: waiting / cancelled 状態の会計には返金ボタンが表示されない
- [ ] AC-9: 返金額が返金可能残額を超える場合、BE がエラーを返し FE にトーストで表示される

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 返金モデル設計 | billing_refunds テーブル独立管理 | Stripe モデル準拠。billings.status を汚染しない。監査証跡として全返金履歴が残る | billings.status に refunded を追加 |
| 返金可能額バリデーション | BE で sum(refunds) を計算して超過チェック | 二重送信等でも DB 側で整合性を保証 | FE のみでチェック |
| 一覧での返金バッジ | GET /accountings レスポンスに total_refunded_amount を追加 | N+1 を避けるため JOIN か COALESCE でまとめて取得 | 別途 refunds API を N回呼ぶ |

## 影響範囲

### DB
- テーブル追加: `billing_refunds` — id, clinic_id, billing_id, amount, reason, refunded_at, created_at

### Backend
- `backend/migrations/001_init.sql` — billing_refunds テーブル追加
- `backend/internal/model/billing_refund.go` — 新規 BillingRefund モデル
- `backend/internal/model/accounting.go` — Billing.Refunds リレーション追加
- `backend/internal/repository/` — refund_repository.go 新規
- `backend/internal/service/` — refund_service.go 新規
- `backend/internal/handler/` — refund_handler.go 新規、accounting_handler.go にルート追加
- `backend/cmd/api/main.go` — DI 配線

### Frontend
- `frontend/src/types/generated/models.ts` — make codegen で自動更新
- `frontend/src/features/accounting/api/` — create-refund.ts, get-refunds.ts 新規
- `frontend/src/features/accounting/api/types.ts` — CreateRefundRequest 追加
- `frontend/src/features/accounting/api/transforms.ts` — refunds フィールド変換追加
- `frontend/src/features/accounting/types/index.ts` — Refund 型追加
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` — 返金セクション追加
- `frontend/src/features/accounting/routes/Accounting.tsx` — 返金バッジ追加

## 参照実装

- `backend/internal/handler/accounting_handler.go` — handler パターン参照
- `backend/internal/service/billing_service.go` — billing service パターン参照
- `features/accounting/routes/AccountingDetail.tsx` — 会計精算画面（返金UI追加先）
- `features/accounting/routes/Accounting.tsx` — 会計一覧（バッジ追加先）

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 返金額超過の二重送信 | 高 | BE でトランザクション内にて sum check → ErrConflict |
| GET /accountings のパフォーマンス劣化 | 中 | LEFT JOIN + COALESCE(SUM, 0) でサブクエリ追加、実行計画確認 |
| make codegen 後の型エラー | 低 | FE-126 実装前に make codegen を先行実行して models.ts を確認 |

## 未解決事項

なし

## 実装順序

1. DB マイグレーション（`billing_refunds` テーブル）
2. Backend モデル追加 → `make codegen`
3. Backend API（handler → service → repository）
4. Frontend API hooks + 型定義
5. Frontend UI（AccountingDetail + Accounting）

## 関連イシュー

- BE-062: [billing_refunds DB + API 実装](../../backend/issues/open/BE-062-billing-refunds-db-and-api.md)
- FE-126: [会計返金 UI 実装](../../frontend/issues/open/FE-126-accounting-refund-ui.md)
