import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Button } from "./button";

const BUTTON_VARIANTS = [
  "default",
  "destructive",
  "outline",
  "secondary",
  "ghost",
  "link",
  "primary",
  "ghost-danger",
] as const;

const BUTTON_SIZES = ["default", "sm", "lg", "icon"] as const;

const BUTTON_VARIANT_SIZE_CASES = BUTTON_VARIANTS.flatMap((variant) =>
  BUTTON_SIZES.map((size) => ({ variant, size })),
);

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

  it.each(BUTTON_VARIANT_SIZE_CASES)(
    "keeps a 44px minimum target for $variant/$size when height and width are overridden",
    ({ variant, size }) => {
      render(
        <Button variant={variant} size={size} className="h-9 w-40">
          検索
        </Button>,
      );

      expect(screen.getByRole("button", { name: "検索" })).toHaveClass(
        "min-h-11",
        "min-w-11",
        "h-9",
        "w-40",
      );
    },
  );

  it.each(BUTTON_VARIANT_SIZE_CASES)(
    "keeps a 44px minimum target for $variant/$size when size is overridden",
    ({ variant, size }) => {
      render(
        <Button variant={variant} size={size} className="size-8">
          操作
        </Button>,
      );

      expect(screen.getByRole("button", { name: "操作" })).toHaveClass(
        "min-h-11",
        "min-w-11",
        "size-8",
      );
    },
  );
});
