# BE-refactor.md — バックエンド リファクタリング（残オープン項目）

- **更新日**: 2026-07-13（第5期 A-1〜G-1 完了突合 + X-16② 実装完了を反映）
- **本書の規約**: 判断待ち・未実装フォローアップのみを記載する。対応済みの詳細・手順は git 履歴が正本（本書には残さない）。判断が出たら下記をそのまま実装単位にしてよい。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 判断待ち項目（PO / オーナー決定後に着手）

判断が出るまで実装しない。決定後は該当案の手順どおり実行する。

> **判断材料の再実測（2026-07-12 / HEAD `4f9550c9`）**: 全項目のアンカー・主張を並列調査 5 レーン + 手動 grep で現 HEAD に対して再検証した。初版の記述に**事実誤認 4 件**（X-16「封筒化が必要」「handler 先例 = owner_request.go」/ F6+H-3「唯一の category 述語」/ last_visit_date「pin 解消が動機」）が見つかり、各項目内で訂正済み。各項目は **決定者 → 判断質問（そのまま転送できる一文）→ 推奨 → 判断材料 → 判断肢別の実装手順 → 検証** で構成する。

### X-15. 状態トグル系 DELETE 4 ルートの "edit" 権限（P6 逸脱） — ✅ DONE (`d89df5f4`)

PO 判断: death 系 2 本 (a) P6 例外明文化（挙動保存・`.claude/refs/gin-architecture-compliance.md` +
`handler/CLAUDE.md` に注記済み）、lstep-opt-out 系 2 本 (c) ルート削除（`DeleteOwnerLstepOptOut`
handler・専用テスト・api.yaml オペレーションごと削除、FE 呼び出し 0 件を確認済み）。詳細は git 履歴参照。

### X-16. 健診一覧のページネーション — ✅ DONE（①②とも完了）

**① 健診期限アラート API 削除 — ✅ DONE (`e4a059ef`)**（消費者 0 件のため `GetCheckupAlerts` /
`GetAlerts` / `FindAlerts` / `toCheckupAlertsResponse` 系 + api.yaml オペレーション + 専用テストを削除済み。
詳細は git 履歴参照）。

**② 健診一覧の実サーバページング移行 — ✅ DONE (`7a3fb9e5` backend / `34b70f2f` api.yaml / `2968d2aa` frontend、2026-07-13)**:
repo `FindByClinicID` を `vaccination_repository` 同型の buildBase closure 化し `([]model.Checkup, int64, error)`
（データ + total）を返すよう改修（Preload 5 連鎖・clinicScope は維持）。service `ListByClinic` に
page/limit 入力・total 出力を追加。handler `ListGlobalCheckups` を `parsePagination` + 実 total の
`newPaginatedResponse` へ置換し、退化封筒（total=全件数・page=1 固定の見せかけページング）を解消。
`api.yaml` の checkup 一覧オペレーションに page/limit クエリパラメータを追記（api.yaml は他ワークストリーム
の未コミット変更と同居していたため hunk 単位で選択的にステージし別コミット化）。
FE 3 ファイル（`features/checkups/api/get-checkups.ts` / `api/types.ts` / `routes/CheckupsList.tsx`）を
medical-records 先例（`{data,total,page,limit}` 消費）で改修。検索・ソートは BE 側に対応クエリパラメータが
存在しないためサーバ移行不可と判断し、現在ページ内のクライアント処理として維持（FE 実装時の設計判断として
記録、着手時点の想定どおり）。詳細は git 履歴参照。

### F6 + H-3. Lstep 系死にコード 4 メソッドの keep/delete — ✅ DONE (`0d192642`)

オーナー判断: (a) 全削除。`BulkAddOwnerTag` / `SyncPetSpeciesTags` / `SyncSeniorTag` /
`FindOwnersByCategoryPurchaseDate` と専用型・実装ファイル（`lstep_tag_sync_pet_species.go` /
`_senior.go` は丸ごと削除）・mock・専用テストを削除済み。`SyncChronicConditionTags` /
`SyncOwnerAnimalClassificationTags` / `HasFoodPurchaseByOwnerSince` と共有ヘルパーは無変更。
復元は本コミットの revert で可能。詳細は git 履歴参照。

### M2. MeResponse.AvatarURL の三方向 drift — 解消済み（コード変更不要）

2026-07-12 再調査時点で、api.yaml（MeResponse）・FE（`transforms.ts` / `types/auth.ts` / fixture）
のいずれにも `avatar_url` / `avatarUrl` 宣言は現存しない（grep 0 件、pnpm キャッシュのみヒット）。
本項目作成時点の記述が既に解消済みの状態を反映していなかったための誤登録。対応不要。

### last_visit_date の wire 形式 — ✅ DONE (`f74c9242`)

LIFF 健康手帳 API の `last_visit_date` を RFC3339 datetime から date-only 文字列
（`In(time.Local).Format(time.DateOnly)`）へ変更済み。api.yaml の format も `date` に統一、
stale だった drift pin も削除。詳細は git 履歴参照。

### CI 構成: Backend ジョブの Lint 直列ゲート — ✅ DONE (`fb18fc18`)

ワークフローオーナー判断: 肢(i)（continue-on-error + 集約 step）採用。Lint step に `id: lint` +
`continue-on-error: true` を追加し、ジョブ末尾（Schema drift check 後）に `if: always()` +
`steps.lint.outcome != 'success'` で fail する `Verify lint outcome` 集約 step を新設。
Lint 失敗が Test を隠す fail-fast マスク構造を解消。集約 step 名はガードレール
`scripts/check-ci-step-order.sh` の gate/risk 判定と衝突しないことを確認済み（PASS 維持）。
詳細は git 履歴参照。

---

## オープン・フォローアップ（実装待ち / 人間作業）

### FOLLOWUP-X14A. MedicineID → DecreaseStock フォールバック（P1） — ✅ DONE (`2c101e40`)

`InventoryID == nil` 時に `MedicineID` を inventory id として `DecreaseStock`（clinic 非参加）へ渡す
フォールバックを廃止。減算は `InventoryID` が明示指定された場合のみ実施し、`DecreaseStock` の
シグネチャは無変更。回帰テスト（MedicineID-only で非呼出／InventoryID 指定時は従来どおり減算）を
追加済み。詳細は git 履歴参照。

### STG クロステナント監査 SQL

正本は `BE-pending.md`「任意検証（必須ではない）」節に集約済み（2026-07-12、reset 後の任意検証として
移管）。本書には残さない・重複させない。

### 別台帳

- `BE_todo.md` PERF-FOLLOWUP-05（password_reset goroutine の shutdown drain）— X-18 完了後の補完タスク。本ファイルのスコープ外。
- `BE-pending.md` — 次期送り・PO 判断待ち・reset 後の任意検証（STG クロステナント監査 SQL）の正本。
