# FE-175: 入院日次記録サブセクション — 追加・保存ボタン権限ガード完全欠落（DailyVitals/CareLog/StaffNotes）

## 概要

入院詳細の「日次記録」タブを構成する 3 つのサブセクションコンポーネントで `usePermission` が一切呼び出されていない。各セクションの「追加」ボタン（SubmitButton）が `canCreate=false` / `canEdit=false` のユーザーにも常時表示される。

## 影響範囲

| ファイル | 問題 UI | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `DailyVitalsSection.tsx` | 「追加」ボタン (行 90-98) | POST `/v1/hospitalizations/*/vitals` | HIGH |
| `DailyCareLogsSection.tsx` | 「追加」ボタン (行 129-137) | POST `/v1/hospitalizations/*/care-logs` | HIGH |
| `DailyStaffNotesSection.tsx` | 「追加」ボタン (行 79-87) | POST `/v1/hospitalizations/*/staff-notes` | HIGH |

加えて `DailyRecordsTab.tsx` 自体も `usePermission` を持たず、「この日の記録を作成」ボタン（行 128-141）に権限チェックがない（FE-159 と連動）。

## 根本原因

```tsx
// DailyVitalsSection.tsx 行 90-98 — usePermission なし ❌
<SubmitButton size="sm">
  追加
</SubmitButton>

// DailyCareLogsSection.tsx 行 129-137 — usePermission なし ❌
<SubmitButton size="sm">
  追加
</SubmitButton>

// DailyStaffNotesSection.tsx 行 79-87 — usePermission なし ❌
<SubmitButton size="sm">
  追加
</SubmitButton>
```

`DailyRecordsTab.tsx` が usePermission を持たず、子コンポーネント（各セクション）に権限情報を props で渡していない。子コンポーネント自身も権限チェックを持たない。

## 影響

`canCreate=false` / `canEdit=false` のユーザーが入院詳細ページを開き日次記録タブを表示すると：
1. バイタル（体重・体温・心拍数等）の「追加」ボタンが表示 → POST → 403
2. ケアログ（投薬・処置等の記録）の「追加」ボタンが表示 → POST → 403
3. スタッフメモの「追加」ボタンが表示 → POST → 403

## 修正方針

### 方針 A: DailyRecordsTab で usePermission を取得し子に渡す

```tsx
// DailyRecordsTab.tsx
const { canCreate, canEdit } = usePermission("hospitalization");

// 「記録を作成」ボタンを canCreate でガード
{canCreate ? (
  <Button onClick={handleCreateDailyRecord}>この日の記録を作成</Button>
) : null}

// 子コンポーネントに canCreate/canEdit を渡す
<DailyVitalsSection canCreate={canCreate} canEdit={canEdit} />
<DailyCareLogsSection canCreate={canCreate} canEdit={canEdit} />
<DailyStaffNotesSection canCreate={canCreate} canEdit={canEdit} />
```

```tsx
// DailyVitalsSection.tsx
interface DailyVitalsSectionProps {
  canCreate?: boolean;
  canEdit?: boolean;
}
export function DailyVitalsSection({ canCreate = false, canEdit = false }: DailyVitalsSectionProps) {
  return (
    // ...
    {canCreate ? <SubmitButton>追加</SubmitButton> : null}
  );
}
```

### 方針 B: 各サブセクションで直接 usePermission を呼ぶ

```tsx
// DailyVitalsSection.tsx
const { canCreate, canEdit } = usePermission("hospitalization");
```

方針 A が推奨（usePermission の呼び出しを 1 箇所に集約できるため）。

## 優先度

**HIGH** — 入院患者のバイタル・ケアログ・スタッフメモは医療記録の中核データ。`canCreate=false` ユーザーが誤ってデータを追加しようとして 403 エラーが発生する。

## 関連ファイル

- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` (行 90-98)
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx` (行 129-137)
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyStaffNotesSection.tsx` (行 79-87)
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx` (行 128-141: 記録作成ボタン)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-159（DailyRecordsTab「記録を作成」ボタン未ガード）、FE-160（入院管理タブの削除ボタン未ガード）
