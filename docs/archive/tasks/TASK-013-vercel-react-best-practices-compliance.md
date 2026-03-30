# TASK-013: Vercel React Best Practices 準拠リファクタリング（全ドメイン）

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

全フロントエンド feature のコードを Vercel React Best Practices（CODING_RULES.md Section 12）に準拠させる。
参照実装 `features/owners/` をベンチマークとし、ドメイン単位でリファクタリングを実施する。

## 依頼内容（原文）

> ドメイン毎に、ドメイン関連のコードがvercel-react-best-practicesのコード規約に準拠するようにしてください。
> 例えば、1ドメイン=飼主・ペット　みたいな感じです。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | ドメイン分割（dashboard は「当日の受付」で正しいか） | dashboard は「当日の受付」画面 |
| 2 | 既に準拠している feature はイシュー作成不要か | yes |
| 3 | hospital-settings の react-hook-form → useTransition 置換でよいか | yes |

## 監査結果

### 準拠済み（イシュー不要）

| Feature | 根拠 |
|---------|------|
| owners (参照実装) | ベンチマーク |
| accounting | useDeferredValue, useCallback, useMemo, hoisted constants 全準拠 |
| vaccinations | useDeferredValue, hoisted constants, useCallback 全準拠 |
| examinations | useDeferredValue, hoisted constants, useCallback 全準拠 |
| auth | memo(), useCallback, hoisted constants 全準拠 |
| master | hoisted constants, navigation のみ（フォームなし）|
| estimates | useCallback, hoisted constants, NavigationBlocker 全準拠 |
| inventory | useCallback, hoisted constants 全準拠 |
| trimming | memo() 3セクション分割, lazy+Suspense, useDeferredValue, useMemo 全準拠 |
| pets | API のみ（UI route なし）|

### 要修正（5ドメイン・5 feature）

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | 病院設定: react-hook-form 除去 + useTransition 化 + 型統一 | FE | FE-055 | - | [x] |
| 2 | 予約・シフト: memo() 追加 + barrel index 除去 + useTransition 化 | FE | FE-056 | - | [x] |
| 3 | 入院: memo() 追加 + hoisted JSX + barrel index 除去 | FE | FE-057 | - | [x] |
| 4 | カルテ: memo() 追加 + hoisted static data | FE | FE-058 | - | [x] |
| 5 | 当日の受付: memo() 追加 + useMemo for column list | FE | FE-059 | - | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: 全対象 feature で `npm run build` がエラーなくパスする
- [x] AC-2: 全対象 feature で `npm run lint` がエラーなくパスする
- [ ] AC-3: 各イシューの完了条件にリストされた Vercel ルールが全て適用されている
- [ ] AC-4: 既存の UI 動作に変更がない（見た目・操作フロー一切変更なし）
- [ ] AC-5: barrel index (`components/index.ts`, `api/index.ts`) が対象 feature から除去されている

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| hospital-settings フォーム管理 | useTransition + useState | プロジェクト標準（owners パターン） | react-hook-form 継続使用 |
| memo() 適用基準 | 50行以上 or props 受け取りの子コンポーネント | Vercel ガイドライン + owners 参照実装 | 全コンポーネント一律 memo |
| barrel index 対応 | 削除 + 直接 import に変更 | tree-shaking + CLAUDE.md ルール | 残して lint 除外 |

## 影響範囲

### Frontend

- `frontend/src/features/hospital-settings/` — react-hook-form 除去、useTransition 化、型統一
- `frontend/src/features/reservations/` — memo() 追加、barrel index 除去
- `frontend/src/features/shifts/` — barrel index 除去、lazy init 修正、useTransition 追加
- `frontend/src/features/hospitalization/` — memo() 追加、hoisted JSX、barrel index 除去
- `frontend/src/features/medical-records/` — memo() 追加、static data 巻き上げ
- `frontend/src/features/dashboard/` — memo() 追加、useMemo for column list

### Backend
- 変更なし

### DB
- 変更なし

## 参照実装

- `features/owners/` — 全パターンのベンチマーク
  - `OwnerForm.tsx` — memo() セクション分割、useCallback 安定化、lazy+Suspense
  - `useOwnerForm.ts` — useTransition フォーム pending 管理
  - `OwnersList.tsx` — useDeferredValue 検索、useMemo フィルタ
- `features/trimming/` — 3カラム memo 分割、lazy modal、useDeferredValue 履歴検索

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| react-hook-form 除去で hospital-settings フォームの動作が変わる | 高 | owners パターンに忠実に実装 + 手動テスト |
| memo() 追加で props 比較コストが増える | 低 | 50行以上のコンポーネントのみ対象 |
| barrel index 除去で import パスの変更漏れ | 中 | `npm run build` で型エラー検出 |

## 未解決事項

- なし

## 実装順序

1. FE-055: 病院設定（Critical・独立性高い）
2. FE-056: 予約・シフト
3. FE-057: 入院
4. FE-058: カルテ
5. FE-059: 当日の受付
※ 全イシュー間に依存関係なし。並行実装可能。

## 関連イシュー

- FE-055: [病院設定 — react-hook-form 除去 + Vercel Best Practices 準拠](../../frontend/issues/open/FE-055-hospital-settings-vercel-compliance.md)
- FE-056: [予約・シフト — memo/barrel index/useTransition 準拠](../../frontend/issues/open/FE-056-reservations-shifts-vercel-compliance.md)
- FE-057: [入院 — memo/hoisted JSX/barrel index 準拠](../../frontend/issues/open/FE-057-hospitalization-vercel-compliance.md)
- FE-058: [カルテ — memo/hoisted static data 準拠](../../frontend/issues/open/FE-058-medical-records-vercel-compliance.md)
- FE-059: [当日の受付 — memo/useMemo 準拠](../../frontend/issues/open/FE-059-dashboard-vercel-compliance.md)
