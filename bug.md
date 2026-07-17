# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。

## Open

### BUG-412:【HIGH・潜在的データ不可視化】在庫一覧が backend デフォルト limit=20 のみ取得しページネーション未対応（既にローカルデモデータで顕在化）／予防接種・トリミング一覧にも同型の潜在リスク

- **症状**: `InventoryList.tsx`（在庫管理）の取得経路 `use-inventory.ts` → `frontend/src/features/inventory/api/inventory.ts` の `getInventoryItems`/`GetInventoryItemsParams` は `category`/`status` のみを送り `page`/`limit` を一切送らない。backend `inventory_handler.go` の `ListInventory` は `parsePagination`（未指定時 `limit=20`）を適用するため、20件を超える在庫は一覧・フィルタ・検索から不可視化する。BUG-411（会計・検査一覧の偽ページネーション、2026-07-17 修正済み）と同型のバグ。
- **実測件数（ローカル seed）**: `backend/migrations/seeds/003_demo/inventory_items.csv` = **30件**（ヘッダ除く実データ行数で確認）。30>20のため、**ローカルデモデータの時点で既に10件が一覧に出ない**。旧 BUG-411 エントリが「low risk」と判定した根拠（"inventory_items=30件、実データ小さいため相対的に低リスク"）は自己矛盾していた（30自体が既に既定 limit=20 を超えている）。BUG-411 対応中の派生監査（2026-07-17）で発覚。
- **関連する潜在リスク（同一取得パターン。現状データ0件のため無症状だが、旧 bug.md の「date-scoped だから安全」という前提自体が誤り）**:
  - `VaccinationList.tsx`（予防接種一覧）: `use-vaccinations.ts` → `get-vaccinations.ts` の `getVaccinations` は `start_date`/`end_date` をユーザーが日付フィルタを使った時だけ送る任意パラメータで、未指定時は無条件で `parsePagination` の既定 `limit=20` に落ちる（date-scoped ではない）。seed 0件のため今は無症状。
  - `TrimmingList.tsx`（トリミング一覧）: `use-trimming-records.ts` → `get-trimmings.ts` の `getTrimmings` は明示的に `page:1, limit:HISTORY_FETCH_LIMIT`（`frontend/src/config/fetch-limits.ts` で `100` = backend `defaultMaxPaginationLimit`）を送信しており `limit=20` のバグではない。ただし1リクエストで取得できる上限が100件で、それ以降のページネーション導線が無い。seed 0件のため今は無症状だが、100件超のトリミング実績が蓄積すると同型の不可視化が起きる。
- **修正方針**: BUG-411（会計・検査、#266 owners と同型）で確立したサーバサイド page/limit 転送パターンをそのまま踏襲する。`InventoryList` を優先（既に顕在化）。`VaccinationList`/`TrimmingList` は本番データ移行（#250）後の実件数を見て優先度判断（`TrimmingList` は limit=100→カーソル/オフセット継続取得への拡張が別途必要）。
- **発見**: 2026-07-17（BUG-411 対応中の派生監査。be411-other-screens-audit subagent による調査、コーディネーターが claim を再検証済み）。

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
