# FE-174: 複数マスタ設定ページ — usePermission 完全欠落（ChiefComplaint・HospitalizationSettings・Insurance・InterviewTemplate）

## 概要

以下の 4 つのマスタ設定ページで `usePermission` が一切呼び出されていない。CRUD 操作（追加・編集・削除）に対する RBAC ガードが完全に欠如している。

## 影響範囲

| ファイル | マスタ設定 | 深刻度 |
|---------|-----------|--------|
| `ChiefComplaintSettings.tsx` | 主訴マスタ | HIGH |
| `HospitalizationSettings.tsx` | 入院設定マスタ（病室・種別等） | HIGH |
| `InsuranceSettings.tsx` | 保険マスタ | HIGH |
| `InterviewTemplateSettings.tsx` | 問診テンプレートマスタ | HIGH |

## 根本原因

```tsx
// ChiefComplaintSettings.tsx — usePermission なし ❌
export function ChiefComplaintSettings() {
  // usePermission(ResourceMasterMedical) が呼ばれていない
  // 追加・編集・削除ボタン: 権限チェックなし
}

// HospitalizationSettings.tsx — usePermission なし ❌
// InsuranceSettings.tsx — usePermission なし ❌
// InterviewTemplateSettings.tsx — usePermission なし ❌
```

これらのページは `CompanySettings.tsx`・`OccupationSettings.tsx` 等と同様のパターンで、RBAC 実装が抜け落ちている（FE-168 参照）。

## 各設定ページのリスク評価

| 設定 | リスク | 理由 |
|------|--------|------|
| ChiefComplaint（主訴） | HIGH | 診断・カルテ入力に使用されるマスタ。誤削除で診療フローに影響 |
| HospitalizationSettings（入院設定） | HIGH | 病室・入院種別マスタ。誤変更で入院患者管理に影響 |
| Insurance（保険） | HIGH | 保険割合等が正確でないと会計計算に誤りが生じる |
| InterviewTemplate（問診テンプレート） | HIGH | テンプレート削除で問診フォームが機能不全になる |

## 修正方針

FE-168 の修正方針と同様。各ページで適切なリソースを使って `usePermission` を呼び出す。

```tsx
// ChiefComplaintSettings.tsx（例）
const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);

// 追加ボタンをガード
{canCreate ? <AddButton onClick={handleAdd}>追加</AddButton> : null}

// SidePanel に readOnly/onDelete を正しく渡す
<MasterSidePanel
  readOnly={!canEdit}
  onDelete={item !== null && canDelete ? () => onDeleteRequest(item) : undefined}
>
```

各ページで使うべき resource は以下の通り（要確認）：
- `ChiefComplaintSettings` → `ResourceMasterMedical`
- `HospitalizationSettings` → `ResourceMasterHospitalization`
- `InsuranceSettings` → `ResourceMasterAccounting`（または専用リソース）
- `InterviewTemplateSettings` → `ResourceMasterMedical`

## 優先度

**HIGH** — いずれも実際のデータ変更操作をガードしていない。特に保険マスタと入院設定は会計・入院管理の基盤データ。

## 関連ファイル

- `frontend/src/features/master/routes/ChiefComplaintSettings.tsx`
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/InsuranceSettings.tsx`
- `frontend/src/features/master/routes/InterviewTemplateSettings.tsx`
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-168（CompanySettings 等の独自実装マスタ設定）、FE-173（StaffSettings）
