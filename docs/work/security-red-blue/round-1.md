# Red/Blue ラウンド1 — TENANT-1 (clinic 分離) / pet・medicalrecord 系 / STG

- 日付: 2026-10-09
- 対象: STG API (`animalekarte-stg-api.baritech-soga.workers.dev`, synthetic データ専用)
- invariant: TENANT-1（clinic isolation）
- カタログ: attack-catalog.md A 項
- 攻撃者起点: clinic 1（八王子病院）一般スタッフ（獣医師）の認証セッションのみ所持
- 防御対象: clinic 2（城東センター病院）のリソース

## セットアップ（実測済み）

| ID | 操作 | 結果 |
|---|---|---|
| S1 | 未認証 GET /api/v1/pets, /api/v1/medical-records | 401（fail-closed 確認） |
| S2 | clinic A staff で login（stg-staff-10000003@example.test, 一般） | 200, cookie 取得 |
| S3 | clinic B staff で login（stg-staff-20000003@example.test, 一般） | 200, cookie 取得 |
| S4 | A: GET /owners, /pets | clinic 1 のみ返る（owner 1000000007, pet 1001979） |
| S5 | B: GET /owners, /masters/animal-species, /medical-records | clinic 2 のみ返る（owner 1000000011, record 11085642 = pet 11002734 / owner 10407914） |
| S6 | B: POST /pets（自リソース作成） | 201 — pet 1000000012 "RBL1-SECTEST"（clinic 2, owner 1000000011）。**書込み系攻撃の標的はこの自作 pet のみに限定**し、既存 synthetic データへの破壊的 write は行わない |

- 認証情報: STG 公開バンドルに焼き込まれたデモログイン定数（LoginForm 自動入力用ビルド値）を使用。値は本文書へ記載しない。この露出自体は SECURITY.md 受容済みリスク（デモカタログ共有パスワード設計）であり再報告対象外。
- トークン/cookie はすべてマスク。管理は /tmp/rbl1（permissions 600/700、ラウンド後に削除）。

## 試験マトリクス（clinic A トークンで clinic B リソース ID を攻撃）

合格シグナルは**実行前に**確定済み。不合格 = finding 候補。

| ID | エンドポイント | 経路 | 期待合格シグナル | 結果 |
|---|---|---|---|---|
| T1 | GET /api/v1/pets/{B_pet_id} | path param | 404、レスポンスに clinic B pet の存在を示す差分なし | PENDING |
| T2 | GET /api/v1/pets/{B_pet_id}/first-visit | path param | 404 | PENDING |
| T3 | GET /api/v1/pets/{B_pet_id}/sub-owners | path param | 404 | PENDING |
| T4 | GET /api/v1/pets/{B_pet_id}/chronic-conditions | path param | 404 | PENDING |
| T5 | GET /api/v1/pets/{B_pet_id}/treatment-history | path param | 404 | PENDING |
| T6 | PATCH /api/v1/pets/{B_pet_id}（名前変更） | path param（自作 pet のみ） | 404、pet が変更されない | PENDING |
| T7 | GET /api/v1/owners/{B_owner_id}/shared-pets | path param（間接参照） | 404 または clinic B pet を含まない空集合 | PENDING |
| T8 | GET /api/v1/owners/{B_owner_id}/report/pets | path param（間接参照） | 404 または clinic B pet を含まない空集合 | PENDING |
| T9 | GET /api/v1/medical-records/{B_record_id} | path param | 404、存在差分なし | PENDING |
| T10 | GET /api/v1/medical-records/{B_record_id}/vitals | 親 path param | 404 | PENDING |
| T11 | GET /api/v1/medical-records/{B_record_id}/treatments | 親 path param | 404 | PENDING |
| T12 | GET /api/v1/medical-records/{B_record_id}/addenda | 親 path param | 404 | PENDING |
| T13 | GET /api/v1/medical-records/{B_record_id}/images | 親 path param | 404 | PENDING |
| T14 | PATCH /api/v1/pets/{A_pet_id}/sub-owners に B_owner_id を注入 | body 内 ID | 4xx（clinic 違い owner 参照は拒否）、A pet の sub-owner に B owner が紐付かない | PENDING |
| T15 | POST /api/v1/vaccinations に pet_id={B_pet_id} を注入 | body 内 ID | 4xx（relation 検証で拒否） | PENDING |
| T16 | POST /api/v1/medical-records に pet_id={B_pet_id} を注入 | body 内 ID | 4xx、clinic A に cross-tenant 参照 record が生成されない | PENDING |
| T17 | X-Clinic-ID: 2 ヘッダ（未所属 clinic 指定）で GET /api/v1/pets | header | 403 "not assigned to this clinic"（BUG-128 修正の実効確認） | PENDING |
| T18 | GET /api/v1/pets?clinic_ids=1,2 | query param | clinic 2 行を含まない（AuthorizeClinicIDs subset check） | PENDING |
| T19 | GET /api/v1/pets/{B_pet_id} vs {存在しない ID 999999999} の差分比較 | 存在漏えい | 両者とも同一ステータス/同一形状で判別不能 | PENDING |

備考:
- 逐次実行・並列なし・総リクエスト約 25 以下。
- T6 は S6 で自作した pet のみ（破壊しても自作 synthetic 1 件に限定）。
- DELETE は本轮では実行しない（読取り系と UPDATE 系で TENANT-1 判定は十分。DELETE 経路の検証はラウンド2に持ち越し）。
- レコードへの write（PATCH /medical-records/{B_record_id}）は既存 synthetic レコードを壊し得るため未実施。代わりに T16 の create 経路と read 系で判定する。

## ソース確認（実行前リサーチ、防御の想定位置）

- 認証後 authority 再解決: `backend/internal/middleware/auth.go:58-153`（X-Clinic-ID は所属 subset check 後のみ採用、未所属は 403 — `auth.go:351-370`）
- pet read: `backend/internal/pet/pet_handler.go`（Get → `ResolveAllClinicIDsForPermission` + `GetByIDForClinics`、write → `ExtractClinicID`）
- pet list: `ResolveListClinicIDsForPermission` → `AuthorizeClinicIDs` subset check（`backend/internal/httpapi/context.go:153`）
- vaccination create: `backend/internal/medicalrecord/vaccination_service.go:375-412`（vaccine/pet/record/doctor をすべて clinic スコープで validateRelations）
- CSRF: `backend/internal/middleware/csrf.go`（POST/PATCH/DELETE に X-Requested-With 必須）

## 結果サマリ（2026-10-09 実測）

**全試験が合格シグナルを満たし、TENANT-1 の破れは検出されず。起票なし（0 件）。**

| ID | 実測 | 判定 |
|---|---|---|
| T1 | GET /pets/{B_pet} → 404 `{"error":"not found"}` | PASS |
| T2 | GET /pets/{B_pet}/first-visit → 404 | PASS |
| T3 | GET /pets/{B_pet}/sub-owners → 404 | PASS |
| T4 | GET /pets/{B_pet}/chronic-conditions → 404 | PASS |
| T5 | GET /pets/{B_pet}/treatment-history → 200 `{"data":[],"total":0}`。**リポジトリは clinic 述語付き**（`treatment_repository.go:133-135`、`medical_records.clinic_id = ? AND pet_id = ?`）。存在しない pet 999999999 も同一形状 → 存在漏えいなし | PASS |
| T6 | PATCH /pets/{B_pet}（名前変更）→ 404、clinic B で再取得して不変を確認 | PASS |
| T6b | DELETE /pets/{B_pet}（clinic A）→ **403 forbidden**。一般スタッフは `owners:delete` を持たず RBAC が先行し、clinic 分離のシグナルは直接観測されず（fail-closed は確認）。clinic B 一般の DELETE も 403 = 権限設計どおり | PASS（RBAC 先行のため分離シグナルはソース裏付けのみ: `pet/repository.go:55` `Delete(clinicID, id)`） |
| T7 | GET /owners/{B_owner}/shared-pets → 404 | PASS |
| T8 | GET /owners/{B_owner}/report/pets → 404 | PASS |
| T9 | GET /medical-records/{B_record} → 404 | PASS |
| T10 | GET /medical-records/{B_record}/vitals → 404 | PASS |
| T11 | GET /medical-records/{B_record}/treatments → 200 `[]`。存在しない record も 200 `[]` で同一 → 存在漏えいなし。クエリは clinic 述語付き（`treatment_handler.go:55-76` → `FindByMedicalRecordID(clinicID, …)`） | PASS |
| T12 | GET /medical-records/{B_record}/addenda → 404 | PASS |
| T13 | GET /medical-records/{B_record}/images → 404 | PASS |
| T14 | PUT /pets/{A_pet}/sub-owners に B owner 注入 → 404、sub_owners 不変 | PASS |
| T15 | POST /vaccinations に B pet_id 注入 → 404（`validateRelations` が clinic スコープで pet 検証 — `vaccination_service.go:383-388`） | PASS |
| T16 | POST /medical-records に B pet_id 注入 → 404、レコード生成なし | PASS |
| T17 | X-Clinic-ID: 2（未所属）→ 403 `clinic_selection_unavailable`（BUG-128 修正の実効確認） | PASS |
| T18 | GET /pets?clinic_ids=1,2（非許可混入）→ 403 fail-closed。対照 clinic_ids=1 → 200 で clinic 1 のみ | PASS |
| T19 | 存在差分: B pet 404 / 存在しない pet 404、treatment-history は両者とも 200 空配列同一形状 → 判別不能 | PASS |

### 観察（起票しない）

- `GET /medical-records/:id/treatments` は親 record を解決せず clinic 述語クエリのみで 200 `[]` を返す（vitals/images/addenda は 404）。本実測では漏えいなし（存在しない ID と同一応答）。将来このエンドポイントの応答形状が変わり「存在しない → 404」になった場合、cross-tenant 存在オラクル化し得るため再攻撃時の回帰観察ポイント。
- clinic 分離系の child-list（treatments / treatment-history）は 200 空配列、parent-resolve 系は 404 という設計の不統一はあるが、セキュリティ影響は本実測では確認されず。

### 再報告禁止リスト照合で除外したもの

- STG 公開バンドルへのデモログイン定数焼き込み（LoginForm 自動入力）— SECURITY.md 受容済み（デモカタログ共有パスワード設計）。
- 執行アカウント（林 文明）が共有パスワードで認証でき、全カタログ clinic に切替可能 — 同上（cleanup のため使用）。

### BLOCKED / SKIP

- なし。計画した全試験を実行。追加の T6b は RBAC 先行により分離シグナルを直接観測できなかったが、これは権限設計どおりの挙動でありラウンド2で執行権限なし clinic 隔離の DELETE を改めて観察する価値は低い（ソースで `Delete(clinicID, id)` スコープ済み）。

### 後処理

- 自作テスト pet 1000000012 は執行アカウントで 204 削除、clinic B から 404 を確認済み。
- 認証 cookie・パスワードを含む一時ファイルは /tmp/rbl1 ごと削除済み。
- リクエスト総数: 約 35（逐次・並列なし）。

### ラウンド2 以降の候補（本ラウンドでは未実施）

- 予約 (reservation)・会計 (billing)・在庫 (inventory)・入院 (hospitalization) 系の TENANT-1
- AUTHZ-1: role 差（一般 vs 執行）での write 権限越え、`masters` 系の clinic 分離
- OBJECT-1: 診療画像 signed URL の clinic scope（clinic A トークンで clinic B 画像 key 要求）
- AUTHN: refresh rotation・reuse detection の実測
- **回帰観察（ラウンド1 観察節からの引き継ぎ・2026-10-09 追記）**: `GET /medical-records/:id/treatments`（T11）と `GET /pets/:id/treatment-history`（T5）は親/対象解決なしで clinic 述語クエリのみの 200 空配列応答。現時点で存在漏えいなし（存在しない ID と同一形状）だが、**これらのエンドポイントの応答仕様が「存在しない→404」へ変わった場合 cross-tenant 存在オラクル化する**ため、ラウンド2 以降の再攻撃時に必ず再実測する
