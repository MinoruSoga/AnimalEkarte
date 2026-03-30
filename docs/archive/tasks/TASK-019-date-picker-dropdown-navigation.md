# TASK-019: 日付ピッカーに年月ドロップダウンナビゲーションを追加

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー（UX改善要望）

---

## 概要

飼主・ペットの登録編集ページの生年月日入力で、月送りボタン（Chevron）しかないため過去の日付へのナビゲーションが非常に不便。react-day-picker v9 の `captionLayout="dropdown"` を活用し、年・月のドロップダウンを追加する。

## 依頼内容（原文）

> 飼主・ペットの登録編集ページの生年月日の日付入力フォームにて、
>
> 生年月日の入力、当月以外の日付に行くときに選択するところの判定が分かりにくくて不便
> 1月と2月でさかのぼりのクリックするところがズレるのが時間がかかる
> ここは誕生日なのでさかのぼることが多いので、今日からさかのぼる場合、10歳とかだと大変になる

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 年の選択範囲（下限）は？ | 100年前 |
| 2 | 生年月日以外の日付ピッカーにも適用するか？ | Yes（全日付ピッカーに適用） |
| 3 | 年の並び順 | デフォルト採用: 新しい順（reverseYears） |
| 4 | 月送りボタンの併用 | デフォルト採用: ドロップダウンのみ |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | NotionDatePicker に captionLayout="dropdown" + startMonth/endMonth を追加 | FE | FE-025 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 飼主登録画面の生年月日フィールドをクリックすると、カレンダー上部に年・月のドロップダウンが表示される
- [ ] AC-2: 年ドロップダウンには現在年から100年前までの年が新しい順で表示される
- [ ] AC-3: 月ドロップダウンから任意の月を選択すると、即座にカレンダーがその月に遷移する
- [ ] AC-4: ペット編集モーダルの生年月日・去勢手術日でも同様にドロップダウンが機能する
- [ ] AC-5: ワクチン接種日・次回接種日でも同様にドロップダウンが機能する
- [ ] AC-6: 履歴フィルタの日付範囲ピッカー（RangeMode）でも同様にドロップダウンが機能する
- [ ] AC-7: 既存の日付選択機能（クリックで日付選択、クリアボタン）が壊れていない
- [ ] AC-8: `npm run build` / `npm run lint` がエラーなしでパスする

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| ドロップダウン実装方法 | react-day-picker v9 の `captionLayout="dropdown"` | 標準機能で追加コード最小 | カスタムSelect + state管理 |
| 年範囲指定 | `startMonth` + `endMonth` | v9 の公式API（fromYear/toYear は deprecated） | fromYear/toYear |
| 年の並び順 | `reverseYears` で新しい順に表示 | 生年月日は近年から選ぶことが多い | デフォルト（古い→新しい） |

## 影響範囲

### Backend
- 変更なし

### Frontend
- `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` — `captionLayout`, `startMonth`, `endMonth` props を Calendar に追加
- `frontend/src/components/ui/calendar.tsx` — ドロップダウン用の classNames 追加（`dropdowns`, `dropdown`, `dropdown_root`, `months_dropdown`, `years_dropdown`）

### 使用箇所（全7箇所、変更不要 — 共有コンポーネント側で対応）
1. `features/owners/routes/OwnerForm.tsx:372` — 飼主生年月日
2. `features/owners/components/PetEditModal.tsx:286` — ペット生年月日
3. `features/owners/components/PetEditModal.tsx:348` — 去勢・避妊手術日
4. `components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx:55` — 履歴フィルタ開始日
5. `components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx:62` — 履歴フィルタ終了日
6. `features/medical-records/components/VaccinationForm.tsx:86` — ワクチン接種日
7. `features/medical-records/components/VaccinationForm.tsx:216` — 次回接種日

## 参照実装

- react-day-picker v9 公式: `captionLayout="dropdown"` + `startMonth` + `endMonth`
- `features/owners/` — NotionDatePicker の使用パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| ドロップダウンのスタイルが Notion 風デザインと合わない | 中 | calendar.tsx の classNames でドロップダウンのスタイルを調整 |
| RangeMode（2ヶ月表示）でのドロップダウン表示崩れ | 低 | RangeMode でも動作確認する |

## 未解決事項

- なし

## 実装順序

1. `calendar.tsx` にドロップダウン用 classNames を追加
2. `NotionDatePicker.tsx` の Calendar に `captionLayout`, `startMonth`, `endMonth` props を追加
3. 全7箇所の動作確認

## 関連イシュー

- FE-025: [日付ピッカーにドロップダウンナビゲーション追加](../../frontend/issues/open/FE-025-date-picker-dropdown-navigation.md)
