# FE-200: `&&` 条件レンダー規約違反（EstimateList・MedicalRecordExamination）

## 概要

`EstimateList.tsx` と `MedicalRecordExamination.tsx` で、プロジェクト規約で禁止されている
`&&` による条件レンダリングを使用している。`0` や空文字がレンダリングされるリスクがあるため、
必ず `condition ? <X /> : null` を使用すること。

## 現状コード

### `frontend/src/features/estimates/routes/EstimateList.tsx:272-280`
```tsx
{filtered.length > 0 && (
  <Pagination
    currentPage={pagination.currentPage}
    totalPages={pagination.totalPages}
    totalCount={pagination.totalCount}
    startIndex={pagination.startIndex}
    endIndex={pagination.endIndex}
    onPageChange={pagination.goToPage}
  />
)}
```

`filtered.length > 0` は boolean になるため実害は少ないが、規約違反。

### `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx:66-69`
```tsx
{!isLoading &&
  examGroups.map((group) => (
    <ExaminationGroup key={group.id} group={group} />
  ))}
```

`examGroups.map(...)` が空配列 `[]` を返す場合、`!isLoading && []` で `[]` がレンダリングされる可能性がある。

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/estimates/routes/EstimateList.tsx` | 272 | 要修正 |
| `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx` | 66-69 | 要修正 |

## 修正方針

### 1. `EstimateList.tsx:272`
```tsx
// Before
{filtered.length > 0 && (
  <Pagination ... />
)}

// After
{filtered.length > 0 ? (
  <Pagination ... />
) : null}
```

### 2. `MedicalRecordExamination.tsx:66-69`
```tsx
// Before
{!isLoading &&
  examGroups.map((group) => (
    <ExaminationGroup key={group.id} group={group} />
  ))}

// After
{!isLoading ? (
  examGroups.map((group) => (
    <ExaminationGroup key={group.id} group={group} />
  ))
) : null}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — 条件レンダー
> `&&` 禁止: `{isLoading && <div>Loading...</div>}`  
> 正しい: `{isLoading ? <div>Loading...</div> : null}`

### `.claude/rules/code-style.md` — Prohibited
> `&&` for conditional rendering（use `? (...) : null` instead）

## 優先度
**Low** — 現状のロジックでは実害がないが、規約統一のために修正すること。

## 関連ファイル
- `frontend/src/features/estimates/routes/EstimateList.tsx:272`
- `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx:66-69`
