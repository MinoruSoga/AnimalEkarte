import { useActionState, useState } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Checkbox } from "./checkbox";

// EMR-252: React 19 fires form.reset() after a <form action> completes. Radix
// restores a mount-time baseline (initialCheckedStateRef / hidden input
// defaultChecked). The wrapper remounts the Radix root keyed on the controlled
// prop so the baseline is always the latest committed value. The plain
// uncontrolled probe input proves the reset actually ran in these tests.

function ControlledCheckboxForm({ onCheckedChange }: { onCheckedChange: (next: boolean) => void }) {
  const [checked, setChecked] = useState(false);
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Checkbox
        aria-label="同意"
        name="agree"
        checked={checked}
        onCheckedChange={(next) => {
          if (next === true) {
            onCheckedChange(true);
            setChecked(true);
          } else if (next === false) {
            onCheckedChange(false);
            setChecked(false);
          }
        }}
      />
      <button type="submit">保存</button>
    </form>
  );
}

function UncontrolledCheckboxForm({
  onCheckedChange,
}: {
  onCheckedChange: (next: boolean) => void;
}) {
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Checkbox
        aria-label="同意"
        name="agree"
        defaultChecked={false}
        onCheckedChange={(next) => {
          if (next === true) onCheckedChange(true);
          else if (next === false) onCheckedChange(false);
        }}
      />
      <button type="submit">保存</button>
    </form>
  );
}

describe("Checkbox + form action reset (EMR-252)", () => {
  it("controlled: post-action form.reset() keeps the latest value and does not refire onCheckedChange with the mount value", async () => {
    const user = userEvent.setup();
    const onCheckedChange = vi.fn();
    render(<ControlledCheckboxForm onCheckedChange={onCheckedChange} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("checkbox", { name: "同意" }));
    expect(screen.getByRole("checkbox", { name: "同意" })).toBeChecked();
    expect(onCheckedChange).toHaveBeenCalledWith(true);

    await user.click(screen.getByRole("button", { name: "保存" }));

    // The uncontrolled probe reverting proves form.reset() ran.
    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await act(async () => {});
    expect(screen.getByRole("checkbox", { name: "同意" })).toBeChecked();
    expect(onCheckedChange).toHaveBeenCalledTimes(1);
  });

  it("uncontrolled: form.reset() restores defaultChecked (native semantics)", async () => {
    const user = userEvent.setup();
    render(<UncontrolledCheckboxForm onCheckedChange={vi.fn()} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("checkbox", { name: "同意" }));
    expect(screen.getByRole("checkbox", { name: "同意" })).toBeChecked();

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await waitFor(() => expect(screen.getByRole("checkbox", { name: "同意" })).not.toBeChecked());
  });
});
