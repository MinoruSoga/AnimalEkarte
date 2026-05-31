import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StaffExcludedReservationTypesSection } from "./StaffSidePanelSections";
import type { ReservationType } from "../api/reservation-types";

function reservationType(overrides: Partial<ReservationType>): ReservationType {
  return {
    id: "1",
    clinicId: "1",
    name: "一般診療",
    color: "#111111",
    isActive: true,
    description: "",
    sortOrder: 1,
    groupId: undefined,
    createdAt: "2026-05-29T00:00:00Z",
    updatedAt: "2026-05-29T00:00:00Z",
    reservationDisplayName: "",
    durationMinutes: 30,
    shortName: "",
    showShortName: false,
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    reservationDayOption: "none",
    isInternal: false,
    category: "general",
    ...overrides,
  };
}

describe("StaffExcludedReservationTypesSection", () => {
  it("対応可能コースとしてカテゴリ別に表示し、チェック状態を更新できる", async () => {
    const onToggle = vi.fn();
    const user = userEvent.setup({ delay: null });
    const medical = reservationType({ id: "1", name: "一般診療", category: "general" });
    const trimming = reservationType({ id: "2", name: "シャンプー", category: "trimming" });

    render(
      <StaffExcludedReservationTypesSection
        activeReservationTypes={[medical, trimming]}
        allReservationTypes={[medical, trimming]}
        capableIdSet={new Set(["1"])}
        isNew={false}
        onToggle={onToggle}
      />,
    );

    expect(screen.getByText("対応可能コース")).toBeInTheDocument();
    expect(screen.getByText("診療")).toBeInTheDocument();
    expect(screen.getByText("トリミング")).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: /一般診療/ })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: /シャンプー/ })).not.toBeChecked();

    await user.click(screen.getByRole("checkbox", { name: /シャンプー/ }));

    expect(onToggle).toHaveBeenCalledWith("2", true);
  });
});
