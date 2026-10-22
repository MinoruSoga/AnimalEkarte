package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

// POST/PATCH wire decode: 省略・true・false・null・非 boolean の契約（EMR-447 A.2）。
func TestMedicineBillDiscount_RequestDecode(t *testing.T) {
	cases := []struct {
		name  string
		body  string
		want  *bool
		isErr bool
	}{
		{"omitted", `{"name":"x"}`, nil, false},
		{"explicit true", `{"name":"x","is_bill_discount_excluded":true}`, boolPtrForBillDiscount(true), false},
		{"explicit false", `{"name":"x","is_bill_discount_excluded":false}`, boolPtrForBillDiscount(false), false},
		{"null", `{"name":"x","is_bill_discount_excluded":null}`, nil, true},
		{"non boolean", `{"name":"x","is_bill_discount_excluded":"yes"}`, nil, true},
	}
	t.Run("create", func(t *testing.T) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				var req createMedicineRequest
				err := json.Unmarshal([]byte(tt.body), &req)
				if tt.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				input, err := req.toServiceInput()
				require.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, input.IsBillDiscountExcluded)
				} else {
					require.NotNil(t, input.IsBillDiscountExcluded)
					assert.Equal(t, *tt.want, *input.IsBillDiscountExcluded)
				}
			})
		}
	})
	t.Run("update", func(t *testing.T) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				var req updateMedicineRequest
				err := json.Unmarshal([]byte(tt.body), &req)
				if tt.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				input, err := req.toServiceInput()
				require.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, input.IsBillDiscountExcluded)
				} else {
					require.NotNil(t, input.IsBillDiscountExcluded)
					assert.Equal(t, *tt.want, *input.IsBillDiscountExcluded)
				}
			})
		}
	})
	// version は任意送信。送信値はそのまま service input へ渡る（service 層で >0 検証）。
	t.Run("update version", func(t *testing.T) {
		var req updateMedicineRequest
		require.NoError(t, json.Unmarshal([]byte(`{"version":7,"is_bill_discount_excluded":true}`), &req))
		input, err := req.toServiceInput()
		require.NoError(t, err)
		require.NotNil(t, input.Version)
		assert.Equal(t, int64(7), *input.Version)
	})
}

// POST: 省略でも service が false を明示保存し、在庫連携の補償 UPDATE 後の
// 実保存版を返し、前値なし監査を同じ tx で記録する（A.2/A.4/A.5）。
func TestMedicineService_BillDiscountCreate(t *testing.T) {
	actor := uint64(5)
	for _, tt := range []struct {
		name  string
		input *bool
		want  bool
	}{
		{"omitted saves explicit false", nil, false},
		{"explicit true", boolPtrForBillDiscount(true), true},
		{"explicit false", boolPtrForBillDiscount(false), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stored *bool
			audit := &insuranceAuditRecorder{}
			repo := &mockMedicineRepository{
				createFn: func(_ context.Context, m *model.Medicine) error {
					m.ID = 9
					stored = m.IsBillDiscountExcluded
					return nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
					// 在庫連携の補償 UPDATE 後に trigger が版を進めた状態を模す。
					return &model.Medicine{ID: 9, ClinicID: 7, Version: 3, IsBillDiscountExcluded: stored}, nil
				},
			}
			inv := &mockInventoryRepository{
				createFn: func(_ context.Context, _ uint64, item *model.InventoryItem) error {
					item.ID = 9001
					return nil
				},
			}
			svc := NewMedicineServiceWithAudit(repo, inv, masterServiceJournalTx{rollback: func() { stored = nil }}, audit)
			got, err := svc.Create(context.Background(), 7, &CreateMedicineInput{Name: "x", ActorID: &actor, IsBillDiscountExcluded: tt.input})
			require.NoError(t, err)
			require.NotNil(t, stored, "create must always write the column (never rely on DB default)")
			assert.Equal(t, tt.want, *stored)
			assert.Equal(t, int64(3), got.Version, "response must carry the stored version after the compensating inventory link update")
			require.NotNil(t, audit.entry)
			assert.True(t, audit.inTx, "audit must share the business transaction")
			assert.Equal(t, "medicine.create", audit.entry.Action)
			assert.Equal(t, model.AuditResourceMedicine, audit.entry.Resource)
			assert.Equal(t, map[string]any{"field": colIsBillDiscountExcluded}, audit.entry.Metadata)
			assert.Nil(t, audit.entry.OldValue, "create has no previous row")
			assert.Equal(t, map[string]any{colIsBillDiscountExcluded: tt.want}, audit.entry.NewValue)
			require.NotNil(t, audit.entry.ActorID)
			assert.Equal(t, actor, *audit.entry.ActorID)
		})
	}
}

// POST は監査依存欠落・監査失敗で全体 rollback（fail-closed、A.5）。
func TestMedicineService_BillDiscountCreateRollback(t *testing.T) {
	actor := uint64(5)
	inv := &mockInventoryRepository{
		createFn: func(_ context.Context, _ uint64, item *model.InventoryItem) error {
			item.ID = 9001
			return nil
		},
	}
	t.Run("missing audit dependency", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 9; return nil },
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				return &model.Medicine{ID: 9}, nil
			},
		}
		svc := NewMedicineServiceWithAudit(repo, inv, masterServiceJournalTx{rollback: func() {}}, nil)
		_, err := svc.Create(context.Background(), 7, &CreateMedicineInput{Name: "x", ActorID: &actor})
		require.Error(t, err)
	})
	t.Run("audit failure rolls back", func(t *testing.T) {
		var stored *bool
		audit := &insuranceAuditRecorder{err: errors.New("audit unavailable")}
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error {
				m.ID = 9
				stored = m.IsBillDiscountExcluded
				return nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				return &model.Medicine{ID: 9, IsBillDiscountExcluded: stored}, nil
			},
		}
		svc := NewMedicineServiceWithAudit(repo, inv, masterServiceJournalTx{rollback: func() { stored = nil }}, audit)
		_, err := svc.Create(context.Background(), 7, &CreateMedicineInput{Name: "x", ActorID: &actor})
		require.Error(t, err)
		assert.Nil(t, stored, "audit failure must roll back the business write")
	})
}

// PATCH: 省略=変更なし（保存 NULL/false/true を保持）、明示送信=書込+前後値監査、
// 新設定送信時は version 必須、旧 PATCH の任意 version は必ず CAS 照合される（A.2/A.4/A.5）。
func TestMedicineService_BillDiscountUpdate(t *testing.T) {
	actor := uint64(5)
	name := "renamed"
	for _, tt := range []struct {
		name           string
		input          *UpdateMedicineInput
		wantUpdate     bool
		wantVersionNil bool
		wantVersion    *int64
		wantErr        bool
		wantAudit      bool
	}{
		{"field omitted keeps old behaviour", &UpdateMedicineInput{Name: &name, ActorID: &actor}, true, true, nil, false, false},
		{"field omitted but version sent is still checked", &UpdateMedicineInput{Name: &name, Version: int64PtrForCAS(7), ActorID: &actor}, true, false, int64PtrForCAS(7), false, false},
		{"field sent without version is rejected", &UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(true), ActorID: &actor}, false, true, nil, true, false},
		{"field sent with non-positive version is rejected", &UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(true), Version: int64PtrForCAS(0), ActorID: &actor}, false, false, nil, true, false},
		{"field true written through map and audited", &UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(true), Version: int64PtrForCAS(7), ActorID: &actor}, true, false, int64PtrForCAS(7), false, true},
		{"field false written through map and audited", &UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(false), Version: int64PtrForCAS(7), ActorID: &actor}, true, false, int64PtrForCAS(7), false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			audit := &insuranceAuditRecorder{}
			var gotCmd *UpdateMedicineInput
			beforeVal := boolPtrForBillDiscount(true)
			repo := &mockMedicineRepository{
				findByIDFn: func(_ context.Context, cid, id uint64) (*model.Medicine, error) {
					if cid != 7 || id != 9 {
						t.Fatal("clinic scope changed")
					}
					return &model.Medicine{ID: 9, ClinicID: 7, IsBillDiscountExcluded: beforeVal}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMedicineInput) (*model.Medicine, error) {
					gotCmd = &cmd
					return &model.Medicine{ID: 9, ClinicID: 7, IsBillDiscountExcluded: cmd.IsBillDiscountExcluded}, nil
				},
			}
			svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, masterServiceJournalTx{rollback: func() {}}, audit)
			got, err := svc.Update(context.Background(), 7, 9, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				assert.Nil(t, gotCmd, "no write may occur on validation error")
				assert.Nil(t, audit.entry)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, gotCmd)
			if tt.wantVersionNil {
				assert.Nil(t, gotCmd.Version)
			} else {
				require.NotNil(t, gotCmd.Version)
				assert.Equal(t, *tt.wantVersion, *gotCmd.Version)
			}
			if !tt.wantAudit {
				assert.Nil(t, audit.entry, "PATCH without the field must not log bill-discount audit")
				return
			}
			require.NotNil(t, audit.entry)
			assert.True(t, audit.inTx)
			assert.Equal(t, "medicine.update", audit.entry.Action)
			assert.Equal(t, map[string]any{"field": colIsBillDiscountExcluded}, audit.entry.Metadata)
			assert.Equal(t, map[string]any{colIsBillDiscountExcluded: *beforeVal}, audit.entry.OldValue)
			assert.Equal(t, map[string]any{colIsBillDiscountExcluded: *tt.input.IsBillDiscountExcluded}, audit.entry.NewValue)
		})
	}
}

// 保存値が raw NULL（BeforeExists=true, Before=nil）でも before=null として監査できる。
func TestMedicineService_BillDiscountUpdateNullBefore(t *testing.T) {
	actor := uint64(5)
	audit := &insuranceAuditRecorder{}
	repo := &mockMedicineRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
			return &model.Medicine{ID: 9, ClinicID: 7, IsBillDiscountExcluded: nil}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMedicineInput) (*model.Medicine, error) {
			return &model.Medicine{ID: 9, ClinicID: 7, IsBillDiscountExcluded: cmd.IsBillDiscountExcluded}, nil
		},
	}
	svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, masterServiceJournalTx{rollback: func() {}}, audit)
	_, err := svc.Update(context.Background(), 7, 9, &UpdateMedicineInput{
		IsBillDiscountExcluded: boolPtrForBillDiscount(true), Version: int64PtrForCAS(3), ActorID: &actor})
	require.NoError(t, err)
	require.NotNil(t, audit.entry)
	assert.Equal(t, map[string]any{colIsBillDiscountExcluded: nil}, audit.entry.OldValue)
	assert.Equal(t, map[string]any{colIsBillDiscountExcluded: true}, audit.entry.NewValue)
}

// 版不一致（409 競合）と監査失敗は書込・監査ともに成功を返さない。
func TestMedicineService_BillDiscountUpdateRollback(t *testing.T) {
	actor := uint64(5)
	t.Run("stale conflict propagates and audits nothing", func(t *testing.T) {
		audit := &insuranceAuditRecorder{}
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
				return &model.Medicine{ID: 9, ClinicID: 7}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				return nil, apperrors.WrapConflict(errMsgMedicineVersionConflict)
			},
		}
		svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, masterServiceJournalTx{rollback: func() {}}, audit)
		_, err := svc.Update(context.Background(), 7, 9, &UpdateMedicineInput{
			IsBillDiscountExcluded: boolPtrForBillDiscount(true), Version: int64PtrForCAS(7), ActorID: &actor})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, audit.entry)
	})
	t.Run("audit failure rolls back", func(t *testing.T) {
		audit := &insuranceAuditRecorder{err: errors.New("audit unavailable")}
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
				return &model.Medicine{ID: 9, ClinicID: 7}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateMedicineInput) (*model.Medicine, error) {
				return &model.Medicine{ID: 9, ClinicID: 7, IsBillDiscountExcluded: cmd.IsBillDiscountExcluded}, nil
			},
		}
		svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, masterServiceJournalTx{rollback: func() {}}, audit)
		_, err := svc.Update(context.Background(), 7, 9, &UpdateMedicineInput{
			IsBillDiscountExcluded: boolPtrForBillDiscount(false), Version: int64PtrForCAS(7), ActorID: &actor})
		require.Error(t, err)
	})
}

// UPDATE map: 省略は列を出さず既存値を保持し、false も明示 SET される（A.2）。
func TestMedicineBuildUpdate_BillDiscount(t *testing.T) {
	fields := buildMedicineUpdate(&UpdateMedicineInput{})
	_, ok := fields[colIsBillDiscountExcluded]
	assert.False(t, ok, "omitted field must not touch the column")
	fields = buildMedicineUpdate(&UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(false)})
	assert.Equal(t, false, fields[colIsBillDiscountExcluded])
	fields = buildMedicineUpdate(&UpdateMedicineInput{IsBillDiscountExcluded: boolPtrForBillDiscount(true)})
	assert.Equal(t, true, fields[colIsBillDiscountExcluded])
}

// 応答は is_bill_discount_excluded（null をそのまま・omitempty なし）と version を必ず返す（A.3/A.4）。
func TestMedicineResponse_BillDiscountAndVersion(t *testing.T) {
	resp := toMedicineResponse(&model.Medicine{ID: 1, Version: 17})
	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"is_bill_discount_excluded":null`, "stored NULL must stay null (never rewritten to false)")
	assert.Contains(t, string(body), `"version":17`)
	stored := true
	resp = toMedicineResponse(&model.Medicine{ID: 1, Version: 3, IsBillDiscountExcluded: &stored})
	body, err = json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"is_bill_discount_excluded":true`)
}

// reader エラーは NULL へ変換せずエラーとして伝播する（A.3）。
func TestMedicineService_GetByIDReaderError(t *testing.T) {
	repo := &mockMedicineRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
			return nil, errors.New("column does not exist")
		},
	}
	svc := NewMedicineServiceWithAudit(repo, &mockInventoryRepository{}, masterServiceJournalTx{rollback: func() {}}, nil)
	got, err := svc.GetByID(context.Background(), 7, 9)
	require.Error(t, err)
	assert.Nil(t, got)
}

// repository の CAS: version は UPDATE の WHERE に入り SET には入らない。
// 0 行は「存在しない」と「版不一致」に区別する。実 PostgreSQL trigger/concurrency 対象外。
func TestMedicineRepository_BillDiscountCASSQLAndOutcomes(t *testing.T) {
	for _, tt := range []struct {
		name     string
		version  *int64
		scenario string
	}{
		{"version sent, updated", int64PtrForCAS(7), "updated"},
		{"version omitted keeps old update", nil, "updated"},
		{"stale conflict", int64PtrForCAS(7), "stale"},
		{"missing row stays not-found", int64PtrForCAS(7), "missing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			conn, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			var updateSQL string
			var updateVars []any
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("capture_cas", func(tx *gorm.DB) {
				updateSQL = tx.Statement.SQL.String()
				updateVars = append([]any(nil), tx.Statement.Vars...)
				if tt.scenario == "updated" {
					tx.RowsAffected = 1
				} else {
					tx.RowsAffected = 0
				}
			}))
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("cas_result", func(tx *gorm.DB) {
				require.Contains(t, tx.Statement.SQL.String(), "clinic_id")
				if tt.scenario == "missing" {
					tx.AddError(gorm.ErrRecordNotFound)
					return
				}
				if row, ok := tx.Statement.Dest.(*model.Medicine); ok {
					row.ID = 17
					row.ClinicID = 2
					row.Version = 8
					tx.RowsAffected = 1
				}
			}))
			repo := &medicineRepository{db: nil}
			ctx := persistence.WithTxValue(t.Context(), db)
			got, err := repo.Update(ctx, 2, 17, UpdateMedicineInput{
				Version: tt.version, IsBillDiscountExcluded: boolPtrForBillDiscount(true)})
			require.Contains(t, updateSQL, "id = $")
			require.Contains(t, updateSQL, "clinic_id")
			require.Contains(t, updateSQL, "is_bill_discount_excluded")
			assert.NotContains(t, updateSQL, `"version"=`, "version must never be written by the app (DB trigger advances it)")
			if tt.version == nil {
				assert.NotContains(t, updateSQL, "version = $", "omitted version must not add a CAS predicate")
			} else {
				require.Contains(t, updateSQL, "version = $")
				require.Contains(t, updateVars, int64(7))
			}
			require.Contains(t, updateVars, uint64(2))
			require.Contains(t, updateVars, uint64(17))
			switch tt.scenario {
			case "updated":
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, int64(8), got.Version, "response must carry the stored version re-read in the same tx")
			case "stale":
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err))
				assert.Nil(t, got)
			case "missing":
				require.Error(t, err)
				assert.True(t, apperrors.IsNotFound(err))
				assert.Nil(t, got)
			}
		})
	}
}

// handler が trusted actor（staff コンテキスト）を service input へ配線し、
// null/非 boolean を binding 層で 400 にする（clinic_id/actor はクライアントから取らない）。
func TestMedicineHandler_BillDiscountWiring(t *testing.T) {
	gin.SetMode(gin.TestMode)
	actorSeen := make(chan *uint64, 1)
	svc := &mockMedicineService{
		createFn: func(_ context.Context, _ uint64, input *CreateMedicineInput) (*model.Medicine, error) {
			actorSeen <- input.ActorID
			return &model.Medicine{ID: 9, IsBillDiscountExcluded: input.IsBillDiscountExcluded}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
			actorSeen <- input.ActorID
			return &model.Medicine{ID: 9, IsBillDiscountExcluded: input.IsBillDiscountExcluded}, nil
		},
	}
	h := NewMedicineHandler(svc)
	r := gin.New()
	r.POST("/t", func(c *gin.Context) { setClinicID(c); setStaffID(c) }, h.CreateMedicine)
	r.PATCH("/t/:id", func(c *gin.Context) { setClinicID(c); setStaffID(c) }, h.UpdateMedicine)

	t.Run("POST sends trusted actor and field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewBufferString(`{"name":"x","is_bill_discount_excluded":true}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		assert.NotNil(t, <-actorSeen)
	})
	t.Run("PATCH sends trusted actor and version", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/t/9", bytes.NewBufferString(`{"is_bill_discount_excluded":false,"version":3}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, <-actorSeen)
	})
	for _, tt := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"POST null", http.MethodPost, "/t", `{"name":"x","is_bill_discount_excluded":null}`},
		{"POST non boolean", http.MethodPost, "/t", `{"name":"x","is_bill_discount_excluded":"yes"}`},
		{"PATCH null", http.MethodPatch, "/t/9", `{"is_bill_discount_excluded":null}`},
		{"PATCH non boolean", http.MethodPatch, "/t/9", `{"is_bill_discount_excluded":1}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
