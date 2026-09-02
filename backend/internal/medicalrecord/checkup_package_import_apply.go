package medicalrecord

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *checkupPackageImportService) importCheckupTypes(
	db *gorm.DB,
	clinicID uint64,
	canonical *CanonicalCheckupPackage,
) (map[string]uint64, int, error) {
	typeIDByKey := make(map[string]uint64, len(canonical.Manifest.Types))
	var roots, children []CheckupPackageTypeDef
	for _, t := range canonical.Manifest.Types {
		if t.ParentKey == nil {
			roots = append(roots, t)
		} else {
			children = append(children, t)
		}
	}
	typesCreated := 0
	for _, pass := range [][]CheckupPackageTypeDef{roots, children} {
		for _, t := range pass {
			var parentID *uint64
			if t.ParentKey != nil {
				id, ok := typeIDByKey[*t.ParentKey]
				if !ok {
					return nil, 0, apperrors.WrapInvalidInput(fmt.Sprintf("parent type %q not resolved", *t.ParentKey))
				}
				parentID = &id
			}
			ns := canonical.Manifest.Namespace
			key := t.Key
			row := model.CheckupType{
				ClinicID:        clinicID,
				Name:            t.Name,
				IsActive:        t.IsActive,
				Description:     t.Description,
				Interval:        t.Interval,
				TargetAge:       t.TargetAge,
				ParentID:        parentID,
				ImportNamespace: &ns,
				ImportKey:       &key,
				SortOrder:       t.SortOrder,
			}
			if err := db.Create(&row).Error; err != nil {
				return nil, 0, apperrors.FromGORM(err, "checkup_type", t.Key)
			}
			typeIDByKey[t.Key] = row.ID
			typesCreated++
		}
	}
	return typeIDByKey, typesCreated, nil
}

func (s *checkupPackageImportService) importCheckupFields(
	db *gorm.DB,
	clinicID uint64,
	canonical *CanonicalCheckupPackage,
	typeIDByKey map[string]uint64,
) (map[string]uint64, int, error) {
	fieldIDByKey := make(map[string]uint64, len(canonical.Manifest.Fields))
	fieldsCreated := 0
	for _, f := range canonical.Manifest.Fields {
		typeID, ok := typeIDByKey[f.TypeKey]
		if !ok {
			return nil, 0, apperrors.WrapInvalidInput(fmt.Sprintf("field type_key %q missing", f.TypeKey))
		}
		opts, err := json.Marshal(f.Options)
		if err != nil {
			return nil, 0, apperrors.Wrap(err, "marshal field options")
		}
		var minV, maxV *float64
		if f.MinValue != nil {
			v, err := strconv.ParseFloat(*f.MinValue, 64)
			if err != nil {
				return nil, 0, apperrors.WrapInvalidInput("invalid min_value")
			}
			minV = &v
		}
		if f.MaxValue != nil {
			v, err := strconv.ParseFloat(*f.MaxValue, 64)
			if err != nil {
				return nil, 0, apperrors.WrapInvalidInput("invalid max_value")
			}
			maxV = &v
		}
		ns := canonical.Manifest.Namespace
		key := f.Key
		row := model.CheckupTypeField{
			ClinicID:        clinicID,
			CheckupTypeID:   typeID,
			Name:            f.Name,
			FieldType:       model.CheckupFieldType(f.FieldType),
			Unit:            f.Unit,
			MinValue:        minV,
			MaxValue:        maxV,
			Options:         datatypes.JSON(opts),
			IsProvisional:   false,
			ImportNamespace: &ns,
			ImportKey:       &key,
			SortOrder:       f.SortOrder,
		}
		if err := db.Create(&row).Error; err != nil {
			return nil, 0, apperrors.FromGORM(err, "checkup_type_field", f.Key)
		}
		fieldIDByKey[f.Key] = row.ID
		fieldsCreated++
	}
	return fieldIDByKey, fieldsCreated, nil
}
