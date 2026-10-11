package trimming

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// makeTrimmingLifecycleAppointment はトリミング区分の予約を1件作成する。
// 一般区分の makeReservation とは異なり LockTrimmingByID / DeleteForTrimming が
// 受理する reservation_type を要求するサービス経路のテストで使う。
func makeTrimmingLifecycleAppointment(t *testing.T, db *gorm.DB, clinicID uint64) *model.Reservation {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "トリミング",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(rt).Error)
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(30 * time.Minute),
		ReservationTypeID: rt.ID,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
	}
	require.NoError(t, db.Create(res).Error)
	return res
}

// makeLifecycleDetail は指定予約へ画像キー付きの trimming detail を直接 INSERT する。
// repo.Create を通さず既存状態を作るため、writer 側のキークレイムを前提にしない。
func makeLifecycleDetail(t *testing.T, db *gorm.DB, clinicID, appointmentID uint64, style, completed string) {
	t.Helper()
	require.NoError(t, db.Create(&model.AppointmentTrimmingDetail{
		ClinicID:       clinicID,
		AppointmentID:  appointmentID,
		StyleImage:     style,
		CompletedImage: completed,
		BWUnit:         model.BodyWeightUnitKg,
	}).Error)
}

func newLifecycleService(t *testing.T, db *gorm.DB) Service {
	t.Helper()
	svc := NewServiceWithAudit(
		reservation.NewReservationRepository(db),
		&mockTrimmingReservationTypeRepository{},
		newAcceptingTrimmingStaffRepository(),
		&mockTrimmingUnavailableTimeRepository{},
		NewAppointmentTrimmingDetailRepository(db),
		newActiveTrimmingCourseRepository(),
		newActiveTrimmingOptionRepository(),
		&testTransactor{db: db},
		noopTrimmingAuditTxLogger{},
	)
	// writer 側 claim の Exists 確認用に既定の durable ストアを配線する
	// （fake 既定は存在扱い）。WithImageSupport を通るテストは後続の
	// bind で実テスト用 uploader に差し替わる。旧実装では no-op。
	bindLifecycleServiceUploader(svc, &fakeTrimmingImageUploader{})
	return svc
}

func newLifecycleImageHandler(t *testing.T, db *gorm.DB) (*Handler, *fakeTrimmingImageUploader) {
	t.Helper()
	uploader := &fakeTrimmingImageUploader{}
	h := newHandlerWithTrimmingSvc(newLifecycleService(t, db))
	h.WithImageSupport(ImageDependencies{Uploader: uploader})
	return h, uploader
}

// waitForTrimmingLockWaiter は新規ロック待機か done のどちらか早い方を待つ。
// 旧実装ではロック待機が発生しないため、タイムアウトで諦めても先へ進める。
func waitForTrimmingLockWaiters(ctx context.Context, db *gorm.DB, done <-chan struct{}, want int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if done != nil {
			select {
			case <-done:
				return
			default:
			}
		}
		var waiting bool
		_ = db.WithContext(ctx).
			Raw("SELECT COUNT(*) >= ? FROM pg_locks WHERE NOT granted", want).
			Scan(&waiting).Error
		if waiting {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDeleteReplacedTrimmingImages_RetainsKeyReReferencedBySameAppointment は
// update1 で破棄されたキー K が、update1 コミット後に同じ予約へ再参照されている
// 状態を再現する。クリーンアップは stale な updated スナップショット (K2) を持つが、
// コミット済みの DB 上の参照を見なければならない。
// 旧実装は exceptAppointmentID で自予約を参照判定から除外するため K を削除してしまう。
func TestDeleteReplacedTrimmingImages_RetainsKeyReReferencedBySameAppointment(t *testing.T) {
	// flag-on: 参照再確認が「参照あり」を返して保持されることを検証する
	// （flag-off の無条件保持ではなく、live 参照チェック経由の保持）。
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(7)
	keyK := "trimming-images/7/re-ref-same.jpg"
	keyK2 := "trimming-images/7/re-ref-same-2.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	// 再参照後のコミット済み状態: 同じ予約が再び K を参照している。
	makeLifecycleDetail(t, db, clinicID, appt.ID, keyK, "")

	h, uploader := newLifecycleImageHandler(t, db)
	// update1 の結果スナップショットは style=K2 (stale)。破棄集合は {K}。
	updated := &model.Reservation{
		ID:       appt.ID,
		ClinicID: clinicID,
		TrimmingDetail: &model.AppointmentTrimmingDetail{
			StyleImage: keyK2,
		},
	}
	h.deleteReplacedTrimmingImages(ctx, clinicID, &TrimmingReplacedImages{Style: keyK}, updated)

	assert.NotContains(t, uploader.deletes, keyK,
		"コミット済みの再参照キーは post-commit クリーンアップで削除してはいけない")
}

// TestDeleteReplacedTrimmingImages_RetainsCrossAppointmentReference は
// 別予約が参照する共有キーの保持契約を固定する（新旧両実装で成立すべき基準契約）。
func TestDeleteReplacedTrimmingImages_RetainsCrossAppointmentReference(t *testing.T) {
	// flag-on: 別予約の live 参照を参照再確認が検出して保持することを検証する。
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(8)
	keyK := "trimming-images/8/cross-appt.jpg"
	keyK2 := "trimming-images/8/cross-appt-new.jpg"

	apptA := makeTrimmingLifecycleAppointment(t, db, clinicID)
	apptB := makeTrimmingLifecycleAppointment(t, db, clinicID)
	makeLifecycleDetail(t, db, clinicID, apptB.ID, keyK, "")

	h, uploader := newLifecycleImageHandler(t, db)
	updated := &model.Reservation{
		ID:       apptA.ID,
		ClinicID: clinicID,
		TrimmingDetail: &model.AppointmentTrimmingDetail{
			StyleImage: keyK2,
		},
	}
	h.deleteReplacedTrimmingImages(ctx, clinicID, &TrimmingReplacedImages{Style: keyK}, updated)

	assert.NotContains(t, uploader.deletes, keyK,
		"別予約のコミット済み参照があるキーは削除してはいけない")
}

// TestDeleteReplacedTrimmingImages_SerializesWithInFlightWriter は
// 参照チェックとオブジェクト削除の間に別予約が同じキーをコミットする
// check-to-delete ウィンドウ競合を再現する。writer トランザクションが
// キーのロックを保持したまま未コミットである間、クリーンアップは
// そのコミットを待ってから参照を再確認しなければならない。
// 旧実装にはこの直列化がなく、未コミット参照を見ずに K を削除する。
func TestDeleteReplacedTrimmingImages_SerializesWithInFlightWriter(t *testing.T) {
	// flag-on: セッションロック取得が writer の xact lock コミットまで
	// ブロックされる直列化経路を検証する（flag-off ではロックを取らない）。
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(9)
	keyK := "trimming-images/9/mid-window.jpg"
	keyK2 := "trimming-images/9/mid-window-new.jpg"

	apptA := makeTrimmingLifecycleAppointment(t, db, clinicID)
	apptB := makeTrimmingLifecycleAppointment(t, db, clinicID)
	// writer 側 claim の Exists 確認を通すため durable ストアを配線する
	// （fake 既定は存在扱い）。旧実装では no-op。
	uploader := &fakeTrimmingImageUploader{}
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)

	// writer tx: 別予約へ K を格納する detail を未コミットのまま保持する。
	txWriter := db.Begin()
	require.NoError(t, txWriter.Error)
	require.NoError(t, repo.Create(
		persistence.WithTxValue(ctx, txWriter),
		&model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: apptB.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		},
	))

	h, _ := newLifecycleImageHandler(t, db)
	updated := &model.Reservation{
		ID:       apptA.ID,
		ClinicID: clinicID,
		TrimmingDetail: &model.AppointmentTrimmingDetail{
			StyleImage: keyK2,
		},
	}
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		h.deleteReplacedTrimmingImages(ctx, clinicID, &TrimmingReplacedImages{Style: keyK}, updated)
	}()

	// 新実装ではクリーンアップが writer のキーロックで待機する。
	// 旧実装では待機が発生しないためタイムアウトで先へ進む。
	waitForTrimmingLockWaiters(ctx, db, cleanupDone, 1, 2*time.Second)
	require.NoError(t, txWriter.Commit().Error)

	select {
	case <-cleanupDone:
	case <-time.After(10 * time.Second):
		require.FailNow(t, "cleanup did not finish")
	}
	assert.NotContains(t, uploader.deletes, keyK,
		"参照チェック後にコミットされたキーは削除してはいけない")
}

// TestDeleteTrimmingImage_RejectsStoreAfterDelete は
// DELETE 済みオブジェクトキーへの再格納が writer 側ガードで拒否される
// ことを確認する（durable tombstone = オブジェクトの不在）。
// 旧実装は参照不在チェック後に削除するだけで
// 書き込み側を拒否する手段がなく、再格納が成功してしまう。
func TestDeleteTrimmingImage_RejectsStoreAfterDelete(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	gin.SetMode(gin.TestMode)
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(1) // setTrimmingWriteContext が c.Set("clinic_id", "1") する
	keyK := "trimming-images/1/deleted-store.jpg"

	h, uploader := newLifecycleImageHandler(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(
		fmt.Sprintf(`{"key":%q}`, keyK)))
	c.Request.Header.Set("Content-Type", "application/json")
	setTrimmingWriteContext(c)
	h.DeleteTrimmingImage(c)

	require.Equal(t, http.StatusNoContent, c.Writer.Status())
	require.Contains(t, uploader.deletes, keyK)

	// 削除済みキーを新規 detail へ格納しようとする writer。
	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)
	err := (&testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	require.Error(t, err, "削除済みオブジェクトキーの再格納は拒否されなければならない")
}

// TestDeleteTrimmingImage_SoftDeletedParentReferenceCleaned は
// 参照行が soft-delete 済み予約配下のみの場合の参照数契約を固定する。
// A3 primary/live-parent 契約: versioned source の write-path 列挙で
// appointments.deleted_at を NULL へ戻す restore 経路が存在しないことが
// 立証済み（evidence/appointments-no-restore-r9.json、versioned bound・
// main 統合後の再検証必須）のため soft-deleted 親配下の参照は数えず、
// オブジェクトは削除される。
func TestDeleteTrimmingImage_SoftDeletedParentReferenceCleaned(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	gin.SetMode(gin.TestMode)
	db := setupImageReferenceTestDB(t)
	clinicID := uint64(1)
	keyK := "trimming-images/1/soft-deleted-parent.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	makeLifecycleDetail(t, db, clinicID, appt.ID, keyK, "")
	require.NoError(t, db.Where("id = ?", appt.ID).Delete(&model.Reservation{}).Error)

	h, uploader := newLifecycleImageHandler(t, db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(
		fmt.Sprintf(`{"key":%q}`, keyK)))
	c.Request.Header.Set("Content-Type", "application/json")
	setTrimmingWriteContext(c)
	h.DeleteTrimmingImage(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status(),
		"primary 契約: soft-deleted 親配下の参照のみのキーは削除される")
	assert.Contains(t, uploader.deletes, keyK)
}

type cleanupEvalCall struct {
	key     string
	deleted bool
	err     error
}

type cleanupEvalSpyService struct {
	Service
	mu    sync.Mutex
	calls []cleanupEvalCall
}

func (s *cleanupEvalSpyService) DeleteUnreferencedImageObject(
	ctx context.Context, clinicID uint64, key string, deleteObject ImageObjectDeleteFunc,
) (bool, error) {
	deleted, err := s.Service.DeleteUnreferencedImageObject(ctx, clinicID, key, deleteObject)
	s.mu.Lock()
	s.calls = append(s.calls, cleanupEvalCall{key: key, deleted: deleted, err: err})
	s.mu.Unlock()
	return deleted, err
}

func (s *cleanupEvalSpyService) bindImageObjectStore(u FileUploader) {
	if b, ok := s.Service.(imageObjectStoreBinder); ok {
		b.bindImageObjectStore(u)
	}
}

func (s *cleanupEvalSpyService) callsFor(key string) []cleanupEvalCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []cleanupEvalCall
	for _, c := range s.calls {
		if c.key == key {
			out = append(out, c)
		}
	}
	return out
}

// TestServiceDelete_CapturesLockedTimeImages は Service.Delete が
// delete tx の行ロック下で捕捉した画像値を返す契約を固定する。
// K1→K2 の更新コミット後の Delete は K2（locked-time 値）を返す。
func TestServiceDelete_CapturesLockedTimeImages(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.StaffClinicAssignment{}, &model.LineCustomer{}))
	ctx := context.Background()
	clinicID := uint64(12)
	keyK1 := "trimming-images/12/locked-old.jpg"
	keyK2 := "trimming-images/12/locked-new.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	makeLifecycleDetail(t, db, clinicID, appt.ID, keyK1, "")

	svc := newLifecycleService(t, db)
	_, _, err := svc.Update(ctx, clinicID, appt.ID, &UpdateTrimmingInput{
		StyleImage: strPtr(keyK2),
		ActorID:    ptrUint64(42),
	})
	require.NoError(t, err)

	removed, err := svc.Delete(ctx, clinicID, appt.ID, ptrUint64(42))
	require.NoError(t, err)
	require.NotNil(t, removed)
	assert.Equal(t, keyK2, removed.Style, "locked-time の画像値を返す")
	assert.Empty(t, removed.Completed)
}

// TestBestEffortDeleteTrimmingImageKeys_CheckErrorRetainsObject は
// 参照チェック不能時にオブジェクトを消さない fail-closed 契約を固定する。
func TestBestEffortDeleteTrimmingImageKeys_CheckErrorRetainsObject(t *testing.T) {
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		deleteUnreferencedImageObjectFn: func(context.Context, uint64, string, ImageObjectDeleteFunc) (bool, error) {
			return false, errors.New("reference check unavailable")
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	h.bestEffortDeleteTrimmingImageKeys(context.Background(), 1,
		"trimming-images/1/check-err.jpg")

	assert.Empty(t, uploader.deletes,
		"参照状況を確認できないキーは削除してはいけない")
}

// TestDeleteUnreferencedImageObject_NotConfiguredFailsClosed は
// 参照確認に必要な依存が無い場合にガードが失敗を返し、呼び出し元が
// オブジェクトを保持する契約を固定する。
func TestDeleteUnreferencedImageObject_NotConfiguredFailsClosed(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	svc := &trimmingService{}
	deleted, err := svc.DeleteUnreferencedImageObject(
		context.Background(), 1, "trimming-images/1/any.jpg",
		func(context.Context, string) error { return nil })
	require.Error(t, err)
	assert.False(t, deleted)
}

// ---- durable object-existence guard (attempt-3) ----

// bindLifecycleRepoUploader は imageObjectStoreBinder 互換の構造的アサーション。
// 旧実装（uploader バインド無し）では no-op になり、durable 実装では
// FileUploader.Exists を claim 側へ配線する。
func bindLifecycleRepoUploader(repo AppointmentTrimmingDetailRepository, uploader FileUploader) {
	if b, ok := repo.(interface{ bindImageObjectStore(FileUploader) }); ok {
		b.bindImageObjectStore(uploader)
	}
}

func bindLifecycleServiceUploader(svc Service, uploader FileUploader) {
	if b, ok := svc.(interface{ bindImageObjectStore(FileUploader) }); ok {
		b.bindImageObjectStore(uploader)
	}
}

// lifecycleTestDSN は testdb.connectTestDatabase と同じ規則で test DB への
// DSN を組み立てる。second-connection / forked-self サブプロセス用。
func lifecycleTestDSN() string {
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		return envDSN
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "ekarte_db"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "ekarte_user"
	}
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		pass = "ekarte_password"
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s_test?sslmode=disable", user, pass, host, port, name)
}

// runTrimmingImageLifecycleHelper は TRIMMING_IMAGE_LIFECYCLE_HELPER=claim で
// 起動された子プロセスの入口。親が削除済みにしたオブジェクトキーを、独立に
// 初期化されたプロセス（in-process ガード状態を引き継がない）から格納しようと
// する cross-process oracle。
func runTrimmingImageLifecycleHelper() int {
	return runTrimmingImageConflictOracleHelper()
}

// TestClaimTrimmingImageKeys_MissingObjectRejected は durable tombstone 契約の
// writer 側: オブジェクトが既に存在しないキーの格納は拒否される。
// 旧実装はオブジェクト存在を見ないため格納が成功する（RED）。
func TestClaimTrimmingImageKeys_MissingObjectRejected(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(13)
	keyK := "trimming-images/13/missing-object.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	uploader := &fakeTrimmingImageUploader{missing: map[string]bool{keyK: true}}
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)

	err := (&testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	require.Error(t, err, "オブジェクトが存在しないキーの格納は拒否されなければならない")
	assert.True(t, apperrors.IsConflict(err),
		"削除済みオブジェクトへの再参照は conflict で拒否")
}

// TestClaimTrimmingImageKeys_ExistsCheckErrorFailsClosed はオブジェクト存在
// チェックがインフラエラーで確認不能な場合、格納を fail-closed で拒否する契約。
// 旧実装は存在チェック自体を持たないため格納が成功する（RED）。
func TestClaimTrimmingImageKeys_ExistsCheckErrorFailsClosed(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(15)
	keyK := "trimming-images/15/exists-err.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	uploader := &fakeTrimmingImageUploader{existsErr: errors.New("head object failed")}
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)

	err := (&testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	require.Error(t, err, "オブジェクト存在を確認できないキーの格納は fail-closed で拒否されなければならない")
	assert.False(t, apperrors.IsConflict(err),
		"存在チェックのインフラエラーは conflict ではなく fail-closed エラー")
}

// TestClaimTrimmingImageKeys_WaitedWriterSeesDeletedObject は
// キーロック待ちでブロックされた writer が、セッションロック保持側の
// 「参照再確認→オブジェクト削除→unlock」クリティカルセクションの後に
// 進み、オブジェクト不在を見て格納を拒否されることを検証する
// （check-to-delete ウィンドウの durable 閉鎖）。
// 旧実装の tombstone はこのシナリオで立たないため格納が成功する（RED）。
func TestClaimTrimmingImageKeys_WaitedWriterSeesDeletedObject(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(16)
	keyK := "trimming-images/16/deleted-while-waiting.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	uploader := &fakeTrimmingImageUploader{}
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)

	// cleanup 側: セッションスコープ advisory lock を pinned conn で保持。
	sqlDB, err := db.DB()
	require.NoError(t, err)
	holder, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	defer holder.Close() //nolint:errcheck // conn の Close 失敗は復旧不可のため無視
	_, err = holder.ExecContext(ctx,
		"SELECT pg_advisory_lock(hashtextextended($1, 0))",
		trimmingImageKeyLockDomain+keyK)
	require.NoError(t, err)

	txWriter := db.Begin()
	require.NoError(t, txWriter.Error)
	writerDone := make(chan error, 1)
	go func() {
		writerDone <- repo.Create(
			persistence.WithTxValue(ctx, txWriter),
			&model.AppointmentTrimmingDetail{
				ClinicID:      clinicID,
				AppointmentID: appt.ID,
				StyleImage:    keyK,
				BWUnit:        model.BodyWeightUnitKg,
			},
		)
	}()

	// writer がキーロック待ちに入るまで待つ（xact lock はセッションロックと衝突する）。
	waitForTrimmingLockWaiters(ctx, db, nil, 1, 5*time.Second)

	// cleanup 側クリティカルセクション: ロック下でオブジェクトを削除してから unlock。
	require.NoError(t, uploader.Delete(ctx, keyK))
	_, err = holder.ExecContext(context.Background(),
		"SELECT pg_advisory_unlock(hashtextextended($1, 0))",
		trimmingImageKeyLockDomain+keyK)
	require.NoError(t, err)

	select {
	case err := <-writerDone:
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err),
			"ロック待ち解放後に削除済みオブジェクトを見た writer は conflict で拒否")
	case <-time.After(10 * time.Second):
		require.FailNow(t, "writer did not finish")
	}
	_ = txWriter.Rollback()
}

// TestBestEffortDeleteTrimmingImageKeys_ObjectDeleteDisabledRetains は
// TRIMMING_IMAGE_OBJECT_DELETE_ENABLED が "1" でない限りオブジェクト削除を
// 行わないロールアウトゲート契約（旧バージョンのインスタンスが存在する間の
// 誤削除防止）。旧実装はフラグを参照しないため削除してしまう（RED）。
func TestBestEffortDeleteTrimmingImageKeys_ObjectDeleteDisabledRetains(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "0")
	db := setupImageReferenceTestDB(t)
	h, uploader := newLifecycleImageHandler(t, db)

	h.bestEffortDeleteTrimmingImageKeys(context.Background(), 1,
		"trimming-images/1/flag-off.jpg")

	assert.Empty(t, uploader.deletes,
		"TRIMMING_IMAGE_OBJECT_DELETE_ENABLED != \"1\" の間は削除をスキップする")
}

// TestBestEffortDeleteTrimmingImageKeys_ObjectDeleteEnabledDeletes は
// フラグ有効時はガード通過後に削除される基準契約（新旧両実装で成立）。
func TestBestEffortDeleteTrimmingImageKeys_ObjectDeleteEnabledDeletes(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	h, uploader := newLifecycleImageHandler(t, db)

	h.bestEffortDeleteTrimmingImageKeys(context.Background(), 1,
		"trimming-images/1/flag-on.jpg")

	assert.Contains(t, uploader.deletes, "trimming-images/1/flag-on.jpg")
}

// TestDeleteUnreferencedImageObject_SecondConnectionRejectsClaim は
// 削除確定後、別コネクション（別接続プール＝別インスタンス相当）の writer が
// 同じキーの格納を拒否される契約。プロセスをまたいだ拒否は
// forked-self テストが担い、ここでは第二接続の基準挙動を固定する。
func TestDeleteUnreferencedImageObject_SecondConnectionRejectsClaim(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(17)
	keyK := "trimming-images/17/second-conn.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	dir := t.TempDir()
	uploader := infra.NewLocalUploader(dir, "/uploads")
	_, err := uploader.Upload(ctx, keyK, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)

	svc := newLifecycleService(t, db)
	h := newHandlerWithTrimmingSvc(svc)
	h.WithImageSupport(ImageDependencies{Uploader: uploader})
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, keyK)
	_, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(keyK)))
	require.True(t, os.IsNotExist(statErr), "cleanup はオブジェクトを削除しなければならない")

	// 別接続プール = 別インスタンス相当の writer。
	db2, err := gorm.Open(postgres.Open(lifecycleTestDSN()), &gorm.Config{})
	require.NoError(t, err)
	repo2 := NewAppointmentTrimmingDetailRepository(db2)
	bindLifecycleRepoUploader(repo2, uploader)
	err = db2.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return repo2.Create(persistence.WithTxValue(ctx, tx), &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	require.Error(t, err, "削除済みオブジェクトキーへの別接続からの再格納は拒否されなければならない")
	assert.True(t, apperrors.IsConflict(err))
}

// TestDeleteUnreferencedImageObject_SubprocessClaimRejected は cross-process
// oracle: 親プロセスが durable ガード経由でオブジェクトを削除した後、
// fork した子プロセス（in-process 状態を引き継がない独立インスタンス）が
// 同じキーを格納しようとして typed conflict で拒否されることを検証する。
// 旧実装の sync.Map tombstone は子プロセスに存在しないため claim が成功し、
// 子が exit 1 を返す（RED）。shared runner は既存オブジェクトキーの
// positive control と typed-conflict marker の両方を要求する。
func TestDeleteUnreferencedImageObject_SubprocessClaimRejected(t *testing.T) {
	runSubprocessTypedConflictOracleTest(t, "claim")
}

// TestUploadThenReferenceLegitimateFlow は upload → reference の正規順序の
// 回帰契約: アップロード済み（存在する）オブジェクトへの参照格納は受理される。
func TestUploadThenReferenceLegitimateFlow(t *testing.T) {
	db := setupImageReferenceTestDB(t)
	ctx := context.Background()
	clinicID := uint64(19)
	keyK := "trimming-images/19/legit.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	uploader := &fakeTrimmingImageUploader{}
	_, err := uploader.Upload(ctx, keyK, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)

	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, uploader)
	err = (&testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			StyleImage:    keyK,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	require.NoError(t, err, "アップロード済みオブジェクトへの参照は受理されなければならない")
}
