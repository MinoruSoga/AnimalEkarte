import { act } from "react";
import { MemoryRouter } from "react-router";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SupportWidget } from "./SupportWidget";

const { clearHistory } = vi.hoisted(() => ({ clearHistory: vi.fn() }));

// Keep the complete support/ConfirmDialog/Radix portal chain real. Only its
// network/auth dependencies and the unrelated bug-report tab are replaced.
vi.mock("./BugReportTab", () => ({ BugReportTab: () => null }));
vi.mock("../api/get-support-chat-status", () => ({
  useGetSupportChatStatus: () => ({ data: { enabled: true } }),
}));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: false }),
}));
vi.mock("@/hooks/use-manual-article-overrides", () => ({
  useGetManualArticleOverrides: () => ({ data: [] }),
}));
vi.mock("../api/get-support-chat-history", () => ({
  useGetSupportChatHistory: () => ({
    data: [{ role: "user", content: "予約の変更方法は？", sources: [] }],
    isSuccess: true,
    isError: false,
    isPending: false,
  }),
}));
vi.mock("../api/send-support-chat", () => ({
  useSendSupportChat: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock("../api/clear-support-chat-history", () => ({
  useClearSupportChatHistory: () => ({ mutate: clearHistory, isPending: false }),
}));

afterEach(() => clearHistory.mockClear());

async function openResetDialog(withSidePeek: boolean) {
  const user = userEvent.setup();
  render(
    <MemoryRouter>
      {withSidePeek ? (
        <aside data-side-peek>
          <button type="button">外部パネルの保存</button>
        </aside>
      ) : null}
      <SupportWidget />
    </MemoryRouter>,
  );
  const external = withSidePeek ? screen.getByRole("button", { name: "外部パネルの保存" }) : null;
  external?.focus();
  const launcher = screen.getByRole("button", { name: "サポート・ヘルプを開く" });
  await user.click(launcher);
  const reset = await screen.findByRole("button", { name: "会話をリセット" });
  await user.click(reset);
  const dialog = await screen.findByRole("alertdialog", {
    name: "会話履歴をリセットしますか？",
  });
  expect(dialog.closest("[data-support-widget]")).toBeNull();
  expect(document.querySelector(`[aria-controls="${CSS.escape(dialog.id)}"]`)).toBeNull();
  expect(dialog).toContainElement(document.activeElement as HTMLElement);
  return { user, external, launcher, reset, dialog };
}

// Radix FocusScope schedules close auto-focus in the next task. Observe that
// cleanup before asserting final focus, rather than accepting a transient focus.
async function flushCloseAutoFocus() {
  await act(async () => {
    await new Promise<void>((resolve) => setTimeout(resolve, 0));
  });
}

describe("SupportWidget — real reset ConfirmDialog portal focus", () => {
  for (const withSidePeek of [false, true]) {
    for (const method of ["close", "Escape"] as const) {
      it(`${withSidePeek ? "SidePeekへ" : "launcherへ"}: reset portalを開いたまま${method}してもfocusが残る`, async () => {
        const { external, launcher, dialog } = await openResetDialog(withSidePeek);
        if (method === "Escape") {
          fireEvent.keyDown(dialog, { key: "Escape" });
        } else {
          // Exercise the existing close callback through the retained launcher.
          // This is a logic regression, not a physical click through a modal.
          fireEvent.click(launcher);
        }
        await waitFor(() => expect(dialog).not.toBeInTheDocument());
        await flushCloseAutoFocus();
        expect(document.querySelector("#support-panel")).toBeNull();
        const target = external ?? launcher;
        expect(target.isConnected).toBe(true);
        expect(target).toHaveFocus();
        expect(clearHistory).not.toHaveBeenCalled();
      });
    }
  }

  it("確認だけをキャンセルするとpanelを保持してreset triggerへ戻す", async () => {
    const { user, reset, dialog } = await openResetDialog(false);
    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(dialog).not.toBeInTheDocument());
    await flushCloseAutoFocus();
    expect(document.querySelector("#support-panel")).not.toBeNull();
    expect(reset).toHaveFocus();
    expect(clearHistory).not.toHaveBeenCalled();
  });
});
