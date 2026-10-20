package medicalrecord

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// HospitalizationCageOccupancyReader supplies the complete ID projection, with
// unresolved evidence represented in the result rather than as a read failure.
type HospitalizationCageOccupancyReader interface {
	CageOccupancyForAssignment(context.Context, uint64, *uint64) (hospitalizationCageOccupancyResponse, error)
}

type hospitalizationCageCurrentAssignment struct {
	State  string  `json:"state"`
	CageID *string `json:"cage_id"`
}

type hospitalizationCageOccupancyResponse struct {
	ClinicID                 string                                `json:"clinic_id"`
	ExcludeHospitalizationID *string                               `json:"exclude_hospitalization_id"`
	State                    string                                `json:"state"`
	OccupiedCageIDs          []string                              `json:"occupied_cage_ids"`
	CandidateCageIDs         []string                              `json:"candidate_cage_ids"`
	CurrentAssignment        *hospitalizationCageCurrentAssignment `json:"current_assignment"`
}

// parseCageOccupancyTarget accepts only the selected query and PostgreSQL bigint
// IDs. Presence with an empty value is invalid, unlike an omitted target.
func parseCageOccupancyTarget(query url.Values) (*uint64, error) {
	for key := range query {
		if key != "exclude_hospitalization_id" {
			return nil, apperrors.WrapInvalidInput("unsupported cage occupancy query")
		}
	}
	values, present := query["exclude_hospitalization_id"]
	if !present {
		return nil, nil
	}
	if len(values) != 1 || len(values[0]) == 0 || values[0][0] < '1' || values[0][0] > '9' {
		return nil, apperrors.WrapInvalidInput("exclude_hospitalization_id must be a positive bigint id")
	}
	for _, digit := range values[0] {
		if digit < '0' || digit > '9' {
			return nil, apperrors.WrapInvalidInput("exclude_hospitalization_id must be a positive bigint id")
		}
	}
	id, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("exclude_hospitalization_id must be a positive bigint id")
	}
	value := uint64(id)
	return &value, nil
}

// requireSelectedClinicHospitalizationWriteGrant rechecks the selected clinic;
// passing middleware using another assigned clinic's grant is insufficient.
func requireSelectedClinicHospitalizationWriteGrant(c *gin.Context) bool {
	if isAdmin, ok := httpapi.PeekIsSystemAdmin(c); ok && isAdmin {
		return true
	}
	check, ok := httpapi.PeekClinicPermissionChecker(c)
	clinicID, hasClinic := httpapi.PeekClinicID(c)
	if ok && hasClinic && clinicID != 0 {
		resource := string(model.ResourceHospitalization)
		if check(c, clinicID, resource, "create") || check(c, clinicID, resource, "edit") {
			return true
		}
	}
	httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
	return false
}

func serveHospitalizationCageOccupancy(c *gin.Context, reader HospitalizationCageOccupancyReader) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !requireSelectedClinicHospitalizationWriteGrant(c) {
		return
	}
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid cage occupancy query"))
		return
	}
	exclude, err := parseCageOccupancyTarget(query)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if exclude != nil && !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceHospitalization), "edit") {
		return
	}
	if reader == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("cage occupancy reader is not configured"))
		return
	}
	result, err := reader.CageOccupancyForAssignment(c.Request.Context(), clinicID, exclude)
	if err != nil {
		if apperrors.IsNotFound(err) && !errors.Is(err, errCageOccupancyReadFailed) {
			httpapi.RespondError(c, err)
			return
		}
		// The selected contract maps acquisition failures to 500, including a
		// dependency returning 503. Keep the cause for request-boundary logging.
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}
