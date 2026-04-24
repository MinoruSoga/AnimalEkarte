# FE-089: 入院一覧が0件表示（APIレスポンス型不一致）

**Status**: Open
**Priority**: High
**Affects**: 入院管理一覧 (`/hospitalization`)
**Date Created**: 2026-03-21
**Related**: BUG-005

## Summary

`features/hospitalization/api/get-hospitalizations.ts` がバックエンドのページネーション形式レスポンス `{ data: [...], total: N }` を配列 `BackendHospitalization[]` として型付けしているため、`data.map()` が `TypeError: data.map is not a function` を throw し、`use-hospitalizations.ts` の catch ブロックで silently 処理されて `hospitalizations` が空配列のままになる。

## 現状のコード

```typescript
// frontend/src/features/hospitalization/api/get-hospitalizations.ts:7-11
export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<BackendHospitalization[]>(
    "/v1/hospitalizations"
  );
  return data.map(transformHospitalization);  // ← TypeError: data.map is not a function
};
```

バックエンドの実際のレスポンス形式（`hospitalization_handler.go:54`）:
```go
c.JSON(http.StatusOK, newPaginatedResponse(hospitalizations, total, page, limit))
// → { "data": [...], "total": 7, "page": 1, "limit": 20 }
```

```typescript
// frontend/src/features/hospitalization/hooks/use-hospitalizations.ts:16-23
const loadData = useCallback(async () => {
    try {
        const data = await getHospitalizations();
        setHospitalizations(data);
    } catch (error) {
        handleApiError(error, "取得");
        // ← catch されて setHospitalizations([]) のまま
    }
}, []);
```

## 必要な変更

### 1. get-hospitalizations.ts でページネーションレスポンスを正しく処理

```typescript
// frontend/src/features/hospitalization/api/get-hospitalizations.ts

import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

// ページネーションレスポンス型を追加
interface HospitalizationPaginatedResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>(
    "/v1/hospitalizations"
  );
  return data.data.map(transformHospitalization);  // data.data で配列を取得
};
```

### 2. useGetHospitalizations hook も同様に修正（使用箇所があれば）

`use-hospitalizations.ts` は `getHospitalizations()` を直接呼んでいるため、Step 1 の修正のみで対応可能。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] 型は `models.ts` から導出

## 依存関係

なし（FE のみ）

## 完了条件

- [ ] 入院管理一覧で7件（または登録件数）が表示される
- [ ] 「入院中」タブ: admitted ステータスのみ表示
- [ ] 「予約」タブ: reserved ステータスのみ表示
- [ ] 「退院済」タブ: discharged ステータスのみ表示
- [ ] 「すべて」タブ: 全件表示
- [ ] `pnpm build` が通る
