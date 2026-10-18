import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Label } from "./label";

describe("Label", () => {
  it("既定は muted-foreground トークンと text-base / mb-2 を使う", () => {
    render(<Label htmlFor="field-1">名前</Label>);
    const label = screen.getByText("名前");
    expect(label.tagName).toBe("LABEL");
    expect(label.className).toContain("text-muted-foreground");
    expect(label.className).not.toContain("text-foreground/65");
    expect(label.className).toContain("text-base");
    expect(label.className).toContain("mb-2");
  });

  it("htmlFor で対応する入力と関連付く", () => {
    render(
      <>
        <Label htmlFor="field-2">メモ</Label>
        <input id="field-2" />
      </>,
    );
    expect(screen.getByLabelText("メモ")).toBeInTheDocument();
  });

  it("呼び出し側の色 className が既定色に勝つ", () => {
    render(
      <Label htmlFor="field-3" className="text-destructive">
        必須
      </Label>,
    );
    const label = screen.getByText("必須");
    expect(label.className).toContain("text-destructive");
    expect(label.className).not.toContain("text-muted-foreground");
  });
});
