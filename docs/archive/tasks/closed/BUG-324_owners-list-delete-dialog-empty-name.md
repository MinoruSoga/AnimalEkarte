# BUG-324: 飼主一覧の削除確認ダイアログで飼主名が空になる

## 概要
飼主一覧 (`/owners`) の操作メニュー「削除」をクリックすると、確認ダイアログに `飼主「」とこの飼主に関連するすべてのペット情報が削除されます。` と表示され、飼主名が空文字になる。

根本原因は API レスポンスが `owner_name` キーを返すのに、フロントエンドが `models.ts` の `Owner.name` を参照しているため、型は合致するが実行時に `undefined` になり `""` にフォールバックする。

## 再現手順

1. `admin@example.com / password` でログイン
2. `/owners` に移動（飼主一覧）
3. 任意の行の操作メニュー（…）をクリック
4. 「削除」をクリック
5. **結果**: `飼主「」とこの飼主に関連する…` と表示（飼主名が空）

## 期待する動作
- 削除確認ダイアログに `飼主「田中 花子」とこの飼主に関連する…` のように実際の飼主名が表示される

## 現状コード

### `frontend/src/features/owners/routes/OwnersList.tsx:383`
```typescript
onClick: () => handleDeleteRequest(pet.ownerId, pet.ownerName),
```

### `frontend/src/lib/transforms/pet.ts:61`
```typescript
ownerName: p.owner?.name ?? "",
```

`p.owner?.name` は `undefined` になる。実際の API レスポンスは `owner_name` フィールドを返すが、`models.ts` 由来の `Owner` 型は `name` を期待しているため。

### `frontend/src/features/owners/loaders.ts:48`
```typescript
return pets.map(pet => transformBackendPetToFrontend({ ...pet, owner }));
```

`owner` は `BackendOwner` (= `models.ts Owner`) として型付けされているが、実際の JSON は `{ owner_name: "田中 花子", ... }` 形式で来る。`owner.name` = `undefined`。

### 比較: 正しい実装（emptyPet 行、同ファイル:56）
```typescript
ownerName: owner.name,  // ← これも同じ理由で undefined になる
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/owners/loaders.ts:46-58` | `owner.name` 参照 → 全行で ownerName が `""` | NG |
| `frontend/src/lib/transforms/pet.ts:61` | `p.owner?.name` 参照 → ownerName が `""` | NG |
| `frontend/src/features/owners/routes/OwnersList.tsx:332` | `{pet.ownerName}` テーブル表示 → 飼主名列が空 | NG（要確認） |
| `frontend/src/features/owners/routes/OwnersList.tsx:383` | `handleDeleteRequest` への `pet.ownerName` 渡し → 削除ダイアログ空 | NG（確認済み） |

## 修正方針

**バックエンド API レスポンス (`ownerResponse`) は `owner_name` を返す（`owner_response.go:55`）。**
`models.ts` の `Owner.name` と乖離している。フロントエンドの loader で正しくマッピングする。

### 1. `frontend/src/features/owners/loaders.ts`

`BackendOwner` を使う際、API レスポンスフィールドを正規化してから transform に渡す。

```typescript
// 修正前（line 48）
return pets.map(pet => transformBackendPetToFrontend({ ...pet, owner }));

// 修正後: API response の owner_name を name に正規化
const normalizedOwner = {
  ...owner,
  name: (owner as Record<string, unknown>).owner_name as string ?? owner.name ?? "",
  name_kana: (owner as Record<string, unknown>).owner_name_kana as string ?? owner.name_kana ?? "",
};
return pets.map(pet => transformBackendPetToFrontend({ ...pet, owner: normalizedOwner }));
```

同様に emptyPet 行（line 56）も修正：

```typescript
// 修正前（line 56）
ownerName: owner.name,

// 修正後
ownerName: (owner as Record<string, unknown>).owner_name as string ?? owner.name ?? "",
ownerNameKana: (owner as Record<string, unknown>).owner_name_kana as string ?? owner.name_kana ?? undefined,
```

### 2. OwnerApiResponse 型を追加（推奨: 型安全化）

```typescript
// loaders.ts に追加
interface OwnerApiListItem extends Omit<BackendOwner, "name" | "name_kana"> {
  owner_name: string;
  owner_name_kana: string;
}

interface OwnersResponse {
  data: OwnerApiListItem[];
  total: number;
  page: number;
  limit: number;
}
```

そして transform 呼び出しを:
```typescript
const normalizedOwner: BackendOwner = {
  ...owner,
  name: owner.owner_name,
  name_kana: owner.owner_name_kana,
};
return pets.map(pet => transformBackendPetToFrontend({ ...pet, owner: normalizedOwner }));
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> 「型安全性最優先: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。」

API レスポンス型は実際の JSON と一致させるべき。`models.ts` の GORM モデル型をそのまま API レスポンス型として使うのは誤り（JSON タグが異なる）。

### `backend/internal/handler/owner_response.go:55`
```go
OwnerName string `json:"owner_name"`
```
バックエンドは明示的に `owner_name` を使う設計。フロントエンドもこれに合わせる必要がある。

## 優先度
**High** — 飼主名列が全行空になるUXバグ。飼主名ソートも無効化される。削除ダイアログで誰を削除するか判別不能。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/owners/loaders.ts:44-83` — ownersLoader、owner正規化処理
- `frontend/src/lib/transforms/pet.ts:57-89` — transformBackendPetToFrontend
- `backend/internal/handler/owner_response.go:53-75` — ownerResponse DTO（JSON `owner_name` を返す）
- `frontend/src/types/generated/models.ts:1114-1140` — `Owner` 型（`name` フィールド）
