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

function renderSection({
  types = [
    reservationType({ id: "1", name: "一般診療", category: "general" }),
    reservationType({ id: "2", name: "シャンプー", category: "trimming" }),
  ],
  capableIdSet = new Set(["1"]),
  isNew = false,
  onToggle = vi.fn(),
}: {
  types?: ReservationType[];
  capableIdSet?: Set<string>;
  isNew?: boolean;
  onToggle?: (id: string, checked: boolean) => void;
} = {}) {
  render(
    <StaffExcludedReservationTypesSection
      activeReservationTypes={types.filter((type) => type.isActive)}
      allReservationTypes={types}
      capableIdSet={capableIdSet}
      isNew={isNew}
      onToggle={onToggle}
    />,
  );
  return { onToggle };
}

describe("StaffExcludedReservationTypesSection", () => {
  it("対応可能 view: カテゴリ別に表示し、チェック状態を更新できる", async () => {
    const { onToggle } = renderSection();
    const user = userEvent.setup({ delay: null });

    expect(screen.getByText("対応区分")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "対応可能" })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "対応不可" })).toBeInTheDocument();
    expect(screen.getByText("診療")).toBeInTheDocument();
    expect(screen.getByText("トリミング")).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: /一般診療/ })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: /シャンプー/ })).not.toBeChecked();

    await user.click(screen.getByRole("checkbox", { name: /シャンプー/ }));

    expect(onToggle).toHaveBeenCalledWith("2", true);
  });

  it("対応不可 view は補集合（有効区分 − capableIds）をチェック済み表示し、切替だけでは書き込まない", async () => {
    const { onToggle } = renderSection();
    const user = userEvent.setup({ delay: null });

    await user.click(screen.getByRole("radio", { name: "対応不可" }));

    // capable = {"1"} なので、対応不可表示では 一般診療=未チェック / シャンプー=チェック
    // （チェック済み集合が有効区分 − capableIds の補集合そのものになる）
    expect(screen.getByRole("checkbox", { name: /一般診療/ })).not.toBeChecked();
    expect(screen.getByRole("checkbox", { name: /シャンプー/ })).toBeChecked();
    expect(onToggle).not.toHaveBeenCalled();

    // 対応可能 view に戻すと元のチェック状態に復元され、依然として書き込みは発生しない
    await user.click(screen.getByRole("radio", { name: "対応可能" }));
    expect(screen.getByRole("checkbox", { name: /一般診療/ })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: /シャンプー/ })).not.toBeChecked();
    expect(onToggle).not.toHaveBeenCalled();
  });

  it("対応不可 view でのトグルは capableIds へ反転して書き戻される", async () => {
    const { onToggle } = renderSection();
    const user = userEvent.setup({ delay: null });

    await user.click(screen.getByRole("radio", { name: "対応不可" }));

    // 対応不可 view でチェックを付ける = 対応可能から外す
    await user.click(screen.getByRole("checkbox", { name: /一般診療/ }));
    expect(onToggle).toHaveBeenCalledWith("1", false);

    // 対応不可 view でチェックを外す = 対応可能へ戻す
    await user.click(screen.getByRole("checkbox", { name: /シャンプー/ }));
    expect(onToggle).toHaveBeenCalledWith("2", true);
  });

  it("新規スタッフは「スタッフ登録後に設定できます」を表示し、表示切替やチェックを出さない", () => {
    renderSection({ isNew: true });

    expect(screen.getByText("スタッフ登録後に設定できます")).toBeInTheDocument();
    expect(screen.queryByRole("radio", { name: "対応不可" })).not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
  });

  it("予約区分一覧が未取得/0件のときは空メッセージのみで、反転表示由来の編集経路を出さない", () => {
    renderSection({ types: [] });

    expect(screen.getByText("予約区分が登録されていません")).toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("radio", { name: "対応不可" })).not.toBeInTheDocument();
  });
});
