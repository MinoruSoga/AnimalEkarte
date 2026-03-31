# BE-035: マスタリポジトリのページネーション欠落

## ステータス
- 優先度: MEDIUM
- カテゴリ: パフォーマンス
- 影響範囲: repository → service → handler（3層全修正）

## 問題

以下の3リポジトリの `FindAll()` がページネーション未実装で全件取得している。
データ増加時にメモリスパイクとレスポンス遅延の原因になる。

## アーキテクチャ上の懸念（実装前に要判断）

1. **Medicine の階層構造**: `parent_id` による親子構造あり。フラットページネーションで親子分断リスク
2. **Staff の費用対効果**: スタッフ数は数十人程度。全層修正コストに対してリターンが小さい
3. **DiagnosisName の FindByCategoryID**: カテゴリ内フィルタもページネーション対象にすべきか要判断

---

## 修正コード案

### ⚠️ 実装上の注意: buildBase() パターン必須

GORM は `*gorm.DB` に条件を蓄積する。`Count()` 後に同じ `q` で `Offset/Limit` を連結すると、
Count の内部状態が残って不正な SQL が生成される。
`owner_repository.go` と同じ `buildBase()` クロージャパターンで Count と Find を分離すること。

```go
// ❌ NG — q を再利用すると Count の状態が Find に漏れる
q := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID)
q.Model(&model.Medicine{}).Count(&total)
q.Offset(offset).Limit(limit).Find(&medicines)

// ✅ OK — buildBase() で毎回新しいクエリビルダを生成
buildBase := func() *gorm.DB {
    return r.db.WithContext(ctx).Model(&model.Medicine{}).Where("clinic_id = ?", clinicID)
}
buildBase().Count(&total)
buildBase().Offset(offset).Limit(limit).Find(&medicines)
```

---

## 1. medicine_repository.go

### 修正前

```go
// Interface
type MedicineRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Medicine, error)
	// ...
}

// 実装
func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Medicine, error) {
	medicines := make([]model.Medicine, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&medicines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find medicines")
	}
	return medicines, nil
}
```

### 修正後

```go
// Interface
type MedicineRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	// ... 他メソッドは変更なし
}

// 実装
func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	medicines := make([]model.Medicine, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.Medicine{}).Where("clinic_id = ?", clinicID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count medicines")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&medicines).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find medicines")
	}
	return medicines, total, nil
}
```

### 対応するサービス層修正 — medicine_service.go

```go
// 修正前
type MedicineService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Medicine, error)
}
func (s *medicineService) List(ctx context.Context, clinicID uint64) ([]model.Medicine, error) {
	return s.repo.FindAll(ctx, clinicID)
}

// 修正後
type MedicineService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
}
func (s *medicineService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit)
}
```

### 対応するハンドラ層修正 — medicine_handler.go

```go
// 修正前
func (h *Handler) ListMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicines, err := h.svc.Medicine.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineResponseList(medicines))
}

// 修正後
func (h *Handler) ListMedicines(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	medicines, total, err := h.svc.Medicine.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toMedicineResponseList(medicines), total, page, limit))
}
```

---

## 2. staff_repository.go

### 修正前

```go
// Interface
type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error)
	// ...
}

// 実装
func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error) {
	staffs := make([]model.Staff, 0)
	q := r.db.WithContext(ctx).Model(&model.Staff{}).Where("clinic_id = ?", clinicID)
	if role != nil {
		q = q.Where("staff_role = ?", *role)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&staffs).Error; err != nil {
		return nil, apperrors.Wrap(err, "find staffs")
	}
	return staffs, nil
}
```

### 修正後

```go
// Interface
type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error)
	// ... 他メソッドは変更なし
}

// 実装
func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Staff{}).Where("clinic_id = ?", clinicID)
		if role != nil {
			q = q.Where("staff_role = ?", *role)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count staffs")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&staffs).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find staffs")
	}
	return staffs, total, nil
}
```

### 対応するサービス層修正 — staff_service.go

```go
// 修正前
type StaffService interface {
	List(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error)
}
func (s *staffService) List(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error) {
	return s.repo.FindAll(ctx, clinicID, role)
}

// 修正後
type StaffService interface {
	List(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error)
}
func (s *staffService) List(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error) {
	return s.repo.FindAll(ctx, clinicID, role, page, limit)
}
```

### 対応するハンドラ層修正 — staff_handler.go

```go
// 修正前
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}
	staffs, err := h.svc.Staff.List(c.Request.Context(), clinicID, role)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponseList(staffs))
}

// 修正後
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	staffs, total, err := h.svc.Staff.List(c.Request.Context(), clinicID, role, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toStaffResponseList(staffs), total, page, limit))
}
```

---

## 3. diagnosis_repository.go（DiagnosisCategory + DiagnosisName）

### 3a. DiagnosisCategoryRepository — 修正前

```go
type DiagnosisCategoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.DiagnosisCategory, error)
	// ...
}

func (r *diagnosisCategoryRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.DiagnosisCategory, error) {
	categories := make([]model.DiagnosisCategory, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis categories")
	}
	return categories, nil
}
```

### 3a. DiagnosisCategoryRepository — 修正後

```go
type DiagnosisCategoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error)
	// ... 他メソッドは変更なし
}

func (r *diagnosisCategoryRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
	categories := make([]model.DiagnosisCategory, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisCategory{}).Where("clinic_id = ?", clinicID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis categories")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis categories")
	}
	return categories, total, nil
}
```

### 3b. DiagnosisNameRepository — 修正前

```go
type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.DiagnosisName, error)
	FindByCategoryID(ctx context.Context, clinicID, categoryID uint64) ([]model.DiagnosisName, error)
	// ...
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.DiagnosisName, error) {
	names := make([]model.DiagnosisName, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, clinicID, categoryID uint64) ([]model.DiagnosisName, error) {
	names := make([]model.DiagnosisName, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND diagnosis_category_id = ?", clinicID, categoryID).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names by category id")
	}
	return names, nil
}
```

### 3b. DiagnosisNameRepository — 修正後

```go
type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	// ... 他メソッドは変更なし
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	names := make([]model.DiagnosisName, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Where("clinic_id = ?", clinicID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis names")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis names")
	}
	return names, total, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	names := make([]model.DiagnosisName, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).
			Where("clinic_id = ? AND diagnosis_category_id = ?", clinicID, categoryID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis names by category id")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis names by category id")
	}
	return names, total, nil
}
```

### 対応するサービス層修正 — diagnosis_service.go

```go
// ---- DiagnosisCategoryService ----

// 修正前
type DiagnosisCategoryService interface {
	List(ctx context.Context, clinicID uint64) ([]model.DiagnosisCategory, error)
}
func (s *diagnosisCategoryService) List(ctx context.Context, clinicID uint64) ([]model.DiagnosisCategory, error) {
	return s.repo.FindAll(ctx, clinicID)
}

// 修正後
type DiagnosisCategoryService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error)
}
func (s *diagnosisCategoryService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit)
}

// ---- DiagnosisNameService ----

// 修正前
type DiagnosisNameService interface {
	List(ctx context.Context, clinicID uint64) ([]model.DiagnosisName, error)
	ListByCategoryID(ctx context.Context, clinicID, categoryID uint64) ([]model.DiagnosisName, error)
}
func (s *diagnosisNameService) List(ctx context.Context, clinicID uint64) ([]model.DiagnosisName, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *diagnosisNameService) ListByCategoryID(ctx context.Context, clinicID, categoryID uint64) ([]model.DiagnosisName, error) {
	return s.repo.FindByCategoryID(ctx, clinicID, categoryID)
}

// 修正後
type DiagnosisNameService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	ListByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
}
func (s *diagnosisNameService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit)
}
func (s *diagnosisNameService) ListByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return s.repo.FindByCategoryID(ctx, clinicID, categoryID, page, limit)
}
```

### 対応するハンドラ層修正 — diagnosis_handler.go

```go
// ---- ListDiagnosisCategories ----

// 修正前
func (h *Handler) ListDiagnosisCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	categories, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiagnosisCategoryResponseList(categories))
}

// 修正後
func (h *Handler) ListDiagnosisCategories(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	categories, total, err := h.svc.DiagnosisCategory.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toDiagnosisCategoryResponseList(categories), total, page, limit))
}

// ---- ListDiagnosisNames ----

// 修正前
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var names any
	var svcErr error
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, parseErr := strconv.ParseUint(catIDStr, 10, 64)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid category_id"))
			return
		}
		result, err := h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), clinicID, catID)
		names = toDiagnosisNameResponseList(result)
		svcErr = err
	} else {
		result, err := h.svc.DiagnosisName.List(c.Request.Context(), clinicID)
		names = toDiagnosisNameResponseList(result)
		svcErr = err
	}
	if svcErr != nil {
		RespondError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, names)
}

// 修正後
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var resp any
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, parseErr := strconv.ParseUint(catIDStr, 10, 64)
		if parseErr != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid category_id"))
			return
		}
		names, total, svcErr := h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), clinicID, catID, page, limit)
		if svcErr != nil {
			RespondError(c, svcErr)
			return
		}
		resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)
	} else {
		names, total, svcErr := h.svc.DiagnosisName.List(c.Request.Context(), clinicID, page, limit)
		if svcErr != nil {
			RespondError(c, svcErr)
			return
		}
		resp = newPaginatedResponse(toDiagnosisNameResponseList(names), total, page, limit)
	}

	c.JSON(http.StatusOK, resp)
}
```

---

## 実装手順

1. **repository 層修正** — Interface + 実装（3ファイル・5メソッド）
2. **service 層修正** — Interface + 実装（3ファイル・5メソッド）
3. **handler 層修正** — 4ハンドラ関数
4. **テスト修正** — mock interface 更新
   - `service/medicine_service_test.go`
   - `service/staff_service_test.go`
   - `service/diagnosis_service_test.go`
5. **検証**: `docker compose exec backend go test ./... -v`

## API レスポンス形式（変更後）

```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

既存の `parsePagination()` (response.go:123) と `newPaginatedResponse()` (handler.go:33) を利用。
デフォルト: `page=1`, `limit=20`。limit 上限: 100。

## フロントエンド影響

レスポンス形式が `[]Item` → `{ data, total, page, limit }` に変わるため、
`features/master/api/` 内の該当 fetch 関数・hooks とリスト表示コンポーネントの修正が必要。
