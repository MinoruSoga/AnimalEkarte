import { describe, expect, it, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationTypeAvailableSlotsSection } from "./ReservationTypeAvailableSlotsSection";
import { ReservationTypeUnavailableTimesSection } from "./ReservationTypeUnavailableTimesSection";

afterEach(() => {
  server.resetHandlers();
});

describe("ReservationType 特定日の日付入力", () => {
  it("予約可能枠と予約不可時間の特定日入力が異なるアクセシブル名を持つ", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () => HttpResponse.json([])),
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json([]),
      ),
    );

    const user = userEvent.setup();
    render(
      <>
        <ReservationTypeAvailableSlotsSection clinicId="1" reservationTypeId="5" />
        <ReservationTypeUnavailableTimesSection clinicId="1" reservationTypeId="5" />
      </>,
      { wrapper: createTestWrapper({ router: true }) },
    );

    await user.click(await screen.findByRole("combobox", { name: "スロット種別" }));
    await user.click(await screen.findByRole("option", { name: "特定日" }));
    await user.click(await screen.findByRole("combobox", { name: "不可時間の種別" }));
    await user.click(await screen.findByRole("option", { name: "特定日" }));

    expect(screen.getByRole("textbox", { name: "予約可能枠の特定日" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "予約不可時間の特定日" })).toBeInTheDocument();
  });
});
