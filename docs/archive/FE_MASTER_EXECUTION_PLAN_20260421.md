# FE Master Features — Execution Plan 2026-04-21

**状況**: スキャン完了、すべての違反を既存タスクにマップ  
**次フェーズ**: 優先度順にタスク実行

---

## 📊 実行計画サマリー

| 優先度 | タスク | パターン | ファイル数 | 推定工数 | 依存関係 |
|--------|--------|---------|----------|--------|---------|
| **1** | TASK-484 | FA7 | 13 | 4h | なし |
| **2** | TASK-491 | FR1 | 3 | 2h | なし |
| **2** | TASK-492 | FR2 | 4-5 | 3h | なし |
| **3** | TASK-488 | FG1 | 2 | 1h | なし |
| **4** | TASK-486 | FA3 | 2 | 1h | PM 決定待機 |

**総工数**: ~11-12時間  
**並列可能**: TASK-491 & TASK-492 (関連性あり、段階的実行推奨)  
**ブロッキング**: TASK-486 (アーキ決定待ち)

---

## 🎯 Phase 1: TASK-484 (FA7 Request Type Derivation)

**優先度**: 最高（型安全性クリティカル）  
**ファイル数**: 13  
**工数**: ~4h

### 対象ファイル

```
frontend/src/features/master/api/
  ├── animal-species.ts (Create, Update)
  ├── cages.ts (Create, Update)
  ├── chief-complaint-types.ts (Create, Update)
  ├── company.ts (Update)
  ├── hospitalization-plans.ts (Create, Update)
  ├── inquiry-templates.ts (Create, Update)
  ├── insurances.ts (Create, Update)
  ├── occupations.ts (Create, Update)
  ├── payment-method-master.ts (Create, Update)
  ├── permission-groups.ts (Create, Update, SetRules)
  ├── staffs.ts (Create, Update)
  ├── trimming.ts (Course Create/Update, Option Create/Update)
  └── types.ts (Create, Update)
```

### 実装ステップ

1. **Omit パターン確認**
   ```typescript
   // 現: export interface CreateXxxRequest { ... }
   // 正: export type CreateXxxRequest = Omit<ModelXxx, 'id' | 'created_at' | 'updated_at'>;
   ```

2. **各ファイルで置き換え**
   - `@/types/generated/models` の Model フィールド確認
   - Omit から除外するフィールド特定
   - 手書き interface → type alias に変更

3. **フォーム検証**
   - 各ファイルの useCreateXxx/useUpdateXxx で型エラーなし
   - API リクエスト payload が backend 期待値と一致

4. **テスト**
   - Build & type-check パス
   - Form submit で API 呼び出し成功

### 成功基準

- [ ] 13ファイルすべてで Create/Update 型が Omit/Partial 導出
- [ ] 手書き interface 完全排除
- [ ] TypeScript エラーなし

---

## 🎯 Phase 2A: TASK-491 (FR1 useMasterCRUD Integration)

**優先度**: 2位（アーキテクチャ一貫性）  
**ファイル数**: 3 → **要スコープ確認** (文書は 2、スキャンは 3)  
**工数**: ~2h

### 対象ファイル

```
frontend/src/features/master/routes/
  ├── DiagnosisSettings.tsx (Line 560 — 手動state)
  ├── ReservationTypeSettings.tsx (Line 247 — 二重entity)
  └── MedicineSettings.tsx ⚠️ (文書未記載、スキャン検出)
```

### 実装ステップ

1. **DiagnosisSettings.tsx**
   - `useMasterCRUD(ResourceMasterDiagnosis)` import
   - 手動 state 削除: `editTarget`, `defaultCategoryId`, `deleteConfirmOpen`
   - Hook から返される state/setter 使用

2. **ReservationTypeSettings.tsx**
   - Primary entity (ReservationType) に hook 使用
   - Secondary entity (Group/Category) 必要に応じて手動 state 継続（スコープ確認）

3. **MedicineSettings.tsx** (要確認)
   - 現在 `useState + useTransition` で CRUD 管理
   - Hook への移行可否判定

### 成功基準

- [ ] 3ファイル useMasterCRUD 使用
- [ ] 手動 CRUD state 削除
- [ ] TypeScript エラーなし

---

## 🎯 Phase 2B: TASK-492 (FR2 useMasterSave Integration)

**優先度**: 2位（TASK-491 と相補）  
**ファイル数**: 4-5 (文書 3、スキャン 4-5)  
**工数**: ~3h

### 対象ファイル

```
frontend/src/features/master/routes/
  ├── DiagnosisSettings.tsx (Line 560 — handleCategorySave, handleNameSave)
  ├── MedicineSettings.tsx (Line 706 — handleSave)
  ├── ReservationTypeSettings.tsx (Line 336 — handleGroupSave, handleCategorySave)
  ├── TrimmingSettings.tsx (Line 620 — handleCourseSave, handleOptionSave)
  └── TreatmentPlanMaster.tsx ⚠️ (スキャン検出、文書未記載)
```

### 実装ステップ

1. **各ファイルで mutation 検出**
   ```typescript
   // 現: mutation.mutate(data, { onSuccess, onError })
   // 正: saveMutation.mutate(data) — hook が onSuccess/onError 処理
   ```

2. **Hook 置き換え**
   - `useMasterSave(ResourceXxx)` import
   - `mutate` を呼び出しのみに簡潔化

3. **useTransition との組み合わせ**
   - `startSaveTransition` は継続 (hook と共存 OK)
   - Hook から `isPending` で loading state 取得

### 成功基準

- [ ] 4-5ファイルで useMasterSave 使用
- [ ] Direct useMutation 完全排除
- [ ] Error handling hook に委譲

---

## 🎯 Phase 3: TASK-488 (FG1 Design Tokens)

**優先度**: 3位（UI 一貫性）  
**ファイル数**: 2 (実スキャン) / 6 (文書)  
**工数**: ~1h

### 対象ファイル（スキャン結果）

```
frontend/src/features/master/routes/
  ├── PermissionGroupSettings.tsx (L202 — color input サイズ)
  └── ReservationTypeGroupSidePanel.tsx (L77 — input レイアウト)
```

**注**: 文書では 6ファイル記載だが、スキャン確認では 2ファイル のみ FG1 違反

### 実装ステップ

1. **LAYOUT tokens 追加**
   ```typescript
   // @/lib/design-tokens.ts に追加
   export const LAYOUT = {
     colorInputSmall: "w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0",
     colorInputMedium: "w-12 h-12 rounded border",
     inputCompact: "px-1.5 py-0.5 rounded-[3px]",
   };
   ```

2. **Hardcode 置き換え**
   - `className="w-7 h-7 ..."` → `className={LAYOUT.colorInputSmall}`

3. **検証**
   - Dynamic color (style={{ backgroundColor }}) は OK
   - Static color は token 使用

### 成功基準

- [ ] 2ファイルで LAYOUT token 使用
- [ ] Tailwind hardcode 完全排除

---

## 🎯 Phase 4: TASK-486 (FA3 Query Key Consistency)

**優先度**: 4位（アーキテクチャ決定待機）  
**ファイル数**: 2 (FA3) / 5 (文書)  
**工数**: ~1h

### 対象ファイル（実違反）

```
frontend/src/features/master/api/
  ├── reservation-type-occupations.ts (FA3 — query key prefix)
  └── reservation-type-unavailable-times.ts (FA3 — query key prefix)
```

**ブロック理由**: Sub-resource query key pattern (hierarchical vs flat) を PM/アーキが決定するまで待機

### 条件付き実装

**If PM 承認: Flat Pattern**
```typescript
const OCCUPATIONS_KEY = ["masters", "reservation-type-occupations"] as const;
```

**If PM 承認: Hierarchical Pattern**
```typescript
const OCCUPATIONS_KEY = (rtId: number) => ["masters", "reservation-types", rtId, "occupations"] as const;
```

---

## ⚠️ スコープ差異と確認事項

| タスク | 文書スコープ | スキャン | 差異 | 対応 |
|--------|-----------|--------|-----|------|
| TASK-484 | 14ファイル | 13ファイル | merchandise-items.ts (FA7) | ✅ 含める |
| TASK-486 | 5ファイル | 2ファイル | company.ts, FA6 違反なし確認 | ⚠️ 再検証待ち |
| TASK-488 | 6ファイル | 2ファイル | ReservationTypeSettings.tsx 他 4つ | ❓ 再確認 (規約準拠かも) |
| TASK-491 | 2ファイル | 3ファイル | MedicineSettings.tsx 追加 | ⚠️ スコープ確認 |
| TASK-492 | 3ファイル | 4-5ファイル | TreatmentPlanMaster.tsx 追加 | ⚠️ スコープ確認 |

**推奨**: TASK-491/492 実行前に、MedicineSettings.tsx と TreatmentPlanMaster.tsx を確認し、スコープを正式決定

---

## 📅 推奨実行スケジュール

```
Week 1:
  Day 1-2: TASK-484 (FA7, 13 files)       [4h]
  Day 3:   TASK-491 + TASK-492 段階1      [2h]
  Day 4:   TASK-492 完了                  [3h]
  Day 5:   TASK-488 (FG1)                 [1h]

Week 2:
  PM 決定後: TASK-486 (FA3)               [1h]
```

**Parallel**: TASK-491/492 は関連性あるため、同時進行可（段階的に）

---

## 🔄 前提条件

1. ✅ スキャン完了 — 22 違反確認
2. ❓ MedicineSettings.tsx、TreatmentPlanMaster.tsx スコープ確認
3. ⏳ TASK-486: PM アーキ決定待ち

---

## ✅ 最終チェックリスト

- [ ] TASK-484 スコープ確定（13 or 14 ファイル）
- [ ] TASK-491/492: MedicineSettings.tsx、TreatmentPlanMaster.tsx 含める判定
- [ ] TASK-488: 6 ファイル中実違反 2 つ確認完了
- [ ] TASK-486: PM 決定待機フラグ立て
- [ ] 開発者アサイン & 実行開始

