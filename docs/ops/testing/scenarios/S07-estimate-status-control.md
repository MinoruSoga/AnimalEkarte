# S07: 見積ステータス制御

> **目的**: 承認・却下で確定した見積が UI・API の両面で変更不能になり、確定前の見積だけが編集・削除できることを証明する。
> **所要目安**: 15分 / **深度**: 深い
> **仕様正本**: [見積一覧](../../../spec/screens/22-estimate-list.md) / [見積作成・編集](../../../spec/screens/23-estimate-form.md) / [見積詳細](../../../spec/screens/26-estimate-detail.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。**seed 003_demo の `estimates.csv` には八王子病院（clinic_id=1）の draft 見積が多数含まれる**（2026-07 以降のデモ seed）。本シナリオのロック／フィルタ検証データはタイトル「S07 検証用*」で**新規作成**し、seed 既存行と混同しないこと（城東 clinic では seed 見積が無い前提でも新規作成で実施可）。
- ログイン: 権限グループ「執行」のスタッフ（estimates の view/create/edit/delete を保有）。
- ステータスは「下書き」「送付済み」「承認済み」「却下」の 4 値（`EstimateStatusBadge` の `STATUS_LABELS`＝draft/sent/approved/rejected）。承認済み・却下＝確定（ロック）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:---|:---|
| 1 | 見積一覧 `/estimates` →「新規見積書登録」。タイトル「S07 検証用A」とヘッダ金額を入力し、ステータス「下書き」で保存 | 一覧に新規行が追加され、ステータスバッジが「下書き」（グレー）。独立画面は明細行を送らない。詳細を開くと明細は空でヘッダ金額だけが出る |
| 2 | 該当行の操作メニュー →「編集」。ステータスを「送付済み」へ変更して保存 | 保存に成功し、バッジが「送付済み」（青）に変わる |
| 3 | 再度編集し、ステータスを「承認済み」へ変更して保存 | バッジが「承認済み」（緑）に変わる |
| 4 | 一覧へ戻り、該当行の操作メニューを開く | 「詳細」のみ表示され、「編集」「削除」の導線が表示されない（`isEstimateLockedStatus`） |
| 5 | 該当見積の詳細画面 `/estimates/:id` を開く | 編集・削除ボタンが表示されない |
| 6 | URL 直叩きで `/estimates/{当該ID}/edit` を開く | トースト「承認済みまたは却下済みの見積書は編集できません」（`ESTIMATE_LOCKED_EDIT_MESSAGE`）後に詳細へリダイレクト |
| 7 | タイトル「S07 検証用B」の見積を「下書き」→「送付済み」→「却下」の順で作成・更新 | バッジが「却下」（赤）になり、#4〜#6 と同様に編集・削除導線が消える |
| 8 | 一覧の状態フィルタで「承認済み」「却下」を順に絞り込む | それぞれ検証用A / 検証用B が含まれる（seed 既存 approved/rejected が無い場合は当該 2 件のみ。仕様正本 22 §1.1 状態フィルタ） |
| 9 | タイトル「S07 検証用C」の見積を「下書き」のまま保存し、一覧から編集（金額変更）→ 保存 → 削除 | （対照）編集・削除とも導線が表示され、いずれも成功する |

## 確認観点

- **確定ロックの不変条件**: 承認済み・却下の見積は Update/Delete が API レベルで拒否される（`backend/internal/billing/estimate_service.go` の `isEstimateLocked` ＋ `estimate_repository.go` の status NOT IN 述語による原子的拒否）。新規作成のステータスは draft/sent のみ許可（approved/rejected 指定は Conflict 拒否）。
- **#6 の実装由来の期待値**: ロック済み見積の編集 URL 直アクセスは `EstimateForm` / `use-estimate-form` が `isEstimateLockedStatus` 判定で toast + detail へ replace。【要実測】**DEFER** — 承認済み見積 ID への `/edit` 直叩き未実施（一覧 smoke のみ）。ユニットは `EstimateForm.test.tsx` が cover。
- **監査証跡**: Create/Update/Delete の監査は **best-effort**（`logEstimateChangeBestEffort` — 監査失敗でも本体は成功）。#1〜#3・#9 実施後に audit_logs へ対応レコードがあることを確認（欠落時はログを確認し、fail-closed とは誤認しない）。後継作成（下記）のみ fail-closed。
- **削除の性質**: 見積の削除は論理削除（仕様正本 22 §2）。#9 の削除後、一覧に再表示されないこと。
- **created_by 検証**: 見積作成者はサービス層で同一クリニック所属を検証される（画面からの通常操作では常に成立するため、逸脱がないことのみ確認）。
- **割引権限**: 見積フォームの割引額欄は discount 権限を持たないユーザーには無効化される（仕様正本 23 §1.2）。「執行」ロールでは入力可能であること。
- **clinic_id 隔離**: 本シナリオで作成した見積が、他クリニック選択時の一覧に現れないこと。

## 異常系

- **API レベルの更新・削除拒否**: 承認済み・却下見積への PATCH/DELETE 拒否はサービス層テストで検証済みのため、本シナリオでは UI 導線消失（#4〜#6）の確認を主とする。ブラウザからの API 直叩き再現は行わない。
- **確定見積の修正が必要になった場合（TASK-012 FINAL B）**: 承認済み・却下は**不可逆**（unlock / 「下書きへ戻す」は存在しない。仕様正本 26 §2.1 と同旨）。訂正は後継ドラフトのみ:
  - `POST /api/v1/estimates/:id/successors`（権限 `estimates:create`）
  - body: `{ "reason": string required min=1 max=500 }`（任意で title/comment/notes 上書き）
  - 201: 新規 draft 見積（新 ID・新 estimate_no・`supersedes_estimate_id` = 原見積 ID）。原行は不変。
  - 監査: action=`supersede` を同一 TX で fail-closed 記録。
  - 確定カルテに紐付く原見積でも後継作成を許可する（カルテ reopen 不要の明示訂正パス）。
  - 詳細画面の「後継ドラフトを作成」（理由 1〜500 字）。201 で新 draft（新 ID・新 estimate_no・`supersedes_estimate_id`）。原行は不変。原見積の明細があればコピーする（ヘッダのみ原見積なら後継も空明細）。

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - seed 003 に八王子 draft 見積が存在することを前提に追記（「見積無し」記述を撤回）
  - バッジ文言を実装どおり「送付済み」「承認済み」に修正
  - Create/Update/Delete 監査を best-effort、successors のみ fail-closed と明記
  - 後継ドラフトは詳細 UI「後継ドラフトを作成」。独立見積はヘッダ金額のみ
  - ルート/ロック実装パス（`isEstimateLockedStatus` / `ESTIMATE_LOCKED_EDIT_MESSAGE`）を現行 main で再確認
