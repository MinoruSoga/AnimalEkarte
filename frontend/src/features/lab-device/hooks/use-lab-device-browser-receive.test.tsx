import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

const { instances } = vi.hoisted(() => ({
  instances: [] as Array<{
    input: {
      onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
      isEnabled: () => boolean;
    };
    stop: ReturnType<typeof vi.fn>;
  }>,
}));

vi.mock("../lib/lab-device-browser-receiver", () => ({
  createLabDeviceBrowserReceiver: (input: {
    onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
    isEnabled: () => boolean;
  }) => {
    const instance = { input, stop: vi.fn().mockResolvedValue(undefined) };
    instances.push(instance);
    return { connect: vi.fn(), retry: vi.fn(), stop: instance.stop };
  },
}));

import { useLabDeviceBrowserReceive } from "./use-lab-device-browser-receive";

afterEach(() => instances.splice(0));

describe("useLabDeviceBrowserReceive scope teardown", () => {
  it("stops the previous clinic instance and never sends its queued frame to the new scope", async () => {
    const clinicAFrame = vi.fn().mockResolvedValue(undefined);
    const clinicBFrame = vi.fn().mockResolvedValue(undefined);
    const { rerender, unmount } = renderHook(
      (input: { clinicId: string; onFrame: typeof clinicAFrame }) =>
        useLabDeviceBrowserReceive({ enabled: true, ...input }),
      { initialProps: { clinicId: "clinic-a", onFrame: clinicAFrame } },
    );
    const previous = instances[0];
    rerender({ clinicId: "clinic-b", onFrame: clinicBFrame });
    await act(async () => previous.input.onFrame({ payloadBase64: "AQ==", deviceHint: "auto" }));
    expect(previous.stop).toHaveBeenCalledTimes(1);
    expect(clinicAFrame).not.toHaveBeenCalled();
    expect(clinicBFrame).not.toHaveBeenCalled();
    unmount();
    expect(instances[1].stop).toHaveBeenCalledTimes(1);
  });

  it("refreshes the callback within a clinic without restarting or switching the old scope", async () => {
    const clinicA1 = vi.fn().mockResolvedValue(undefined);
    const clinicA2 = vi.fn().mockResolvedValue(undefined);
    const clinicB1 = vi.fn().mockResolvedValue(undefined);
    const { rerender, unmount } = renderHook(
      (input: { clinicId: string; onFrame: typeof clinicA1 }) =>
        useLabDeviceBrowserReceive({ enabled: true, ...input }),
      { initialProps: { clinicId: "clinic-a", onFrame: clinicA1 } },
    );
    const activeA = instances[0];
    rerender({ clinicId: "clinic-a", onFrame: clinicA2 });
    expect(instances).toHaveLength(1);
    await act(async () => activeA.input.onFrame({ payloadBase64: "AQ==", deviceHint: "auto" }));
    expect(clinicA1).not.toHaveBeenCalled();
    expect(clinicA2).toHaveBeenCalledTimes(1);

    rerender({ clinicId: "clinic-b", onFrame: clinicB1 });
    await act(async () => activeA.input.onFrame({ payloadBase64: "Ag==", deviceHint: "auto" }));
    expect(clinicA2).toHaveBeenCalledTimes(1);
    expect(clinicB1).not.toHaveBeenCalled();
    unmount();
  });

  it("stops immediately when receive permission is lost", async () => {
    const onFrame = vi.fn().mockResolvedValue(undefined);
    const { rerender } = renderHook(
      (enabled: boolean) => useLabDeviceBrowserReceive({ enabled, clinicId: "clinic-a", onFrame }),
      { initialProps: true },
    );
    const previous = instances[0];
    rerender(false);
    await act(async () => previous.input.onFrame({ payloadBase64: "AQ==", deviceHint: "auto" }));
    expect(previous.stop).toHaveBeenCalledTimes(1);
    expect(onFrame).not.toHaveBeenCalled();
  });
});
