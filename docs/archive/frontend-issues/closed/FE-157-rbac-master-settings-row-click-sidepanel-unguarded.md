# FE-157: マスタ設定ページ — 全マスタページで行クリックが canEdit 無関係に SidePeek を開く + XxxSidePanel が readOnly を無視して「保存」ボタンを表示（システム的欠如）

## 概要

`/settings/*` 配下のマスタ設定ページ（MedicineSettings を除く全ページ）で、以下の 2 つの問題が複合して発生している：

1. `DataTableRow` の `onClick` に `canEdit` チェックがなく、閲覧のみユーザーでも行クリックで SidePeek が開く
2. 各 `XxxSidePanel` コンポーネントが `readOnly` prop を定義・使用していないため、MasterCRUDPage から `readOnly={!canEdit}` を渡しても無視され、「保存」ボタンが表示される

## 影響範囲（全ページ）

| ファイル | 行クリック | SidePanel 保存ボタン | 深刻度 |
|---------|-----------|---------------------|--------|
| `ChiefComplaintSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `HospitalizationSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `InsuranceSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `StaffSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `InterviewTemplateSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `DiagnosisSettings.tsx` | `onClick={() => onEditTargetChange(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `TrimmingSettings.tsx` | `onClick={() => onEditTargetChange(item)}` ❌ | readOnly 無視 ❌ | HIGH |
| `OccupationSettings.tsx` | `onClick={() => onEdit(item)}` ❌ | readOnly 無視 ❌ | HIGH |

**正しい実装（参照）**: `MedicineSettings.tsx:133`
```tsx
<SortableDataTableRow id={medicine.id} onClick={canEdit ? () => onEdit(medicine) : undefined}>
```

## 根本原因

### 問題 1: DataTableRow の onClick に canEdit チェックなし

```tsx
// ❌ 全マスタページ（MedicineSettings 以外）の共通パターン
<DataTableRow key={item.id} onClick={() => onEdit(item)}>
```

MasterCRUDPage の `renderRow` は `canEdit` を引数として受け取るが、各ページの実装では使用していない。

### 問題 2: XxxSidePanel が readOnly prop を型定義に含めていない

```tsx
// ❌ OccupationSettings.tsx の例
const OccupationSidePanel = memo(function OccupationSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: {
  item: Occupation | null;
  onClose: () => void;
  onSave: (d: OccupationFormData) => void;
  // readOnly が定義されていない ❌
}) {
  // MasterCRUDPage から {...props} で readOnly が渡されても
  // 型に定義がないため実際の「保存」ボタン制御に使われていない
```

MasterCRUDPage は `readOnly: !canEdit` を `renderSidePanel` に渡しているが、各 XxxSidePanel がこれを受け取らないため `SidePeekFooter` に `readOnly` が伝わらず、「保存」ボタンが常に表示される。

## 現状の挙動（バグ）

閲覧のみユーザー（canEdit=false）が `/settings/occupations` 等を開くと：
1. 行をクリック → SidePeek が開く（開くべきでない）
2. SidePeek 内に「保存」ボタンが表示される（非表示にすべき）
3. 「保存」ボタンをクリック → API PATCH が発火し 403

## 修正方針

### 方針 A: 各マスタページの DataTableRow onClick を修正

```tsx
// 各マスタページで onClick に canEdit チェックを追加
<DataTableRow
  key={item.id}
  onClick={canEdit ? () => onEdit(item) : undefined}
>
```

`MasterCRUDPage.renderRow` から `canEdit` が引数として渡されているため、これを利用する。

### 方針 B: 各 XxxSidePanel に readOnly prop を追加

```tsx
// OccupationSidePanel の例
const OccupationSidePanel = memo(function OccupationSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,  // ← 追加
}: {
  item: Occupation | null;
  onClose: () => void;
  onSave: (d: OccupationFormData) => void;
  readOnly?: boolean;  // ← 追加
}) {
  return (
    <MasterSidePanel
      ...
      readOnly={readOnly}  // ← SidePeekFooter に伝播
    />
  );
});
```

両方の修正が必要。方針 A のみでは「開けないが保存できる」バグが残る。方針 B のみでは「保存できないが開ける」UX 問題が残る。

## 優先度

**HIGH** — 閲覧のみユーザーがすべてのマスタデータ（職種・入院種別・主訴・保険・スタッフ等）の編集フォームを開け、「保存」ボタンをクリックした際に 403 エラーが発生する。実際のデータ変更は 403 で防がれるが、UI が誤った操作可能感を与える。

## 関連ファイル

- `frontend/src/features/master/components/MasterCRUDPage.tsx`
- `frontend/src/features/master/routes/OccupationSettings.tsx`
- `frontend/src/features/master/routes/ChiefComplaintSettings.tsx`
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/InsuranceSettings.tsx`
- `frontend/src/features/master/routes/StaffSettings.tsx`
- `frontend/src/features/master/routes/InterviewTemplateSettings.tsx`
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`
- `frontend/src/features/master/routes/TrimmingSettings.tsx`
- 発見日: 2026-04-07（RBAC Phase 2 テスト中）
- 参照正実装: `frontend/src/features/master/routes/MedicineSettings.tsx:133`
