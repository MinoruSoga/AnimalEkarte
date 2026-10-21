package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// BcryptCost は bcrypt ハッシュ生成コストの標準値。
// DefaultCost(10) は 2013 年基準で現代では不十分なため 12 を使用する。
const (
	BcryptCost            = 12
	minimumJWTSecretBytes = 32
	// minimumSchedulerInternalTokenBytes は release で要求する scheduler 内部
	// トークンの最小バイト長。JWT_SECRET と同じ 32 を下限とする。
	// 弱い共有秘密では全院バッチの認可境界が崩れる（CWE-521）ため、
	// 未設定・短い値は起動時に fail-loud とする（middleware は空 expected で
	// 全リクエスト 401 の fail-closed — 起動失敗の方が設定ミスを早期に顕在化する）。
	minimumSchedulerInternalTokenBytes = 32
	maxReleaseDBOpenConns              = 10
	maxReleaseDBIdleConns              = 5
)

// Config holds process environment for the API and related commands.
type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	// DBSSLRootCert selects the trust store used for PostgreSQL server certificate verification.
	DBSSLRootCert string
	GinMode       string

	// DB接続プール上限。既定はローカル/AWS前提の従来値。
	// Cloudflare Containers(直結・プーラー無し)は max_instances 分が並存しうるため、
	// wrangler.jsonc 側で低値を明示注入する(PlanetScaleのスロット枯渇防止 — 2026-07-17障害の再発防止)。
	DBMaxOpenConns int
	DBMaxIdleConns int

	// DBSlowQueryMS は slow-query ログの閾値(ms)。0 は記録しない。
	DBSlowQueryMS int

	// DBStatementTimeoutMS は接続プールが張る各セッションの statement_timeout(ms)。
	// 既定 120000 (120s) — 停滞クエリがプール枠を長時間占有し続けた 2026-10-01
	// 障害の再発防止。getEnvInt の性質上 0/負値は既定値にフォールバックする。
	DBStatementTimeoutMS int
	// DBConnectTimeoutSec は新規接続確立のタイムアウト秒。既定 5。
	DBConnectTimeoutSec int

	JWTSecret string

	// SMTP設定（空文字=無効）。LINE アクセストークン・通知先メールはクリニックごとに DB で管理する。
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// FrontendURL はパスワードリセットメール等に含めるフロントエンドのベースURL。
	FrontendURL string

	// IntegrationEncryptionKey はクリニック連携設定（LステップAPIキー等）をAES-256-GCMで暗号化するキー。
	// 32バイトのhex文字列（64文字）。本番では必須。
	IntegrationEncryptionKey string

	// S3 ファイルストレージ設定（shared_files用）
	S3SharedBucket string
	S3SharedRegion string

	// S3Endpoint はカスタム S3 互換エンドポイント（Cloudflare R2 等）。
	// 空文字（既定）の場合は AWS S3 のリージョナルエンドポイント・バーチャルホスト形式を維持する。
	S3Endpoint string

	// StorageType はファイルアップロード先を切り替える ("s3" または空文字=ローカル)。
	StorageType string
	// S3Bucket/S3Region はアップローダー用 S3 バケット（StorageType=s3 のとき必須）。
	// shared_files 用の S3SharedBucket/S3SharedRegion とは別設定。
	S3Bucket string
	S3Region string
	// AWSAccessKeyID/AWSSecretAccessKey are loaded only to fail release startup
	// before the AWS SDK's lazy credential resolution reaches the first upload.
	// Never log these values.
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	// S3PublicBaseURL はブラウザ向けオブジェクト公開 URL の base（R2 custom domain /
	// *.r2.dev / CloudFront 等）。S3 API 接続先の S3Endpoint とは別ホストであり、
	// 設定時はアップロード後 URL の組み立てに優先使用する。空文字（既定）の場合は
	// AWS はバーチャルホスト形式、R2 は API endpoint への暫定フォールバックとなる。
	S3PublicBaseURL string

	// TrustedProxyCIDR は release モードで rate-limit バイパス防止のため必須の ALB CIDR。
	TrustedProxyCIDR string

	// LogLevel はロガー初期化時のレベル切り替えに使用する ("debug" でデバッグレベル)。
	LogLevel string

	// CORSAllowedOrigin は CORS 許可オリジン（カンマ区切り）。空文字ならミドルウェア側で開発既定値を適用する。
	CORSAllowedOrigin string

	TrimmingImageCleanupEnabled bool

	// SchedulerInternalToken protects POST /_internal/scheduled-jobs (DEC-36 / CMD-02).
	// Callers must send header X-Scheduler-Token. Release mode requires an
	// explicit value of at least minimumSchedulerInternalTokenBytes (Validate).
	// Outside release an empty expected token keeps the route registered but
	// middleware fails closed (every request 401).
	SchedulerInternalToken string

	// SupportLLM* はサポートウィジェットのヘルプチャット用 LLM 設定（OpenAI 互換 API）。
	// SUPPORT_LLM_API_KEY 未設定の場合はチャット機能自体が無効になり、
	// フロントエンドは検索型ヘルプへフォールバックする。
	// xAI / Gemini 等の OpenAI 互換エンドポイントも BaseURL 差し替えで利用できる。
	// キーは絶対にログに出力しない。
	SupportLLMBaseURL   string
	SupportLLMAPIKey    string
	SupportLLMModel     string
	SupportLLMTimeoutMS int

	// EMR-410: ハイブリッド検索・外部送信承認（既定はすべて無効/false）。
	// SupportExternalDataApproved=false の間は Sarashina への質問・文書送信を行わない。
	SupportHybridSearchEnabled  bool
	SupportEmbeddingBaseURL     string
	SupportEmbeddingAPIKey      string
	SupportEmbeddingModel       string
	SupportEmbeddingTimeoutMS   int
	SupportExternalDataApproved bool

	// Plane* はバグ報告の Plane ワークアイテム自動起票設定。
	// PLANE_API_KEY 未設定の場合は連携無効（報告はローカル保存のみ、
	// 管理画面からの手動起票は 501）。BaseURL は Plane REST API、
	// WebBaseURL はチケット表示 URL 組立用の Web アプリ側。
	// ProjectID は API パスに使う UUID、ProjectIdentifier は表示 URL 用の
	// プロジェクト識別子（例 "EMR"。API 応答に identifier が無い場合の fallback）。
	// API キーは絶対にログに出力しない。
	PlaneBaseURL           string
	PlaneWebBaseURL        string
	PlaneWorkspaceSlug     string
	PlaneProjectID         string
	PlaneProjectIdentifier string
	PlaneAPIKey            string
	PlaneTimeoutMS         int

	// PlaneEnvironment は Plane 起票の external_id 名前空間に使う信頼済み環境名。
	// APP_ENV から allowlist (development/staging/production) で導出する。
	// debug で未設定/未知値は development 既定。release で PLANE_API_KEY が有効な
	// 場合、未設定/未知値は Validate が拒否する（fail-closed）。
	// リクエスト値・URL など信頼できない入力からは導出しない。
	PlaneEnvironment string

	// PlaneSourceLabelID / PlaneLaneLabelID はバグ報告起票時に送る Plane label
	// UUID（source:manual / lane:dev）。PLANE_API_KEY 有効時は必須で、
	// 未設定・非 UUID 値は Validate が拒否する（label 無しチケットを作らない）。
	PlaneSourceLabelID string
	PlaneLaneLabelID   string

	// DISCORD_BUG_REPORT_WEBHOOK_URL はバグ報告を DB 保存前に Discord チャンネルへ
	// 転送する Incoming Webhook URL。未設定なら通知無効。送信失敗でも報告作成は
	// 続行する（best-effort）。webhook URL は秘密値でありログに出力しない。
	DiscordBugReportWebhookURL string
	DiscordWebhookTimeoutMS    int

	// OperationFingerprintCurrentKeyVersion / OperationFingerprintKeysJSON は
	// operation fingerprint の versioned HMAC keyring。
	// 両方必須・両方未設定のいずれかのみ有効（片方のみ、JSON 重複 member、
	// 非object、非文字列 value、空 map、空白・trim 変化 version、不正 base64、
	// decode 後 32 バイト未満、current version 不在は Validate が拒否する）。
	// key 値は秘密値でありログ・エラー出力へ含めない。
	OperationFingerprintCurrentKeyVersion string
	OperationFingerprintKeysJSON          string
}

// Load reads process environment into Config. Release mode refuses
// development database defaults when those variables are omitted.
func Load() *Config {
	ginMode := getEnv("GIN_MODE", "debug")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ekarte_user")
	dbPass := getEnv("DB_PASSWORD", "ekarte_password")
	dbName := getEnv("DB_NAME", "ekarte_db")
	if ginMode == "release" {
		// Release must never inherit a development target when an environment
		// variable is omitted. Validate reports the specific missing field.
		dbHost = os.Getenv("DB_HOST")
		dbPort = os.Getenv("DB_PORT")
		dbUser = os.Getenv("DB_USER")
		dbPass = os.Getenv("DB_PASSWORD")
		dbName = os.Getenv("DB_NAME")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBHost:        dbHost,
		DBPort:        dbPort,
		DBUser:        dbUser,
		DBPass:        dbPass,
		DBName:        dbName,
		DBSSLMode:     getEnv("DB_SSL_MODE", "disable"),
		DBSSLRootCert: os.Getenv("DB_SSL_ROOT_CERT"),
		GinMode:       ginMode,

		DBMaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 50),
		DBMaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBSlowQueryMS:        getEnvInt("DB_SLOW_QUERY_MS", 300),
		DBStatementTimeoutMS: getEnvInt("DB_STATEMENT_TIMEOUT_MS", 120000),
		DBConnectTimeoutSec:  getEnvInt("DB_CONNECT_TIMEOUT_S", 5),

		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		IntegrationEncryptionKey: os.Getenv("INTEGRATION_ENCRYPTION_KEY"),
		S3SharedBucket:           os.Getenv("S3_SHARED_BUCKET"),
		S3SharedRegion:           getEnv("S3_SHARED_REGION", "ap-northeast-1"),
		S3Endpoint:               os.Getenv("S3_ENDPOINT"),

		StorageType:     os.Getenv("STORAGE_TYPE"),
		S3Bucket:        os.Getenv("S3_BUCKET"),
		S3Region:        os.Getenv("S3_REGION"),
		S3PublicBaseURL: os.Getenv("S3_PUBLIC_BASE_URL"),
		AWSAccessKeyID:  os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv(
			"AWS_SECRET_ACCESS_KEY",
		),

		TrustedProxyCIDR: os.Getenv("TRUSTED_PROXY_CIDR"),

		LogLevel: os.Getenv("LOG_LEVEL"),

		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),

		TrimmingImageCleanupEnabled: os.Getenv("TRIMMING_IMAGE_CLEANUP_ENABLED") == "true",

		SchedulerInternalToken: os.Getenv("SCHEDULER_INTERNAL_TOKEN"),

		SupportLLMBaseURL:   getEnv("SUPPORT_LLM_BASE_URL", "https://api.openai.com/v1"),
		SupportLLMAPIKey:    os.Getenv("SUPPORT_LLM_API_KEY"),
		SupportLLMModel:     getEnv("SUPPORT_LLM_MODEL", "gpt-5-nano"),
		SupportLLMTimeoutMS: getEnvInt("SUPPORT_LLM_TIMEOUT_MS", 30000),
		// EMR-410: ハイブリッド検索・外部送信承認（既定はすべて無効/false）。
		SupportHybridSearchEnabled:  getEnvBool("SUPPORT_HYBRID_SEARCH_ENABLED", false),
		SupportEmbeddingBaseURL:     getEnv("SUPPORT_EMBEDDING_BASE_URL", "https://api.sarashina.ai/v1"),
		SupportEmbeddingAPIKey:      os.Getenv("SUPPORT_EMBEDDING_API_KEY"),
		SupportEmbeddingModel:       getEnv("SUPPORT_EMBEDDING_MODEL", "sarashina3-embedding"),
		SupportEmbeddingTimeoutMS:   getEnvInt("SUPPORT_EMBEDDING_TIMEOUT_MS", 1200),
		SupportExternalDataApproved: getEnvBool("SUPPORT_EXTERNAL_DATA_APPROVED", false),

		PlaneBaseURL:           getEnv("PLANE_BASE_URL", "https://api.plane.so/api"),
		PlaneWebBaseURL:        getEnv("PLANE_WEB_BASE_URL", "https://app.plane.so"),
		PlaneWorkspaceSlug:     os.Getenv("PLANE_WORKSPACE_SLUG"),
		PlaneProjectID:         os.Getenv("PLANE_PROJECT_ID"),
		PlaneProjectIdentifier: os.Getenv("PLANE_PROJECT_IDENTIFIER"),
		PlaneAPIKey:            os.Getenv("PLANE_API_KEY"),
		PlaneTimeoutMS:         getEnvInt("PLANE_TIMEOUT_MS", 10000),
		PlaneEnvironment:       planeEnvironmentFor(os.Getenv("APP_ENV"), ginMode),
		PlaneSourceLabelID:     os.Getenv("PLANE_SOURCE_LABEL_ID"),
		PlaneLaneLabelID:       os.Getenv("PLANE_LANE_LABEL_ID"),

		DiscordBugReportWebhookURL: os.Getenv("DISCORD_BUG_REPORT_WEBHOOK_URL"),

		OperationFingerprintCurrentKeyVersion: os.Getenv("OPERATION_FINGERPRINT_CURRENT_KEY_VERSION"),
		OperationFingerprintKeysJSON:          os.Getenv("OPERATION_FINGERPRINT_KEYS_JSON"),
		DiscordWebhookTimeoutMS:    getEnvInt("DISCORD_WEBHOOK_TIMEOUT_MS", 5000),
	}
}

// Validate は起動設定を検証する。
// dev デフォルト値チェックは全モードで実行する（staging 認証バイパス防止）。
func (c *Config) Validate() error {
	if c.JWTSecret == "dev-secret-change-me" {
		return fmt.Errorf("dev default JWT_SECRET is prohibited; set a secure value via the JWT_SECRET environment variable")
	}
	if c.StorageType == "s3" && (c.S3Bucket == "" || c.S3Region == "") {
		return fmt.Errorf("S3_BUCKET and S3_REGION are required when STORAGE_TYPE=s3")
	}
	if c.StorageType == "s3" && c.S3SharedBucket == "" {
		return fmt.Errorf("S3_SHARED_BUCKET is required when STORAGE_TYPE=s3")
	}
	if c.StorageType != "" && c.StorageType != "s3" {
		return fmt.Errorf("STORAGE_TYPE must be empty for local development or s3")
	}
	// Plane 連携が有効（API key 設定）なら label UUID はモードに関わらず必須。
	// 未設定/不正のまま起票すると label 無しチケットができるため fail-fast。
	if c.PlaneAPIKey != "" {
		if err := validatePlaneLabelID("PLANE_SOURCE_LABEL_ID", c.PlaneSourceLabelID); err != nil {
			return err
		}
		if err := validatePlaneLabelID("PLANE_LANE_LABEL_ID", c.PlaneLaneLabelID); err != nil {
			return err
		}
	}
	// operation fingerprint の keyring は全モードで fail-closed。
	// 片方のみ設定・ malformed なまま起動すると永続化済み fingerprint と
	// 計算値がずれ、同一 operation_key の replay/衝突検出が壊れる。
	if _, _, err := c.OperationFingerprintKeys(); err != nil {
		return err
	}
	if c.GinMode != "release" {
		return nil
	}
	if c.TrustedProxyCIDR == "" {
		return fmt.Errorf("TRUSTED_PROXY_CIDR is required in production (rate-limit bypass risk)")
	}
	if err := validateReleaseTrustedProxyCIDR(c.TrustedProxyCIDR); err != nil {
		return err
	}
	if err := validateReleaseCORSOrigins(c.CORSAllowedOrigin); err != nil {
		return err
	}
	if err := validateReleaseFrontendURL(c.FrontendURL); err != nil {
		return err
	}
	if err := validateReleaseDatabaseTarget(c); err != nil {
		return err
	}
	if err := validateReleaseDBSSLMode(c.DBSSLMode); err != nil {
		return err
	}
	if c.DBSSLRootCert != "system" {
		return fmt.Errorf("DB_SSL_ROOT_CERT must be system in release mode")
	}
	if c.DBMaxOpenConns <= 0 || c.DBMaxIdleConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS and DB_MAX_IDLE_CONNS must be positive in release mode")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS in release mode")
	}
	if c.DBMaxOpenConns > maxReleaseDBOpenConns || c.DBMaxIdleConns > maxReleaseDBIdleConns {
		return fmt.Errorf(
			"DB_MAX_OPEN_CONNS must be at most %d and DB_MAX_IDLE_CONNS at most %d in release mode",
			maxReleaseDBOpenConns,
			maxReleaseDBIdleConns,
		)
	}
	if c.StorageType != "s3" {
		return fmt.Errorf("STORAGE_TYPE must be s3 in release mode")
	}
	if strings.TrimSpace(c.AWSAccessKeyID) == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID must be explicitly set when STORAGE_TYPE=s3 in release mode")
	}
	if strings.TrimSpace(c.AWSSecretAccessKey) == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY must be explicitly set when STORAGE_TYPE=s3 in release mode")
	}
	if err := validateOptionalReleaseHTTPSURL("S3_ENDPOINT", c.S3Endpoint); err != nil {
		return err
	}
	if err := validateOptionalReleaseHTTPSURL("S3_PUBLIC_BASE_URL", c.S3PublicBaseURL); err != nil {
		return err
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must be explicitly set in release mode")
	}
	if len(c.JWTSecret) < minimumJWTSecretBytes {
		return fmt.Errorf(
			"JWT_SECRET must be at least %d bytes in release mode",
			minimumJWTSecretBytes,
		)
	}
	// DEC-36 / CMD-02: Worker→Container の X-Scheduler-Token 共有秘密。
	// STG/本番とも secrets.required に含まれるため release では必須。
	if c.SchedulerInternalToken == "" {
		return fmt.Errorf("SCHEDULER_INTERNAL_TOKEN must be explicitly set in release mode")
	}
	if len(c.SchedulerInternalToken) < minimumSchedulerInternalTokenBytes {
		return fmt.Errorf(
			"SCHEDULER_INTERNAL_TOKEN must be at least %d bytes in release mode",
			minimumSchedulerInternalTokenBytes,
		)
	}
	if c.DBPass == "" || c.DBPass == "ekarte_password" {
		return fmt.Errorf("DB_PASSWORD must be explicitly set in release mode")
	}
	if c.IntegrationEncryptionKey == "" {
		return fmt.Errorf("INTEGRATION_ENCRYPTION_KEY must be explicitly set in release mode")
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST must be explicitly set in release mode")
	}
	if c.SMTPUser == "" {
		return fmt.Errorf("SMTP_USER must be explicitly set in release mode")
	}
	if c.SMTPPass == "" {
		return fmt.Errorf("SMTP_PASS must be explicitly set in release mode")
	}
	if c.SMTPFrom == "" {
		return fmt.Errorf("SMTP_FROM must be explicitly set in release mode")
	}
	if c.SMTPPort != "465" && c.SMTPPort != "587" {
		return fmt.Errorf("SMTP_PORT must be 465 (TLS) or 587 (STARTTLS) in release mode, got %s", c.SMTPPort)
	}
	if strings.ContainsAny(c.SMTPFrom, "\r\n") {
		return fmt.Errorf("SMTP_FROM must not contain CR or LF characters")
	}
	if _, err := mail.ParseAddress(c.SMTPFrom); err != nil {
		return fmt.Errorf("SMTP_FROM must be a valid email address")
	}
	if os.Getenv("LIFF_MOCK") == "true" {
		return fmt.Errorf("LIFF_MOCK must not be set in release mode")
	}
	if c.SupportLLMAPIKey != "" {
		if err := validateOptionalReleaseHTTPSURL("SUPPORT_LLM_BASE_URL", c.SupportLLMBaseURL); err != nil {
			return err
		}
		if err := validateOptionalReleaseHTTPSURL("SUPPORT_EMBEDDING_BASE_URL", c.SupportEmbeddingBaseURL); err != nil {
			return err
		}
	}
	if c.PlaneAPIKey != "" {
		if strings.TrimSpace(c.PlaneWorkspaceSlug) == "" || strings.TrimSpace(c.PlaneProjectID) == "" {
			return fmt.Errorf("PLANE_WORKSPACE_SLUG and PLANE_PROJECT_ID are required when PLANE_API_KEY is set")
		}
		if c.PlaneEnvironment == "" {
			return fmt.Errorf("APP_ENV must be development, staging, or production when PLANE_API_KEY is set")
		}
		if err := validateOptionalReleaseHTTPSURL("PLANE_BASE_URL", c.PlaneBaseURL); err != nil {
			return err
		}
		if err := validateOptionalReleaseHTTPSURL("PLANE_WEB_BASE_URL", c.PlaneWebBaseURL); err != nil {
			return err
		}
	}
	if c.DiscordBugReportWebhookURL != "" {
		if err := validateOptionalReleaseHTTPSURL("DISCORD_BUG_REPORT_WEBHOOK_URL", c.DiscordBugReportWebhookURL); err != nil {
			return err
		}
	}
	return nil
}

func validateReleaseTrustedProxyCIDR(rawCIDR string) error {
	_, network, err := net.ParseCIDR(rawCIDR)
	if err != nil {
		return fmt.Errorf("TRUSTED_PROXY_CIDR must be a valid CIDR")
	}
	lastAddress := make(net.IP, len(network.IP))
	for i := range network.IP {
		lastAddress[i] = network.IP[i] | ^network.Mask[i]
	}
	if !isPrivateOrLoopbackIP(network.IP) || !isPrivateOrLoopbackIP(lastAddress) {
		return fmt.Errorf("TRUSTED_PROXY_CIDR must be contained within a private or loopback network")
	}
	return nil
}

func isPrivateOrLoopbackIP(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback()
}

func validateReleaseCORSOrigins(rawOrigins string) error {
	if strings.TrimSpace(rawOrigins) == "" {
		return fmt.Errorf("CORS_ALLOWED_ORIGIN must be explicitly set in release mode")
	}
	for origin := range strings.SplitSeq(rawOrigins, ",") {
		parsed, err := url.Parse(strings.TrimSpace(origin))
		if err != nil ||
			parsed.Scheme != "https" ||
			parsed.Host == "" ||
			parsed.User != nil ||
			parsed.Path != "" ||
			parsed.RawQuery != "" ||
			parsed.Fragment != "" ||
			isLocalHostname(parsed.Hostname()) {
			return fmt.Errorf("CORS_ALLOWED_ORIGIN must contain only HTTPS origins in release mode")
		}
	}
	return nil
}

func validateReleaseFrontendURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil ||
		parsed.Scheme != "https" ||
		parsed.Host == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		isLocalHostname(parsed.Hostname()) {
		return fmt.Errorf("FRONTEND_URL must be an absolute HTTPS URL without user info, query, or fragment in release mode")
	}
	return nil
}

func validateReleaseDatabaseTarget(c *Config) error {
	host := strings.TrimSpace(c.DBHost)
	normalizedHost := strings.TrimSuffix(strings.Trim(host, "[]"), ".")
	if host == "" ||
		strings.ContainsAny(host, " \t\r\n/") ||
		isLocalHostname(normalizedHost) ||
		strings.EqualFold(normalizedHost, "localhost.localdomain") {
		return fmt.Errorf("DB_HOST must be an explicit non-loopback database host in release mode")
	}
	port, err := strconv.Atoi(strings.TrimSpace(c.DBPort))
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("DB_PORT must be an integer between 1 and 65535 in release mode")
	}
	user := strings.TrimSpace(c.DBUser)
	if user == "" || user == "ekarte_user" {
		return fmt.Errorf("DB_USER must be explicitly set to a non-development user in release mode")
	}
	name := strings.TrimSpace(c.DBName)
	if name == "" || name == "ekarte_db" {
		return fmt.Errorf("DB_NAME must be explicitly set to a non-development database in release mode")
	}
	return nil
}

func isLocalHostname(hostname string) bool {
	normalized := strings.TrimSuffix(strings.ToLower(hostname), ".")
	if normalized == "localhost" {
		return true
	}
	ip := net.ParseIP(normalized)
	return ip != nil && ip.IsLoopback()
}

func validateReleaseDBSSLMode(mode string) error {
	switch mode {
	case "verify-full":
		return nil
	default:
		return fmt.Errorf("DB_SSL_MODE must be verify-full in release mode")
	}
}

// planeEnvironmentFor は APP_ENV から Plane external_id 名前空間用の環境名を導出する。
// allowlist (development/staging/production) 以外は release では空文字
// （PLANE_API_KEY 有効時に Validate が拒否）、それ以外では development 既定とする。
func planeEnvironmentFor(appEnv, ginMode string) string {
	switch normalized := strings.ToLower(strings.TrimSpace(appEnv)); normalized {
	case "development", "staging", "production":
		return normalized
	default:
		if ginMode == "release" {
			return ""
		}
		return "development"
	}
}

// validatePlaneLabelID は Plane label 設定値が UUID 形式かを検証する。
// enabled 連携で label 無し起票を作らないための fail-fast。
func validatePlaneLabelID(name, raw string) error {
	if _, err := uuid.Parse(strings.TrimSpace(raw)); err != nil {
		return fmt.Errorf("%s must be a valid label UUID when PLANE_API_KEY is set", name)
	}
	return nil
}

func validateOptionalReleaseHTTPSURL(name, rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil ||
		parsed.Scheme != "https" ||
		parsed.Host == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return fmt.Errorf("%s must be an absolute HTTPS URL without user info, query, or fragment in release mode", name)
	}
	return nil
}

func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode, JapanTimeZone,
	)
	// connect_timeout は pgx 側の接続確立タイムアウト(秒)。0 は省略し pgx 既定に従う。
	if c.DBConnectTimeoutSec > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", c.DBConnectTimeoutSec)
	}
	// statement_timeout は未認識キーとして startup runtime parameter に流れ
	// PostgreSQL 側の文タイムアウト(ms)になる。0 は省略しサーバー既定に従う。
	if c.DBStatementTimeoutMS > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%d", c.DBStatementTimeoutMS)
	}
	if c.DBSSLRootCert != "" {
		dsn += " sslrootcert=" + c.DBSSLRootCert
	}
	return dsn
}

// OperationFingerprintKeys は OPERATION_FINGERPRINT_KEYS_JSON を厳格解析し、
// version→32バイト以上の HMAC key map と current version を返す。
// 両方未設定なら (nil, "", nil)。片方のみ・重複 member・非 object・非文字列
// value・空 map・空白または trim で変化する version・不正/非 canonical
// base64・decode 後 32 バイト未満・current version 不在はすべて error。
// エラーメッセージに version・key の値は含めない（秘密値の漏洩防止）。
func (c *Config) OperationFingerprintKeys() (map[string][]byte, string, error) {
	if c.OperationFingerprintKeysJSON == "" && c.OperationFingerprintCurrentKeyVersion == "" {
		return nil, "", nil
	}
	if c.OperationFingerprintKeysJSON == "" || c.OperationFingerprintCurrentKeyVersion == "" {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON and OPERATION_FINGERPRINT_CURRENT_KEY_VERSION must be set together")
	}
	current := c.OperationFingerprintCurrentKeyVersion
	if strings.TrimSpace(current) != current {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_CURRENT_KEY_VERSION must not be blank or carry surrounding whitespace")
	}
	raw := []byte(c.OperationFingerprintKeysJSON)
	if !utf8.Valid(raw) {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON is not valid UTF-8")
	}
	// token walk で member を1つずつ検査する。map への一度デコードでは
	// 重複 member が黙って後勝ちになるため使わない（fail-closed）。
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON is not valid JSON")
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON must be a JSON object")
	}
	keys := make(map[string][]byte)
	for dec.More() {
		kt, terr := dec.Token()
		if terr != nil {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON is not valid JSON")
		}
		version, ok := kt.(string)
		if !ok || version == "" || strings.TrimSpace(version) != version {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON contains an invalid key version")
		}
		if _, dup := keys[version]; dup {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON contains a duplicate key version")
		}
		vt, terr := dec.Token()
		if terr != nil {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON is not valid JSON")
		}
		encoded, ok := vt.(string)
		if !ok {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON values must be base64 strings")
		}
		decoded, derr := base64.StdEncoding.DecodeString(encoded)
		if derr != nil || base64.StdEncoding.EncodeToString(decoded) != encoded {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON contains an invalid base64 key")
		}
		if len(decoded) < 32 {
			return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON contains a key shorter than 32 bytes")
		}
		keys[version] = decoded
	}
	if _, terr := dec.Token(); terr != nil {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON is not valid JSON")
	}
	// More() は top-level では trailing `}` `]` を見逃すため、EOF で閉じる。
	if tok, terr := dec.Token(); terr != io.EOF || tok != nil {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON has trailing content")
	}
	if len(keys) == 0 {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_KEYS_JSON must not be empty")
	}
	if _, ok := keys[current]; !ok {
		return nil, "", fmt.Errorf("OPERATION_FINGERPRINT_CURRENT_KEY_VERSION is not present in OPERATION_FINGERPRINT_KEYS_JSON")
	}
	return keys, current, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt は正の整数として解釈できる場合のみ環境変数値を採用する。
// 不正値でプール上限が 0(=無制限相当の誤設定)になる事故を防ぐため、失敗時は既定値へフォールバックする。
func getEnvBool(key string, defaultVal bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return defaultVal
	}
	return v == "1" || v == "true" || v == "yes"
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}
