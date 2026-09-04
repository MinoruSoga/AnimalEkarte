import { describe, expect, it, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationTypeAvailableSlotsSection } from "./ReservationTypeAvailableSlotsSection";

function createWrapper() {
  return createTestWrapper({ router: true });
}

afterEach(() => {
  server.resetHandlers();
});

describe("ReservationTypeAvailableSlotsSection", () => {
  it("予約可能枠を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json([
          {
            id: 1,
            clinic_id: 1,
            reservation_type_id: 5,
            available_type: "weekly",
            day_of_week: 1,
            start_time: "09:45",
            is_active: true,
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
          {
            id: 2,
            clinic_id: 1,
            reservation_type_id: 5,
            available_type: "specific",
            specific_date: "2026-06-01",
            start_time: "14:00",
            is_active: true,
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
        ]),
      ),
    );

    render(<ReservationTypeAvailableSlotsSection clinicId="1" reservationTypeId="5" />, {
      wrapper: createWrapper(),
    });

    expect(await screen.findByText("予約可能枠")).toBeInTheDocument();
    expect(await screen.findByText("毎週月曜日")).toBeInTheDocument();
    expect(screen.getByText("2026-06-01")).toBeInTheDocument();
    expect(screen.getAllByText("09:45").length).toBeGreaterThan(0);
    expect(screen.getAllByText("14:00").length).toBeGreaterThan(0);
  });
});
