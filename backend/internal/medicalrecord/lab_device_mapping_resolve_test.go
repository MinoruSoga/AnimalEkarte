package medicalrecord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestResolveItems_SpeciesEmptyDoesNotReadSpecs(t *testing.T) {
	fieldID := uint64(4)
	repo := &mappingSpecRepo{
		masters: []model.LabDeviceItemMaster{{
			ClinicID: 1, SourceType: "fuji_nx600", DeviceItemCode: "BUN-P",
			Unit: "mg/dl", ValueShape: model.LabDeviceValueShapeNumeric,
			ExamTypeFieldID: &fieldID, IsActive: true,
		}},
		specs: []model.LabDeviceMappingSpec{{
			ClinicID: 1, SourceType: "fuji_nx600", Species: "canine",
			DeviceItemCode: "BUN-P", Unit: "g/dl", ValueShape: model.LabDeviceValueShapeNumeric,
			ExamTypeFieldID: 9,
		}},
		fields: map[uint64]model.ExamTypeField{
			4: {ID: 4, ExamTypeID: 2, Name: "BUN"},
			9: {ID: 9, ExamTypeID: 2, Name: "BUN-dog"},
		},
	}
	svc := NewLabDeviceItemMasterService(repo)
	ctx := context.Background()

	res, err := svc.ResolveItems(ctx, 1, "fuji_nx600", []string{"BUN-P"})
	require.NoError(t, err)
	assert.Zero(t, repo.specCalls)
	require.Len(t, res.Mapped, 1)
	assert.Equal(t, uint64(4), res.Mapped[0].ExamTypeFieldID)

	res, err = svc.ResolveItemsForSpecies(ctx, 1, "fuji_nx600", "canine", []string{"BUN-P"})
	require.NoError(t, err)
	require.Len(t, res.Mapped, 1)
	assert.Equal(t, uint64(9), res.Mapped[0].ExamTypeFieldID)
	assert.Equal(t, "g/dl", res.Mapped[0].Unit)
	assert.Equal(t, 1, repo.specCalls)

	res, err = svc.ResolveItemsForSpecies(ctx, 1, "fuji_nx600", "feline", []string{"BUN-P"})
	require.NoError(t, err)
	require.Len(t, res.Mapped, 1)
	assert.Equal(t, uint64(4), res.Mapped[0].ExamTypeFieldID)
	assert.Equal(t, 2, repo.specCalls)
}

type mappingSpecRepo struct {
	masters   []model.LabDeviceItemMaster
	specs     []model.LabDeviceMappingSpec
	fields    map[uint64]model.ExamTypeField
	specCalls int
}

func (r *mappingSpecRepo) FindByClinicSourceCodes(context.Context, uint64, string, []string) ([]model.LabDeviceItemMaster, error) {
	return append([]model.LabDeviceItemMaster(nil), r.masters...), nil
}

func (r *mappingSpecRepo) FindMappingSpecs(_ context.Context, _ uint64, _ string, species string, _ []string) ([]model.LabDeviceMappingSpec, error) {
	r.specCalls++
	out := make([]model.LabDeviceMappingSpec, 0)
	for _, spec := range r.specs {
		if spec.Species == species {
			out = append(out, spec)
		}
	}
	return out, nil
}

func (r *mappingSpecRepo) FindExamTypeFields(context.Context, uint64, []uint64) (map[uint64]model.ExamTypeField, error) {
	return r.fields, nil
}

func (r *mappingSpecRepo) List(context.Context, uint64, string) ([]model.LabDeviceItemMaster, error) {
	return nil, nil
}
func (r *mappingSpecRepo) FindByID(context.Context, uint64, uint64) (*model.LabDeviceItemMaster, error) {
	return nil, nil
}
func (r *mappingSpecRepo) EnsureCatalog(context.Context, []model.LabDeviceItemMaster) (int64, error) {
	return 0, nil
}
func (r *mappingSpecRepo) Update(context.Context, uint64, uint64, UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error) {
	return nil, nil
}
func (r *mappingSpecRepo) FindExamTypeField(context.Context, uint64, uint64) (*model.ExamTypeField, error) {
	return nil, nil
}
func (r *mappingSpecRepo) FindExamType(context.Context, uint64, uint64) (*model.ExaminationType, error) {
	return nil, nil
}
func (r *mappingSpecRepo) ListDevices(context.Context, uint64) ([]model.LabDevice, error) {
	return nil, nil
}
func (r *mappingSpecRepo) FindDeviceByID(context.Context, uint64, uint64) (*model.LabDevice, error) {
	return nil, nil
}
func (r *mappingSpecRepo) CreateDevice(context.Context, *model.LabDevice) error { return nil }
func (r *mappingSpecRepo) UpdateDevice(context.Context, uint64, uint64, UpdateLabDeviceInput) (*model.LabDevice, error) {
	return nil, nil
}
func (r *mappingSpecRepo) EnsureDevices(context.Context, []model.LabDevice) (int64, error) {
	return 0, nil
}
