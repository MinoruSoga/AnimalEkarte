package handler

import (
	"net/url"
	"strconv"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

type unbilledItemsQuery struct {
	PetID string
}

func newUnbilledItemsQuery(values url.Values) unbilledItemsQuery {
	return unbilledItemsQuery{PetID: values.Get("pet_id")}
}

func (q unbilledItemsQuery) toPetID() (uint64, error) {
	if q.PetID == "" {
		return 0, apperrors.WrapInvalidInput("pet_id は必須です")
	}
	petID, err := strconv.ParseUint(q.PetID, 10, 64)
	if err != nil || petID == 0 {
		return 0, apperrors.WrapInvalidInput("pet_id の形式が不正です")
	}
	return petID, nil
}
