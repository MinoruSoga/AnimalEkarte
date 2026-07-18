package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/procedure"
)

// ProcedureRepository is a stable facade alias for the procedure domain package.
type ProcedureRepository = procedure.Repository

// NewProcedureRepository constructs the procedure repository.
func NewProcedureRepository(db *gorm.DB) ProcedureRepository {
	return procedure.New(db)
}
