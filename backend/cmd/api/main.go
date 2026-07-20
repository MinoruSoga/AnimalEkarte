package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/handler"
	"github.com/animal-ekarte/backend/internal/infra"
	appCrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// manualArticleAuditAdapter adapts the shared audit kernel (internal/service.AuditService —
// BE9-2A classification: "keep") to internal/manualarticle's own minimal AuditLogger
// interface, so internal/manualarticle itself never imports internal/service (ADR-006
// "aggregator 非経由"). This is exactly the kind of narrow, composition-root-only bridge
// BE9-2B item 3 anticipates ("DI を main.go だけに限定せず、型安全な composition を維持する").
type manualArticleAuditAdapter struct {
	audit service.AuditService
}

func (a manualArticleAuditAdapter) LogEntry(ctx context.Context, entry manualarticle.AuditEntry) error {
	return a.audit.LogEntry(ctx, &service.AuditLogInput{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}

// medicalRecordAuditTxAdapter adapts internal/service's tx-internal audit logger
// (AuditTxLogger.LogEntryTx over *service.AuditLogInput) to internal/medicalrecord's own
// consumer-side AuditTxLogger view (LogEntryTx over *medicalrecord.AuditEntry), so
// internal/medicalrecord never imports internal/service (ADR-006 aggregator 非経由). Used by
// medicalrecord's checkupFieldResultService for its #211 fail-closed tx-internal deletion
// audit. Moved here from internal/service/medicalrecord_middle_state.go in BE9-2D Batch C
// (mirrors manualArticleAuditAdapter above).
type medicalRecordAuditTxAdapter struct{ inner service.AuditTxLogger }

func (a medicalRecordAuditTxAdapter) LogEntryTx(ctx context.Context, e *medicalrecord.AuditEntry) error {
	return a.inner.LogEntryTx(ctx, &service.AuditLogInput{
		ClinicID:   e.ClinicID,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		OldValue:   e.OldValue,
		NewValue:   e.NewValue,
		Metadata:   e.Metadata,
	})
}

// labAuditAdapter adapts the shared audit kernel's non-tx logger (AuditService.LogEntry over
// *service.AuditLogInput) to internal/medicalrecord's consumer-side AuditLogger view (LogEntry
// over *medicalrecord.AuditEntry), so internal/medicalrecord never imports internal/service
// (ADR-006 aggregator 非経由). Used by medicalrecord's labAuditLogger for its best-effort
// (non-tx) lab-import audit trail. Moved here from internal/service/lab_middle_state.go in
// BE9-2D sub-batch③ Batch C (mirrors medicalRecordAuditTxAdapter / manualArticleAuditAdapter).
type labAuditAdapter struct{ audit service.AuditService }

func (a labAuditAdapter) LogEntry(ctx context.Context, e *medicalrecord.AuditEntry) error {
	return a.audit.LogEntry(ctx, &service.AuditLogInput{
		ClinicID:   e.ClinicID,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		OldValue:   e.OldValue,
		NewValue:   e.NewValue,
		Metadata:   e.Metadata,
	})
}

func main() {
	// 設定読み込み（ロガー初期化より先に行い、LOG_LEVEL を含む全設定を config.Config に一元化する）
	cfg := config.Load()

	// ロガー初期化
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	logger.Init(logger.Config{
		Level:  logLevel,
		Format: "json",
		Output: os.Stdout,
	})
	logger.Info("starting Animal Ekarte API v2.0 (45 tables)")

	if err := config.ConfigureTimeZone(); err != nil {
		slog.Error("timezone configuration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// H2: TRUSTED_PROXY_CIDR（release必須）・STORAGE_TYPE=s3 時の S3_BUCKET/S3_REGION 必須検証は
	// cfg.Validate() に集約済み（G9-2）
	if err := cfg.Validate(); err != nil {
		slog.Error("config validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	gin.SetMode(cfg.GinMode)

	// DB接続
	db, err := repository.NewDB(cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected")

	// リポジトリ初期化
	repos := repository.NewRepositories(db)

	// 連携設定の暗号化 cipher 初期化（INTEGRATION_ENCRYPTION_KEY 未設定時は dev モード・暗号化なし）。
	// lstep 連携（clinic_integrations）と LINE 予約設定（line_reservation_settings）で共有する（H-4）。
	// NewServices に渡すため NewServices 呼び出しより前に初期化する。
	var integrationCipher *appCrypto.AESGCMCipher
	if cfg.IntegrationEncryptionKey != "" {
		c, err := appCrypto.NewAESGCMCipher(cfg.IntegrationEncryptionKey)
		if err != nil {
			logger.Error("failed to initialize AES-GCM cipher", slog.String("error", err.Error()))
			os.Exit(1)
		}
		integrationCipher = c
		logger.Info("integration cipher: AES-256-GCM enabled")
	} else {
		logger.Info("INTEGRATION_ENCRYPTION_KEY not set: running without encryption (dev mode)")
	}

	// 共有ファイルストレージ初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	// NewServices に渡すため NewServices 呼び出しより前に初期化する（G9-1）。
	var sharedStorage infra.FileStorage
	if cfg.StorageType == "s3" {
		s3fs, err := infra.NewS3FileStorage(context.Background(), cfg.S3SharedBucket, cfg.S3SharedRegion, cfg.S3Endpoint)
		if err != nil {
			logger.Error("failed to initialize S3FileStorage", slog.String("error", err.Error()))
			os.Exit(1)
		}
		sharedStorage = s3fs
		logger.Info("shared file storage: S3", slog.String("bucket", cfg.S3SharedBucket))
	} else {
		sharedStorage = infra.NewLocalFileStorage("/app/uploads/shared", "http://localhost:"+cfg.Port+"/uploads/shared")
		logger.Info("shared file storage: local filesystem")
	}

	// サービス初期化（G9-1: 旧・二段階DIを単一段階に統合。LstepLifecycle/LstepTag/SharedFile/
	// ChronicCondition/LineSend/LineLink/LstepTagSummary/CheckupSync/LstepDeliveryMonitor/
	// LstepTriggerPriority/LstepDeliveryTrigger/MedicalRecord/Checkup/LstepBatch/LstepCsvImport/
	// LstepAnalytics はすべて service.NewServices 内で一括構築される）
	svcs := service.NewServices(repos, &service.ReservationNotificationConfig{
		SMTPHost:    cfg.SMTPHost,
		SMTPPort:    cfg.SMTPPort,
		SMTPUser:    cfg.SMTPUser,
		SMTPPass:    cfg.SMTPPass,
		SMTPFrom:    cfg.SMTPFrom,
		FrontendURL: cfg.FrontendURL,
	}, integrationCipher, sharedStorage, cfg.JWTSecret)

	// ファイルアップローダー初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	var uploader infra.FileUploader
	if cfg.StorageType == "s3" {
		// S3_BUCKET/S3_REGION 必須検証は cfg.Validate() で起動時に済んでいる（G9-2）
		s3Up, err := infra.NewS3Uploader(context.Background(), cfg.S3Bucket, cfg.S3Region, cfg.S3Endpoint, cfg.S3PublicBaseURL)
		if err != nil {
			logger.Error("failed to initialize S3 uploader", slog.String("error", err.Error()))
			os.Exit(1)
		}
		uploader = s3Up
		logger.Info("file uploader: S3", slog.String("bucket", cfg.S3Bucket), slog.String("region", cfg.S3Region))
		// S3 API endpoint（R2 等）を使う構成で公開 base URL が未設定の場合、オブジェクト
		// 公開 URL は API ホストを指しブラウザから参照できない。推測ドメインは捏造せず、
		// 誤設定として起動時に警告する（P2-5: S3_PUBLIC_BASE_URL への実値投入は USER 運用）。
		if cfg.S3Endpoint != "" && cfg.S3PublicBaseURL == "" {
			logger.Warn("S3_PUBLIC_BASE_URL 未設定: オブジェクト公開 URL が S3 API ホストを指しブラウザから参照できません。R2 の公開ドメイン(custom domain / *.r2.dev)を S3_PUBLIC_BASE_URL に設定してください",
				slog.String("s3_endpoint", cfg.S3Endpoint))
		}
	} else {
		uploader = infra.NewLocalUploader("/app/uploads", "/uploads")
		logger.Info("file uploader: local filesystem")
	}

	// ハンドラー初期化
	h := handler.New(cfg, svcs, repos, uploader)

	// アプリケーションライフタイムコンテキスト（バックグラウンドゴルーチン管理用）
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// LSTEP-BE-014: ノーショウ検知バッチ — 毎時0分に起動し 10/15/20 時 (JST) のみ実行
	// time.Local は main() 冒頭の config.ConfigureTimeZone() で Asia/Tokyo 確定済み（G9-3）
	go runScheduled(appCtx, "no-show batch", hourlyTick, func(ctx context.Context) error {
		triggerHours := map[int]bool{10: true, 15: true, 20: true}
		if !triggerHours[time.Now().Hour()] {
			return nil
		}
		return svcs.LstepBatch.RunNoShowCheckAllClinics(ctx)
	})

	// LSTEP-BE-004: 休眠検知バッチ — 毎日 02:00 JST に実行
	go runScheduled(appCtx, "dormant detection batch", dailyAt2AM, svcs.LstepBatch.RunDormantDetectionAllClinics)

	// FEAT-383: 自動配信トリガーバッチ — 毎時0分に起動（10:00 JST 固定）
	go runScheduled(appCtx, "delivery trigger batch", hourlyTick, svcs.LstepBatch.RunDeliveryTriggerBatchAllClinics)

	// ルーター設定
	r := gin.New()
	// H2: Set trusted proxies to prevent rate-limit bypass via X-Forwarded-For spoofing
	// TRUSTED_PROXY_CIDR validation is done earlier via cfg.Validate(), so only build list here
	var trustedProxies []string
	if cfg.GinMode == "release" {
		// Production: trust ALB CIDR
		// e.g., TRUSTED_PROXY_CIDR="10.0.0.0/8" for AWS ALB in private subnet
		if cfg.TrustedProxyCIDR != "" {
			trustedProxies = []string{cfg.TrustedProxyCIDR}
			slog.Info("rate-limit: trusting ALB CIDR")
		}
	} else {
		// Development: trust localhost only
		trustedProxies = []string{"127.0.0.1"}
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		// Log but continue - SetTrustedProxies failure is non-fatal
		slog.Warn("failed to set trusted proxies", slog.Any("error", err))
	}
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigin))
	r.Use(middleware.RequestLoggingMiddleware())
	// BUG-067: POST/PATCH/PUT ボディから NULL バイトを除去（PostgreSQL エラー防止）
	r.Use(middleware.SanitizeNullBytes())
	protected := h.RegisterRoutes(appCtx, r)

	// BE9-2B pilot: internal/manualarticle is composed here directly — it does not go
	// through handler.Handler/service.Services/repository.Repositories. It reuses the same
	// protected *gin.RouterGroup (so it inherits middleware.Auth's clinic_id/user_id/
	// is_system_admin context) and the transitional h.RequirePermission method value (auth
	// domain has not migrated yet; see internal/manualarticle/handler.go's PermissionMiddleware
	// doc comment).
	manualArticleHandler := manualarticle.NewHandler(
		manualarticle.NewManualArticleService(manualarticle.New(db)),
		manualArticleAuditAdapter{audit: svcs.Audit},
		h.RequirePermission,
	)
	manualArticleHandler.RegisterRoutes(protected)

	// BE9-2C/2D (medicalrecord slice): same aggregator-non-経由 pattern as the BE9-2B
	// manualarticle pilot above. The master-CRUD entities (diagnosis/exam/chief-complaint,
	// BE9-2C) plus checkup/checkup-field-result/checkup-type/vaccine/vaccination/inquiry/
	// inquiry-template/prescription (BE9-2D) plus the lab import/report saga (BE9-2D
	// sub-batch③, wired just below the checkup services) are composed here directly from repos
	// rather than via svcs.* (their Services fields were removed — nothing else read them).
	// Unlike the 2C master handlers (which never wrote audit entries), the 2D services need the
	// same LSTEP tag-sync / delivery-trigger deps and tx/audit boundary the pre-move NewServices
	// wired: the LSTEP deps come from svcs.LstepTagSync / svcs.LstepDeliveryTrigger (still
	// constructed inside NewServices), tx from repository.NewTransactor(repos.DB()), and the
	// tx-internal audit logger via medicalRecordAuditTxAdapter (lab's best-effort non-tx audit
	// uses labAuditAdapter). checkupSvc is kept in scope for its graceful-shutdown drain
	// (checkupSvc.Wait() below, replacing the old svcs.Checkup.Wait()).
	mrTx := repository.NewTransactor(repos.DB())
	mrAuditTxLogger, ok := svcs.Audit.(service.AuditTxLogger)
	if !ok {
		logger.Error("DI wiring error: AuditService concrete does not implement AuditTxLogger (#211 tx-internal audit)")
		os.Exit(1)
	}
	checkupSvc := medicalrecord.NewCheckupService(repos.Checkup, repos.MedicalRecord, repos.CheckupType, svcs.LstepDeliveryTrigger, svcs.LstepTagSync)
	checkupFieldResultSvc := medicalrecord.NewCheckupFieldResultService(repos.Checkup, repos.MedicalRecord, repos.CheckupTypeField, repos.CheckupFieldResult, medicalRecordAuditTxAdapter{inner: mrAuditTxLogger}, mrTx)
	checkupTypeSvc := medicalrecord.NewCheckupTypeService(repos.CheckupType)
	vaccineSvc := medicalrecord.NewVaccineService(repos.Vaccine)
	vaccinationSvc := medicalrecord.NewVaccinationService(repos.Vaccination, repos.Vaccine, svcs.LstepTagSync)
	prescriptionSvc := medicalrecord.NewPrescriptionService(repos.Prescription, repos.MedicalRecord, svcs.LstepTagSync, mrTx)
	inquirySvc := medicalrecord.NewInquiryService(repos.Inquiry, repos.ChiefComplaintType)
	inquiryTemplateSvc := medicalrecord.NewInquiryTemplateService(repos.InquiryTemplate)

	// lab import/report saga (BE9-2D sub-batch③): moved from internal/service NewServices to
	// here (leaf domain, no facade). labImportJobSvc is a single instance shared by the
	// LabImportJob reads (get/events) and the LabResultImport commit path — preserving the
	// pre-move sharing. examinationImportRepo/examinationReportRepo=repos.Examination,
	// petFinder=repos.Pet, medicalRecordFinder=repos.MedicalRecord, ExamTypeRepository=
	// repos.ExaminationType, dupChecker=medicalrecord.NewLabImportDuplicateCheckerDB (the
	// repository facade was deleted in Batch B), and the non-tx best-effort audit trail flows
	// through labAuditAdapter (mirrors the checkup tx-audit adapter above).
	labImportJobSvc := medicalrecord.NewLabImportJobService(
		medicalrecord.NewLabImportJobRepository(repos.DB()),
		medicalrecord.NewLabImportEventRepository(repos.DB()),
	)
	labResultImportSvc := medicalrecord.NewLabResultImportService(
		labImportJobSvc,
		medicalrecord.NewLabImportExaminationService(
			repos.Examination,
			medicalrecord.NewLabImportDuplicateCheckerDB(repos.DB()),
			repos.ExaminationType,
			repos.Pet,
			repos.MedicalRecord,
		),
	)
	labAuditLogger := medicalrecord.NewLabAuditLogger(labAuditAdapter{audit: svcs.Audit})
	labReportQuerySvc := medicalrecord.NewLabReportQueryService(repos.Examination)

	// vital / clinical-plan / medical-record-image (BE9-2D sub-batch④a): moved from
	// internal/service NewServices to here (their Services fields were removed). Same wiring the
	// pre-move NewServices used: repos.* (Batch A facade aliases), mrTx, and svcs.Audit as the
	// vital audit sink — svcs.Audit satisfies medicalrecord's vitalAuditLogger view directly (no
	// adapter; signature is LogVitalChange field-for-field). The vital / image handlers take
	// svcs.MedicalRecord as their medicalRecordGetter (the faithful port of the pre-move
	// verifyMedicalRecordOwnership → svc.MedicalRecord.GetByID), and the image handler takes the
	// same infra.FileUploader (uploader) that internal/handler.New injected.
	vitalSvc := medicalrecord.NewVitalService(repos.Vital, repos.MedicalRecord, svcs.Audit, mrTx)
	clinicalPlanSvc := medicalrecord.NewClinicalPlanService(repos.ClinicalPlan, repos.MedicalRecord, repos.DiagnosisType, repos.DiagnosisName)
	medicalRecordImageSvc := medicalrecord.NewMedicalRecordImageService(repos.MedicalRecordImage, repos.MedicalRecord, mrTx)

	medicalRecordHandler := medicalrecord.NewHandler(
		medicalrecord.NewDiagnosisHandler(
			medicalrecord.NewDiagnosisTypeService(repos.DiagnosisType),
			medicalrecord.NewDiagnosisNameService(repos.DiagnosisName, repos.DiagnosisType),
		),
		medicalrecord.NewExamTypeHandler(medicalrecord.NewExamTypeService(repos.ExaminationType)),
		medicalrecord.NewChiefComplaintHandler(medicalrecord.NewChiefComplaintTypeService(repos.ChiefComplaintType)),
		medicalrecord.NewCheckupHandler(checkupSvc, checkupFieldResultSvc),
		medicalrecord.NewCheckupTypeHandler(checkupTypeSvc),
		medicalrecord.NewVaccineHandler(vaccineSvc),
		medicalrecord.NewVaccinationHandler(vaccinationSvc),
		medicalrecord.NewPrescriptionHandler(prescriptionSvc),
		medicalrecord.NewInquiryHandler(inquirySvc),
		medicalrecord.NewInquiryTemplateHandler(inquiryTemplateSvc),
		medicalrecord.NewLabImportHandler(labResultImportSvc, labImportJobSvc, labAuditLogger),
		medicalrecord.NewLabReportHandler(labReportQuerySvc),
		medicalrecord.NewVitalHandler(vitalSvc, svcs.MedicalRecord),
		medicalrecord.NewClinicalPlanHandler(clinicalPlanSvc),
		medicalrecord.NewMedicalRecordImageHandler(medicalRecordImageSvc, svcs.MedicalRecord, uploader),
		h.RequirePermission,
	)
	medicalRecordHandler.RegisterRoutes(protected)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting",
		slog.String("port", cfg.Port),
	)

	// Graceful shutdown
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("server goroutine panic", slog.Any("panic", r))
			}
		}()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// シャットダウン処理（30秒タイムアウト）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}

	// PERF-FOLLOWUP-05: パスワードリセットメール送信は fire-and-forget goroutine のため、
	// server.Shutdown の HTTP drain だけでは goroutine が孤児化する。ここで明示的に drain する。
	logger.Info("draining in-flight password reset email goroutines...")
	svcs.PasswordReset.Wait()

	// BE-refactor.md B-1: 予約通知（LINE/メール）と健診フォローアップトリガーも
	// fire-and-forget goroutine のため、同様に明示的に drain する。
	logger.Info("draining in-flight reservation notification goroutines...")
	svcs.ReservationNotifier.Wait()
	logger.Info("draining in-flight checkup followup trigger goroutines...")
	checkupSvc.Wait()

	logger.Info("server stopped")
}
