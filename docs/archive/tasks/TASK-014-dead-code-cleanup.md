# TASK-014: デッドコード クリーンアップ（全ドメイン）

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

全ドメインのフロントエンド・バックエンドコードを調査し、未使用の関数・ファイル・型定義・インポート・再エクスポートを削除する。

## 依頼内容（原文）

> ドメイン毎に、ドメイン関連のコードのを確認して、デットコードをクリーンアップして。
> 例えば、1ドメイン=飼主・ペット　みたいな感じです。

## 仕様確認ログ

確認事項なし。デッドコードの定義は明確（エクスポートされているが一度もインポートされていない関数・型、コメントのみのファイル、deprecated 再エクスポート）。

## 監査結果

### バックエンド
**全ドメインでデッドコードなし。** 全 handler メソッドが `main.go` に登録済み、全 service/repository メソッドが handler から呼び出されている。BE イシューは発行しない。

### フロントエンド
**体系的パターン**: 7 feature に未使用フィルタ API 関数（`getXxxByPetId`, `getXxxByOwnerId`, `getXxxByStatus` + 対応 hook）が存在。UI はメインリスト API + クライアントサイドフィルタで実装されており、これらは一度も使われていない。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | 飼主・ペット・認証: dead auth files + 再エクスポート除去 | FE | FE-060 | - | [x] |
| 2 | 診察・ワクチン・カルテ: 未使用フィルタ API 関数除去 | FE | FE-061 | - | [x] |
| 3 | 会計・見積・在庫: 未使用フィルタ API 関数 + 空ファイル除去 | FE | FE-062 | - | [x] |
| 4 | 入院・予約・シフト: 未使用フィルタ API 関数除去 | FE | FE-063 | - | [x] |
| 5 | トリミング・当日の受付: 未使用フィルタ API + import 除去 | FE | FE-064 | - | [x] |
| 6 | 共通基盤: 未使用 hooks・deprecated 再エクスポート除去 | FE | FE-065 | - | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: 全イシューのデッドコードが削除されている
- [x] AC-2: `npm run build` がエラーなくパスする
- [x] AC-3: `npm run lint` がエラーなくパスする
- [ ] AC-4: 既存の UI 動作に変更がない
- [ ] AC-5: 削除対象の関数・ファイルが他のどこからもインポートされていないことを `grep` で確認済み

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 未使用フィルタ API 関数 | 削除 | 現在のアーキテクチャではクライアントサイドフィルタを採用しており、これらは不要 | コメントアウトして残す |
| barrel index 内の再エクスポート | 削除行のみ除去（index.ts は残す） | 他の有効な export が同じ index.ts に存在する | index.ts ごと削除 |

## 影響範囲

### Backend
- 変更なし（デッドコードなし）

### Frontend
- `features/examinations/api/` — 未使用関数削除
- `features/vaccinations/api/` — 未使用関数削除
- `features/medical-records/api/` — 未使用関数削除
- `features/accounting/api/` — 未使用関数削除
- `features/hospitalization/api/` — 未使用関数削除
- `features/reservations/api/` — 未使用関数削除
- `features/trimming/api/` + `routes/` — 未使用関数・import 削除
- `features/auth/hooks/` — dead ファイル削除
- `hooks/` — 未使用 hook・deprecated 再エクスポート削除
- `main.tsx` — コメント削除
- `features/inventory/types/index.ts` — 空ファイル削除

## 参照実装

- なし（削除のみのタスク）

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 削除した関数が実は別ブランチで使われている | 低 | git で復元可能。削除前に grep で確認済み |

## 未解決事項

- なし

## 実装順序

全イシュー間に依存関係なし。並行実装可能。
1. FE-060 〜 FE-065 を任意の順序で実施
2. 最後に `npm run build` + `npm run lint` で全体確認

## 関連イシュー

- FE-060: [飼主・ペット・認証 — dead files + 再エクスポート除去](../../frontend/issues/open/FE-060-owners-pets-auth-dead-code.md)
- FE-061: [診察・ワクチン・カルテ — 未使用フィルタ API 除去](../../frontend/issues/open/FE-061-exam-vacc-records-dead-code.md)
- FE-062: [会計・見積・在庫 — 未使用フィルタ API + 空ファイル除去](../../frontend/issues/open/FE-062-accounting-estimates-inventory-dead-code.md)
- FE-063: [入院・予約・シフト — 未使用フィルタ API 除去](../../frontend/issues/open/FE-063-hospitalization-reservations-shifts-dead-code.md)
- FE-064: [トリミング・当日の受付 — 未使用フィルタ API + import 除去](../../frontend/issues/open/FE-064-trimming-dashboard-dead-code.md)
- FE-065: [共通基盤 — 未使用 hooks・deprecated 再エクスポート除去](../../frontend/issues/open/FE-065-shared-dead-code.md)
