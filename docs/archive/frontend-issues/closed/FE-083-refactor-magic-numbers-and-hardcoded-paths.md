# FE-083: マジックナンバー・ハードコード URL の定数化

## 背景

全 `src/` フォルダの命名規則監査で検出された、マジックナンバーとハードコード URL パスの修正。
可読性と保守性の観点で定数に抽出する。

## 依存

- なし（単独で実施可能）

## 要件

### 1. `features/hospitalization/` — マジックナンバー定数化（3箇所）

**ファイル**: `hooks/use-hospitalization-form.ts`

`7 * 86400000`（7日間のミリ秒）が3箇所に重複している。モジュールレベル定数に抽出する。

```typescript
// ファイル先頭に追加
const MS_PER_DAY = 86_400_000;
const DEFAULT_HOSPITALIZATION_DAYS = 7;

// 行43, 106, 252 を置換
// before
new Date(Date.now() + 7 * 86400000)

// after
new Date(Date.now() + DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY)
```

### 2. `features/inventory/` — ハードコード URL パスの修正（1箇所）

**ファイル**: `routes/InventoryList.tsx:161`

```typescript
// before
navigate(`/inventory/${id}`);

// after
import { paths } from "@/config/paths";
navigate(paths.inventory.detail.getHref(id));
```

**補足**: 同ファイル157行目は正しく `paths.inventory.new.getHref()` を使っており、161行目だけが漏れている。

### 3. `features/vaccinations/` — マジック文字列の定数化（1箇所）

**ファイル**: `hooks/use-vaccination-form.ts:32`

```typescript
// before
const DEFAULT_FORM: VaccinationFormState = {
  // ...
  nextScheduleType: "4weeks",
  // ...
};

// after
const DEFAULT_NEXT_SCHEDULE_TYPE = "4weeks" as const;

const DEFAULT_FORM: VaccinationFormState = {
  // ...
  nextScheduleType: DEFAULT_NEXT_SCHEDULE_TYPE,
  // ...
};
```

## 受入条件

- [ ] `use-hospitalization-form.ts` の `7 * 86400000` が全3箇所で定数 `DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY` に置換されている
- [ ] `InventoryList.tsx` のハードコード URL が `paths.inventory.detail.getHref(id)` に変更されている
- [ ] `use-vaccination-form.ts` の `"4weeks"` が定数化されている
- [ ] `docker compose exec frontend pnpm build` が成功
