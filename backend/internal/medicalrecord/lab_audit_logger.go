package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// LabAuditLogger は lab import 操作の監査イベントを記録するインターフェース。
//
// # PII 安全ポリシー
//
// audit payload には以下を含めない:
//   - 患者名・飼い主名・電子メール・電話番号・住所 (PHI)
//   - raw lab result values（検査値そのもの）
//   - source_fingerprint（バッチ識別子: 再現可能な識別情報）以外のペイロード内容
//   - リクエストボディ全体
//
// 許可するメタデータ: job_id, clinic_id, actor_id, source_type, status, counts, error_category
//
// # 失敗ポリシー
//
// 監査ログ書き込みの失敗は lab import フローを中断しない。
// 失敗した場合は slog.WarnContext でログを残すが、呼び出し元にエラーを返さない。
// これは audit_logs テーブルが best-effort の観測目的であり、
// 書き込み失敗が import の整合性に影響しないためである。
//
// Phase 4A スコープ: preview / commit / source_blocked / failure の各操作をカバーする。
type LabAuditLogger interface {
	// LogPreviewRequested は preview エンドポイントへのリクエスト受信を記録する。
	LogPreviewRequested(ctx context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int)
	// LogCommitRequested は commit エンドポイントへのリクエスト受信を記録する。
	LogCommitRequested(ctx context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int)
	// LogCommitSucceeded は commit の正常完了を記録する。
	LogCommitSucceeded(ctx context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, counts CommitAuditCounts)
	// LogCommitFailed は commit の失敗を記録する（job 作成前の失敗も含む）。
	// errorCategory には model.LabAuditErrorCategory の定数を使用すること。free-form string は受け付けない。
	LogCommitFailed(ctx context.Context, clinicID uint64, actorID *uint64, errorCategory model.LabAuditErrorCategory)
	// LogSourceBlocked は drwan / manual 等のブロック対象ソースへの操作を記録する。
	// reason には model.LabBlockedReason の定数を使用すること。free-form string は受け付けない。
	LogSourceBlocked(ctx context.Context, clinicID uint64, actorID *uint64, sourceType, operation string, reason model.LabBlockedReason)
	// LogRevertSucceeded は compensating revert（persisted → reverted）の成功を記録する。
	// commit 成功監査と混同しない（TASK-032 / DEC-57）。
	LogRevertSucceeded(ctx context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, retractedExamCount int)
}

// CommitAuditCounts は commit 成功時の集計カウンタ。PII を含まない。
type CommitAuditCounts struct {
	RowCount       int
	PersistedCount int
	DuplicateCount int
	FailedCount    int
}

// labAuditLogger は非tx の AuditLogger（共有監査カーネルの LogEntry view）を使って
// lab import 監査イベントを記録する。
type labAuditLogger struct {
	audit AuditLogger
}

// NewLabAuditLogger は LabAuditLogger を初期化して返す。
func NewLabAuditLogger(audit AuditLogger) LabAuditLogger {
	return &labAuditLogger{audit: audit}
}

func (l *labAuditLogger) LogPreviewRequested(ctx context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int) {
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportPreviewRequested, map[string]any{
		"source_type": sourceType,
		"row_count":   rowCount,
	})
}

func (l *labAuditLogger) LogCommitRequested(ctx context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int) {
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportCommitRequested, map[string]any{
		"source_type": sourceType,
		"row_count":   rowCount,
	})
}

func (l *labAuditLogger) LogCommitSucceeded(ctx context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, counts CommitAuditCounts) {
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportCommitSucceeded, map[string]any{
		"job_id":          jobID.String(),
		"row_count":       counts.RowCount,
		"persisted_count": counts.PersistedCount,
		"duplicate_count": counts.DuplicateCount,
		"failed_count":    counts.FailedCount,
	})
}

func (l *labAuditLogger) LogCommitFailed(ctx context.Context, clinicID uint64, actorID *uint64, errorCategory model.LabAuditErrorCategory) {
	if !model.ValidLabAuditErrorCategory(errorCategory) {
		slog.WarnContext(ctx, "lab audit: LogCommitFailed called with invalid errorCategory — skipping (fail-closed)",
			"clinic_id", clinicID)
		return
	}
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportCommitFailed, map[string]any{
		"error_category": string(errorCategory),
	})
}

func (l *labAuditLogger) LogSourceBlocked(ctx context.Context, clinicID uint64, actorID *uint64, sourceType, operation string, reason model.LabBlockedReason) {
	if !model.ValidLabBlockedReason(reason) {
		slog.WarnContext(ctx, "lab audit: LogSourceBlocked called with invalid reason — skipping (fail-closed)",
			"reason", string(reason), "clinic_id", clinicID)
		return
	}
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportSourceBlocked, map[string]any{
		"source_type": sourceType,
		"operation":   operation,
		"reason":      string(reason),
	})
}

func (l *labAuditLogger) LogRevertSucceeded(ctx context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, retractedExamCount int) {
	l.logBestEffort(ctx, clinicID, actorID, model.AuditActionLabImportRevertSucceeded, map[string]any{
		"job_id":               jobID.String(),
		"retracted_exam_count": retractedExamCount,
	})
}

// logBestEffort は AuditLogger.LogEntry を呼び出し、失敗時は warn ログを残して無視する。
// 失敗が lab import フローを中断しないようにするためのベストエフォート設計。
func (l *labAuditLogger) logBestEffort(ctx context.Context, clinicID uint64, actorID *uint64, action string, metadata map[string]any) {
	actorType := auditActorTypeFor(actorID)
	if err := l.audit.LogEntry(ctx, &AuditEntry{
		ClinicID:  &clinicID,
		ActorID:   actorID,
		ActorType: actorType,
		Action:    action,
		Resource:  model.AuditResourceLabImport,
		Metadata:  metadata,
	}); err != nil {
		slog.WarnContext(ctx, "lab audit log failed (best-effort)",
			"action", action, "clinic_id", clinicID, "error", err)
	}
}
