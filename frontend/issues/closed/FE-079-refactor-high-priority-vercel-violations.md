# FE-079: Vercel React Best Practices 違反修正 (High Priority)

## 背景

全 `src/` フォルダの Vercel React Best Practices 監査を実施し、即時修正が必要な違反を7件検出した。
tree-shaking 阻害・本番 console 汚染・不要な再レンダーが含まれる。

## 依存

- なし（単独で実施可能）

## 要件

### 1. `features/master/` — console 汚染削除 (3件)

**ファイル**: `hooks/use-service-type-color-map.ts`

| 行 | 違反ルール | 修正内容 |
|----|-----------|---------|
| 68 | `console.warn` 禁止 | 行を削除 |
| 76 | `console.debug` 禁止 | 行を削除 |
| 83 | `console.log` 禁止 | 行を削除 |

### 2. `features/hospitalization/` — barrel import 修正 (3件)

**ファイル**: `hooks/use-hospitalizations.ts`, `hooks/use-hospitalization-detail.ts`, `hooks/use-hospitalization-form.ts`

| 違反ルール | 修正内容 |
|-----------|---------|
| `bundle-barrel-imports` | `from "../api"` を直接ファイル import に変更 |

```typescript
// before
import { getHospitalizations, updateHospitalization } from "../api";

// after
import { getHospitalizations } from "../api/get-hospitalizations";
import { updateHospitalization } from "../api/update-hospitalization";
```

各ファイルで `../api` から import している全シンボルを特定し、対応する個別ファイルからの import に書き換える。

### 3. `features/master/` — useCallback deps からオブジェクト除外 (1件)

**ファイル**: `hooks/use-master-save.ts:69`

| 違反ルール | 修正内容 |
|-----------|---------|
| `rerender-dependencies` | `crud.editTarget` をプリミティブに分解 |

```typescript
// before
const handleSave = useCallback((...) => {
  // ...
}, [crud.editTarget, crud.handleClose, crud.startSaveTransition, ...]);

// after
const editTargetId = crud.editTarget && typeof crud.editTarget === 'object' ? crud.editTarget.id : null;
const isNewMode = crud.editTarget === "new";
const handleSave = useCallback((...) => {
  // ...
}, [editTargetId, isNewMode, crud.handleClose, crud.startSaveTransition, ...]);
```

## 受入条件

- [ ] `use-service-type-color-map.ts` から console.warn / console.debug / console.log が全削除されている
- [ ] `hospitalization/hooks/` の 3 ファイルが直接ファイル import に変更されている
- [ ] `use-master-save.ts` の useCallback deps にオブジェクトが含まれていない
- [ ] `docker compose exec frontend npm run lint` がエラー 0
- [ ] `docker compose exec frontend npm run build` が成功
