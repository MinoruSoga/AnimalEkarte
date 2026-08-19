# S02: 検査異常値ハイライトと確定ロック

> **目的**: 基準値範囲外の検査値が自動でハイライトされ重要値の見落としを防ぐこと、および確定した検査記録が編集不可となり臨床データの真正性が守られることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/12-examinations-list.md](../../../spec/screens/12-examinations-list.md)・[screens/13-examinations-form.md](../../../spec/screens/13-examinations-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: vet ロール（examinations の create/edit 権限あり）。
- 対象ペット: ペット検索で「ステータス=生存」の犬を 1 頭選ぶ。
- 検査種別: 検査種別マスタのうち、基準値（Min〜Max）が定義された測定項目テンプレを持つ種別（例: 名称に「血液」を含む種別。WBC/RBC 等に基準値あり）。
- 手順 8 は seed 003_demo の「閲覧専用」権限グループ（examinations は view のみ）を割り当てたスタッフで実施できること。
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
| 8 | seed 003_demo の「閲覧専用」権限グループを割り当てたスタッフでログインし、確定操作・結果入力を試行する | examinations の view は許可されるためフォーム・ステータス・結果項目は表示されるが、フォーム全体が無効化され保存ボタンは表示されない（[13 概要](../../../spec/screens/13-examinations-form.md)）。【要実測】実ブラウザで値を変更・保存できないこと。**runtime 2026-08-01 BLOCKED**: 第2アカウント（閲覧専用）未用意 |

## 確認観点

- 異常値判定はバックエンド（`backend/internal/medicalrecord/examination_service.go` の `computeExamResultStatus` → `assessExamResult`）に集約。フロント（`ExamItemsTable`）は `status` / `isAssessed` / `isAbnormal` の表示専任（再計算しない）。
- 基準値未設定・未評価時は **未判定** バッジ（sr-only「基準値未設定のため判定していない」）。HIGH/LOW は `isAssessed` かつ status high/low のときのみ。
- ロックは二段: `confirmed` は全ロック。`completed` かつ `current_revision_version == nil` は結果・削除シール。確定解除は `examination-unconfirm`（ハイフン）。通常 edit では不可。
- カルテ詳細の「検査」タブと検査一覧の結果がリアルタイムに同期すること（[12 §2](../../../spec/screens/12-examinations-list.md)）。
- 検査一覧の進捗フィルタ（依頼中/検査中/結果入力済み/完了/確定）の全 5 ステータスで、現在取得済みのページ内から対象検査が正しく抽出できること（BE enum: pending/in_progress/result_entered/completed/confirmed。フィルタはクライアント側適用 — [12 §1.1](../../../spec/screens/12-examinations-list.md)）。
- **observed FAIL / DEFER（2026-08-01 browser）**: 検査詳細で基準値文字列（例 WBC「55～195」）は表示されるが判定は常に『未判定』— 異常値入力でも HIGH/LOW 未発火。犬猫切替の基準値次元は未証明（【要実測】残）。実装上はマスタ基準値が解決されないと `is_assessed=false` になる。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定済み検査に対する項目一括更新 API（PUT items）を直接送信する（USER が API クライアントで実施・任意） | バックエンドが拒否する（例: 「確定済みの検査は編集できません」— [13 §2.1](../../../spec/screens/13-examinations-form.md)） |
| A2 | 検査オーダーを取り消す（一覧から削除） | 論理削除され一覧から消える（[12 API連携](../../../spec/screens/12-examinations-list.md)）。【要実測】確定済み検査の削除可否（BE: 「確定済みの検査は削除できません」/ 確定履歴ありも拒否）。**runtime 2026-08-01 BLOCKED**: 一覧サンプルは依頼中のみ（確定済なし）。依頼中に削除ボタンは存在 |

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - 異常値 UI ラベルを H/L → **HIGH/LOW**（`ExamItemsTable`）に合わせて修正
  - inclusive range・未判定バッジ・status 5 値の BE enum をソースに合わせて明記
  - 確定解除は専用権限がある旨を手順 7 に注記（本シナリオのロック確認は通常 edit）

- runtime 2026-08-07: **PARTIAL** (auth unlocked) — DB `exam_reference_ranges` COUNT=20; pure unit green. Authenticated API: create exam type=血液検査（院内）+ `exam_type_field_id` on living dog pet → PUT items yields **high/low/normal** with `is_assessed=true` (300→high, 1→low, 6/10/17→normal, inclusive). Seed pending exams with free-text `normal_value` only (no field id) stay `is_assessed=false` even when range text shows. Playwright `examinations-flow` 5/5 PASS (list/select-pet/new/detail/search). Full UI HIGH/LOW badge still needs browser entry on form (not asserted in e2e). Cleanup: probe exams deleted.
