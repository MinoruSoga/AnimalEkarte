package auth

import (
	"encoding/json"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToMeResponse_AccountingDocumentSectionOrderEncodesAsArray(t *testing.T) {
	tests := []struct {
		name         string
		sectionOrder pq.StringArray
		want         string
	}{
		{name: "nil", sectionOrder: nil, want: `[]`},
		{name: "empty", sectionOrder: pq.StringArray{}, want: `[]`},
		{
			name:         "configured",
			sectionOrder: pq.StringArray{"clinic_header", "items_table"},
			want:         `["clinic_header","items_table"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := LoginResponse{
				User: ToMeResponse(
					nil,
					&model.Account{IsSystemAdmin: true},
					"1",
					nil,
					[]model.Clinic{{
						ID:                             1,
						Name:                           "Test Clinic",
						IsActive:                       true,
						AccountingDocumentSectionOrder: tt.sectionOrder,
					}},
					nil,
				),
			}

			encoded, err := json.Marshal(response)
			require.NoError(t, err)

			var payload struct {
				User struct {
					Clinic struct {
						AccountingDocumentSectionOrder json.RawMessage `json:"accounting_document_section_order"`
					} `json:"clinic"`
				} `json:"user"`
			}
			require.NoError(t, json.Unmarshal(encoded, &payload))
			require.JSONEq(t, tt.want, string(payload.User.Clinic.AccountingDocumentSectionOrder))
		})
	}
}
