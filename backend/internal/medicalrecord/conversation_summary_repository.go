package medicalrecord

// conversation_summary_repository.go — EMR-485 非表示会話要約（来院カルテ単位）の保存/読戻し adapter。
//
// 範囲: 既存の medical_records / model.MedicalRecord を再利用し、conversation_summary 系3列だけを
// 保存・読戻しする。route 登録・handler・api.yaml・model・migration はこのファイルでは扱わない
// （Phase2 の音声/文字起こし/外部サービス/手動入力formも作らない）。

import (
	"context"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	// conversationSummaryResource は apperrors 用のリソース名（テーブル名ではない）。
	conversationSummaryResource = "conversation_summary"
	// conversationSummaryConflictMessage は楽観ロック不一致の client 向けメッセージ。
	conversationSummaryConflictMessage = "他のユーザーが会話要約を更新しました。再読み込みしてください"
)

// ConversationSummary は medical_records の会話要約3列の読戻し結果。
//
// 契約（EMR-485）:
//   - 未保存（旧未保存データ）は Summary / Revision / Source がすべて nil。
//     「不明(unknown)」を推定値・空文字・0 で置き換えない。
//   - Revision はサーバ側採番のみ。初回保存で 1、保存のたびに +1（呼び出し元は指定しない）。
//   - Source は保存元を識別する非空白文字列。DB CHECK ck_conversation_summary_metadata と
//     同じ条件をコード側でも検証する（AutoMigrate のテストDBには CHECK が載らないため）。
//   - json タグを持たず、route/handler へ公開しない非表示の保存専用データ。
type ConversationSummary struct {
	MedicalRecordID uint64
	Summary         *string
	Revision        *int64
	Source          *string
}

// SaveConversationSummaryCommand は保存の入力。Revision はサーバ側で採番するため受け取らない。
type SaveConversationSummaryCommand struct {
	ClinicID        uint64
	MedicalRecordID uint64
	Summary         string
	Source          string
	// ExpectedRevision は楽観ロック用の期待版。
	//   nil : 照合スキップ（最終書き込み優先。冪等再送もこの経路）
	//   0   : 「未保存（3列すべて NULL）」を期待（CHECK 制約により保存済み revision は 1 以上）
	//   n>0 : 現在の revision が n のときだけ保存する
	//
	// 再送例外: 0行更新のときは保存済み内容が本commandの意図（summary+source）と
	// 一致するかを先に判定する。一致する再送（同一要求の再送・期待版が実際の版より
	// 古い/未来値の場合を含む）は冪等成功で返し、内容が異なる期待版不一致だけを
	// conflict にする（Save の doc comment を正本とする）。
	ExpectedRevision *int64
}

// validate は保存経路の入力契約を守る。
//
//   - clinic/record 特定・source 非空白・revision>0 は DB CHECK
//     ck_conversation_summary_metadata と同じ条件のコード側検証
//     （AutoMigrate のテストDBには CHECK が載らないため）。
//   - 空本文（空白のみ）の拒否は DB CHECK より狭い**追加の防御的判断**であり
//     承認済み契約ではない。空本文の保存可否が確定した場合はこの検証を見直す。
func (c SaveConversationSummaryCommand) validate() error {
	if c.ClinicID == 0 || c.MedicalRecordID == 0 {
		return apperrors.WrapInvalidInput("医院とカルテを特定できません")
	}
	if strings.TrimSpace(c.Summary) == "" {
		return apperrors.WrapInvalidInput("会話要約の本文が空です")
	}
	if strings.TrimSpace(c.Source) == "" {
		return apperrors.WrapInvalidInput("会話要約の保存元を指定してください")
	}
	if c.ExpectedRevision != nil && *c.ExpectedRevision < 0 {
		return apperrors.WrapInvalidInput("期待する会話要約の版が不正です")
	}
	return nil
}

// ConversationSummaryRepository は会話要約の保存/読戻し。
// 他医院のカルテは存在有無を漏らさず NotFound として扱う。
type ConversationSummaryRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*ConversationSummary, error)
	Save(ctx context.Context, cmd SaveConversationSummaryCommand) (*ConversationSummary, error)
}

type conversationSummaryRepository struct {
	db *gorm.DB
}

// NewConversationSummaryRepository は ConversationSummaryRepository を初期化して返す。
func NewConversationSummaryRepository(db *gorm.DB) ConversationSummaryRepository {
	return &conversationSummaryRepository{db: db}
}

// FindByMedicalRecordID は clinic_id + medical_record_id で1件に限定して3列を返す。
// 他医院・不存在・論理削除は NotFound（clinic 不一致を存在有無で漏らさない）。
func (r *conversationSummaryRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*ConversationSummary, error) {
	var row model.MedicalRecord
	err := persistence.DBOrTx(ctx, r.db).
		Select("id", "conversation_summary", "conversation_summary_revision", "conversation_summary_source").
		Where("id = ? AND clinic_id = ?", medicalRecordID, clinicID).
		First(&row).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, conversationSummaryResource, strconv.FormatUint(medicalRecordID, 10))
	}
	return &ConversationSummary{
		MedicalRecordID: row.ID,
		Summary:         row.ConversationSummary,
		Revision:        row.ConversationSummaryRevision,
		Source:          row.ConversationSummarySource,
	}, nil
}

// Save は会話要約3列のみを原子的に更新する。
//
//   - 対象は clinic_id + medical_record_id で限定する（他医院・他カルテ・他の診療記録列は触らない）。
//     UpdateColumns を使うため updated_at 等の自動タイムスタンプも進めない
//     （非表示保存が通常カルテの「最終更新」に見えることを防ぐ）。
//   - revision は UPDATE 式 COALESCE(conversation_summary_revision, 0) + 1 で採番する。
//     read-modify-write を避けるため、同時保存でも版が重複しない。旧未保存（NULL）からの
//     初回保存は 1 になる。
//   - 成功時は RETURNING で自分の書込みrevisionを同一statement内で受け取る。
//     UPDATE と読戻しの間に他writerが割り込んでも、自分の commit 済み結果を
//     正しく返せる（lost update / 誤conflict の競合窗口が存在しない）。
//   - 同一内容（summary と source が不変）の行は WHERE で除外するため、再送しても revision は
//     据え置き（no-op 成功）。0 行更新のときは読戻しで分類する: 内容一致は冪等成功、
//     対象なしは NotFound、期待版不一致は conflict。
//   - 0 行更新経路は commit 済みの書込みを伴わないため、backend/CLAUDE.md
//     「commit済みの成功を後段の再取得errorで失敗応答へ反転させない」に触れない。
func (r *conversationSummaryRepository) Save(ctx context.Context, cmd SaveConversationSummaryCommand) (*ConversationSummary, error) {
	if err := cmd.validate(); err != nil {
		return nil, err
	}
	// RETURNING で自分の書込みrevisionを同一statement内で受け取る。
	// UPDATE と読戻しの間に他writerが割り込んでも、commit 済みの自分の結果を正しく返せる
	// （lost update / 誤conflict の競合窗口が存在しない）。updated_at は触れない。
	var revision int64
	query := `UPDATE medical_records SET
		conversation_summary = ?,
		conversation_summary_source = ?,
		conversation_summary_revision = COALESCE(conversation_summary_revision, 0) + 1
	WHERE id = ? AND clinic_id = ?
		AND (conversation_summary IS DISTINCT FROM ? OR conversation_summary_source IS DISTINCT FROM ?)`
	args := []any{cmd.Summary, cmd.Source, cmd.MedicalRecordID, cmd.ClinicID, cmd.Summary, cmd.Source}
	if cmd.ExpectedRevision != nil {
		// 保存済み revision は CHECK 制約により 1 以上なので、0 は「未保存」と一意に対応する。
		query += " AND COALESCE(conversation_summary_revision, 0) = ?"
		args = append(args, *cmd.ExpectedRevision)
	}
	result := persistence.DBOrTx(ctx, r.db).Raw(query, args...).Scan(&revision)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, conversationSummaryResource, strconv.FormatUint(cmd.MedicalRecordID, 10))
	}
	if result.RowsAffected == 1 {
		rev := revision
		return &ConversationSummary{
			MedicalRecordID: cmd.MedicalRecordID,
			Summary:         &cmd.Summary,
			Source:          &cmd.Source,
			Revision:        &rev,
		}, nil
	}

	// 0 行更新: commit 済みの書込みを伴わないため、読戻しで分類する。
	current, err := r.FindByMedicalRecordID(ctx, cmd.ClinicID, cmd.MedicalRecordID)
	if err != nil {
		return nil, err // 対象なし（他医院を含む）は NotFound
	}
	if conversationSummaryMatches(current, cmd) {
		// 同一内容の再送（期待版が古い場合を含む）。revision も含め保存済みの値をそのまま返す。
		return current, nil
	}
	if cmd.ExpectedRevision != nil && conversationSummaryRevisionValue(current) != *cmd.ExpectedRevision {
		return nil, apperrors.WrapConflict(conversationSummaryConflictMessage)
	}
	// 内容が異なり、期待版指定もないのに 0 行更新になる経路はない（WHERE は内容差分だけ）。
	// 到達した場合は競合状態として conflict に倒す（fail-closed）。
	return nil, apperrors.WrapConflict(conversationSummaryConflictMessage)
}

// conversationSummaryRevisionValue は未保存（nil）を 0 として扱う（CHECK 制約と整合）。
func conversationSummaryRevisionValue(row *ConversationSummary) int64 {
	if row == nil || row.Revision == nil {
		return 0
	}
	return *row.Revision
}

// conversationSummaryMatches は保存済み内容が cmd と同一かを返す（冪等再送の判定）。
func conversationSummaryMatches(row *ConversationSummary, cmd SaveConversationSummaryCommand) bool {
	if row == nil || row.Summary == nil || row.Source == nil {
		return false
	}
	return *row.Summary == cmd.Summary && *row.Source == cmd.Source
}
