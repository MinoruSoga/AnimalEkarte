# S23: 複数医院所属 — 拠点スコープと他院操作の境界

> **目的**: 複数医院に所属する actor が拠点切替（`?clinics=`）で一覧を正しく絞り込み、他医院の行では操作が抑制され、非第一所属の医院で作成したカルテの帰属（`entered_by_clinic`）が正しく記録されることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 深い
> **仕様正本**: [specification.md §1.1](../../../spec/specification.md)・[screens/03-owners-list.md](../../../spec/screens/03-owners-list.md)・[screens/05-medical-records-list.md](../../../spec/screens/05-medical-records-list.md)。関連 tracker: `BUG2-MR-ENTERED-BY-CLINIC`・`SLACK-CROSS-CLINIC`（PO 裁定待ち）・Plane 他院検索調査。

## 前提条件

- ローカルの使い捨て環境で、actor が所属する **2 つ以上の clinic**（A = 第一所属、B = 別所属）と、actor が**所属しない** clinic C を用意する。
- clinic A/B にそれぞれ owner/pet/カルテの fixture。固定 ID・汎用 seed は仮定しない。
- actor は両医院で必要な権限（owners/medical-records 等）を持つ attached account。
- `SLACK-CROSS-CLINIC`（他院患者の検索・受付の境界）は PO 裁定待ち。本シナリオは**現行の拠点スコープ実装**の受入であり、境界変更を前提にしない。
- 試験後に作成した fixture を削除する。
- 依存シナリオ: S14（検索 AND と医院スコープ）・S13（identity-links）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | カルテ一覧を開き、拠点を clinic A のみにする | clinic A のカルテだけが出る。URL は `?clinics=` を反映し、単一の現在医院選択では `clinicIds` を送らない（`useClinicScope` 既定） |
| 2 | 拠点トグルで clinic B（非第一所属だが所属はある）を選択する | URL が `?clinics=clinicB` に変わり、その `clinicIds` が API へ送られて B のカルテだけが出る |
| 3 | clinic B のカルテ行を開く | 他医院行は `isOtherClinic` で編集・レポート・削除が抑制され、参照中心になる。行の `clinicId` が医院名で表示される（`clinicNameById`） |
| 4 | clinic B を選択した状態で新規カルテを作成する | 所属が確認された医院 B でカルテが作られ、レコードの `clinic_id` が現医院 A に置き換わらず B のまま記録される（`BUG2-MR-ENTERED-BY-CLINIC` の修正後挙動）。`entered_by` は B でのスタッフ所属が検証される（`AssertEnteredByActor`）。再読込で帰属が保持される |
| 5 | 拠点を複数（A+B）選択する | `isMultiClinic` で両医院のデータが出る。各行の医院が区別できる |
| 6 | 飼主一覧・他の一覧でも同じ `?clinics=` スコープが効くことを確認する | 拠点スコープは `useClinicScope` 経由で画面横断。画面ごとに勝手な絞り込みにならない |

## 確認観点

- 拠点スコープは `useClinicScope`（`frontend/src/hooks/`）が `?clinics=` → `selectedClinicIds`・`isMultiClinic`・`assignedClinics`・`clinicNameById` を提供する。クライアント側フィルタだけで他医院を隠す設計ではない。
- 他院行の操作抑制は `MedicalRecordsListPanels` の `isOtherClinic`（record.clinicId ≠ currentClinicId）で編集/削除等を抑止する。参照は許容、変更は現医院のみという現行仕様。
- カルテの `clinic_id` 帰属の保持は `BUG2-MR-ENTERED-BY-CLINIC` の修正対象。`entered_by` はスタッフ ID（FK）で、`AssertEnteredByActor` がそのレコードの医院でのスタッフ所属を検証する。非第一所属で作成しても帰属が現医院に上書きされないことを再読込で確認する。
- `SLACK-CROSS-CLINIC`（山梨の他院患者の受付等）は対象医院・権限範囲の PO 裁定待ち。本シナリオは現行の拠点スコープの受入で、境界変更を期待値にしない。
- API 直叩きで所属外の `clinic_ids` を送った場合の拒否は L2 認可の範囲。ここでは UI 経路のスコープと帰属を主対象とする。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | URL の `?clinics=` に所属しない clinic C の ID を直接入れて開く | clinic C のデータは返らない/表示されない（サーバー側で所属を検証）。UI が勝手に他院を見せない |
| A2 | 他医院行で編集・削除を試みる（ボタンが出ていれば押す） | 抑止される（非活性または拒否）。他院データを変更できない |
| A3 | clinic B でカルテ作成中に拠点を A に切り替える | 編集中のデータを失わない（`NavigationBlocker` 等）か、明確に破棄される。半端な帰属のカルテが残らない |
| A4 | `?clinics=` を空・不正値にして開く | デフォルト（現医院）へ安全にフォールバックし、全医院が見える状態にしない |

## 実装突合

- 変更サマリ:
  - `useClinicScope` の `?clinics=` / `selectedClinicIds` / `isMultiClinic` / `assignedClinics` / `clinicNameById` を現行コードと突合
  - `MedicalRecordsListPanels` の `isOtherClinic` による他院行抑制と、単一非現在医院選択時の `clinicIds` 送信を現行コード・テストと突合
  - `BUG2-MR-ENTERED-BY-CLINIC` の帰属保持（`clinic_id` + `entered_by` 所属検証 `AssertEnteredByActor`）を手順 4、`SLACK-CROSS-CLINIC` の PO 裁定待ちを前提条件に明記
