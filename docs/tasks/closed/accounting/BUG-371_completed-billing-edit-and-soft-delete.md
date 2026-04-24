# BUG-371: 精算済会計の修正・論理削除（権限あり時）

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: HIGH（業務運用阻害）
**Affects**: `features/accounting`, `accounting_service.go`

## 実装結果（2026-04-14）

### 完了
- BE: `AccountingService` インターフェース / 実装に `Cancel(ctx, clinicID, id)` を追加
  - 既存値取得 → 二重キャンセル防止（既に `cancelled` なら `ErrConflict`）
  - `UpdateFields` で `status=cancelled` 遷移
- BE: `CancelAccounting` handler 新設（POST `/accountings/:id/cancel`）
- BE: ルート変更: `accountings.DELETE("/:id", ...)` 廃止 → `accountings.POST("/:id/cancel", ...)`
- BE: build / vet 成功
- FE: `api/cancel-accounting.ts` 新規作成
- FE: `AccountingDetail.tsx`
  - `ConfirmDialog` 2 個追加（修正確認 + キャンセル確認）
  - 精算済修正: `editConfirmedRef` + `formRef.requestSubmit()` パターン
    （確認 OK 後、formAction 再実行）
  - キャンセル: `handleCancelConfirm` + `useTransition` + `queryClient.invalidateQueries` + 一覧遷移
  - SubmitButton から `isCompleted` disable 条件削除。ラベル「精算完了済み」→「修正を保存する」
  - ヘッダーに「会計をキャンセル」ボタン（`canDelete && status !== "cancelled"` のみ表示）
- FE: build / lint 成功（エラー 0）

### 設計上の注意
- キャンセルモーダルの `variant="destructive"` で赤系 UI
- ルート `POST /:id/cancel` は `accounting:delete` 権限を要求
- 旧 `DeleteAccounting` / `accountingService.Delete` は残存（ルート未登録のため未使用。別イシューで撤去検討）

### 未完了（別イシュー化候補）
- 修正履歴の監査ログ（精算済修正のトレーサビリティ）
- 修正機能と返金機能の UI ガイダンス
- 不要コード: `accounting_service.Delete`（ルート未登録）の撤去

**依頼元（原文）**:

> 消去修正などは不正防止のため執務のみが権限がある現状です。清算を行ってから修正できないのですが、できるようにしてほしいです
>
> 補足: 現在の実装では権限制御はしております。ただ、削除ができるかは覚えていないです

---

## 概要

`status=completed` の会計に対して、権限保有者（`accounting:edit` / `accounting:delete`）が編集・論理削除できるようにする。BE は既に許可しているが、FE で `isCompleted` 一律ブロックされている。削除は **論理削除（status=cancelled）** で実装し、ハード削除は行わない。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| Q1 | 完了後の修正範囲 | **(A) すべての項目**（明細・支払方法・受領額・保険率） |
| Q2 | 完了後修正の安全策 | **(B) 修正開始時に確認モーダル**「精算済みの会計を修正します」 |
| Q3 | 削除機能 | **(C) 論理削除**（`status=cancelled` に変更）。ハード削除はしない |
| Q4 | 修正後の status | **(A) `completed` のまま即時保存**（再度の確定ボタン不要） |
| Q5 | 「執務」権限 | 既存の `accounting:edit` / `accounting:delete` で十分 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | 精算済会計のキャンセル API（status=cancelled、現行 DELETE は廃止） | BE | BE-111 | - |
| 2 | 完了後 UI ロック解除 + 修正確認モーダル + キャンセルボタン | FE | FE-249 | #1 |

## 受入条件（Acceptance Criteria）

### 編集（Q1=A, Q2=B, Q4=A）
- [ ] **AC-1**: `accounting:edit` 権限保有者が `status=completed` の会計詳細を開いた時、明細追加/削除・支払方法変更・受領額変更・保険適用率変更が可能（`disabled` 解除）
- [ ] **AC-2**: 「会計を確定する」ボタンが「精算完了済み」固定ではなく「修正を保存する」に変わる
- [ ] **AC-3**: 修正開始時（最初の編集アクション時、または保存ボタン押下時）に確認モーダル「**精算済みの会計を修正します。よろしいですか?**」を表示。OK で実行、キャンセルで操作中止
- [ ] **AC-4**: 保存後、status は **`completed` のまま**（waiting に戻さない）。`completed_at` も保持
- [ ] **AC-5**: 権限なし（`accounting:edit` を持たないロール）は従来通り `disabled` のまま
- [ ] **AC-6**: 修正後の `payments` レコードも upsert される（`UpsertPayment` 既存ロジック踏襲）

### 削除（Q3=C 論理削除）
- [ ] **AC-7**: `accounting:delete` 権限保有者の画面に「**会計をキャンセル**」ボタンが表示される（`status` 問わず）
- [ ] **AC-8**: ボタン押下で確認モーダル「**この会計をキャンセルします。元に戻せません**」表示。OK で実行
- [ ] **AC-9**: 実行時は **新規 API `POST /accountings/:id/cancel`** を呼び出し、`status=cancelled` に変更（ハード削除しない）
- [ ] **AC-10**: 既存 `DELETE /accountings/:id`（ハード削除）は本タスクで **削除する**（FE から呼ばないだけでなく、BE のルート登録自体を撤去）
- [ ] **AC-11**: キャンセル後、画面は会計一覧に遷移
- [ ] **AC-12**: キャンセル済 (`status=cancelled`) の会計は再度キャンセル不可（ボタン非表示 or 409）

### 共通
- [ ] **AC-13**: 既存の返金機能（`RefundSection`）はそのまま維持。修正機能と並存
- [ ] **AC-14**: 既存の `waiting` 状態の編集動作は変更なし

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 削除方式 | **論理削除（status=cancelled）** | Q3=C。会計データの監査性確保。完全削除は不正の温床 | ハード削除 / 別テーブル退避 |
| キャンセル API | **新規 `POST /accountings/:id/cancel`** | RESTful 的にステータス遷移を表現。既存 DELETE と意味が異なる | 既存 PATCH に status=cancelled で対応 |
| 既存 DELETE ルート | **撤去** | 「論理削除のみ」ポリシーを徹底するため、ハード削除経路を残さない | 残す（後方互換） — ただし運用で使わなくなるので撤去推奨 |
| 修正後の status | **`completed` のまま** | Q4=A。再確定フローは業務を煩雑化 | `waiting` に戻す |
| 修正確認モーダル | **保存ボタン押下時に1回** | UX 阻害最小。最初の編集アクションで出すと頻繁すぎる | 最初の編集アクション時に出す |
| 権限制御 | **既存 `accounting:edit` / `accounting:delete` 流用** | Q5。新ロール追加は運用負担増 | 新規 `accounting:cancel` 権限追加 |

## 影響範囲

### Backend
- `backend/internal/handler/handler.go:180` — `accountings.DELETE("/:id", ...)` を **削除**、`accountings.POST("/:id/cancel", ...)` に置き換え
- `backend/internal/handler/accounting_handler.go:192` — `DeleteAccounting` を `CancelAccounting` にリネーム / 内部実装を `status=cancelled` PATCH に変更
- `backend/internal/service/accounting_service.go:257` — `Delete(ctx, ...)` を `Cancel(ctx, ...)` にリネーム、内部で `UpdateFields(..., {status: "cancelled"})` を呼ぶ実装に変更
- `backend/internal/service/accounting_service.go:101` — `AccountingService` インターフェース更新

### Frontend
- `frontend/src/features/accounting/api/cancel-accounting.ts` — **新規**作成（`POST /api/accountings/:id/cancel`）
- `frontend/src/features/accounting/api/index.ts` 相当の export 追加
- `frontend/src/features/accounting/routes/AccountingDetail.tsx:608-617` — `disabled={... || isCompleted}` から `isCompleted` 条件を除去、ラベル分岐変更
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` — 修正確認モーダル追加、キャンセルボタン追加、キャンセル確認モーダル追加

### DB
- **変更なし**

## 参照実装

- `backend/internal/service/accounting_service.go:165-202` — Update メソッド（status 制限なしを確認）
- `frontend/src/features/accounting/routes/AccountingDetail.tsx:1168-1176` — 既存 `RefundSection` の確認モーダルパターン
- `frontend/src/components/shared/ConfirmDialog/` — 確認モーダル共通コンポーネント

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 既存 `DELETE /accountings/:id` を呼ぶ他コードの存在 | 高 | grep で全件確認。FE は `delete-accounting.ts` 不在を確認済。他 BE テストや外部システム連携を要確認 |
| 完了後の修正が会計レポート（売上集計）に影響 | 高 | 修正履歴の追跡が将来必要になる可能性。本 issue では監査ログ実装は **対象外** だが、別 issue で追跡推奨 |
| 修正と返金の使い分け不明瞭 | 中 | 将来的に画面ガイダンス追加検討。本 issue では既存 `RefundSection` をそのまま並存させる |
| `cancelled` 状態の会計が `AccountingList` 画面で表示されるか | 中 | 既存 `AccountingList.tsx:78` に `cancelled` フィルタ既存。動作確認のみ |
| `cancelled` 状態の会計に紐づく `billing_items` の扱い | 中 | 削除しない（参照は残す）。論理削除の主旨に整合 |
| キャンセル後の再請求運用 | 低 | 仕様外。新規会計を作り直す運用 |

## 未解決事項

- なし

## 実装順序

1. BE-111: API 実装（DELETE 廃止 → `POST /:id/cancel` 新設）
2. FE-249: 画面実装（UI ロック解除 + 確認モーダル + キャンセル API クライアント）

## 関連イシュー

- BE-111: 精算済会計の論理削除 API（DELETE 廃止 / cancel 新設）
- FE-249: 完了後修正 UI ロック解除 + キャンセル機能
