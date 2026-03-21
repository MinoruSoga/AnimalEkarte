---
status: closed
closed_at: 2026-03-16
---

# [master] API stub 関数・dead code の削除

## 優先度
低

## 種別
コード品質 / 保守性

## 対象ファイル
- `frontend/src/features/master/api/get-master-items.ts`
- `frontend/src/features/master/api/get-master-item.ts`
- `frontend/src/features/master/api/transforms.ts`
- `frontend/src/features/master/api/create-master-item.ts`
- `frontend/src/features/master/api/update-master-item.ts`
- `frontend/src/features/master/api/index.ts`

## 問題一覧

### 1. `transforms.ts` が実質空ファイル

`features/master/api/transforms.ts` にはコメントのみ記述されており、実装がない。
transform 関数は各 API ファイル内（`diagnosis.ts` 等）に直接書かれている。
空のファイルは削除するか、実際の transform 関数を集約すること。

### 2. `getMasterItems` / `useGetMasterItems` が常に空配列を返す stub

```ts
// get-master-items.ts
export async function getMasterItems(): Promise<MasterItem[]> {
  return [];  // stub - 常に空配列
}
```

この関数を呼び出しているコンポーネントが存在する場合、静かに空データを返すバグになっている。
呼び出し元を確認し、stub を削除するか実装を移行すること。

### 3. `getMasterItem` / `useGetMasterItem` が常に null を返す

同様に常に `null` を返す stub。削除すること。

### 4. `create-master-item.ts` / `update-master-item.ts` の `inquiry_template` 分岐（dead code）

`MASTER_CATEGORY_ENDPOINT` に `inquiry_template` キーが存在しないため、分岐に到達したら
`throw new Error` になる到達不能分岐。削除すること。

### 5. `index.ts` の `useGetMasterItemsByCategoryNew` alias

```ts
// index.ts
export { useGetMasterItemsByCategory as useGetMasterItemsByCategoryNew } from "./get-master-items";
```

`New` サフィックスは移行期の一時的な命名の名残。移行完了後は alias を削除し、
`useGetMasterItemsByCategory` のみを export すること。

## 修正方針

各 stub 関数について、呼び出し元が存在するか `Grep` で確認してから削除する。
呼び出し元が存在する場合は、呼び出し元を適切な実装済み hook に変更する。
