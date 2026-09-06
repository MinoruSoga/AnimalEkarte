# 会計精算 詳細仕様書 (Accounting Detail)

## 概要
- **画面の目的**: 診療・サービス費用の精緻な計算、保険適用処理、支払い決済、およびインボイス制度で必要となる登録番号・税率別内訳等の帳票項目の出力。
- **URLパターン**: 
  - 新規作成: `/accounting/new?petId=xxx`
  - 編集・精算: `/accounting/:id`
- **アクセス権限**: ルートは `ResourceAccounting`（view/create）。画面内確定・編集（`canSubmit`）は **加えて `ResourceCashRegisterClose` の view** が必須。キャンセル=`ResourceAccountingCancel`、確定後修正=`ResourceAccountingPostCloseEdit`。

---

## 1. 画面構成

### 1.1 明細管理エリア (左カラム)
- **項目リスト**: カルテから自動連携された項目に加え、物販商品や手動項目（自由入力）を統合管理。
- **税率の個別制御**: 
    - 項目ごとに標準税率 (10%)、軽減税率 (8%)、非課税、内税/外税を個別に変更可能。
- **計算サマリ**: 税抜小計、消費税額、税込合計額をリアルタイム表示（税率別の内訳表示は印刷帳票のみ）。

### 1.2 支払い・保険エリア (右カラム)
- **ペット保険窓口精算**:
    - 明細ごとの保険適用フラグに基づき、保険会社が支払う割合（50% / 70% / 90% / 100% の選択式）に応じた控除額を自動算出。飼主の窓口負担率ではない。`InsuranceCard` の選択値を使い、保険マスタの `coverage_rate` からは自動設定しない。
- **決済処理 (Payment Splits)**: 
    - **複数支払対応**: 「現金 5,000 円 ＋ クレジットカード残額」といった複合的な支払い（分割決済）をサポート。
    - **お釣り計算**: お預かり金額からの自動計算補助。

保険会社へのオンライン請求送信（レセプト連携）は未実装。[#240](https://github.com/MinoruSoga/AnimalEkarte/issues/240) は PO の「着手しない」判断で Closed となっており、上記の窓口精算機能と区別する。

---

## 2. 業務・安全機能

### 2.1 精算済みデータの保護と修正
- **修正確認モーダル**: すでに「精算済」となった会計を修正・再保存しようとする際、「精算済みの会計を修正します」という確認ダイアログを強制表示。
- **レジ締め済み期間の編集**: 対象日（新規は JST 当日）がレジ締め確定済みの場合、修正理由（`post_close_reason`）を必須とする。理由が空、または `accounting-post-close-edit:edit` が無いときは確定ボタンを物理ブロックする。BE の `POST /accountings/complete` も FK 解決より先に理由を検証する。
- **会計キャンセル**: 誤請求等の場合、論理削除 (`status=cancelled`) を行うことで、監査証跡を残しつつ請求を無効化。

**確定日時の保護**: 精算済み会計の再保存では `completed_at` を送信せず、最初の確定日時を維持する。バックエンドはクライアントによる確定日時の指定を400で拒否する。支払内訳・金額・締め後修正理由はPATCHで送り、成功後に支払表示・一覧を更新する。新規会計の決済確定は専用Complete経路を使い、未確定の既存会計をPATCHで確定することはできない。

### 2.2 インボイス制度への対応
- **帳票形式**: A4 縦。診療明細書と領収書を 1 枚に集約して印字。
- **必須要件の網羅**: 
    - 適格請求書発行事業者登録番号。
    - 適用税率ごとの区分記載（8%対象、10%対象）。
    - 税率別消費税額。

---

## 3. 技術仕様

### 3.1 集中計算エンジン (`calculations.ts`)
臨床（カルテ会計確認・入院）と会計で同一のロジック (`calculateBillingTotals`) を使用。印刷帳票 (`AccountingDocument`) は共通計算結果を props で受け取って再計算を避け、レイヤー間の丸め差を抑制します。帳票に印字するインボイス用の税率別内訳（8%/10%）は、丸め規則を固定した純関数 `calcTaxBreakdown`（`tax-breakdown.ts`）で算出します。

### 3.2 フロントエンド状態管理
- **`useActionState`**: 保存処理中の重複クリックを無効化（React 19）。
- **`paymentSplits`**: 複合決済の状態を `use-accounting-detail-state.ts` で一元管理。明細更新・返金・キャンセル等の非同期操作は React 19 の `useTransition` で処理中状態を管理。

### 3.3 新規会計確定の再試行

- `Idempotency-Key` はUUID形式が必須。同一医院・同一キー・同一内容の再送は既存会計を返し、支払・監査を再作成しない（初回201、再送200）。同じキーで内容が異なる場合、または削除済み会計に使ったキーの再利用は409となる。
- 画面は失敗後も、入力内容が同じ間はキーを保持して再試行に使う。入力内容の変更や画面再読込では新しいキーになる。通信切断で保存結果が不明なときは、会計一覧で登録状態を確認してから操作する。
- 別ドメインのAPIへ接続する環境では、このヘッダーのCORS許可も必要。バックエンドの固定allowlistに `Idempotency-Key` を含める。デプロイ後の確認は [Vercel STG検証手順](../../ops/deploy/VERCEL-FRONTEND-STAGING-TEST.md) を参照する。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/accountings/:id` | 会計詳細および関連明細の取得 | `accounting` | `view` |
| POST | `/api/v1/accountings/complete` | 新規会計の原子確定（明細・支払・監査。`Idempotency-Key` 必須） | `accounting` | `create` |
| POST | `/api/v1/accountings` | レガシー新規作成（本画面の確定経路では使わない） | `accounting` | `create` |
| PATCH | `/api/v1/accountings/:id` | 既存会計の更新・精算済データの修正（未確定会計の決済確定は不可。確定は `POST /accountings/complete`） | `accounting` | `edit` |
| POST | `/api/v1/accountings/:id/refunds` | 理由を伴う部分返金の記録 | `accounting` | `create` |
| POST | `/api/v1/accountings/:id/credit-correction` | 確定済みカード金額の確定後訂正 | `accounting-post-close-edit` | `edit` |
| POST | `/api/v1/accountings/:id/cancel` | 会計のキャンセル（論理削除） | `accounting-cancel` | `edit` |

---
