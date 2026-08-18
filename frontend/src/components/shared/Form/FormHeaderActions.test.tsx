import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { FormHeaderActions } from "./FormHeaderActions";

describe("FormHeaderActions", () => {
  it("キャンセルと確定を右並びに出し、キャンセルで onCancel する", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();

    render(
      <FormHeaderActions
        onCancel={onCancel}
        submitLabel="保存"
      />,
    );

    expect(screen.getByRole("button", { name: "キャンセル" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
