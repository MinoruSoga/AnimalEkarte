package medicalrecord

// Treatment package master service (EMR-518 §3/§4/§6/§7, EMR-521).
// All catalog writes join one lock order: clinic row -> catalog row -> package
// IDs ascending -> component. Reads never create catalog rows and never invent
// version 1. Conflicts/rejections are zero-write. trusted clinic/actor come only
// from server auth context; body/query clinic/actor are never accepted.

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Per-typed-source current view grant (§7 / api contract):
// checkup_type -> ResourceCheckups, merchandise_item -> ResourceMasterMerchandise,
// the other five medical masters -> ResourceMasterMedical. manual_other needs none.
var treatmentPackageKindViewResource = map[string]model.Resource{
	model.TreatmentPackageItemKindConsultation:    model.ResourceMasterMedical,
	model.TreatmentPackageItemKindProcedure:       model.ResourceMasterMedical,
	model.TreatmentPackageItemKindMedicine:        model.ResourceMasterMedical,
	model.TreatmentPackageItemKindExaminationType: model.ResourceMasterMedical,
	model.TreatmentPackageItemKindVaccine:         model.ResourceMasterMedical,
	model.TreatmentPackageItemKindCheckupType:     model.ResourceCheckups,
	model.TreatmentPackageItemKindMerchandiseItem: model.ResourceMasterMerchandise,
}

// packageViewAuthorizer is the consumer-side view of the per-clinic RBAC check
// (composition root adapts auth.HasPermissionInClinic). One miss is fail-closed.
type treatmentPackageViewAuthorizer interface {
	Granted(ctx context.Context, clinicID uint64, resource model.Resource, action string) bool
}

// Minimal per-kind master read views (consumer-side, ADR-006 準拠).
type packageConsultationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
}
type packageProcedureFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
}
type packageMedicineFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
}
type packageExamTypeFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
}
type packageVaccineFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
}
type packageCheckupTypeFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
}
type packageMerchandiseItemFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
}

// TreatmentPackageDependencies wires typed-source readers, the per-clinic grant
// check, the shared transaction boundary, and tx-internal audit. All are required
// for writes; missing dependencies fail closed before any mutation.
type TreatmentPackageDependencies struct {
	Transactor       Transactor
	Audit            AuditTxLogger
	Authorizer       treatmentPackageViewAuthorizer
	Consultations    packageConsultationFinder
	Procedures       packageProcedureFinder
	Medicines        packageMedicineFinder
	ExamTypes        packageExamTypeFinder
	Vaccines         packageVaccineFinder
	CheckupTypes     packageCheckupTypeFinder
	MerchandiseItems packageMerchandiseItemFinder
}

type CreateTreatmentPackageInput struct {
	ActorID     *uint64
	Name        string
	Description string
	IsActive    bool
	Components  []TreatmentPackageComponentInput
}

type UpdateTreatmentPackageInput struct {
	Version     int64 // editing-start version CAS; never replaced by a fresh read.
	ActorID     *uint64
	Name        *string
	Description *string
	IsActive    *bool
	Components  *[]TreatmentPackageComponentInput
}

type TreatmentPackageService interface {
	List(ctx context.Context, clinicID uint64) ([]TreatmentPackageView, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*TreatmentPackageView, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTreatmentPackageInput) (*TreatmentPackageView, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTreatmentPackageInput) (*TreatmentPackageView, error)
	Delete(ctx context.Context, clinicID, id uint64, version int64, actorID uint64) error
	Reorder(ctx context.Context, clinicID uint64, orderVersion int64, ids []uint64, actorID uint64) error
}

type treatmentPackageService struct {
	repo TreatmentPackageRepository
	deps TreatmentPackageDependencies
}

func NewTreatmentPackageService(repo TreatmentPackageRepository, deps TreatmentPackageDependencies) TreatmentPackageService {
	return &treatmentPackageService{repo: repo, deps: deps}
}

func (s *treatmentPackageService) requireDeps() error {
	if s.deps.Transactor == nil || s.deps.Audit == nil || s.deps.Authorizer == nil ||
		s.deps.Consultations == nil || s.deps.Procedures == nil || s.deps.Medicines == nil ||
		s.deps.ExamTypes == nil || s.deps.Vaccines == nil || s.deps.CheckupTypes == nil ||
		s.deps.MerchandiseItems == nil {
		return apperrors.WrapInternalServerError("treatment package dependencies are required")
	}
	return nil
}

func (s *treatmentPackageService) List(ctx context.Context, clinicID uint64) ([]TreatmentPackageView, int64, error) {
	if s.deps.Authorizer == nil {
		return nil, 0, apperrors.WrapInternalServerError("treatment package authorizer is required")
	}
	revision, err := s.repo.GetCatalogRevision(ctx, clinicID)
	if err != nil {
		return nil, 0, err
	}
	pkgs, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list treatment packages")
	}
	ids := make([]uint64, 0, len(pkgs))
	for i := range pkgs {
		ids = append(ids, pkgs[i].ID)
	}
	items, err := s.repo.FindLiveItemsByPackageIDs(ctx, clinicID, ids)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list treatment package items")
	}
	byPackage := make(map[uint64][]model.TreatmentPackageItem, len(pkgs))
	for i := range items {
		byPackage[items[i].PackageID] = append(byPackage[items[i].PackageID], items[i])
	}
	views := make([]TreatmentPackageView, 0, len(pkgs))
	for i := range pkgs {
		pkgItems := byPackage[pkgs[i].ID]
		if err := s.authorizeItemKinds(ctx, clinicID, pkgItems); err != nil {
			// 権限不足の構成を含む read は 403 で全体非開示（部分返却しない）。
			return nil, 0, err
		}
		views = append(views, TreatmentPackageView{Package: pkgs[i], Items: pkgItems})
	}
	return views, revision, nil
}

func (s *treatmentPackageService) GetByID(ctx context.Context, clinicID, id uint64) (*TreatmentPackageView, error) {
	if s.deps.Authorizer == nil {
		return nil, apperrors.WrapInternalServerError("treatment package authorizer is required")
	}
	pkg, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.FindLiveItemsByPackageIDs(ctx, clinicID, []uint64{id})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list treatment package items")
	}
	if err := s.authorizeItemKinds(ctx, clinicID, items); err != nil {
		return nil, err
	}
	return &TreatmentPackageView{Package: *pkg, Items: items}, nil
}

// authorizeItemKinds requires the current per-type view grant for every kind
// present. manual_other needs none. 失権 component を含む read/write は拒否。
func (s *treatmentPackageService) authorizeItemKinds(ctx context.Context, clinicID uint64, items []model.TreatmentPackageItem) error {
	seen := map[string]struct{}{}
	for i := range items {
		kind := items[i].ItemKind
		if _, ok := seen[kind]; ok {
			continue
		}
		seen[kind] = struct{}{}
		res, ok := treatmentPackageKindViewResource[kind]
		if !ok {
			return apperrors.WrapInternalServerError("unknown package item kind")
		}
		if !s.deps.Authorizer.Granted(ctx, clinicID, res, "view") {
			return apperrors.WrapForbidden("forbidden")
		}
	}
	return nil
}

// validateComponents checks each typed source inside the write transaction:
// same clinic (repo FindByID scoping), active, and current view grant. manual
// input is already schema-validated at the boundary; master values are never
// copied and no manual default is invented.
func (s *treatmentPackageService) validateComponents(ctx context.Context, clinicID uint64, components []TreatmentPackageComponentInput) error {
	authorized := map[model.Resource]bool{}
	for _, c := range components {
		res, ok := treatmentPackageKindViewResource[c.Kind]
		if !ok {
			return apperrors.WrapInvalidInput("item.kind の値が不正です")
		}
		if !authorized[res] {
			if !s.deps.Authorizer.Granted(ctx, clinicID, res, "view") {
				return apperrors.WrapForbidden("forbidden")
			}
			authorized[res] = true
		}
		if c.Kind == model.TreatmentPackageItemKindManualOther {
			continue
		}
		if err := s.validateTypedSource(ctx, clinicID, c); err != nil {
			return err
		}
	}
	return nil
}

func (s *treatmentPackageService) validateTypedSource(ctx context.Context, clinicID uint64, c TreatmentPackageComponentInput) error {
	active := false
	var err error
	switch c.Kind {
	case model.TreatmentPackageItemKindConsultation:
		var m *model.Consultation
		m, err = s.deps.Consultations.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindProcedure:
		var m *model.Procedure
		m, err = s.deps.Procedures.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindMedicine:
		var m *model.Medicine
		m, err = s.deps.Medicines.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindExaminationType:
		var m *model.ExaminationType
		m, err = s.deps.ExamTypes.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindVaccine:
		var m *model.Vaccine
		m, err = s.deps.Vaccines.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindCheckupType:
		var m *model.CheckupType
		m, err = s.deps.CheckupTypes.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	case model.TreatmentPackageItemKindMerchandiseItem:
		var m *model.MerchandiseItem
		m, err = s.deps.MerchandiseItems.FindByID(ctx, clinicID, c.TypedID)
		if m != nil {
			active = m.IsActive
		}
	}
	if err != nil {
		return err // missing/cross-clinic rows stay NotFound via scoped FindByID.
	}
	if !active {
		return apperrors.WrapUnprocessable("無効なマスタはパッケージ構成にできません")
	}
	return nil
}

func packageItemFromInput(clinicID uint64, c TreatmentPackageComponentInput, position int) model.TreatmentPackageItem {
	pos := position
	item := model.TreatmentPackageItem{
		ClinicID:     clinicID,
		ComponentKey: c.ComponentKey,
		ItemKind:     c.Kind,
		Quantity:     c.Quantity,
		Position:     &pos,
	}
	switch c.Kind {
	case model.TreatmentPackageItemKindConsultation:
		item.ConsultationID = &c.TypedID
	case model.TreatmentPackageItemKindProcedure:
		item.ProcedureID = &c.TypedID
	case model.TreatmentPackageItemKindMedicine:
		item.MedicineID = &c.TypedID
	case model.TreatmentPackageItemKindExaminationType:
		item.ExamTypeID = &c.TypedID
	case model.TreatmentPackageItemKindVaccine:
		item.VaccineID = &c.TypedID
	case model.TreatmentPackageItemKindCheckupType:
		item.CheckupTypeID = &c.TypedID
	case model.TreatmentPackageItemKindMerchandiseItem:
		item.MerchandiseItemID = &c.TypedID
	case model.TreatmentPackageItemKindManualOther:
		if c.Manual != nil {
			raw := json.RawMessage(append([]byte(nil), c.Manual.Raw...))
			item.ManualDefinition = &raw
		}
	}
	return item
}

func (s *treatmentPackageService) Create(ctx context.Context, clinicID uint64, input *CreateTreatmentPackageInput) (*TreatmentPackageView, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if input.ActorID == nil || *input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required")
	}
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	if len(input.Components) == 0 {
		return nil, apperrors.WrapInvalidInput("components は1件以上必要です")
	}
	if err := s.requireDeps(); err != nil {
		return nil, err
	}
	var view *TreatmentPackageView
	err := s.deps.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.LockClinicRow(txCtx, clinicID); err != nil {
			return err
		}
		revision, err := s.repo.LockCatalogRow(txCtx, clinicID)
		if err != nil {
			return err
		}
		if err := s.validateComponents(txCtx, clinicID, input.Components); err != nil {
			return err
		}
		maxOrder, err := s.repo.MaxSortOrder(txCtx, clinicID)
		if err != nil {
			return err
		}
		now := time.Now()
		pkg := &model.TreatmentPackage{
			ClinicID:    clinicID,
			Name:        input.Name,
			Description: input.Description,
			IsActive:    input.IsActive,
			SortOrder:   maxOrder + 1, // create は末尾追加。
			Version:     1,
			CreatedBy:   *input.ActorID,
			UpdatedBy:   *input.ActorID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		items := make([]model.TreatmentPackageItem, 0, len(input.Components))
		for i, c := range input.Components {
			items = append(items, packageItemFromInput(clinicID, c, i))
		}
		if err := s.repo.Create(txCtx, pkg, items); err != nil {
			return err
		}
		if err := s.repo.IncrementCatalogRevision(txCtx, clinicID, revision); err != nil {
			return err
		}
		pkgID := pkg.ID
		if err := s.deps.Audit.LogEntryTx(txCtx, &AuditEntry{
			ClinicID: &clinicID, ActorID: input.ActorID, ActorType: model.AuditActorTypeStaff,
			Action: model.AuditActionTreatmentPackageCreate, Resource: model.AuditResourceTreatmentPackage, ResourceID: &pkgID,
			NewValue: map[string]any{"name": pkg.Name, "is_active": pkg.IsActive, "components": len(items)},
		}); err != nil {
			return err
		}
		view = &TreatmentPackageView{Package: *pkg, Items: items}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "treatment package created", slog.Uint64("clinic_id", clinicID), slog.Uint64("package_id", view.Package.ID))
	return view, nil
}

func (s *treatmentPackageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTreatmentPackageInput) (*TreatmentPackageView, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if input.Version < 1 {
		return nil, apperrors.WrapInvalidInput("positive editing-start version is required")
	}
	if input.ActorID == nil || *input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required")
	}
	if input.Name == nil && input.Description == nil && input.IsActive == nil && input.Components == nil {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	if err := s.requireDeps(); err != nil {
		return nil, err
	}
	var view *TreatmentPackageView
	err := s.deps.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		now := time.Now()
		if err := s.repo.LockClinicRow(txCtx, clinicID); err != nil {
			return err
		}
		locked, err := s.repo.LockPackagesByIDAsc(txCtx, clinicID, []uint64{id})
		if err != nil {
			return err
		}
		if len(locked) == 0 {
			return apperrors.WrapNotFound("treatment_package", "not found")
		}
		current := locked[0]
		if current.Version != input.Version {
			// zero-write conflict。
			return apperrors.WrapConflict("編集後に変更されています。再読み込みしてから保存してください")
		}
		// F1: components 未指定の更新でも、lock 済み package の live items を書込み前に検査する。
		// 権限不足の型を含む場合は 403 で全体非開示（zero-write rollback）。
		liveItems, err := s.repo.FindLiveItemsByPackageIDs(txCtx, clinicID, []uint64{id})
		if err != nil {
			return err
		}
		if err := s.authorizeItemKinds(txCtx, clinicID, liveItems); err != nil {
			return err
		}
		fields := map[string]any{}
		changed := false
		if input.Name != nil && *input.Name != current.Name {
			if err := validateRequiredName(*input.Name); err != nil {
				return err
			}
			fields["name"] = *input.Name
			changed = true
		}
		if input.Description != nil && *input.Description != current.Description {
			fields["description"] = *input.Description
			changed = true
		}
		if input.IsActive != nil && *input.IsActive != current.IsActive {
			fields["is_active"] = *input.IsActive
			changed = true
		}
		if input.Components != nil {
			if err := s.validateComponents(txCtx, clinicID, *input.Components); err != nil {
				return err
			}
			diff, applyErr := s.diffComponents(txCtx, clinicID, id, *input.Components, now)
			if applyErr != nil {
				return applyErr
			}
			if diff {
				changed = true
			}
		}
		if !changed {
			// no actual change: no fake revision（zero-write no-op）。
			// 応答 items は lock 後に検査済みの liveItems をそのまま使う。
			view = &TreatmentPackageView{Package: current, Items: liveItems}
			return nil
		}
		updated, err := s.repo.UpdateFieldsCAS(txCtx, clinicID, id, input.Version, fields)
		if err != nil {
			return err
		}
		if err := s.deps.Audit.LogEntryTx(txCtx, &AuditEntry{
			ClinicID: &clinicID, ActorID: input.ActorID, ActorType: model.AuditActorTypeStaff,
			Action: model.AuditActionTreatmentPackageUpdate, Resource: model.AuditResourceTreatmentPackage, ResourceID: &id,
			OldValue: map[string]any{"name": current.Name, "is_active": current.IsActive, "version": current.Version},
			NewValue: map[string]any{"name": updated.Name, "is_active": updated.IsActive, "version": updated.Version},
		}); err != nil {
			return err
		}
		items, err := s.repo.FindLiveItemsByPackageIDs(txCtx, clinicID, []uint64{id})
		if err != nil {
			return err
		}
		// 応答 items にも同じ検査を適用し、権限不足の型を含むなら 403 で全体非開示。
		if err := s.authorizeItemKinds(txCtx, clinicID, items); err != nil {
			return err
		}
		view = &TreatmentPackageView{Package: *updated, Items: items}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "treatment package updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("package_id", id))
	return view, nil
}

// diffComponents applies the ordered full replacement under the locked parent:
// existing key update / new key insert / missing key -> removed_at+position=NULL
// (kept row; keys never reused). Returns whether anything changed.
func (s *treatmentPackageService) diffComponents(ctx context.Context, clinicID, packageID uint64, desired []TreatmentPackageComponentInput, now time.Time) (bool, error) {
	existing, err := s.repo.FindAllItemsForPackage(ctx, clinicID, packageID)
	if err != nil {
		return false, err
	}
	liveByKey := make(map[uuid.UUID]*model.TreatmentPackageItem, len(existing))
	removedKeys := make(map[uuid.UUID]struct{}, len(existing))
	for i := range existing {
		if existing[i].RemovedAt == nil {
			liveByKey[existing[i].ComponentKey] = &existing[i]
		} else {
			removedKeys[existing[i].ComponentKey] = struct{}{}
		}
	}
	seen := make(map[uuid.UUID]struct{}, len(desired))
	keep := make([]model.TreatmentPackageItem, 0, len(desired))
	for i, c := range desired {
		if _, ok := removedKeys[c.ComponentKey]; ok {
			return false, apperrors.WrapInvalidInput("removed component_key は再利用できません")
		}
		seen[c.ComponentKey] = struct{}{}
		item := packageItemFromInput(clinicID, c, i)
		if prev, ok := liveByKey[c.ComponentKey]; ok {
			item.ID = prev.ID
		}
		keep = append(keep, item)
	}
	var removeIDs []uint64
	for i := range existing {
		if existing[i].RemovedAt != nil {
			continue
		}
		if _, ok := seen[existing[i].ComponentKey]; !ok {
			removeIDs = append(removeIDs, existing[i].ID)
		}
	}
	if len(removeIDs) == 0 && componentsUnchanged(existing, keep) {
		return false, nil
	}
	if err := s.repo.ReplaceItems(ctx, clinicID, packageID, keep, removeIDs, now); err != nil {
		return false, err
	}
	return true, nil
}

// componentsUnchanged reports whether keep preserves every live component's
// key, kind, FK, quantity text, manual payload, and position order.
func componentsUnchanged(existing []model.TreatmentPackageItem, keep []model.TreatmentPackageItem) bool {
	var live []model.TreatmentPackageItem
	for i := range existing {
		if existing[i].RemovedAt == nil {
			live = append(live, existing[i])
		}
	}
	if len(live) != len(keep) {
		return false
	}
	byKey := make(map[uuid.UUID]*model.TreatmentPackageItem, len(live))
	for i := range live {
		byKey[live[i].ComponentKey] = &live[i]
	}
	for i := range keep {
		prev, ok := byKey[keep[i].ComponentKey]
		if !ok {
			return false
		}
		if prev.ItemKind != keep[i].ItemKind || prev.ID != keep[i].ID {
			return false
		}
		if prev.Position == nil || *prev.Position != *keep[i].Position {
			return false
		}
		if exactNumericCanonical(prev.Quantity) != exactNumericCanonical(keep[i].Quantity) {
			return false
		}
		if derefPackageFK(prev.ConsultationID) != derefPackageFK(keep[i].ConsultationID) ||
			derefPackageFK(prev.ProcedureID) != derefPackageFK(keep[i].ProcedureID) ||
			derefPackageFK(prev.MedicineID) != derefPackageFK(keep[i].MedicineID) ||
			derefPackageFK(prev.ExamTypeID) != derefPackageFK(keep[i].ExamTypeID) ||
			derefPackageFK(prev.VaccineID) != derefPackageFK(keep[i].VaccineID) ||
			derefPackageFK(prev.CheckupTypeID) != derefPackageFK(keep[i].CheckupTypeID) ||
			derefPackageFK(prev.MerchandiseItemID) != derefPackageFK(keep[i].MerchandiseItemID) {
			return false
		}
		if string(derefRaw(prev.ManualDefinition)) != string(derefRaw(keep[i].ManualDefinition)) {
			return false
		}
	}
	return true
}

func derefRaw(r *json.RawMessage) []byte {
	if r == nil {
		return nil
	}
	return *r
}

func (s *treatmentPackageService) Delete(ctx context.Context, clinicID, id uint64, version int64, actorID uint64) error {
	if version < 1 || actorID == 0 {
		return apperrors.WrapInvalidInput("positive version and staff actor are required")
	}
	if err := s.requireDeps(); err != nil {
		return err
	}
	err := s.deps.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		now := time.Now()
		if err := s.repo.LockClinicRow(txCtx, clinicID); err != nil {
			return err
		}
		// delete は catalog を package より先に lock する。
		revision, err := s.repo.LockCatalogRow(txCtx, clinicID)
		if err != nil {
			return err
		}
		locked, err := s.repo.LockPackagesByIDAsc(txCtx, clinicID, []uint64{id})
		if err != nil {
			return err
		}
		if len(locked) == 0 {
			return apperrors.WrapNotFound("treatment_package", "not found")
		}
		// F1: 削除も書込み前に live items の型別権限を検査する（zero-write）。
		liveItems, err := s.repo.FindLiveItemsByPackageIDs(txCtx, clinicID, []uint64{id})
		if err != nil {
			return err
		}
		if err := s.authorizeItemKinds(txCtx, clinicID, liveItems); err != nil {
			return err
		}
		if err := s.repo.SoftDeleteCAS(txCtx, clinicID, id, version, actorID, now); err != nil {
			return err
		}
		if err := s.repo.IncrementCatalogRevision(txCtx, clinicID, revision); err != nil {
			return err
		}
		return s.deps.Audit.LogEntryTx(txCtx, &AuditEntry{
			ClinicID: &clinicID, ActorID: &actorID, ActorType: model.AuditActorTypeStaff,
			Action: model.AuditActionTreatmentPackageDelete, Resource: model.AuditResourceTreatmentPackage, ResourceID: &id,
			OldValue: map[string]any{"version": version},
		})
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "treatment package deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("package_id", id))
	return nil
}

func (s *treatmentPackageService) Reorder(ctx context.Context, clinicID uint64, orderVersion int64, ids []uint64, actorID uint64) error {
	if orderVersion < 1 || actorID == 0 {
		return apperrors.WrapInvalidInput("positive order_version and staff actor are required")
	}
	if err := s.requireDeps(); err != nil {
		return err
	}
	err := s.deps.Transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.LockClinicRow(txCtx, clinicID); err != nil {
			return err
		}
		revision, err := s.repo.LockCatalogRow(txCtx, clinicID)
		if err != nil {
			return err
		}
		if revision != orderVersion {
			// expected catalog revision mismatch -> zero-write conflict。
			return apperrors.WrapConflict("並び順が更新されています。再読み込みしてから保存してください")
		}
		current, err := s.repo.FindAll(txCtx, clinicID)
		if err != nil {
			return err
		}
		currentIDs := make([]uint64, 0, len(current))
		for i := range current {
			currentIDs = append(currentIDs, current[i].ID)
		}
		locked, err := s.repo.LockPackagesByIDAsc(txCtx, clinicID, currentIDs)
		if err != nil {
			return err
		}
		// 非削除全件（inactive 含む）と完全一致のみ受理。foreign/欠落/重複は全体拒否。
		if len(ids) != len(locked) {
			return apperrors.WrapInvalidInput("並び替えは現在の全パッケージを対象にしてください")
		}
		wanted := make(map[uint64]int, len(ids))
		for i, id := range ids {
			wanted[id] = i
		}
		currentOrder := make(map[uint64]int, len(current))
		for i := range current {
			currentOrder[current[i].ID] = i
		}
		moved := false
		// UPDATE は id 昇順で行い行ロック取得順を確定させる。
		for i := range locked {
			id := locked[i].ID
			target, ok := wanted[id]
			if !ok {
				return apperrors.WrapInvalidInput("並び替え対象外の id が含まれています")
			}
			if target != locked[i].SortOrder && currentOrder[id] != target {
				if err := s.repo.UpdateSortOrderBumpVersion(txCtx, clinicID, id, target); err != nil {
					return err
				}
				moved = true
			}
		}
		for id := range wanted {
			if _, ok := currentOrder[id]; !ok {
				return apperrors.WrapInvalidInput("存在しない id が含まれています")
			}
		}
		if !moved {
			return nil // no-op は版不変。
		}
		if err := s.repo.IncrementCatalogRevision(txCtx, clinicID, revision); err != nil {
			return err
		}
		return s.deps.Audit.LogEntryTx(txCtx, &AuditEntry{
			ClinicID: &clinicID, ActorID: &actorID, ActorType: model.AuditActorTypeStaff,
			Action: model.AuditActionTreatmentPackageReorder, Resource: model.AuditResourceTreatmentPackage,
			NewValue: map[string]any{"count": len(ids)},
		})
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "treatment packages reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
