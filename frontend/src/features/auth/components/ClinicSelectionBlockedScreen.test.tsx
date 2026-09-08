import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ClinicSelectionBlockedScreen } from "./ClinicSelectionBlockedScreen";
import {
  pauseClinicWrites,
  resetClinicSelectionRecoveryForTests,
  recoverClinicSelectionOnce,
} from "@/lib/clinic-selection-recovery";

vi.mock("@/lib/current-clinic", () => ({
  setStoredClinicId: vi.fn(),
}));

vi.mock("@/lib/react-query", () => ({
  queryClient: { clear: vi.fn() },
}));

describe("ClinicSelectionBlockedScreen", () => {
  afterEach(() => {
    resetClinicSelectionRecoveryForTests();
    vi.unstubAllGlobals();
  });

  it("offers logout when no clinic remains", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ clinics: [] }),
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
  });

  it("renders nothing while clinic selection is available", () => {
    pauseClinicWrites();
    const { container } = render(<ClinicSelectionBlockedScreen onLogout={async () => undefined} />);
    expect(container).toBeEmptyDOMElement();
  });
});
