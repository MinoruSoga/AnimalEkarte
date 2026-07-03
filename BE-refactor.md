# バックエンド リファクタリング計画

- **作成日**: 2026-07-02
- **最終更新**: 2026-07-04
- **対象**: `backend/`（Go 1.25 / Gin / GORM / handler -> service -> repository）
- **状態**: 完了

## 結論

本計画で扱っていた BE-refactor 項目は対応済み。

完了済み項目の詳細な実施手順・状態メモは、重複した履歴情報としてこの文書から削除した。実装内容の正本は現在のコード、テスト、CI 設定、git 履歴とする。

主な完了コミット:

- `5f0fd548` `fix: BE-refactor R1-2/R3-2/R3-5 完了分をcommit`
- `61b85d7a` `fix: CI失敗2件を解消 (dbOrTx inventory allowlist / codegen drift)`

最終確認:

- GitHub Actions push event run `28672433856`
- Security Scan / Performance Tests / E2E Tests / CI: success
- Backend job: Build / Lint / Test / Coverage ratchet / Schema drift check: PASS

## 削除方針

完了済みの個別項目（R1-1〜R3-7）の作業計画、手順、検証メモは削除済み。今後この文書では、完了済み作業の再掲ではなく、BE-refactor 本体から切り離したフォローアップだけを管理する。

## 現在残すフォローアップ

以下は BE-refactor 本体の未対応ではなく、別トラックで扱う項目。

### `exam_results` の DB レベル複合 FK

`exam_results` と参照先 `exam_type_fields` はどちらも `clinic_id` 列を持たず、clinic は `exam_type_fields -> exam_types -> clinics` の2段先にある。

そのため、`checkup_field_results` と同じ additive migration では `(id, clinic_id)` 複合 FK を張れない。対応する場合は `clinic_id` 列追加と backfill を伴う非 additive な schema change になるため、PO / architect 判断の別タスクとして扱う。

現行防御はアプリ層の `FindByID(clinicID)` ガードで維持する。

### `break_hours` 実データ監査

R1-3 の fail-closed 化は実装済み。ただし STG / 本番の既存 `break_hours` データが想定形式と異なる場合、該当 clinic の LINE 予約が拒否される可能性がある。

STG / 本番 DB 監査はこの計画の自動実行対象外。実施する場合は、承認済み read-only 経路で別タスクとして行う。

### `pwreset` 30s timeout

`pwreset` が効かない・30s timeout する問題は挙動修正トラックであり、リファクタ対象外。

追跡先:

- `docs/tasks/open/PERF-FOLLOWUP-05-pwreset-30s-timeout-unreliable.md`

## 今後の扱い

この文書は完了済み計画のアーカイブとして残す。新しい backend リファクタ項目が発生した場合は、この文書へ完了済み項目を再追加せず、別のタスク文書または issue として起票する。
