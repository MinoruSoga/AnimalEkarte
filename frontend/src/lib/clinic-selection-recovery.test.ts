import { afterEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import {
  CLINIC_SELECTION_UNAVAILABLE,
  isClinicSelectionUnavailable,
  recoverClinicSelectionOnce,
  retryClinicSelectionRecovery,
  clearClinicSelectionRecovery,
  resetClinicSelectionRecoveryForTests,
  areClinicWritesPaused,
  getClinicSelectionBlockReason,
  cancelPendingClinicSelectionRecovery,
} from "@/lib/clinic-selection-recovery";

const mocks = vi.hoisted(() => ({
  setStoredClinicId: vi.fn(),
  queryClient: { clear: vi.fn() },
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), warning: vi.fn() },
}));

vi.mock("@/lib/current-clinic", () => ({
  setStoredClinicId: (...args: unknown[]) => mocks.setStoredClinicId(...args),
}));

vi.mock("@/lib/react-query", () => ({
  queryClient: mocks.queryClient,
}));

describe("clinic-selection-recovery", () => {
  afterEach(() => {
    resetClinicSelectionRecoveryForTests();
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it("detects only the clinic-selection error code", () => {
    expect(isClinicSelectionUnavailable({ error_code: CLINIC_SELECTION_UNAVAILABLE })).toBe(true);
    expect(isClinicSelectionUnavailable({ error_code: "FORBIDDEN" })).toBe(false);
    expect(isClinicSelectionUnavailable({ message: "forbidden" })).toBe(false);
    expect(isClinicSelectionUnavailable({ code: CLINIC_SELECTION_UNAVAILABLE })).toBe(false);
  });

  it("stores the resolved clinic and does not retry writes", async () => {
    mocks.setStoredClinicId.mockReturnValue(true);
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          main_clinic_id: "2",
          clinics: [{ clinic_id: "2", clinic_name: "城東", is_main: true }],
        }),
      }),
    );

    await recoverClinicSelectionOnce();
    expect(areClinicWritesPaused()).toBe(true);
    expect(getClinicSelectionBlockReason()).toBe("none");
    expect(mocks.setStoredClinicId).toHaveBeenCalledWith("2");
    expect(mocks.queryClient.clear).toHaveBeenCalled();
    expect(reload).toHaveBeenCalled();
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining("/v1/me"),
      expect.objectContaining({
        method: "GET",
        credentials: "include",
      }),
    );
    const [, init] = vi.mocked(fetch).mock.calls[0];
    expect(JSON.stringify(init)).not.toContain("X-Clinic-ID");
  });

  it("blocks the screen when no active clinic remains", async () => {
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        json: async () => ({ error_code: CLINIC_SELECTION_UNAVAILABLE }),
      }),
    );

    await recoverClinicSelectionOnce();
    expect(areClinicWritesPaused()).toBe(true);
    expect(getClinicSelectionBlockReason()).toBe("no-clinic");
    expect(mocks.setStoredClinicId).not.toHaveBeenCalled();
    expect(reload).not.toHaveBeenCalled();
  });

  it("blocks the screen when /me recovery fails", async () => {
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network")));

    await recoverClinicSelectionOnce();
    expect(areClinicWritesPaused()).toBe(true);
    expect(getClinicSelectionBlockReason()).toBe("recovery-failed");
    expect(mocks.setStoredClinicId).not.toHaveBeenCalled();
    expect(reload).not.toHaveBeenCalled();
  });

  it("does not treat storage failure as a successful clinic switch", async () => {
    mocks.setStoredClinicId.mockReturnValue(false);
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          main_clinic_id: "2",
          clinics: [{ clinic_id: "2", clinic_name: "城東", is_main: true }],
        }),
      }),
    );

    await recoverClinicSelectionOnce();
    expect(areClinicWritesPaused()).toBe(true);
    expect(getClinicSelectionBlockReason()).toBe("recovery-failed");
    expect(reload).not.toHaveBeenCalled();
  });

  it("does not restart automatic recovery for later polling errors", async () => {
    const request = vi.fn().mockRejectedValue(new Error("network"));
    vi.stubGlobal("fetch", request);
    await recoverClinicSelectionOnce();
    await recoverClinicSelectionOnce();
    expect(request).toHaveBeenCalledTimes(1);
    expect(areClinicWritesPaused()).toBe(true);

    await retryClinicSelectionRecovery();
    expect(request).toHaveBeenCalledTimes(2);
    await recoverClinicSelectionOnce();
    expect(request).toHaveBeenCalledTimes(2);
  });

  it.each([401, 403, 503])("does not mistake an ordinary %s for no clinic", async (status) => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status,
        json: async () => ({ error: "unavailable" }),
      }),
    );
    await recoverClinicSelectionOnce();
    expect(getClinicSelectionBlockReason()).toBe("recovery-failed");
  });

  it("cancel between setStoredClinicId and toast/reload ignores mutations", async () => {
    mocks.setStoredClinicId.mockImplementation(() => {
      cancelPendingClinicSelectionRecovery();
      return true;
    });
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          main_clinic_id: "2",
          clinics: [{ clinic_id: "2", clinic_name: "城東", is_main: true }],
        }),
      }),
    );

    await recoverClinicSelectionOnce();
    expect(mocks.setStoredClinicId).toHaveBeenCalledWith("2");
    expect(toast.warning).not.toHaveBeenCalled();
    expect(mocks.queryClient.clear).not.toHaveBeenCalled();
    expect(reload).not.toHaveBeenCalled();
  });

  it("ignores a recovery response that arrives after logout", async () => {
    const reload = vi.fn();
    vi.stubGlobal("window", { location: { reload } });
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
    const pending = recoverClinicSelectionOnce();
    clearClinicSelectionRecovery();
    finish(new Response(JSON.stringify({ main_clinic_id: "2", clinics: [{ clinic_id: "2" }] })));
    await pending;
    expect(mocks.setStoredClinicId).not.toHaveBeenCalled();
    expect(reload).not.toHaveBeenCalled();
    expect(getClinicSelectionBlockReason()).toBe("none");
    expect(areClinicWritesPaused()).toBe(false);
  });
});
