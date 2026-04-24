# FE-117: src/types/index.ts の TrimmingRecord を TrimmingUI にリネーム

**Status**: Closed
**Priority**: Medium
**Affects**: frontend/src/types/index.ts, features/trimming/api/, features/trimming/routes/
**Date Created**: 2026-03-25
**Related**: TASK-027

## Summary

`src/types/index.ts:207` の `TrimmingRecord`（フロントエンド UI 型）と `models.ts:1252` の `TrimmingRecord`（バックエンドモデル型）が同名で混同を招く。
フロントエンド型を `TrimmingUI` にリネームし、名前衝突を解消する。

## 現状のコード

```typescript
// frontend/src/types/index.ts:207 — フロントエンド UI 型（string IDs, 日本語 status）
export interface TrimmingRecord {
  id: string;
  date: string;
  petId?: string;
  ownerId?: string;
  petNumber: string;
  petName: string;
  ownerName: string;
  species: string;
  weight: string;
  styleRequest: string;
  staff: string;
  status: "完了" | "予約" | "進行中";
  // Form fields
  staffId?: string;
  courseId?: string;
  optionIds?: string[];
  bw?: string;
  bwUnit?: "Kg" | "g";
  bt?: string;
  usedShampoo?: string;
  usedRibbon?: string;
  remarks?: string;
}

// frontend/src/types/generated/models.ts:1252 — バックエンドモデル型（number IDs, snake_case）
export interface TrimmingRecord {
  id: number /* uint64 */;
  clinic_id: number /* uint64 */;
  date: string;
  pet_id?: number /* uint64 */;
  // ... （別物）
}
```

**インポート状況（grep 確認済み — 6ファイル）:**
```typescript
// frontend/src/features/trimming/api/get-trimming.ts:3
import type { TrimmingRecord } from "@/types";

// frontend/src/features/trimming/api/get-trimmings.ts:3
import type { TrimmingRecord } from "@/types";

// frontend/src/features/trimming/api/create-trimming.ts:4
import type { TrimmingRecord } from "@/types";

// frontend/src/features/trimming/api/update-trimming.ts:4
import type { TrimmingRecord } from "@/types";

// frontend/src/features/trimming/api/transforms.ts:1
import type { TrimmingRecord } from "@/types";

// frontend/src/features/trimming/routes/TrimmingList.tsx:37
import type { TrimmingRecord } from "@/types";
```

## 必要な変更

### 1. `src/types/index.ts:207` — interface 名をリネーム

```typescript
// Before
export interface TrimmingRecord {

// After
export interface TrimmingUI {
```

### 2. `features/trimming/api/get-trimming.ts` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";

export const getTrimming = async (id: string): Promise<TrimmingRecord> => {
export const getTrimmingsByPetId = async (petId: string): Promise<TrimmingRecord[]> => {

// After
import type { TrimmingUI } from "@/types";

export const getTrimming = async (id: string): Promise<TrimmingUI> => {
export const getTrimmingsByPetId = async (petId: string): Promise<TrimmingUI[]> => {
```

### 3. `features/trimming/api/get-trimmings.ts` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";
export const getTrimmings = async (): Promise<TrimmingRecord[]> => {

// After
import type { TrimmingUI } from "@/types";
export const getTrimmings = async (): Promise<TrimmingUI[]> => {
```

### 4. `features/trimming/api/transforms.ts` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";
export function transformTrimming(data: BackendTrimming): TrimmingRecord {
  const statusMap: Record<string, TrimmingRecord["status"]> = {

// After
import type { TrimmingUI } from "@/types";
export function transformTrimming(data: BackendTrimming): TrimmingUI {
  const statusMap: Record<string, TrimmingUI["status"]> = {
```

### 5. `features/trimming/api/create-trimming.ts` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";
export const createTrimming = async (...): Promise<TrimmingRecord> => {

// After
import type { TrimmingUI } from "@/types";
export const createTrimming = async (...): Promise<TrimmingUI> => {
```

### 6. `features/trimming/api/update-trimming.ts` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";
export const updateTrimming = async (...): Promise<TrimmingRecord> => {

// After
import type { TrimmingUI } from "@/types";
export const updateTrimming = async (...): Promise<TrimmingUI> => {
```

### 7. `features/trimming/routes/TrimmingList.tsx` — import と型注釈を更新

```typescript
// Before
import type { TrimmingRecord } from "@/types";
// interface TrimmingTableRowProps
  record: TrimmingRecord;
  onDeleteClick: (record: TrimmingRecord) => void;
// handlers
  const handleDeleteClick = useCallback((record: TrimmingRecord) => {

// After
import type { TrimmingUI } from "@/types";
// interface TrimmingTableRowProps
  record: TrimmingUI;
  onDeleteClick: (record: TrimmingUI) => void;
// handlers
  const handleDeleteClick = useCallback((record: TrimmingUI) => {
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）— `TrimmingUI` は transform の戻り型として維持（`ReturnType` への変更は別タスク）

## 依存関係

なし（FE-115 / FE-116 と独立）

## 完了条件

- [ ] `src/types/index.ts` に `interface TrimmingRecord` が存在せず `interface TrimmingUI` が存在する
- [ ] trimming feature の 6 ファイル（`get-trimming.ts`, `get-trimmings.ts`, `create-trimming.ts`, `update-trimming.ts`, `transforms.ts`, `TrimmingList.tsx`）が `TrimmingUI` を使用している
- [ ] `pnpm build` 型エラーゼロ
- [ ] `pnpm lint` エラーゼロ
