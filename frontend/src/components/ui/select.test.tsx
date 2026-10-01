import { useActionState, useState } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Dialog, DialogContent, DialogDescription, DialogTitle } from "./dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";

describe("Select inside Dialog", () => {
  it("opens and selects an option without a focus loop", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>予約区分</DialogTitle>
          <DialogDescription>予約区分を選択します。</DialogDescription>
          <Select onValueChange={onValueChange}>
            <SelectTrigger aria-label="予約区分">
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="consultation">診察</SelectItem>
              <SelectItem value="vaccination">予防接種</SelectItem>
            </SelectContent>
          </Select>
        </DialogContent>
      </Dialog>,
    );

    await user.click(screen.getByRole("combobox", { name: "予約区分" }));
    await user.click(screen.getByRole("option", { name: "診察" }));

    expect(onValueChange).toHaveBeenCalledWith("consultation");
    expect(screen.getByRole("combobox", { name: "予約区分" })).toHaveTextContent("診察");
  });
});

// EMR-252: React 19 fires form.reset() after a <form action> completes. Radix
// restores a mount-time baseline (initialValueRef + hidden select defaultValue).
// The wrapper remounts the Radix root keyed on the controlled prop so the
// baseline is always the latest committed value. The plain uncontrolled probe
// input proves the reset actually ran in these tests.

function ControlledSelectForm({ onValueChange }: { onValueChange: (next: string) => void }) {
  const [value, setValue] = useState("a");
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Select
        name="choice"
        value={value}
        onValueChange={(next) => {
          onValueChange(next);
          setValue(next);
        }}
      >
        <SelectTrigger aria-label="区分">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">診察</SelectItem>
          <SelectItem value="b">予防</SelectItem>
        </SelectContent>
      </Select>
      <button type="submit">保存</button>
    </form>
  );
}

function UncontrolledSelectForm({ onValueChange }: { onValueChange: (next: string) => void }) {
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Select name="choice" defaultValue="a" onValueChange={onValueChange}>
        <SelectTrigger aria-label="区分">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">診察</SelectItem>
          <SelectItem value="b">予防</SelectItem>
        </SelectContent>
      </Select>
      <button type="submit">保存</button>
    </form>
  );
}

async function selectOption(user: ReturnType<typeof userEvent.setup>, optionName: string) {
  await user.click(screen.getByRole("combobox", { name: "区分" }));
  await user.click(screen.getByRole("option", { name: optionName }));
}

function formOf(element: HTMLElement): HTMLFormElement {
  const form = element.closest("form");
  if (!form) throw new Error("probe is not inside a form");
  return form;
}

describe("Select + form action reset (EMR-252)", () => {
  it("controlled: post-action form.reset() keeps the latest value and does not refire onValueChange with the mount value", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(<ControlledSelectForm onValueChange={onValueChange} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await selectOption(user, "予防");
    expect(screen.getByRole("combobox", { name: "区分" })).toHaveTextContent("予防");
    expect(onValueChange).toHaveBeenCalledWith("b");

    await user.click(screen.getByRole("button", { name: "保存" }));

    // The uncontrolled probe reverting proves form.reset() ran.
    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await act(async () => {});
    expect(screen.getByRole("combobox", { name: "区分" })).toHaveTextContent("予防");
    expect(onValueChange).toHaveBeenCalledTimes(1);
  });

  it("controlled: hidden native select / FormData returns the latest value after reset", async () => {
    const user = userEvent.setup();
    render(<ControlledSelectForm onValueChange={vi.fn()} />);

    const probe = screen.getByLabelText("probe");
    const form = formOf(probe);
    expect(new FormData(form).get("choice")).toBe("a");

    await selectOption(user, "予防");
    expect(new FormData(form).get("choice")).toBe("b");

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await act(async () => {});
    expect(new FormData(form).get("choice")).toBe("b");
  });

  it("uncontrolled: form.reset() restores defaultValue (native semantics)", async () => {
    const user = userEvent.setup();
    render(<UncontrolledSelectForm onValueChange={vi.fn()} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await selectOption(user, "予防");
    expect(screen.getByRole("combobox", { name: "区分" })).toHaveTextContent("予防");

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await waitFor(() =>
      expect(screen.getByRole("combobox", { name: "区分" })).toHaveTextContent("診察"),
    );
  });
});
