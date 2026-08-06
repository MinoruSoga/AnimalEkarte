# S06: カルテ確定 Lock と監査証跡

> **目的**: 確定（Lock）したカルテが編集・削除不可となり法的記録としての真正性が守られること、および一連の CRUD 操作が `audit_logs` に操作者・時刻付きで追跡されることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/05-medical-records-list.md](../../../spec/screens/05-medical-records-list.md)・[screens/06-medical-records-form.md](../../../spec/screens/06-medical-records-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: vet ロール（medical-records の create/edit 権限あり）。
- 対象ペット: ペット検索で「ステータス=生存」のペットを 1 頭選ぶ。
- `audit_logs` の確認は管理 UI が存在しないため DB 参照（**USER 実施**）。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | カルテ新規作成 → ペット選択 → 「問診」タブで主訴、「診察/治療プラン」タブで所見（S/O/A）・診断名を入力し保存（[06 §1/§2.2](../../../spec/screens/06-medical-records-form.md)） | 保存成功。カルテ一覧にステータス「作成中」で表示される（[05 §1.2](../../../spec/screens/05-medical-records-list.md)) |
| 2 | 「治療」タブで処置明細を 1 件追加する | 明細は行単位で即時保存される（[06 §2.2](../../../spec/screens/06-medical-records-form.md)） |
| 3 | カルテ本体（主訴等）を変更して再保存する | 保存成功。「作成中」の間は臨床内容の追記・変更が可能（[05 §1.2](../../../spec/screens/05-medical-records-list.md)） |
| 4 | カルテを確定（Lock）する: 画面右下フローティングの「確定する」（`MedicalRecordFloatingActions`。`canEdit`・保存済み・未確定のときのみ）→ `MedicalRecordFinalizeDialog` で「確定後はカルテを編集できなくなります。修正が必要な場合は訂正追記（addendum）を使用してください。この操作は元に戻せません。」を確認し「確定する」（[06 §2.3](../../../spec/screens/06-medical-records-form.md)） | ステータスが「確定済」（BE `finalized`）に変わる（[05 §1.2](../../../spec/screens/05-medical-records-list.md)）。詳細に確定済みバナーが出て「確定する」ボタンは消える |
| 5 | 確定済みカルテを開き、属性を変更して保存を試行する | 主要タブは fieldset 一括無効（`isFinalized` または `!canSubmit` — [06 §2.3](../../../spec/screens/06-medical-records-form.md)）。保存ボタンも disabled。直接 API 更新は 409: 「確定済みカルテは編集できません。訂正追記 (addendum) を使用してください」（`medical_record_crud.go`） |
| 6 | 確定済みカルテの削除導線を確認する | **詳細**: `canDelete={canDelete && !isFinalized}` のためフローティング「削除」は出ない。**一覧**: 現行 UI は確定済でも canDelete なら「削除」メニューが出うる — 実行時 BE が拒否する（draft 以外は削除不可）。直接 API: 「確定済みまたは下書き以外の診療記録は削除できません」（[05 §3.1](../../../spec/screens/05-medical-records-list.md)） |
| 7 | 確定済みカルテの各タブ下部「追記する」から訂正追記（addendum）を行う（修正内容・修正理由が必須 — `AddendumModal`） | 保存後、時刻・スタッフ ID 付きの時系列リストとして表示され、リロード後も永続する（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）。addendum は fieldset 外（`MedicalRecordAddenda`） |
| 8 | 確定の解除（確定済 → 作成中へ戻す）を試行する | 解除（unfinalize）API はバックエンドに存在しない（`medical_record_crud.go` に該当メソッドなし）。確定は一方向遷移であり、確定後の修正経路は訂正追記（addendum）のみ。解除機能が必要な場合は GAP-1 とは別に要件を起票する（2026-07-16 Fable 代理決裁） |
| 9 | 手順 1〜4 の操作について `audit_logs` を DB で確認する（**USER 実施**。例: resource がカルテ関連の行を時刻降順で参照） | 作成（create）・更新（update）・確定の各操作が、操作者（actor_id）・時刻（created_at）・変更内容（old/new_value）付きで記録されている（[specification.md §2.1](../../../spec/specification.md)「全テーブルの CRUD を audit_logs で追跡」） |

## 確認観点

- 確定ロックは DB 制約ではなくサービス層のガード（`backend/internal/medicalrecord/medical_record_crud.go` — [05 §3.1](../../../spec/screens/05-medical-records-list.md)）。UI ステータス: 作成中=`draft` / 確定済=`finalized`。
- 子エンティティ（治療・検査・バイタル等）の書込と確定処理は行ロックで直列化される（`LockByIDForUpdate`。確定と同時書込のレースで確定済みカルテに子データが混入しないこと）。
- カルテ確定（finalize）・訂正追記（addendum）の監査は本体操作と同一トランザクションで fail-closed に記録され、監査失敗時は本体操作もロールバックされる。カルテ作成（create）・通常更新（update）の監査は臨床継続性のため best-effort を維持する（`medical_record_crud.go`・`medical_record_addendum_service.go`）。削除監査も best-effort。
- 会計完了によるカルテの自動確定は存在しない（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）— 確定は明示操作のみ。
- カルテ一覧のステータスフィルタ（作成中/確定済）で確定後の当該カルテが抽出できること（[05 §1.1](../../../spec/screens/05-medical-records-list.md)）。
- フローティング「削除」は問診タブかつ未確定時のみ（詳細）。一覧の削除メニューは確定状態で隠していない点に注意（BE が最終防衛）。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定済みカルテの「治療」タブで明細の追加・削除を試行する | UI 上は `<fieldset disabled>` によりタブ内の追加/削除ボタンが無効化され操作できない（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）。直接 API を叩いた場合、治療（`treatment_service.go`）・検査・バイタル・処方・健診・画像の各子リソースはバックエンド側でも確定済み親カルテへの書込を拒否する |
| A2 | medical-records の edit 権限を持たないロールで確定済み/作成中カルテを開く | 保存操作が権限制御で不可（[06 概要](../../../spec/screens/06-medical-records-form.md): 保存可否はコンポーネント内の権限制御）。fieldset は `!canSubmit` でも disabled |

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - 確定ダイアログ文言を実装全文に更新
  - 編集拒否メッセージに addendum 誘導句を追加（409）
  - 削除拒否メッセージを「確定済みまたは下書き以外の診療記録は削除できません」に修正
  - 一覧は確定済でも削除メニューが出うる / 詳細は非表示、と UI 差を明記

