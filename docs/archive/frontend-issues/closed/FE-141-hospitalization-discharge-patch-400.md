# FE-141: 退院処理ボタンが PATCH 400 エラー（フロントの送信データ不正）

**Status**: Open
**Priority**: High
**Affects**: features/hospitalization/
**Date Created**: 2026-03-29
**Related**: BUG-046

---

## Summary

入院詳細ページの「退院処理」実行時に `PATCH /api/v1/hospitalizations/:id` が 400 を返す。
API に直接 `{ "status": "discharged", "end_date": "2026-03-28T00:00:00Z" }` を送ると 200 成功するため、
フロントエンドの送信リクエストボディが不正（必須フィールド欠落またはフォーマット不一致）。

---

## 実装手順

### 1. 原因調査

退院処理ハンドラで実際に送信しているボディを確認：

```bash
grep -rn "discharge\|discharged\|退院処理" frontend/src/features/hospitalization/
```

Network タブで PATCH リクエストのペイロードを確認し、以下をチェック：
- `status: "discharged"` が含まれているか
- `end_date` が ISO 8601 形式（`"2026-03-29T00:00:00Z"`）で含まれているか
- 余分な未知フィールドが含まれていないか

### 2. 修正パターン（推定）

```typescript
// ❌ 現在（推測）: end_date 未設定 or フォーマット不一致
await patchHospitalization(id, { status: "discharged" });

// ✅ 修正: end_date を今日の日付で送信
const today = new Date().toISOString();
await patchHospitalization(id, {
  status: "discharged",
  end_date: today,
});
```

または `end_date` フィールド名が `endDate`（camelCase）になっている可能性：

```typescript
// ❌ camelCase → バックエンドが認識しない
{ status: "discharged", endDate: today }

// ✅ snake_case
{ status: "discharged", end_date: today }
```

### 3. 確認事項

API の UpdateHospitalizationInput（`backend/internal/service/hospitalization_service.go`）と
フロントの送信型（`features/hospitalization/api/types.ts`）のフィールド名・型が一致しているか照合する。

---

## 受入条件

- [ ] 「退院処理を実行」ボタンクリックで `PATCH /api/v1/hospitalizations/:id` → 200
- [ ] ステータスが「退院済」に変更される
- [ ] 一覧「退院済」タブに遷移する
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
