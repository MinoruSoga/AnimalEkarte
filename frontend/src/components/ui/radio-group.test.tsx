import { useActionState, useState } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { RadioGroup, RadioGroupItem } from "./radio-group";

// EMR-252: React 19 fires form.reset() after a <form action> completes. Radix
// restores a mount-time baseline (initialValueRef). The wrapper remounts the
// Radix root keyed on the controlled prop so the baseline is always the latest
// committed value. The plain uncontrolled probe input proves the reset
// actually ran in these tests.

function ControlledRadioForm({ onValueChange }: { onValueChange: (next: string) => void }) {
  const [value, setValue] = useState("a");
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <RadioGroup
        aria-label="区分"
        name="kind"
        value={value}
        onValueChange={(next) => {
          onValueChange(next);
          setValue(next);
        }}
      >
        <RadioGroupItem value="a" aria-label="診察" />
        <RadioGroupItem value="b" aria-label="予防" />
      </RadioGroup>
      <button type="submit">保存</button>
    </form>
  );
}

function UncontrolledRadioForm({ onValueChange }: { onValueChange: (next: string) => void }) {
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <RadioGroup aria-label="区分" name="kind" defaultValue="a" onValueChange={onValueChange}>
        <RadioGroupItem value="a" aria-label="診察" />
        <RadioGroupItem value="b" aria-label="予防" />
      </RadioGroup>
      <button type="submit">保存</button>
    </form>
  );
}

describe("RadioGroup + form action reset (EMR-252)", () => {
  it("controlled: post-action form.reset() keeps the latest value and does not refire onValueChange with the mount value", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(<ControlledRadioForm onValueChange={onValueChange} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("radio", { name: "予防" }));
    expect(screen.getByRole("radio", { name: "予防" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "診察" })).not.toBeChecked();
    expect(onValueChange).toHaveBeenCalledWith("b");

    await user.click(screen.getByRole("button", { name: "保存" }));

    // The uncontrolled probe reverting proves form.reset() ran.
    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await act(async () => {});
    expect(screen.getByRole("radio", { name: "予防" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "診察" })).not.toBeChecked();
    expect(onValueChange).toHaveBeenCalledTimes(1);
  });

  it("uncontrolled: form.reset() restores defaultValue (native semantics)", async () => {
    const user = userEvent.setup();
    render(<UncontrolledRadioForm onValueChange={vi.fn()} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("radio", { name: "予防" }));
    expect(screen.getByRole("radio", { name: "予防" })).toBeChecked();

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await waitFor(() => {
      expect(screen.getByRole("radio", { name: "診察" })).toBeChecked();
      expect(screen.getByRole("radio", { name: "予防" })).not.toBeChecked();
    });
  });
});
