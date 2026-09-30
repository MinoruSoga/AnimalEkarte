import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { C } from "@/lib/design-tokens";
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

  it.each(["default", "primary"] as const)(
    "%s variant は brand と同じ primary teal と pressed 色を使う",
    (variant) => {
      render(<Button variant={variant}>保存</Button>);

      const button = screen.getByRole("button", { name: "保存" });
      expect(button).toHaveClass(
        C.bgActionPrimary,
        C.hoverBgActionPrimary,
        C.textOnActionPrimary,
        C.hoverTextOnActionPrimary,
        C.activeTextOnActionPrimary,
      );
    },
  );

  it("link variant は通常文字向けの accessible primary text を使う", () => {
    render(<Button variant="link">詳細</Button>);

    expect(screen.getByRole("button", { name: "詳細" })).toHaveClass(C.textActionPrimary);
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

/**
 * BUG-BUTTON-FOCUS-INVISIBLE (EMR-207): the shared cva base ended in
 * `outline-none` with no `:focus-visible` replacement, so keyboard users saw
 * no focus indicator on any variant (WCAG 2.2 AA 2.4.7 / 2.4.11). The shared
 * base must carry a brand-colored focus-visible ring — including a ring
 * offset so it stays visible on the brand-filled default/primary variants —
 * while pointer-only focus still shows nothing.
 */
describe("Button :focus-visible indicator", () => {
  const FOCUS_VISIBLE_CLASSES = [
    "focus-visible:ring-2",
    C.focusVisibleRingActionPrimary,
    "focus-visible:ring-offset-1",
    "focus-visible:ring-offset-background",
  ] as const;

  it.each(BUTTON_VARIANT_SIZE_CASES)(
    "composes a visible brand focus ring for $variant/$size",
    ({ variant, size }) => {
      render(
        <Button variant={variant} size={size}>
          保存
        </Button>,
      );

      const button = screen.getByRole("button", { name: "保存" });
      for (const className of FOCUS_VISIBLE_CLASSES) {
        expect(button, className).toHaveClass(className);
      }
      expect(button).toHaveClass("outline-none");
      expect(button).not.toHaveAttribute("tabindex");
    },
  );

  it("keeps every ring class keyboard-gated (no plain focus/ring indicator)", () => {
    render(<Button>保存</Button>);

    const button = screen.getByRole("button", { name: "保存" });
    const pointerIndicators = button.className
      .split(/\s+/)
      .filter((token) => /^ring-\d/.test(token) || /^focus:/.test(token));
    expect(pointerIndicators).toEqual([]);
  });

  it("preserves natural tab order into and between buttons", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <Button>最初</Button>
        <Button variant="outline">次</Button>
      </div>,
    );

    const first = screen.getByRole("button", { name: "最初" });
    const next = screen.getByRole("button", { name: "次" });
    await user.tab();
    expect(first).toHaveFocus();
    await user.tab();
    expect(next).toHaveFocus();
  });

  it("keeps the focus-visible ring classes on the asChild element", () => {
    render(
      <Button asChild>
        <a href="/settings">設定</a>
      </Button>,
    );

    const link = screen.getByRole("link", { name: "設定" });
    for (const className of FOCUS_VISIBLE_CLASSES) {
      expect(link).toHaveClass(className);
    }
  });
});

/**
 * design-states.md §3 (FE-DS1): loading は disabled + aria-busy + スピナーで
 * 表現し、テキストを維持する。クリック抑止・スクリーンリーダー通知・
 * 幅確保（ラベル維持）の3点を満たすのが標準仕様。
 */
describe("Button loading state", () => {
  it("disables the button and marks aria-busy while loading", () => {
    render(<Button loading>保存中</Button>);

    const button = screen.getByRole("button", { name: "保存中" });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("aria-busy", "true");
    expect(button).toHaveAttribute("data-loading", "true");
  });

  it("does not fire onClick while loading", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(
      <Button loading onClick={onClick}>
        送信
      </Button>,
    );

    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onClick).not.toHaveBeenCalled();
  });

  it("renders a hidden spinner next to the preserved label", () => {
    render(<Button loading>処理中</Button>);

    const button = screen.getByRole("button", { name: "処理中" });
    const spinner = button.querySelector("svg.animate-spin");
    expect(spinner).not.toBeNull();
    expect(spinner).toHaveAttribute("aria-hidden", "true");
    expect(button).toHaveTextContent("処理中");
  });

  it("asChild はスピナーを描画しないが aria-busy は付く", () => {
    render(
      <Button asChild loading>
        <a href="/save">保存</a>
      </Button>,
    );

    const link = screen.getByRole("link", { name: "保存" });
    expect(link).toHaveAttribute("aria-busy", "true");
    expect(link.querySelector("svg")).toBeNull();
  });
});
