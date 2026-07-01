package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockCampaignRepository struct {
	repository.CampaignRepository
	findAllFn                  func(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	findByIDFn                 func(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	createFn                   func(ctx context.Context, m *model.Campaign) (*model.Campaign, error)
	updateFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error)
	replaceTargetsFn           func(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
	reorderFn                  func(ctx context.Context, clinicID uint64, ids []uint64) error
	findApplicableForItemFn    func(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error)
	findAllApplicableForItemFn func(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error)
}

func (m *mockCampaignRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.Campaign{}, nil
}

func (m *mockCampaignRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Campaign{ID: id, ClinicID: clinicID}, nil
}

func (m *mockCampaignRepository) Create(ctx context.Context, campaign *model.Campaign) (*model.Campaign, error) {
	if m.createFn != nil {
		return m.createFn(ctx, campaign)
	}
	return campaign, nil
}

func (m *mockCampaignRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.Campaign{ID: id, ClinicID: clinicID}, nil
}

func (m *mockCampaignRepository) ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error {
	if m.replaceTargetsFn != nil {
		return m.replaceTargetsFn(ctx, campaignID, categories, itemIDs)
	}
	return nil
}

func (m *mockCampaignRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockCampaignRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockCampaignRepository) FindApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error) {
	if m.findApplicableForItemFn != nil {
		return m.findApplicableForItemFn(ctx, clinicID, date, category, merchandiseItemID)
	}
	return nil, nil
}

func (m *mockCampaignRepository) FindAllApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error) {
	if m.findAllApplicableForItemFn != nil {
		return m.findAllApplicableForItemFn(ctx, clinicID, date, category, merchandiseItemID)
	}
	return nil, nil
}

func TestCampaignService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findAllFn: func(_ context.Context, clinicID uint64) ([]model.Campaign, error) {
				return []model.Campaign{{ID: 1, ClinicID: clinicID}}, nil
			},
		}
		svc := NewCampaignService(repo)
		res, err := svc.List(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.Campaign, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCampaignService(repo)
		res, err := svc.List(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCampaignService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Campaign, error) {
				return &model.Campaign{ID: id, ClinicID: clinicID}, nil
			},
		}
		svc := NewCampaignService(repo)
		res, err := svc.GetByID(ctx, 1, 100)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), res.ID)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCampaignService(repo)
		res, err := svc.GetByID(ctx, 1, 100)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestCampaignService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success rate discount", func(t *testing.T) {
		repo := &mockCampaignRepository{
			createFn: func(_ context.Context, m *model.Campaign) (*model.Campaign, error) {
				m.ID = 1
				return m, nil
			},
		}
		svc := NewCampaignService(repo)
		input := &CreateCampaignInput{
			Name:             "Autumn Sale",
			StartDate:        time.Now(),
			EndDate:          time.Now().Add(24 * time.Hour),
			DiscountType:     model.CampaignDiscountTypeRate,
			DiscountValue:    10.0,
			IsActive:         true,
			SortOrder:        1,
			TargetCategories: []model.ItemCategory{model.ItemCategoryFood},
			TargetItemIDs:    []uint64{10},
		}
		res, err := svc.Create(ctx, 1, input)
		assert.NoError(t, err)
		assert.Equal(t, "Autumn Sale", res.Name)
	})

	t.Run("invalid name", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		input := &CreateCampaignInput{
			Name: "",
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid period", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		input := &CreateCampaignInput{
			Name:      "Bad Period",
			StartDate: time.Now().Add(24 * time.Hour),
			EndDate:   time.Now(),
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid rate discount", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		input := &CreateCampaignInput{
			Name:          "Bad Discount",
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 105.0,
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid amount discount", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		input := &CreateCampaignInput{
			Name:          "Bad Discount",
			DiscountType:  model.CampaignDiscountTypeAmount,
			DiscountValue: -100,
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("invalid discount type", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		input := &CreateCampaignInput{
			Name:          "Bad Discount",
			DiscountType:  "invalid",
			DiscountValue: 10,
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockCampaignRepository{
			createFn: func(_ context.Context, _ *model.Campaign) (*model.Campaign, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCampaignService(repo)
		input := &CreateCampaignInput{
			Name:          "DB Error Test",
			DiscountType:  model.CampaignDiscountTypeAmount,
			DiscountValue: 100,
		}
		_, err := svc.Create(ctx, 1, input)
		assert.Error(t, err)
	})
}

func TestCampaignService_Update(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success update fields", func(t *testing.T) {
		current := &model.Campaign{
			ID:            100,
			ClinicID:      1,
			Name:          "Original",
			StartDate:     now,
			EndDate:       now.Add(time.Hour),
			DiscountType:  model.CampaignDiscountTypeAmount,
			DiscountValue: 500,
		}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Campaign, error) {
				assert.Equal(t, "Updated", fields["name"])
				return current, nil
			},
		}
		svc := NewCampaignService(repo)
		newName := "Updated"
		input := &UpdateCampaignInput{
			Name: &newName,
		}
		_, err := svc.Update(ctx, 1, 100, input)
		assert.NoError(t, err)
	})

	t.Run("nil input", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		_, err := svc.Update(ctx, 1, 100, nil)
		assert.Error(t, err)
	})

	t.Run("no field update", func(t *testing.T) {
		current := &model.Campaign{ID: 100, ClinicID: 1}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
		}
		svc := NewCampaignService(repo)
		input := &UpdateCampaignInput{}
		_, err := svc.Update(ctx, 1, 100, input)
		assert.Error(t, err)
	})

	t.Run("invalid optional name", func(t *testing.T) {
		current := &model.Campaign{ID: 100, ClinicID: 1}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
		}
		svc := NewCampaignService(repo)
		badName := ""
		input := &UpdateCampaignInput{Name: &badName}
		_, err := svc.Update(ctx, 1, 100, input)
		assert.Error(t, err)
	})

	t.Run("targets replace only", func(t *testing.T) {
		current := &model.Campaign{ID: 100, ClinicID: 1}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, categories []model.ItemCategory, itemIDs []uint64) error {
				assert.Len(t, categories, 1)
				assert.Len(t, itemIDs, 1)
				return nil
			},
		}
		svc := NewCampaignService(repo)
		cats := []model.ItemCategory{model.ItemCategoryFood}
		ids := []uint64{20}
		input := &UpdateCampaignInput{
			TargetCategories: &cats,
			TargetItemIDs:    &ids,
		}
		_, err := svc.Update(ctx, 1, 100, input)
		assert.NoError(t, err)
	})

	t.Run("replace targets db error", func(t *testing.T) {
		current := &model.Campaign{ID: 100, ClinicID: 1}
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return current, nil
			},
			replaceTargetsFn: func(_ context.Context, _ uint64, _ []model.ItemCategory, _ []uint64) error {
				return errors.New("db error")
			},
		}
		svc := NewCampaignService(repo)
		cats := []model.ItemCategory{model.ItemCategoryFood}
		input := &UpdateCampaignInput{
			TargetCategories: &cats,
		}
		_, err := svc.Update(ctx, 1, 100, input)
		assert.Error(t, err)
	})
}

func TestCampaignService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Campaign, error) {
				return &model.Campaign{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		}
		svc := NewCampaignService(repo)
		err := svc.Delete(ctx, 1, 100)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewCampaignService(repo)
		err := svc.Delete(ctx, 1, 100)
		assert.Error(t, err)
	})

	t.Run("db delete error", func(t *testing.T) {
		repo := &mockCampaignRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Campaign, error) {
				return &model.Campaign{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db delete error")
			},
		}
		svc := NewCampaignService(repo)
		err := svc.Delete(ctx, 1, 100)
		assert.Error(t, err)
	})
}

func TestCampaignService_Reorder(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCampaignRepository{
			reorderFn: func(_ context.Context, _ uint64, ids []uint64) error {
				assert.Equal(t, []uint64{3, 2, 1}, ids)
				return nil
			},
		}
		svc := NewCampaignService(repo)
		err := svc.Reorder(ctx, 1, []uint64{3, 2, 1})
		assert.NoError(t, err)
	})

	t.Run("empty IDs error", func(t *testing.T) {
		svc := NewCampaignService(&mockCampaignRepository{})
		err := svc.Reorder(ctx, 1, []uint64{})
		assert.Error(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockCampaignRepository{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return errors.New("db error")
			},
		}
		svc := NewCampaignService(repo)
		err := svc.Reorder(ctx, 1, []uint64{1})
		assert.Error(t, err)
	})
}
