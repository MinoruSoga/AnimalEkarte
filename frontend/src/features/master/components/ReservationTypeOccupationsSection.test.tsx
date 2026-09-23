import { describe, expect, it, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationTypeOccupationsSection } from "./ReservationTypeOccupationsSection";

afterEach(() => {
  server.resetHandlers();
});

function makeOccupation(id: number, name: string) {
  return {
    id,
    clinic_id: 1,
    name,
    description: "",
    sort_order: id,
    is_active: true,
    created_at: "2026-09-23T00:00:00Z",
    updated_at: "2026-09-23T00:00:00Z",
  };
}

describe("ReservationTypeOccupationsSection", () => {
  it("BE が裸配列を返しても紐付け職種バッジを表示する（EMR-209）", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/occupations", () =>
        HttpResponse.json([
          {
            id: 10,
            clinic_id: 1,
            reservation_type_id: 5,
            occupation_id: 110,
            occupation: makeOccupation(110, "V04職種"),
            created_at: "2026-09-23T00:00:00Z",
          },
        ]),
      ),
      http.get("/api/v1/masters/occupations", () =>
        HttpResponse.json([makeOccupation(110, "V04職種")]),
      ),
    );

    render(<ReservationTypeOccupationsSection clinicId="1" reservationTypeId="5" />, {
      wrapper: createTestWrapper(),
    });

    expect(await screen.findByText("V04職種")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "V04職種 の紐付けを解除" })).toBeInTheDocument();
  });
});
