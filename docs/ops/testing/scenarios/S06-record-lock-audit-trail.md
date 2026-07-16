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
| 4 | カルテを確定（Lock）する: 画面上に確定ボタンは存在しないため、ステータスを finalized に更新する PATCH を送信する（**USER が API クライアントで実施** — [06 §2.3](../../../spec/screens/06-medical-records-form.md)） | カルテ一覧のステータスが「確定済」に変わる（[05 §1.2](../../../spec/screens/05-medical-records-list.md)） |
| 5 | 確定済みカルテを開き、属性を変更して保存を試行する | バックエンドが更新を 409 で拒否する（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）。エラー「確定済みカルテは編集できません」（[05 §3.1](../../../spec/screens/05-medical-records-list.md)） |
| 6 | 確定済みカルテを一覧から削除しようとする | 拒否される: 「確定済みの診療記録は削除できません」（[05 §3.1](../../../spec/screens/05-medical-records-list.md)） |
| 7 | 確定済みカルテへ訂正追記（addendum）を行う | 訂正は追記のみ許可される（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）。【要実測】追記の UI 経路 |
| 8 | 確定の解除（確定済 → 作成中へ戻す）を試行する | 【要実測】確定解除の可否と要求権限。仕様文書に明記なし |
| 9 | 手順 1〜4 の操作について `audit_logs` を DB で確認する（**USER 実施**。例: resource がカルテ関連の行を時刻降順で参照） | 作成（create）・更新（update）・確定の各操作が、操作者（actor_id）・時刻（created_at）・変更内容（old/new_value）付きで記録されている（[specification.md §2.1](../../../spec/specification.md)「全テーブルの CRUD を audit_logs で追跡」） |

## 確認観点

- 確定ロックは DB 制約ではなくサービス層のガード（`backend/internal/service/medical_record_crud.go` — [05 §3.1](../../../spec/screens/05-medical-records-list.md)）。
- 子エンティティ（治療・検査・バイタル等）の書込と確定処理は行ロックで直列化される（`LockByIDForUpdate`。確定と同時書込のレースで確定済みカルテに子データが混入しないこと）。
- `audit_logs` の記録は本体操作と同一トランザクションで行われること（監査欠落のサイレント失敗がないこと）。
- 会計完了によるカルテの自動確定は存在しない（[06 §2.3](../../../spec/screens/06-medical-records-form.md)）— 確定は明示操作のみ。
- カルテ一覧のステータスフィルタ（作成中/確定済）で確定後の当該カルテが抽出できること（[05 §1.1](../../../spec/screens/05-medical-records-list.md)）。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定済みカルテの「治療」タブで明細の追加・削除を試行する | 【要実測】子リソース書込も拒否されること（[06 §2.3](../../../spec/screens/06-medical-records-form.md) は本体更新の拒否を明記。子リソースへの適用範囲は明記なし） |
| A2 | medical-records の edit 権限を持たないロールで確定済み/作成中カルテを開く | 保存操作が権限制御で不可（[06 概要](../../../spec/screens/06-medical-records-form.md): 保存可否はコンポーネント内の権限制御） |
