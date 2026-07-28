package reservation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestBuildReservationTypeUpdate(t *testing.T) {
	t.Run("empty input produces empty field map", func(t *testing.T) {
		fields := buildReservationTypeUpdate(&UpdateReservationTypeInput{})
		assert.Empty(t, fields)
	})

	t.Run("all scalar fields set are mapped to their column names", func(t *testing.T) {
		name := "一般診療"
		color := "#FF0000"
		isActive := true
		description := "説明"
		sortOrder := 3
		category := "general"
		dispName := "診療"
		duration := 30
		maxConcurrent := 2
		shortName := "一般"
		showShortName := true
		visible := true
		comment := "コメント"
		imageURL := "https://example.com/image.png"
		dayOption := "weekday"
		isInternal := false

		input := &UpdateReservationTypeInput{
			Name:                   &name,
			Color:                  &color,
			IsActive:               &isActive,
			Description:            &description,
			SortOrder:              &sortOrder,
			Category:               &category,
			ReservationDisplayName: &dispName,
			DurationMinutes:        &duration,
			MaxConcurrent:          &maxConcurrent,
			ShortName:              &shortName,
			ShowShortName:          &showShortName,
			ReservationVisible:     &visible,
			ReservationComment:     &comment,
			ReservationImageURL:    &imageURL,
			ReservationDayOption:   &dayOption,
			IsInternal:             &isInternal,
		}
		fields := buildReservationTypeUpdate(input)

		assert.Equal(t, name, fields[colReservationTypeName])
		assert.Equal(t, color, fields[colReservationTypeColor])
		assert.Equal(t, isActive, fields[colReservationTypeIsActive])
		assert.Equal(t, description, fields[colReservationTypeDescription])
		assert.Equal(t, sortOrder, fields[colReservationTypeSortOrder])
		assert.Equal(t, model.ReservationTypeCategory(category), fields[colReservationTypeCategory])
		assert.Equal(t, dispName, fields[colReservationTypeReservationDispName])
		assert.Equal(t, duration, fields[colReservationTypeDurationMinutes])
		assert.Equal(t, maxConcurrent, fields[colReservationTypeMaxConcurrent])
		assert.Equal(t, shortName, fields[colReservationTypeShortName])
		assert.Equal(t, showShortName, fields[colReservationTypeShowShortName])
		assert.Equal(t, visible, fields[colReservationTypeReservationVisible])
		assert.Equal(t, comment, fields[colReservationTypeReservationComment])
		assert.Equal(t, imageURL, fields[colReservationTypeReservationImageURL])
		assert.Equal(t, dayOption, fields[colReservationTypeReservationDayOpt])
		assert.Equal(t, isInternal, fields[colReservationTypeIsInternal])
	})

	t.Run("ClearMaxConcurrent takes precedence over MaxConcurrent", func(t *testing.T) {
		maxConcurrent := 5
		input := &UpdateReservationTypeInput{
			ClearMaxConcurrent: true,
			MaxConcurrent:      &maxConcurrent,
		}
		fields := buildReservationTypeUpdate(input)
		assert.Contains(t, fields, colReservationTypeMaxConcurrent)
		assert.Nil(t, fields[colReservationTypeMaxConcurrent])
	})

	t.Run("MaxConcurrent alone (no clear) sets the value", func(t *testing.T) {
		maxConcurrent := 5
		input := &UpdateReservationTypeInput{MaxConcurrent: &maxConcurrent}
		fields := buildReservationTypeUpdate(input)
		assert.Equal(t, maxConcurrent, fields[colReservationTypeMaxConcurrent])
	})

	t.Run("ClearGroupID clears group_id", func(t *testing.T) {
		input := &UpdateReservationTypeInput{ClearGroupID: true}
		fields := buildReservationTypeUpdate(input)
		assert.Contains(t, fields, colReservationTypeGroupID)
		assert.Nil(t, fields[colReservationTypeGroupID])
	})

	t.Run("GroupID sets group_id when not clearing", func(t *testing.T) {
		groupID := uint64(9)
		input := &UpdateReservationTypeInput{GroupID: &groupID}
		fields := buildReservationTypeUpdate(input)
		assert.Equal(t, groupID, fields[colReservationTypeGroupID])
	})

	t.Run("ClearParentID clears parent_id", func(t *testing.T) {
		input := &UpdateReservationTypeInput{ClearParentID: true}
		fields := buildReservationTypeUpdate(input)
		assert.Contains(t, fields, colReservationTypeParentID)
		assert.Nil(t, fields[colReservationTypeParentID])
	})

	t.Run("ParentID sets parent_id when not clearing", func(t *testing.T) {
		parentID := uint64(4)
		input := &UpdateReservationTypeInput{ParentID: &parentID}
		fields := buildReservationTypeUpdate(input)
		assert.Equal(t, parentID, fields[colReservationTypeParentID])
	})
}
