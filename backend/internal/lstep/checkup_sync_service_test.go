package lstep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CheckupSyncRepository モック ----

type mockCheckupSyncRepository struct {
	findCheckupSyncPreviewFn func(ctx context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error)
}

func (m *mockCheckupSyncRepository) FindCheckupSyncPreview(ctx context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
	if m.findCheckupSyncPreviewFn != nil {
		return m.findCheckupSyncPreviewFn(ctx, params)
	}
	return nil, nil
}

// ---- ヘルパー ----
// ptrString は accounting_service_test.go で同パッケージ内に既に定義されているため再定義しない。

func newCheckupSyncSvcForPreview(rows []CheckupSyncPreviewRow) CheckupSyncService {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return rows, nil
		},
	}
	return NewCheckupSyncService(
		repo,
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{},
		&mockAuditService{},
	)
}

// ---- deriveExclusionReason: 除外理由の優先度判定 (ISSUE-005) ----

func TestDeriveExclusionReason(t *testing.T) {
	tests := []struct {
		name         string
		hasLine      bool
		isOptOut     bool
		hasLivingPet bool
		want         *string
	}{
		{"全条件OK -> 送信可能（nil）", true, false, true, nil},
		{"opt-out が最優先", true, true, true, ptrString(exclusionReasonOptOut)},
		{"opt-out > 生存ペットなし", true, true, false, ptrString(exclusionReasonOptOut)},
		{"opt-out > LINE未連携", false, true, true, ptrString(exclusionReasonOptOut)},
		{"opt-out > 全部 NG", false, true, false, ptrString(exclusionReasonOptOut)},
		{"生存ペットなし > LINE未連携", false, false, false, ptrString(exclusionReasonNoLivingPet)},
		{"生存ペットなし単独", true, false, false, ptrString(exclusionReasonNoLivingPet)},
		{"LINE未連携単独", false, false, true, ptrString(exclusionReasonLineUnlinked)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveExclusionReason(tt.hasLine, tt.isOptOut, tt.hasLivingPet)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

// ---- PreviewCheckupSync: 件数集計と除外理由付与 (ISSUE-005) ----

func TestCheckupSyncService_PreviewCheckupSync_ExclusionAggregation(t *testing.T) {
	lineID := "U1234567890"
	rows := []CheckupSyncPreviewRow{
		// 1) 送信可能（has_line && !opt_out && living_pet > 0）
		{OwnerID: 1, OwnerName: "送信可能", LineUserID: &lineID, LstepOptOut: false, PetNamesCSV: "ポチ", LivingPetCount: 1},
		// 2) opt-out: ExclusionReason="Lステップ配信停止中"
		{OwnerID: 2, OwnerName: "OPT-OUT", LineUserID: &lineID, LstepOptOut: true, PetNamesCSV: "タマ", LivingPetCount: 1},
		// 3) 死亡ペットのみ: LivingPetCount=0
		{OwnerID: 3, OwnerName: "死亡のみ", LineUserID: &lineID, LstepOptOut: false, PetNamesCSV: "", LivingPetCount: 0},
		// 4) LINE未連携: LineUserID=nil
		{OwnerID: 4, OwnerName: "LINE未連携", LineUserID: nil, LstepOptOut: false, PetNamesCSV: "ハチ", LivingPetCount: 1},
		// 5) opt-out かつ 死亡ペットのみ → opt-out が優先
		{OwnerID: 5, OwnerName: "OPT-OUT&死亡", LineUserID: &lineID, LstepOptOut: true, PetNamesCSV: "", LivingPetCount: 0},
	}

	svc := newCheckupSyncSvcForPreview(rows)

	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 全体件数とサマリー
	assert.Equal(t, 5, result.TotalCount)
	assert.Equal(t, 1, result.EligibleCount, "送信可能は ID=1 の1件のみ")
	assert.Equal(t, 4, result.LineLinkedCount, "LINE連携は ID=1,2,3,5 の4件（opt-out や 生存ペットなしでも line_user_id は別軸でカウント）")
	assert.Equal(t, 2, result.OptOutCount, "opt-out は ID=2,5 の2件")
	assert.Equal(t, 2, result.NoLivingPetCount, "生存ペットなしは ID=3,5 の2件")

	// 各行の除外理由
	assert.Nil(t, result.Owners[0].ExclusionReason, "送信可能 -> nil")
	assert.True(t, result.Owners[0].HasLivingPet)

	if assert.NotNil(t, result.Owners[1].ExclusionReason) {
		assert.Equal(t, exclusionReasonOptOut, *result.Owners[1].ExclusionReason)
	}
	if assert.NotNil(t, result.Owners[2].ExclusionReason) {
		assert.Equal(t, exclusionReasonNoLivingPet, *result.Owners[2].ExclusionReason)
		assert.False(t, result.Owners[2].HasLivingPet)
	}
	if assert.NotNil(t, result.Owners[3].ExclusionReason) {
		assert.Equal(t, exclusionReasonLineUnlinked, *result.Owners[3].ExclusionReason)
		assert.False(t, result.Owners[3].HasLine)
	}
	// opt-out かつ 死亡ペットのみ → opt-out 優先
	if assert.NotNil(t, result.Owners[4].ExclusionReason) {
		assert.Equal(t, exclusionReasonOptOut, *result.Owners[4].ExclusionReason)
	}
}

// ---- PreviewCheckupSync: 空結果のサマリー ----

func TestCheckupSyncService_PreviewCheckupSync_EmptyResult(t *testing.T) {
	svc := newCheckupSyncSvcForPreview([]CheckupSyncPreviewRow{})

	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, 0, result.TotalCount)
	assert.Equal(t, 0, result.EligibleCount)
	assert.Equal(t, 0, result.LineLinkedCount)
	assert.Equal(t, 0, result.OptOutCount)
	assert.Equal(t, 0, result.NoLivingPetCount)
	assert.Empty(t, result.Owners)
}

// ---- CreateCheckupSync: サーバ側二重防御 (ISSUE-007) ----

// newCheckupSyncSvcForCreate は CreateCheckupSync のテスト用に、
// 偽 lstep サーバ・各種モックを束ねたサービスを返す。
// addTagFn は AddTag リクエストを受けた際の検証フック（nil 可）。
func newCheckupSyncSvcForCreate(
	owners map[uint64]*model.Owner,
	livingPetCounts map[uint64]int64,
	addTagFn func(lineUserID, tagName string) int, // 戻り値は HTTPステータスコード
) (CheckupSyncService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /contacts/{lineUserID}/tags の AddTag のみ想定。
		status := http.StatusOK
		if addTagFn != nil {
			// LineUserID は path から取り出す（`/contacts/{id}/tags`）。簡易抽出で十分。
			path := r.URL.Path
			lineUserID := ""
			if len(path) > len("/contacts/") {
				rest := path[len("/contacts/"):]
				for i := 0; i < len(rest); i++ {
					if rest[i] == '/' {
						lineUserID = rest[:i]
						break
					}
				}
			}
			status = addTagFn(lineUserID, "")
		}
		w.WriteHeader(status)
	}))

	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}

	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			result := make([]*model.Owner, 0, len(ids))
			for _, id := range ids {
				if o, ok := owners[id]; ok {
					result = append(result, o)
				}
			}
			return result, nil
		},
	}

	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = livingPetCounts[id]
			}
			return m, nil
		},
	}

	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		ownerRepo,
		petRepo,
		&mockLstepTagCacheRepository{},
		settingsSvc,
		&mockAuditService{},
	)
	return svc, server
}

// TestCheckupSyncService_CreateCheckupSync_SkipsByExclusionReason は
// 直接 POST された場合でも opt-out / 生存ペットなし / LINE未連携の owner が
// 確実にスキップされ、有効な owner のみに送信されることを検証する（ISSUE-007）。
func TestCheckupSyncService_CreateCheckupSync_SkipsByExclusionReason(t *testing.T) {
	// 除外ロジック検証のため deploy write gate を有効化し、eligible のみ HTTP 1 を観測する。
	t.Setenv(lstepapi.EnvWriteAPIEnabled, "true")

	lineID1 := "U_eligible"
	lineID2 := "U_optout" // opt-out なので使われない
	lineID3 := "U_no_living_pet"
	// owner4 は LINE未連携 (LineUserID=nil)
	lineID5 := "U_optout_no_pet" // opt-out + 生存ペットなし → opt-out 優先

	owners := map[uint64]*model.Owner{
		1: {ID: 1, LineUserID: &lineID1, LstepOptOut: false},
		2: {ID: 2, LineUserID: &lineID2, LstepOptOut: true},
		3: {ID: 3, LineUserID: &lineID3, LstepOptOut: false},
		4: {ID: 4, LineUserID: nil, LstepOptOut: false},
		5: {ID: 5, LineUserID: &lineID5, LstepOptOut: true}, // opt-out 優先
	}
	livingPetCounts := map[uint64]int64{
		1: 2, // 生存ペットあり
		2: 1, // 生存ペットあるが opt-out で先にスキップされるため呼ばれない（呼ばれても問題ない値）
		3: 0, // 死亡ペットのみ
		4: 1, // 生存ペットあるが LINE 未連携
		5: 0, // opt-out + 生存ペットなし（opt-out 優先で先に分岐）
	}

	addTagCalls := make(map[string]int)
	svc, server := newCheckupSyncSvcForCreate(owners, livingPetCounts, func(lineUserID, _ string) int {
		addTagCalls[lineUserID]++
		return http.StatusOK
	})
	defer server.Close()

	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1, 2, 3, 4, 5},
		TagName:     "campaign_health_2026",
	}, nil)

	assert.NoError(t, err)
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 1, result.SuccessCount, "送信成功は ID=1 のみ")
	assert.Equal(t, 4, result.SkippedCount, "opt-out (ID=2,5) / 生存ペットなし (ID=3) / LINE未連携 (ID=4) の合計4件")
	assert.Equal(t, 0, result.FailedCount)
	assert.Empty(t, result.FailedOwnerIDs)

	// 両 gate true（deploy exact true + clinic is_sync_enabled default true）で eligible のみ HTTP 1。
	assert.Equal(t, 1, addTagCalls[lineID1], "eligible owner のみ AddTag HTTP 1")
	assert.Equal(t, 0, addTagCalls[lineID2], "opt-out owner には呼ばれない")
	assert.Equal(t, 0, addTagCalls[lineID3], "死亡ペットのみ owner には呼ばれない（誤配信防止）")
	assert.Equal(t, 0, addTagCalls[lineID5], "opt-out + 生存ペットなし owner には呼ばれない")
}

// TestCheckupSyncService_CreateCheckupSync_AllSkipped は
// 全 owner がスキップ対象でも他に影響しないこと、エラーにならないことを検証する。
func TestCheckupSyncService_CreateCheckupSync_AllSkipped(t *testing.T) {
	lineID := "U_dead_only"
	owners := map[uint64]*model.Owner{
		10: {ID: 10, LineUserID: &lineID, LstepOptOut: false},
	}
	livingPetCounts := map[uint64]int64{10: 0} // 死亡ペットのみ

	addTagCalls := 0
	svc, server := newCheckupSyncSvcForCreate(owners, livingPetCounts, func(_, _ string) int {
		addTagCalls++
		return http.StatusOK
	})
	defer server.Close()

	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{10},
		TagName:     "campaign_health_2026",
	}, nil)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 1, result.SkippedCount)
	assert.Equal(t, 0, addTagCalls, "AddTag は一切呼ばれない（誤配信防止）")
}

// TestCheckupSyncService_CreateCheckupSync_NoLstepConfigured は
// Lステップ未設定時の早期エラー挙動を維持していることを検証する（既存挙動の保護）。
func TestCheckupSyncService_CreateCheckupSync_NoLstepConfigured(t *testing.T) {
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "", "", "", nil // 未設定
		},
	}
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		settingsSvc,
		&mockAuditService{},
	)

	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "checkup_2026",
	}, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "未設定時は invalid input エラーであるべき")
}

// TestCheckupSyncService_CreateCheckupSync_AutoManagedTagRejected は
// 自動管理タグの直接付与が拒否される既存挙動を保護する。
func TestCheckupSyncService_CreateCheckupSync_AutoManagedTagRejected(t *testing.T) {
	buildClientCalled := false
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{},
		&mockAuditService{},
	).(*checkupSyncService)
	svc.buildClientFn = func(_ context.Context, _ uint64) (lstepapi.Client, error) {
		buildClientCalled = true
		return &mockLstepAPIClient{}, nil
	}

	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "CPM_CORE", // 固定プレフィックスの自動管理タグ
	}, nil)
	if assert.Error(t, err) {
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), "自動管理タグ")
	}
	assert.False(t, buildClientCalled, "自動管理タグは外部クライアント構築前に拒否されること")
}

// ---- ISSUE-009: 追加フィルタ ----

// TestCheckupSyncService_PreviewCheckupSync_PassesExtendedFiltersToRepo は
// 追加クエリ（年齢/慢性疾患/累計診療費/年間来院/最終健診）が repository に正しく転送されることを検証する。
func TestCheckupSyncService_PreviewCheckupSync_PassesExtendedFiltersToRepo(t *testing.T) {
	var captured FindCheckupSyncPreviewParams

	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			captured = *params
			return nil, nil
		},
	}
	svc := NewCheckupSyncService(
		repo,
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{},
		&mockAuditService{},
	)

	minAge := 7
	maxAge := 12
	chronicTrue := true
	minTotal := int64(50_000)
	minAnnual := int64(2)
	lastBefore := mustParseDate(t, "2026-01-31")
	lastAfter := mustParseDate(t, "2025-01-01")

	_, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType:         "annual",
		Species:             "dog",
		MinAgeYears:         &minAge,
		MaxAgeYears:         &maxAge,
		HasChronicCondition: &chronicTrue,
		CPMStage:            "cpm_core",
		MinTotalAmount:      &minTotal,
		MinAnnualVisitCount: &minAnnual,
		LastCheckupBefore:   &lastBefore,
		LastCheckupAfter:    &lastAfter,
	}, nil)
	assert.NoError(t, err)

	// CPMStage は service 層 post-filter なので repository には渡らない。
	assert.Equal(t, "dog", captured.Species)
	if assert.NotNil(t, captured.MinAgeYears) {
		assert.Equal(t, 7, *captured.MinAgeYears)
	}
	if assert.NotNil(t, captured.MaxAgeYears) {
		assert.Equal(t, 12, *captured.MaxAgeYears)
	}
	if assert.NotNil(t, captured.HasChronicCondition) {
		assert.True(t, *captured.HasChronicCondition)
	}
	if assert.NotNil(t, captured.MinTotalAmount) {
		assert.Equal(t, int64(50_000), *captured.MinTotalAmount)
	}
	if assert.NotNil(t, captured.MinAnnualVisitCount) {
		assert.Equal(t, int64(2), *captured.MinAnnualVisitCount)
	}
	if assert.NotNil(t, captured.LastCheckupBefore) {
		assert.Equal(t, "2026-01-31", captured.LastCheckupBefore.Format("2006-01-02"))
	}
	if assert.NotNil(t, captured.LastCheckupAfter) {
		assert.Equal(t, "2025-01-01", captured.LastCheckupAfter.Format("2006-01-02"))
	}
}

// TestCheckupSyncService_PreviewCheckupSync_CPMStageFilter は
// CPM ステージ指定時に対象外 owner が除外され、サマリー集計も連動することを検証する。
func TestCheckupSyncService_PreviewCheckupSync_CPMStageFilter(t *testing.T) {
	lineID := "U_owner"

	// CPM 値域: cpm_dormant は最終来院から240日超または来院なし。
	// row1: 来院なし → cpm_dormant
	// row2: 直近来院あり、初診1年以上、年間3回以上、LTV 80,000円以上 → cpm_noah
	now := time.Now()
	recent := now.AddDate(0, 0, -10)      // 直近 10 日前
	firstLongAgo := now.AddDate(-2, 0, 0) // 初診 2 年前
	dormantLast := now.AddDate(-1, -6, 0) // 18 ヶ月前 → 240日超
	dormantFirst := now.AddDate(-3, 0, 0)

	rows := []CheckupSyncPreviewRow{
		{
			OwnerID: 1, OwnerName: "活発", LineUserID: &lineID, LstepOptOut: false,
			PetNamesCSV: "ポチ", LivingPetCount: 1,
			LastVisitDate: &recent, FirstVisitDate: &firstLongAgo,
			TotalVisitCount: 5, AnnualVisitCount: 4,
			TotalAmount: 100_000, MaxSingleVisitAmount: 30_000,
		},
		{
			OwnerID: 2, OwnerName: "休眠", LineUserID: &lineID, LstepOptOut: false,
			PetNamesCSV: "タマ", LivingPetCount: 1,
			LastVisitDate: &dormantLast, FirstVisitDate: &dormantFirst,
			TotalVisitCount: 1, AnnualVisitCount: 0,
			TotalAmount: 5_000, MaxSingleVisitAmount: 5_000,
		},
	}

	svc := newCheckupSyncSvcForPreview(rows)

	// cpm_noah で絞り込み → ID=1 のみ
	resultNoah, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType: "annual",
		CPMStage:    "cpm_noah",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, resultNoah) {
		assert.Equal(t, 1, resultNoah.TotalCount, "cpm_noah で絞ると 1 件")
		if assert.Len(t, resultNoah.Owners, 1) {
			assert.Equal(t, uint64(1), resultNoah.Owners[0].OwnerID)
			assert.Equal(t, "cpm_noah", resultNoah.Owners[0].CPMStage)
		}
	}

	// cpm_dormant で絞り込み → ID=2 のみ
	resultDormant, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType: "annual",
		CPMStage:    "cpm_dormant",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, resultDormant) {
		assert.Equal(t, 1, resultDormant.TotalCount, "cpm_dormant で絞ると 1 件")
		if assert.Len(t, resultDormant.Owners, 1) {
			assert.Equal(t, uint64(2), resultDormant.Owners[0].OwnerID)
			assert.Equal(t, "cpm_dormant", resultDormant.Owners[0].CPMStage)
		}
	}

	// 未指定（CPMStage == ""）→ 全件
	resultAll, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType: "annual",
	}, nil)
	assert.NoError(t, err)
	if assert.NotNil(t, resultAll) {
		assert.Equal(t, 2, resultAll.TotalCount, "CPM フィルタ未指定なら全件返る")
	}
}

// TestCheckupSyncService_PreviewCheckupSync_ExposesAdditionalFields は
// 追加レスポンスフィールド（年齢・慢性疾患・累計診療費・年間来院・最終健診日）が
// repository から service に正しく伝播することを検証する。
func TestCheckupSyncService_PreviewCheckupSync_ExposesAdditionalFields(t *testing.T) {
	lineID := "U_owner"
	minAge := 5
	maxAge := 10
	lastCheckup := mustParseDate(t, "2025-09-15")
	now := time.Now()
	lastVisit := now.AddDate(0, 0, -5)
	firstVisit := now.AddDate(-2, 0, 0)

	rows := []CheckupSyncPreviewRow{
		{
			OwnerID: 1, OwnerName: "山田", LineUserID: &lineID, LstepOptOut: false,
			PetNamesCSV: "ポチ", LivingPetCount: 1,
			LastVisitDate: &lastVisit, FirstVisitDate: &firstVisit,
			MinPetAgeYears: &minAge, MaxPetAgeYears: &maxAge,
			HasChronicCondition:  true,
			TotalAmount:          123_456,
			AnnualVisitCount:     3,
			TotalVisitCount:      5,
			MaxSingleVisitAmount: 40_000,
			LastCheckupDate:      &lastCheckup,
		},
	}

	svc := newCheckupSyncSvcForPreview(rows)

	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType: "annual",
	}, nil)
	assert.NoError(t, err)
	if !assert.NotNil(t, result) || !assert.Len(t, result.Owners, 1) {
		return
	}

	o := result.Owners[0]
	if assert.NotNil(t, o.MinPetAgeYears) {
		assert.Equal(t, 5, *o.MinPetAgeYears)
	}
	if assert.NotNil(t, o.MaxPetAgeYears) {
		assert.Equal(t, 10, *o.MaxPetAgeYears)
	}
	assert.True(t, o.HasChronicCondition)
	assert.Equal(t, int64(123_456), o.TotalAmount)
	assert.Equal(t, int64(3), o.AnnualVisitCount)
	if assert.NotNil(t, o.LastCheckupDate) {
		assert.Equal(t, "2025-09-15", o.LastCheckupDate.Format("2006-01-02"))
	}
	// CPM ステージは row の集計値から純粋関数で算出される。
	assert.NotEmpty(t, o.CPMStage, "CPM ステージは常にいずれかの値が入る")
}

// mustParseDate は YYYY-MM-DD のテスト用ヘルパー。
func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("invalid test date %s: %v", s, err)
	}
	return parsed
}

// ---- ISSUE-010: 監査ログ metadata 永続化 ----

// spyCheckupAuditService は LogLstepOperationWithMetadata の引数を捕捉する spy mock。
// 既存の mockAuditService はすべて nil を返す簡易モックのため、ISSUE-010 の検証用に専用 spy を用意する。
type spyCheckupAuditService struct {
	mockAuditService // 既存の互換実装を継承（Log / LogAuthLogin / LogLstepOperation）
	capturedAction   string
	capturedResource string
	capturedMetadata any
	capturedActorID  *uint64
	capturedClinicID uint64
	called           bool
}

func (s *spyCheckupAuditService) LogLstepOperationWithMetadata(_ context.Context, clinicID uint64, actorID *uint64, action, resource string, _ *uint64, metadata any) error {
	s.called = true
	s.capturedClinicID = clinicID
	s.capturedActorID = actorID
	s.capturedAction = action
	s.capturedResource = resource
	s.capturedMetadata = metadata
	return nil
}

// newCheckupSyncSvcForAudit は audit 検証用 spy を注入したサービスを返す。
func newCheckupSyncSvcForAudit(rows []CheckupSyncPreviewRow, spy *spyCheckupAuditService) CheckupSyncService {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
			return rows, nil
		},
	}
	return NewCheckupSyncService(
		repo,
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&mockLstepSettingsService{},
		spy,
	)
}

// TestCheckupSyncService_PreviewCheckupSync_PersistsMetadata は ISSUE-010 受入条件を検証する:
//   - audit_logs.metadata に total_count / eligible_count / line_linked_count /
//     opt_out_count / no_living_pet_count / species / last_visit_before / last_visit_after
//     が記録される。
func TestCheckupSyncService_PreviewCheckupSync_PersistsMetadata(t *testing.T) {
	lineID := "U_test"
	rows := []CheckupSyncPreviewRow{
		{OwnerID: 1, OwnerName: "送信可能", LineUserID: &lineID, LstepOptOut: false, PetNamesCSV: "ポチ", LivingPetCount: 1},
		{OwnerID: 2, OwnerName: "OPT-OUT", LineUserID: &lineID, LstepOptOut: true, PetNamesCSV: "タマ", LivingPetCount: 1},
		{OwnerID: 3, OwnerName: "死亡のみ", LineUserID: &lineID, LstepOptOut: false, PetNamesCSV: "", LivingPetCount: 0},
		{OwnerID: 4, OwnerName: "LINE未連携", LineUserID: nil, LstepOptOut: false, PetNamesCSV: "ハチ", LivingPetCount: 1},
	}
	lastBefore := mustParseDate(t, "2026-01-31")
	lastAfter := mustParseDate(t, "2025-01-01")

	spy := &spyCheckupAuditService{}
	svc := newCheckupSyncSvcForAudit(rows, spy)

	staffID := uint64(99)
	_, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType:     "annual",
		Species:         "dog",
		LastVisitBefore: &lastBefore,
		LastVisitAfter:  &lastAfter,
	}, &staffID)
	assert.NoError(t, err)
	assert.True(t, spy.called, "audit ログが呼び出されている")
	assert.Equal(t, "checkup_sync_preview", spy.capturedAction)
	assert.Equal(t, "owner", spy.capturedResource)
	assert.Equal(t, uint64(1), spy.capturedClinicID)
	if assert.NotNil(t, spy.capturedActorID) {
		assert.Equal(t, staffID, *spy.capturedActorID)
	}

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok, "metadata は map[string]any で渡される") {
		return
	}

	// 集計件数（受入条件: total/eligible/line_linked/opt_out/no_living_pet）
	assert.Equal(t, 4, meta["total_count"])
	assert.Equal(t, 1, meta["eligible_count"], "送信可能は ID=1 のみ")
	assert.Equal(t, 3, meta["line_linked_count"], "LINE連携 ID=1,2,3")
	assert.Equal(t, 1, meta["opt_out_count"])
	assert.Equal(t, 1, meta["no_living_pet_count"])

	// フィルタ条件（受入条件: species / last_visit_before / last_visit_after）
	filter, ok := meta["filter"].(map[string]any)
	if !assert.True(t, ok, "filter は object として渡される") {
		return
	}
	assert.Equal(t, "annual", filter["checkup_type"])
	assert.Equal(t, "dog", filter["species"])
	assert.Equal(t, "2026-01-31", filter["last_visit_before"])
	assert.Equal(t, "2025-01-01", filter["last_visit_after"])
}

// TestCheckupSyncService_PreviewCheckupSync_PersistsMetadata_NilFilters は
// 任意フィルタが nil の場合、metadata.filter にそのキーが含まれないこと（誤った "" / 0 を残さない）を検証する。
func TestCheckupSyncService_PreviewCheckupSync_PersistsMetadata_NilFilters(t *testing.T) {
	spy := &spyCheckupAuditService{}
	svc := newCheckupSyncSvcForAudit([]CheckupSyncPreviewRow{}, spy)

	_, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{
		CheckupType: "annual",
	}, nil)
	assert.NoError(t, err)
	assert.True(t, spy.called)

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	filter, ok := meta["filter"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "annual", filter["checkup_type"])
	// nil フィルタは含まれない（後段の SQL 検索で意味のない条件を残さない）
	_, hasLastVisitBefore := filter["last_visit_before"]
	assert.False(t, hasLastVisitBefore)
	_, hasMinAge := filter["min_age_years"]
	assert.False(t, hasMinAge)
	_, hasCPM := filter["cpm_stage"]
	assert.False(t, hasCPM)
}

// TestCheckupSyncService_CreateCheckupSync_PersistsMetadata は CreateCheckupSync 実行時に
// 件数とスキップ理由内訳が metadata として永続化されることを検証する（ISSUE-010）。
func TestCheckupSyncService_CreateCheckupSync_PersistsMetadata(t *testing.T) {
	// 成功経路の metadata を検証するため deploy write gate を有効化する。
	t.Setenv(lstepapi.EnvWriteAPIEnabled, "true")

	// CreateCheckupSync は HTTP モックが必要なため newCheckupSyncSvcForCreate と組み合わせる。
	lineID1 := "U_eligible"
	owners := map[uint64]*model.Owner{
		1: {ID: 1, LineUserID: &lineID1, LstepOptOut: false},
		2: {ID: 2, LineUserID: nil, LstepOptOut: false}, // LINE未連携
	}
	livingPetCounts := map[uint64]int64{1: 1, 2: 1}

	// 既存の newCheckupSyncSvcForCreate は spy を差し込めないため、ここで直接組み立てる。
	spy := &spyCheckupAuditService{}

	// httptest サーバ（200 を返す素朴な lstep API モック）。
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			result := make([]*model.Owner, 0, len(ids))
			for _, id := range ids {
				if o, ok := owners[id]; ok {
					result = append(result, o)
				}
			}
			return result, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = livingPetCounts[id]
			}
			return m, nil
		},
	}

	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		ownerRepo,
		petRepo,
		&mockLstepTagCacheRepository{},
		settingsSvc,
		spy,
	)

	staffID := uint64(7)
	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1, 2},
		TagName:     "campaign_health_2026",
	}, &staffID)
	assert.NoError(t, err)
	assert.True(t, spy.called)
	assert.Equal(t, "checkup_sync", spy.capturedAction)

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "checkup_sync_execute", meta["operation"])
	assert.Equal(t, "annual", meta["checkup_type"])
	assert.Equal(t, "campaign_health_2026", meta["tag_name"])
	assert.Equal(t, 2, meta["requested_count"])
	assert.Equal(t, 1, meta["success_count"])
	assert.Equal(t, 1, meta["skipped_count"], "ID=2 は LINE 未連携でスキップ")
	assert.Equal(t, 1, meta["skipped_line_unlinked"])
	assert.Equal(t, 0, meta["skipped_opt_out"])
	assert.Equal(t, 0, meta["skipped_no_living_pet"])
}

// ================================================================
// buildClient: 分岐網羅（IsSyncEnabled / GetRawCredentials / buildClientFn 差し替え）
// ================================================================

func TestCheckupSyncService_buildClient(t *testing.T) {
	t.Run("buildClientFn が設定されている場合はそれを使う（テスト注入経路）", func(t *testing.T) {
		called := false
		svc := &checkupSyncService{
			buildClientFn: func(_ context.Context, clinicID uint64) (lstepapi.Client, error) {
				called = true
				assert.Equal(t, uint64(1), clinicID)
				return nil, nil
			},
		}
		_, err := svc.buildClient(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("IsSyncEnabled がエラーを返す場合はラップして返す", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := &checkupSyncService{settingsSvc: settingsSvc}

		_, err := svc.buildClient(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("同期無効の場合は nil, nil を返す", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, nil
			},
		}
		svc := &checkupSyncService{settingsSvc: settingsSvc}

		client, err := svc.buildClient(context.Background(), 1)
		assert.NoError(t, err)
		assert.Nil(t, client)
	})

	t.Run("GetRawCredentials がエラーを返す場合はラップして返す", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "", "", "", errors.New("db error")
			},
		}
		svc := &checkupSyncService{settingsSvc: settingsSvc}

		_, err := svc.buildClient(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("apiKeyが空の場合は nil, nil を返す", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "", "https://example.com", "", nil
			},
		}
		svc := &checkupSyncService{settingsSvc: settingsSvc}

		client, err := svc.buildClient(context.Background(), 1)
		assert.NoError(t, err)
		assert.Nil(t, client)
	})

	t.Run("apiKeyが設定されている場合はクライアントを返す", func(t *testing.T) {
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "test-key", "https://example.com", "", nil
			},
		}
		svc := &checkupSyncService{settingsSvc: settingsSvc}

		client, err := svc.buildClient(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// ================================================================
// Write dual-gate: deploy LSTEP_WRITE_API_ENABLED + clinic is_sync_enabled
// ================================================================

// TestCheckupSyncService_CreateCheckupSync_WriteGateDisabled_NoReceiptUpdate は
// deploy gate 無効時に real httpLstepClient が ErrWriteDisabled を返し、
// HTTP 0・Success 0・tag cache receipt (UpsertTag) 0 であることを証明する。
// （delivery fired / receipt 更新に相当する成功経路副作用が走らないこと）
func TestCheckupSyncService_CreateCheckupSync_WriteGateDisabled_NoReceiptUpdate(t *testing.T) {
	// 既定: LSTEP_WRITE_API_ENABLED 未設定/非 true → write disabled
	t.Setenv(lstepapi.EnvWriteAPIEnabled, "false")

	lineID := "U_gate_disabled"
	owners := map[uint64]*model.Owner{
		1: {ID: 1, LineUserID: &lineID, LstepOptOut: false},
	}
	livingPetCounts := map[uint64]int64{1: 1}

	httpHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpHits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	upsertCalls := 0
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			upsertCalls++
			return nil
		},
	}

	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			result := make([]*model.Owner, 0, len(ids))
			for _, id := range ids {
				if o, ok := owners[id]; ok {
					result = append(result, o)
				}
			}
			return result, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = livingPetCounts[id]
			}
			return m, nil
		},
	}

	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		ownerRepo,
		petRepo,
		tagCache,
		settingsSvc,
		&mockAuditService{},
	)

	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_gate_test",
	}, nil)
	assert.NoError(t, err)
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 0, result.SuccessCount, "gate error must not count as success")
	assert.Equal(t, 1, result.FailedCount, "ErrWriteDisabled is recorded as failed owner")
	assert.Equal(t, []uint64{1}, result.FailedOwnerIDs)
	assert.Equal(t, 0, httpHits, "deploy gate false: 0 HTTP requests")
	assert.Equal(t, 0, upsertCalls, "receipt/tag-cache update must stay 0 on gate error")
}

// TestCheckupSyncService_CreateCheckupSync_ClinicSyncDisabled_ZeroHTTP は
// clinic is_sync_enabled=false のとき client 未生成で HTTP 0・Create が設定エラーになることを証明する。
func TestCheckupSyncService_CreateCheckupSync_ClinicSyncDisabled_ZeroHTTP(t *testing.T) {
	// deploy gate を true にしても clinic flag false なら write しない（二重 gate）。
	t.Setenv(lstepapi.EnvWriteAPIEnabled, "true")

	httpHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpHits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		settingsSvc,
		&mockAuditService{},
	)

	_, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_clinic_off",
	}, nil)
	assert.Error(t, err)
	assert.Equal(t, 0, httpHits, "clinic is_sync_enabled=false: 0 HTTP requests")
}

// TestCheckupSyncService_CreateCheckupSync_BothGatesTrue_HTTPAndReceipt は
// deploy gate true + clinic sync true のとき HTTP 1 と tag cache receipt 1 を証明する。
func TestCheckupSyncService_CreateCheckupSync_BothGatesTrue_HTTPAndReceipt(t *testing.T) {
	t.Setenv(lstepapi.EnvWriteAPIEnabled, "true")

	lineID := "U_both_gates"
	owners := map[uint64]*model.Owner{
		1: {ID: 1, LineUserID: &lineID, LstepOptOut: false},
	}
	livingPetCounts := map[uint64]int64{1: 1}

	httpHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpHits++
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	upsertCalls := 0
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, clinicID, ownerID uint64, tagName, category, _ string) error {
			upsertCalls++
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), ownerID)
			assert.Equal(t, "campaign_both_gates", tagName)
			assert.Equal(t, "manual", category)
			return nil
		},
	}

	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
			result := make([]*model.Owner, 0, len(ids))
			for _, id := range ids {
				if o, ok := owners[id]; ok {
					result = append(result, o)
				}
			}
			return result, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64]int64, error) {
			m := make(map[uint64]int64, len(ownerIDs))
			for _, id := range ownerIDs {
				m[id] = livingPetCounts[id]
			}
			return m, nil
		},
	}

	svc := NewCheckupSyncService(
		&mockCheckupSyncRepository{},
		ownerRepo,
		petRepo,
		tagCache,
		settingsSvc,
		&mockAuditService{},
	)

	result, err := svc.CreateCheckupSync(context.Background(), 1, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    []uint64{1},
		TagName:     "campaign_both_gates",
	}, nil)
	assert.NoError(t, err)
	if !assert.NotNil(t, result) {
		return
	}
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Equal(t, 1, httpHits, "both gates true: HTTP 1")
	assert.Equal(t, 1, upsertCalls, "success path updates tag cache receipt once")
}
