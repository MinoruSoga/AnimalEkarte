# src/hooks — Shared Global Hooks

## 責務

このディレクトリは**クロスカッティングな共有フック**のみ配置する。

- ✅ ここに置くもの: 複数 feature にまたがる共有フック（認証、権限、ページタイトル、モーダル状態など）
- ❌ ここに置かないもの: 特定 feature 専用のフック → `features/xxx/hooks/` に配置する

### 例外: フック本体ではない支持ファイル

`auth-context.ts`（`createContext` の定義のみ、フックではない）は例外として本ディレクトリに同居する。唯一の消費者 `use-auth.ts` と同じ場所に置くのが最も凝集的で、`features/auth` に置くと層逆転（features → hooks/lib への依存は許可・逆方向は禁止）になるため。命名規則(`use-kebab-case.ts`)の対象外（FE7-4(a)・2026-07-18）。

## 命名規則

```
use-kebab-case.ts   例: use-modal-state.ts, use-pagination.ts
```

テストファイル: `use-xxx.test.ts` を同ディレクトリに並置する。

## React フックルール (MANDATORY)

```typescript
// ✅ トップレベルでのみ呼び出す
function useMyHook() {
  const [state, setState] = useState(false);  // 常にトップレベル
}

// ❌ 条件分岐・ループ内での呼び出し禁止
if (condition) {
  const [state] = useState(false);
}
```

## 安定参照 (useCallback / useMemo)

コールバックを外部コンポーネントへ渡す場合は `useCallback` でラップする。
依存配列を省略しない。

```typescript
// ✅
const handleClose = useCallback(() => setOpen(false), []);

// ❌
const handleClose = () => setOpen(false);  // 毎レンダーで新しい参照
```

## Query Cache 共有パターン

複数 feature が同じエンティティを参照する場合、このディレクトリに shared hook を置いて query key を統一する。

```typescript
import { queryKeys } from "@/lib/query-keys";

// src/hooks/use-pet.ts — 18 feature から参照
export function useGetPet(petId: string) {
  return useQuery({
    queryKey: queryKeys.pets.detail(petId),  // features/pets と同じキーでキャッシュ共有
    ...
  });
}
```

フック一覧・参照元 feature 数は `ls` / `grep -rl` で再構築可能なため個別列挙しない（2026-07-21 doctor監査でderivableと判定・削除）。
