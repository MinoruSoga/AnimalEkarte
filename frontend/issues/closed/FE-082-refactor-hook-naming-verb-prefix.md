# FE-082: フィルタリング Hook の動詞なし命名を修正

## 背景

全 `src/` フォルダの命名規則監査を実施した結果、4つのフィルタリング Wrapper Hook が動詞なし命名（`useXxx`）になっていることを検出。
プロジェクト規約では API Hook は `useGet` + Entity、フィルタリング Hook は `useFilter` + Entity を求めている。

## 依存

- なし（単独で実施可能）

## 要件

### 1. Hook 関数名のリネーム（4件）

各ファイルの関数名を変更し、全 import 箇所も更新する。

| ファイル | 現在の関数名 | 修正後 |
|---------|------------|--------|
| `features/trimming/hooks/use-trimming-records.ts` | `useTrimmingRecords` | `useFilterTrimmingRecords` |
| `features/vaccinations/hooks/use-vaccinations.ts` | `useVaccinations` | `useFilterVaccinations` |
| `features/examinations/hooks/use-examination-records.ts` | `useExaminationRecords` | `useFilterExaminationRecords` |
| `features/medical-records/hooks/use-medical-records.ts` | `useMedicalRecords` | `useFilterMedicalRecords` |

**理由**: これらは `useGetXxx()` API Hook のラッパーで、検索・フィルタリングロジックを追加している。`useGet` と紛らわしいため `useFilter` 接頭辞で意図を明確にする。

### 2. 曖昧な `data` 変数名の修正（3件）

各 Hook 内の `useQuery` 戻り値の `data` を、意味のある名前に変更する。

| ファイル | 行 | 現在 | 修正後 |
|---------|-----|------|--------|
| `features/trimming/hooks/use-trimming-records.ts` | 11 | `const { data = [] } = useGetTrimmings()` | `const { data: trimmingRecords = [] } = useGetTrimmings()` |
| `features/vaccinations/hooks/use-vaccinations.ts` | 9 | `const { data = [] } = useGetVaccinations(filters)` | `const { data: vaccinationsData = [] } = useGetVaccinations(filters)` |
| `features/examinations/hooks/use-examination-records.ts` | 9 | `const { data = [] } = useGetExaminations(filters)` | `const { data: examinationsData = [] } = useGetExaminations(filters)` |

### 3. hooks/index.ts の export 更新

各 feature の `hooks/index.ts` が存在する場合、export 名も更新する。

```typescript
// before
export { useTrimmingRecords } from "./use-trimming-records";

// after
export { useFilterTrimmingRecords } from "./use-trimming-records";
```

### 4. 呼び出し元の import 更新

各 Hook を使用しているコンポーネント（routes/ 内のファイル）の import を全て更新する。
IDE のリネーム機能を使えば安全に一括変更できる。

## 受入条件

- [ ] 4つの Hook 関数名が `useFilter` + Entity に変更されている
- [ ] 3つの `data` 変数が具体的な名前に変更されている
- [ ] 全 import 箇所が更新されている
- [ ] `docker compose exec frontend npm run build` が成功
- [ ] `docker compose exec frontend npm run lint` がエラー 0
