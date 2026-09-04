# 同一飼主・ペット連携 仕様書 (Identity Links)

## 概要

- **画面の目的**: 所属医院内の飼主・ペットを手動で identity group にリンク／unlink し、連携先ペットの最小治療履歴を参照する（Phase 1 workbench）。
- **URLパターン**: `/identity-links`（`operations-routes.tsx` に lazy 登録。サイドバー常設ナビは持たない）。
- **アクセス権限**:
  - `identity-links:view` — ページ表示。未所持時は `IdentityLinksPage` が `/` へ `Navigate`（`ResourceIdentityLinks` + `usePermission`）。
  - `identity-links:edit` — リンク／unlink 操作の表示。閲覧のみのときは status バナー「閲覧のみ（link/unlink 権限なし）」。
- **スコープ**: 所属医院内の手動リンクのみ。権限のない医院 ID はサーバ側で拒否（UI 説明文と一致）。

---

## 1. 画面構成

### 1.1 シェル

- `PageLayout` タイトル「同一飼主・ペット連携」。
- `resource={ResourceIdentityLinks}`。
- 内部実装は `IdentityLinksWorkbench`（同一ファイル内）。

### 1.2 飼主リンク

- 検索入力（placeholder「氏名・カナ・電話」）。`useDebouncedValue` 300ms 後に `searchOwnersForLink`。
- ヒット一覧をトグル選択（clinic_id + owner_id）。
- `canEdit` 時: 「飼主をリンク」（選択 2 件以上）、メンバーごとの unlink。メンバー選択時に `findOwnerIdentityGroupByMember` で逆引きし、そのメンバーの group ID が解決した場合だけ unlink を有効化する。

### 1.3 ペットリンク

- 親の飼主 identity group（`ownerGroupId`）が必要である旨を表示。
- 検索入力（placeholder「ペット名・番号」）→ `searchPetsForLink`（300ms debounce）。
- `canEdit` 時: 「ペットをリンク」（ownerGroupId ありかつ選択 2 件以上）、メンバーごとの unlink。メンバー選択時に `findPetIdentityGroupByMember` で逆引きし、そのメンバーの group ID が解決した場合だけ unlink を有効化する。
- 選択ペットごとに「連携履歴」ボタン → `getLinkedTreatmentHistory`（view/edit どちらでも可）。結果は plain text /「（履歴なし）」。

### 1.4 空・エラー・pending

- 検索クエリ空: ヒット一覧をクリア（専用 empty コピーなし）。
- 失敗: `role="alert"` のインラインエラー。
- 閲覧のみ: `role="status"` バナー。
- ミューテーション／履歴取得中は `useTransition` の pending でボタン disable。

---

## 2. 主要フロー・制約

1. 飼主を 2 件以上選択 → 飼主グループ作成 → `ownerGroupId` 保持。既存メンバーを選択した場合はメンバー逆引き GET で group ID を解決する。
2. その group を親にしてペットを 2 件以上選択 → ペットグループ作成。既存ペットの選択時も同様にメンバー逆引き GET を行う。
3. group ID はメンバーごとに保持する。そのメンバーの逆引きが成功した場合だけ unlink を有効化し、別メンバーの session group ID を使って unlink を有効化しない。
4. クロス医院の無断リンクはサーバ拒否。fail-closed の権限デフォルト（新規医院テンプレートに identity-links を付けない）。
5. **親飼主 group の全医院セット必須（mutation）**: ペット group 作成（`CreatePetGroup`）では、actor が親 owner group の **anchor（`CreatedClinicID`）＋全 active owner member の clinic** および全 pet clinic に所属していること。1 医院でも欠けると Forbidden・ゼロ書き込み。any-member フォールバックは禁止（`assertActorCoversOwnerGroupClinics`）。既存 pet group の Add/Unlink は `assertCanManagePetGroup`（pet group anchor + owner-group anchor + 全 pet member clinics）。閲覧（GET/search/history）は actor clinic でフィルタ可。

---

## 3. 臨床安全・アクセシビリティ

- 本画面は臨床記録の直接編集ではなく identity 運用。誤リンク防止のため **edit 必須** で create/unlink を隠す。
- 検索・履歴の失敗を空成功にしない（alert 表示）。
- セクションは `aria-label`（飼主リンク／ペットリンク）。status / alert ロールを使用。

---

## 4. 技術仕様

### 構成

| 要素 | 役割 |
|:---|:---|
| `IdentityLinksPage` | view ゲート + workbench 委譲 |
| `IdentityLinksWorkbench` | 検索・選択・link/unlink・履歴 |
| `usePermission` / `ResourceIdentityLinks` | view / edit |
| `useDebouncedValue` | 検索 300ms |
| `identity-links-api.ts` | HTTP クライアント |

### API連携（本画面が呼ぶもの）

| メソッド | エンドポイント | 用途 | 必須アクション |
|:---|:---|:---|:---|
| GET | `/api/v1/identity-links/owners/search` | 飼主検索 (`q`, `limit`) | `view` |
| GET | `/api/v1/identity-links/pets/search` | ペット検索 | `view` |
| GET | `/api/v1/identity-links/owners/:clinic_id/:owner_id/group` | 選択した飼主メンバーの group 逆引き | `view` |
| GET | `/api/v1/identity-links/pets/:clinic_id/:pet_id/group` | 選択したペットメンバーの group 逆引き | `view` |
| POST | `/api/v1/identity-links/owner-groups` | 飼主グループ作成 | `edit` |
| DELETE | `/api/v1/identity-links/owner-groups/:groupId/members` | 飼主メンバー unlink | `edit` |
| POST | `/api/v1/identity-links/pet-groups` | ペットグループ作成（`owner_group_id` 必須） | `edit` |
| DELETE | `/api/v1/identity-links/pet-groups/:groupId/members` | ペットメンバー unlink | `edit` |
| GET | `/api/v1/identity-links/pets/:clinicId/:petId/treatment-history` | 連携治療履歴（`include_linked`, page, limit） | `view` |

### BE 追加ルート（Phase 1 UI 未使用・OpenAPI 契約済み）

| メソッド | エンドポイント | 用途 | 必須アクション |
|:---|:---|:---|:---|
| GET | `/api/v1/identity-links/owner-groups/:id` | group 取得（可視メンバーのみ） | `view` |
| POST | `/api/v1/identity-links/owner-groups/:id/members` | 飼主メンバー追加 | `edit` |
| GET | `/api/v1/identity-links/pet-groups/:id` | pet group 取得 | `view` |
| POST | `/api/v1/identity-links/pet-groups/:id/members` | ペットメンバー追加 | `edit` |

OpenAPI 正本: `backend/docs/api.yaml`（`/identity-links/*`）。ルート drift gate は `backend/internal/apicontract/openapi_route_drift_test.go` が identitylink package を walk。

### 関連

- ルート: `frontend/src/app/routes/operations-routes.tsx`
- Feature: `frontend/src/features/identity-links/`
- 製品 leaf 数の正本は `route-inventory.test.tsx`。本ファイルは **画面仕様インデックス上の 1 葉**であり、product leaf 数と「番号付き md ファイル数」は一致しない。

---

## 5. Phase 1 Acceptance Criteria（証拠）

| AC | 要件 | 証拠 |
|:---|:---|:---|
| AC-1 | 手動 link/unlink（飼主・ペット） | BE: `CreateOwnerGroup` / `UnlinkOwnerMember` / `CreatePetGroup` / `UnlinkPetMember` + FE workbench ボタン。OpenAPI: POST/DELETE owner-groups・pet-groups members |
| AC-2 | clinic-scoped linked treatment history | `ListLinkedTreatmentHistory` + `include_linked` 相関 `(clinic_id,pet_id)` のみ。FE: `getLinkedTreatmentHistory`。Test: `TestListLinkedTreatmentHistory_*` |
| AC-3 | mixed / hidden / cross-clinic IDs は **全体 reject・部分書き込みなし** | `assertOwnerRefsInActorScope` / `assertPetRefsInActorScope` + lock length check。Tests: `TestCreateOwnerGroup_RejectsMixedCrossClinic_NoPartialWrite`, `TestCreatePetGroup_RejectsMixedCrossClinic_NoPartialWrite`, `TestAddOwnerMembers_RejectsMixedCrossClinic_NoPartialWrite`, `TestAddPetMembers_RejectsMixedCrossClinic_NoPartialWrite`, `TestCreateOwnerGroup_RejectsHiddenOwner_NoPartialWrite` |
| AC-3b | 親 owner group の **全医院**（anchor + active members）を actor がカバーしない CreatePetGroup は Forbidden・ゼロ書き込み | `assertActorCoversOwnerGroupClinics`。Tests: `TestCreatePetGroup_RejectsMissingParentOwnerAnchorClinic_NoPartialWrite`, `TestCreatePetGroup_RejectsMissingParentOwnerMemberClinic_NoPartialWrite`, `TestCreatePetGroup_AllowsWhenActorCoversAllParentOwnerAndPetClinics` |
| AC-4 | audit は business write と同一 tx・fail-closed | `writeAudit` 失敗で callback error。PHI（name/phone）を audit payload に含めない（IDs のみ）。Tests: `TestCreateOwnerGroup_AuditFailureRollsBack`, `TestCreateOwnerGroup_NilAuditFailClosed`, `TestCreateOwnerGroup_SuccessWritesAuditWithoutPHI` |
| AC-5 | 権限 fail-closed（view/edit） | routes: GET=view, link/unlink=edit。FE: view なし → `/` Navigate、edit なし → 閲覧バナー。Tests: `handler_permission_test.go`, `IdentityLinksPage.test.tsx` |
| AC-6 | OpenAPI と RegisterRoutes の route 一致 | identity-links 全 13 ルートを `api.yaml` に記載。reverse-lookup 2 本を known-missing allowlist から除去済み |

**Out of scope (Phase 2 / DEC-46)**: 自動 link、merge、候補提示 UI。

---
