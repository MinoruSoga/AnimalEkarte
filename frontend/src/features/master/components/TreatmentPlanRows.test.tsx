import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { TreatmentPlanRow } from "./TreatmentPlanRows";
import type { TreatmentVirtualRow } from "../lib/treatment-plan-tab-content-model";

const ROOT_ROW = {
  type: "root",
  isExpanded: false,
  item: {
    id: "root-1",
    name: "診察",
    price: 0,
    isActive: true,
    description: "",
    sortOrder: 1,
    children: [
      {
        id: "child-1",
        name: "一般診察",
        parentId: "root-1",
        price: 1000,
        isActive: true,
        description: "",
        sortOrder: 1,
      },
    ],
  },
} satisfies TreatmentVirtualRow;

describe("TreatmentPlanRow", () => {
  it("展開ボタンはglyphを維持したまま44px以上の操作領域を持つ", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onToggleExpanded = vi.fn();

    render(
      <table>
        <tbody>
          <TreatmentPlanRow
            row={ROOT_ROW}
            canEdit={false}
            onEdit={onEdit}
            onToggleExpanded={onToggleExpanded}
          />
        </tbody>
      </table>,
    );

    const toggle = screen.getByRole("button", {
      name: "治療プラン 診察 (ID root-1) の子項目を展開",
    });
    expect(toggle).toHaveClass("size-[22px]", "min-h-11", "min-w-11");
    expect(toggle.querySelector("svg")).toBeInTheDocument();

    await user.click(toggle);

    expect(onToggleExpanded).toHaveBeenCalledWith("root-1");
    expect(onEdit).not.toHaveBeenCalled();
  });
});
