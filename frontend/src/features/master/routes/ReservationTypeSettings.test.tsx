import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { LAYOUT } from "@/lib/design-tokens";
import { ReservationTypeSettings } from "./ReservationTypeSettings";

const pageLayoutPropsSpy = vi.fn();

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: (props: { title: string; maxWidth?: string; children: ReactNode }) => {
    pageLayoutPropsSpy(props);
    return (
      <div data-testid="page-layout-mock" data-max-width={props.maxWidth}>
        <h1>{props.title}</h1>
        {props.children}
      </div>
    );
  },
}));

vi.mock("../hooks/use-reservation-type-settings", () => ({
  useReservationTypeSettings: () => ({
    groupsRaw: [],
    activeGroups: [],
    categoryDefaultGroupId: undefined,
    categoryCrud: {
      filteredItems: [],
      activeFilters: [],
      setActiveFilters: vi.fn(),
      searchTerm: "",
      setSearchTerm: vi.fn(),
      editTarget: null,
      panelItem: null,
      pendingDelete: null,
      handleClose: vi.fn(),
      handleDeleteCancel: vi.fn(),
      handleDeleteConfirm: vi.fn(),
    },
    groupCrud: {
      editTarget: null,
      panelItem: null,
      pendingDelete: null,
      handleClose: vi.fn(),
      handleDeleteCancel: vi.fn(),
      handleDeleteConfirm: vi.fn(),
    },
    groupSave: { handleSave: vi.fn() },
    categorySave: { handleSave: vi.fn() },
    canCreate: true,
    canEdit: true,
    canDelete: true,
    dirty: { discardDialog: null },
    handleCategoryEdit: vi.fn(),
    handleGroupEdit: vi.fn(),
    handleCategoryAddInGroup: vi.fn(),
    handleGroupAdd: vi.fn(),
    handleGroupDeleteRequest: vi.fn(),
    handleCategoryDeleteRequest: vi.fn(),
    handleDirtyChange: vi.fn(),
  }),
}));

vi.mock("../components/ReservationTypeSettingsContent", () => ({
  ReservationTypeSettingsContent: () => <div data-testid="reservation-content-mock" />,
}));

vi.mock("../components/ReservationTypeSettingsSidePanels", () => ({
  ReservationTypeSettingsSidePanels: () => null,
}));

vi.mock("../components/ReservationTypeDeleteDialogs", () => ({
  ReservationTypeDeleteDialogs: () => null,
}));

describe("ReservationTypeSettings 少列マスタ幅制限（EMR-333）", () => {
  it("PageLayout に tableNarrow (max-w-5xl) が渡される", () => {
    render(
      <MemoryRouter>
        <ReservationTypeSettings />
      </MemoryRouter>,
    );

    expect(pageLayoutPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "予約区分",
        maxWidth: LAYOUT.pageContentMaxWidth.tableNarrow,
      }),
    );
    expect(screen.getByTestId("page-layout-mock")).toHaveAttribute(
      "data-max-width",
      LAYOUT.pageContentMaxWidth.tableNarrow,
    );
  });
});
