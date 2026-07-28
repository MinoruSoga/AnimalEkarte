package reservation

// reservation_type_preload_clinic_isolation_test.go
// P0 lint follow-up（read 監査と同じ bar）: reservation_type の Group/Parent/Children Preload の
// clinic_id 述語が実際にクロステナント混入を防ぐことを runtime で実証する。
// これらは b3638d5e が見逃し、P0 静的 lint(preload_clinic_scope_lint_test.go)が捕捉・修正した
// 11 サイトのうち reservation_type 系。静的 lint は述語の存在を、本テストは隔離の動作を保証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupReservationTypePreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// setupTestDB が reservation_type_category ENUM を DROP CASCADE するため category 列が消える。
	// AutoMigrate で列・テーブルを再整備する。
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationTypeGroup{}, &model.ReservationType{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	return db
}

// TestReservationTypeRepository_GroupParentChildren_CrossClinicPreloadIsolation は
// Group/Parent/Children の clinic_id 述語が別クリニックのマスタ混入を防ぐことを動作で証明する。
// legit ケース(同一clinicはPreloadされる)と cross ケース(別clinicは nil)の対で、
// 述語が「効いている」かつ「正規データを壊していない」ことを同時に示す（anti-vacuous）。
func TestReservationTypeRepository_GroupParentChildren_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupReservationTypePreloadTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	groupA := makeReservationTypeGroup(t, db, clinicA, "医院Aのグループ")
	groupB := makeReservationTypeGroup(t, db, clinicB, "医院Bのグループ")

	// (legit) 同一 clinic の Group は Preload される（述語が正規データを壊さない）
	legit := makeReservationTypeLinked(t, db, clinicA, "正規区分", &groupA.ID, nil)
	gotLegit, err := repo.FindByID(ctx, clinicA, legit.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.Group, "同一クリニックの Group は Preload されるべき")
	assert.Equal(t, groupA.ID, gotLegit.Group.ID)

	// (i) Group: 別クリニックの group_id を植え付けた区分 → Group は混入しない
	crossGroup := makeReservationTypeLinked(t, db, clinicA, "越境Group区分", &groupB.ID, nil)
	gotCrossGroup, err := repo.FindByID(ctx, clinicA, crossGroup.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCrossGroup.Group, "別クリニックの Group マスタが混入してはならない")

	// (ii) Parent: 別クリニックの reservation_type を親に植え付け → Parent は混入しない
	parentB := makeReservationTypeLinked(t, db, clinicB, "医院Bの親区分", nil, nil)
	crossParent := makeReservationTypeLinked(t, db, clinicA, "越境Parent区分", nil, &parentB.ID)
	gotCrossParent, err := repo.FindByID(ctx, clinicA, crossParent.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCrossParent.Parent, "別クリニックの Parent 区分が混入してはならない")

	// (iii) Children: 別クリニックの子(parent_id=A親)は Children に含まれない
	parentA := makeReservationTypeLinked(t, db, clinicA, "医院Aの親区分", nil, nil)
	childA := makeReservationTypeLinked(t, db, clinicA, "医院Aの子区分", nil, &parentA.ID)
	_ = makeReservationTypeLinked(t, db, clinicB, "越境子区分", nil, &parentA.ID) // B の子が A 親を参照
	gotWithChildren, err := repo.FindByIDWithChildren(ctx, clinicA, parentA.ID)
	require.NoError(t, err)
	require.Len(t, gotWithChildren.Children, 1, "Children は同一クリニックの子のみであるべき")
	assert.Equal(t, childA.ID, gotWithChildren.Children[0].ID)
}
