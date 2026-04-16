# BUG-179: AccountingDetail.tsx（1204行）が memo() 未適用でパフォーマンス問題

## 概要

`features/accounting/routes/AccountingDetail.tsx` は 1204 行の超大型コンポーネントであるが `memo()` が適用されていない。会計詳細画面は複数の `useState`、`useCallback`、複数の API クエリを持ち、親コンポーネントからの不要な再レンダーで全体が再計算される。プロジェクト規約では「独立した大きいセクションは `memo()` で囲む」ことが必須。

## 再現手順

1. 会計詳細画面を開く
2. React DevTools Profiler で記録開始
3. 親コンポーネントで何らかの状態変更（サイドバーの開閉等）を行う
4. **結果**: AccountingDetail.tsx 全体が再レンダーされる（1204行分の再計算が発生）

## 期待する動作

```tsx
// ✅ 大型コンポーネントは memo() でラップ
export const AccountingDetail = memo(function AccountingDetail() {
  // 1204行のコンポーネント
});
```

## 現状コード

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:1付近`
```tsx
// ❌ memo() なし
export function AccountingDetail() {
  // 複数の useState
  const [activeTab, setActiveTab] = useState("items");
  const [isEditing, setIsEditing] = useState(false);
  // 複数の API クエリ
  const { data: accounting } = useGetAccountingDetail(id);
  const { data: merchandiseItems } = useGetAllMerchandiseItems(clinicId);
  // ... 1200行続く
}
```

### 比較: 正しい実装（参照実装）
```tsx
// ✅ MedicalRecordForm.tsx:58
export const MedicalRecordForm = memo(function MedicalRecordForm() {
  // 556行
});

// ✅ TrimmingForm.tsx — 685行で内部サブコンポーネントを memo 分割
const LeftColumn = memo(function LeftColumn({ ... }) { ... });
const MiddleColumn = memo(function MiddleColumn({ ... }) { ... });
const RightColumn = memo(function RightColumn({ ... }) { ... });
```

## 影響範囲

| 対象ファイル | 行数 | memo 適用 | 状態 |
|---|---|---|---|
| `features/accounting/routes/AccountingDetail.tsx` | 1204行 | ❌ | 未修正 |
| `features/accounting/routes/AccountingList.tsx` | 406行 | ❌ | 未修正（同時対応推奨） |

## 修正方針

### 1. コンポーネント全体を memo() でラップ

```tsx
import { memo } from 'react';

// Before
export function AccountingDetail() { ... }

// After
export const AccountingDetail = memo(function AccountingDetail() { ... });
```

### 2. 内部を論理セクションに分割して各 memo 化（推奨）

1204 行は巨大すぎる。以下のセクションに分割して個別に memo 化することを推奨:

```tsx
// 請求明細セクション
const ItemListSection = memo(function ItemListSection({ items, onEdit }: Props) { ... });

// 割引・調整セクション
const DiscountSection = memo(function DiscountSection({ discounts, onAdd }: Props) { ... });

// 支払い合計セクション
const PaymentSummarySection = memo(function PaymentSummarySection({ total }: Props) { ... });
```

### 3. useCallback でハンドラ安定化（memo の前提条件）

```tsx
// memo 効果を出すために、渡すハンドラは useCallback で安定化
const handleTabChange = useCallback((tab: string) => {
  setActiveTab(tab);
}, []);

const handleAddItem = useCallback((item: BillingItem) => {
  setItems(prev => [...prev, item]);
}, []);
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Performance Rules
> `rerender-memo`: 独立した大きいセクションは `memo()` で囲む。必ず props ハンドラを `useCallback` で安定化すること。

### `.claude/rules/performance-rules.md`
> React memo() で不要再レンダー排除

### `.claude/CLAUDE.md` — Shared Component memo()
> `DataTable`, `NotionFilter`, `Pagination`, `SidePeekPanel` は `memo()` 適用済み。新規共有コンポーネントも同様に適用すること。

### プロジェクト内参照実装
- `features/medical-records/routes/MedicalRecordForm.tsx:58` — `memo()` 正しく適用（556行）
- `features/trimming/routes/TrimmingForm.tsx` — LeftColumn/MiddleColumn/RightColumn に分割してmemo適用

## 優先度
**Low** — 機能的な問題はなく、現時点ではパフォーマンス上の体感差も小さい。ただし 1204 行という規模は将来のメンテナンス性にも問題があるため、大幅な機能追加前に対応を推奨。

## 関連チケット
なし

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/features/accounting/routes/AccountingList.tsx`（同時対応推奨）
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` — 参照実装
