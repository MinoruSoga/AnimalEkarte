package reservation

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// lastHandlerSegment は gin の RouteInfo.Handler（full path 付き関数名）から末尾の
// メソッド名だけを取り出す（internal/medicalrecord の同名ヘルパーと同一実装）。
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}

// TestRegisterRoutes_Snapshot は reservation 側の BE9-2C route-snapshot 回帰チェック
// （internal/medicalrecord/routes_snapshot_test.go の先例を踏襲）。R①〜R④ で
// internal/handler/testdata/route_snapshot.golden から 48 route（reservation-type-groups 6 +
// reservation-types 6 + unavailable-times 3 + available-slots 3 + occupations 3 +
// LINE 管理用予約区分 7+予約スタッフ 7+スケジュール 3+予約CRUD 7+LINE管理予約 3）を drop し、本 package の RegisterRoutes が登録する。
// permission 引数は capture されない — RBAC parity は routes.go の逐語転記でレビュー担保。
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewReservationTypeHandler(nil, nil, nil, nil),
		NewReservationTypeGroupHandler(nil),
		NewReservationTypeLiffHandler(nil),
		NewReservationStaffHandler(nil),
		NewReservationScheduleHandler(nil),
		NewReservationHandler(nil, nil, nil, nil),
		NewReservationAdminHandler(nil, nil),
		noopPermission,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	want := "" +
		"DELETE /api/v1/clinics/:clinic_id/reservation-staffs/:staffId DeleteReservationStaff\n" +
		"DELETE /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/schedules/:date DeleteReservationSchedule\n" +
		"DELETE /api/v1/clinics/:clinic_id/reservation-types/:id DeleteReservationTypeLiff\n" +
		"DELETE /api/v1/clinics/:clinic_id/reservations/:reservationId DeleteReservationAdmin\n" +
		"DELETE /api/v1/masters/reservation-type-groups/:id DeleteReservationTypeGroup\n" +
		"DELETE /api/v1/masters/reservation-types/:id DeleteReservationType\n" +
		"DELETE /api/v1/masters/reservation-types/:id/available-slots/:available_slot_id DeleteAvailableSlot\n" +
		"DELETE /api/v1/masters/reservation-types/:id/occupations/:occupation_id UnlinkReservationTypeOccupation\n" +
		"DELETE /api/v1/masters/reservation-types/:id/unavailable-times/:unavailable_time_id DeleteUnavailableTime\n" +
		"DELETE /api/v1/reservations/:id DeleteReservation\n" +
		"GET /api/v1/clinics/:clinic_id/reservation-staffs ListReservationStaffs\n" +
		"GET /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/schedules ListReservationSchedules\n" +
		"GET /api/v1/clinics/:clinic_id/reservation-types ListReservationTypeLiffs\n" +
		"GET /api/v1/clinics/:clinic_id/reservations ListReservationsAdmin\n" +
		"GET /api/v1/masters/reservation-type-groups ListReservationTypeGroups\n" +
		"GET /api/v1/masters/reservation-type-groups/:id GetReservationTypeGroup\n" +
		"GET /api/v1/masters/reservation-types ListReservationTypes\n" +
		"GET /api/v1/masters/reservation-types/:id GetReservationType\n" +
		"GET /api/v1/masters/reservation-types/:id/available-slots ListAvailableSlots\n" +
		"GET /api/v1/masters/reservation-types/:id/occupations ListReservationTypeOccupations\n" +
		"GET /api/v1/masters/reservation-types/:id/unavailable-times ListUnavailableTimes\n" +
		"GET /api/v1/reservations ListReservations\n" +
		"GET /api/v1/reservations/:id GetReservation\n" +
		"GET /api/v1/reservations/available-times GetReservationAvailableTimes\n" +
		"PATCH /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/sort-order UpdateReservationStaffSortOrder\n" +
		"PATCH /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/status UpdateReservationStaffStatus\n" +
		"PATCH /api/v1/clinics/:clinic_id/reservation-types/:id/sort-order UpdateReservationTypeLiffSortOrder\n" +
		"PATCH /api/v1/clinics/:clinic_id/reservation-types/:id/status UpdateReservationTypeLiffStatus\n" +
		"PATCH /api/v1/masters/reservation-type-groups/:id UpdateReservationTypeGroup\n" +
		"PATCH /api/v1/masters/reservation-type-groups/reorder ReorderReservationTypeGroups\n" +
		"PATCH /api/v1/masters/reservation-types/:id UpdateReservationType\n" +
		"PATCH /api/v1/masters/reservation-types/reorder ReorderReservationTypes\n" +
		"PATCH /api/v1/reservations/:id UpdateReservation\n" +
		"PATCH /api/v1/reservations/:id/reservation-route UpdateReservationReservationRoute\n" +
		"POST /api/v1/clinics/:clinic_id/reservation-staffs CreateReservationStaff\n" +
		"POST /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/image UploadReservationStaffImage\n" +
		"POST /api/v1/clinics/:clinic_id/reservation-types CreateReservationTypeLiff\n" +
		"POST /api/v1/clinics/:clinic_id/reservation-types/:id/image UploadReservationTypeLiffImage\n" +
		"POST /api/v1/clinics/:clinic_id/reservations CreateReservationAdmin\n" +
		"POST /api/v1/masters/reservation-type-groups CreateReservationTypeGroup\n" +
		"POST /api/v1/masters/reservation-types CreateReservationType\n" +
		"POST /api/v1/masters/reservation-types/:id/available-slots CreateAvailableSlot\n" +
		"POST /api/v1/masters/reservation-types/:id/occupations LinkReservationTypeOccupation\n" +
		"POST /api/v1/masters/reservation-types/:id/unavailable-times CreateUnavailableTime\n" +
		"POST /api/v1/reservations CreateReservation\n" +
		"PUT /api/v1/clinics/:clinic_id/reservation-staffs/:staffId UpdateReservationStaff\n" +
		"PUT /api/v1/clinics/:clinic_id/reservation-staffs/:staffId/schedules/:date UpsertReservationSchedule\n" +
		"PUT /api/v1/clinics/:clinic_id/reservation-types/:id UpdateReservationTypeLiff\n"

	assert.Equal(t, want, got)
}
