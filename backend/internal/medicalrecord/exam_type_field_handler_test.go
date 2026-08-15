package medicalrecord

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestExamReferenceRangeHandler_RequiresNonNullRanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewExamTypeHandler(&mockExamTypeService{})

	for _, body := range []string{`{}`, `{"ranges":null}`} {
		t.Run(body, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(body))
			context.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "fieldId", Value: "2"}}
			setClinicID(context)

			handler.ReplaceExaminationTypeFieldReferenceRanges(context)

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func TestExamReferenceRangeHandler_ReturnsCompleteFieldMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &mockExamTypeService{
		replaceReferenceRangesFn: func(
			_ context.Context,
			clinicID, examTypeID uint64,
			command *ReplaceReferenceRangesCommand,
		) (*ExamTypeFieldResult, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), examTypeID)
			assert.Equal(t, uint64(20), command.ExamTypeFieldID)
			require.Len(t, command.Ranges, 1)
			return &ExamTypeFieldResult{
				Field: model.ExamTypeField{
					ID:         20,
					ExamTypeID: 10,
					Name:       "WBC",
					Unit:       "10^3/μL",
					SortOrder:  3,
				},
				ReferenceRanges: []model.ExamReferenceRange{{
					ID: 30, ExamTypeFieldID: 20, AnimalSpeciesID: 40,
				}},
			}, nil
		},
	}
	handler := NewExamTypeHandler(service)
	recorder := httptest.NewRecorder()
	requestContext, _ := gin.CreateTestContext(recorder)
	requestContext.Request = httptest.NewRequest(
		http.MethodPut,
		"/",
		bytes.NewBufferString(`{"ranges":[{"animal_species_id":40}]}`),
	)
	requestContext.Params = gin.Params{{Key: "id", Value: "10"}, {Key: "fieldId", Value: "20"}}
	setClinicID(requestContext)

	handler.ReplaceExaminationTypeFieldReferenceRanges(requestContext)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"id":20`)
	assert.Contains(t, recorder.Body.String(), `"name":"WBC"`)
	assert.Contains(t, recorder.Body.String(), `"unit":"10^3/μL"`)
	assert.Contains(t, recorder.Body.String(), `"animal_species_id":40`)
}

func TestExamReferenceRangeHandler_EmptyArrayClears(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewExamTypeHandler(&mockExamTypeService{})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(`{"ranges":[]}`))
	context.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "fieldId", Value: "2"}}
	setClinicID(context)

	handler.ReplaceExaminationTypeFieldReferenceRanges(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"reference_ranges":[]`)
}
