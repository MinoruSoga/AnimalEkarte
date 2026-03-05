package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// Service defines the interface for the business logic used by the handler.
type Service interface {
	service.PetService
	service.OwnerService
	service.MedicalRecordService
	service.ReservationService     // from reservation.go
	service.MasterItemService      // from master.go
	service.HospitalizationService // from hospitalization.go
	service.AccountingService      // from accounting.go
	service.ExaminationService     // from examination.go
	service.VaccinationService     // from vaccination.go
	service.TrimmingService        // from trimming.go
	service.ClinicService          // from clinic.go
	service.StaffService           // from clinic.go
	service.InventoryItemService   // from inventory.go
	GetDB() (interface{ DB() *gorm.DB }, error)
}

// Handler contains the HTTP handlers.
type Handler struct {
	svc Service
}

// New creates a new Handler with the given service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ErrorResponse defines the standard error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// ミドルウェア設定
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.CORS())

	// ヘルスチェック
	r.GET("/health", h.Health)

	// API v1
	v1 := r.Group("/api/v1")
	v1.GET("/", h.Welcome)

	// Pets CRUD
	v1.GET("/pets", h.GetPets)
	v1.GET("/pets/:id", h.GetPet)
	v1.POST("/pets", h.CreatePet)
	v1.PUT("/pets/:id", h.UpdatePet)
	v1.DELETE("/pets/:id", h.DeletePet)

	// Owners CRUD
	v1.GET("/owners", h.GetAllOwners)
	v1.POST("/owners", h.CreateOwner)

	// Owner nested routes (register before /:id)
	v1.GET("/owners/:id/medical-records", h.GetMedicalRecordsByOwnerID)
	v1.GET("/owners/:id/reservations", h.GetReservationsByOwnerID)
	v1.GET("/owners/:id/hospitalizations", h.GetHospitalizationsByOwnerID)
	v1.GET("/owners/:id/accountings", h.GetAccountingByOwnerID)
	v1.GET("/owners/:id/examinations", h.GetExaminationsByOwnerID)
	v1.GET("/owners/:id/vaccinations", h.GetVaccinationsByOwnerID)
	v1.GET("/owners/:id/trimmings", h.GetTrimmingsByOwnerID)

	// Owner single resource
	v1.GET("/owners/:id", h.GetOwnerByID)
	v1.PUT("/owners/:id", h.UpdateOwner)
	v1.DELETE("/owners/:id", h.DeleteOwner)

	// Medical Records CRUD
	v1.GET("/medical-records", h.GetAllMedicalRecords)
	v1.GET("/medical-records/paginated", h.GetMedicalRecordsWithPagination)
	v1.GET("/medical-records/:id", h.GetMedicalRecord)
	v1.POST("/medical-records", h.CreateMedicalRecord)
	v1.PUT("/medical-records/:id", h.UpdateMedicalRecord)
	v1.DELETE("/medical-records/:id", h.DeleteMedicalRecord)

	// Pet-based medical records
	v1.GET("/pets/:id/medical-records", h.GetMedicalRecordsByPetID)

	// Reservations CRUD
	v1.GET("/reservations", h.GetAllReservations)
	v1.GET("/reservations/:id", h.GetReservationByID)
	v1.POST("/reservations", h.CreateReservation)
	v1.PUT("/reservations/:id", h.UpdateReservation)
	v1.DELETE("/reservations/:id", h.DeleteReservation)

	// Pet-based reservations
	v1.GET("/pets/:id/reservations", h.GetReservationsByPetID)

	// Master Items CRUD
	v1.GET("/master-items", h.GetAllMasterItems)
	v1.GET("/master-items/:id", h.GetMasterItemByID)
	v1.GET("/master-items/category/:category", h.GetMasterItemsByCategory)
	v1.POST("/master-items", h.CreateMasterItem)
	v1.PUT("/master-items/:id", h.UpdateMasterItem)
	v1.DELETE("/master-items/:id", h.DeleteMasterItem)

	// Hospitalizations CRUD
	v1.GET("/hospitalizations", h.GetAllHospitalizations)
	v1.GET("/hospitalizations/status/:status", h.GetHospitalizationsByStatus)
	v1.GET("/hospitalizations/:id", h.GetHospitalizationByID)
	v1.POST("/hospitalizations", h.CreateHospitalization)
	v1.PUT("/hospitalizations/:id", h.UpdateHospitalization)
	v1.DELETE("/hospitalizations/:id", h.DeleteHospitalization)

	// Pet-based hospitalizations
	v1.GET("/pets/:id/hospitalizations", h.GetHospitalizationsByPetID)

	// Accountings CRUD
	v1.GET("/accountings", h.GetAllAccounting)
	v1.GET("/accountings/status/:status", h.GetAccountingByStatus)
	v1.GET("/accountings/:id", h.GetAccountingByID)
	v1.POST("/accountings", h.CreateAccounting)
	v1.PUT("/accountings/:id", h.UpdateAccounting)
	v1.DELETE("/accountings/:id", h.DeleteAccounting)

	// Pet-based accountings
	v1.GET("/pets/:id/accountings", h.GetAccountingByPetID)

	// Examinations CRUD
	v1.GET("/examinations", h.GetAllExaminations)
	v1.GET("/examinations/status/:status", h.GetExaminationsByStatus)
	v1.GET("/examinations/:id", h.GetExaminationByID)
	v1.POST("/examinations", h.CreateExamination)
	v1.PUT("/examinations/:id", h.UpdateExamination)
	v1.DELETE("/examinations/:id", h.DeleteExamination)

	// Pet-based examinations
	v1.GET("/pets/:id/examinations", h.GetExaminationsByPetID)

	// Vaccinations CRUD
	v1.GET("/vaccinations", h.GetAllVaccinations)
	v1.GET("/vaccinations/:id", h.GetVaccinationByID)
	v1.POST("/vaccinations", h.CreateVaccination)
	v1.PUT("/vaccinations/:id", h.UpdateVaccination)
	v1.DELETE("/vaccinations/:id", h.DeleteVaccination)

	// Pet-based vaccinations
	v1.GET("/pets/:id/vaccinations", h.GetVaccinationsByPetID)

	// Trimmings CRUD
	v1.GET("/trimmings", h.GetAllTrimmings)
	v1.GET("/trimmings/status/:status", h.GetTrimmingsByStatus)
	v1.GET("/trimmings/:id", h.GetTrimmingByID)
	v1.POST("/trimmings", h.CreateTrimming)
	v1.PUT("/trimmings/:id", h.UpdateTrimming)
	v1.DELETE("/trimmings/:id", h.DeleteTrimming)

	// Pet-based trimmings
	v1.GET("/pets/:id/trimmings", h.GetTrimmingsByPetID)

	// Clinics CRUD
	v1.GET("/clinics", h.GetAllClinics)
	v1.POST("/clinics", h.CreateClinic)

	// Clinic nested routes (register before /:id)
	v1.GET("/clinics/:id/staffs", h.GetStaffByClinicID)

	// Clinic single resource
	v1.GET("/clinics/:id", h.GetClinicByID)
	v1.PUT("/clinics/:id", h.UpdateClinic)
	v1.DELETE("/clinics/:id", h.DeleteClinic)

	// Staffs CRUD
	v1.GET("/staffs", h.GetAllStaff)
	v1.GET("/staffs/:id", h.GetStaffByID)
	v1.POST("/staffs", h.CreateStaff)
	v1.PUT("/staffs/:id", h.UpdateStaff)
	v1.DELETE("/staffs/:id", h.DeleteStaff)

	// Inventory Items CRUD
	v1.GET("/inventory-items", h.GetAllInventoryItems)
	v1.GET("/inventory-items/:id", h.GetInventoryItemByID)
	v1.GET("/inventory-items/category/:category", h.GetInventoryItemsByCategory)
	v1.GET("/inventory-items/status/:status", h.GetInventoryItemsByStatus)
	v1.POST("/inventory-items", h.CreateInventoryItem)
	v1.PUT("/inventory-items/:id", h.UpdateInventoryItem)
	v1.DELETE("/inventory-items/:id", h.DeleteInventoryItem)
}

// Health godoc
// @Summary ヘルスチェック
// @Description APIの稼働状態を確認します
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"message":   "Animal Ekarte API is running",
	}

	// DB接続確認
	if db, err := h.svc.GetDB(); err == nil {
		if sqlDB, err := db.DB().DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				status["database"] = "connected"
			} else {
				status["database"] = "disconnected"
				status["database_error"] = err.Error()
			}
		} else {
			status["database"] = "error"
			status["database_error"] = err.Error()
		}
	} else {
		status["database"] = "error"
		status["database_error"] = err.Error()
	}

	c.JSON(http.StatusOK, status)
}

// Welcome godoc
// @Summary ウェルカムメッセージ
// @Description APIのウェルカムメッセージを返します
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func (h *Handler) Welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Animal Ekarte API",
	})
}
