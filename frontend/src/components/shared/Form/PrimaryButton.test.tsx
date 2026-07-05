import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { PrimaryButton } from "./PrimaryButton";

describe("PrimaryButton", () => {
  it("既定（colorVariant 未指定）は旧 accent ブルーを使う（既存画面の見た目を変えない）", () => {
    render(<PrimaryButton>新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgAccent);
  });

  it('colorVariant="brand" は DESIGN.md button-primary（brand blue + pill）を使う', () => {
    render(<PrimaryButton colorVariant="brand">新規登録</PrimaryButton>);
    const button = screen.getByRole("button", { name: "新規登録" });
    expect(button.className).toContain(C.bgBrand);
    expect(button.className).toContain("rounded-full");
    expect(button.className).not.toContain(C.bgAccent);
  });
});
