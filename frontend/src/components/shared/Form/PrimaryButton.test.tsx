import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { PrimaryButton } from "./PrimaryButton";

describe("PrimaryButton", () => {
  it("既定（colorVariant 未指定）は DESIGN.md button-primary（brand blue + pill）を使う", () => {
    render(<PrimaryButton>新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgBrand);
    expect(button.className).toContain("rounded-full");
    expect(button.className).not.toContain(C.bgAccent);
  });

  it('colorVariant="brand" を明示指定しても同じ結果になる（後方互換）', () => {
    render(<PrimaryButton colorVariant="brand">新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgBrand);
    expect(button.className).toContain("rounded-full");
    expect(button.className).not.toContain(C.bgAccent);
  });

  it('colorVariant="default" は旧 accent ブルーを使う opt-out', () => {
    render(<PrimaryButton colorVariant="default">新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgAccent);
  });
});
