package medicalrecord

// hospitalization_service_cage_swap.go — EMR-306 review: ボード DnD の入れ替えを
// 単一トランザクションで行う経路。従来の2本の独立 PATCH では、片方がケージガード
// （連結メンバーへの 422 / 占有グループへの 409）で拒否されてももう片方が commit
// され得る — 特に preflight 判定と PATCH の間に別セッションがケージを連結した
// TOCTOU ケースで半スワップ（両患者が片側ケージ）が残った。ここでは両方の更新を
// 同一 tx に入れ、どちらかがガードで落ちれば両方ロールバックする。

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// petRowLockFinder は SwapCages が入れ替え対象ペットのローカル行を同一 tx で FOR UPDATE するための最小面。
// petdomain.RowLockingRepository（CompleteRepository にのみ載せる narrow capability）と
// 構造的に互換で、注入された petRepo が面を持たなければ 500 fail-closed にする
// （doctorVerifier / CageMergeRepository の型アサーション拡張と同じ先例）。
// 死亡判定はこのローカル行ロックだけでは成立しない（死亡は共通 subject の現行 revision が持つ）ため、
// ロック後に petLife（EMR-491 D6 reader）が subject を FOR SHARE ロックして判定する。
type petRowLockFinder interface {
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// SwapCages は2件の入院の cage_id を単一トランザクションで交換する。
// expectedCageIDs は (idA, idB) と同順の表示時割当で、不一致は409で拒否する。
// 連結ケージが絡む場合でも通常の割当ガードが各移動に適用され、構成ケージ宛は 422、
// 当事者以外が占有するケージ/グループ宛は 409 で両方ロールバックする（半スワップ状態を残さない）。
func (s *hospitalizationService) SwapCages(ctx context.Context, clinicID, idA, idB uint64, expectedCageIDs [2]uint64) (*model.Hospitalization, *model.Hospitalization, error) {
	if idA == idB {
		return nil, nil, apperrors.WrapInvalidInput("同一の入院は入れ替えできません")
	}
	if expectedCageIDs[0] == 0 || expectedCageIDs[1] == 0 {
		return nil, nil, apperrors.WrapInvalidInput("表示時のケージIDを2件指定してください")
	}
	if s.transactor == nil {
		return nil, nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	petLocker, ok := s.petRepo.(petRowLockFinder)
	if !ok {
		// Codex PR606 #26 再指摘: 死亡判定はペット行ロック必須。面を持たない
		// 注入 repo（テスト用の FindByID だけのモック等）では tx 直列化を
		// 保証できないため、黙って通常 read には退避せず fail-closed にする。
		return nil, nil, apperrors.WrapInternalServerError("cage swap requires a pet repository with row locking")
	}
	if s.petLife == nil {
		return nil, nil, apperrors.WrapInternalServerError("cage swap requires a pet life state reader")
	}
	var outA, outB *model.Hospitalization
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 行ロックは入院 id 昇順で取得 — スワップ同士や Update（同じく行ロック先行）
		// とのロック順を一致させデッドロックを防ぐ。
		firstID, secondID := idA, idB
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}
		first, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, firstID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization")
		}
		if first == nil {
			return apperrors.WrapNotFound("hospitalization", strconv.FormatUint(firstID, 10))
		}
		second, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, secondID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization")
		}
		if second == nil {
			return apperrors.WrapNotFound("hospitalization", strconv.FormatUint(secondID, 10))
		}

		// 表示時の割当と、両行ロック後の現在値を元の入力順で照合する。
		// ID昇順のロック取得は expectedCageIDs と入院IDの対応を変更しない。
		inputOrder := [2]*model.Hospitalization{first, second}
		if first.ID != idA {
			inputOrder[0], inputOrder[1] = second, first
		}
		for i, hospitalization := range inputOrder {
			if hospitalization.CageID == nil || *hospitalization.CageID != expectedCageIDs[i] {
				return hospitalizationCageConflict(codeCageAssignmentChanged, "ケージの割当が変更されています。最新の入院ボードを確認してください")
			}
		}

		// Codex PR606 #26 再指摘: 死亡判定も同じ tx で行う。通常 FindByID だと
		// 「判定 → 別セッションの RecordDeath commit → 本 tx commit」の直列化が
		// 成立せず、死亡直後のペットを入れ替え得る。ペット行は pet_id 昇順・重複排除で
		// FOR UPDATE してから契約判定する（入院行の id 昇順ロックと同じ規約で
		// ロック順を揃え、デッドロックを防ぐ）。
		// EMR-491 D6: ローカル pet 行ロックだけでは死亡判定の直列化にならないため、各ペットの
		// ローカル行ロック直後に共通 subject の FOR SHARE ロックと現行 revision の再解決を同一 tx で行う。
		petIDs := []uint64{first.PetID}
		if second.PetID != first.PetID {
			petIDs = append(petIDs, second.PetID)
			if petIDs[0] > petIDs[1] {
				petIDs[0], petIDs[1] = petIDs[1], petIDs[0]
			}
		}
		for _, petID := range petIDs {
			pet, err := petLocker.LockByIDForUpdate(txCtx, clinicID, petID)
			if err != nil {
				return apperrors.Wrap(err, "failed to verify pet status")
			}
			if pet == nil {
				return apperrors.WrapNotFound("pet", "status")
			}
			if err := validatePetNotDeceased(txCtx, s.petLife, clinicID, petID, "死亡したペットの入院はケージを入れ替えできません"); err != nil {
				return err
			}
		}

		if first.CageID == nil || second.CageID == nil {
			return apperrors.WrapInvalidInput("ケージ未割当の入院は入れ替えできません")
		}
		if *first.CageID == *second.CageID {
			return apperrors.WrapInvalidInput("同一ケージ内の入れ替えは不要です")
		}

		// 割当ガードはターゲットケージ id の昇順で実行し、2グループのケージ行ロック
		// 取得順を merge/unmerge・単体割当と同じ昇順規約に揃える。
		moves := []struct {
			target uint64
		}{
			{*second.CageID},
			{*first.CageID},
		}
		if moves[1].target < moves[0].target {
			moves[0], moves[1] = moves[1], moves[0]
		}
		// Lock both cage groups before reading either subject's stay. Keep the
		// existing cage -> stay order and the original ambient transaction.
		groups := make([][]model.Cage, 0, len(moves))
		for _, mv := range moves {
			group, err := s.lockCageAssignmentGroup(txCtx, clinicID, mv.target)
			if err != nil {
				return err
			}
			groups = append(groups, group)
		}
		// Neither subject may disappear through self exclusion before its own
		// current stay has been resolved as reserved/admitted.
		if err := s.validateCageAssignmentSubjects(txCtx, clinicID, []*model.Hospitalization{first, second}); err != nil {
			return err
		}
		swapping := []uint64{first.ID, second.ID}
		for _, group := range groups {
			if err := s.validateCageAssignmentOccupancy(txCtx, clinicID, group, swapping); err != nil {
				return err
			}
		}

		// 更新も行ロックと同じ昇順で適用（各行は既に FOR UPDATE 済み）。
		updatedFirst, err := s.hospRepo.Update(txCtx, clinicID, first.ID,
			UpdateHospitalizationInput{CageID: second.CageID})
		if err != nil {
			return apperrors.Wrap(err, "failed to update hospitalization cage")
		}
		updatedSecond, err := s.hospRepo.Update(txCtx, clinicID, second.ID,
			UpdateHospitalizationInput{CageID: first.CageID})
		if err != nil {
			return apperrors.Wrap(err, "failed to update hospitalization cage")
		}
		if first.ID == idA {
			outA, outB = updatedFirst, updatedSecond
		} else {
			outA, outB = updatedSecond, updatedFirst
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	slog.InfoContext(ctx, "hospitalization cages swapped",
		slog.Uint64("hospitalization_id_a", idA),
		slog.Uint64("hospitalization_id_b", idB),
		slog.Uint64("clinic_id", clinicID))
	return outA, outB, nil
}
