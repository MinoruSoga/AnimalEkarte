package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// EMR-449 A01 fixture F（merchandise_items）。実 DB / trigger / 実行は NOT_RUN
// （Lead 指示）。service/repo/handler の unit レベルで表現できる範囲を検査する。

type merchandiseBillDiscountAuditRecorder struct {
	entry  *AuditEntry
	failFn error
}

func (r *merchandiseBillDiscountAuditRecorder) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	if r.failFn != nil {
		return r.failFn
	}
	r.entry = entry
	return nil
}

// rollbackTrackingTransactor は fn の error を rollback 扱いとして記録する。
type rollbackTrackingTransactor struct {
	rolledBack bool
	committed  bool
}

func (m *rollbackTrackingTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		m.rolledBack = true
		return err
	}
	m.committed = true
	return nil
}

func merchandiseTestActorID() *uint64 { id := uint64(12); return &id }

// --- POST: 省略 → 明示 false 保存・create 監査・実保存版 ---

func TestMerchandiseItemService_Create_BillDiscount(t *testing.T) {
	tests := []struct {
		name       string
		input      *bool
		wantStored bool
	}{
		{name: "omitted saves explicit false", input: nil, wantStored: false},
		{name: "explicit true", input: boolPtrForBillDiscountTest(true), wantStored: true},
		{name: "explicit false", input: boolPtrForBillDiscountTest(false), wantStored: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *model.MerchandiseItem
			repo := &mockMerchandiseItemRepository{
				createFn: func(_ context.Context, item *model.MerchandiseItem) error {
					created = item
					item.ID = 5
					return nil
				},
				// 補償 UPDATE 後の同 TX 再読取: trigger が版を進めた実保存版を返す。
				findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
					return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 2, IsBillDiscountExcluded: boolPtrForBillDiscountTest(tt.wantStored)}, nil
				},
			}
			audit := &merchandiseBillDiscountAuditRecorder{}
			svc := NewMerchandiseItemServiceWithAudit(repo, &rollbackTrackingTransactor{}, audit)
			got, err := svc.Create(t.Context(), 7, &CreateMerchandiseItemInput{
				Name: "テスト品", Category: "goods", TaxType: "excluded",
				IsBillDiscountExcluded: tt.input, ActorID: merchandiseTestActorID(),
			})
			require.NoError(t, err)
			require.NotNil(t, created)
			require.NotNil(t, created.IsBillDiscountExcluded, "create must not leave the column to the DB default")
			assert.Equal(t, tt.wantStored, *created.IsBillDiscountExcluded)
			// A.4: create 補償後の実保存版を返す（初期1固定や再読取前の値ではない）。
			assert.Equal(t, int64(2), got.Version)
			// A.5: create 監査は before 行なし（BeforeExists=false）。
			require.NotNil(t, audit.entry)
			assert.Nil(t, audit.entry.OldValue)
			assert.Equal(t, map[string]any{colIsBillDiscountExcluded: tt.wantStored}, audit.entry.NewValue)
			assert.Equal(t, map[string]any{"field": colIsBillDiscountExcluded}, audit.entry.Metadata)
			assert.Equal(t, merchandiseItemAuditActionCreate, audit.entry.Action)
			assert.Equal(t, merchandiseItemAuditResource, audit.entry.Resource)
		})
	}
}

// --- 監査依存の fail-closed（旧 constructor・書込0） ---

func TestMerchandiseItemService_Create_MissingAuditFailsClosed(t *testing.T) {
	created := false
	repo := &mockMerchandiseItemRepository{
		createFn: func(context.Context, *model.MerchandiseItem) error { created = true; return nil },
	}
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	_, err := svc.Create(t.Context(), 7, &CreateMerchandiseItemInput{Name: "x", Category: "goods", TaxType: "excluded", ActorID: merchandiseTestActorID()})
	require.Error(t, err)
	assert.False(t, created, "zero writes without the audit dependency")
}

// --- PATCH: 省略は既存 NULL/false/true を保持し監査しない ---

func TestMerchandiseItemService_Update_FieldOmittedKeepsStoredValue(t *testing.T) {
	audit := &merchandiseBillDiscountAuditRecorder{}
	repo := &mockMerchandiseItemRepository{
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			assert.Nil(t, cmd.IsBillDiscountExcluded, "omitted field must not enter the update map")
			assert.NotNil(t, cmd.Version)
			assert.Equal(t, int64(7), *cmd.Version)
			return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 8}, nil
		},
	}
	// 旧 constructor（監査なし）でも新設定を含まない PATCH は従来どおり成功する。
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	name := "改名"
	got, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{Name: &name, Version: int64PtrForMerchandiseTest(7)})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, audit.entry, "no audit when the field is not sent")
}

// --- PATCH: 明示 true/false は必ず SET（false をゼロ値 skip しない）+ 同 TX 監査 ---

func TestMerchandiseItemService_Update_ExplicitValueWritesAndAudits(t *testing.T) {
	tests := []struct {
		name        string
		sent        bool
		beforeValue *bool // nil = 既存行の raw NULL
		afterValue  bool
	}{
		{name: "raw NULL to true", sent: true, beforeValue: nil, afterValue: true},
		{name: "true to false", sent: false, beforeValue: boolPtrForBillDiscountTest(true), afterValue: false},
		{name: "false to true", sent: true, beforeValue: boolPtrForBillDiscountTest(false), afterValue: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := &merchandiseBillDiscountAuditRecorder{}
			repo := &mockMerchandiseItemRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
					return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 7, IsBillDiscountExcluded: tt.beforeValue}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
					require.NotNil(t, cmd.IsBillDiscountExcluded, "explicit value must be SET even when false")
					assert.Equal(t, tt.sent, *cmd.IsBillDiscountExcluded)
					require.NotNil(t, cmd.Version)
					assert.Equal(t, int64(7), *cmd.Version, "CAS version reaches the repository")
					return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 8, IsBillDiscountExcluded: boolPtrForBillDiscountTest(tt.afterValue)}, nil
				},
			}
			tx := &rollbackTrackingTransactor{}
			svc := NewMerchandiseItemServiceWithAudit(repo, tx, audit)
			got, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{
				IsBillDiscountExcluded: boolPtrForBillDiscountTest(tt.sent),
				Version:                int64PtrForMerchandiseTest(7),
				ActorID:                merchandiseTestActorID(),
			})
			require.NoError(t, err)
			assert.True(t, tx.committed, "business write and audit share one transaction")
			require.NotNil(t, audit.entry)
			// raw NULL の before は create と混同しない（OldValue に field: null を記録）。
			wantBefore := map[string]any{colIsBillDiscountExcluded: nil}
			if tt.beforeValue != nil {
				wantBefore = map[string]any{colIsBillDiscountExcluded: *tt.beforeValue}
			}
			assert.Equal(t, wantBefore, audit.entry.OldValue)
			assert.Equal(t, map[string]any{colIsBillDiscountExcluded: tt.afterValue}, audit.entry.NewValue)
			assert.Equal(t, merchandiseItemAuditActionUpdate, audit.entry.Action)
			assert.Equal(t, int64(8), got.Version)
		})
	}
}

// --- PATCH: 新設定を含む PATCH は編集開始 version 必須 ---

func TestMerchandiseItemService_Update_FieldSentRequiresVersion(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			t.Fatal("mutation must not run")
			return nil, nil
		},
	}
	svc := NewMerchandiseItemServiceWithAudit(repo, &rollbackTrackingTransactor{}, &merchandiseBillDiscountAuditRecorder{})
	_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{IsBillDiscountExcluded: boolPtrForBillDiscountTest(true), ActorID: merchandiseTestActorID()})
	require.Error(t, err)
}

// --- PATCH: version のみ送信（業務 field なし）は 400・書込0 ---

func TestMerchandiseItemService_Update_VersionOnlyIsInvalid(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			t.Fatal("mutation must not run")
			return nil, nil
		},
	}
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{Version: int64PtrForMerchandiseTest(7)})
	require.Error(t, err)
}

// --- PATCH: 送信 version が 0/負は 400・書込0 ---

func TestMerchandiseItemService_Update_SentVersionMustBePositive(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			t.Fatal("mutation must not run")
			return nil, nil
		},
	}
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	for _, v := range []int64{0, -1} {
		_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{Version: int64PtrForMerchandiseTest(v)})
		require.Error(t, err, "version %d must be rejected", v)
	}
}

// --- PATCH: 版不一致は既存の競合 error（409）・監査なし ---

func TestMerchandiseItemService_Update_VersionConflictDoesNotAudit(t *testing.T) {
	audit := &merchandiseBillDiscountAuditRecorder{}
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 9}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			return nil, apperrors.WrapConflict("stale master")
		},
	}
	tx := &rollbackTrackingTransactor{}
	svc := NewMerchandiseItemServiceWithAudit(repo, tx, audit)
	got, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{
		IsBillDiscountExcluded: boolPtrForBillDiscountTest(true),
		Version:                int64PtrForMerchandiseTest(7),
		ActorID:                merchandiseTestActorID(),
	})
	require.Error(t, err)
	require.True(t, apperrors.IsConflict(err))
	require.Nil(t, got)
	assert.True(t, tx.rolledBack, "conflict rolls the transaction back")
	assert.Nil(t, audit.entry, "conflict must not audit")
}

// --- 監査失敗は全 rollback ---

func TestMerchandiseItemService_Update_AuditFailureRollsBack(t *testing.T) {
	sentinel := errors.New("audit sink down")
	audit := &merchandiseBillDiscountAuditRecorder{failFn: sentinel}
	updated := false
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 7}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			updated = true
			return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 8}, nil
		},
	}
	tx := &rollbackTrackingTransactor{}
	svc := NewMerchandiseItemServiceWithAudit(repo, tx, audit)
	_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{
		IsBillDiscountExcluded: boolPtrForBillDiscountTest(true),
		Version:                int64PtrForMerchandiseTest(7),
		ActorID:                merchandiseTestActorID(),
	})
	require.ErrorIs(t, err, sentinel)
	assert.True(t, updated, "business write ran inside the transaction")
	assert.True(t, tx.rolledBack, "audit failure rolls the whole write back")
}

func TestMerchandiseItemService_Create_AuditFailureRollsBack(t *testing.T) {
	sentinel := errors.New("audit sink down")
	audit := &merchandiseBillDiscountAuditRecorder{failFn: sentinel}
	repo := &mockMerchandiseItemRepository{
		createFn: func(_ context.Context, item *model.MerchandiseItem) error { item.ID = 5; return nil },
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 1}, nil
		},
	}
	tx := &rollbackTrackingTransactor{}
	svc := NewMerchandiseItemServiceWithAudit(repo, tx, audit)
	_, err := svc.Create(t.Context(), 7, &CreateMerchandiseItemInput{Name: "x", Category: "goods", TaxType: "excluded", ActorID: merchandiseTestActorID()})
	require.ErrorIs(t, err, sentinel)
	assert.True(t, tx.rolledBack)
}

// --- 新設定の PATCH で監査依存欠落は fail-closed・書込0 ---

func TestMerchandiseItemService_Update_MissingAuditFailsClosed(t *testing.T) {
	updated := false
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 7}, nil
		},
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			updated = true
			return nil, nil
		},
	}
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{
		IsBillDiscountExcluded: boolPtrForBillDiscountTest(true),
		Version:                int64PtrForMerchandiseTest(7),
		ActorID:                merchandiseTestActorID(),
	})
	require.Error(t, err)
	assert.False(t, updated, "zero writes without the audit dependency")
}

// --- GET: 保存値 raw NULL は null のまま返し、false へ書換えない ---

func TestMerchandiseItemResponse_PreservesNullAndVersion(t *testing.T) {
	raw, err := json.Marshal(mustMerchandiseResponse(t, &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 17}))
	require.NoError(t, err)
	assert.True(t, bytes.Contains(raw, []byte(`"is_bill_discount_excluded":null`)), "stored NULL stays null: %s", raw)
	assert.True(t, bytes.Contains(raw, []byte(`"version":17`)), "stored version is returned: %s", raw)

	raw, err = json.Marshal(mustMerchandiseResponse(t, &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 17, IsBillDiscountExcluded: boolPtrForBillDiscountTest(true)}))
	require.NoError(t, err)
	assert.True(t, bytes.Contains(raw, []byte(`"is_bill_discount_excluded":true`)), "%s", raw)
}

// --- read 失敗は error として伝播し、null の成功応答へ変換しない ---

func TestMerchandiseItemService_ReaderErrorIsNotConvertedToNull(t *testing.T) {
	sentinel := errors.New("column missing")
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.MerchandiseItem, error) { return nil, sentinel },
	}
	svc := NewMerchandiseItemService(repo, &mockMerchandiseItemTransactor{})
	_, err := svc.GetByID(t.Context(), 7, 5)
	require.ErrorIs(t, err, sentinel)
}

// --- repository: CAS SQL と outcome 分類（DryRun・実 DB なし） ---

func TestMerchandiseItemRepository_CASSQLAndOutcomes(t *testing.T) {
	for _, scenario := range []string{"updated", "stale", "missing", "no-cas"} {
		t.Run(scenario, func(t *testing.T) {
			db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			conn, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			var updateSQL string
			var updateVars []any
			reads := 0
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("capture_cas", func(tx *gorm.DB) {
				updateSQL = tx.Statement.SQL.String()
				updateVars = append([]any(nil), tx.Statement.Vars...)
				if scenario == "updated" || scenario == "no-cas" {
					tx.RowsAffected = 1
				} else {
					tx.RowsAffected = 0
				}
			}))
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("cas_result", func(tx *gorm.DB) {
				reads++
				require.Contains(t, tx.Statement.SQL.String(), "clinic_id")
				if scenario == "missing" {
					tx.AddError(gorm.ErrRecordNotFound)
					return
				}
				row := tx.Statement.Dest.(*model.MerchandiseItem)
				row.ID = 17
				row.ClinicID = 2
				row.Version = 8
				tx.RowsAffected = 1
			}))
			repo := &merchandiseItemRepository{db: nil}
			ctx := persistence.WithTxValue(t.Context(), db)
			fields := map[string]any{colIsBillDiscountExcluded: false}
			var version *int64
			if scenario != "no-cas" {
				version = int64PtrForMerchandiseTest(7)
			}
			err = repo.update(ctx, 2, 17, fields, version)
			require.Contains(t, updateSQL, "id = $")
			require.Contains(t, updateSQL, "clinic_id")
			setClause := strings.Split(updateSQL, " WHERE ")[0]
			// A-6: GORM は SET 値を placeholder へ bind する。SQL は列名 + $n、
			// literal false は Statement.Vars に入る。
			require.Contains(t, setClause, `"is_bill_discount_excluded"=$`, "explicit false must be SET via a bind placeholder")
			require.Contains(t, updateVars, false, "explicit false must reach Statement.Vars")
			require.NotContains(t, setClause, `"version"`, "database trigger owns advancement")
			if scenario == "no-cas" {
				require.NotContains(t, updateSQL, "version = $", "legacy PATCH without version has no CAS")
				require.NoError(t, err)
				assert.Equal(t, 0, reads)
				return
			}
			require.Contains(t, updateSQL, "version = $")
			require.Contains(t, updateVars, int64(7))
			require.Contains(t, updateVars, uint64(2))
			switch scenario {
			case "updated":
				require.NoError(t, err)
				assert.Equal(t, 0, reads)
			case "stale":
				require.Error(t, err)
				require.True(t, apperrors.IsConflict(err))
				assert.Equal(t, 1, reads)
			case "missing":
				require.Error(t, err)
				require.True(t, apperrors.IsNotFound(err))
				assert.Equal(t, 1, reads)
			}
		})
	}
}

// --- repository: ambient-TX FindByID のロック強度（A-2/D2） ---

func TestMerchandiseItemRepository_FindByIDLockStrength(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	var queries []string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("capture_find_lock", func(tx *gorm.DB) {
		queries = append(queries, tx.Statement.SQL.String())
	}))
	repo := &merchandiseItemRepository{db: db}
	txCtx := persistence.WithTxValue(t.Context(), db)
	// 参照用途（campaign attachment serialization）: 既定は FOR SHARE。
	_, _ = repo.FindByID(txCtx, 2, 17)
	// 更新 TX の前値読取: FOR UPDATE で先に排他ロックする。
	_, _ = repo.FindByID(withMerchandiseItemForUpdate(txCtx), 2, 17)
	// ambient tx なし: ロック句なし。
	_, _ = repo.FindByID(t.Context(), 2, 17)
	require.Len(t, queries, 3)
	require.Contains(t, queries[0], "FOR SHARE")
	require.Contains(t, queries[1], "FOR UPDATE")
	require.NotContains(t, queries[2], "FOR ")
}

// --- update TX: FOR UPDATE の前値読取が UPDATE・監査より先（A-2/D2） ---

func TestMerchandiseItemService_Update_ForUpdateReadPrecedesWrite(t *testing.T) {
	var events []string
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(ctx context.Context, _, id uint64) (*model.MerchandiseItem, error) {
			if locked, _ := ctx.Value(merchandiseItemForUpdateKey{}).(bool); locked {
				events = append(events, "forUpdateRead")
			} else {
				events = append(events, "read")
			}
			return &model.MerchandiseItem{ID: id, ClinicID: 7, Version: 7}, nil
		},
		updateFieldsFn: func(context.Context, uint64, uint64, UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
			events = append(events, "update")
			return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 8}, nil
		},
	}
	svc := NewMerchandiseItemServiceWithAudit(repo, &rollbackTrackingTransactor{}, &merchandiseBillDiscountAuditRecorder{})
	_, err := svc.Update(t.Context(), 7, 5, &UpdateMerchandiseItemInput{
		IsBillDiscountExcluded: boolPtrForBillDiscountTest(true),
		Version:                int64PtrForMerchandiseTest(7),
		ActorID:                merchandiseTestActorID(),
	})
	require.NoError(t, err)
	// 存在確認 read → FOR UPDATE 前値読取 → UPDATE。排他ロックは UPDATE より先。
	require.Equal(t, []string{"read", "forUpdateRead", "update"}, events)
}

// --- handler: PATCH version 境界（null/0/負/文字列/小数は 400） ---

func TestMerchandiseItemHandler_PatchVersionBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, version := range []string{`,"version":null`, `,"version":0`, `,"version":-1`, `,"version":"7"`, `,"version":1.5`} {
		t.Run(version, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) { setClinicID(c); setStaffIDForMerchandiseTest(c) })
			r.PATCH("/:id", NewHandler(nil, &mockMerchandiseItemService{updateFn: func(context.Context, uint64, uint64, *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
				t.Fatal("mutation must not run")
				return nil, nil
			}}, allowInventoryTestPermission).UpdateMerchandiseItem)
			req := httptest.NewRequest(http.MethodPatch, "/5", strings.NewReader(`{"name":"changed"`+version+`}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// --- handler: actor は trusted request context 由来（body に actor を受け取らない） ---

func TestMerchandiseItemHandler_ActorFromRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { setClinicID(c); setStaffIDForMerchandiseTest(c) })
	r.PATCH("/:id", NewHandler(nil, &mockMerchandiseItemService{updateFn: func(_ context.Context, _, _ uint64, in *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
		require.NotNil(t, in.ActorID)
		assert.Equal(t, uint64(7), *in.ActorID)
		return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 8}, nil
	}}, allowInventoryTestPermission).UpdateMerchandiseItem)
	req := httptest.NewRequest(http.MethodPatch, "/5", strings.NewReader(`{"name":"changed","version":7}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMerchandiseItemHandler_CreateActorFromRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { setClinicID(c); setStaffIDForMerchandiseTest(c) })
	r.POST("/", NewHandler(nil, &mockMerchandiseItemService{createFn: func(_ context.Context, _ uint64, in *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
		require.NotNil(t, in.ActorID)
		assert.Equal(t, uint64(7), *in.ActorID)
		return &model.MerchandiseItem{ID: 5, ClinicID: 7, Version: 1}, nil
	}}, allowInventoryTestPermission).CreateMerchandiseItem)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"new","category":"goods","tax_type":"excluded"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, bytes.Contains(w.Body.Bytes(), []byte(`"is_bill_discount_excluded":`)), "response always contains the key: %s", w.Body.String())
}

// --- handler: POST の null / 非 boolean は 400 ---

func TestMerchandiseItemHandler_CreateBillDiscountWireRejections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, payload := range []string{`{"name":"x","category":"goods","tax_type":"excluded","is_bill_discount_excluded":null}`, `{"name":"x","category":"goods","tax_type":"excluded","is_bill_discount_excluded":"true"}`} {
		t.Run(payload, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) { setClinicID(c); setStaffIDForMerchandiseTest(c) })
			r.POST("/", NewHandler(nil, &mockMerchandiseItemService{createFn: func(context.Context, uint64, *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
				t.Fatal("mutation must not run")
				return nil, nil
			}}, allowInventoryTestPermission).CreateMerchandiseItem)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// --- handler: 版不一致は 409 ---

func TestMerchandiseItemHandler_UpdateVersionConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { setClinicID(c); setStaffIDForMerchandiseTest(c) })
	r.PATCH("/:id", NewHandler(nil, &mockMerchandiseItemService{updateFn: func(_ context.Context, _, _ uint64, in *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
		if in.Version == nil || *in.Version != 7 {
			t.Fatal("expected version changed")
		}
		return nil, apperrors.WrapConflict("stale master")
	}}, allowInventoryTestPermission).UpdateMerchandiseItem)
	req := httptest.NewRequest(http.MethodPatch, "/5", strings.NewReader(`{"name":"changed","version":7}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- helpers ---

func setStaffIDForMerchandiseTest(c *gin.Context) {
	c.Set("user_id", "7")
}

func int64PtrForMerchandiseTest(v int64) *int64 { return &v }

func mustMerchandiseResponse(t *testing.T, m *model.MerchandiseItem) merchandiseItemResponse {
	t.Helper()
	resp, err := toMerchandiseItemResponse(m)
	if err != nil {
		t.Fatalf("response conversion failed: %v", err)
	}
	return resp
}
