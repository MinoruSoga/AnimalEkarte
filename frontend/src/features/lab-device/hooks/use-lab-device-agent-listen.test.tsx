import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { LabDeviceAgentClient } from "../lib/lab-device-agent";
import { useLabDeviceAgentListen } from "./use-lab-device-agent-listen";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

function client(overrides: Partial<LabDeviceAgentClient> = {}): LabDeviceAgentClient {
  return {
    claim: vi.fn(async () => {}),
    health: vi.fn(async () => ({
      status: "running", openPorts: 2, configuredPorts: 2, pending: 0, rejected: 0, overflow: 0, inputOverflow: 0,
      portDiscoveryFailures: 0, portOpenFailures: 0, lastErrorCategory: "none",
      queueFailures: 0, portCloseFailures: 0,
      responseFailures: 0,
    })),
    frames: vi.fn(async () => []),
    ack: vi.fn(async () => {}),
    reject: vi.fn(async () => {}),
    ...overrides,
  };
}

describe("useLabDeviceAgentListen", () => {
  it("polls immediately without Web Serial and reports monitored ports", async () => {
    const agent = client();
    const { result } = renderHook(() => useLabDeviceAgentListen({
      enabled: true,
      clinicId: "clinic-2",
      client: agent,
      onFrame: async () => {},
    }));

    await waitFor(() => expect(result.current.connected).toBe(true));
    expect(result.current.openPorts).toBe(2);
    expect(result.current.configuredPorts).toBe(2);
    expect(agent.health).toHaveBeenCalledTimes(1);
    expect(agent.frames).toHaveBeenCalledTimes(1);
  });

  it("does not overlap polls and aborts on unmount", async () => {
    vi.useFakeTimers();
    let resolveFrames!: () => void;
    const frames = vi.fn(() => new Promise<[]>(resolve => { resolveFrames = () => resolve([]); }));
    const agent = client({ frames });
    const { unmount } = renderHook(() => useLabDeviceAgentListen({
      enabled: true,
      clinicId: "clinic-2",
      client: agent,
      onFrame: async () => {},
    }));

    await act(async () => { await Promise.resolve(); });
    await act(async () => { await vi.advanceTimersByTimeAsync(5000); });
    expect(frames).toHaveBeenCalledTimes(1);

    resolveFrames();
    await act(async () => { await Promise.resolve(); });
    unmount();
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(frames).toHaveBeenCalledTimes(1);
  });

  it("does not poll when disabled and recovers to disconnected after agent failure", async () => {
    const health = vi.fn().mockResolvedValueOnce({
      status: "degraded", openPorts: 1, configuredPorts: 2, pending: 1, rejected: 1, overflow: 0,
      inputOverflow: 0, portDiscoveryFailures: 1, portOpenFailures: 2,
      lastErrorCategory: "port_open_failed",
      queueFailures: 0, portCloseFailures: 0,
      responseFailures: 0,
    }).mockRejectedValueOnce(new Error("offline"));
    const agent = client({ health });
    const { result, rerender } = renderHook(
      ({ enabled }) => useLabDeviceAgentListen({
        enabled, clinicId: "clinic-2", client: agent, onFrame: async () => {},
      }),
      { initialProps: { enabled: false } },
    );
    expect(result.current.connected).toBe(false);
    expect(health).not.toHaveBeenCalled();

    rerender({ enabled: true });
    await waitFor(() => expect(result.current.degraded).toBe(true));
    expect(result.current.portOpenFailures).toBe(2);
    expect(result.current.lastErrorCategory).toBe("port_open_failed");
    await act(async () => { await new Promise(resolve => window.setTimeout(resolve, 800)); });
    await waitFor(() => expect(result.current.connected).toBe(false));
  });

  it("backs off after a retryable backend delivery failure", async () => {
    vi.useFakeTimers();
    const frames = vi.fn(async () => [{ id: "frame-1", payloadBase64: "AA==", receivedAt: "now" }]);
    const agent = client({ frames });
    renderHook(() => useLabDeviceAgentListen({
      enabled: true,
      clinicId: "clinic-2",
      client: agent,
      onFrame: async () => { throw Object.assign(new Error("offline"), { status: 500 }); },
    }));

    await act(async () => { await Promise.resolve(); await Promise.resolve(); });
    expect(frames).toHaveBeenCalledTimes(1);
    await act(async () => { await vi.advanceTimersByTimeAsync(1000); });
    expect(frames).toHaveBeenCalledTimes(1);
    await act(async () => { await vi.advanceTimersByTimeAsync(600); });
    expect(frames).toHaveBeenCalledTimes(2);
  });

  it("does not expose the previous clinic snapshot while switching clinics", async () => {
    const claim = vi.fn(async (clinicId: string, signal?: AbortSignal) => {
      if (clinicId === "clinic-3") {
        await new Promise<void>((resolve) => signal?.addEventListener("abort", () => resolve(), { once: true }));
      }
    });
    const agent = client({ claim });
    const { result, rerender, unmount } = renderHook(
      ({ clinicId }) => useLabDeviceAgentListen({
        enabled: true, clinicId, client: agent, onFrame: async () => {},
      }),
      { initialProps: { clinicId: "clinic-2" } },
    );
    await waitFor(() => expect(result.current.connected).toBe(true));

    act(() => {
      rerender({ clinicId: "clinic-3" });
    });

    expect(result.current.connected).toBe(false);
    expect(result.current.openPorts).toBe(0);
    unmount();
  });
});
