# AnimalEkarte — TODO

> 更新: 2026-07-23（BE-refactor.md 未解消技術債の調査確定分 7 件を起票）

## 運用

- 本書は、エージェントが直ちに着手できる未完了タスクの台帳とする。
- タスクは「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で追加する。
- 対応済みセクションは削除し、完了記録は git 履歴を正本とする。
- GitHub Issue と対応するタスクは Issue の state を実測し、Issue 一覧を本書へ重複掲出しない。

## 正本の境界

| 内容 | 正本 |
|------|------|
| 着手可能な実装タスク | 本書の「個別タスク詳細」 |
| GitHub Issue の state・一覧 | GitHub Issues |
| BE9 構造移行・進捗・検証証跡 | [`BE-refactor.md`](BE-refactor.md) |
| FE デザイン準拠・リファクタリング計画 | [`FE-refactor.md`](FE-refactor.md) |
| 今フェーズで着手しない事項 | [`phase2.html`](phase2.html) |
| 着手保留・任意検証の BE 技術債 | [`BE-pending.md`](BE-pending.md) |
| PO 判断・USER 実操作・P0 ブロッカー | [`q&a.html`](q&a.html) |

## 個別タスク詳細

> 2026-07-23 起票: BE-refactor.md「未解消の技術債」の調査確定分（21エージェント並列調査+敵対検証済み）。全て BE9 構造移行と分離した独立 behavior batch とし、baseline test を RED→GREEN で先に固定する。DEC-11〜13 は同日 Fable 代理決裁で確定（全て案A・裁定全文 = q&a.html・実装 = BUG-426〜428）。死亡ペットガードは SD-10 裁定により phase2.html 一括監査へ統合済み（本書へは起票しない）。

### BUG-421: trimming course 使用数カウントの details.clinic_id 述語欠落（LOW・防御多層化）

- 事象: `CountUsageByTrimmingCourseID`（`backend/internal/trimming/trimming_course_repository.go:74-84`）が `appointments.clinic_id` のみ制約し、`appointment_trimming_details.clinic_id` に直接述語を置かない。
- 調査確定: 書込3系統（trimming domain / LINE・LIFF / csvimport cutover）は全て clinic 強制済みで現行の混入経路なし。DB 層は単一列 FK のみ・RLS は owner 接続非適用のため破損行は表現可能（defense-in-depth 欠落のみ）。
- 修正: count クエリへ `Where("appointment_trimming_details.clinic_id = ?", clinicID)` を1行追加。`CountUsageByTrimmingOptionID` は options に clinic_id 列が無く JOIN 依存のままが正・変更不要。複合 FK 化は次回 trimming 系 DDL 変更に相乗りする別判断。
- 受け入れ: 既存 `TestTrimmingCourseRepository_CountUsageByTrimmingCourseID` へ「detail.clinic_id が親 appointment と異なる破損行を数えない」subtest（db.Create 直接挿入）を追加し RED→GREEN。

### BUG-422: trimming CUD に actor-aware な同一 tx 監査がない（MEDIUM）

- 事象: trimming の全書込経路（Create / 既存予約への detail 追加 / Update / Delete、`backend/internal/trimming/trimming_service.go`）に audit_logs への永続監査がない。slog のみで actor 情報なし。特に Delete は RBAC を通れば痕跡ゼロ。
- 修正（既存パターンの完全再利用・新規機構なし）: ① model に `AuditActionTrimmingCreate/Update/Delete` + `AuditResourceTrimming` 追加 ② `trimming/ports.go` へ medicalrecord/service_deps.go と同型の consumer-side `AuditEntry`+`AuditTxLogger` ③ 4 つの WithTx closure 内で fail-closed `LogEntryTx`（監査失敗で本体 rollback）④ `cmd/api/main.go` に adapter 配線。Delete を最優先に同一 PR で対称実装。
- 受け入れ: `hospitalization_discharge_audit_test.go` と同型の同一 tx fail-closed テスト + 4 経路の監査行 field 検証。trimming マスタ 3 種の CUD 監査は本タスク対象外（要否は別途起票判断）。

### BUG-423: trimming マスタ削除の CountUsage→soft delete 非原子 TOCTOU（MEDIUM）

- 事象: course / option / course_type の Delete が FindByID→CountUsage→Delete を transactor なしの独立 3 クエリで実行（`trimming_course_service.go:175-193` 等）。確認と削除の間に新規使用（staff API・公衆到達の LINE 予約）が commit されると使用中マスタが消え、FK ON DELETE SET NULL で detail.course_id が黙って NULL 化し得る。
- 修正: tx 内 delete-then-recheck 方式 — 各 service.Delete へ Transactor 注入、WithTx 内で Delete（排他行ロック）→ CountUsage 再実行 → >0 なら Conflict rollback。注意: course_type は write 側（course Create/Update の tx 化 + type FindByID の SHARE lock 分岐追加）もセットでないと閉じない。`DeleteScopedByID` の DBOrTx 化は trimming 限定で先行（repohelpers 全域への横展開は別タスク）。
- 受け入れ: 実 DB table-driven TOCTOU テスト 3 マスタ分（明示 tx で使用行 INSERT 未 commit → Delete 発行 → commit 順序別に Conflict/成功を検証）RED→GREEN。

### BUG-424: trimming 予約検証の fail-open・tx 不参加・上限欠落 3 点（MEDIUM）

- (a) fail-open: `ValidateReservationTypeAvailableTime`（unavailableRepo==nil で全スキップ）と `CheckReservationTypeCapacity`（repo nil で即 nil）に対し、trimming service は course/option repo と異なり nil ガードなし（本番 DI は non-nil のため潜在パス・規約「依存欠落は fail-closed」違反）。修正 = trimming service 側へ fail-closed ガード追加。共有ヘルパ自体の fail-closed 化は全 consumer 棚卸し後の別判断。
- (b) unavailable-time read が `DBOrTx` 不参加で予約確定 tx 外 read（writer 側 CreateUnavailableTime/DeleteUnavailableTime も tx/lock なしの追加証拠あり）。修正 = repo FindAll の DBOrTx 化 + tx 参加テスト。
- (c) Update の `option_ids` に件数上限なし（Create は max=50）。修正 = `trimming_request.go:136` へ `binding:"max=50,dive"` 追加 + 境界テスト。clinic advisory booking lock を長時間占有する運用 DoS ベクタの封鎖。

### BUG-425: lstep PutMappingsForTag の非 transactional replace（MEDIUM）

- 事象: PUT /api/v1/lstep-tag-code-mappings/:tag_name が soft-delete 一括 + ループ Create を全て autocommit 単文で実行（`lstep_tag_code_mapping_service.go:63-`、repo は `r.db` 直参照で DBOrTx 不参加）。途中失敗で「旧設定全消し + 新設定先頭数件」が永続化され、tag sync batch（健康予防タグ判定）が部分集合を正として読むサイレント誤判定になる。
- 修正: ① repo の Create/SoftDelete 系 3 method を `repohelpers.DBOrTx` 化 ② service へ lifecycleTransactor と同型の consumer-side WithTx を注入し置換全体を包む ③ composition 1 行。Find 系を DBOrTx 化する場合は WithTx 外呼び出しの実測必須（既知知見）。delete〜create 間の並行読取窓が閉じることも成果に明記。
- 受け入れ: N 回目 Create で失敗するデコレータ repo による実 DB 統合テストで「部分更新が残らない」rollback を実証 RED→GREEN。

### TEST-ROUTES-01: 全 domain route 合成 smoke test（BE9-2F 着手前に導入）

- 事象: route 合成は `cmd/api/main.go` のみで、cross-package の path/method 衝突（gin path merge 共存中の /medical-records・/accountings 等）は per-domain snapshot が全 green のまま起動 panic まで検出不能（BUG-246 の教訓が BE9 の route 分散で失効中）。
- 実装（調査で実現性確認済み）: `backend/cmd/api/route_composition_smoke_test.go` 1 file（~100 行）。gin.TestMode で main.go と同順に旧 handler + manualarticle/medicalrecord/reservation/billing/lstep（+ RegisterLiffRoutes / RegisterWebhookRoutes）を nil 依存 + noop closure で登録し `assert.NotPanics`。DB・stub 不要（handler.New の svc nil 登録は既存 `TestRegisterRoutes_NoPanic` が実証済み）。

### FMT-BE-01: gofmt 未整形 4 file の独立 format batch

- 実測（2026-07-23・コンテナ内 `gofmt -l`）: `internal/handler/auth_session_test.go` / `internal/service/staff_service_account.go` / `internal/service/token_service.go` / `internal/service/token_service_test.go` の 4 本（BE-refactor.md 旧記載「7 本」は BE9 縦移動で 3 本解消済み・同 doc は訂正済み）。
- 実施: 4 file 限定 `gofmt -w` → `gofmt -l` 0 本確認 + scoped build。BE9 規約どおり独立 format batch として他変更と混ぜない。

### BUG-426: trimming completed 直接受理の閉鎖（DEC-11 裁定 = 案A）

- 裁定: q&a.html DEC-11（2026-07-23 Fable 代理決裁）。3 アプリ全域実測で completed 送信 consumer は会計完了（billing 正規経路）のみ — 閉鎖で壊れるフローなし。
- 実装: ① `trimming_request.go:66,123` の oneof から completed を除去 ② `reservation_intent_repository.go` の `validateTrimmingStatusTransition` へ completed 遷移拒否 + `CreateForTrimming` へ completed 新規作成拒否（no_show 拒否 :454-456 と同型）③ `trimming_request_test.go:108-124` の completed 透過 assert を非 terminal 値へ変更 + 拒否テスト追加。
- 受け入れ: 拒否テスト RED→GREEN。会計経路 baseline（`TestReservationRepository_CompleteForAccounting` 系）green 維持。route snapshot 無変更。
- 後続（本タスクに含めない）: 一般予約 PATCH（`reservation_request.go:220`）の同型閉鎖は consumer 再実測を添えて別 batch で判断。

### BUG-427: checkup-sync owner_ids 上限 100 + actor 欠落時の即 abort（DEC-12 裁定 = 両案A）

- 裁定: q&a.html DEC-12（2026-07-23 Fable 代理決裁）。2 batch 構成で実施する。
- Batch 1（cap・FE 追随と同一 batch 必須）: `checkup_sync_request.go:146` へ max=100 + `api.yaml:8095` へ maxItems:100 + 境界テスト + FE checkup-sync 画面の分割送信 or 選択上限 UX 追随（BE 先行で 422 を出さないこと）。
- Batch 2（401 即 abort・挙動変更）: `checkup_sync_handler.go` GET/POST の ExtractStaffID !ok で即 return 化 + characterization テスト（handler_test :185 の「401 でも続行」）の期待反転。
- 受け入れ: 両 batch とも RED→GREEN。401 後の 200 二重書込が機構的に不能であることをテストで固定。

### BUG-428: 予約キャンセル後 draft cleanup の可視化 + LIFF/admin 経路配線（DEC-13 裁定 = 案A）

- 裁定: q&a.html DEC-13（2026-07-23 Fable 代理決裁）。
- 実装: ① `medical_record_auto_create.go` の cleanup 失敗を WARN→ERROR 格上げ + 監査行追加（意図的な見積依存 Conflict と DB 障害の文言分離）② `liff_service_reservations.go` / `appointment_admin_service.go` へ medicalRecordAutoCreator view（`service_deps.go:42-45`）を注入し同じ best-effort cleanup を配線（`cmd/api/main.go` DI 変更含む）③ 配線テスト（キャンセル時に呼ばれること）+ 失敗可視化テスト。
- 受け入れ: RED→GREEN。回復経路は既存の手動 DELETE /medical-records/:id を維持（新機構なし）。
- 再開条件の記録: cleanup 失敗の監査行が月 1 件以上恒常観測されたら実測を添えて案B/C を再起案（PERF-AUDIT-TX P2 と同一基準）。
