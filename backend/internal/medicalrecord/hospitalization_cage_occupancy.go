package medicalrecord

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// CageOccupancyForAssignment reads the selected-clinic projection through the
// same repository/core as the save guards. It does not reserve a candidate.
func (s *hospitalizationService) CageOccupancyForAssignment(ctx context.Context, clinicID uint64, excludeHospitalizationID *uint64) (hospitalizationCageOccupancyResponse, error) {
	source, ok := s.cageRepo.(hospitalizationCageOccupancySource)
	if !ok {
		return hospitalizationCageOccupancyResponse{}, errCageOccupancyReadFailed
	}
	if persistence.TxFromContext(ctx) != nil {
		return readHospitalizationCageOccupancy(ctx, clinicID, excludeHospitalizationID, s.hospRepo, source)
	}
	if s.transactor == nil {
		return hospitalizationCageOccupancyResponse{}, errCageOccupancyReadFailed
	}
	var result hospitalizationCageOccupancyResponse
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		result, err = readHospitalizationCageOccupancy(txCtx, clinicID, excludeHospitalizationID, s.hospRepo, source)
		return err
	})
	return result, err
}

// GetCageOccupancy serves the selected EMR-461/740 ID-only read contract.
func (h *HospitalizationHandler) GetCageOccupancy(c *gin.Context) {
	reader, _ := h.service.(HospitalizationCageOccupancyReader)
	serveHospitalizationCageOccupancy(c, reader)
}
