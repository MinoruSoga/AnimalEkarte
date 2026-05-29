import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";
import { ReservationDetailModal } from "./ReservationDetailModal";

vi.mock("@/features/owners", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));

vi.mock("@/hooks/use-reservation-type-color-map", () => ({
  useReservationTypeColorMap: () => ({
    getColor: () => ({ dotStyle: { backgroundColor: "rgb(0, 0, 0)" } }),
  }),
}));

const baseReservation: Reservation = {
  id: "88",
  start: new Date("2026-05-29T03:30:00.000Z"),
  end: new Date("2026-05-29T04:00:00.000Z"),
  ownerName: "山田花子",
  petName: "ポチ",
  petType: "トイプードル",
  visitType: "revisit",
  type: "一般診察",
  category: "general",
  reservationTypeId: "1",
  doctor: "担当者A",
  doctorId: "33",
  isDesignated: false,
  status: "checked_in",
  notes: undefined,
  petId: "10",
  ownerId: "20",
  source: "manual",
  reservationRoute: "reception",
};

describe("ReservationDetailModal", () => {
  it("通常診療予約ではカルテ作成ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });

  it("トリミングカテゴリの予約では区分名に依存せずトリミング記録作成ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          type: "シャンプーコース",
          category: "trimming",
          reservationTypeId: "9",
        }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /トリミング記録作成/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });
});
