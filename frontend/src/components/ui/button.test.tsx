import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Button } from "./button";

describe("Button", () => {
  it("defaults to type button", () => {
    render(<Button>開く</Button>);

    expect(screen.getByRole("button", { name: "開く" })).toHaveAttribute("type", "button");
  });

  it("keeps explicit submit type", () => {
    render(<Button type="submit">保存</Button>);

    expect(screen.getByRole("button", { name: "保存" })).toHaveAttribute("type", "submit");
  });

  it("does not inject type when rendering as child", () => {
    render(
      <Button asChild>
        <a href="/settings">設定</a>
      </Button>,
    );

    expect(screen.getByRole("link", { name: "設定" })).not.toHaveAttribute("type");
  });
});
