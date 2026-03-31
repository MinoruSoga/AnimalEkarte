# TASK-036: FE barrel index.ts 群の削除

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: 低
**領域**: Frontend

---

## 概要

プロジェクトルール「barrel index 経由 import 禁止・直接ファイル import」に従い、
各 feature の `api/index.ts`, `routes/index.ts`, `hooks/index.ts` が整備されているが、
**router.tsx・各コンポーネントはすべて直接ファイルをimportしており、これらのバレルは一切 import されていない**。

将来も使われる可能性がなく、むしろコーディングルール違反を誘発するリスクがあるため削除する。

---

## 削除対象ファイル（25ファイル）

```
features/accounting/api/index.ts
features/accounting/routes/index.ts
features/dashboard/api/index.ts
features/dashboard/components/index.ts
features/examinations/api/index.ts
features/examinations/hooks/index.ts
features/examinations/routes/index.ts
features/hospitalization/api/index.ts
features/hospitalization/hooks/index.ts
features/hospitalization/routes/index.ts
features/inventory/api/index.ts
features/inventory/routes/index.ts
features/master/api/index.ts
features/master/routes/index.ts
features/medical-records/api/index.ts
features/medical-records/components/index.ts
features/medical-records/components/BillingReviewSection/index.ts
features/medical-records/components/CheckupsTab/index.ts
features/medical-records/components/ClinicalPlanSection/index.ts
features/medical-records/components/TreatmentsTab/index.ts
features/medical-records/components/VitalsTab/index.ts
features/medical-records/hooks/index.ts
features/medical-records/routes/index.ts
features/owners/api/index.ts
features/owners/routes/index.ts
features/trimming/api/index.ts
features/trimming/hooks/index.ts
features/trimming/routes/index.ts
features/vaccinations/api/index.ts
features/vaccinations/hooks/index.ts
features/vaccinations/routes/index.ts
```

**除外（alive）**:
- `features/auth/index.ts` — `router.tsx:7` から `import { AuthProvider } from "@/features/auth"` で使用中
- `features/auth/types/index.ts` — `RequirePermission.tsx` から `import type { ResourceAction } from "@/features/auth/types"` で使用中
- `features/*/*/types/index.ts` — 型定義が実際に記述されているものは個別判断

---

## 注意事項

- 削除前に各ファイルが import されていないことを再確認すること（`grep -rn "from.*features/xxx/api\"" src/`）
- `types/index.ts` ファイルは内容が「コメントのみ」のものだけ削除（TASK-033 で対応済み）、実型定義があるものは残す

---

## 受入条件

- [ ] 対象ファイルがすべて削除されている
- [ ] `docker compose exec frontend npm run build` 成功
- [ ] `docker compose exec frontend npm run lint` エラー 0 件
