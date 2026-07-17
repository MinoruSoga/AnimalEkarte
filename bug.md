# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。

## Open

### BUG-413:【要監視・#250後に判定】予防接種・トリミング一覧が同型のページネーション不可視化リスクを抱える（現状 seed 0件で無症状）

- **経緯**: BUG-412（在庫一覧の偽ページネーション）の派生監査（2026-07-17）で発見。BUG-412 は 2026-07-17 に修正済み（`InventoryList`/`use-inventory.ts`/`frontend/src/features/inventory/api/inventory.ts` をサーバサイド page/limit 転送 + 実 total 消費に変更。backend `inventory_repository.go` は元々 clinic-scoped 実 COUNT を WHERE 適用後に返しており変更不要だった）。本エントリは BUG-412 調査で判明した**同一パターンの潜在リスク**のうち、今回は修正しないと判断したものを引き続き追跡する。
- **VaccinationList.tsx（予防接種一覧）**: `use-vaccinations.ts` → `get-vaccinations.ts` の `getVaccinations` は `start_date`/`end_date` を日付フィルタ使用時のみ送る任意パラメータで、未指定時は無条件で backend `parsePagination` の既定 `limit=20` に落ちる（date-scoped ではない）。`backend/migrations/seeds/003_demo/vaccinations.csv` は現状 0 件（ヘッダのみ）で無症状・検証不能。
- **TrimmingList.tsx（トリミング一覧）**: `use-trimming-records.ts` → `get-trimmings.ts` の `getTrimmings` は明示的に `page:1, limit:HISTORY_FETCH_LIMIT`（`frontend/src/config/fetch-limits.ts` で `100` = backend `defaultMaxPaginationLimit`）を送信しており `limit=20` の欠陥ではないが、1リクエストの上限100件を超えた場合の継続取得導線が無い。トリミング実績は `reservation_type.category='trimming'` の予約（`appointments.csv`）であり、現状 0 件で無症状・検証不能。
- **今回修正しない理由**: 両画面とも seed が 0 件のため、ページネーション化しても page=2 が別レコードを返すことを実測で反証できず、「直った」ことを検証できない。加えて両フックとも医師/ステータス/種別等のクライアントサイドフィルタを持ち、BUG-411/412 と同様「フィルタがページ内スコープに退行しないか」の再設計が必要（Trimming はさらに limit=100→カーソル/オフセット継続取得という追加拡張が要る）。検証不能な状態での臨床データ一覧の書き換えは、修正漏れより悪い結果（未検証の新規回帰）を生みうるため見送った。
- **要監視**: #250（本番データ移行）でこの2画面の実件数が判明した時点で、実件数が limit（Vaccination=20 / Trimming=100）を超えるか再確認し、超える場合は本エントリを起票根拠に BUG-411/412 と同型の修正を行う。#250 の受け入れ条件にこの再確認を明示的に含めること。
- **発見**: 2026-07-17（BUG-412 対応中の派生調査、コーディネーターが claim を再検証済み）。

### BUG-414:【LOW・既存バグ】在庫一覧のカテゴリ/ステータス絞り込みで「以外」条件（is_not）が無効化されず、単に無条件表示になる

- **症状**: `InventoryList.tsx` の `category`/`statusFilterEntry` 導出ロジック（category は L112-116, status は L117-120 付近）は `condition === "is"` のときのみ値をサーバへ渡し、それ以外（`is_not` 含む）は無条件に `"all"` へフォールバックする。コメントは「"is" 条件のみサーバーサイド、"is_not" はクライアントサイドで処理」と書かれているが、`is_not` を除外フィルタとして適用するクライアントサイドのロジックはコード上どこにも存在しない（`use-inventory.ts` にも `InventoryList.tsx` にも無し）。結果、ユーザーが「カテゴリが医薬品以外」等を選択しても**何も絞り込まれず全件が表示される**（サイレントな無絞り込み）。
- **調査の起点**: `frontend/src/features/inventory/routes/InventoryList.tsx` のコメント（102行目付近）と実装の乖離。修正方針は、コメント通りクライアントサイドで `is_not` の除外フィルタを実装するか、"is_not" 自体を `PropertyFilter` の `conditions` から除外するか（`CONDITIONS_NO_EMPTY` の内容を確認）のいずれか。
- **発見**: 2026-07-17（BUG-412 対応中の investigate subagent による調査、コード実読で確認）。

### BUG-408:【設計判断待ち】予防接種フォームのワクチン選択に動物種(species)フィルタが無い

- **経緯**: 旧 BUG-408/401（予防接種フォームが `VACCINE_TYPE_ITEMS` というハードコード2択で、ワクチンマスタを一切クエリせず、誤った vaccine_id を永続化していたデータ破損）は **2026-07-17 に修正済み**。`VaccinationFormPanels.tsx` が姉妹フォーム（`MedicalRecordVaccination.tsx`）と同じ `useGetAllVaccinesMaster()` パターンで実マスタを `isActive` フィルタ付きでクエリするよう変更し、`use-vaccination-form.ts` の `VACCINE_SCHEDULE_MAP`（"1"/"2"固定キー）も選択ワクチンの `interval` フィールドから次回予定を導出する方式に置き換えた（回帰テスト: `use-vaccination-form.test.ts` の「BUG-401」節、`VaccinationFormPanels.test.tsx`）。
- **残存する設計判断（未決・本エントリが追跡するのはこれのみ）**: マスタ項目を選択ペットの動物種でどう絞り込むか。姉妹フォームも動物種フィルタを持たないため、上記修正は parity に留め新規リスクを増やしていないが、犬患者に猫用ワクチン（またはその逆）が選択可能な状態そのものは残っている。選択ペットのデータ型に `species` が含まれておらず型拡張が必要になる可能性がある。
- **調査の起点**: `usePetSelection` が返す pet オブジェクトの型定義、`VaccineItem.species` フィールド（`frontend/src/lib/transforms/treatment.ts`）との突合。
- **発見**: 2026-07-17（BUG-401 調査中）。vaccine_id 誤保存自体は同日中に修正済み。

### BUG-410:【要検証】カルテ編集保存時に構造化診断（diagnosis1/2）が再投入されず上書きされる可能性

- **症状（未確認・要検証）**: BUG-406 の修正で問診（Inquiry）は再読込時に正しく復元されるようになったが、`use-apply-medical-record.ts` / `transforms.ts` は assessment/plan の自由記述は再投入する一方、diagnosis1/2 の category/name ID を再投入していない（`transformMedicalRecord` がそもそも露出しない）。保存アクションが未ロードの診断マスタ state を送信する実装であれば、通常の編集保存で保存済みの構造化診断が意図せず上書き・クリアされる可能性がある。
- **調査の起点**: カルテ編集保存の送信ペイロードが、未ロードの diagnosis1/2 を null 化して送っていないかを確認する。
- **発見**: 2026-07-17（healthcare-reviewer による BUG-406 修正の独立監査で指摘、MEDIUM）。

## 直近クローズ（次回整理で削除）

- **BUG-409**（ペット死亡ステータスの二重管理が外側フォーム経由で再発しうる）: **修正済み 2026-07-18**（commit 74652f72）。再現テストで確認: `PetCareSection.tsx` の生死ラジオが `deceased_at`/監査ログと独立に `status` のみを書き換え可能で、①生存ペットで死亡ラジオ→保存 = status=死亡・deceased_at=null・監査ログ無し、②ダイアログ死亡登録後に生存ラジオ→保存 = status=生存・deceased_at残存、の2経路で不整合を実際に生成できた（`transformUpdatePetRequest`/backend `buildPetUpdate` いずれも deceased_at に触れない）。修正はラジオを現在値表示専用(disabled)にし、生死変更を監査付き `PetDeceasedRecordButton` 経由に一本化（transform への deceased_at 追加は無監査の第二死亡経路を新設するため不採用）。
- **BUG-404**（入院デイリー記録 GET/ケアログ POST 全 500）: **修正済み 2026-07-17**（commit 58c653df）。根因 = TIME 列を `time.Time` で Scan（書込成功・読取全滅）+ 永続テスト DB のスキーマドリフトがテストを素通りさせていた（自己修復 ALTER 追加済み）。次回シナリオ再実行で最終確認したら本行を削除。
