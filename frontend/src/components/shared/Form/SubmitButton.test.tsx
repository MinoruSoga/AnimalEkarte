import { describe, it, expect } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { C } from "@/lib/design-tokens";
import { SubmitButton } from "./SubmitButton";

function renderInForm(children: React.ReactNode) {
  return render(<form action={async () => {}}>{children}</form>);
}

describe("SubmitButton", () => {
  it("既定（colorVariant 未指定）は brand と同じ primary teal + pill を使う", () => {
    renderInForm(<SubmitButton>保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.className).toContain(C.bgActionPrimarySolid);
    expect(button.className).toContain(C.textOnActionPrimary);
    expect(button.className).toContain(C.hoverTextOnActionPrimary);
    expect(button.className).toContain(C.activeTextOnActionPrimary);
    expect(button.className).toContain("rounded-full");
  });

  it('colorVariant="brand" は AA 適合の brand Solid 塗りを使う', () => {
    renderInForm(<SubmitButton colorVariant="brand">保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.className).toContain(C.bgBrandIdentitySolid);
    expect(button.className).not.toContain(C.bgBrandIdentity);
    expect(button.className).toContain("text-white");
    expect(button.className).not.toContain("text-black");
    expect(button.className).toContain("text-base");
    expect(button.className).toContain("font-medium");
    expect(button.className).not.toContain("text-xl");
    expect(button.className).not.toContain("font-bold");
    expect(button.className).not.toContain("font-semibold");
    expect(button.className).toContain("rounded-full");
  });

  it('colorVariant="default" は semantic primary の互換 alias', () => {
    renderInForm(<SubmitButton colorVariant="default">保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.className).toContain(C.bgActionPrimarySolid);
  });
});

describe("SubmitButton size variants (EMR-400)", () => {
  it.each([undefined, "default", "primary", "brand", "destructive"] as const)(
    "colorVariant=%s 既定サイズは h-11 クラスを含む",
    (colorVariant) => {
      renderInForm(<SubmitButton colorVariant={colorVariant}>保存</SubmitButton>);
      const button = screen.getByRole("button", { name: "保存" });
      expect(button.classList.contains("h-11")).toBe(true);
    },
  );

  it.each([undefined, "default", "primary", "brand", "destructive"] as const)(
    "colorVariant=%s size=lg は h-12 を使い h-11 を残さない",
    (colorVariant) => {
      renderInForm(
        <SubmitButton colorVariant={colorVariant} size="lg">
          保存
        </SubmitButton>,
      );
      const button = screen.getByRole("button", { name: "保存" });
      expect(button.classList.contains("h-12")).toBe(true);
      expect(button.classList.contains("h-11")).toBe(false);
    },
  );

  it("呼び出し側 className の h-9 オーバーライドを保持する", () => {
    renderInForm(<SubmitButton className="h-9">保存</SubmitButton>);
    const button = screen.getByRole("button", { name: "保存" });
    expect(button.classList.contains("h-9")).toBe(true);
  });

  it("明示 disabled=true/false が反映される", () => {
    const { unmount } = renderInForm(<SubmitButton disabled>保存</SubmitButton>);
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    unmount();
    renderInForm(<SubmitButton disabled={false}>保存</SubmitButton>);
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
  });

  it("親フォームの action pending 中は disabled + loadingText になり、解決後に戻る", async () => {
    const user = userEvent.setup();
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    render(
      <form action={() => gate}>
        <SubmitButton>保存</SubmitButton>
      </form>,
    );

    await user.click(screen.getByRole("button", { name: "保存" }));
    const loadingButton = await screen.findByRole("button", { name: "保存中..." });
    expect(loadingButton).toBeDisabled();

    await act(async () => {
      release();
    });
    await waitFor(() => expect(screen.getByRole("button", { name: "保存" })).toBeEnabled());
  });
});
