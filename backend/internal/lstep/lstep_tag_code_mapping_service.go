package lstep

// File organization (this file only, non-mandatory convention — order is not enforced by any
// gate; see service/CLAUDE.md): const → buildFunc → interface → struct → constructor → methods

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ConfigurableTagNames は管理画面から設定可能なタグ名一覧。
var ConfigurableTagNames = []string{
	HlthHealthcheckDoneTag,
	PrevFilariaTag,
	PrevFleaTickTag,
	LtvFoodPurchaseTag,
}

// PutMappingEntry は PUT リクエストの 1 code_type エントリ。
type PutMappingEntry struct {
	CodeType     string
	Codes        []string
	SpeciesScope string
	AgeMin       *int
}

// LstepTagCodeMappingService はコード→タグ設定の管理 CRUD を担う。
type LstepTagCodeMappingService interface {
	// ListMappings は clinic の全タグコードマッピングを返す。
	ListMappings(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	// PutMappingsForTag は指定 tagName のマッピングを entries で全置換する（DELETE + INSERT）。
	PutMappingsForTag(ctx context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error)
}

type tagCodeMappingTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

type lstepTagCodeMappingService struct {
	repo       LstepTagCodeMappingRepository
	transactor tagCodeMappingTransactor
}

// NewLstepTagCodeMappingService はサービスを生成する。
func NewLstepTagCodeMappingService(
	repo LstepTagCodeMappingRepository,
	transactor tagCodeMappingTransactor,
) LstepTagCodeMappingService {
	return &lstepTagCodeMappingService{
		repo:       repo,
		transactor: transactor,
	}
}

func (s *lstepTagCodeMappingService) ListMappings(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	mappings, err := s.repo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list lstep tag code mappings")
	}
	return mappings, nil
}

func (s *lstepTagCodeMappingService) PutMappingsForTag(ctx context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error) {
	if !isConfigurableTag(tagName) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("tag %q is not configurable", tagName))
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("lstep tag code mapping transaction dependency is required")
	}
	// Cardinality must be rejected before WithTx so SoftDelete/Create never run (SEC-CS-F11).
	if err := validatePutMappingEntriesCardinality(entries); err != nil {
		return nil, err
	}
	for i := range entries {
		if err := validatePutMappingEntry(&entries[i]); err != nil {
			return nil, err
		}
	}

	created := make([]*model.LstepTagCodeMapping, 0, len(entries))
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.SoftDeleteByClinicIDAndTagName(txCtx, clinicID, tagName); err != nil {
			return apperrors.Wrap(err, "failed to delete existing mappings")
		}

		for i := range entries {
			m := &model.LstepTagCodeMapping{
				ClinicID: clinicID,
				TagName:  tagName,
				CodeType: entries[i].CodeType,
				Codes:    entries[i].Codes,
			}
			if entries[i].SpeciesScope != "" {
				speciesScope := entries[i].SpeciesScope
				m.SpeciesScope = &speciesScope
			}
			if entries[i].AgeMin != nil {
				m.AgeMin = entries[i].AgeMin
			}
			if err := s.repo.Create(txCtx, m); err != nil {
				return apperrors.Wrap(err, "failed to create mapping")
			}
			created = append(created, m)
		}

		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to replace lstep tag code mappings")
	}

	return created, nil
}

func isConfigurableTag(tagName string) bool {
	for _, t := range ConfigurableTagNames {
		if t == tagName {
			return true
		}
	}
	return false
}

// validatePutMappingEntriesCardinality enforces request-level size caps (SEC-CS-F11).
// Must run before WithTx so over-limit input never reaches SoftDelete/Create.
func validatePutMappingEntriesCardinality(entries []PutMappingEntry) error {
	if len(entries) > MaxTagCodeMappingEntries {
		return apperrors.WrapInvalidInput(
			fmt.Sprintf("entries must contain at most %d items", MaxTagCodeMappingEntries),
		)
	}
	totalCodes := 0
	for i := range entries {
		n := len(entries[i].Codes)
		if n > MaxTagCodeMappingCodesPerEntry {
			return apperrors.WrapInvalidInput(
				fmt.Sprintf("codes must contain at most %d items per entry", MaxTagCodeMappingCodesPerEntry),
			)
		}
		totalCodes += n
		if totalCodes > MaxTagCodeMappingTotalCodes {
			return apperrors.WrapInvalidInput(
				fmt.Sprintf("codes total must be at most %d across all entries", MaxTagCodeMappingTotalCodes),
			)
		}
	}
	return nil
}

// validatePutMappingEntry enforces code_type / species_scope / age_min / codes bounds (G2C-02).
func validatePutMappingEntry(e *PutMappingEntry) error {
	switch e.CodeType {
	case model.CodeTypeCheckupType, model.CodeTypePrescription, model.CodeTypeMerchandiseItem:
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid code_type: %s", e.CodeType))
	}
	if len(e.Codes) == 0 {
		return apperrors.WrapInvalidInput("codes must contain at least one entry")
	}
	// Per-entry length is also enforced in validatePutMappingEntriesCardinality;
	// keep the empty-check here and individual code shape bounds below.
	for _, c := range e.Codes {
		if c == "" {
			return apperrors.WrapInvalidInput("codes must not contain empty values")
		}
		if len(c) > 64 {
			return apperrors.WrapInvalidInput("code length must be <= 64")
		}
	}
	switch e.SpeciesScope {
	case "", "dog", "cat", "other", "all":
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid species_scope: %s", e.SpeciesScope))
	}
	if e.AgeMin != nil && *e.AgeMin < 0 {
		return apperrors.WrapInvalidInput("age_min must be >= 0")
	}
	return nil
}
