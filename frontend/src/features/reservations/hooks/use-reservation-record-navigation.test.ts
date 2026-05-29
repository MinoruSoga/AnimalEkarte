import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";

import { useReservationRecordNavigation } from "./use-reservation-record-navigation";

const baseReservation: Reservation = {
  id: "88",
  start: new Date("2026-05-29T03:30:00.000Z"),
  end: new Date("2026-05-29T04:00:00.000Z"),
  ownerName: "山田花子",
  petName: "ポチ",
  visitType: "first",
  type: "シャンプー",
  category: "trimming",
  reservationTypeId: "3",
  doctor: "",
  isDesignated: false,
  status: "checked_in",
  source: "manual",
  reservationRoute: "reception",
};

describe("useReservationRecordNavigation", () => {
  it("petId 未確定の予約でもペット選択へ appointmentId と visitDate を引き継ぐ", () => {
    const navigate = vi.fn();
    const { result } = renderHook(() => useReservationRecordNavigation({ navigate }));

    act(() => {
      result.current.handleCreateRecord(baseReservation);
    });

    expect(result.current.petSelectConfirmOpen).toBe(true);

    act(() => {
      result.current.handlePetSelectConfirm();
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=88&visitDate=2026-05-29",
      {
        state: {
          from: "/reservations",
          appointmentId: "88",
          visitDate: "2026-05-29",
        },
      },
    );
  });

  it("petId 確定済みの予約は作成画面へ appointmentId と visitDate を引き継ぐ", () => {
    const navigate = vi.fn();
    const { result } = renderHook(() => useReservationRecordNavigation({ navigate }));

    act(() => {
      result.current.handleCreateRecord({ ...baseReservation, petId: "10" });
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/new?petId=10&appointmentId=88&visitDate=2026-05-29",
      {
        state: {
          from: "/reservations",
          appointmentId: "88",
          visitDate: "2026-05-29",
        },
      },
    );
  });

  it("ペットホテル予約はマスタ名の部分一致で入院登録へ遷移する", () => {
    const navigate = vi.fn();
    const { result } = renderHook(() => useReservationRecordNavigation({ navigate }));

    act(() => {
      result.current.handleCreateRecord({
        ...baseReservation,
        type: "ペットホテル",
        category: "general",
        petId: "10",
      });
    });

    expect(navigate).toHaveBeenCalledWith(
      "/hospitalization/new?petId=10&appointmentId=88&visitDate=2026-05-29",
      {
        state: {
          from: "/reservations",
          appointmentId: "88",
          visitDate: "2026-05-29",
        },
      },
    );
  });
});
