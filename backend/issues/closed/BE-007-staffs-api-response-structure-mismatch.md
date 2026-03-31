# BE-007: スタッフAPI - レスポンス構造の不一致

## 問題

医師選択モーダルでスタッフリストを読み込むと、以下のエラーが発生する：

```
TypeError: staffs.filter is not a function
```

---

## 根本原因

StaffSelectionModal.tsx はスタッフリストが配列（`BackendStaff[]`）であることを期待しているが、実際のAPIレスポンスはラップされた構造になっている可能性：

**期待される構造（現在の実装）:**
```typescript
// useGetStaffs では以下を想定
const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
// data が BackendStaff[] として返される
```

**実際のAPIレスポンス（推測）:**
```json
{
  "data": [
    { "id": "...", "name": "...", "staff_role": "...", "is_active": true },
    ...
  ]
}
```

---

## エラースタック

```
at StaffSelectionModal (StaffSelectionModal.tsx:26:103)
→ staffs.filter is not a function
```

**StaffSelectionModal.tsx:24**
```typescript
const { data: staffs = [], isLoading } = useGetStaffs();
// staffs が配列でない
```

---

## 修正対応

### API エンドポイント `/v1/masters/staffs` の確認

スタッフ一覧を返すAPIエンドポイントが以下のいずれかを確認：

1. **直接配列を返すべき**: 現在のフロント実装と一致
   ```go
   c.JSON(http.StatusOK, staffs)  // []Staff
   ```

2. **ラップされた構造を返している**: フロント実装と不一致
   ```go
   c.JSON(http.StatusOK, gin.H{"data": staffs})  // { "data": []Staff }
   ```

### 修正方法

**方法A: APIエンドポイント修正（推奨）**

`/v1/masters/staffs` が **直接配列を返す** ように修正。これが REST API の標準パターン。

```go
// 修正後: 直接配列をレスポンス
c.JSON(http.StatusOK, staffs)
```

**方法B: フロント側対応（代替案）**

APIレスポンス構造に合わせてフロント実装を修正：

```typescript
export function useGetStaffs() {
  return useQuery({
    queryKey: ["masters", "staffs"],
    queryFn: async () => {
      const { data } = await axios.get<{ data: BackendStaff[] }>("/v1/masters/staffs");
      return data.data;  // ここでアンラップ
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
  });
}
```

---

## テスト環境

- 記録 ID: 17
- テスト日時: 2026-03-16 12:50 JST

---

## 優先度

**🔴 HIGH** - ユーザー操作ブロック・医師選択機能が完全に使用不可

---

## ブロッカー

- `/v1/masters/staffs` APIエンドポイントの応答構造確認が必要
- 他のマスタAPIエンドポイント（診断、医薬品等）も同じ問題を持つ可能性があるため確認推奨
