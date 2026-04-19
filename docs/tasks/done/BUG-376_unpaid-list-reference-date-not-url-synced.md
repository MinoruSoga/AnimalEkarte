# BUG-376: 未納者一覧の基準日が URL クエリパラメータに同期されない

**作成日**: 2026-04-14
**Status**: Open
**Priority**: LOW
**Affects**: `frontend/src/features/accounting/routes/UnpaidCustomerList.tsx`
**関連**: BUG-370 (月末未納者一覧), BUG-374 (ブラウザテスト)

## 概要

BUG-370 の実装仕様書 TC-370-11 は「ブラウザリロード → URL クエリパラメータ復元（タブ + 基準日）」を求めているが、現実装では `group_by` のみ URL 同期され、基準日（reference_date）は URL に反映されない。ページリロードで基準日は今日にリセットされる。

## 再現

1. `/accounting/unpaid` を開く
2. 基準日ピッカーを 2026-03-20 に変更
3. URL が `/accounting/unpaid?group_by=owner` のまま（基準日クエリなし）
4. F5 リロード
5. 基準日が 2026-04-14（今日）に戻る

## 根拠コード

`frontend/src/features/accounting/routes/UnpaidCustomerList.tsx:44-60`

```tsx
const [searchParams, setSearchParams] = useSearchParams();
const groupBy: GroupBy = (searchParams.get("group_by") as GroupBy) === "billing" ? "billing" : "owner";
// ← reference_date の searchParams 同期コードが存在しない
```

## 修正方針

1. `useState` 管理の基準日を `searchParams.get("reference_date")` 由来に変更
2. ピッカー変更ハンドラで `setSearchParams` に `reference_date=YYYY-MM-DD` を設定
3. 不正な日付文字列（past 3 年超など）は今日にフォールバック

## 影響

- ブックマーク・URL 共有時に基準日が保持されない → 運用で過去日付を共有不可
- 影響度 LOW（機能自体は動作、UX の劣化のみ）

## 完了条件

- [ ] `/accounting/unpaid?group_by=billing&reference_date=2026-03-20` 直接アクセスで両パラメータ復元
- [ ] ピッカー変更時に URL 更新（replace mode）
- [ ] リロード耐性確認

## 検証履歴

- 2026-04-14 BUG-374 ブラウザテスト中に発見（TC-370-11 NG）
