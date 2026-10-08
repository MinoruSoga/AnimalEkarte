import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";
import { ReservationDetailModal } from "./ReservationDetailModal";

vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));

// EMR-227: detail modal は petId から pet.status を引いて死亡区分を表示層へ渡す。
// このテストは QueryClientProvider なしで描画するためフックを固定する。
vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: undefined }),
}));

const getColorMock = vi.fn(() => ({
  style: { backgroundColor: "rgb(0, 0, 0)" },
  dotStyle: { backgroundColor: "rgb(0, 0, 0)" },
  hex: "#000000",
}));

vi.mock("../hooks/use-reservation-type-color-map", () => ({
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

describe("ReservationDetailModal — tentative (EMR-308)", () => {
  // tentative は start/end が意図日 00:00–00:15 JST の sentinel。実時刻ではないので
  // 時刻表示・ステータスセレクト・カルテ作成・編集の各導線を出さず、
  // 「予約を確定」（confirm-tentative 経路）のみを出す。
  // 意図日は未来に固定する — 過去日だと isExpiredTentativeReservation が真になり
  // ラベルが 期限超過・未対応 に変わるため。
  const tentativeReservation: Reservation = {
    ...baseReservation,
    id: "tentative-1",
    start: new Date("2999-06-29T00:00:00.000Z"),
    end: new Date("2999-06-29T00:15:00.000Z"),
    status: "tentative" as Reservation["status"],
    isTentative: true,
    intendedDate: "2999-06-29",
  };

  it("次回予定・未予約 ラベルと 時刻未設定 を表示する", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onCreateRecord={vi.fn()}
        onStatusChange={vi.fn()}
        onEdit={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.getByText("次回予定・未予約")).toBeInTheDocument();
    expect(screen.getByText("時刻未設定（予定日のみ）")).toBeInTheDocument();
  });

  it("tentative 行ではカルテ作成・編集・ステータス変更・削除を出さない", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onCreateRecord={vi.fn()}
        onStatusChange={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    // カルテ作成は appointmentId 付き遷移で sentinel 行に直接紐づき昇格しないため出さない
    expect(screen.queryByRole("button", { name: /カルテ作成/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /編集/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("意図日が過去の tentative は 期限超過・未対応 ラベルになる", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...tentativeReservation,
          start: new Date("2000-01-01T00:00:00.000Z"),
          end: new Date("2000-01-01T00:15:00.000Z"),
          intendedDate: "2000-01-01",
        }}
        onConfirmTentative={vi.fn()}
      />,
    );

    expect(screen.getByText("期限超過・未対応")).toBeInTheDocument();
    expect(screen.queryByText("次回予定・未予約")).not.toBeInTheDocument();
  });

  it("onConfirmTentative 指定時は予約を確定ボタンで予約を渡す", async () => {
    const user = userEvent.setup();
    const onConfirmTentative = vi.fn();
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={tentativeReservation}
        onConfirmTentative={onConfirmTentative}
      />,
    );

    await user.click(screen.getByRole("button", { name: "予約を確定" }));
    expect(onConfirmTentative).toHaveBeenCalledWith(tentativeReservation);
  });

  it("onConfirmTentative 省略時は確定ボタンを出さない（fail-closed）", () => {
    render(
      <ReservationDetailModal isOpen={true} onClose={vi.fn()} reservation={tentativeReservation} />,
    );

    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
  });

  it("昇格済み（hadIntendedDate のみ残る）行は通常予約として扱う", () => {
    render(
      <ReservationDetailModal
        isOpen={true}
        onClose={vi.fn()}
        reservation={{
          ...baseReservation,
          hadIntendedDate: true,
          intendedDate: "2026-05-29",
        }}
        onCreateRecord={vi.fn()}
        onEdit={vi.fn()}
        onStatusChange={vi.fn()}
        onConfirmTentative={vi.fn()}
      />,
    );

    // isTentative=false のため tentative ラベルも確定ボタンも出ず、カルテ作成は出る
    expect(screen.queryByText("次回予定・未予約")).not.toBeInTheDocument();
    expect(screen.queryByText("期限超過・未対応")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "予約を確定" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });
});
