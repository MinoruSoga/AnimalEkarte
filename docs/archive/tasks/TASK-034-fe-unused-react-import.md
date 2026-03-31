# TASK-034: FE medical-records コンポーネントの未使用 React import 削除

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 高
**領域**: Frontend

---

## 概要

`tsc --noUnusedLocals` で検出された、`features/medical-records/components/` 配下4ファイルの未使用 `React` import を削除する。

React 19 では JSX Transform により `import React from "react"` は不要。

---

## 対象ファイル

| ファイル | エラー |
|---------|-------|
| `src/features/medical-records/components/MedicalRecordBillCheck.tsx:1` | `React` is declared but its value is never read |
| `src/features/medical-records/components/MedicalRecordEstimate.tsx:1` | `React` is declared but its value is never read |
| `src/features/medical-records/components/MedicalRecordExamination.tsx:2` | `React` is declared but its value is never read |
| `src/features/medical-records/components/MedicalRecordImage.tsx:2` | `React` is declared but its value is never read |

---

## 対応内容

各ファイルの先頭 import から `React` を削除する。

```diff
- import React, { lazy, memo, Suspense, useState, useMemo, useCallback } from "react";
+ import { lazy, memo, Suspense, useState, useMemo, useCallback } from "react";
```

（各ファイルで使っているフックに応じて named import のみ残す）

---

## 受入条件

- [ ] 4ファイルすべてから `React` の名前付き/デフォルト import が削除されている
- [ ] `docker compose exec frontend npm run build` 成功
- [ ] `docker compose exec frontend npm run lint` 警告 7 件以下（shadcn/ui 由来のみ）
