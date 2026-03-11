package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	svc *service.Services
}

// New はHandlerを初期化して返す
func New(svc *service.Services) *Handler {
	return &Handler{svc: svc}
}

// PaginatedResponse はページネーション付きレスポンスの共通構造
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// RegisterRoutes はすべてのルートを登録する
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// 飼主
		owners := api.Group("/owners")
		owners.GET("", h.ListOwners)
		owners.POST("", h.CreateOwner)
		owners.GET("/:id", h.GetOwner)
		owners.PUT("/:id", h.UpdateOwner)
		owners.DELETE("/:id", h.DeleteOwner)

		// ペット
		pets := api.Group("/pets")
		pets.GET("", h.ListPets)
		pets.POST("", h.CreatePet)
		pets.GET("/:id", h.GetPet)
		pets.PUT("/:id", h.UpdatePet)
		pets.DELETE("/:id", h.DeletePet)

		// 予約
		reservations := api.Group("/reservations")
		reservations.GET("", h.ListReservations)
		reservations.POST("", h.CreateReservation)
		reservations.GET("/:id", h.GetReservation)
		reservations.PUT("/:id", h.UpdateReservation)
		reservations.DELETE("/:id", h.DeleteReservation)

		// カルテ
		records := api.Group("/medical-records")
		records.GET("", h.ListMedicalRecords)
		records.POST("", h.CreateMedicalRecord)
		records.GET("/:id", h.GetMedicalRecord)
		records.PUT("/:id", h.UpdateMedicalRecord)
		records.DELETE("/:id", h.DeleteMedicalRecord)

		// 入院
		hospitalizations := api.Group("/hospitalizations")
		hospitalizations.GET("", h.ListHospitalizations)
		hospitalizations.POST("", h.CreateHospitalization)
		hospitalizations.GET("/:id", h.GetHospitalization)
		hospitalizations.PUT("/:id", h.UpdateHospitalization)
		hospitalizations.DELETE("/:id", h.DeleteHospitalization)

		// 会計
		accountings := api.Group("/accountings")
		accountings.GET("", h.ListAccountings)
		accountings.POST("", h.CreateAccounting)
		accountings.GET("/:id", h.GetAccounting)
		accountings.PUT("/:id", h.UpdateAccounting)

		// トリミング
		trimmings := api.Group("/trimmings")
		trimmings.GET("", h.ListTrimmings)
		trimmings.POST("", h.CreateTrimming)
		trimmings.GET("/:id", h.GetTrimming)
		trimmings.PUT("/:id", h.UpdateTrimming)

		// 在庫
		inventory := api.Group("/inventory")
		inventory.GET("", h.ListInventory)
		inventory.POST("", h.CreateInventory)
		inventory.GET("/:id", h.GetInventory)
		inventory.PUT("/:id", h.UpdateInventory)

		// マスタ
		masters := api.Group("/masters")
		masters.GET("/staffs", h.ListStaffs)
		masters.POST("/staffs", h.CreateStaff)
		masters.PUT("/staffs/:id", h.UpdateStaff)
		masters.DELETE("/staffs/:id", h.DeleteStaff)

		masters.GET("/cages", h.ListCages)
		masters.POST("/cages", h.CreateCage)
		masters.PUT("/cages/:id", h.UpdateCage)
		masters.DELETE("/cages/:id", h.DeleteCage)

		masters.GET("/medicines", h.ListMedicines)
		masters.POST("/medicines", h.CreateMedicine)
		masters.PUT("/medicines/:id", h.UpdateMedicine)
		masters.DELETE("/medicines/:id", h.DeleteMedicine)

		masters.GET("/vaccines", h.ListVaccines)
		masters.POST("/vaccines", h.CreateVaccine)
		masters.PUT("/vaccines/:id", h.UpdateVaccine)
		masters.DELETE("/vaccines/:id", h.DeleteVaccine)

		masters.GET("/insurances", h.ListInsurances)
		masters.POST("/insurances", h.CreateInsurance)
		masters.PUT("/insurances/:id", h.UpdateInsurance)
		masters.DELETE("/insurances/:id", h.DeleteInsurance)

		masters.GET("/service-types", h.ListServiceTypes)
		masters.POST("/service-types", h.CreateServiceType)
		masters.PUT("/service-types/:id", h.UpdateServiceType)
		masters.DELETE("/service-types/:id", h.DeleteServiceType)

		masters.GET("/consultations", h.ListConsultations)
		masters.POST("/consultations", h.CreateConsultation)
		masters.PUT("/consultations/:id", h.UpdateConsultation)
		masters.DELETE("/consultations/:id", h.DeleteConsultation)

		masters.GET("/procedures", h.ListProcedures)
		masters.POST("/procedures", h.CreateProcedure)
		masters.PUT("/procedures/:id", h.UpdateProcedure)
		masters.DELETE("/procedures/:id", h.DeleteProcedure)

		masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
		masters.POST("/hospitalization-plans", h.CreateHospitalizationPlan)
		masters.PUT("/hospitalization-plans/:id", h.UpdateHospitalizationPlan)
		masters.DELETE("/hospitalization-plans/:id", h.DeleteHospitalizationPlan)

		masters.GET("/trimming-courses", h.ListTrimmingCourses)
		masters.POST("/trimming-courses", h.CreateTrimmingCourse)
		masters.PUT("/trimming-courses/:id", h.UpdateTrimmingCourse)
		masters.DELETE("/trimming-courses/:id", h.DeleteTrimmingCourse)

		masters.GET("/trimming-options", h.ListTrimmingOptions)
		masters.POST("/trimming-options", h.CreateTrimmingOption)
		masters.PUT("/trimming-options/:id", h.UpdateTrimmingOption)
		masters.DELETE("/trimming-options/:id", h.DeleteTrimmingOption)

		masters.GET("/examination-types", h.ListExaminationTypes)
		masters.POST("/examination-types", h.CreateExaminationType)
		masters.PUT("/examination-types/:id", h.UpdateExaminationType)
		masters.DELETE("/examination-types/:id", h.DeleteExaminationType)

		masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
		masters.POST("/diagnosis-categories", h.CreateDiagnosisCategory)
		masters.PUT("/diagnosis-categories/:id", h.UpdateDiagnosisCategory)
		masters.DELETE("/diagnosis-categories/:id", h.DeleteDiagnosisCategory)

		masters.GET("/diagnosis-names", h.ListDiagnosisNames)
		masters.POST("/diagnosis-names", h.CreateDiagnosisName)
		masters.PUT("/diagnosis-names/:id", h.UpdateDiagnosisName)
		masters.DELETE("/diagnosis-names/:id", h.DeleteDiagnosisName)

		masters.GET("/checkup-types", h.ListCheckupTypes)
		masters.POST("/checkup-types", h.CreateCheckupType)
		masters.PUT("/checkup-types/:id", h.UpdateCheckupType)
		masters.DELETE("/checkup-types/:id", h.DeleteCheckupType)

		// クリニック設定
		api.GET("/clinic", h.GetClinicInfo)
		api.PUT("/clinic", h.UpdateClinicInfo)
		api.GET("/clinics", h.ListClinics)
	}
}
