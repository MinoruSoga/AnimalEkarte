import { fireEvent, render, screen } from "@testing-library/react";
import { MessageSquare } from "lucide-react";
import { describe, expect, it, vi } from "vitest";

import { HospitalizationNoteCard } from "./HospitalizationNoteCard";

describe("HospitalizationNoteCard", () => {
  it("カードは内容高さに追随し DESIGN.md の24px余白・heading-3見出しに従う", () => {
    render(
      <HospitalizationNoteCard
        id="owner_request"
        title="主訴"
        icon={MessageSquare}
        value=""
        onChange={vi.fn()}
      />,
    );

    const heading = screen.getByRole("heading", { name: "主訴" });
    expect(heading).toHaveClass("text-heading-3", "font-bold");

    const card = heading.closest("div.rounded-lg");
    expect(card).toHaveClass("p-6");
    expect(card).not.toHaveClass("h-full");
    expect(card).not.toHaveClass("p-2");
  });

  it("textarea は共有 Textarea 既定の text-sm(15px) に任せ value/onChange/aria を維持する", () => {
    const onChange = vi.fn();
    render(
      <HospitalizationNoteCard
        id="owner_request"
        title="主訴"
        icon={MessageSquare}
        value="元気がない"
        onChange={onChange}
        placeholder="主訴を入力..."
      />,
    );

    const textarea = screen.getByRole("textbox", { name: "主訴" });
    expect(textarea).toHaveValue("元気がない");
    expect(textarea).toHaveAttribute("aria-labelledby", "owner_request-label");
    expect(textarea).toHaveClass("min-h-[200px]");
    expect(textarea).not.toHaveClass("text-base");

    fireEvent.change(textarea, { target: { value: "嘔吐" } });
    expect(onChange).toHaveBeenCalledWith("嘔吐");
  });
});
