import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { PropertyToggleButton } from "./PropertyToggleButton";

function ToggleForm({ onSubmit }: { onSubmit: (data: FormData) => void }) {
  const [pressed, setPressed] = useState(false);

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit(new FormData(event.currentTarget));
      }}
    >
      <PropertyToggleButton label="有効" pressed={pressed} onPressedChange={setPressed}>
        {pressed ? "有効" : "無効"}
      </PropertyToggleButton>
      <button type="submit">保存</button>
    </form>
  );
}

describe("PropertyToggleButton", () => {
  it("exposes its pressed state, toggles with Space, and does not submit the parent form", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(data: FormData) => void>();
    render(<ToggleForm onSubmit={onSubmit} />);

    const toggle = screen.getByRole("button", { name: "有効を切り替え" });
    expect(toggle).toHaveAttribute("type", "button");
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    toggle.focus();
    expect(toggle).toHaveFocus();
    await user.keyboard(" ");

    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(toggle).toHaveTextContent("有効");
    expect(onSubmit).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "保存" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit.mock.calls[0]?.[0].get("有効")).toBeNull();
  });
});
