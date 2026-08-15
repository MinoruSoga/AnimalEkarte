import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ShiftTemplateToolbar } from "./ShiftTemplateSettingsParts";

describe("ShiftTemplateToolbar", () => {
  it("新規登録はlabelとglyphを維持したまま44px以上の操作領域を持つ", async () => {
    const user = userEvent.setup();
    const onCreate = vi.fn();

    render(<ShiftTemplateToolbar count={8} onCreate={onCreate} />);

    const createButton = screen.getByRole("button", { name: "新規登録" });
    expect(createButton).toHaveClass("min-h-11", "min-w-11");
    expect(createButton).toHaveTextContent("新規登録");
    expect(createButton.querySelector("svg")).toBeInTheDocument();

    await user.click(createButton);

    expect(onCreate).toHaveBeenCalledTimes(1);
  });
});
