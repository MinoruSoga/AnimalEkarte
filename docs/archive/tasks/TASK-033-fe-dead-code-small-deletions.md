# TASK-033: FE 確定デッドコード削除（小規模）

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 高
**領域**: Frontend

---

## 概要

静的解析（ESLint + tsc --noUnusedLocals + import grep）で確定したフロントエンドの小規模デッドコードを削除する。
すべて削除または1行修正で完結する作業。

---

## 対象ファイル一覧

### 1. `src/lib/zod.ts` — 完全削除

Zod バリデーションヘルパー6関数を export しているが、プロジェクト全体で **一切 import されていない**。

```
requiredString / optionalString / requiredEmail
requiredPhone / optionalPhone / postalCode
```

**対応**: ファイルごと削除。

---

### 2. `src/features/accounting/index.ts` — 完全削除

```ts
export { Accounting, AccountingDetail, AccountingPetSelection } from "./routes";
```

- `Accounting` は存在しない（正しくは `AccountingList`）→ TypeScript エラー発生中
- ファイル自体がどこからも import されていない

**対応**: ファイルごと削除。

---

### 3. `src/features/master/types/index.ts` — 完全削除

```ts
// Master feature types
// Add feature-specific types here as needed
```

コメントのみのプレースホルダー。型定義なし・import もなし。

**対応**: ファイルごと削除。

---

### 4. `src/features/examinations/types/index.ts` — 完全削除

```ts
// Examinations feature types
// Add feature-specific types here as needed
```

同上。

**対応**: ファイルごと削除。

---

### 5. `src/stores/` — 空ディレクトリ削除

ファイルが存在しない空ディレクトリ。Sidebar の collapsed 状態は `Sidebar.tsx` 内の `useState` でローカル管理済みのためグローバルストア不要。

**対応**: `src/stores/` ディレクトリを削除（`.gitkeep` 等も含め）。

---

## 受入条件

- [ ] `src/lib/zod.ts` が削除されている
- [ ] `src/features/accounting/index.ts` が削除されている
- [ ] `src/features/master/types/index.ts` が削除されている
- [ ] `src/features/examinations/types/index.ts` が削除されている
- [ ] `src/stores/` ディレクトリが削除されている
- [ ] `docker compose exec frontend npm run lint` エラー 0 件
- [ ] `docker compose exec frontend npm run build` 成功

## 備考

- 削除後に他ファイルへの影響がないことをビルドで確認すること
- `zod` パッケージ自体（`package.json`）は他で直接 `import { z } from "zod"` として使われているため削除しない
