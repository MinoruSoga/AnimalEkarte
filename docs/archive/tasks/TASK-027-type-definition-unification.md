# TASK-027: 型定義統一リファクタリング

**作成日**: 2026-03-25
**ステータス**: Open
**依頼元**: 型定義の重複・命名衝突を調査した結果、修正が必要な箇所が判明したため

---

## 概要

`src/types/` と `status-colors.ts` に型定義の重複・命名衝突が存在する。
`ReservationStatus` / `VisitType` / `TrimmingRecord` の3点を修正し、
型の単一定義源を確立する。

## 依頼内容（原文）

> 型定義統一リファクタリング: ReservationStatus/VisitType の重複定義解消・src/types/index.ts の旧型削除・TrimmingRecord 重複排除

## 仕様確認ログ

確認事項なし（コードベース調査により全仕様が確定）

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | ReservationStatus を proper union に修正 + status-colors.ts の重複削除 | FE | FE-115 | - | [x] |
| 2 | status-colors.ts の未使用 VisitType エクスポートを削除 | FE | FE-116 | - | [x] |
| 3 | src/types/index.ts の TrimmingRecord を TrimmingUI にリネーム | FE | FE-117 | - | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: `ReservationStatus` の定義が `src/types/index.ts` のみ（proper union `"confirmed"|"pending"|...`）になっており、`status-colors.ts` に重複定義がない
- [x] AC-2: `VisitType` の定義が `src/types/index.ts` の `"first"|"revisit"` のみ（`status-colors.ts` に `export type VisitType` が存在しない）
- [x] AC-3: `src/types/index.ts` に `TrimmingRecord` interface が存在せず（削除またはリネーム）、`trimming` feature が `TrimmingUI` を使用している
- [x] AC-4: `npm run build` 型エラーゼロ
- [x] AC-5: `npm run lint` エラーゼロ

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| ReservationStatus の定義方法 | `src/types/index.ts` の現状維持（変更不要） | `ReservationAppointment.status` が手書き proper union のため `ReservationAppointment["status"]` は既に型安全。変更は不要だった | `typeof RESERVATION_STATUS_VALUES[number]` への変更（不要と判明） |
| status-colors.ts の重複 ReservationStatus 処理 | export を削除 + `@/types` から import | 誰もインポートしていない（grep 確認済み）。キャスト用途のみ `@/types` から import して使用 | export を `@/types` から re-export に変更 |
| VisitType（status-colors.ts）の処理 | export を削除（内部型 `VisitTypeDisplay` に変更） | 誰もインポートしていない（grep 確認済み）。`@/types` の `"first"\|"revisit"` が canonical | `@/types` から re-export |
| TrimmingRecord のリネーム先 | `TrimmingUI` | `models.ts` の `TrimmingRecord`（BE型）と明確に区別。`UI` suffix でフロントエンド専用型であることを明示 | 削除（utils/→trimming.tsへの移動は影響大） |

## 影響範囲

### DB
- 変更なし

### Backend
- 変更なし

### Frontend

**FE-115 対象:**
- `frontend/src/utils/constants/status-colors.ts` — 重複 `ReservationStatus` export 削除（line 21）+ `import type { ReservationStatus } from "@/types"` 追加
- `frontend/src/types/index.ts` — 変更不要（`ReservationAppointment["status"]` は既に proper union）

**FE-116 対象:**
- `frontend/src/utils/constants/status-colors.ts` — 未使用 `VisitType` export 削除（line 62）

**FE-117 対象:**
- `frontend/src/types/index.ts` — `TrimmingRecord` → `TrimmingUI` にリネーム（line 207）
- `frontend/src/features/trimming/api/get-trimming.ts` — import 更新
- `frontend/src/features/trimming/api/get-trimmings.ts` — import 更新
- `frontend/src/features/trimming/api/create-trimming.ts` — import 更新
- `frontend/src/features/trimming/api/update-trimming.ts` — import 更新
- `frontend/src/features/trimming/api/transforms.ts` — import 更新
- `frontend/src/features/trimming/routes/TrimmingList.tsx` — import 更新

## 参照実装

型の正規化パターンとして `src/types/owner.ts` の構造を参照。

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| `ReservationStatus` を proper union に変更した際の型エラー | 低 | `RESERVATION_STATUS_VALUES` に全値が定義済み。既存の `RESERVATION_STATUS_LABELS: Record<ReservationStatus, string>` も同値を使用しており整合 |
| `TrimmingRecord` リネームによる import 漏れ | 低 | 影響ファイルは4件（grep 確認済み）。build で検出可能 |

## 未解決事項

なし

## 実装順序

FE-115 → FE-116 → FE-117 の順（互いに独立しているが、同一ファイルへの変更があるため順次実施）

## 関連イシュー

- FE-115: [ReservationStatus proper union 修正](../../frontend/issues/open/FE-115-reservation-status-proper-union.md)
- FE-116: [status-colors.ts 未使用 VisitType 削除](../../frontend/issues/open/FE-116-remove-unused-visittype-export.md)
- FE-117: [TrimmingRecord を TrimmingUI にリネーム](../../frontend/issues/open/FE-117-rename-trimming-record-to-trimming-ui.md)
