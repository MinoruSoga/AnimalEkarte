import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LAYOUT } from "@/lib/design-tokens";

import { MedicineSidePanel } from "./MedicineSidePanel";

// EMR-411 C-3 の対象は motion wrapper の位置契約のみ。Body 本体は別責務のため stub 化する。
vi.mock("./MedicineSidePanelBody", () => ({
  MedicineSidePanelBody: () => <div data-testid="medicine-side-panel-body" />,
}));

describe("MedicineSidePanel motion wrapper (EMR-411 C-3)", () => {
  const baseProps = {
    selectedMedicine: null,
    isCategory: false,
    categoryMedicines: [],
    panelDuration: 0,
    onCloseEdit: vi.fn(),
    onSave: vi.fn(),
    onDeleteRequest: vi.fn(),
  } as const;

  it("編集中の wrapper が overlay 位置トークンと max-w-full を持つ", () => {
    render(<MedicineSidePanel isEditing {...baseProps} />);

    const wrapper = screen.getByTestId("medicine-side-panel-body").parentElement;
    expect(wrapper).not.toBeNull();

    for (const cls of LAYOUT.sidePeek.position.split(" ")) {
      expect(wrapper).toHaveClass(cls);
    }
    expect(wrapper).toHaveClass("max-w-full");
    expect(wrapper).toHaveClass("shrink-0");
  });

  it("非編集中は何もレンダリングしない", () => {
    render(<MedicineSidePanel isEditing={false} {...baseProps} />);

    expect(screen.queryByTestId("medicine-side-panel-body")).not.toBeInTheDocument();
  });
});
