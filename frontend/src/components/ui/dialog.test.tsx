import * as React from "react";
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { Dialog, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from "./dialog";

/**
 * BUG-DIALOG-FOCUS-RESTORE (EMR-64): a controlled Dialog rendered without a
 * DialogTrigger (e.g. the treatment-plan search dialog) has no Radix restore
 * target, so closing it dropped DOM focus to document.body and produced an
 * aria-focus warning. The shared wrapper must record the element focused at
 * open and restore it on close for every close path.
 */
describe("Dialog focus restore", () => {
  function ControlledDialog({ onCloseAutoFocus }: { onCloseAutoFocus?: (event: Event) => void }) {
    const [open, setOpen] = React.useState(false);
    return (
      <>
        <button type="button" onClick={() => setOpen(true)}>
          検索を開く
        </button>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogContent onCloseAutoFocus={onCloseAutoFocus}>
            <DialogTitle>治療検索</DialogTitle>
            <DialogDescription>内容</DialogDescription>
          </DialogContent>
        </Dialog>
      </>
    );
  }

  it("returns focus to the controlled opener after Escape close", async () => {
    const user = userEvent.setup();
    render(<ControlledDialog />);

    const opener = screen.getByRole("button", { name: "検索を開く" });
    await user.click(opener);

    const content = await screen.findByRole("dialog");
    expect(content).toContain(document.activeElement);

    await user.keyboard("{Escape}");
    await waitFor(() => expect(opener).toHaveFocus());
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("returns focus to the controlled opener after the built-in close button", async () => {
    const user = userEvent.setup();
    render(<ControlledDialog />);

    const opener = screen.getByRole("button", { name: "検索を開く" });
    await user.click(opener);
    await screen.findByRole("dialog");

    await user.click(screen.getByRole("button", { name: "Close" }));
    await waitFor(() => expect(opener).toHaveFocus());
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("returns focus to the controlled opener on a programmatic open->false transition", async () => {
    const user = userEvent.setup();
    function Harness({ open }: { open: boolean }) {
      return (
        <>
          <button type="button">検索を開く</button>
          <Dialog open={open} onOpenChange={() => undefined}>
            <DialogContent>
              <DialogTitle>治療検索</DialogTitle>
              <DialogDescription>内容</DialogDescription>
            </DialogContent>
          </Dialog>
        </>
      );
    }
    const { rerender } = render(<Harness open={false} />);

    const opener = screen.getByRole("button", { name: "検索を開く" });
    // Focus the invoking control, then drive the controlled open prop directly:
    // this is the open->false transition a consumer performs without Radix Close.
    await user.click(opener);
    rerender(<Harness open={true} />);
    await screen.findByRole("dialog");

    rerender(<Harness open={false} />);
    await waitFor(() => expect(opener).toHaveFocus());
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("keeps Radix DialogTrigger restore working (trigger path)", async () => {
    const user = userEvent.setup();
    render(
      <Dialog>
        <DialogTrigger asChild>
          <button type="button">トリガーで開く</button>
        </DialogTrigger>
        <DialogContent>
          <DialogTitle>治療検索</DialogTitle>
          <DialogDescription>内容</DialogDescription>
        </DialogContent>
      </Dialog>,
    );

    const trigger = screen.getByRole("button", { name: "トリガーで開く" });
    await user.click(trigger);
    await screen.findByRole("dialog");

    await user.keyboard("{Escape}");
    await waitFor(() => expect(trigger).toHaveFocus());
  });

  it("lets a consumer onCloseAutoFocus with preventDefault win over the shared restore", async () => {
    const user = userEvent.setup();
    render(
      <>
        <button type="button" data-testid="sentinel">
          sentinel
        </button>
        <ControlledDialog
          onCloseAutoFocus={(event) => {
            event.preventDefault();
            screen.getByTestId("sentinel").focus();
          }}
        />
      </>,
    );

    const opener = screen.getByRole("button", { name: "検索を開く" });
    await user.click(opener);
    await screen.findByRole("dialog");

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.getByTestId("sentinel")).toHaveFocus());
    expect(opener).not.toHaveFocus();
  });

  it("defers to the Radix default when the recorded opener is unmounted before close", async () => {
    const user = userEvent.setup();
    function UnmountingOpener() {
      const [open, setOpen] = React.useState(false);
      return (
        <>
          {!open ? (
            <button type="button" onClick={() => setOpen(true)}>
              検索を開く
            </button>
          ) : null}
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent>
              <DialogTitle>治療検索</DialogTitle>
              <DialogDescription>内容</DialogDescription>
            </DialogContent>
          </Dialog>
        </>
      );
    }
    render(<UnmountingOpener />);

    const opener = screen.getByRole("button", { name: "検索を開く" });
    await user.click(opener);
    await screen.findByRole("dialog");
    expect(screen.queryByRole("button", { name: "検索を開く" })).not.toBeInTheDocument();

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    // The stale element must not be focused; the Radix default applies.
    expect(document.activeElement).not.toBe(opener);
    expect(document.activeElement).toBe(document.body);
  });
});
