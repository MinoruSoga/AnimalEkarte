package medicalrecord

import (
	"context"
	"log/slog"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CageService ----

// CreateCageInput はケージ作成の入力DTO
type CreateCageInput struct {
	Name        string
	CageType    string
	CageSize    string
	Price       *int64
	IsActive    bool
	Description string
	SortOrder   int
}

// UpdateCageInput はケージ更新のサービス入力 DTO
type UpdateCageInput struct {
	Name        *string
	CageType    *string
	CageSize    *string
	Price       *int64
	IsActive    *bool
	Description *string
	SortOrder   *int
}

const (
	colCageName        = "name"
	colCageCageType    = "cage_type"
	colCageCageSize    = "cage_size"
	colCagePrice       = "price"
	colCageIsActive    = "is_active"
	colCageDescription = "description"
	colCageSortOrder   = "sort_order"
)

func buildCageUpdate(input *UpdateCageInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCageName] = *input.Name
	}
	if input.CageType != nil {
		fields[colCageCageType] = model.CageType(*input.CageType)
	}
	if input.CageSize != nil {
		fields[colCageCageSize] = model.CageSize(*input.CageSize)
	}
	if input.Price != nil {
		fields[colCagePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colCageIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colCageDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colCageSortOrder] = *input.SortOrder
	}
	return fields
}

type CageService interface {
	List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// CageMerger はケージ一時連結（EMR-306）の書込み面。*cageService が実装する。
// CageService へはメソッドを追加しない（INTERFACE STABILITY: 非所有モックが
// CageService を実装しているためコンパイルを壊さない）。ハンドラは
// service.(CageMerger) の型アサーションで取得する。
type CageMerger interface {
	// Merge は memberCageIDs を代表ケージ representativeID のもとに連結し、
	// 代表ケージを返す。memberCageIDs は1件以上・重複なし・同一医院・有効・
	// 未連結・グループ非代表・占有なしのケージのみ許容する。
	Merge(ctx context.Context, clinicID, representativeID uint64, memberCageIDs []uint64) (*model.Cage, error)
	// Unmerge は representativeID が率いるグループの全構成ケージの連結を解除する。
	Unmerge(ctx context.Context, clinicID, representativeID uint64) error
}

// CageOccupancyReader は連結ピッカーの占有読み取り面（EMR-306 review）。
// *cageService が実装し、ハンドラーは service.(CageOccupancyReader) の
// 型アサーションで取得する。GET /hospitalizations（hospitalization:view）を
// 経由せず master-hospitalization:edit 権限で占有集合を返すための面。
type CageOccupancyReader interface {
	// OccupiedCageIDs は未削除の admitted/reserved 入院が乗っているケージ id を返す。
	OccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error)
}

type cageService struct {
	repo       CageRepository
	transactor Transactor
}

// NewCageService constructs CageService.
// transactor is required for Delete: lock + usage check + soft-delete share one
// ambient transaction (SEC-CS-F13; merchandise/procedure atomic-delete pattern).
func NewCageService(repo CageRepository, transactor Transactor) CageService {
	return &cageService{repo: repo, transactor: transactor}
}

func (s *cageService) List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	result, err := s.repo.FindAll(ctx, clinicID, cageType)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list cage")
	}
	return result, nil
}
func (s *cageService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get cage")
	}
	return result, nil
}
func (s *cageService) Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	// Service は Handler 以外から直接呼ばれる可能性があるため防御的バリデーションを維持する
	if err := validateCageType(input.CageType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate cage type")
	}
	if err := validateCageSize(input.CageSize); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate cage size")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	cage := &model.Cage{
		ClinicID:    clinicID,
		Name:        input.Name,
		CageType:    model.CageType(input.CageType),
		CageSize:    model.CageSize(input.CageSize),
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, cage); err != nil {
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Name,
			apperrors.ConstraintCageName,
			apperrors.CodeCageNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to create cage")
	}
	slog.InfoContext(ctx, "cage created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("cage_id", cage.ID))
	return cage, nil
}
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get cage")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	// Service は Handler 以外から直接呼ばれる可能性があるため防御的バリデーションを維持する
	if input.CageType != nil {
		if err := validateCageType(*input.CageType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate cage type")
		}
	}
	if input.CageSize != nil {
		if err := validateCageSize(*input.CageSize); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate cage size")
		}
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if len(buildCageUpdate(input)) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	// EMR-306: is_active=false の無効化は連結状態との競合を FOR UPDATE で直列化して判定する。
	if input.IsActive != nil && !*input.IsActive {
		return s.deactivateCage(ctx, clinicID, id, input)
	}
	cage, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, cageUpdateError(err, input.Name)
	}
	slog.InfoContext(ctx, "cage updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("cage_id", id))
	return cage, nil
}

// cageUpdateError は repo Update 系の失敗を共通変換する: cage 名の unique 制約
// 違反は 409 conflict、それ以外は internal wrap。Update と deactivateCage で共有する。
func cageUpdateError(err error, name *string) error {
	nameForConflict := ""
	if name != nil {
		nameForConflict = *name
	}
	if conflict := apperrors.AsNameUniqueConflict(
		err,
		nameForConflict,
		apperrors.ConstraintCageName,
		apperrors.CodeCageNameConflict,
	); conflict != nil {
		return conflict
	}
	return apperrors.Wrap(err, "failed to update cage")
}
func (s *cageService) Delete(ctx context.Context, clinicID, id uint64) error {
	// SEC-CS-F13: FOR UPDATE → usage count → soft-delete in one ambient tx so a
	// concurrent hospitalization assignment cannot sneak between check and delete.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("cage write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		cage, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get cage")
		}
		// EMR-306: 連結中のケージ（構成・代表の双方）は削除不可。解除してから削除する。
		if cage.MergedIntoCageID != nil {
			return apperrors.WrapConflict("連結中のケージは削除できません")
		}
		// 代表判定は merge 面を必須とする fail-closed（Merge/Unmerge/入院割当ガードと同一姿勢。
		// 面なし注入で代表判定が素通りすると、連結中の代表を削除できてしまうため）。
		mergeRepo := s.mergeRepo()
		if mergeRepo == nil {
			return apperrors.WrapInternalServerError("cage merge repository dependency is required")
		}
		memberCounts, err := mergeRepo.CountMergedMembers(txCtx, clinicID, []uint64{id})
		if err != nil {
			return apperrors.Wrap(err, "failed to count merged members")
		}
		if memberCounts[id] > 0 {
			return apperrors.WrapConflict("連結グループの代表ケージは削除できません")
		}
		count, err := s.repo.CountUsageByCageID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check cage usage")
		}
		if count > 0 {
			return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete cage")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "cage deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("cage_id", id))
	return nil
}

func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	// SEC-CS-F04: bound cardinality + uniqueness before fan-out UPDATE per id.
	if err := httpapi.ValidateReorderIDs(ids); err != nil {
		return err
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder cage")
	}
	slog.InfoContext(ctx, "cage reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

// ---- cage merge (EMR-306) ----

// mergeRepo は任意の連結データアクセス面を解決する。CageRepository のみを実装する
// 非所有モックでは nil を返す。全呼び出し側（Merge/Unmerge・Delete/deactivateCage の
// 連結ガード・入院割当ガード）は nil を fail-closed（500）とする — 面なし注入で
// 代表判定が素通りすると連結中の代表を削除/無効化できてしまうため。
func (s *cageService) mergeRepo() CageMergeRepository {
	mergeRepo, _ := s.repo.(CageMergeRepository)
	return mergeRepo
}

var _ CageMerger = (*cageService)(nil)
var _ CageOccupancyReader = (*cageService)(nil)

// OccupiedCageIDs は連結ピッカー向けの占有ケージ id 集合を返す。
// mergeRepo 面がない注入では 500 fail-closed（他の連結ガードと同一姿勢）。
func (s *cageService) OccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error) {
	mergeRepo := s.mergeRepo()
	if mergeRepo == nil {
		return nil, apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	ids, err := mergeRepo.ListOccupiedCageIDs(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list occupied cages")
	}
	return ids, nil
}

// maxCageMergeMembers は1回の連結で指定できる構成ケージ数の上限。
// mergeCageRequest の binding max と同一値を保つ（gin tag はリテラル必須）。
const maxCageMergeMembers = 32

func (s *cageService) Merge(ctx context.Context, clinicID, representativeID uint64, memberCageIDs []uint64) (*model.Cage, error) {
	mergeRepo := s.mergeRepo()
	if mergeRepo == nil {
		return nil, apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("cage write transaction dependency is required")
	}
	if len(memberCageIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("member_cage_ids は1件以上指定してください")
	}
	// SEC-CS-F04 同型: member 配列のファンアウト UPDATE を上限で制限する。
	// 運用上の連結は隣接する少数ケージ（2〜4室）を想定しており、上限は異常入力の
	// 早期拒否用（binding の max タグはリテラル必須のためサービス側でも検査する）。
	if len(memberCageIDs) > maxCageMergeMembers {
		return nil, apperrors.WrapInvalidInput("member_cage_ids は32件以内で指定してください")
	}
	seen := make(map[uint64]struct{}, len(memberCageIDs))
	for _, memberID := range memberCageIDs {
		if memberID == representativeID {
			return nil, apperrors.WrapInvalidInput("代表ケージ自身は連結メンバーに指定できません")
		}
		if _, dup := seen[memberID]; dup {
			return nil, apperrors.WrapInvalidInput("member_cage_ids に重複したIDが含まれています")
		}
		seen[memberID] = struct{}{}
	}

	// 代表+構成を id 昇順に並べ、全行を同一順序で FOR UPDATE ロックする（デッドロック回避）。
	allIDs := append([]uint64{representativeID}, memberCageIDs...)
	sort.Slice(allIDs, func(i, j int) bool { return allIDs[i] < allIDs[j] })

	var representative *model.Cage
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		cages, err := mergeRepo.LockCagesForUpdate(txCtx, clinicID, allIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock cages")
		}
		for i := range cages {
			cage := &cages[i]
			if !cage.IsActive {
				return apperrors.WrapConflict("無効化されているケージは連結できません")
			}
			if cage.MergedIntoCageID != nil {
				return apperrors.WrapConflict("すでに連結済みのケージは連結できません")
			}
			if cage.ID == representativeID {
				representative = cage
			}
		}
		// 単一階層のみ: 代表・構成のどれも既存グループの代表であってはならない（ネスト拒否）。
		memberCounts, err := mergeRepo.CountMergedMembers(txCtx, clinicID, allIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to count merged members")
		}
		for _, n := range memberCounts {
			if n > 0 {
				return apperrors.WrapConflict("連結グループの代表ケージは別のグループに連結できません")
			}
		}
		occupied, err := mergeRepo.CountOccupiedCages(txCtx, clinicID, allIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to check cage occupancy")
		}
		for _, n := range occupied {
			if n > 0 {
				return apperrors.WrapConflict("入院中または予約中のケージは連結できません")
			}
		}
		if err := mergeRepo.SetMergedInto(txCtx, clinicID, representativeID, memberCageIDs); err != nil {
			return apperrors.Wrap(err, "failed to merge cages")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "cages merged",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("representative_cage_id", representativeID),
		slog.Int("member_count", len(memberCageIDs)))
	return representative, nil
}

func (s *cageService) Unmerge(ctx context.Context, clinicID, representativeID uint64) error {
	mergeRepo := s.mergeRepo()
	if mergeRepo == nil {
		return apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("cage write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 代表+構成を1文の id 昇順 FOR UPDATE で原子的にロック（Merge と同一の
		// 昇順ロック規約でデッドロックを避け、集合の発見とロックを同時に行う）。
		group, err := mergeRepo.LockCageGroupForUpdate(txCtx, clinicID, representativeID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock cage group")
		}
		var representative *model.Cage
		groupIDs := make([]uint64, 0, len(group))
		memberCount := 0
		for i := range group {
			groupIDs = append(groupIDs, group[i].ID)
			if group[i].ID == representativeID {
				representative = &group[i]
			} else {
				memberCount++
			}
		}
		if representative.MergedIntoCageID != nil {
			return apperrors.WrapConflict("連結メンバーのケージは連結解除できません")
		}
		if memberCount == 0 {
			return apperrors.WrapConflict("連結グループを持たないケージは連結解除できません")
		}
		// EMR-306: 占有検査は構成ケージだけでなく代表を含むグループ全体に対して行う。
		// 入院は代表ケージにのみ紐付くため、構成ケージ限定の検査は常に空振りし、
		// 占有中の連結室を誤って解除できてしまう。spec 文言の意図（占有中は解除不可）
		// に合わせてグループ単位で判定する。
		occupied, err := mergeRepo.CountOccupiedCages(txCtx, clinicID, groupIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to check cage occupancy")
		}
		for _, n := range occupied {
			if n > 0 {
				return apperrors.WrapConflict("入院中または予約中のケージを含む連結は解除できません")
			}
		}
		if err := mergeRepo.ClearMergedInto(txCtx, clinicID, representativeID); err != nil {
			return apperrors.Wrap(err, "failed to unmerge cages")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "cages unmerged",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("representative_cage_id", representativeID))
	return nil
}

// deactivateCage は is_active=false 更新を FOR UPDATE 行ロック下で実行し、
// Merge との競合を直列化する（EMR-306）。連結中の構成ケージおよびグループ代表
// ケージは、解除するまで無効化できない。merged_into_cage_id 自体はここでは触らない。
func (s *cageService) deactivateCage(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("cage write transaction dependency is required")
	}
	var cage *model.Cage
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get cage")
		}
		// EMR-306: 連結中のケージ（構成・代表の双方）は無効化不可。解除してから行う。
		if locked.MergedIntoCageID != nil {
			return apperrors.WrapConflict("連結中のケージは無効化できません")
		}
		// 代表判定は merge 面を必須とする fail-closed（Delete と同一姿勢）。
		mergeRepo := s.mergeRepo()
		if mergeRepo == nil {
			return apperrors.WrapInternalServerError("cage merge repository dependency is required")
		}
		memberCounts, err := mergeRepo.CountMergedMembers(txCtx, clinicID, []uint64{id})
		if err != nil {
			return apperrors.Wrap(err, "failed to count merged members")
		}
		if memberCounts[id] > 0 {
			return apperrors.WrapConflict("連結グループの代表ケージは無効化できません")
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, *input)
		if err != nil {
			return cageUpdateError(err, input.Name)
		}
		cage = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "cage deactivated", slog.Uint64("clinic_id", clinicID), slog.Uint64("cage_id", id))
	return cage, nil
}
