import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { SidePeekTitleInput } from "./SidePeekTitleInput";

describe("SidePeekTitleInput (EMR-401 C5-4)", () => {
  it("既定 placeholder は 名称を入力 で、マスタ名称必須として aria-required を持つ", () => {
    render(<SidePeekTitleInput id="master-title" value="" onChange={vi.fn()} />);

    const input = screen.getByPlaceholderText("名称を入力");
    expect(input).toHaveAttribute("aria-required", "true");
    expect(screen.queryByPlaceholderText("無題")).not.toBeInTheDocument();
  });

  it("error 時は input の id 由来の describedby でエラー要素と結びつく", () => {
    render(
      <SidePeekTitleInput
        id="master-title"
        value=""
        onChange={vi.fn()}
        error="名称を入力してください"
      />,
    );

    const input = screen.getByPlaceholderText("名称を入力");
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAttribute("aria-describedby", "master-title-error");
    expect(document.getElementById("master-title-error")).toHaveTextContent(
      "名称を入力してください",
    );
  });

  it("error なしでは aria-describedby を付けない", () => {
    render(<SidePeekTitleInput id="master-title" value="x" onChange={vi.fn()} />);

    expect(screen.getByPlaceholderText("名称を入力")).not.toHaveAttribute("aria-describedby");
  });

  it("明示 placeholder（domain 指定）は既定値で上書きしない", () => {
    render(
      <SidePeekTitleInput
        id="master-title"
        value=""
        onChange={vi.fn()}
        placeholder="新規スタッフ登録"
      />,
    );

    expect(screen.getByPlaceholderText("新規スタッフ登録")).toBeInTheDocument();
  });
});
