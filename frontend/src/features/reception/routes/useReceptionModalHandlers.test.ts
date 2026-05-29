import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useReceptionModalHandlers } from "./useReceptionModalHandlers";
import type { ReceptionAppointment } from "../api/types";

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-06-01",
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

function renderHandlers(updateAppointment = vi.fn()) {
  return renderHook(() => useReceptionModalHandlers({
    advanceStatus: vi.fn(),
    cancelAppointment: vi.fn(),
    updateAppointment,
  }));
}

describe("useReceptionModalHandlers", () => {
  it("受付予約編集では appointment の visitDate と time から開始日時を作る", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });

    expect(result.current.editingAppointment?.start?.toISOString()).toBe("2026-06-01T00:45:00.000Z");
    expect(result.current.editingAppointment?.end?.toISOString()).toBe("2026-06-01T01:45:00.000Z");
  });

  it("編集保存後のローカル appointment も JST 基準の visitDate を保持する", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date("2026-06-01T09:45:00+09:00"),
          visitType: "revisit",
          type: "一般診察",
          doctor: "担当者A",
        },
        [],
      );
    });

    expect(updateAppointment).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "101",
        time: "09:45",
        visitDate: "2026-06-01",
      }),
    );
  });
});
