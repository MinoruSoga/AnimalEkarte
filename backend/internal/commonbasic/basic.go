package commonbasic

import "github.com/animal-ekarte/backend/internal/model"

// Ptrs はスライス要素そのものへのポインタ列を返す。Hydrate* の結果が元のスライスの要素へ反映される。
func Ptrs[T any](items []T) []*T {
	out := make([]*T, len(items))
	for i := range items {
		out[i] = &items[i]
	}
	return out
}

// OwnerOf は hydrate 済みの owner から現行 basic revision を返す。未 hydrate は R-1 の unresolved（422）で、
// 空値や廃止列へのフォールバックはしない。返す値は読み取り専用として扱うこと。
func OwnerOf(owner *model.Owner) (*model.CommonOwnerBasicRevision, error) {
	if owner == nil {
		return nil, unresolved("owner is nil")
	}
	if owner.Basic == nil {
		return nil, unresolved("owner %d/%d: basic information is not hydrated", owner.ClinicID, owner.ID)
	}
	return owner.Basic, nil
}

// PetOf は OwnerOf の pet 版。
func PetOf(pet *model.Pet) (*model.CommonPetBasicRevision, error) {
	if pet == nil {
		return nil, unresolved("pet is nil")
	}
	if pet.Basic == nil {
		return nil, unresolved("pet %d/%d: basic information is not hydrated", pet.ClinicID, pet.ID)
	}
	return pet.Basic, nil
}

// MustOwnerBasic は戻り値にエラーを持てない固定シグネチャの変換関数向け（R-NIL）。未 hydrate は panic する。
func MustOwnerBasic(owner *model.Owner) *model.CommonOwnerBasicRevision {
	basic, err := OwnerOf(owner)
	if err != nil {
		panic(err)
	}
	return basic
}

// MustPetBasic は MustOwnerBasic の pet 版。
func MustPetBasic(pet *model.Pet) *model.CommonPetBasicRevision {
	basic, err := PetOf(pet)
	if err != nil {
		panic(err)
	}
	return basic
}
