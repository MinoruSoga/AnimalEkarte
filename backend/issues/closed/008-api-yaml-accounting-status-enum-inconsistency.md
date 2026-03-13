---
status: open
---

# [api.yaml] GET /accounting の status クエリパラメータ enum が実装と不一致

## 背景

`GET /v1/accounting` のクエリパラメータ `status` の enum 値が api.yaml 内で
実装モデル（`BillingStatus`）と一致していない。

## 問題

```yaml
# api.yaml L2387: GET /accounting クエリパラメータ
- name: status
  in: query
  schema:
    type: string
    enum: [pending, paid, cancelled]   # ← 間違い
```

```go
// model/accounting.go: 実装の BillingStatus enum
BillingStatusWaiting   BillingStatus = "waiting"    // ← yaml にない
BillingStatusCompleted BillingStatus = "completed"  // ← yaml にない
BillingStatusCancelled BillingStatus = "cancelled"  // ← 一致
BillingStatusPending   BillingStatus = "pending"    // ← 一致
// "paid" は存在しない
```

**差分まとめ**:

| 値 | yaml | 実装 |
|----|------|------|
| `waiting` | ❌ ない | ✅ ある |
| `completed` | ❌ ない | ✅ ある |
| `cancelled` | ✅ ある | ✅ ある |
| `pending` | ✅ ある | ✅ ある |
| `paid` | ✅ ある（誤り）| ❌ ない |

なお Billing スキーマ本体（api.yaml L408）は正しく定義されている:
```yaml
# api.yaml L408: Billing schema の status フィールド（正しい）
enum: [waiting, completed, cancelled, pending]
```

つまり同じファイル内でスキーマ定義とクエリパラメータの enum が矛盾している。

## 修正方針

`GET /accounting` のクエリパラメータ enum を実装に合わせる:

```yaml
- name: status
  in: query
  schema:
    type: string
    enum: [waiting, completed, cancelled, pending]  # ← 修正
```

## 完了条件

- [ ] `GET /accounting` の `status` クエリパラメータ enum が `[waiting, completed, cancelled, pending]` になっている
- [ ] api.yaml 内で Billing の status enum 値が全箇所で一致している
- [ ] Swagger UI でドロップダウンに正しい値が表示される
