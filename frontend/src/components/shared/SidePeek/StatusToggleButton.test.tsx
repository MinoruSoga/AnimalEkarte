import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { StatusToggleButton } from "./StatusToggleButton";

/**
 * EMR-411 C-2: ステータス切替ボタンは toggle ボタンとして現在の押下状態を
 * aria-pressed で公開する（スクリーンリーダーが「今どちらか」を読み上げられるように）。
 */
describe("StatusToggleButton aria-pressed (EMR-411 C-2)", () => {
  function StatefulToggle({ initial }: { initial: boolean }) {
    const [isActive, setIsActive] = useState(initial);
    return <StatusToggleButton isActive={isActive} onToggle={() => setIsActive((v) => !v)} />;
  }

  it("状態に追従して aria-pressed が false→true→false と遷移する", async () => {
    const user = userEvent.setup();
    render(<StatefulToggle initial={false} />);

    const toggle = screen.getByRole("button", { name: "ステータスを切り替え" });
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByText("無効")).toBeInTheDocument();

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText("有効")).toBeInTheDocument();

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByText("無効")).toBeInTheDocument();
  });
});
