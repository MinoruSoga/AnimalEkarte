import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { UnifiedTabs, UnifiedTabsRoot } from "./UnifiedTabs";

const ITEMS = [
  { value: "first", label: "最初" },
  { value: "second", label: "次" },
] as const;

describe("UnifiedTabs interaction surface", () => {
  it("全tab triggerを44px以上のnative buttonとして描画する", () => {
    render(
      <UnifiedTabs items={ITEMS} value="first" onValueChange={() => undefined} />,
    );

    for (const tab of screen.getAllByRole("tab")) {
      expect(tab.tagName).toBe("BUTTON");
      expect(tab).toHaveClass("min-h-11");
      expect(tab).toHaveClass("shrink-0");
    }
  });

  it("tab list は overflow-x-auto で全項目到達可能にする (BUG-458)", () => {
    const { container } = render(
      <UnifiedTabs items={ITEMS} value="first" onValueChange={() => undefined} />,
    );
    const list = container.querySelector('[role="tablist"]');
    expect(list).toHaveClass("overflow-x-auto", "max-w-full");
  });

  it("transition中のheadless rootをaria-busyで通知する", () => {
    const { container } = render(
      <UnifiedTabsRoot
        value="first"
        onValueChange={() => undefined}
        ariaBusy
      >
        <div>content</div>
      </UnifiedTabsRoot>,
    );

    expect(container.firstElementChild).toHaveAttribute("aria-busy", "true");
  });
});
