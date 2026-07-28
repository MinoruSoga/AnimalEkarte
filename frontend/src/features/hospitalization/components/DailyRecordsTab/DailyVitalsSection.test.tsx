import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DailyVitalsSection } from "./DailyVitalsSection";

describe("DailyVitalsSection responsive layout", () => {
  it("バイタル入力はmobileで全幅1列、sm以上で既存の2列に戻る", async () => {
    const user = userEvent.setup();
    render(
      <DailyVitalsSection
        vitals={[]}
        onAddVital={vi.fn()}
        isPending={false}
        canCreate
      />,
    );

    await user.click(screen.getByRole("button", { name: "追加" }));

    const vitalsGrid = screen.getByLabelText("体温 (℃)").closest('[class*="grid-cols"]');
    expect(vitalsGrid).toHaveClass("w-full", "grid-cols-1", "sm:grid-cols-2");
    expect(vitalsGrid).not.toHaveClass("grid-cols-2");
  });
});
