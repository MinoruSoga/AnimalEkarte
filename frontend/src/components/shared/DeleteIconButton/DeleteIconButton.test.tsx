import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DeleteIconButton } from "./DeleteIconButton";

describe("DeleteIconButton", () => {
  it("44x44px のタッチターゲットを提供する", () => {
    render(<DeleteIconButton className="size-7" onClick={() => {}} />);

    expect(screen.getByRole("button", { name: "削除" })).toHaveClass("min-h-11", "min-w-11");
  });

  it("タッチターゲットを広げてもアイコンの視覚サイズを維持する", () => {
    render(<DeleteIconButton onClick={() => {}} />);

    const icon = screen.getByRole("button", { name: "削除" }).querySelector("svg");
    expect(icon).toHaveClass("size-5");
  });
});
