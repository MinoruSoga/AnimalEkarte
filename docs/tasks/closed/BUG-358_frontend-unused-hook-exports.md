# BUG-358: 未使用 React Query hook の export 削除

## 概要

11件の React Query hook が export されているが、どのコンポーネント/ルートからも呼び出されていない。UI 未実装または raw 関数を直接使用しているケース。

## 優先度

**LOW** — tree-shaking でバンドルには含まれない。公開 API の肥大化防止のため export を閉じるのが望ましいが、将来的な UI 追加で使用される可能性がある。

## 対象（11件）

| ファイル | hook | 状態 |
|---------|------|------|
| `accounting/api/create-accounting.ts:18` | `useCreateAccounting` | UI 未実装 |
| `accounting/api/update-accounting.ts:20` | `useUpdateAccounting` | UI 未実装 |
| `hospitalization/api/daily-records.ts:93` | `useGetDailyRecords` | DailyRecordsTab が直接 API 呼び出し |
| `inventory/api/inventory.ts:79` | `useDeleteInventoryItem` | raw 関数を直接使用 |
| `line-reservation/api/update-line-reservation-setting.ts:17` | `useUpdateLineReservationSetting` | raw 関数を直接使用 |
| `master/api/cages.ts:67` | `useGetCageById` | 未使用 |
| `master/api/company.ts:83` | `useUpdateCompany` | raw 関数を直接使用 |
| `master/api/insurances.ts:121` | `useReorderInsurances` | 未使用 |
| `master/api/medicines.ts:38` | `useGetMedicineById` | 未使用 |
| `master/api/occupations.ts:115` | `useReorderOccupations` | 未使用 |
| `master/api/reservation-type-groups.ts:139` | `useReorderReservationTypeGroups` | index.ts re-export は BUG-357 で削除済み |

## 推奨対応

各 hook の `export` keyword を削除し、ファイル内部のみに閉じる。hook 関数自体は将来の UI 追加用に残存可。

## 検出方法

デッドコードスキャン（2026-04-14）— grep による全シンボル参照カウント

## タグ

- `dead-code`
- `low-priority`
