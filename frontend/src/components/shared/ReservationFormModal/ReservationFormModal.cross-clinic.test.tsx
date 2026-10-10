import { describe, it, expect, afterEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import { server } from "@/testing/mocks/node";
import { http, HttpResponse } from "msw";
import { ReservationFormModal } from "./ReservationFormModal";
import type { Reservation } from "@/types";
import { createWrapper, silentApiHandlers, noop } from "./ReservationFormModal.test-helpers";
import { setStoredClinicId } from "@/lib/current-clinic";

/**
 * Codex PR606 再指摘 #1: 拠点横断 tentative 確定・編集フォームが
 * 「グローバル選択中の医院」ではなく「対象行の clinicId」で
 * clinic-scoped query を発行することを X-Clinic-ID / path で固定する。
 * localStorage の選択中医院は clinic-1、対象行は clinic-2 にして
 * 既定フォールバックではなく上書きが効いていることを区別できるようにする。
 */

const TARGET_CLINIC = "clinic-2";
const OTHER_CLINIC = "clinic-1";

// ヘッダ capture: endpoint label → 受信した X-Clinic-ID
const seenHeaders = new Map<string, string | null>();
// path capture: endpoint label → :clinicId パラメータ
const seenPathClinicIds = new Map<string, string>();

const captureHeader =
  (label: string, body: unknown = []) =>
  ({ request }: { request: Request }) => {
    seenHeaders.set(label, request.headers.get("x-clinic-id"));
    return HttpResponse.json(body);
  };

const PET_WIRE = {
  id: 10,
  clinic_id: 2,
  owner_id: 20,
  name: "ポチ",
  status: "alive",
  version: 1,
  phone: "090-0000-0000",
  owner: { id: 20, name: "山田", phone: "090-0000-0000" },
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const crossClinicHandlers = [
  // MSW は先着マッチ — capture ハンドラを silentApiHandlers より先に置く。
  // （silent 側の同一 URL ハンドラは到達しないが、animal-species/pets-list は有効）
  http.get("/api/v1/clinic-holidays", captureHeader("clinic-holidays")),
  http.get("/api/v1/masters/reservation-types", captureHeader("reservation-types")),
  http.get("/api/v1/masters/staffs", captureHeader("staffs")),
  http.get("/api/v1/shifts/on-duty-staffs", captureHeader("on-duty-staffs")),
  http.get("/api/v1/reservations/available-times", captureHeader("available-times")),
  http.get(
    "/api/v1/masters/reservation-types/:typeId/unavailable-times",
    captureHeader("unavailable-times"),
  ),
  // path 埋め込み系も backend 同様に header で scope する:
  // X-Clinic-ID と :clinicId が不一致なら 403（interceptor 補完ミスを潰す）。
  http.get("/api/v1/clinics/:clinicId/reservation-staffs", ({ params, request }) => {
    seenPathClinicIds.set("reservation-staffs", String(params.clinicId));
    seenHeaders.set("reservation-staffs", request.headers.get("x-clinic-id"));
    if (request.headers.get("x-clinic-id") !== String(params.clinicId)) {
      return HttpResponse.json({ error: "clinic scope mismatch" }, { status: 403 });
    }
    return HttpResponse.json([]);
  }),
  http.get("/api/v1/pets/:id", captureHeader("pet-detail", PET_WIRE)),
  http.get("/api/v1/clinics/:clinicId/owners/:ownerId/lstep/tags", ({ params, request }) => {
    seenPathClinicIds.set("owner-line-tags", String(params.clinicId));
    seenHeaders.set("owner-line-tags", request.headers.get("x-clinic-id"));
    if (request.headers.get("x-clinic-id") !== String(params.clinicId)) {
      return HttpResponse.json({ error: "clinic scope mismatch" }, { status: 403 });
    }
    return HttpResponse.json({
      line_user_id: null,
      is_linked: false,
      lstep_opt_out: false,
      tags: [],
      fetched_at: "2024-01-01T00:00:00Z",
    });
  }),
  http.get("/api/v1/reservations/:id", captureHeader("reservation-detail", {})),
  ...silentApiHandlers,
];

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem("auth_current_clinic:v1");
  seenHeaders.clear();
  seenPathClinicIds.clear();
});

const futureStart = () => new Date(2999, 5, 20, 10, 0);

const confirmInitialData: Partial<Reservation> = {
  petId: "10",
  start: futureStart(),
  end: new Date(2999, 5, 20, 11, 0),
  visitType: "revisit",
  type: "1",
  doctor: "",
  isDesignated: false,
  status: "confirmed",
};

describe("ReservationFormModal — 拠点横断 clinicId スコープ（Codex 再指摘 #1）", () => {
  it("確定モード: clinic-scoped query は全て対象行の clinicId で発行される", async () => {
    server.use(...crossClinicHandlers);
    // 選択中医院は別の医院（clinic-1）— 既定フォールバックなら全て clinic-1 になる
    setStoredClinicId(OTHER_CLINIC);

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onConfirmTentative={noop}
        tentativeConfirmId="t-1"
        initialData={confirmInitialData}
        canEdit={true}
        clinicId={TARGET_CLINIC}
      />,
      { wrapper: createWrapper() },
    );

    // 患者取得 → owner tag 取得まで連鎖するのを待つ
    await waitFor(() => expect(seenHeaders.get("pet-detail")).toBe(TARGET_CLINIC));
    await waitFor(() => expect(seenPathClinicIds.get("owner-line-tags")).toBe(TARGET_CLINIC));

    // path 埋め込み系は header も対象行の医院と一致する（backend は header で scope）
    expect(seenHeaders.get("reservation-staffs")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("owner-line-tags")).toBe(TARGET_CLINIC);

    // ヘッダ系エンドポイント（master/availability）は全て対象行の医院
    expect(seenHeaders.get("clinic-holidays")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("reservation-types")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("staffs")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("on-duty-staffs")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("available-times")).toBe(TARGET_CLINIC);
    expect(seenHeaders.get("unavailable-times")).toBe(TARGET_CLINIC);

    // path 埋め込み系エンドポイントも対象行の医院
    expect(seenPathClinicIds.get("reservation-staffs")).toBe(TARGET_CLINIC);
  }, 20000);

  it("編集モード: GET /reservations/:id も対象行の clinicId で発行される", async () => {
    server.use(...crossClinicHandlers);
    setStoredClinicId(OTHER_CLINIC);

    const editInitialData: Partial<Reservation> = {
      ...confirmInitialData,
      id: "42",
    };

    render(
      <ReservationFormModal
        isOpen={true}
        onClose={noop}
        onSave={noop}
        initialData={editInitialData}
        canEdit={true}
        clinicId={TARGET_CLINIC}
      />,
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(seenHeaders.get("reservation-detail")).toBe(TARGET_CLINIC));
  }, 20000);
});
