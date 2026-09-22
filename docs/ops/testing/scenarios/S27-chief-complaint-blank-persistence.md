# S27: 主訴 — 区分空欄許可・本文保持・意図的解除

> **目的**: 主訴が区分未選択（空欄）でも本文を失わず保存でき、既存の区分を意図的に解除した場合もその解除が保持され、再読込・記録切替で値が破綻しないことを実 UI で納品前に証明する。
> **所要目安**: 15分 / **深度**: 中
> **仕様正本**: [screens/06-medical-records-form.md](../../../spec/screens/06-medical-records-form.md)。関連 tracker: `SLACK-COMPLAINT`（主訴区分空欄許可・実 UI 解除は検証待ち）。

## 前提条件

- ローカルの使い捨て clinic、または承認済みの専用 UAT tenant。
- 主訴区分マスタの fixture と、区分未選択・区分選択済みの両方の状態を作れる medical-records 権限の attached account。
- 区分未選択は医院依頼で確定済みの許可（可否を再質問しない）。
- 依存シナリオ: なし。問診全体の保存は V01、履歴導線は S16 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 主訴区分を**選択せず**、主訴本文だけ入力して保存する | 区分空欄でも本文が保存される（`chief_complaint_type_id: null` を許容）。エラーで保存できない・本文が消えるのは FAIL |
| 2 | 再読込する | 区分は空のまま、本文は保持される。`use-apply-medical-record` が `chiefComplaintTypeId ?? null` で hydrate する |
| 3 | 区分を選択して本文を変えて保存 → 再読込 | 区分・本文ともに保持される |
| 4 | 選択済みの区分を**実 UI で解除する**（クリア操作。SearchableSelect の clearable） | 解除操作が実 UI で到達可能で、解除後に区分が空になる。解除手段がなければ `SLACK-COMPLAINT` 残件として記録し、mock 上の解除で済ませない |
| 5 | 解除して保存 → 再読込 | 区分が空のまま保持される。前の区分が勝手に復活しない |
| 6 | 別のカルテ/ペットへ切り替えて戻る | 記録ごとの区分・本文が混ざらない。記録切替で別レコードの値が残らない |

## 確認観点

- 区分は `updateInquiryMutation` の `chief_complaint_type_id: number | null` で null を許容。空欄は仕様（依頼済み）でありエラーにしない。
- hydrate は `use-apply-medical-record` が `chiefComplaintTypeId ?? null` を setter へ渡す。再読込で区分が `0` や空文字に化けない。
- **実 UI での解除到達性**は `SLACK-COMPLAINT` の残件。テストでは mock の空値 callback を使っていたため、実際の `SearchableSelect`（`clearable`）で解除操作ができるかを本シナリオで確認する。到達できなければ既存 Issue へ接続する。
- 本文保持は区分の有無に関わらず必須。区分だけ変えたとき本文が消える・区分を外すと本文が消えるのは FAIL。
- 記録切替での値の持ち越し（別レコードの区分が残る）は FAIL。レコードごとの独立性を確認する。

## 実装突合

- 変更サマリ:
  - `use-apply-medical-record` の `chiefComplaintTypeId ?? null` hydrate、`InterviewChiefComplaint` の `chiefComplaintTypeId`/`setChiefComplaintTypeId`・`clearable` SearchableSelect、`use-medical-record-save-action` の `chief_complaint_type_id: number | null` を現行コードと突合
  - `SLACK-COMPLAINT` の「区分空欄で本文保存・実 UI 解除・再読込保持」を手順 1–6 に対応づけ、実 UI 解除の未到達を既存 Issue 接続として明記
