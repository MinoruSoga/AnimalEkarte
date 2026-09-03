import { describe, it, expect, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationRouteSelect } from "./ReservationRouteSelect";
import type { ReservationRoute } from "@/types/reservation-route";

const RESERVATION_ID = "res-test-1";

// transformReservation が通せる最低限レスポンス
const minimalReservationApiResponse = {
  id: 1,
  clinic_id: 1,
  start_time: "2026-05-04T10:00:00Z",
  end_time: "2026-05-04T10:30:00Z",
  visit_type: "revisit",
  reservation_type_id: 1,
  is_designated: false,
  status: "confirmed",
  notes: "",
  created_at: "2026-05-04T00:00:00Z",
  updated_at: "2026-05-04T00:00:00Z",
  source: "manual",
  is_staff_delegated: false,
  customer_fields: {},
};

function createWrapper() {
  return createTestWrapper({ router: true });
}

function renderSelect(value: ReservationRoute | null, disabled = false) {
  render(
    <ReservationRouteSelect
      reservationId={RESERVATION_ID}
      value={value}
      disabled={disabled}
    />,
    { wrapper: createWrapper() }
  );
}

afterEach(() => {
  server.resetHandlers();
});

// ─────────────────────────────────────────────────────────────
// A: 初期表示
// ─────────────────────────────────────────────────────────────

describe("ReservationRouteSelect — A: 初期表示", () => {
  it("value=null → 「未選択」プレースホルダーが表示される", () => {
    renderSelect(null);
    expect(screen.getByTestId("reservation-route-trigger")).toBeInTheDocument();
    expect(screen.getByText("未選択")).toBeInTheDocument();
  });

  it("value='line' → 「LINE」が表示される", () => {
    renderSelect("line");
    expect(screen.getByText("LINE")).toBeInTheDocument();
  });

  it("value='phone' → 「電話」が表示される", () => {
    renderSelect("phone");
    expect(screen.getByText("電話")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 選択肢の表示
// ─────────────────────────────────────────────────────────────

describe("ReservationRouteSelect — B: 選択肢の表示", () => {
  it("トリガークリック → 5選択肢（LINE/電話/受付/診察室/記録入力）がすべて表示される", async () => {
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "LINE" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "電話" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "受付" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "診察室" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "記録入力" })).toBeInTheDocument();
    });
  });

  it("value が設定済みの場合「クリア」選択肢が表示される", async () => {
    renderSelect("reception");
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "クリア" })).toBeInTheDocument();
    });
  });

  it("value=null の場合「クリア」選択肢は表示されない", async () => {
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "LINE" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "クリア" })).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: PATCH エンドポイント呼び出し
// ─────────────────────────────────────────────────────────────

describe("ReservationRouteSelect — C: PATCH エンドポイント呼び出し", () => {
  it("「受付」を選択 → PATCH に route:'reception' が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/reservations/${RESERVATION_ID}/reservation-route`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({ ...minimalReservationApiResponse, reservation_route: "reception" });
        }
      )
    );
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "受付" }));
    await user.click(screen.getByRole("option", { name: "受付" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ route: "reception" }));
    });
  });

  it("「診察室」を選択 → PATCH に route:'exam_room' が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/reservations/${RESERVATION_ID}/reservation-route`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({ ...minimalReservationApiResponse, reservation_route: "exam_room" });
        }
      )
    );
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "診察室" }));
    await user.click(screen.getByRole("option", { name: "診察室" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ route: "exam_room" }));
    });
  });

  it("「クリア」選択 → PATCH に route:'' が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/reservations/${RESERVATION_ID}/reservation-route`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({ ...minimalReservationApiResponse, reservation_route: null });
        }
      )
    );
    renderSelect("line");
    const user = userEvent.setup();
    await user.click(screen.getByTestId("reservation-route-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "クリア" }));
    await user.click(screen.getByRole("option", { name: "クリア" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ route: "" }));
    });
  });
});

// ─────────────────────────────────────────────────────────────
// D: disabled 状態
// ─────────────────────────────────────────────────────────────

describe("ReservationRouteSelect — D: disabled 状態", () => {
  it("disabled=true → Select トリガーが disabled 属性を持つ", () => {
    renderSelect(null, true);
    const trigger = screen.getByTestId("reservation-route-trigger");
    expect(trigger).toBeDisabled();
  });
});
