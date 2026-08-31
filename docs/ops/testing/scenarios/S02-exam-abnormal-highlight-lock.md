# S02: 検査異常値ハイライトと確定ロック

> **目的**: 基準値範囲外の検査値が自動でハイライトされ重要値の見落としを防ぐこと、および確定した検査記録が編集不可となり臨床データの真正性が守られることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/12-examinations-list.md](../../../spec/screens/12-examinations-list.md)・[screens/13-examinations-form.md](../../../spec/screens/13-examinations-form.md)

## 前提条件

- ローカルの使い捨て clinic に、生存ペット、`exam_type_field_id` が付いた検査項目、数値 Min/Max 基準範囲を承認済み fixture/import 手順で作成する。
- attached account を 2 つ用意する。一方は examinations create/edit、もう一方は examinations view のみを持つ。汎用 admin や seed 由来アカウントを仮定しない。
- 試験後に作成した検査記録、検査項目、アカウントを削除する。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 検査一覧から「新規検査登録」→ ペット選択 → 検査フォームで前提の検査種別を選択 | 選択した種別に基づき測定項目テーブルが動的生成され、項目名・単位・その動物種の基準値が表示される（[13 §1.2](../../../spec/screens/13-examinations-form.md)） |
| 2 | ある項目に基準値上限を超える測定値を入力して保存し、画面を再読込 | 当該項目が **HIGH** バッジ（赤 / `C.bgDanger`）および行ハイライトになる。判定はバックエンドが導出し保存・再読込後に反映（`status=high`・`is_abnormal` — [13 §1.2/§2.2](../../../spec/screens/13-examinations-form.md)）。UI ラベルは「H」ではなく **HIGH**（`ExamItemsTable`） |
| 3 | 別の項目に基準値下限未満の値を入力して保存・再読込 | **LOW** バッジとして **status-blue**（`C.textStatusBlue` / `C.bgStatusBlueLight`）でハイライト。仕様 13 の teal 表記とは実装トークンが異なる |
| 4 | 基準値ちょうど（Min または Max と同値）を入力して保存・再読込 | **正常扱い**（`computeExamResultStatus` / `assessExamResult` は inclusive range: `v < min` → low、`v > max` → high） |
| 5 | 基準値範囲内の値を入力して保存・再読込 | ハイライトなし（normal・通常表示） |
| 6 | 検査を「完了」にして保存し、再オープンする | `completed` かつ revision 無しは結果・削除を封印する（完了シール）。保存ボタンは消える。ステータスを「確定」に変えると保存 UI が一時的に戻る |
| 6b | ステータスドロップダウンから「確定」を**選択して保存**する（独立確定ボタンは無い）。一覧へ戻り再度開く | **ドラフトで「確定」を選んだだけではロックしない**。**サーバに confirmed が保存された後**、再オープン時にステータス/項目が無効化され保存ボタンが消える（[13 §2.1](../../../spec/screens/13-examinations-form.md)） |
| 7 | 確定済み（サーバ status=確定 / `confirmed`）の検査を開き、測定値の変更・保存を試行する | 編集がロックされ変更を保存できない。FE の persisted lock ＋ BE の confirmed 拒否の二重ガード。確定解除 UI は `examination-unconfirm:edit` があるときだけ（`ExaminationUnconfirmDialog`）。通常 edit では出ない |
| 7b | `examination-unconfirm:edit` 付きアカウントで確定解除（理由 1〜500 字）→ 印刷 | `POST /examinations/:id/unconfirm` が成功し再編集できる。印刷は `GET /examinations/:id/print-snapshot`（confirmed は official、それ以外は draft 透かし） |
| 8 | examinations view のみを持つ専用 attached accountでログインし、確定操作・結果入力を試行する | examinations の view は許可されるためフォームは見えるが、fieldset 無効・保存ボタンなし。第2アカウントが無い環境はBLOCKEDとして記録し、このscenarioを完了扱いにしない |

## 確認観点

- 異常値判定はバックエンド（`backend/internal/medicalrecord/examination_service.go` の `computeExamResultStatus` → `assessExamResult`）に集約。フロント（`ExamItemsTable`）は `status` / `isAssessed` / `isAbnormal` の表示専任（再計算しない）。
- 基準値未設定・未評価時は **未判定** バッジ（sr-only「基準値未設定のため判定していない」）。HIGH/LOW は `isAssessed` かつ status high/low のときのみ。
- ロックは二段: `confirmed` は全ロック。`completed` かつ `current_revision_version == nil` は結果・削除シール。確定解除は `examination-unconfirm`（ハイフン）。通常 edit では不可。
- カルテ詳細の「検査」タブと検査一覧の結果がリアルタイムに同期すること（[12 §2](../../../spec/screens/12-examinations-list.md)）。
- 検査一覧の進捗フィルタ（依頼中/検査中/結果入力済み/完了/確定）の全 5 ステータスで、現在取得済みのページ内から対象検査が正しく抽出できること（BE enum: pending/in_progress/result_entered/completed/confirmed。フィルタはクライアント側適用 — [12 §1.1](../../../spec/screens/12-examinations-list.md)）。
- **「未判定」の条件**: `exam_type_field_id` がない free-text 結果は基準範囲文字列があっても `is_assessed=false`。手順 2〜3 は fixture の field ID 付き項目で high/low/normal を確認する。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定済み検査に対する項目一括更新 API（PUT items）を直接送信する（USER が API クライアントで実施・任意） | バックエンドが拒否する（例: 「確定済みの検査は編集できません」— [13 §2.1](../../../spec/screens/13-examinations-form.md)） |
| A2 | 検査オーダーを取り消す（一覧から削除） | 依頼中は論理削除され一覧から消える。確定済み（および完了シール）は BE「確定済みの検査は削除できません」/ 完了済みは「完了済みの検査は削除できません」 |

---

## 実装突合

- 変更サマリ:
  - 異常値 UI ラベルを H/L → **HIGH/LOW**（`ExamItemsTable`）に合わせて修正
  - inclusive range・未判定バッジ・status 5 値の BE enum をソースに合わせて明記
  - 確定解除は専用権限がある旨を手順 7 に注記（本シナリオのロック確認は通常 edit）
