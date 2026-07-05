import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { SubmitButton } from "./SubmitButton";

function renderInForm(children: React.ReactNode) {
  return render(<form action={async () => {}}>{children}</form>);
}

describe("SubmitButton", () => {
  it("既定（colorVariant 未指定）は STYLE.confirmPrimary を使う", () => {
    renderInForm(<SubmitButton>保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.className).toContain(C.bgAccent);
  });

  it('colorVariant="brand" は DESIGN.md button-primary（brand blue + pill）を使う', () => {
    renderInForm(<SubmitButton colorVariant="brand">保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.className).toContain(C.bgBrand);
    expect(button.className).toContain("rounded-full");
    expect(button.className).not.toContain(C.bgAccent);
  });
});
