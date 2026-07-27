import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { PrimaryButton } from "./PrimaryButton";

describe("PrimaryButton", () => {
  it("既定（colorVariant 未指定）は brand と同じ primary teal + pill を使う", () => {
    render(<PrimaryButton>新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgActionPrimary);
    expect(button.className).toContain(C.textOnActionPrimary);
    expect(button.className).toContain(C.hoverTextOnActionPrimary);
    expect(button.className).toContain(C.activeTextOnActionPrimary);
    expect(button.className).toContain("rounded-full");
  });

  it('colorVariant="brand" は brand teal を明示的に使う', () => {
    render(<PrimaryButton colorVariant="brand">新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgBrandIdentity);
    expect(button.className).toContain("text-white");
    expect(button.className).not.toContain("text-black");
    expect(button.className).toContain("text-xl");
    expect(button.className).toContain("font-bold");
    expect(button.className).not.toContain("font-semibold");
    expect(button.className).toContain("rounded-full");
  });

  it('colorVariant="default" は semantic primary の互換 alias', () => {
    render(<PrimaryButton colorVariant="default">新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgActionPrimary);
  });
});
