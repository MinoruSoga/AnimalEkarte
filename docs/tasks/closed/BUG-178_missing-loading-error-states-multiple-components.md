# BUG-178: ローディング・エラー状態未処理（AccountingDetail・TrimmingForm・ClinicalPlanSection・MedicalRecordVaccination）

## 概要

4 つのコンポーネント/ルートで `useQuery` 系フックを使いながら `isLoading` / `isError` チェックを行っていない。データ取得中に空データが表示されたり、API エラー時に何も起きずに静かに失敗するサイレント障害が発生している。

## 再現手順

1. ネットワーク速度を低速に制限（DevTools → Network → Slow 3G）
2. 以下の画面を開く
3. **結果**: ローディングスピナーが表示されず、空のテーブルや空のフォームが一瞬表示されてからデータが補充される

## 期待する動作

- データ取得中は `LoadingFallback` またはスケルトン UI を表示する
- エラー時は `ErrorFallback` またはエラーメッセージを表示する

## 現状コード

### `frontend/src/features/accounting/routes/AccountingDetail.tsx`
```tsx
// ❌ isLoading / isError チェックなし（3つのクエリ）
const { data: merchandiseItems = [] } = useGetAllMerchandiseItems(clinicId);
const { data: refunds = [] } = useGetRefunds(accountingId);
const { data: pet } = useGetPet(petId);
// → ローディング中は空配列で表示、エラー時はサイレント失敗
```

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:464`
```tsx
// ❌ isLoading / isError チェックなし
const { data: petTrimmings = [] } = useGetTrimmingsByPetId(selectedPet?.id ?? "");
// → ペット選択後、施術履歴が空で表示されることがある
```

### `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx`
```tsx
// ❌ isLoading / isError チェックなし
const { data: clinicalPlan } = useGetClinicalPlan(medicalRecordId);
// → データ取得前に空のフォームが表示される
```

### `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx`
```tsx
// ❌ isLoading チェックなし（2つのクエリ）
const { data: petVaccinations = [] } = useGetPetVaccinations(petId);
const { data: vaccinesMaster = [] } = useGetAllVaccinesMaster(clinicId);
// → ワクチン一覧が空で表示されてからデータが補充される
```

### 比較: 正しい実装（EstimateDetail.tsx）
```tsx
// ✅ 正しい実装
const { data: estimate, isLoading, isError } = useGetEstimate(id);
if (isLoading) return <LoadingFallback />;
if (isError || !estimate) return <ErrorFallback />;
```

## 影響範囲

| 対象ファイル | 未処理のクエリ数 | 影響 | 状態 |
|---|---|---|---|
| `features/accounting/routes/AccountingDetail.tsx` | 3クエリ | 会計詳細画面の空表示 | 未修正 |
| `features/trimming/routes/TrimmingForm.tsx:464` | 1クエリ | 施術履歴の一時的な空表示 | 未修正 |
| `features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx` | 1クエリ | 診療計画の空フォーム表示 | 未修正 |
| `features/medical-records/components/MedicalRecordVaccination.tsx` | 2クエリ | ワクチン一覧の空表示 | 未修正 |

## 修正方針

### 1. `AccountingDetail.tsx` — ページレベルで isLoading/isError 追加
```tsx
const { data: accounting, isLoading, isError } = useGetAccountingDetail(id);
const { data: merchandiseItems = [] } = useGetAllMerchandiseItems(clinicId);
const { data: refunds = [] } = useGetRefunds(accountingId);

if (isLoading) return <LoadingFallback />;
if (isError || !accounting) return <ErrorFallback message="会計データの取得に失敗しました" />;
```

### 2. `TrimmingForm.tsx:464` — isLoading 追加
```tsx
const { data: petTrimmings = [], isLoading: isTrimmingsLoading } =
  useGetTrimmingsByPetId(selectedPet?.id ?? "");

// 施術履歴テーブル表示時に判定
{isTrimmingsLoading ? <LoadingSpinner /> : (
  <PetTrimmingHistory trimmings={petTrimmings} />
)}
```

### 3. `ClinicalPlanSection.tsx` — isLoading/isError 追加
```tsx
const { data: clinicalPlan, isLoading, isError } = useGetClinicalPlan(medicalRecordId);

if (isLoading) return <div>読み込み中...</div>;
if (isError) return <div>診療計画の取得に失敗しました</div>;
```

### 4. `MedicalRecordVaccination.tsx` — isLoading 追加
```tsx
const { data: petVaccinations = [], isLoading: isVaccinationsLoading } =
  useGetPetVaccinations(petId);
const { data: vaccinesMaster = [], isLoading: isMasterLoading } =
  useGetAllVaccinesMaster(clinicId);

const isLoading = isVaccinationsLoading || isMasterLoading;
if (isLoading) return <LoadingFallback />;
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
エラーハンドリング: `catch` ブロックでは必ず `handleApiError` を呼び出す。
コンポーネントのローディング状態は適切に管理すること。

### プロジェクト内参照実装
- `features/estimates/routes/EstimateDetail.tsx` — `isLoading`, `isError` を含む完全な実装
- `features/owners/routes/OwnersList.tsx` — エラー/ローディング状態の正しい処理

## 優先度
**Medium** — 機能的な問題よりも UX の問題。ただし `AccountingDetail` は業務の中核画面であり、エラー時のサイレント失敗は現場に混乱をもたらす可能性がある。

## 関連チケット
- BUG-163: MedicalRecordForm の null return（同カテゴリ）
- BUG-165: リスト系ページのローディング/エラー状態欠如

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx`
