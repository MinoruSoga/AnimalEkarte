# BUG-297: 大型コンポーネントのmemo()不足（rerender-memo）

## 概要

Vercel React Best Practices `rerender-memo` 規則に違反するコンポーネントが5件確認された。
コールバックpropsを受け取る大型コンポーネントが `memo()` でラップされていないため、親コンポーネントの状態変化のたびに不要な再レンダリングが発生している。

## 対象コンポーネント

| ファイル | 行数 | 主な問題 |
|---------|------|---------|
| `features/reservations/components/WeekView.tsx` | 569行 | drag-and-drop カレンダー。親のdateState変化で毎回再計算 |
| `features/accounting/components/AccountingDocument.tsx` | 194行 | 印刷ドキュメント。1200行の親(AccountingDetail)の状態変化で毎回再レンダリング |
| `features/auth/components/ChangePasswordDialog.tsx` | 129行 | Sidebarに常時mount。auth状態変化で毎回再レンダリング |
| `features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx` | 145行 | `onRegisterSave` コールバックprops。親の再レンダリングでuseEffect再実行リスク |
| `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 322行 | lazy()で遅延ロード済みだがdialog open中に親再レンダリングで再実行 |

**注記**: PetEditModal(632行) は `lazy()` + `Suspense` で既に最適化済み。
**注記**: WeekView の内部コンポーネント (TimeSidebar, AppointmentCard, DayColumn) は既に `memo()` 適用済み。外側のWeekViewも適用が必要。

## 修正パターン

```typescript
// Before
export function WeekView({ ... }: WeekViewProps) { ... }

// After
export const WeekView = memo(function WeekView({ ... }: WeekViewProps) { ... });
// + import に memo を追加
```

**重要**: memo()の効果を最大化するため、親コンポーネントがコールバックを`useCallback`で安定化していることを確認すること。

## ステータス

- [x] ドキュメント作成
- [x] 実装完了（5件すべてmemo()適用済み）
