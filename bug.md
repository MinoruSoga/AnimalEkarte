# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。

## Open

### BUG-408:【データ整合性疑い】予防接種フォームがワクチンマスタを一切参照せず、UIラベルと実際のvaccine_idが動物種で食い違う

- **症状**: 予防接種の新規登録フォーム（`frontend/src/features/vaccinations/components/VaccinationFormPanels.tsx`）のワクチン選択は `VACCINE_TYPE_ITEMS`（value:"1"=混合ワクチン, value:"2"=狂犬病ワクチン の2件のみ）というハードコード定数で、ワクチンマスタ（GET /api/v1/masters/vaccines、アクティブ11件）を一切クエリしていない。マスタの id=2 は「ワクチン猫」（猫用）だが、UI上のラベルは「狂犬病ワクチン」で犬患者にも選択可能 — 選択したラベルと実際に永続化される vaccine_id の動物種が食い違っている疑いが強い。
- **影響範囲調査が必要**: 既存の犬患者ワクチン記録に誤った vaccine_id が入っている可能性がある。
- **未決の設計判断**（実装前に責任者の決定が必要）: (a) マスタ項目を選択ペットの動物種でどう絞り込むか（選択ペットのデータ型に species が含まれておらず型拡張が必要になる可能性）。(b) `use-vaccination-form.ts` の `VACCINE_SCHEDULE_MAP`（"1"/"2"の2値のみで次回接種予定を自動計算）を実マスタの interval 等メタデータとどう統合するか。カルテ内蔵フォーム（`features/medical-records/components/MedicalRecordVaccination.tsx`）は `useGetAllVaccinesMaster` でマスタを参照するが isActive のみでフィルタし動物種フィルタが無く、同じく誤った動物種のワクチンが選択できてしまう。
- **調査の起点**: BUG-401 調査中に発見（マスタ11件に対しフォームが2件しか出ない症状の根因はマスタ未参照そのものであり、「フィルタが厳しすぎる」ではなかった）。既存パターンをそのまま移植すると動物種フィルタが皆無になり臨床リスクを新設するため、単純なバグ修正ではなく機能設計として扱う。
- **発見**: 2026-07-17（BUG-401 調査中に判明。BUG-401 自体は本エントリへ統合しクローズ）。

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
