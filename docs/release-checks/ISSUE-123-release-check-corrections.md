# Issue #123 リリースチェック — ローカル訂正と Docker 代替検証

> **このドキュメントの位置づけ**
> GitHub Issue #123（STGリリース前確認チェックリスト）の本文・コメントには、現行コード/スキーマと矛盾する記述が2件ある。
> GitHub 本体の編集はスコープ外のため、ここに**ローカル訂正**と**ローカル Docker による代替検証結果**を集約する。
>
> ⚠️ **STG リリース判断は依然 BLOCKED**。本書はローカルでの**機構（メカニズム）検証**であり、STG ランタイム上の実機証跡ではない。
>
> - 作成日: 2026-06-22
> - 対象 Issue: #123（関連 #114〜#122）
> - 真実の源泉（優先順位）: 実行可能チェック（テスト/スキーマ/生成型）> 現行コード挙動 > Issue 本文

---

## 1. 訂正サマリー（誤 → 正）

| 項目 | Issue 本文の記述（誤） | 現行コードの実態（正） | 根拠 |
|------|----------------------|----------------------|------|
| **P1-3** | DB検証SQL で `WHERE status = 'unpaid'` | `WHERE status = 'waiting'` | `billing_status` ENUM は `('waiting','completed','cancelled','pending')`。`'unpaid'` は**ENUM に存在しない値**で、実行すると `invalid input value for enum billing_status: "unpaid"` エラーになる |
| **P1-5** | `change_amount` 欠落で 400（"change_amount is required" 相当のバインドレベル必須） | 当該 cash リクエストの 400 は**サービス層の現金整合性バリデーション**由来。`change_amount` はバインドレベルでは必須ではない | `accounting_service_builders.go:74-103`（`validatePaymentSplits`）/ `accounting_request.go:193` |

---

## 2. P1-3 訂正詳細 — 月次未納繰越の検証 SQL

### 誤（Issue 本文の例）

```sql
-- ❌ 'unpaid' は billing_status ENUM に存在しない → 実行時 ENUM エラー
SELECT
  COALESCE(SUM(CASE WHEN scheduled_date < '2026-06-01' THEN total_amount ELSE 0 END), 0) AS prev_month_carryover,
  COALESCE(SUM(CASE WHEN scheduled_date >= '2026-06-01' AND scheduled_date <= '2026-06-30' THEN total_amount ELSE 0 END), 0) AS curr_month_unpaid,
  COALESCE(SUM(total_amount), 0) AS next_month_carryover
FROM billings
WHERE clinic_id = 1
  AND status = 'unpaid'           -- ← 誤
  AND scheduled_date <= '2026-06-30';
```

### 正（現行実装 `accounting_repository_unpaid.go: FindMonthlyUnpaidCarryover` に一致）

```sql
-- ✅ 未納の定義は status = 'waiting'（会計待ち）
SELECT
  COALESCE(SUM(CASE WHEN scheduled_date < '2026-06-01' THEN total_amount ELSE 0 END), 0) AS prev_month_carryover,
  COALESCE(SUM(CASE WHEN scheduled_date >= '2026-06-01' AND scheduled_date <= '2026-06-30' THEN total_amount ELSE 0 END), 0) AS curr_month_unpaid,
  COALESCE(SUM(total_amount), 0) AS next_month_carryover
FROM billings
WHERE clinic_id = 1
  AND status = 'waiting'          -- ← 正
  AND deleted_at IS NULL          -- ソフトデリート除外（実装と整合）
  AND scheduled_date <= '2026-06-30';
```

**等式 `次月繰越 = 前月繰越 + 当月未払` の根拠**:
実装は対象集合を `scheduled_date <= lastDay` に限定し、その内部を `< firstDay`（前月繰越）と `firstDay〜lastDay`（当月未払）に分割している。
両者は対象集合の排他的・網羅的な分割なので、`次月繰越（= 全体合計）= 前月繰越 + 当月未払` は**代数的恒等式**として常に成立する。

**未納のステータス語彙について**: このシステムに「unpaid」という状態は存在しない。会計未完了（＝未納）は `billing_status = 'waiting'`（会計待ち）で表現される。`completed` が会計済み、`cancelled` がキャンセル、`pending` が保留。

---

## 3. P1-5 訂正詳細 — `change_amount` と現金会計の 400

### Issue 本文（P1-5 / P2-4）の記述の問題点

Issue は「`change_amount` 欠落 → 400（`change_amount is required` 相当）」と説明しているが、これは**2点で不正確**:

1. **400 の発火源が違う**。ドキュメント記載の cash リクエスト
   ```json
   {"payment_splits":[{"method":"cash","amount":1000}]}
   ```
   が 400 になるのは、`change_amount` が無いからではなく、**`received_amount`（預り金）が省略され 0 になり、`received_amount(0) < amount(1000)` となるため**。
   サービス層 `validatePaymentSplits`（`accounting_service_builders.go:91`）が
   `apperrors.WrapInvalidInput("現金の預り金が不足しています")` を返す。エラー文言は "change_amount is required" ではない。

2. **`change_amount` はバインドレベルで必須ではない**。
   - `paymentSplitRequest.ChangeAmount` は `int64` + `binding:"min=0"`（`accounting_request.go:193`）。`min=0` は**非負を強制するだけで、必須化はしない**（省略時は Go のゼロ値 0 になり `min=0` を通過する）。
   - トップレベル `updateAccountingRequest.ChangeAmount` は `*int64` でバインド制約なし（`accounting_request.go:241`）＝任意。
   - 実際、`received_amount >= amount` を満たしつつ `change_amount` を省略すると、`change(0) == received - amount` が成立する限り**バリデーションを通過する**（例: received=1000, amount=1000, change省略 → 0==0 で OK）。

### 正しい説明

- ドキュメント記載の cash リクエストは 400 を返す。ただしその理由は**サービス層の現金整合性検証**（預り金 < 請求額）であって、`change_amount` のバインド必須ルールではない。
- `change_amount` の整合性が問われるのは「現金 split で `received_amount >= amount` のとき、`change_amount == received_amount - amount` でなければ 400（`お釣り計算が不正です`）」というルール（`accounting_service_builders.go:94`）。

### サービス層バリデーション（現金 split）の全体像

```go
// accounting_service_builders.go: validatePaymentSplits（要旨）
if s.Method == model.PaymentMethodCash {
    if s.ReceivedAmount < s.Amount {
        return apperrors.WrapInvalidInput("現金の預り金が不足しています")  // ← P1-5 の documented request はここ
    }
    if s.ChangeAmount != s.ReceivedAmount-s.Amount {
        return apperrors.WrapInvalidInput("お釣り計算が不正です")
    }
}
```

> テスターへ: P1-5/P2-4 の「400 が返る」という**結果の合否判定はそのまま有効**（documented request は 400 を返す）。
> ただし NG 切り分け時、エラー文言は `現金の預り金が不足しています` であり、`change_amount` を JSON に足しただけでは解決しない（`received_amount` を `amount` 以上にする必要がある）。

---

## 4. ローカル Docker 代替検証結果（P1-1〜P1-6）

> ⚠️ これらは**ローカル機構検証**。STG ランタイム証跡ではない。STG リリース可否は別途 STG 実機確認が必要。

| 項目 | 検証方法 | 結果 | 証跡 |
|------|----------|------|------|
| **P1-1** 権限Seed 12行 | ローカル Docker DB へ読み取り専用 SQL | ✅ LOCAL PASS | `permission_group_rules` で `accounting-cancel`/`accounting-post-close-edit` × 6グループ = 12行。group 1/3/5 `can_edit=t`、2/4/6 `can_edit=f`（期待マトリクスと完全一致） |
| **P1-2** キャンセル予約非表示 | フロント単体テスト + コード確認 | ✅ LOCAL PASS | `ReservationManagement.tsx:17` `filterCalendarAppointments` が `status !== "cancelled"` でフィルタ（no_show 維持）。`ReservationManagement.filter.test.ts` 5件 GREEN |
| **P1-3** 繰越等式 | リポジトリテスト（`ekarte_db_test`）+ SQL 訂正 | ✅ LOCAL PASS | `TestFindMonthlyUnpaidCarryover_*` 9件 GREEN（`_NextMonthCarryoverEquality` 等式 / `_StatusFilter` waiting限定）。実装は `status='waiting'` を使用 |
| **P1-4** base_date 残骸なし | フロント全文 grep | ✅ LOCAL PASS | `frontend/src/` の非テストコードに `base_date`/`baseDate`/`BaseDate` 0件 |
| **P1-5** 現金会計 400 | サービス単体テスト | ✅ LOCAL PASS（機構） | `TestValidatePaymentSplits` 全ブランチ GREEN。「現金: 預り金不足」(received<amount→400) を含む。400 の源はサービス層整合性検証（§3 参照） |
| **P1-6** 一般ロール キャンセル不可 | ルート+Seed+ミドルウェア確認 | ✅ LOCAL PASS（機構） / ⚠️ 実機 HTTP 403 は BLOCKED | `handler.go:278` cancel ルートは `RequirePermission("accounting-cancel","edit")`。一般グループ(2)は `can_edit=false`（P1-1）→ `permission_middleware.go:13-17` で 403 abort。実 HTTP 往復はアプリ起動+認証が必要で未実施 |

### 実行コマンド（参考・再現用）

```bash
# P1-1（読み取り専用 SELECT）
docker compose exec -T db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
  "SELECT group_id, resource, can_view, can_edit FROM permission_group_rules \
   WHERE resource IN ('"'"'accounting-cancel'"'"','"'"'accounting-post-close-edit'"'"') ORDER BY resource, group_id;"'

# P1-3 + P1-5（使い捨て ekarte_db_test を使用。dev DB は変更しない）
docker compose run --rm --no-deps -T --entrypoint go backend \
  test ./internal/service/ ./internal/repository/ \
  -run "TestValidatePaymentSplits|TestFindMonthlyUnpaidCarryover" -v

# P1-2（フロント単体）
docker compose run --rm --no-deps -T frontend \
  npx vitest run src/features/reservations/routes/__tests__/ReservationManagement.filter.test.ts

# P1-4（コード grep）
grep -rn "base_date\|baseDate\|BaseDate" frontend/src/ | grep -v test
```

---

## 5. STG リリース判断

| 区分 | 状態 |
|------|------|
| **ローカル機構検証** | P1-1〜P1-6 すべて LOCAL PASS（P1-6 の実 HTTP 403 のみ未実施） |
| **STG リリース可否** | 🔴 **BLOCKED** — STG ランタイム上の実機証跡（実 DB seed 適用、実フロント描画、実 API 403/400）が無い |

ローカルで機構が正しく実装されていることは確認できたが、これは「STG デプロイ後にその機構が実環境で動作している」ことを保証しない。STG リリース可否は Issue #123 のチェックリストを STG 実機で実施して初めて判定できる。

---

## 6. GitHub Issue #123 への反映（要承認・スコープ外）

以下は GitHub 本体の編集（外部書き込み）であり、明示承認が必要なため**未実施**。承認時の推奨編集テキスト:

- **P1-3 の DB検証SQL**: `status = 'unpaid'` を `status = 'waiting'`（＋ `deleted_at IS NULL`）に置換。「'unpaid' は ENUM に存在せず実行エラーになる」旨を注記。
- **P1-5 / P2-4 の期待結果**: 「400 の理由は現金整合性検証（預り金 < 請求額）であり、`change_amount` のバインド必須ではない」「`change_amount` は `min=0`（非負）制約のみで必須ではない」旨に修正。
