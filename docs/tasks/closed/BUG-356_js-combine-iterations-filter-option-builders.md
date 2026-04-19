# BUG-356: js-combine-iterations — フィルタ選択肢ビルダーの多段イテレーション [CLOSED]

## 概要

リスト画面のフィルタ選択肢を動的生成する際、`.map().filter()` → `Set` → `.sort().map()` の 3-4 パスチェーンが 9 箇所に散在している。`js-combine-iterations` ルール（Vercel React Best Practices）に違反。

## 優先度

**LOW** — 全箇所 `useMemo` 内で実行されており、データセットもクリニック単位（<500件）のため実測パフォーマンス影響は無視できる。ただし系統的パターンのため、共通ユーティリティ抽出で可読性・保守性が改善される。

## 現在のコード（違反パターン）

```tsx
// 3-4パスのイテレーション
const doctorOptions = Array.from(
  new Set(allRecords.map((r) => r.doctor).filter(Boolean))
)
  .sort()
  .map((d) => ({ value: d, label: d }));
```

## 推奨修正

共通ユーティリティ関数を作成し、1パス + sort で済ませる:

```tsx
// lib/utils.ts or utils/array.ts
export function uniqueSortedOptions<T>(
  items: T[],
  accessor: (item: T) => string | null | undefined,
): { value: string; label: string }[] {
  const set = new Set<string>();
  for (const item of items) {
    const val = accessor(item);
    if (val) set.add(val);
  }
  return Array.from(set)
    .sort()
    .map((v) => ({ value: v, label: v }));
}

// 使用例
const doctorOptions = uniqueSortedOptions(allRecords, (r) => r.doctor);
```

## 対象ファイル（9箇所）

| ファイル | 行 | フィールド |
|---------|-----|----------|
| `features/medical-records/routes/MedicalRecords.tsx` | 102-104 | doctor |
| `features/medical-records/routes/MedicalRecords.tsx` | 105-107 | species |
| `features/examinations/routes/ExaminationsList.tsx` | 101-103 | testType |
| `features/examinations/routes/ExaminationsList.tsx` | 104-106 | doctor |
| `features/vaccinations/routes/VaccinationList.tsx` | 91-93 | doctor |
| `features/trimming/routes/TrimmingList.tsx` | 180-182 | species |
| `features/trimming/routes/TrimmingList.tsx` | 183-185 | staff |
| `features/hospitalization/routes/HospitalizationList.tsx` | 113-115 | species |
| `features/master/routes/ReservationTypeSettings.tsx` | 255 | filter+map |

## 検出方法

Vercel React Best Practices 深掘り監査（2026-04-14）— `js-combine-iterations` ルール検査

## タグ

- `performance`
- `js-combine-iterations`
- `refactor`
- `low-priority`
