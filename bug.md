# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。

## Open

### BUG-411:【CRITICAL・データ喪失リスク】会計・検査一覧が backend デフォルト limit=20 で切られた配列に対し偽のクライアントページネーションを行い、フルデータ規模で大半のレコードが不可視化する

- **症状**: `AccountingList.tsx`（会計一覧）と `ExaminationsList.tsx`（検査一覧）は `useGetAccountings`/`useGetExaminations` 呼び出しで `page`/`limit` を一切送らない。バックエンドの `parsePagination`（`backend/internal/handler/query_helpers.go:27`）は未指定時 `limit=20` をデフォルト適用するため、実際には常に「最新（もしくは任意順）20件」だけが返る。フロント側はこの20件配列を `usePagination()`（`frontend/src/hooks/use-pagination.ts`）でクライアントページネーションするため `totalPages` は常に1、`Pagination` コンポーネントは非表示、件数表示も「20件」等の**偽の全件数**を示す。ユーザーには「全件表示されている」ように見えるが、実データの大半が一覧にもフィルタにも検索にも一切現れない。
- **実測件数（ローカルフルデータ）**: `billings` テーブル **392,105件**（会計）、`exams` テーブル **14,533件**（検査） — いずれも表示されるのは先頭20件のみ。#266 のような白画面クラッシュにはならない（over-fetch ではなく under-fetch）ため気づかれにくく、**会計データの大半が一覧上「存在しないもの」として扱われる**という点で #266 より発見しにくく深刻。
- **同型パターンの他候補**: `AccountingList`/`ExaminationsList` の他に、`InventoryList`/`VaccinationList`/`TrimmingList` も同じ `usePagination()` 消費側だが、これらは対応する取得APIが date-scoped もしくは実データ件数が小さい（`inventory_items`=30件、`vaccinations`=0件、ローカル実測）ため本チケットでは相対的に低リスクと判定。ただし本番の実データ移行（#250）後は再確認が必要。
- **調査の起点**: `frontend/src/features/accounting/api/get-accountings.ts`（`getAccountings` の `params` に `page`/`limit` が無い）、`frontend/src/hooks/use-examinations.ts`（同様）。修正方針は #266 の owners 対応（サーバサイド search/page/limit を URL 経由で loader/hook に転送し、`usePagination()` の client-side slicing をやめる）と同型。
- **発見**: 2026-07-17（#266 の fetch-limits 横断棚卸し中に発見。`frontend/src/config/fetch-limits.ts` 経由の各所は SD-18 で「直近100件表示」の可視キャップ表示済みで別枠 — 本件は fetch-limits.ts を経由しない、キャップ表示すら無い"偽ページネーション"のため独立起票）。

### BUG-408:【設計判断待ち】予防接種フォームのワクチン選択に動物種(species)フィルタが無い

- **経緯**: 旧 BUG-408/401（予防接種フォームが `VACCINE_TYPE_ITEMS` というハードコード2択で、ワクチンマスタを一切クエリせず、誤った vaccine_id を永続化していたデータ破損）は **2026-07-17 に修正済み**。`VaccinationFormPanels.tsx` が姉妹フォーム（`MedicalRecordVaccination.tsx`）と同じ `useGetAllVaccinesMaster()` パターンで実マスタを `isActive` フィルタ付きでクエリするよう変更し、`use-vaccination-form.ts` の `VACCINE_SCHEDULE_MAP`（"1"/"2"固定キー）も選択ワクチンの `interval` フィールドから次回予定を導出する方式に置き換えた（回帰テスト: `use-vaccination-form.test.ts` の「BUG-401」節、`VaccinationFormPanels.test.tsx`）。
- **残存する設計判断（未決・本エントリが追跡するのはこれのみ）**: マスタ項目を選択ペットの動物種でどう絞り込むか。姉妹フォームも動物種フィルタを持たないため、上記修正は parity に留め新規リスクを増やしていないが、犬患者に猫用ワクチン（またはその逆）が選択可能な状態そのものは残っている。選択ペットのデータ型に `species` が含まれておらず型拡張が必要になる可能性がある。
- **調査の起点**: `usePetSelection` が返す pet オブジェクトの型定義、`VaccineItem.species` フィールド（`frontend/src/lib/transforms/treatment.ts`）との突合。
- **発見**: 2026-07-17（BUG-401 調査中）。vaccine_id 誤保存自体は同日中に修正済み。

### BUG-409:【要検証】ペット死亡ステータスの二重管理が外側フォーム経由で再発しうる

- **症状（未確認・要検証）**: BUG-407 の修正で死亡登録サブダイアログの即時保存時に `status`/`deceased_at` を同一 Update で同期するようにしたが、`PetEditModal` の生死ラジオ（外側フォーム）が `deceased_at` と独立に `status` を変更できる可能性があり、外側の「更新」経由の保存が `status` のみ書き込み `deceased_at` を触らない実装であれば、再度 `status=生存` かつ `deceased_at` 残存という不整合状態を生みうる。
- **調査の起点**: `PetEditModal` の生死ラジオ状態と `transformUpdatePetRequest`（PATCH /pets/:id 系）が `status` と `deceased_at` の両方を一貫して扱っているかを確認する。
- **発見**: 2026-07-17（healthcare-reviewer による BUG-407 修正の独立監査で指摘、MEDIUM）。

### BUG-410:【要検証】カルテ編集保存時に構造化診断（diagnosis1/2）が再投入されず上書きされる可能性

- **症状（未確認・要検証）**: BUG-406 の修正で問診（Inquiry）は再読込時に正しく復元されるようになったが、`use-apply-medical-record.ts` / `transforms.ts` は assessment/plan の自由記述は再投入する一方、diagnosis1/2 の category/name ID を再投入していない（`transformMedicalRecord` がそもそも露出しない）。保存アクションが未ロードの診断マスタ state を送信する実装であれば、通常の編集保存で保存済みの構造化診断が意図せず上書き・クリアされる可能性がある。
- **調査の起点**: カルテ編集保存の送信ペイロードが、未ロードの diagnosis1/2 を null 化して送っていないかを確認する。
- **発見**: 2026-07-17（healthcare-reviewer による BUG-406 修正の独立監査で指摘、MEDIUM）。

## 直近クローズ（次回整理で削除）

- **BUG-404**（入院デイリー記録 GET/ケアログ POST 全 500）: **修正済み 2026-07-17**（commit 58c653df）。根因 = TIME 列を `time.Time` で Scan（書込成功・読取全滅）+ 永続テスト DB のスキーマドリフトがテストを素通りさせていた（自己修復 ALTER 追加済み）。次回シナリオ再実行で最終確認したら本行を削除。
