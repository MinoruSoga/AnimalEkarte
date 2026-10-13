# Red/Blue ラウンド2 — TENANT-1 (clinic 分離) / reservation・billing・inventory・hospitalization 系 / ローカル Docker スタック

- 日付: 2026-10-09
- 対象: ローカル Docker スタック API (`http://127.0.0.1:8080`、main worktree 既定ポート。seed データ)
- invariant: TENANT-1（clinic isolation）+ ラウンド1 T5/T11 回帰観察
- カタログ: attack-catalog.md A 項
- 攻撃者起点: clinic 1（八王子病院）一般スタッフ（獣医師・高橋 純子）の認証セッションのみ所持
- 防御対象: clinic 2（城東センター病院）のリソース
- 備考（2026-10-09 実測）: STG バンドルからデモパスワード定数が消失していた（`SEEDLOGIN_DEMO_PASSWORD` 未注入の再デプロイと推定。fail-closed 方向のため脆弱性ではないが、STG のデモ自動入力は壊れている運用観察として記録）ため、スキル安全ゲート1の許可対象であるローカル Docker スタック（稼働中・ログインは公開定数）で実施する。

## セットアップ（実測済み）

| ID | 操作 | 結果 |
|---|---|---|
| S1 | 未認証 GET /api/v1/reservations, /api/v1/accountings | 401（fail-closed 確認） |
| S2 | clinic A staff（stg-staff-10000003）login | 200, cookie 取得 |
| S3 | clinic B staff（stg-staff-20000003）login | 200, cookie 取得 |
| S4 | B: GET /api/v1/reservations?limit=3 | clinic 2 の reservation **11085643**（owner 10407914, pet 11002734） |
| S5 | B: GET /api/v1/accountings?limit=3 | clinic 2 の accounting **11391301**（medical_record 11398719, pet 11015313） |
| S6 | B: GET /api/v1/estimates?limit=3 | clinic 2 の estimate **11000330**（draft） |
| S7 | B: GET /api/v1/inventory?limit=3 | clinic 2 の inventory item ID を取得 |
| S8 | B: POST /api/v1/hospitalizations（B owner/pet で自作・seed に B の入院が無いため） | 201 → B_hosp_id（**書込み系攻撃の標的はこの自作住院のみに限定**） |
| S9 | A: GET /api/v1/hospitalizations?limit=2 | clinic 1 の hospitalization id=2 が返る（A 側 baseline） |

## 試験マトリクス（clinic A トークンで clinic B リソース ID を攻撃）

合格シグナルは**実行前に**確定済み。不合格 = finding 候補。

| ID | エンドポイント | 経路 | 期待合格シグナル | 結果 |
|---|---|---|---|---|
| T1 | GET /api/v1/reservations/{11085643} | path param | 404、B reservation の内容を返さない | PENDING |
| T2 | GET /api/v1/clinics/2/reservations | **path 内 clinic_id** | 403/404。clinic 2 行を一切返さない | PENDING |
| T3 | GET /api/v1/accountings/{11391301} | path param | 404 | PENDING |
| T4 | GET /api/v1/accountings/{11391301}/refunds | 親 path param | 404 | PENDING |
| T5 | GET /api/v1/estimates/{11000330} | path param | 404 | PENDING |
| T6 | GET /api/v1/hospitalizations/{B_hosp_id} | path param（自作のみ） | 404 | PENDING |
| T7 | GET /api/v1/hospitalizations/{B_hosp_id}/daily-records | 親 path param | 404 | PENDING |
| T8 | GET /api/v1/inventory/{B_item_id} | path param | 404 | PENDING |
| T9 | GET /api/v1/inventory?limit=20（A として） | list scope | clinic 2 の item を含まない | PENDING |
| T10 | PATCH /api/v1/reservations/{11085643}（notes に marker） | path param | 404、B 再取得で notes 不変 | PENDING |
| T11 | PATCH /api/v1/inventory/{B_item_id}（数量 marker） | path param | 4xx、B 再取得で不変 | PENDING |
| T12 | POST /api/v1/accountings/{11391301}/refunds（最小 body） | 親 path param | 4xx、refund 生成なし（B 再取得で total_refunded 不変） | PENDING |
| T13 | POST /api/v1/reservations に pet_id={11002734}（B pet）を注入 | body 内 ID | 4xx（relation 検証で拒否） | PENDING |
| T14 | POST /api/v1/hospitalizations に B pet_id を注入 | body 内 ID | 4xx | PENDING |
| T15 | GET /api/v1/masters/reservation-types/{B_master_id} | master path param | 404（masters は clinic 固有カタログ） | PENDING |
| T16 | X-Clinic-ID: 2 ヘッダ（未所属 clinic 指定）で GET /api/v1/accountings | header | 403 clinic_selection_unavailable（BUG-128 回帰・別リソース） | PENDING |
| T17 | 【回帰観察・ラウンド1 T5/T11 引き継ぎ】GET /api/v1/medical-records/999999999/treatments vs GET /api/v1/medical-records/{11398719}/treatments（B record）を A として | 存在漏えい | 両者とも同一ステータス/同一形状で判別不能（現在は 200 空配列設計） | PENDING |
| T18 | DELETE /api/v1/hospitalizations/{B_hosp_id}（A として・自作のみ） | path param | 404/403、削除されない（B から再取得可能） | PENDING |

備考:
- 逐次実行・並列なし・総リクエスト約 35 以下。
- 破壊的 write は S8 で自作する B 入院 1 件と、T10/T11/T12 の marker（いずれも隔離成立なら不変・破れた場合のみ変化し finding 証跡となる）に限定。既存 seed データへの無条件破壊 write は行わない。
- 認証 cookie は /tmp/rbl2（permissions 600/700、ラウンド後に削除）。

## ソース確認（実行前リサーチ、防御の想定位置）

- 予約の管理面: `backend/internal/reservation/routes.go:143` `clinics.Group("/reservations")` — clinic_id はパス由来。スコープ検証は reservationCRUD ハンドラ
- 会計: `backend/internal/billing/routes.go:130-141` GetAccounting / refunds
- 見積: `backend/internal/billing/routes.go:94` GetEstimate
- 入院: `backend/internal/medicalrecord/routes_hospitalization.go:11-28`（GET/:id, daily-records, care-plan-items）
- 在庫: `backend/internal/inventory/routes.go:13-16`（単数形 `/inventory`）
- X-Clinic-ID: `backend/internal/middleware/auth.go:351-370`（所属 subset check 後のみ採用）

## 結果サマリ（2026-10-09 実測）

**全18試験が合格シグナルを満たし、TENANT-1 の破れは検出されず。起票なし（0 件）。**

| ID | 実測 | 判定 |
|---|---|---|
| T1 | GET /reservations/{11085643} → 404 `{"error":"not found"}` | PASS |
| T2 | GET /clinics/2/reservations（A）→ 200 だが**返るのは clinic 1（自分側）のトリミング予約のみ**（id 1000000000 = clinic_id 1 を詳細で確認・B は同 ID 404）。B 側は同経路で 200 `[]`（B の予約は 1930 年で月次窓外）。**clinic 2 行は一切流出せず** | PASS（観察①参照） |
| T3 | GET /accountings/{11391301} → 404 | PASS |
| T4 | GET /accountings/{11391301}/refunds → 404 | PASS |
| T5 | GET /estimates/{11000330} → 404 | PASS |
| T6 | GET /hospitalizations/{5}（自作）→ 404 | PASS |
| T7 | GET /hospitalizations/{5}/daily-records → 404 | PASS |
| T8 | GET /inventory/{7} → 404 | PASS |
| T9 | GET /inventory?limit=50（A）→ `clinic_id` は 1 のみ | PASS |
| T10 | PATCH /reservations/{11085643}（notes marker）→ 404、B 再取得で notes `""` 不変 | PASS |
| T11 | PATCH /inventory/{7}（quantity 499）→ 403、B 再取得で 500 不変 | PASS |
| T12 | POST /accountings/{11391301}/refunds → 403、B 再取得で total_refunded 0 不変 | PASS |
| T13 | POST /reservations に B pet_id=11002734 注入 → 404（relation 検証） | PASS |
| T14 | POST /hospitalizations に B pet 注入 → 初回 400（end_date 必須が先行）、end_date 付き再実行で **404（relation 検証で拒否）** | PASS |
| T15 | GET /masters/reservation-types/2（B 側 master）→ 404 | PASS |
| T16 | X-Clinic-ID: 2（未所属）で GET /accountings → 403 `clinic_selection_unavailable`（BUG-128 回帰・会計リソースでも実効確認） | PASS |
| T17 | 【回帰観察】/medical-records/999999999/treatments も /medical-records/11398719/treatments（B record）も A からは **200 `[]` 同一形状** → 判別不能。ラウンド1 T5/T11 の状況不変 | PASS |
| T18 | DELETE /hospitalizations/{5}（A）→ 403、B 再取得で 200（id=5 存続） | PASS |

### 観察（起票しない）

- **観察① clinics/:clinic_id パスとスコープ**: `GET/POST/DELETE /api/v1/clinics/:clinic_id/reservations`（admin 三点）は `appointment_admin_handler.go:47` の `extractSelectedClinicGrant` が**セッションの選択clinic付与**からclinicを導出し、パスの `:clinic_id` はスコープ判定に不参与（任意の値で自分側のデータが返る）。実測では他clinic行の流出なし（fail-safe 方向）だが、パス意味と実装の不一致は API 設計上の分かりにくさとして記録する。将来この経路がパス参env照に変更される際は TENANT-1 再攻撃対象。
- **観察② STG デモパスワードのバンドル消失**: 2026-10-09 実測で、現行 STG フロントエンドバンドル（auth chunk 含む全16 asset）に `VITE_DEMO_LOGIN_PASSWORD` の値が存在しない（docs/ops/deploy/stg/STG-DEMO-DATA-LIFECYCLE.md の「stg bundle に値が含まれることは意図的」と乖離）。`SEEDLOGIN_DEMO_PASSWORD` 未注入の経路で再デプロイされたと推定。fail-closed 方向のため脆弱性ではないが、STG のデモボタン自動入力が壊れている運用観察。本ラウンドをローカルスタック実施した直接の理由。
- **観察③ 検証順序**: POST /hospitalizations は body 相互検証（end_date 必須）が pet/owner の clinic スコープ検証より先行する。単発では relation 検証に到達しないため、注入試験は完全な body で実行する必要がある（T14 再実行で対応）。
- B 側一般スタッフの DELETE は RBAC（hospitalizations:delete 無し）が clinic 分離より先行して 403 — ラウンド1 T6b と同型の権限設計どおり。

### 再報告禁止リスト照合で除外したもの

- 該当なし（本ラウンドの対象はローカル Docker スタック）。

### BLOCKED / SKIP

- なし。計画した全試験を実行（T14 は検証順序の関係で end_date 付き再実行）。

### 後処理

- 自作テスト hospitalization id=5 は、B 一般の DELETE が 403（RBAC 先行）のため、**執行アカウント（林 文明・全clinic所属）+ X-Clinic-ID: 2 で 204 削除**。
- 認証 cookie（a/b/x jar）・取得バンドル類は /tmp/rbl2 ごと削除済み。
- リクエスト総数: 約 38（逐次・並列なし）。

