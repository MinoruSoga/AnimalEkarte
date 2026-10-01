import { useActionState, useState } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Switch } from "./switch";

// EMR-252: React 19 fires form.reset() after a <form action> completes. Radix
// restores a mount-time baseline (initialCheckedStateRef / hidden input
// defaultChecked). The wrapper remounts the Radix root keyed on the controlled
// prop so the baseline is always the latest committed value. The plain
// uncontrolled probe input proves the reset actually ran in these tests.

function ControlledSwitchForm({ onCheckedChange }: { onCheckedChange: (next: boolean) => void }) {
  const [checked, setChecked] = useState(false);
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Switch
        aria-label="通知"
        name="notify"
        checked={checked}
        onCheckedChange={(next) => {
          onCheckedChange(next);
          setChecked(next);
        }}
      />
      <button type="submit">保存</button>
    </form>
  );
}

function UncontrolledSwitchForm({ onCheckedChange }: { onCheckedChange: (next: boolean) => void }) {
  const [, formAction] = useActionState(async (): Promise<{ ok: boolean }> => ({ ok: true }), {
    ok: false,
  });
  return (
    <form action={formAction}>
      <input aria-label="probe" name="probe" defaultValue="pristine" />
      <Switch
        aria-label="通知"
        name="notify"
        defaultChecked={false}
        onCheckedChange={onCheckedChange}
      />
      <button type="submit">保存</button>
    </form>
  );
}

describe("Switch + form action reset (EMR-252)", () => {
  it("controlled: post-action form.reset() keeps the latest value and does not refire onCheckedChange with the mount value", async () => {
    const user = userEvent.setup();
    const onCheckedChange = vi.fn();
    render(<ControlledSwitchForm onCheckedChange={onCheckedChange} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("switch", { name: "通知" }));
    expect(screen.getByRole("switch", { name: "通知" })).toHaveAttribute("aria-checked", "true");
    expect(onCheckedChange).toHaveBeenCalledWith(true);

    await user.click(screen.getByRole("button", { name: "保存" }));

    // The uncontrolled probe reverting proves form.reset() ran.
    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await act(async () => {});
    expect(screen.getByRole("switch", { name: "通知" })).toHaveAttribute("aria-checked", "true");
    expect(onCheckedChange).toHaveBeenCalledTimes(1);
  });

  it("uncontrolled: form.reset() restores defaultChecked (native semantics)", async () => {
    const user = userEvent.setup();
    render(<UncontrolledSwitchForm onCheckedChange={vi.fn()} />);

    const probe = screen.getByLabelText("probe");
    await user.type(probe, "+edited");

    await user.click(screen.getByRole("switch", { name: "通知" }));
    expect(screen.getByRole("switch", { name: "通知" })).toHaveAttribute("aria-checked", "true");

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(probe).toHaveValue("pristine"));
    await waitFor(() =>
      expect(screen.getByRole("switch", { name: "通知" })).toHaveAttribute("aria-checked", "false"),
    );
  });
});
