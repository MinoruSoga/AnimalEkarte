import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";
import { ReservationDetailModal } from "./ReservationDetailModal";

vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));

const getColorMock = vi.fn(() => ({
  style: { backgroundColor: "rgb(0, 0, 0)" },
  dotStyle: { backgroundColor: "rgb(0, 0, 0)" },
  hex: "#000000",
}));

vi.mock("@/hooks/use-reservation-type-color-map", () => ({
  useReservationTypeColorMap: () => ({
    getColor: getColorMock,
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
  beforeEach(() => {
    getColorMock.mockReset();
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgb(0, 0, 0)" },
      dotStyle: { backgroundColor: "rgb(0, 0, 0)" },
      hex: "#000000",
    });
  });

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

  it("ペットホテル予約ではマスタ名の部分一致で入院・ホテル登録ボタンを表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          type: "ペットホテル",
          category: "general",
          reservationTypeId: "12",
        }}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /入院・ホテル登録/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
  });

  it("無効区分では DialogTitle と予約区分に（無効）を付ける (BUG-016)", () => {
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgba(255, 102, 0, 0.1)" },
      dotStyle: { backgroundColor: "#ff6600" },
      hex: "#ff6600",
      isInactive: true,
    });

    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "一般診察（無効）" })).toBeInTheDocument();
    const typeRow = screen.getByText("予約区分").parentElement;
    expect(typeRow?.textContent).toContain("一般診察（無効）");
  });

  it("有効区分では（無効）を付けない (BUG-016)", () => {
    getColorMock.mockReturnValue({
      style: { backgroundColor: "rgba(17, 17, 17, 0.1)" },
      dotStyle: { backgroundColor: "#111111" },
      hex: "#111111",
      isInactive: false,
    });

    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={baseReservation}
        onCreateRecord={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "一般診察" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /（無効）/ })).not.toBeInTheDocument();
    const typeRow = screen.getByText("予約区分").parentElement;
    expect(typeRow?.textContent).toContain("一般診察");
    expect(typeRow?.textContent).not.toContain("（無効）");
  });
});
