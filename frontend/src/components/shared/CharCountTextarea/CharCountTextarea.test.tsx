import { useState } from "react";
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { C } from "@/lib/design-tokens";
import { CharCountTextarea } from "./CharCountTextarea";

/**
 * EMR-372 (IU-004): CharCountTextarea composes the shared Textarea and
 * STYLE.textarea, so it must surface the same design-states.md §1 canonical
 * keyboard focus contract — `focus-visible:ring-2` +
 * `C.focusVisibleRingActionPrimary` + `ring-offset-1` — without pointer-focus
 * indicators, even when callers add layout classes via textareaClassName.
 */
const CANONICAL_RING_CLASSES = ["focus-visible:ring-2", C.focusVisibleRingActionPrimary] as const;

const CANONICAL_OFFSET_CLASSES = ["ring-offset-1", "focus-visible:ring-offset-1"] as const;

function expectCanonicalFocusVisible(className: string) {
  const tokens = className.split(/\s+/);
  for (const required of CANONICAL_RING_CLASSES) {
    expect(tokens, required).toContain(required);
  }
  const hasRingOffset = tokens.some((token) =>
    (CANONICAL_OFFSET_CLASSES as readonly string[]).includes(token),
  );
  expect(hasRingOffset, "ring-offset-1").toBe(true);
}

function expectNoPointerFocusIndicators(className: string) {
  const pointerIndicators = className
    .split(/\s+/)
    .filter((token) => /^focus:/.test(token) || /^ring-\d/.test(token));
  expect(pointerIndicators).toEqual([]);
}

function ControlledCharCountTextarea({ maxLength = 10 }: { maxLength?: number }) {
  const [value, setValue] = useState("");
  return <CharCountTextarea value={value} onChange={setValue} maxLength={maxLength} />;
}

describe("CharCountTextarea", () => {
  it("renders the controlled value and the live character count", () => {
    render(<CharCountTextarea value="あいう" onChange={() => {}} maxLength={10} />);

    expect(screen.getByRole("textbox")).toHaveValue("あいう");
    expect(screen.getByText("3 / 10 文字")).toBeInTheDocument();
  });

  it("forwards typed input through onChange and updates the count", async () => {
    const user = userEvent.setup();
    render(<ControlledCharCountTextarea />);

    const textarea = screen.getByRole("textbox");
    await user.type(textarea, "abc");
    expect(textarea).toHaveValue("abc");
    expect(screen.getByText("3 / 10 文字")).toBeInTheDocument();
  });

  it("marks the counter as danger at the maxLength boundary", () => {
    render(<CharCountTextarea value="12345" onChange={() => {}} maxLength={5} />);

    expect(screen.getByText("5 / 5 文字")).toHaveClass(C.danger);
  });

  it("keeps the counter muted below maxLength", () => {
    render(<CharCountTextarea value="1234" onChange={() => {}} maxLength={5} />);

    expect(screen.getByText("4 / 5 文字")).toHaveClass(C.text40);
  });

  it("passes maxLength and disabled through to the textarea", () => {
    render(<CharCountTextarea value="" onChange={() => {}} maxLength={7} disabled />);

    const textarea = screen.getByRole("textbox");
    expect(textarea).toHaveAttribute("maxlength", "7");
    expect(textarea).toBeDisabled();
  });

  it("keeps the canonical focus-visible ring alongside caller classes", () => {
    render(<CharCountTextarea value="" onChange={() => {}} textareaClassName="min-h-0" />);

    const textarea = screen.getByRole("textbox");
    expect(textarea).toHaveClass("min-h-0");
    expectCanonicalFocusVisible(textarea.className);
    expectNoPointerFocusIndicators(textarea.className);
  });
});
