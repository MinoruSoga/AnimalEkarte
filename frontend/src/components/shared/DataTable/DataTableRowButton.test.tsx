import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DataTableRowButton } from "./DataTableRowButton";

describe("DataTableRowButton", () => {
  it("固有accessible nameを持つ44px以上のnative buttonを描画する", () => {
    render(
      <DataTableRowButton aria-label="ケージ操作: 1番ケージ ID 1">1番ケージ</DataTableRowButton>,
    );

    const button = screen.getByRole("button", { name: "ケージ操作: 1番ケージ ID 1" });
    expect(button.tagName).toBe("BUTTON");
    expect(button).toHaveAttribute("type", "button");
    expect(button).toHaveClass("min-h-11", "min-w-11");
  });

  it("button props と追加classNameをnative buttonへ渡す", async () => {
    const onClick = vi.fn();
    const user = userEvent.setup();
    render(
      <DataTableRowButton
        aria-label="詳細を開く: 見積 ID 42"
        className="justify-end"
        onClick={onClick}
      >
        詳細
      </DataTableRowButton>,
    );

    const button = screen.getByRole("button", { name: "詳細を開く: 見積 ID 42" });
    expect(button).toHaveClass("justify-end");

    await user.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });
});
