package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	cfg                *config.Config
	svc                *service.Services
	liffCustomerLookup middleware.LineCustomerLookup
	liffSettingLookup  middleware.LineReservationSettingLookup
	uploader           infra.FileUploader
}

// New はHandlerを初期化して返す
// consumer側narrow interface方針: repos は LiffAuth 用の narrow interface 2 件のみを抽出して保持する（全 repository 集約は保持しない）。
func New(cfg *config.Config, svc *service.Services, repos *repository.Repositories, uploader infra.FileUploader) *Handler {
	h := &Handler{cfg: cfg, svc: svc, uploader: uploader}
	if repos != nil {
		h.liffCustomerLookup = repos.LineCustomerMgr
		h.liffSettingLookup = repos.LineReservationSetting
	}
	return h
}

// PaginatedResponse is a DEPRECATED facade (BE9-2C): moved to
// internal/httpapi.PaginatedResponse (BE9-2A target:httpapi, ADR-006 — see
// query_helpers.go's file header for why this extraction happened now). Delete once BE9-2F
// migrates every remaining internal/handler file to internal/httpapi directly.
type PaginatedResponse[T any] = httpapi.PaginatedResponse[T]

// newPaginatedResponse はPaginatedResponseを型推論で生成するヘルパー
func newPaginatedResponse[T any](data T, total int64, page, limit int) PaginatedResponse[T] {
	return httpapi.NewPaginatedResponse(data, total, page, limit)
}

// Health はサーバーの稼働状態を返す
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RegisterRoutes はすべてのルートを登録する。
// ctx はバックグラウンドゴルーチン（rate limiter cleanup 等）のライフタイム管理に使用する。
//
// BE9-2B: protected *gin.RouterGroup を返すようになった。manualarticle のような
// aggregator 非経由の domain package は、この group を cmd/api/main.go 経由で受け取り、
// 自身の RegisterRoutes(protected) を呼ぶ（*handler.Handler を経由しない）。
func (h *Handler) RegisterRoutes(ctx context.Context, r *gin.Engine) *gin.RouterGroup {
	// Health check エンドポイント（ルートレベル）
	r.GET("/health", h.Health)

	// Static file serving for uploaded images（BUG-141: ディレクトリリスティング無効化）
	r.StaticFS("/uploads", gin.Dir("/app/uploads", false))

	api := r.Group("/api/v1")

	// 認証関連（保護なし）— ログインにはレートリミット適用（BUG-130: ブルートフォース対策）
	loginRateStore := middleware.NewRateLimitStore(ctx)
	// パスワードリセット系も独立したストアでレートリミット（3回/分・バースト3）
	passwordRateStore := middleware.NewRateLimitStore(ctx)
	api.POST("/login", middleware.RateLimit(loginRateStore, 5.0/60, 5), h.Login) // 5回/分
	api.POST("/logout", h.Logout)
	api.POST("/auth/refresh", h.RefreshToken) // BUG-136: refresh token エンドポイント
	api.POST("/auth/forgot-password", middleware.RateLimit(passwordRateStore, 3.0/60, 3), h.ForgotPassword)
	api.POST("/auth/reset-password", middleware.RateLimit(passwordRateStore, 3.0/60, 3), h.ResetPassword)

	protected := api.Group("")
	var auditSvc service.AuditService
	var staffSvc service.StaffService
	if h.svc != nil {
		auditSvc = h.svc.Audit
		staffSvc = h.svc.Staff
	}
	protected.Use(middleware.Auth(h.tokenSvc(), h.cfg.GinMode == "release", auditSvc, staffSvc))
	protected.Use(middleware.RequireXRequestedWith())
	// NOTE: SanitizeNullBytes は main.go でグローバル登録済み（BUG-067）

	protected.GET("/me", h.GetMe)
	protected.PUT("/users/me/password", h.ChangeMyPassword) // BUG-148: 自分のパスワード変更

	// BUG-020: 各リソースの write 操作に権限チェックを適用
	h.registerOwnerRoutesWithAuth(protected)
	h.RegisterPetRoutes(protected)
	// 予約 CRUD routes は internal/reservation.RegisterRoutes へ移動（BE9-2C R③）
	h.registerMedicalRecordRoutesWithAuth(protected)
	// hospitalizations 系 route: BE9-2D ⑤⑥で全て internal/medicalrecord の RegisterRoutes へ移動。
	h.registerAccountingRoutesWithAuth(protected)
	h.registerTrimmingRoutesWithAuth(protected)
	// examinations 系 route: BE9-2D ⑦ で internal/medicalrecord の RegisterRoutes へ移動。
	// BE9-2D: /vaccinations moved to internal/medicalrecord.Handler.RegisterRoutes
	// (composed directly in cmd/api/main.go, ADR-006 aggregator 非経由).
	h.registerInventoryRoutesWithAuth(protected)
	h.RegisterMasterRoutes(protected)
	h.RegisterClinicRoutes(protected)
	h.registerEstimateRoutesWithAuth(protected)
	h.RegisterShiftRoutes(protected)
	h.RegisterShiftTemplateRoutes(protected)
	h.RegisterClinicHolidayRoutes(protected)
	h.RegisterCompanyRoutes(protected)
	// BE9-2D: /checkups (+ /checkups/field-results) moved to
	// internal/medicalrecord.Handler.RegisterRoutes (composed in cmd/api/main.go).
	h.RegisterBillingItemRoutes(protected)
	h.RegisterLineReservationRoutes(protected)
	// FEAT-368: 集計・締め機能
	h.RegisterClosingSettingsRoutes(protected)
	h.RegisterCashRegisterRoutes(protected)
	h.RegisterAccountingReportRoutes(protected)
	h.RegisterPaymentMethodMasterRoutes(protected)

	// BE9-2D sub-batch③: /lab-imports and /lab-reports moved to
	// internal/medicalrecord.Handler.RegisterRoutes (composed in cmd/api/main.go, ADR-006
	// aggregator 非経由).

	// LSTEP / LINE連携
	h.RegisterLstepSettingsRoutes(protected)
	h.RegisterSharedFileRoutes(protected)
	h.RegisterAggregationRoutes(protected)
	// LSTEP-BE-020: タグ集計・タグ別飼い主検索
	h.RegisterLstepTagSummaryRoutes(protected)
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	h.RegisterCheckupSyncRoutes(protected)
	// Q23: トリガー優先順位設定
	h.RegisterLstepTriggerPriorityRoutes(protected)
	// FEAT-379: タグコードマッピング設定
	h.RegisterLstepTagCodeMappingRoutes(protected)
	// 自動管理タグプレフィックス・条件タグ・送信目的タグ設定
	h.RegisterLstepTagConfigRoutes(protected)
	// FEAT-384: 自動配信トリガー監視
	h.RegisterLstepDeliveryMonitorRoutes(protected)
	// FEAT-385: Lステップ CSV インポート・分析
	h.RegisterLstepCsvImportRoutes(protected)
	h.RegisterLstepAnalyticsRoutes(protected)

	// LIFF公開API（JWT認証なし・LINE IDトークン認証）
	// LIFF 公開 API routes は internal/reservation.RegisterLiffRoutes へ移動（BE9-2C R⑤・main.go 配線）

	// BE-021: LINE Webhook（JWT認証なし・HMAC-SHA256署名検証）
	r.POST("/api/line/webhook", h.ReceiveLineWebhook)

	return protected
}

// registerOwnerRoutesWithAuth は飼主ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
//
// BE-refactor.md F-2 は本関数を table-driven 化する計画だったが、
// internal/apicontract/openapi_route_drift_test.go の go/ast 静的ウォーカーが
// owners.PATCH("/literal", ...) 等の呼び出しで method/path が両方リテラル文字列である
// ことを前提に docs/api.yaml とのドリフト検出を行っており（動的な値は
// "dynamic path cannot be statically resolved" として unresolved 扱いにし意図的に
// fail loud する設計 — 同ファイルのコメント参照）、struct スライス + ループへの
// table 化はこの機械ゲートを恒久的に無効化する。write 側 clinic-scope review-coverage ゲートの taint 解析断念と
// 同じ理由（静的解析の対象外にすると"検証した気にさせる"だけの偽ガードになる）で
// ゲート側を緩めることは選ばない。よって table-driven 化は行わず、計画の実質的な
// 意図（エイリアス非対称の理由を明文化する）のみをコメントで果たす。
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListOwners)
	owners.GET("/:id", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwner)
	owners.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreateOwner)
	owners.PATCH("/:id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwner)
	owners.DELETE("/:id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwner)
	// BE-005: LINE User ID 連携・解除
	owners.PATCH("/:id/line-user-id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineUserID)
	// ISSUE-001: FE統一エンドポイントのエイリアス。co側に無い理由は未文書化（現状維持）
	owners.PATCH("/:id/line", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineUserID)
	// co側に無い理由は未文書化（現状維持）
	owners.DELETE("/:id/line", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLine)
	// BE-017: Lステップオプトアウト（旧エンドポイント互換保持）
	owners.POST("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLstepOptOut)
	// ISSUE-001: 統合opt-outエンドポイント。co側に無い理由は未文書化（現状維持）
	owners.PATCH("/:id/lstep/opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLstepOptOut)
	// FEAT-381: 配信除外・転院・LINE ID確認
	owners.PATCH("/:id/delivery-exclusion", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryExclusion)
	// FEAT-381-2: 配信注意フラグ
	owners.PATCH("/:id/delivery-caution", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryCaution)
	owners.PATCH("/:id/transfer-status", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerTransferStatus)
	owners.PATCH("/:id/line-id-confirm", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineIDConfirm)
	// BE-019: Lステップタグ CRUD
	owners.GET("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	owners.POST("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.AddOwnerLstepTag)
	owners.DELETE("/:id/lstep/tags/:tag_name", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)
	// BE-013: LINE個別送信。owners側は /lstep/send・/lstep/send-history エイリアスを含む4ルート、
	// co側はそのうち2ルートのみ（下記）— co側に2ルート無い理由は未文書化（現状維持）
	h.RegisterLineSendRoutes(owners)
	// BE-021: LINE User ID 自動取得・飼い主紐付けトークン発行。co側に無い理由は未文書化（現状維持）
	owners.POST("/:id/line/link-token", h.RequirePermission(string(model.ResourceOwners), "edit"), h.GenerateLineLinkToken)

	// ISSUE-001/ISSUE-002: FE が /clinics/:clinic_id/owners/:id/... プレフィックスで呼ぶルートのエイリアス
	// extractClinicID は JWT コンテキストから clinic_id を取得するため :clinic_id URL パラムは無視される
	co := rg.Group("/clinics/:clinic_id/owners")
	co.PATCH("/:id/line-user-id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineUserID)
	co.PATCH("/:id/delivery-exclusion", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryExclusion)
	co.PATCH("/:id/delivery-caution", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryCaution)
	co.PATCH("/:id/transfer-status", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerTransferStatus)
	co.PATCH("/:id/line-id-confirm", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineIDConfirm)
	co.POST("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLstepOptOut)
	co.GET("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	co.POST("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.AddOwnerLstepTag)
	co.DELETE("/:id/lstep/tags/:tag_name", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)
	co.POST("/:id/line/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.SendLineMessage)
	co.GET("/:id/line/send-logs", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
	// FEAT-385: 飼主の最新 Lステップ友だち属性（owners側に対応ルートなし）
	co.GET("/:id/lstep/friend-attributes", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepOwnerFriendAttributes)
}

// registerMedicalRecordRoutesWithAuth はカルテルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	// BE9-2D ⑦: カルテ本体 CRUD/recommendation-reason/addenda は internal/medicalrecord の
	// RegisterRoutes へ移動。billing-confirmation（billing 残留 domain）のみここに残る
	// （/medical-records group は gin path merge で共存）。
	h.RegisterBillingConfirmationRoutes(records)
}

// registerTrimmingRoutesWithAuth はトリミングルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerTrimmingRoutesWithAuth(rg *gin.RouterGroup) {
	trimmings := rg.Group("/trimmings")
	trimmings.GET("", h.RequirePermission(string(model.ResourceTrimming), "view"), h.ListTrimmings)
	trimmings.GET("/:id", h.RequirePermission(string(model.ResourceTrimming), "view"), h.GetTrimming)
	trimmings.POST("", h.RequirePermission(string(model.ResourceTrimming), "create"), h.CreateTrimming)
	trimmings.PATCH("/:id", h.RequirePermission(string(model.ResourceTrimming), "edit"), h.UpdateTrimming)
	trimmings.DELETE("/:id", h.RequirePermission(string(model.ResourceTrimming), "delete"), h.DeleteTrimming)
}

// registerAccountingRoutesWithAuth は会計ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerAccountingRoutesWithAuth(rg *gin.RouterGroup) {
	accountings := rg.Group("/accountings")
	accountings.GET("", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListAccountings)
	// BUG-370: 月末未納者一覧
	accountings.GET("/unpaid", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListUnpaidBillings)
	// #182: 会計画面表示用 飼主未納残高
	accountings.GET("/unpaid-balance", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetOwnerUnpaidBalance)
	// #114: 月次未納繰越集計
	accountings.GET("/unpaid-monthly", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetUnpaidMonthlySummary)
	// BUG-368: レジ締め日次集計
	accountings.GET("/daily-summary", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetDailySummary)
	accountings.GET("/:id", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetAccounting)
	accountings.GET("/:id/refunds", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListRefunds)
	accountings.POST("", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateAccounting)
	accountings.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateAccounting)
	// #189: 確定済みカード金額の確定後訂正（専用フロー）。確定/締め後訂正の既存権限 accounting-post-close-edit:edit を再利用。
	accountings.POST("/:id/credit-correction", h.RequirePermission(string(model.ResourceAccountingPostCloseEdit), "edit"), h.CorrectCreditPayment)
	// BUG-371 / #118: DELETE は廃止し論理削除 (POST /:id/cancel) に統合。専用権限 accounting-cancel を使用する。
	accountings.POST("/:id/cancel", h.RequirePermission(string(model.ResourceAccountingCancel), "edit"), h.CancelAccounting)
	accountings.POST("/:id/refunds", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateRefund)
}

// registerInventoryRoutesWithAuth は在庫ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerInventoryRoutesWithAuth(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.RequirePermission(string(model.ResourceInventory), "view"), h.ListInventory)
	inventory.GET("/:id", h.RequirePermission(string(model.ResourceInventory), "view"), h.GetInventory)
	inventory.POST("", h.RequirePermission(string(model.ResourceInventory), "create"), h.CreateInventory)
	inventory.PATCH("/:id", h.RequirePermission(string(model.ResourceInventory), "edit"), h.UpdateInventory)
	inventory.DELETE("/:id", h.RequirePermission(string(model.ResourceInventory), "delete"), h.DeleteInventory)
}

// registerEstimateRoutesWithAuth は見積書ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerEstimateRoutesWithAuth(rg *gin.RouterGroup) {
	estimates := rg.Group("/estimates")
	estimates.GET("", h.RequirePermission(string(model.ResourceEstimates), "view"), h.ListEstimates)
	estimates.GET("/:id", h.RequirePermission(string(model.ResourceEstimates), "view"), h.GetEstimate)
	estimates.POST("", h.RequirePermission(string(model.ResourceEstimates), "create"), h.CreateEstimate)
	estimates.PATCH("/:id", h.RequirePermission(string(model.ResourceEstimates), "edit"), h.UpdateEstimate)
	estimates.DELETE("/:id", h.RequirePermission(string(model.ResourceEstimates), "delete"), h.DeleteEstimate)
}
