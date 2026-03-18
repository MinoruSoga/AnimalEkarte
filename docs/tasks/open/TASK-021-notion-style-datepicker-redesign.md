# TASK-021: 日付ピッカーを Notion 風にリデザイン

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

現在の `NotionDatePicker` コンポーネントに年月ヘッダーの重複表示バグがあり、加えて本物の Notion 日付ピッカーの UX に合わせたリデザインを行う。

## 依頼内容（原文）

> 飼主ペット登録編集ページなどの生年月日の日付選択フォームがおかしいです。
> Notionと同じ日付選択にして。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | バグ修正のみ(A) or Notion風リデザイン(B)？ | B（Notion風リデザイン） |

## 確認済みバグ（ブラウザテスト実施済み 2026-03-18）

- **年月ヘッダー重複表示**: 「2026 2026」「3月 3月」と年・月がそれぞれ2回表示
- **原因**: `calendar.tsx` の `caption_label` クラスが `captionLayout="dropdown"` 時に非表示になっていない（react-day-picker v9.13.0）
- **スクリーンショット**: ブラウザで `/owners/1` を開き、「飼主生年月日」をクリックして確認

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | NotionDatePicker を Notion 風にリデザイン + calendar.tsx バグ修正 | FE | FE-074 | - | [ ] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `/owners/1` の「飼主生年月日」をクリック → 年月ヘッダーが重複せず、年・月が1回ずつ表示される
- [ ] AC-2: カレンダーヘッダーに年月テキスト（例:「2026年 3月」）と `<` `>` 矢印で月送りナビゲーションが動作する
- [ ] AC-3: 年月テキストをクリック → 月選択グリッド（1月〜12月）が表示され、月をクリックでカレンダーに戻る
- [ ] AC-4: 月選択グリッド上部の年テキストに `<` `>` があり、年を前後に切り替えられる
- [ ] AC-5: テキスト入力欄に「2020/03/18」「2020-03-18」「20200318」等を入力 → カレンダーがその月に遷移し、日付が選択される
- [ ] AC-6: 「Today」ボタンで今日の日付が即時選択される
- [ ] AC-7: 「Clear」ボタン（既存の × アイコン）で日付がクリアされる
- [ ] AC-8: ペット編集モーダル（`PetEditModal`）の生年月日・去勢日でも同じ挙動
- [ ] AC-9: `DateRangePicker` / range モードの日付ピッカーも同様にヘッダーバグが修正されている
- [ ] AC-10: 既存の全日付ピッカー使用箇所（飼主birthDate、ペットbirthDate/neuteredDate、履歴フィルタ）で回帰なし

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 年月ナビゲーション方式 | Notion風：月グリッドビュー切り替え | Notion本家のUXに合わせる | react-day-picker の `captionLayout="dropdown"`（ネイティブselect） |
| テキスト入力 | ポップオーバー内上部にテキスト入力欄 | Notion本家と同じ配置 | トリガーボタン自体をinputにする |
| 日付パース | 手動パース（YYYY/MM/DD, YYYY-MM-DD, YYYYMMDD） | 外部ライブラリ不要、シンプル | date-fns parse（オーバースペック） |

## 影響範囲

### DB
- 変更なし

### Backend
- 変更なし

### Frontend
- `frontend/src/components/ui/calendar.tsx` — `caption_label` クラスバグ修正
- `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` — Notion風リデザイン（月グリッドビュー、テキスト入力、Todayボタン）
- `frontend/src/components/shared/DateRangePicker/DateRangePicker.tsx` — ヘッダーバグ修正の恩恵を受ける（calendar.tsx の修正で自動的に直る）

## 参照実装

- 本物の Notion 日付ピッカー — 年月ヘッダークリックで月グリッド表示、テキスト入力、Today ボタン
- `features/owners/routes/OwnerForm.tsx:372` — 飼主生年月日での使用例
- `features/owners/components/PetEditModal.tsx:283-294` — ペット生年月日での使用例

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| react-day-picker の内部構造変更で `captionLayout` の挙動が変わる可能性 | 低 | `captionLayout="dropdown"` をやめて独自ナビゲーションに切り替えることで依存を減らす |
| range モードへの影響 | 中 | range モードも同じ calendar.tsx を使うため、リデザイン後にHistoryFilterPanel等で回帰テスト |

## 未解決事項

- なし

## 実装順序

1. FE-074: calendar.tsx バグ修正 + NotionDatePicker Notion風リデザイン

## 関連イシュー

- FE-074: [日付ピッカー Notion風リデザイン](../../frontend/issues/open/FE-074-notion-style-datepicker-redesign.md)
