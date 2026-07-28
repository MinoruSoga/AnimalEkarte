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
| 4 | カルテを確定（Lock）する: 画面右下のフローティングアクションの「確定する」ボタン（`MedicalRecordFloatingActions`。編集権限あり・保存済み・未確定の場合のみ表示）をクリックし、`MedicalRecordFinalizeDialog` で「確定後はカルテを編集できなくなります…この操作は元に戻せません」の警告を確認した上で「確定する」を押す（[06 §2.3](../../../spec/screens/06-medical-records-form.md)） | カルテ一覧のステータスが「確定済」に変わる（[05 §1.2](../../../spec/screens/05-medical-records-list.md)）。カルテ詳細ヘッダーに「確定済」バッジが表示され、「確定する」ボタンは消える |
| 5 | 確定済みカルテを開き、属性を変更して保存を試行する | 主要タブの入力欄・追加/削除ボタンが UI 上で無効化されており（`<fieldset disabled>` による一括ガード、[06 §2.3](../../../spec/screens/06-medical-records-form.md)）操作自体を試行できない。仮に直接 API を叩いた場合はバックエンドが更新を 409 で拒否する。エラー「確定済みカルテは編集できません」（[05 §3.1](../../../spec/screens/05-medical-records-list.md)） |
| 6 | 確定済みカルテを一覧から削除しようとする、またはカルテ詳細のフローティングアクションを確認する | 一覧・詳細いずれも削除操作の導線が表示されない（`MedicalRecordFloatingActions` は確定済みで削除ボタンを非表示）。直接 API を叩いた場合は「確定済みの診療記録は削除できません」で拒否される（[05 §3.1](../../../spec/screens/05-medical-records-list.md)） |
| 7 | 確定済みカルテの各タブ下部「追記する」ボタンから訂正追記（addendum）を行う（修正内容・修正理由が必須のモーダル — 2026-07-17 実測で確定） | 保存後、同タブ下部に時刻・スタッフ付きの時系列リストとして表示され、リロード後も永続する（[06 §2.3](../../../spec/screens/06-medical-records-form.md)） |
| 8 | 確定の解除（確定済 → 作成中へ戻す）を試行する | 解除（unfinalize）API はバックエンドに存在しない（`medical_record_crud.go` に該当メソッドなし）。確定は一方向遷移であり、確定後の修正経路は訂正追記（addendum）のみ。解除機能が必要な場合は GAP-1 とは別に要件を起票する（2026-07-16 Fable 代理決裁） |
| 9 | 手順 1〜4 の操作について `audit_logs` を DB で確認する（**USER 実施**。例: resource がカルテ関連の行を時刻降順で参照） | 作成（create）・更新（update）・確定の各操作が、操作者（actor_id）・時刻（created_at）・変更内容（old/new_value）付きで記録されている（[specification.md §2.1](../../../spec/specification.md)「全テーブルの CRUD を audit_logs で追跡」） |

## 確認観点

- 確定ロックは DB 制約ではなくサービス層のガード（`backend/internal/medicalrecord/medical_record_crud.go` — [05 §3.1](../../../spec/screens/05-medical-records-list.md)）。
- 子エンティティ（治療・検査・バイタル等）の書込と確定処理は行ロックで直列化される（`LockByIDForUpdate`。確定と同時書込のレースで確定済みカルテに子データが混入しないこと）。
- カルテ確定（finalize）・訂正追記（addendum）の監査は本体操作と同一トランザクションで fail-closed に記録され、監査失敗時は本体操作もロールバックされる。カルテ作成（create）・通常更新（update）の監査は臨床継続性のため best-effort を維持する（`medical_record_crud.go`・`medical_record_addendum_service.go`）。
- 会計完了によるカルテの自動確定は存在しない（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）— 確定は明示操作のみ。
- カルテ一覧のステータスフィルタ（作成中/確定済）で確定後の当該カルテが抽出できること（[05 §1.1](../../../spec/screens/05-medical-records-list.md)）。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定済みカルテの「治療」タブで明細の追加・削除を試行する | UI 上は `<fieldset disabled>` によりタブ内の追加/削除ボタンが無効化され操作できない（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）。直接 API を叩いた場合、治療（`treatment_service.go`）・検査・バイタル・処方・健診・画像の各子リソースはバックエンド側でも確定済み親カルテへの書込を拒否する |
| A2 | medical-records の edit 権限を持たないロールで確定済み/作成中カルテを開く | 保存操作が権限制御で不可（[06 概要](../../../spec/screens/06-medical-records-form.md): 保存可否はコンポーネント内の権限制御） |
