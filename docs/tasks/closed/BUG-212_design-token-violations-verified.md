# BUG-212: デザイントークン違反（スポットチェック確認済み）

| 項目 | 内容 |
|------|------|
| 優先度 | **Medium** |
| カテゴリ | デザイントークン |
| 注意 | スポットチェックで確認済みの箇所のみ記載。他にも同種の違反が存在する可能性あり |

## 確認済み違反

### hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx

| 行 | 現状 | 補足 |
|-----|------|------|
| 88 | `text-blue-500` | デザイントークン (`C.text40` 等) と混在使用 |
| 105 | `bg-gray-50` | 同行で `C.text40`, `C.borderMedium` は使用済み |

### accounting/routes/AccountingDetail.tsx

| 行 | 現状 |
|-----|------|
| 148 | `text-blue-500 bg-blue-50` |
| 387 | `bg-gray-50` |

### utils/constants/status-colors.ts（全体）

ファイル全体が Tailwind カラーリテラル。`bg-emerald-500`, `text-sky-700` 等が直接文字列で定義。
このファイルは複数の feature から参照されるため影響範囲が広い。

### components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:13-16

```typescript
import { useGetAllConsultations } from "@/features/master/api/consultations";
import { useGetAllProcedures } from "@/features/master/api/procedures";
```

`useGetAllConsultations` と `useGetAllProcedures` は `features/master/index.ts` に export されていない。
Deep Import 違反（Feature Indexing 違反）が確認済み。

## 未確認の箇所

以下は当初の監査で指摘されたが、スポットチェック未実施のため正確性は保証しない：
- hospitalization feature の他コンポーネント（50箇所+の指摘）
- medical-records feature（30箇所+の指摘）
- reservations / shifts feature のデザイントークン

これらを修正する場合は、対象ファイルを実際に読んで確認してから作業すること。
