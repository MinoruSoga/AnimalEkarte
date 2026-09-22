# S16: 問診抜粋 → カルテ詳細への履歴導線

> **目的**: カルテ編集画面の問診パネル「問診抜粋（履歴）」から同一ペットの過去カルテ詳細へ遷移し、戻っても文脈が失われないこと、および処置が未移行の古い記録でも詳細を開けることを納品前に証明する。
> **所要目安**: 10分 / **深度**: 薄い
> **仕様正本**: [screens/06-medical-records-form.md](../../../spec/screens/06-medical-records-form.md)・[screens/05-medical-records-list.md](../../../spec/screens/05-medical-records-list.md)。検証キュー: `UAT-Q2-HISTORY-NAV`（[todo-verification.md](../../../../todo-verification.md)）。

## 前提条件

- ローカルの使い捨て clinic、または承認済みの専用 UAT tenant。
- 同一ペットに複数件のカルテ（少なくとも 2 件、うち 1 件は治療明細が空または未移行の記録）を持つ fixture。固定 ID は仮定しない。
- actor は medical-records の参照権限を持つ attached account。
- 依存シナリオ: なし。問診の保存自体は S27（主訴）・V01 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 対象ペットのカルテ編集を開き、問診タブを表示する | 右パネル等に「問診抜粋」として同一ペットの過去記録一覧が出る（現在の記録を除く） |
| 2 | 抜粋の 1 行をクリックする | その記録のカルテ詳細/編集へ遷移する（`InterviewHistory` の `paths.medicalRecords.detail.getHref(item.id)`）。別ペット・別記録へ飛ばない |
| 3 | ブラウザの戻る、または画面内の戻る操作で戻る | 元のカルテ編集画面へ戻り、表示中のタブ・患者ヘッダーが維持される。未保存入力があれば `NavigationBlocker` が離脱確認を出す |
| 4 | 治療明細が空（または未移行）の過去記録の抜粋行をクリックする | 明細が空でも詳細画面は開け、エラー・真っ白画面にならない。空の治療タブは空状態として表示される |
| 5 | 別のペットのカルテで同じ操作をする | そのペット自身の履歴だけが抜粋に出る。他人・他ペットの記録が混ざらない |

## 確認観点

- `InterviewHistory`（`frontend/src/features/medical-records/components/`）の各行は `paths.medicalRecords.detail.getHref` で詳細へリンクする。履歴パネルの見出しは「問診抜粋」。
- 処置未移行の記録でも詳細が開けることは `UAT-Q2-HISTORY-NAV` の必須条件。治療明細の移行完了を意味しない（旧列→canonical 移行は `UAT-Q2-TREATMENTS-IMPORT` の別ゲート）。
- 戻る操作で未保存変更を捨てさせないことは [13 §2.3](../../../spec/screens/13-examinations-form.md) と同じ `NavigationBlocker` パターン。dirty 状態での離脱は確認ダイアログを経由する。
- 他人のカルテが履歴に混ざる場合は FAIL（`todo.md#product-bugs` で重複確認のうえ Linear 追跡）。行が空で出ないだけなら fixture 不足の可能性を先に切り分ける。

## 実装突合

- 変更サマリ:
  - `InterviewHistory` のリンク先 `paths.medicalRecords.detail.getHref` と「問診抜粋」見出しを現行コードと突合
  - `UAT-Q2-HISTORY-NAV` の「同一ペット行→詳細→戻る」「処置未移行でも開ける」を手順 2–4 に対応づけ
  - 未保存保護（`NavigationBlocker`）を手順 3 に組み込み、履歴導線と編集保護の併存を確認する形にした
