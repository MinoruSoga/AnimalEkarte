import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { SessionRestoreError } from "./SessionRestoreError";

describe("SessionRestoreError", () => {
  it("announces the restore failure with role=alert, not role=status", () => {
    render(<SessionRestoreError onRetry={() => undefined} onSwitchToLogin={() => undefined} />);

    expect(screen.getByRole("alert")).toHaveTextContent("ログイン状態を確認できませんでした");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByText("ノア動物病院")).toBeInTheDocument();
  });

  it("exposes keyboard-reachable retry and login-switch actions without logout", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    const onSwitchToLogin = vi.fn();
    render(<SessionRestoreError onRetry={onRetry} onSwitchToLogin={onSwitchToLogin} />);

    const retry = screen.getByRole("button", { name: "再試行" });
    const loginSwitch = screen.getByRole("button", { name: "ログイン切替" });
    expect(retry).toBeEnabled();
    expect(loginSwitch).toBeEnabled();
    expect(screen.queryByRole("button", { name: "ログアウト" })).not.toBeInTheDocument();

    retry.focus();
    expect(retry).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(onRetry).toHaveBeenCalledOnce();

    loginSwitch.focus();
    expect(loginSwitch).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(onSwitchToLogin).toHaveBeenCalledOnce();
  });
});
