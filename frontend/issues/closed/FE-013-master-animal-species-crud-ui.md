# FE-013: マスタ設定画面 — 動物種類（animal_species）管理UI追加

**Status**: Open
**Priority**: Medium
**Affects**: master feature — マスタ設定
**Date Created**: 2026-03-17
**Related**: TASK-001, BE-040

## Summary

マスタ設定画面に動物種類（animal_species）のCRUD管理UIを追加する。現在は List API のみで、フロントエンドのマスタ設定画面には動物種類の管理セクションがない。既存のマスタ設定画面パターン（diagnosis_category 等）に準拠して実装する。

## 現状のコード

### 既存の animal_species API hook

```typescript
// frontend/src/features/owners/api/get-animal-species.ts:1-17
export const getAnimalSpecies = async (): Promise<BackendAnimalSpecies[]> => {
  const { data } = await axios.get<BackendAnimalSpecies[]>("/v1/masters/animal-species");
  return data;
};

export const useGetAnimalSpecies = () => {
  return useQuery({
    queryKey: ["masters", "animal-species"],
    queryFn: getAnimalSpecies,
    staleTime: QUERY_STALE_TIMES.STATIC,
  });
};
```

**注意**: 現在 `owners/api/` に配置されている。master feature に移動すべき。

### マスタ設定画面の構造

```typescript
// frontend/src/features/master/ — 既存のマスタ設定パターンを確認する必要あり
// diagnosis, medicine, treatment-item 等の管理UIが参照実装
```

### 自動生成型

```typescript
// frontend/src/types/generated/models.ts:88-100
export interface AnimalSpecies {
  id: number /* uint64 */;
  name: string;
  is_active: boolean;
  sort_order: number /* int */;
  created_at: string;
  updated_at: string;
}
```

## 必要な変更

### 1. API hooks（master feature に配置）

```typescript
// frontend/src/features/master/api/animal-species/ に以下を新規作成:

// get-animal-species.ts — List + Get
// create-animal-species.ts — Create
// update-animal-species.ts — Update
// delete-animal-species.ts — Delete
// reorder-animal-species.ts — Reorder

// 既存の owners/api/get-animal-species.ts は master に移動
// PetEditModal からの参照は master feature の API を import に変更
// ※ feature間import禁止ルールに注意:
//   PetEditModal（owners feature）から master API を直接 import できない
//   → app/pages/ 経由で注入するか、共有 hooks に移動する
```

### 2. 型定義

```typescript
// frontend/src/features/master/api/animal-species/types.ts
// models.ts から導出
import type { AnimalSpecies as BackendAnimalSpecies } from "@/types/generated/models";

export type CreateAnimalSpeciesRequest = Pick<BackendAnimalSpecies, "name" | "is_active" | "sort_order">;
export type UpdateAnimalSpeciesRequest = Partial<CreateAnimalSpeciesRequest>;
export type ReorderAnimalSpeciesRequest = { ids: number[] };
```

### 3. マスタ設定画面にセクション追加

```typescript
// frontend/src/features/master/ の既存パターンに準拠
// テーブル表示: 名前 | 有効/無効 | 並べ替え | 編集 | 削除
// SidePeek パネルで編集（既存パターンに合わせる）
// ドラッグ&ドロップで並べ替え（Reorder API）
```

### 4. 既存の owners/api/get-animal-species.ts の移行

```typescript
// owners/api/get-animal-species.ts を削除
// PetEditModal の useGetAnimalSpecies の参照元を変更
// feature間import禁止ルールへの対応:
//   Option A: hooks/use-master-items.ts に統合（共有hooks）
//   Option B: app/pages/ で注入（依存逆転パターン）
```

## UI 操作フロー

1. ユーザーがマスタ設定画面を開く
2. サイドバーまたはタブから「動物種類」を選択
3. 登録済みの動物種類一覧がテーブル表示される（犬, 猫, 鳥, うさぎ, ハムスター, その他）
4. 「追加」ボタンで SidePeek が開き、名前を入力して保存
5. 行クリックで SidePeek が開き、名前・有効/無効を編集
6. ドラッグ&ドロップで表示順を変更
7. 削除ボタンで確認ダイアログ → ペットで使用中なら「使用中のため削除できません」エラー

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] feature間import禁止（owners → master 直接 import 不可）
- [ ] UI仕様は既存マスタ設定のパターンに準拠（Figma デザインなし）

## 依存関係

- **BE-040** が先に完了している必要がある（Create/Update/Delete/Reorder API が必要）
- `make codegen` は不要（model 変更なし）
- 既存のマスタ設定画面パターン（diagnosis_category 等）を参照実装とする

## 完了条件

- [ ] master feature に animal-species CRUD API hooks 追加
- [ ] マスタ設定画面に動物種類セクション追加
- [ ] 一覧表示（名前, 有効/無効, 並べ替えハンドル）
- [ ] 追加・編集（SidePeek パネル）
- [ ] 削除（使用中チェック付き確認ダイアログ）
- [ ] 並べ替え（ドラッグ&ドロップ → Reorder API）
- [ ] owners/api/get-animal-species.ts の移行（feature間import禁止対応）
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
