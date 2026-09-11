import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ClinicSelectionBlockedScreen } from "./ClinicSelectionBlockedScreen";
import {
  pauseClinicWrites,
  resetClinicSelectionRecoveryForTests,
  recoverClinicSelectionOnce,
  clearClinicSelectionRecovery,
  CLINIC_SELECTION_UNAVAILABLE,
} from "@/lib/clinic-selection-recovery";

vi.mock("@/lib/current-clinic", () => ({
  setStoredClinicId: vi.fn(),
}));

vi.mock("@/lib/react-query", () => ({
  queryClient: { clear: vi.fn() },
}));

describe("ClinicSelectionBlockedScreen", () => {
  afterEach(() => {
    cleanup();
    resetClinicSelectionRecoveryForTests();
    vi.unstubAllGlobals();
  });

  it("offers logout when no clinic remains", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        json: async () => ({ error_code: CLINIC_SELECTION_UNAVAILABLE }),
      }),
    );
    await act(async () => {
      await recoverClinicSelectionOnce();
    });
    const onLogout = vi.fn().mockResolvedValue(undefined);
    render(<ClinicSelectionBlockedScreen onLogout={onLogout} />);

    expect(screen.getByRole("alertdialog")).toHaveTextContent("利用できる医院がありません");
    await userEvent.click(screen.getByRole("button", { name: "ログアウト" }));
    expect(onLogout).toHaveBeenCalled();
  });

  it("offers retry when recovery failed", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network")));
    await act(async () => {
      await recoverClinicSelectionOnce();
    });
    render(<ClinicSelectionBlockedScreen onLogout={async () => undefined} />);

    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ログアウト" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "再試行" }));
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));
  });

  it("renders nothing while clinic selection is available", () => {
    pauseClinicWrites();
    const { container } = render(<ClinicSelectionBlockedScreen onLogout={async () => undefined} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("keeps logout available while a manual retry is pending", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network")));
    await act(async () => {
      await recoverClinicSelectionOnce();
    });
    let finish: (response: Response) => void = () => undefined;
    vi.stubGlobal(
      "fetch",
      vi.fn(
        () =>
          new Promise<Response>((resolve) => {
            finish = resolve;
          }),
      ),
    );
    const onLogout = vi.fn(async () => {
      clearClinicSelectionRecovery();
    });
    render(<ClinicSelectionBlockedScreen onLogout={onLogout} />);
    await userEvent.click(screen.getByRole("button", { name: "再試行" }));
    expect(screen.getByRole("button", { name: "再試行" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "ログアウト" })).toBeEnabled();
    await userEvent.click(screen.getByRole("button", { name: "ログアウト" }));
    await act(async () => {
      finish(new Response(null, { status: 503 }));
    });
    expect(onLogout).toHaveBeenCalledOnce();
  });
});
