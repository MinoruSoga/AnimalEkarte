import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { C, STYLE } from "@/lib/design-tokens";
import { Textarea } from "./textarea";

/**
 * EMR-372 (IU-002): design-states.md §1 requires keyboard focus to use the
 * canonical focus-visible indicator — `focus-visible:ring-2` +
 * `C.focusVisibleRingActionPrimary` + `ring-offset-1` — on every interactive
 * element, while pointer-driven focus shows nothing (`focus:` border/shadow/bg
 * classes are forbidden indicators). The contract must hold for the shared
 * Textarea standalone, when composed with `STYLE.textarea`, and on the
 * `STYLE.textarea` token itself (raw `<textarea className={STYLE.textarea}>`
 * consumers such as LineSendPanel bypass the component).
 *
 * design-states.md names `ring-offset-1`; the established codebase precedent
 * (Button, select triggers) gates it as `focus-visible:ring-offset-1`.
 * Both satisfy the spec — only the offset width matters when a ring renders.
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

describe("Textarea :focus-visible indicator", () => {
  it("composes the canonical keyboard focus ring with offset, pointer-silent", () => {
    render(<Textarea aria-label="メモ" />);

    const textarea = screen.getByRole("textbox", { name: "メモ" });
    expectCanonicalFocusVisible(textarea.className);
    expectNoPointerFocusIndicators(textarea.className);
  });

  it("keeps the canonical contract when combined with STYLE.textarea", () => {
    render(<Textarea aria-label="メモ" className={STYLE.textarea} />);

    const textarea = screen.getByRole("textbox", { name: "メモ" });
    expectCanonicalFocusVisible(textarea.className);
    expectNoPointerFocusIndicators(textarea.className);
  });

  it("preserves natural tab order into and between textareas", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <Textarea aria-label="最初" />
        <Textarea aria-label="次" />
      </div>,
    );

    const first = screen.getByRole("textbox", { name: "最初" });
    const next = screen.getByRole("textbox", { name: "次" });
    await user.tab();
    expect(first).toHaveFocus();
    await user.tab();
    expect(next).toHaveFocus();
  });
});

describe("STYLE.textarea token", () => {
  it("carries the canonical focus-visible ring with offset for raw textarea consumers", () => {
    expectCanonicalFocusVisible(STYLE.textarea);
  });

  it("has no pointer-focus indicator classes", () => {
    expectNoPointerFocusIndicators(STYLE.textarea);
  });
});

describe("Textarea invalid / native prop regression", () => {
  it("keeps aria-invalid error styling and passes the attribute through", () => {
    render(<Textarea aria-label="メモ" aria-invalid="true" />);

    const textarea = screen.getByRole("textbox", { name: "メモ" });
    expect(textarea).toHaveAttribute("aria-invalid", "true");
    expect(textarea).toHaveClass(
      "aria-invalid:border-destructive",
      "aria-invalid:bg-destructive/5",
    );
  });

  it("keeps native disabled behaviour and styling hooks", () => {
    render(<Textarea aria-label="メモ" disabled />);

    const textarea = screen.getByRole("textbox", { name: "メモ" });
    expect(textarea).toBeDisabled();
    expect(textarea).toHaveClass("disabled:cursor-not-allowed", "disabled:opacity-50");
  });

  it("stays editable: accepts typed input and forwards onChange", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Textarea aria-label="メモ" defaultValue="" onChange={onChange} />);

    const textarea = screen.getByRole("textbox", { name: "メモ" });
    await user.type(textarea, "あ");
    expect(onChange).toHaveBeenCalled();
    expect(textarea).toHaveValue("あ");
  });
});
