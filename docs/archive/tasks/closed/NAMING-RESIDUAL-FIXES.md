# NAMING-RESIDUAL-FIXES: 命名リネーム残存修正

- **作成日**: 2026-04-11
- **ステータス**: CLOSED
- **前提チケット**: TABLE-NAMING-AUDIT（Phase 1-3 完了済み）

---

## 概要

全フォルダ監査の結果、以下 7 件の旧名残存を検出。全て修正が必要。

---

## 検出結果

### 1. Frontend: `ResourceMasterReservationType` の値が旧名

**ファイル**: `frontend/src/types/generated/models.ts:1071`
```typescript
export const ResourceMasterReservationType: Resource = "master-reservation-category";
// → "master-reservation-type" に変更
```
**原因**: models.ts は codegen 生成だが、Go model の定数値が旧名のまま。
**修正**: `backend/internal/model/audit.go`（または該当モデル）の定数値を修正 → codegen 再実行

---

### 2. Frontend: router.tsx のパスに旧名

**ファイル**: `frontend/src/app/router.tsx:734`
```typescript
path: "reservation-category",
// → "reservation-type" に変更
```

---

### 3. Frontend: paths.ts のパスに旧名

**ファイル**: `frontend/src/config/paths.ts:284-285`
```typescript
path: "/settings/diagnosis-category",
getHref: () => "/settings/diagnosis-category",
// → "/settings/diagnosis-type" に変更
```

---

### 4. Frontend: category-config.ts の settingsPath に旧名

**ファイル**: `frontend/src/features/master/constants/category-config.ts:187`
```typescript
settingsPath: "/settings/diagnosis-category",
// → "/settings/diagnosis-type" に変更
```

---

### 5. Frontend: PermissionRuleTable.tsx の表示ラベルキーに旧名

**ファイル**: `frontend/src/features/master/components/PermissionRuleTable.tsx:29`
```typescript
"master-reservation-category": "予約区分",
// → "master-reservation-type": "予約区分" に変更
```

---

### 6. Frontend: diagnosis.ts の query key に旧名

**ファイル**: `frontend/src/features/master/api/diagnosis.ts:69`
```typescript
const DIAGNOSIS_CATEGORIES_KEY = ["masters", "diagnosis-categories"] as const;
// → DIAGNOSIS_TYPES_KEY = ["masters", "diagnosis-types"] に変更
```

---

### 7. Migration: `master-insurance` は誤検出

**ファイル**: `backend/migrations/003_seed_demo.sql:266` 等
```sql
(1, 'master-insurance', true, true, true, true),
```
**判定**: これは permission_group_rules の `resource` カラムの値 `"master-insurance"`。`insurance` はテーブル名 `insurances` のリソース識別子であり、旧カラム名ではない。**変更不要**。

---

## 修正計画

| # | ファイル | 修正内容 | 影響 |
|---|---------|---------|------|
| 1 | Go model (audit.go等) | `"master-reservation-category"` → `"master-reservation-type"` | codegen → models.ts 自動更新 |
| 2 | router.tsx:734 | `"reservation-category"` → `"reservation-type"` | URL パス変更 |
| 3 | paths.ts:284-285 | `"/settings/diagnosis-category"` → `"/settings/diagnosis-type"` | URL パス変更 |
| 4 | category-config.ts:187 | settingsPath を更新 | URL パス変更 |
| 5 | PermissionRuleTable.tsx:29 | リソースキー名を更新 | 表示のみ |
| 6 | diagnosis.ts:69 | 変数名 + query key 値を更新 | キャッシュキー変更 |
