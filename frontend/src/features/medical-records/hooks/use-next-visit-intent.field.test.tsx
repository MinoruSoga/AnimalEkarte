import { cleanup, render, screen, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import { toJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "@/lib/transforms/reservation";

import { NextVisitDateField } from "../components/NextVisitDateField";
import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/hooks/use-get-reservations", () => ({ useGetReservations: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({ usePermission: vi.fn() }));

function appointment(instant: string, status: Reservation["status"] = "confirmed"): Reservation {
  return {
    id: instant,
    status,
    start: toJSTWallDate(new Date(instant)),
    end: toJSTWallDate(new Date(instant)),
    ownerName: "飼主",
    petName: "患者",
    petType: undefined,
    visitType: "revisit",
    type: "診療",
    category: "general",
    reservationTypeId: undefined,
    doctor: "担当医",
    doctorId: undefined,
    isDesignated: false,
    notes: undefined,
    petId: "10",
    ownerId: "3",
    source: "manual",
    reservationRoute: null,
    clinicId: "2",
  };
}
function tentative(date = "2026-10-15"): Reservation & { intendedDate: string; updatedAt: string } {
  return {
    ...appointment(`${date}T00:00:00+09:00`),
    id: "intent",
    status: "tentative" as Reservation["status"],
    intendedDate: date,
    updatedAt: `${date}T00:00:00+09:00`,
  };
}
function mockReservations(items: Reservation[]) {
  vi.mocked(useGetReservations).mockReturnValue({ data: items } as ReturnType<
    typeof useGetReservations
  >);
}
const setIntentMutation = { mutateAsync: vi.fn().mockResolvedValue({}) };
const clearIntentMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };

function IntentField() {
  const intent = useNextVisitIntent({
    petId: "10",
    clinicId: "2",
    isPetDeceased: false,
    setIntentMutation,
    clearIntentMutation,
  });
  return (
    <NextVisitDateField
      value={intent.intentFieldValue}
      onChange={intent.handleIntentDateChange}
      setDisabled={!intent.canSetIntent}
      clearDisabled={!intent.canClearIntent}
    />
  );
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(new Date("2026-10-14T00:00:00Z"));
  vi.clearAllMocks();
  mockReservations([]);
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: false,
  });
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("useNextVisitIntent + NextVisitDateField mutation gating", () => {
  it("未編集の focus→blur では PUT を発行しない", async () => {
    const user = userEvent.setup();
    mockReservations([tentative("2026-10-15")]);
    render(<IntentField />);
    const input = screen.getByRole("textbox", { name: "次回来院推奨日" });
    await user.click(input);
    await user.tab();
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("別の有効日へ編集すると PUT を1回発行する", async () => {
    const user = userEvent.setup();
    mockReservations([tentative("2026-10-15")]);
    render(<IntentField />);
    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    await user.clear(input);
    await user.type(input, "2026/10/20");
    await user.tab();
    await waitFor(() => expect(setIntentMutation.mutateAsync).toHaveBeenCalledTimes(1));
    expect(setIntentMutation.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ intendedDate: "2026-10-20" }),
    );
  });

  it("編集後 Enter→blur でも親再レンダー中の重複 commit で PUT は1回に留まる", async () => {
    const user = userEvent.setup();
    mockReservations([tentative("2026-10-15")]);
    render(<IntentField />);
    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    await user.clear(input);
    await user.type(input, "2026/10/20");
    await user.keyboard("{Enter}");
    await user.tab();
    await waitFor(() => expect(setIntentMutation.mutateAsync).toHaveBeenCalledTimes(1));
    expect(setIntentMutation.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ intendedDate: "2026-10-20" }),
    );
  });

  it("クリア権限があるときクリア操作は DELETE を1回発行する", async () => {
    const user = userEvent.setup();
    mockReservations([tentative("2026-10-15")]);
    render(<IntentField />);
    await user.click(screen.getByRole("button", { name: "クリア" }));
    await waitFor(() => expect(clearIntentMutation.mutateAsync).toHaveBeenCalledTimes(1));
  });

  it("canEdit なしではクリアも設定も mutation を発行しない", async () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: false,
      canDelete: false,
    });
    const user = userEvent.setup();
    mockReservations([tentative("2026-10-15")]);
    render(<IntentField />);
    const clearButton = screen.getByRole("button", { name: "クリア" });
    expect(clearButton).toBeDisabled();
    await user.click(clearButton);
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });
});
