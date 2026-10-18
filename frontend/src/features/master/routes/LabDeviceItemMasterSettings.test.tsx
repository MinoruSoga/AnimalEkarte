import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LabDeviceItemMasterSettings } from "./LabDeviceItemMasterSettings";

// EMR-411 C-3/B10: ルートの workspace wrapper と header actions の DOM 契約だけを
// 検証するため、データ取得系は最小 stub に置き換える。
vi.mock("../hooks/use-lab-device-item-master-settings", () => ({
  useLabDeviceItemMasterSettings: () => ({
    canCreate: true,
    canEdit: true,
    ensureMutation: { isPending: false, mutate: vi.fn() },
    handleNew: vi.fn(),
    fromBoard: false,
    sourceFromQuery: null,
    devicesFetched: true,
    rows: [],
    handleEdit: vi.fn(),
    showPanel: false,
    selectedId: null,
    selectedRow: null,
    selectedItems: [],
    examTypes: [],
    unusedSourceTypes: [],
    readOnly: false,
    createMutation: { isPending: false },
    saveConfigurationMutation: { isPending: false },
    handleClose: vi.fn(),
    handleSave: vi.fn(),
    handleDirtyChange: vi.fn(),
    dirty: { discardDialog: null },
  }),
}));

vi.mock("../api/lab-devices", () => ({
  useGetLabDevices: () => ({ isPending: false, isError: false }),
}));
vi.mock("../api/lab-device-item-masters", () => ({
  useGetLabDeviceItemMasters: () => ({ isPending: false, isError: false }),
}));
vi.mock("../api/exam-types-master", () => ({
  useGetAllExaminationTypes: () => ({ isPending: false, isError: false }),
}));
vi.mock("../components/LabDeviceItemMasterSidePanel", () => ({
  LabDeviceItemMasterSidePanel: () => <div data-testid="lab-side-panel" />,
}));
vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: (props: { headerAction?: ReactNode; children?: ReactNode; className?: string }) => (
    <div data-testid="page-layout-mock" className={props.className}>
      {props.headerAction}
      {props.children}
    </div>
  ),
}));

describe("LabDeviceItemMasterSettings workspace layout (EMR-411)", () => {
  it("C-3: ルートは relative flex h-full を保ち <1280 で grid 化してヘッダ行を残す", () => {
    const { container } = render(
      <MemoryRouter>
        <LabDeviceItemMasterSettings />
      </MemoryRouter>,
    );

    const root = container.firstElementChild;
    expect(root).toHaveClass("relative", "flex", "h-full");
    // 1280 未満: 1列×[auto, 1fr] grid — overlay パネルが header 行を覆わない契約
    expect(root).toHaveClass(
      "max-xl:grid",
      "max-xl:grid-cols-1",
      "max-xl:grid-rows-[auto_minmax(0,1fr)]",
    );
    // リスト wrapper / PageLayout が contents 化して header・body を grid 行に出す
    const listWrap = root?.firstElementChild;
    expect(listWrap).toHaveClass("flex-1", "min-w-0", "max-xl:contents");
    expect(listWrap?.querySelector('[data-testid="page-layout-mock"]')).toHaveClass(
      "max-xl:contents",
    );
  });

  it("B10: ヘッダ操作のルートが flex-wrap・justify-end を持つ", () => {
    render(
      <MemoryRouter>
        <LabDeviceItemMasterSettings />
      </MemoryRouter>,
    );

    const ensure = screen.getByRole("button", { name: "既定項目を用意" });
    const actions = ensure.parentElement;
    expect(actions).toHaveClass("flex", "flex-wrap", "items-center", "justify-end", "gap-2");
  });
});
