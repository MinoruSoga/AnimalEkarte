# BUG-314: LAYOUT.modal トークンのギャップ — カスタムモーダルサイズが9箇所に散在

## ステータス: CLOSED — デザインレビュー待ちとして据え置き

## 概要
`LAYOUT.modal` に定義されていないカスタムモーダルサイズが複数の DialogContent に直接ハードコードされている。現在のトークン（sm=480px, md=512px, lg=768px, xl=1000px, full=1200px）のギャップを埋めるか、各カスタムサイズを意図的なものとして文書化する必要がある。

## 現状コード（未トークン化のモーダルサイズ一覧）

| ファイル | 行 | 現在のサイズ | 最近傍トークン | 差分 |
|---------|-----|-----------|--------------|------|
| `features/reservations/components/ReservationDetailModal.tsx` | 87 | `sm:max-w-[420px]` | modal.sm (480px) | −60px |
| `components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | 192 | `sm:max-w-[560px] max-h-[80vh]` | modal.md (512px) | +48px |
| `components/shared/MasterSelectModal/MasterSelectModal.tsx` | 59 | `sm:max-w-[500px]` | modal.md (512px) | −12px |
| `components/shared/OwnerSearchModal/OwnerSearchModal.tsx` | 115 | `sm:max-w-2xl max-h-[80vh]` | modal.lg (768px) | −96px（2xl=672px） |
| `features/accounting/routes/AccountingDetail.tsx` | 680 | `sm:max-w-sm` | modal.sm (480px) | −96px（sm=384px） |
| `features/accounting/routes/AccountingDetail.tsx` | 1164 | `max-w-4xl h-[90vh]` | modal.xl (1000px) | +114px（4xl=896px + h制約違い） |
| `features/auth/components/ChangePasswordDialog.tsx` | 64 | `sm:max-w-md` | modal.md (512px) | −64px（md=448px） |
| `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 176 | `sm:max-w-md` | modal.md (512px) | −64px（md=448px） |
| `features/medical-records/components/VitalsModal.tsx` | 16 | `max-w-4xl max-h-[85vh]` | modal.xl (1000px) | +114px（4xl=896px + h違い） |
| `features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` | 153 | `max-w-sm` | modal.sm (480px) | −96px（sm=384px） |
| `features/line-reservation/components/LinkedLineCustomers.tsx` | 169 | `max-w-md` | modal.md (512px) | −64px（md=448px） |

## 問題の本質

現在の LAYOUT.modal トークンは5段階（sm/md/lg/xl/full）だが、実際のUIには：
- **384px** (`sm:max-w-sm`): ChangePasswordDialog、AccountingDetail一部、DailyVitalsSection
- **448px** (`sm:max-w-md`): ShiftFormDialog、LinkedLineCustomers
- **420px** (`sm:max-w-[420px]`): ReservationDetailModal
- **500px** (`sm:max-w-[500px]`): MasterSelectModal
- **560px** (`sm:max-w-[560px]`): TreatmentSearchDialog
- **672px** (`sm:max-w-2xl`): OwnerSearchModal

これらをそのままトークン化するとサイズが乱立する。整理方針が必要。

## 修正方針（2択）

### 案A: 既存トークンに吸収（推奨）
各モーダルのサイズを最近傍の既存トークンに合わせてリサイズする。
- 例: `sm:max-w-[420px]` → `LAYOUT.modal.sm`（480px）に拡大
- 例: `sm:max-w-md` (448px) → `LAYOUT.modal.md`（512px）に拡大
- デザインレビューが必要

### 案B: 中間トークンを追加
よく使われるサイズを新規トークンとして追加する：
```ts
// frontend/src/lib/design-tokens.ts
modal: {
  xs:   "sm:max-w-sm",          // 384px — ← 追加
  sm:   "sm:max-w-[480px]",     // 480px — 既存
  md:   "sm:max-w-lg",          // 512px — 既存
  lg:   "sm:max-w-3xl",         // 768px — 既存
  xl:   "sm:max-w-[1000px] max-h-[90vh]",  // 既存
  full: "w-[98%] sm:max-w-[1200px] h-[90vh]",  // 既存
}
```
→ 560px や 420px など中途半端なサイズはそれぞれ最近傍に吸収する

### 前提条件
各モーダルのサイズは意図的に設計されている可能性があるため、**デザインレビューなしに一括変更するのは危険**。まずデザイン確認後、案A/Bを選択する。

## 影響範囲

| 対象 | ファイル | 行 |
|------|---------|-----|
| ReservationDetailModal | `features/reservations/components/ReservationDetailModal.tsx` | 87 |
| TreatmentSearchDialog | `components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | 192 |
| MasterSelectModal | `components/shared/MasterSelectModal/MasterSelectModal.tsx` | 59 |
| OwnerSearchModal | `components/shared/OwnerSearchModal/OwnerSearchModal.tsx` | 115 |
| AccountingDetail (払戻) | `features/accounting/routes/AccountingDetail.tsx` | 680 |
| AccountingDetail (大) | `features/accounting/routes/AccountingDetail.tsx` | 1164 |
| ChangePasswordDialog | `features/auth/components/ChangePasswordDialog.tsx` | 64 |
| ShiftFormDialog | `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 176 |
| VitalsModal | `features/medical-records/components/VitalsModal.tsx` | 16 |
| DailyVitalsSection | `features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` | 153 |
| LinkedLineCustomers | `features/line-reservation/components/LinkedLineCustomers.tsx` | 169 |
| トークン定義 | `frontend/src/lib/design-tokens.ts` | LAYOUT.modal セクション |

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`, `LAYOUT`) を使用する。
> **PROHIBITED**: Hexカラー・px サイズのハードコードは厳禁。

## 優先度
**Low** — 機能影響なし。デザインレビューが前提条件のため優先度は低い。トークン整備の観点では価値あり。

## 関連チケット
- なし（このスキャンで初めて全数把握）

## 備考
このチケットは **デザインレビュー待ち**。エンジニア単独では「どのサイズに吸収するか」を決定できない。Figmaデザインファイルを参照してモーダルサイズの設計意図を確認した上で実施すること。
