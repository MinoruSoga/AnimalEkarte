# BUG-378: `/settings/clinic` 一覧に新規作成 clinic が表示されない

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: HIGH (管理機能の不完全性 — system admin でも新規作成した医院が UI 不可視 → 管理不能)
**Affects**: `features/hospital-settings`, `backend/internal/handler/clinic_handler.go`

---

## 概要

`admin@noavet.jp` (`is_system_admin=true`) で `/settings/clinic` 新規登録 → `POST /api/v1/clinics` は **201 Created** + DB 作成成功、直後の `GET /api/v1/clinics` では**新 clinic が返らない**。UI 一覧は 3 件のまま。

## 原因 (コード確認済み)

`backend/internal/handler/clinic_handler.go:16-44` の `ListClinics`：

```go
func (h *Handler) ListClinics(c *gin.Context) {
    scope := c.Query("scope")
    if scope == "all" {
        // master-staff.can_view 権限者: 全 clinic を返す
    }
    // デフォルト: staff_clinic_assignments 経由
    clinics, err := h.svc.Clinic.ListClinicsByStaffID(c.Request.Context(), staffID)
    ...
}
```

Frontend `frontend/src/features/hospital-settings/api/clinics.ts:55` は `scope` クエリなしで GET する：

```ts
const { data } = await axios.get<BackendClinic[]>("/v1/clinics");
```

新規作成された clinic は `staff_clinic_assignments` に登録されないため、作成した staff 自身からも見えない。

## 再現手順

1. admin@noavet.jp でログイン
2. `/settings/clinic` → 「+ 新規登録」 → 院名入力 → 保存
3. `POST /api/v1/clinics` → **201 Created** (id=6)
4. UI が直後に実行する `GET /api/v1/clinics` → **id=3,4,5 のみ**返却
5. 新規作成された id=6 の clinic は一覧に表示されない = UI から編集・削除・状態確認が**一切できない**

## 期待動作

`/settings/clinic` マスタ管理画面では、**全ての clinic** (自分が未割当のものも含む) が一覧表示される。

## 修正方針 (優先順位付き)

### Option A (推奨): Frontend `useGetClinics` が `?scope=all` を送る

```ts
// frontend/src/features/hospital-settings/api/clinics.ts:55
const { data } = await axios.get<BackendClinic[]>("/v1/clinics", { params: { scope: "all" } });
```

- Backend は `master-staff.can_view` 権限をチェックするため (`clinic_handler.go:20`)、権限のない一般スタッフは 403 になる
- settings/clinic は管理者用画面なので妥当
- Backend の resource は `master-staff` だが、実体はクリニック管理。これを機に `master-clinic` or `hospital-settings` に変えるべきか要検討（別タスク）

### Option B: POST 後に作成者の staff_clinic_assignments に自動 INSERT

`backend/internal/service/clinic_service.go` の `CreateClinic` で以下を追加：

```go
// 作成者を新 clinic に自動割当
if err := s.staffAssignmentRepo.Create(ctx, &model.StaffClinicAssignment{
    StaffID: creatorStaffID,
    ClinicID: clinic.ID,
}); err != nil { ... }
```

- メリット: 他の一覧画面 (サイドバー等) でも作成直後に見える
- デメリット: is_system_admin が全 clinic を割当されるわけではないので、根本解決にならない

### Option C (推奨併用): `master-clinic` resource を追加

`hospital-settings` 権限 (医院情報の閲覧/編集) と、`master-clinic` 権限 (clinic 全一覧の CRUD 管理権限) を分離。settings/clinic は後者で制御。

## 関連

- **BUG-377**: 権限モデル不整合 (is_system_admin 二重ガード)。本件と合わせて clinic 管理系の権限設計を再考する価値あり
- 作成されて UI 不可視のテストデータ `clinics.id=6 name="BUG377検証院"` を DB から直接削除する必要あり (または Backend DELETE で)

## 確認事項

- [ ] `scope=all` で is_system_admin 以外のロール (master-staff.can_view あり) の挙動確認
- [ ] staff_clinic_assignments に未登録の clinic で UpdateClinic が動作するか (是否 404)
- [ ] seed の admin@noavet.jp が全 3 clinic に割当済みか確認
