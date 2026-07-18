package service

// P13 definition order: const → buildFunc → interface → struct → constructor → methods

import (
	"context"
	"fmt"

	slog "log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

type lstepTagCodeMappingService struct {
	repo repository.LstepTagCodeMappingRepository
}

// NewLstepTagCodeMappingService はサービスを生成する。
func NewLstepTagCodeMappingService(repo repository.LstepTagCodeMappingRepository) LstepTagCodeMappingService {
	return &lstepTagCodeMappingService{repo: repo}
}

func (s *lstepTagCodeMappingService) ListMappings(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	mappings, err := s.repo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list lstep tag code mappings", "clinic_id", clinicID, "error", err)
		return nil, apperrors.Wrap(err, "failed to list lstep tag code mappings")
	}
	return mappings, nil
}

func (s *lstepTagCodeMappingService) PutMappingsForTag(ctx context.Context, clinicID uint64, tagName string, entries []PutMappingEntry) ([]*model.LstepTagCodeMapping, error) {
	if !isConfigurableTag(tagName) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("tag %q is not configurable", tagName))
	}

	if err := s.repo.SoftDeleteByClinicIDAndTagName(ctx, clinicID, tagName); err != nil {
		slog.ErrorContext(ctx, "failed to soft-delete lstep tag code mappings", "clinic_id", clinicID, "tag_name", tagName, "error", err)
		return nil, apperrors.Wrap(err, "failed to delete existing mappings")
	}

	created := make([]*model.LstepTagCodeMapping, 0, len(entries))
	for i := range entries {
		m := &model.LstepTagCodeMapping{
			ClinicID: clinicID,
			TagName:  tagName,
			CodeType: entries[i].CodeType,
			Codes:    entries[i].Codes,
		}
		if entries[i].SpeciesScope != "" {
			s := entries[i].SpeciesScope
			m.SpeciesScope = &s
		}
		if entries[i].AgeMin != nil {
			m.AgeMin = entries[i].AgeMin
		}
		if err := s.repo.Create(ctx, m); err != nil {
			slog.ErrorContext(ctx, "failed to create lstep tag code mapping", "clinic_id", clinicID, "tag_name", tagName, "error", err)
			return nil, apperrors.Wrap(err, "failed to create mapping")
		}
		created = append(created, m)
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
