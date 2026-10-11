package trimming

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	trimmingConflictControlMarker = "TRIMMING_CLAIM_POSITIVE_CONTROL_OK"
	trimmingDeletedConflictMarker = "TRIMMING_DELETED_KEY_CONFLICT"
)

type existsFailAfterControlUploader struct {
	FileUploader
	targetKey string
	armed     bool
}

func (u *existsFailAfterControlUploader) Exists(ctx context.Context, key string) (bool, error) {
	if u.armed && key == u.targetKey {
		return false, errors.New("oracle injected exists failure")
	}
	exists, err := u.FileUploader.Exists(ctx, key)
	if exists {
		u.armed = true
	}
	return exists, err
}

func runTrimmingImageConflictOracleHelper() int {
	db, err := gorm.Open(postgres.Open(lifecycleTestDSN()), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper db open: %v\n", err)
		return 2
	}
	clinicID, err := strconv.ParseUint(os.Getenv("TRIMMING_HELPER_CLINIC_ID"), 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper clinic id: %v\n", err)
		return 2
	}
	apptID, err := strconv.ParseUint(os.Getenv("TRIMMING_HELPER_APPOINTMENT_ID"), 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper appointment id: %v\n", err)
		return 2
	}
	controlApptID, err := strconv.ParseUint(os.Getenv("TRIMMING_HELPER_CONTROL_APPOINTMENT_ID"), 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper control appointment id: %v\n", err)
		return 2
	}
	key := os.Getenv("TRIMMING_HELPER_KEY")
	controlKey := os.Getenv("TRIMMING_HELPER_CONTROL_KEY")
	uploadDir := os.Getenv("TRIMMING_HELPER_UPLOAD_DIR")
	if key == "" || controlKey == "" || uploadDir == "" {
		fmt.Fprintln(os.Stderr, "helper requires KEY, CONTROL_KEY and UPLOAD_DIR")
		return 2
	}
	uploader := infra.NewLocalUploader(uploadDir, "/uploads")
	var store FileUploader = uploader
	if os.Getenv("TRIMMING_HELPER_FORCE_EXISTS_ERROR") == "1" {
		store = &existsFailAfterControlUploader{FileUploader: uploader, targetKey: key}
	}
	repo := NewAppointmentTrimmingDetailRepository(db)
	bindLifecycleRepoUploader(repo, store)
	ctx := context.Background()

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.Create(persistence.WithTxValue(ctx, tx), &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: controlApptID,
			StyleImage:    controlKey,
			BWUnit:        model.BodyWeightUnitKg,
		})
	}); err != nil {
		fmt.Fprintf(os.Stderr, "positive control claim failed: %v\n", err)
		return 3
	}
	fmt.Fprintln(os.Stdout, trimmingConflictControlMarker)

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.Create(persistence.WithTxValue(ctx, tx), &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: apptID,
			StyleImage:    key,
			BWUnit:        model.BodyWeightUnitKg,
		})
	})
	if err == nil {
		fmt.Fprintln(os.Stderr, "DEFECT: deleted key claim succeeded in a fresh process")
		return 1
	}
	if !apperrors.IsConflict(err) {
		fmt.Fprintf(os.Stderr, "deleted key claim rejected with non-conflict error: %v\n", err)
		return 4
	}
	fmt.Fprintln(os.Stdout, trimmingDeletedConflictMarker)
	return 0
}

func oracleChildEnv(mode string, extra ...string) []string {
	var filtered []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "TRIMMING_HELPER_") ||
			strings.HasPrefix(kv, "TRIMMING_IMAGE_LIFECYCLE_HELPER=") {
			continue
		}
		filtered = append(filtered, kv)
	}
	return append(append(filtered, "TRIMMING_IMAGE_LIFECYCLE_HELPER="+mode), extra...)
}

func runSubprocessTypedConflictOracleTest(t *testing.T, mode string) {
	t.Helper()
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), supersededImageCleanupTimeout)
	defer cancel()
	clinicID := uint64(18)
	keyK := "trimming-images/18/typed-conflict.jpg"
	controlKey := "trimming-images/18/control-ok.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	controlAppt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	dir := t.TempDir()
	uploader := infra.NewLocalUploader(dir, "/uploads")
	_, err := uploader.Upload(ctx, keyK, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)
	_, err = uploader.Upload(ctx, controlKey, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)

	svc := newLifecycleService(t, db)
	h := newHandlerWithTrimmingSvc(svc)
	h.WithImageSupport(ImageDependencies{Uploader: uploader})
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, keyK)
	_, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(keyK)))
	require.True(t, os.IsNotExist(statErr),
		"前提: durable ガードがオブジェクトを削除済みであること")
	_, statErr = os.Stat(filepath.Join(dir, filepath.FromSlash(controlKey)))
	require.NoError(t, statErr,
		"前提: positive control 用オブジェクトは存在したままであること")

	exe, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = oracleChildEnv(mode,
		"TRIMMING_HELPER_CLINIC_ID="+strconv.FormatUint(clinicID, 10),
		"TRIMMING_HELPER_APPOINTMENT_ID="+strconv.FormatUint(appt.ID, 10),
		"TRIMMING_HELPER_CONTROL_APPOINTMENT_ID="+strconv.FormatUint(controlAppt.ID, 10),
		"TRIMMING_HELPER_KEY="+keyK,
		"TRIMMING_HELPER_CONTROL_KEY="+controlKey,
		"TRIMMING_HELPER_UPLOAD_DIR="+dir,
	)
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err,
		"子プロセスは exit 0 でなければならない (positive control + typed conflict): %s", out)
	require.Contains(t, string(out), trimmingConflictControlMarker,
		"positive control marker が無い場合は claim 経路を通っていない")
	require.Contains(t, string(out), trimmingDeletedConflictMarker,
		"typed conflict marker が無い場合は conflict 以外の拒否（または marker 欠落）")
}

func TestDeleteUnreferencedImageObject_SubprocessTypedConflictOracle(t *testing.T) {
	runSubprocessTypedConflictOracleTest(t, "conflict-claim")
}

func TestSubprocessConflictOracle_ForcedExistsErrorExits4(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	db := setupImageReferenceTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), supersededImageCleanupTimeout)
	defer cancel()
	clinicID := uint64(18)
	keyK := "trimming-images/18/forced-exists-err.jpg"
	controlKey := "trimming-images/18/forced-control.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	controlAppt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	dir := t.TempDir()
	uploader := infra.NewLocalUploader(dir, "/uploads")
	_, err := uploader.Upload(ctx, keyK, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)
	_, err = uploader.Upload(ctx, controlKey, strings.NewReader("img"), "image/jpeg")
	require.NoError(t, err)

	svc := newLifecycleService(t, db)
	h := newHandlerWithTrimmingSvc(svc)
	h.WithImageSupport(ImageDependencies{Uploader: uploader})
	h.bestEffortDeleteTrimmingImageKeys(ctx, clinicID, keyK)
	_, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(keyK)))
	require.True(t, os.IsNotExist(statErr), "前提: 対象オブジェクトは削除済み")

	exe, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = oracleChildEnv("conflict-claim",
		"TRIMMING_HELPER_CLINIC_ID="+strconv.FormatUint(clinicID, 10),
		"TRIMMING_HELPER_APPOINTMENT_ID="+strconv.FormatUint(appt.ID, 10),
		"TRIMMING_HELPER_CONTROL_APPOINTMENT_ID="+strconv.FormatUint(controlAppt.ID, 10),
		"TRIMMING_HELPER_KEY="+keyK,
		"TRIMMING_HELPER_CONTROL_KEY="+controlKey,
		"TRIMMING_HELPER_UPLOAD_DIR="+dir,
		"TRIMMING_HELPER_FORCE_EXISTS_ERROR=1",
	)
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "injected exists error の子は非ゼロ終了")
	require.Equal(t, 4, exitErr.ExitCode(),
		"conflict 以外の拒否は exit 4: %s", out)
	assert.Contains(t, string(out), trimmingConflictControlMarker,
		"positive control は成功して marker が出る")
	assert.NotContains(t, string(out), trimmingDeletedConflictMarker,
		"conflict 以外の拒否で conflict marker を出してはいけない")
}

func TestSubprocessConflictOracle_BrokenChildEmitsNoMarkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), supersededImageCleanupTimeout)
	defer cancel()
	exe, err := os.Executable()
	require.NoError(t, err)

	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = oracleChildEnv("conflict-claim")
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "設定不全の子は非ゼロで終了しなければならない")
	require.Equal(t, 2, exitErr.ExitCode(), "設定不全は exit 2: %s", out)
	assert.NotContains(t, string(out), trimmingDeletedConflictMarker)
	assert.NotContains(t, string(out), trimmingConflictControlMarker)
}

func TestSubprocessConflictOracle_UnreachableDBEmitsNoMarkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), supersededImageCleanupTimeout)
	defer cancel()
	exe, err := os.Executable()
	require.NoError(t, err)

	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = oracleChildEnv("conflict-claim",
		"TRIMMING_HELPER_CLINIC_ID=18",
		"TRIMMING_HELPER_APPOINTMENT_ID=1",
		"TRIMMING_HELPER_CONTROL_APPOINTMENT_ID=2",
		"TRIMMING_HELPER_KEY=trimming-images/18/typed-conflict.jpg",
		"TRIMMING_HELPER_CONTROL_KEY=trimming-images/18/control-ok.jpg",
		"TRIMMING_HELPER_UPLOAD_DIR=/nonexistent-dir-for-oracle",
		"TEST_DATABASE_URL=postgres://ekarte_user:ekarte_password@127.0.0.1:1/ekarte_db_test?sslmode=disable",
	)
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "DB 接続不能の子は非ゼロで終了しなければならない")
	require.Equal(t, 2, exitErr.ExitCode(), "DB 接続不能は exit 2: %s", out)
	assert.NotContains(t, string(out), trimmingDeletedConflictMarker)
	assert.NotContains(t, string(out), trimmingConflictControlMarker)
}
