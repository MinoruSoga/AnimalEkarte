# BUG-011 回帰確認メモ

## 対象

トリミング＋診察（医療記録由来明細）の統合会計 / 診察のみの新規会計確定が `POST /api/v1/accountings/complete` で 201 になること。

## 修正概要

1. **FE** `use-accounting-completion-action.ts`: complete body に `medical_record_id` を載せる（会計本体 or 明細の一意値）。
2. **FE/BE unbilled**: treatment 由来未請求候補に仮想 `medical_record_id` を返す（`BillingItem.MedicalRecordID` gorm:"-"）。
3. **BE** `resolveCompleteMedicalRecordID`: FE 未送信でも treatment から親カルテを一意解決して billing ヘッダにセット（明示値不一致・複数カルテは拒否）。

## 確認マトリクス

| ケース | 期待 | 確認手段 |
|--------|------|----------|
| トリミングのみ | 201（従来どおり、medical_record_id 不要） | E2E / 手動。unit: complete mock open-period |
| 診察のみ（treatment 付き） | 201、billing.medical_record_id が treatment 親と一致 | BE resolve + unbilled MR 付与 unit |
| トリミング＋診察 統合 | 201 | 上記の合成。E2E 推奨 |
| treatment 複数で親カルテ不一致 | 400 参照組み合わせ不正 | BE resolve ロジック（conflict） |
| 明示 medical_record_id と treatment 親の不一致 | 400 | BE resolve |

## 自動検証（worktree マウント）

```text
# BE（worktree backend を /app に bind、entrypoint なし）
go test ./internal/billing/ -count=1 \
  -run 'TestResolveCompleteMedicalRecordID|TestBillingItemService_GetUnbilledItems_IncludesMedicalAndTrimming|TestAccountingService_CompleteAccounting_OpenPeriod|TestAccountingService_CompleteAccounting_DBSuccess'
# => ok

# FE
pnpm test:run src/features/accounting/api/transforms.test.ts src/features/accounting/api/complete-accounting.test.ts
# => 28 passed
```

## 手動（Needs Human / QA）

1. 同日同ペットでトリミング予約と一般診察を会計待ちまで進める
2. `/accounting/new?petId=X` で統合明細表示を確認
3. 支払設定後「会計を確定する」→ `POST .../accountings/complete` が **201**
4. Network で body に `medical_record_id` が載ること（FE 経路）、または未載でも BE 解決で 201
5. 対照: トリミングのみ 201 / 診察のみ 201

## 非対象

他 BUG、migrate、push/merge
