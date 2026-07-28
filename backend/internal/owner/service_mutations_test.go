package owner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		petCount     int64
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:         "deletes owner successfully",
			clinicID:     1,
			id:           10,
			petCount:     0,
			repoErr:      nil,
			wantErr:      false,
			wantNF:       false,
			wantConflict: false,
		},
		{
			name:         "returns not found error when owner does not exist",
			clinicID:     1,
			id:           999,
			petCount:     0,
			repoErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNF:       true,
			wantConflict: false,
		},
		{
			name:         "returns error on repository failure",
			clinicID:     1,
			id:           10,
			petCount:     0,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNF:       false,
			wantConflict: false,
		},
		{
			name:         "returns conflict error when owner has pets",
			clinicID:     1,
			id:           10,
			petCount:     2,
			repoErr:      nil,
			wantErr:      true,
			wantNF:       false,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				countPetsByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.petCount, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOwnerService_UpdateDeliveryExclusion(t *testing.T) {
	reason := "  配信不要希望  "
	longReason := strings.Repeat("あ", 101)
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		input        UpdateDeliveryExclusionInput
		findErr      error
		updateErr    error
		wantErr      bool
		wantNotFound bool
		wantInvalid  bool
	}{
		{
			name:     "sets delivery_excluded=true with reason",
			clinicID: 1,
			id:       10,
			input:    UpdateDeliveryExclusionInput{Excluded: true, Reason: &reason},
			wantErr:  false,
		},
		{
			name:     "sets delivery_excluded=false without reason",
			clinicID: 1,
			id:       10,
			input:    UpdateDeliveryExclusionInput{Excluded: false, Reason: nil},
			wantErr:  false,
		},
		{
			name:         "returns not found when owner does not exist",
			clinicID:     1,
			id:           999,
			input:        UpdateDeliveryExclusionInput{Excluded: true},
			findErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "returns error on repository update failure",
			clinicID:  1,
			id:        10,
			input:     UpdateDeliveryExclusionInput{Excluded: true},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:        "returns invalid input when reason is too long",
			clinicID:    1,
			id:          10,
			input:       UpdateDeliveryExclusionInput{Excluded: true, Reason: &longReason},
			wantErr:     true,
			wantInvalid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFields map[string]any
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.Owner{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
					capturedFields = fields
					return tt.updateErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.UpdateDeliveryExclusion(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, owner)
				assert.Equal(t, tt.input.Excluded, capturedFields[colDeliveryExcluded])
				assert.Equal(t, tt.input.Excluded, capturedFields[colLstepOptOut])
				if tt.input.Excluded {
					assert.IsType(t, time.Time{}, capturedFields[colLstepOptOutAt])
					gotReason, ok := capturedFields[colDeliveryExcludedReason].(*string)
					assert.True(t, ok)
					if assert.NotNil(t, gotReason) {
						assert.Equal(t, "配信不要希望", *gotReason)
					}
					gotOptOutReason, ok := capturedFields[colLstepOptOutReason].(*string)
					assert.True(t, ok)
					if assert.NotNil(t, gotOptOutReason) {
						assert.Equal(t, "配信不要希望", *gotOptOutReason)
					}
				} else {
					assert.Nil(t, capturedFields[colDeliveryExcludedReason])
					assert.Nil(t, capturedFields[colLstepOptOutAt])
					assert.Nil(t, capturedFields[colLstepOptOutReason])
				}
			}
		})
	}
}

func TestOwnerService_UpdateTransferStatus(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		input        UpdateTransferStatusInput
		membership   model.MembershipType
		findErr      error
		updateErr    error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:     "sets is_transferred=true and transfer_at",
			clinicID: 1,
			id:       10,
			input:    UpdateTransferStatusInput{IsTransferred: true},
			wantErr:  false,
		},
		{
			name:       "sets is_transferred=false and clears transfer_at",
			clinicID:   1,
			id:         10,
			input:      UpdateTransferStatusInput{IsTransferred: false},
			membership: model.MembershipTypeTransferred,
			wantErr:    false,
		},
		{
			name:         "returns not found when owner does not exist",
			clinicID:     1,
			id:           999,
			input:        UpdateTransferStatusInput{IsTransferred: true},
			findErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "returns error on repository update failure",
			clinicID:  1,
			id:        10,
			input:     UpdateTransferStatusInput{IsTransferred: true},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFields map[string]any
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.Owner{ID: tt.id, ClinicID: tt.clinicID, MembershipType: tt.membership}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
					capturedFields = fields
					return tt.updateErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.UpdateTransferStatus(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, owner)
				assert.Equal(t, tt.input.IsTransferred, capturedFields[colIsTransferred])
				if tt.input.IsTransferred {
					_, ok := capturedFields[colTransferAt].(time.Time)
					assert.True(t, ok, "transfer_at should be time.Time when is_transferred=true")
					assert.Equal(t, model.MembershipTypeTransferred, capturedFields[colMembershipType])
				} else {
					assert.Nil(t, capturedFields[colTransferAt])
					if tt.membership == model.MembershipTypeTransferred {
						assert.Equal(t, model.MembershipTypeNonMember, capturedFields[colMembershipType])
					}
				}
			}
		})
	}
}

func TestOwnerService_LinkLineUserID(t *testing.T) {
	lineUserID := "Uabc123"
	existingOwnerID := uint64(99)
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		lineUserID      *string
		findErr         error
		findByLineIDRet *model.Owner
		updateLineErr   error
		wantErr         bool
		wantConflict    bool
		wantNotFound    bool
	}{
		{
			name:       "links LINE user id successfully",
			clinicID:   1,
			id:         10,
			lineUserID: &lineUserID,
			wantErr:    false,
		},
		{
			name:       "unlinks LINE user id (nil) successfully",
			clinicID:   1,
			id:         10,
			lineUserID: nil,
			wantErr:    false,
		},
		{
			name:            "returns 409 conflict when LINE id belongs to another owner",
			clinicID:        1,
			id:              10,
			lineUserID:      &lineUserID,
			findByLineIDRet: &model.Owner{ID: existingOwnerID, ClinicID: 1},
			wantErr:         true,
			wantConflict:    true,
		},
		{
			name:            "no conflict when LINE id belongs to same owner",
			clinicID:        1,
			id:              10,
			lineUserID:      &lineUserID,
			findByLineIDRet: &model.Owner{ID: 10, ClinicID: 1},
			wantErr:         false,
		},
		{
			name:         "returns not found when owner does not exist",
			clinicID:     1,
			id:           999,
			lineUserID:   &lineUserID,
			findErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:          "returns error on repository update failure",
			clinicID:      1,
			id:            10,
			lineUserID:    &lineUserID,
			updateLineErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.Owner{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
					return tt.findByLineIDRet, nil
				},
				updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
					return tt.updateLineErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			err := svc.LinkLineUserID(context.Background(), tt.clinicID, tt.id, tt.lineUserID, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOwnerService_ConfirmLineID(t *testing.T) {
	lineUserID := "U1234567890"
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		lineUserID   *string
		findErr      error
		updateErr    error
		wantErr      bool
		wantNotFound bool
		wantInvalid  bool
	}{
		{
			name:       "sets line_id_confirmed_at to current time",
			clinicID:   1,
			id:         10,
			lineUserID: &lineUserID,
			wantErr:    false,
		},
		{
			name:        "returns invalid input when line user id is not linked",
			clinicID:    1,
			id:          10,
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:         "returns not found when owner does not exist",
			clinicID:     1,
			id:           999,
			findErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:       "returns error on repository update failure",
			clinicID:   1,
			id:         10,
			lineUserID: &lineUserID,
			updateErr:  errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFields map[string]any
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.Owner{ID: tt.id, ClinicID: tt.clinicID, LineUserID: tt.lineUserID}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
					capturedFields = fields
					return tt.updateErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.ConfirmLineID(context.Background(), tt.clinicID, tt.id, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, owner)
				_, ok := capturedFields[colLineIDConfirmedAt].(time.Time)
				assert.True(t, ok, "line_id_confirmed_at should be set to a time.Time value")
				_, hasConfirmedBy := capturedFields[colLineIDConfirmedBy]
				assert.True(t, hasConfirmedBy, "line_id_confirmed_by should be present in update fields")
			}
		})
	}
}

func TestOwnerService_UpdateDeliveryCaution(t *testing.T) {
	longReason := strings.Repeat("あ", 101)
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		input        UpdateDeliveryCautionInput
		findErr      error
		updateErr    error
		wantErr      bool
		wantNotFound bool
		wantInvalid  bool
	}{
		{
			name:     "sets delivery_caution=true with reason",
			clinicID: 1,
			id:       10,
			input:    UpdateDeliveryCautionInput{Caution: true, Reason: "  注意が必要  "},
			wantErr:  false,
		},
		{
			name:     "sets delivery_caution=false without reason",
			clinicID: 1,
			id:       10,
			input:    UpdateDeliveryCautionInput{Caution: false, Reason: ""},
			wantErr:  false,
		},
		{
			name:         "returns not found when owner does not exist",
			clinicID:     1,
			id:           999,
			input:        UpdateDeliveryCautionInput{Caution: true},
			findErr:      apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "returns error on repository update failure",
			clinicID:  1,
			id:        10,
			input:     UpdateDeliveryCautionInput{Caution: true},
			updateErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:        "returns invalid input when reason is too long",
			clinicID:    1,
			id:          10,
			input:       UpdateDeliveryCautionInput{Caution: true, Reason: longReason},
			wantErr:     true,
			wantInvalid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFields map[string]any
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.Owner{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
					capturedFields = fields
					return tt.updateErr
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, nil)

			owner, err := svc.UpdateDeliveryCaution(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, owner)
				assert.Equal(t, tt.input.Caution, capturedFields[colDeliveryCaution])
				if tt.input.Caution && strings.TrimSpace(tt.input.Reason) != "" {
					gotReason, ok := capturedFields[colDeliveryCautionReason].(*string)
					assert.True(t, ok)
					if assert.NotNil(t, gotReason) {
						assert.Equal(t, strings.TrimSpace(tt.input.Reason), *gotReason)
					}
				} else {
					assert.Nil(t, capturedFields[colDeliveryCautionReason])
				}
			}
		})
	}
}
