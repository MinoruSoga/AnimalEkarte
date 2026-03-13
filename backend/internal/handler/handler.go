package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	cfg *config.Config
	svc *service.Services
}

// New はHandlerを初期化して返す
func New(cfg *config.Config, svc *service.Services) *Handler {
	return &Handler{cfg: cfg, svc: svc}
}

// PaginatedResponse はページネーション付きレスポンスの共通構造
type PaginatedResponse[T any] struct {
	Data  T     `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// newPaginatedResponse はPaginatedResponseを型推論で生成するヘルパー
func newPaginatedResponse[T any](data T, total int64, page, limit int) PaginatedResponse[T] {
	return PaginatedResponse[T]{Data: data, Total: total, Page: page, Limit: limit}
}

// Health はサーバーの稼働状態を返す
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RegisterRoutes はすべてのルートを登録する
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// ヘルスチェック（認証不要）
	api.GET("/health", h.Health)

	// 認証不要
	api.POST("/login", h.Login)
	api.POST("/logout", h.Logout)

	// 認証必須グループ
	protected := api.Group("")
	protected.Use(middleware.Auth(h.cfg.JWTSecret))
	{
		// 認証ユーザー情報
		protected.GET("/me", h.GetMe)

		// 飼主
		owners := protected.Group("/owners")
		owners.GET("", h.ListOwners)
		owners.POST("", h.CreateOwner)
		owners.GET("/:id", h.GetOwner)
		owners.PATCH("/:id", h.UpdateOwner)
		owners.DELETE("/:id", h.DeleteOwner)

		// ペット
		pets := protected.Group("/pets")
		pets.GET("", h.ListPets)
		pets.POST("", h.CreatePet)
		pets.GET("/:id", h.GetPet)
		pets.PATCH("/:id", h.UpdatePet)
		pets.DELETE("/:id", h.DeletePet)

		// 予約
		reservations := protected.Group("/reservations")
		reservations.GET("", h.ListReservations)
		reservations.POST("", h.CreateReservation)
		reservations.GET("/:id", h.GetReservation)
		reservations.PATCH("/:id", h.UpdateReservation)
		reservations.DELETE("/:id", h.DeleteReservation)

		// カルテ
		records := protected.Group("/medical-records")
		records.GET("", h.ListMedicalRecords)
		records.POST("", h.CreateMedicalRecord)
		records.GET("/:id", h.GetMedicalRecord)
		records.PATCH("/:id", h.UpdateMedicalRecord)
		records.DELETE("/:id", h.DeleteMedicalRecord)

		// 入院
		hospitalizations := protected.Group("/hospitalizations")
		hospitalizations.GET("", h.ListHospitalizations)
		hospitalizations.POST("", h.CreateHospitalization)
		hospitalizations.GET("/:id", h.GetHospitalization)
		hospitalizations.PATCH("/:id", h.UpdateHospitalization)
		hospitalizations.DELETE("/:id", h.DeleteHospitalization)

		// 会計
		accountings := protected.Group("/accountings")
		accountings.GET("", h.ListAccountings)
		accountings.POST("", h.CreateAccounting)
		accountings.GET("/:id", h.GetAccounting)
		accountings.PATCH("/:id", h.UpdateAccounting)
		accountings.DELETE("/:id", h.DeleteAccounting)

		// トリミング
		trimmings := protected.Group("/trimmings")
		trimmings.GET("", h.ListTrimmings)
		trimmings.POST("", h.CreateTrimming)
		trimmings.GET("/:id", h.GetTrimming)
		trimmings.PATCH("/:id", h.UpdateTrimming)
		trimmings.DELETE("/:id", h.DeleteTrimming)

		// 検査
		examinations := protected.Group("/examinations")
		examinations.GET("", h.ListExaminations)
		examinations.POST("", h.CreateExamination)
		examinations.GET("/:id", h.GetExamination)
		examinations.PATCH("/:id", h.UpdateExamination)
		examinations.DELETE("/:id", h.DeleteExamination)

		// 予防接種
		vaccinations := protected.Group("/vaccinations")
		vaccinations.GET("", h.ListVaccinations)
		vaccinations.POST("", h.CreateVaccination)
		vaccinations.GET("/:id", h.GetVaccination)
		vaccinations.PATCH("/:id", h.UpdateVaccination)
		vaccinations.DELETE("/:id", h.DeleteVaccination)

		// 在庫
		inventory := protected.Group("/inventory")
		inventory.GET("", h.ListInventory)
		inventory.POST("", h.CreateInventory)
		inventory.GET("/:id", h.GetInventory)
		inventory.PATCH("/:id", h.UpdateInventory)
		inventory.DELETE("/:id", h.DeleteInventory)

		// マスタ
		masters := protected.Group("/masters")
		masters.GET("/animal-species", h.ListAnimalSpecies)

		masters.GET("/staffs", h.ListStaffs)
		masters.POST("/staffs", h.CreateStaff)
		masters.PATCH("/staffs/:id", h.UpdateStaff)
		masters.DELETE("/staffs/:id", h.DeleteStaff)

		masters.GET("/cages", h.ListCages)
		masters.POST("/cages", h.CreateCage)
		masters.PATCH("/cages/:id", h.UpdateCage)
		masters.DELETE("/cages/:id", h.DeleteCage)

		masters.GET("/medicines", h.ListMedicines)
		masters.POST("/medicines", h.CreateMedicine)
		masters.PATCH("/medicines/:id", h.UpdateMedicine)
		masters.DELETE("/medicines/:id", h.DeleteMedicine)

		masters.GET("/vaccines", h.ListVaccines)
		masters.POST("/vaccines", h.CreateVaccine)
		masters.PATCH("/vaccines/:id", h.UpdateVaccine)
		masters.DELETE("/vaccines/:id", h.DeleteVaccine)

		masters.GET("/insurances", h.ListInsurances)
		masters.POST("/insurances", h.CreateInsurance)
		masters.PATCH("/insurances/:id", h.UpdateInsurance)
		masters.DELETE("/insurances/:id", h.DeleteInsurance)

		masters.GET("/service-types", h.ListServiceTypes)
		masters.POST("/service-types", h.CreateServiceType)
		masters.PATCH("/service-types/:id", h.UpdateServiceType)
		masters.DELETE("/service-types/:id", h.DeleteServiceType)

		masters.GET("/consultations", h.ListConsultations)
		masters.POST("/consultations", h.CreateConsultation)
		masters.PATCH("/consultations/:id", h.UpdateConsultation)
		masters.DELETE("/consultations/:id", h.DeleteConsultation)

		masters.GET("/procedures", h.ListProcedures)
		masters.POST("/procedures", h.CreateProcedure)
		masters.PATCH("/procedures/:id", h.UpdateProcedure)
		masters.DELETE("/procedures/:id", h.DeleteProcedure)

		masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
		masters.POST("/hospitalization-plans", h.CreateHospitalizationPlan)
		masters.PATCH("/hospitalization-plans/:id", h.UpdateHospitalizationPlan)
		masters.DELETE("/hospitalization-plans/:id", h.DeleteHospitalizationPlan)

		masters.GET("/trimming-courses", h.ListTrimmingCourses)
		masters.POST("/trimming-courses", h.CreateTrimmingCourse)
		masters.PATCH("/trimming-courses/:id", h.UpdateTrimmingCourse)
		masters.DELETE("/trimming-courses/:id", h.DeleteTrimmingCourse)

		masters.GET("/trimming-options", h.ListTrimmingOptions)
		masters.POST("/trimming-options", h.CreateTrimmingOption)
		masters.PATCH("/trimming-options/:id", h.UpdateTrimmingOption)
		masters.DELETE("/trimming-options/:id", h.DeleteTrimmingOption)

		masters.GET("/examination-types", h.ListExaminationTypes)
		masters.POST("/examination-types", h.CreateExaminationType)
		masters.PATCH("/examination-types/:id", h.UpdateExaminationType)
		masters.DELETE("/examination-types/:id", h.DeleteExaminationType)

		masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
		masters.POST("/diagnosis-categories", h.CreateDiagnosisCategory)
		masters.PATCH("/diagnosis-categories/:id", h.UpdateDiagnosisCategory)
		masters.DELETE("/diagnosis-categories/:id", h.DeleteDiagnosisCategory)

		masters.GET("/diagnosis-names", h.ListDiagnosisNames)
		masters.POST("/diagnosis-names", h.CreateDiagnosisName)
		masters.PATCH("/diagnosis-names/:id", h.UpdateDiagnosisName)
		masters.DELETE("/diagnosis-names/:id", h.DeleteDiagnosisName)

		masters.GET("/checkup-types", h.ListCheckupTypes)
		masters.POST("/checkup-types", h.CreateCheckupType)
		masters.PATCH("/checkup-types/:id", h.UpdateCheckupType)
		masters.DELETE("/checkup-types/:id", h.DeleteCheckupType)

		// 法人・クリニック設定
		protected.GET("/company", h.GetCompany)
		protected.PATCH("/company", h.UpdateCompany)
		protected.GET("/clinics", h.ListClinics)
		protected.POST("/clinics", h.CreateClinic)
		protected.GET("/clinics/:id", h.GetClinic)
		protected.PATCH("/clinics/:id", h.UpdateClinic)
		protected.DELETE("/clinics/:id", h.DeleteClinic)
	}
}
