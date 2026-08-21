import { afterEach, describe, expect, it, vi } from "vitest";

import {
  LAB_DEVICE_AGENT_URL,
  createLabDeviceAgentClient,
  drainLabDeviceAgentFrames,
} from "./lab-device-agent";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("lab device agent client", () => {
  it("reads raw base64 from the loopback-only agent", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => new Response(JSON.stringify(
      String(input).endsWith("/claim")
        ? { owner: "owner-1" }
        : { frames: [{ id: "frame-1", payload_base64: "Av8D", received_at: "2026-08-20T12:00:00Z" }] },
    ), { status: 200, headers: { "Content-Type": "application/json" } }));
    const client = createLabDeviceAgentClient(fetcher);

    await client.claim("clinic-2");
    await expect(client.frames()).resolves.toEqual([
      { id: "frame-1", payloadBase64: "Av8D", receivedAt: "2026-08-20T12:00:00Z" },
    ]);
    expect(fetcher).toHaveBeenCalledWith(`${LAB_DEVICE_AGENT_URL}/frames`, expect.objectContaining({ method: "GET" }));
  });

  it("rejects malformed agent responses", async () => {
    const client = createLabDeviceAgentClient(async (input) => new Response(
      JSON.stringify(String(input).endsWith("/claim")
        ? { owner: "owner-1" }
        : { frames: [{ id: "frame-1", payload_base64: 123 }] }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    ));
    await client.claim("clinic-2");
    await expect(client.frames()).rejects.toThrow("invalid lab device agent response");
  });

  it("maps health and sends explicit frame decisions", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/health")) {
        return new Response(JSON.stringify({
          status: "degraded",
          open_ports: 2,
          configured_ports: 2,
          queue: { pending: 1, rejected: 2, overflow: 3 },
          input_overflow: 4,
          port_discovery_failures_total: 5,
          port_open_failures_total: 6,
          queue_failures_total: 7,
          port_close_failures_total: 8,
          response_failures_total: 9,
          last_error_category: "port_open_failed",
        }), { status: 200, headers: { "Content-Type": "application/json" } });
      }
      if (url.endsWith("/claim")) {
        return new Response(JSON.stringify({ owner: "owner-1" }), { status: 200 });
      }
      return new Response(null, { status: 204 });
    });
    const client = createLabDeviceAgentClient(fetcher);

    await expect(client.health()).resolves.toEqual({
      status: "degraded", openPorts: 2, configuredPorts: 2, pending: 1, rejected: 2, overflow: 3, inputOverflow: 4,
      portDiscoveryFailures: 5, portOpenFailures: 6, lastErrorCategory: "port_open_failed",
      queueFailures: 7, portCloseFailures: 8,
      responseFailures: 9,
    });
    await client.claim("clinic-2");
    await client.ack("frame / 1");
    await client.reject("frame-2");
    expect(fetcher).toHaveBeenCalledWith(
      `${LAB_DEVICE_AGENT_URL}/frames/frame%20%2F%201/ack`,
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("reports non-successful health and decision responses", async () => {
    const client = createLabDeviceAgentClient(async () => new Response(null, { status: 503 }));
    await expect(client.health()).rejects.toThrow("503");
    await expect(client.ack("frame-1")).rejects.toThrow("503");
  });

  it("renews a live claim and reacquires it after a lease conflict", async () => {
    let claimCount = 0;
    let frameCount = 0;
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/claim")) {
        claimCount += 1;
        return new Response(JSON.stringify({ owner: `owner-${claimCount}` }), { status: 200 });
      }
      frameCount += 1;
      if (frameCount === 1) {
        return new Response(null, { status: 409 });
      }
      return new Response(JSON.stringify({ frames: [] }), { status: 200 });
    });
    const client = createLabDeviceAgentClient(fetcher);
    await client.claim("clinic-2");
    await client.claim("clinic-2");
    expect(claimCount).toBe(2);
    await expect(client.frames()).rejects.toThrow("409");
    await client.claim("clinic-2");
    await expect(client.frames()).resolves.toEqual([]);
    expect(claimCount).toBe(3);
  });

  it("rejects an invalid claim response and use before claim", async () => {
    const client = createLabDeviceAgentClient(async () => new Response(JSON.stringify({ owner: "" }), { status: 200 }));
    await expect(client.frames()).rejects.toThrow("not claimed");
    await expect(client.claim("clinic-2")).rejects.toThrow("invalid lab device agent response");
  });
});

describe("drainLabDeviceAgentFrames", () => {
  it("sends bytes unchanged with auto detection and acknowledges only success", async () => {
    const receive = vi.fn(async () => {});
    const ack = vi.fn(async () => {});
    const client = {
      claim: vi.fn(),
      health: vi.fn(),
      frames: vi.fn(async () => [{ id: "frame-1", payloadBase64: "Av8D", receivedAt: "2026-08-20T12:00:00Z" }]),
      ack,
      reject: vi.fn(),
    };

    await drainLabDeviceAgentFrames({ client, receive });

    expect(receive).toHaveBeenCalledWith({ payloadBase64: "Av8D", deviceHint: "auto" });
    expect(ack).toHaveBeenCalledWith("frame-1");
  });

  it("retains retryable failures and rejects only invalid payloads", async () => {
    const frames = [
      { id: "invalid", payloadBase64: "AA==", receivedAt: "2026-08-20T12:00:00Z" },
      { id: "retry", payloadBase64: "AQ==", receivedAt: "2026-08-20T12:00:01Z" },
      { id: "later", payloadBase64: "Ag==", receivedAt: "2026-08-20T12:00:02Z" },
    ];
    const reject = vi.fn(async () => {});
    const client = { claim: vi.fn(), health: vi.fn(), frames: vi.fn(async () => frames), ack: vi.fn(), reject };
    const receive = vi.fn(async (input: { payloadBase64: string }) => {
      throw Object.assign(new Error("receive failed"), { status: input.payloadBase64 === "AA==" ? 400 : 500 });
    });

    const result = await drainLabDeviceAgentFrames({ client, receive });

    expect(reject).toHaveBeenCalledWith("invalid");
    expect(client.ack).not.toHaveBeenCalled();
    expect(reject).not.toHaveBeenCalledWith("retry");
    expect(receive).toHaveBeenCalledTimes(2);
    expect(result.retryableFailure).toBe(true);
  });

  it("passes abort signals through terminal decisions and stops before delivery", async () => {
    const controller = new AbortController();
    const receive = vi.fn(async () => {});
    const ack = vi.fn(async () => {});
    const client = {
      claim: vi.fn(),
      health: vi.fn(),
      frames: vi.fn(async () => [{ id: "frame-1", payloadBase64: "AA==", receivedAt: "now" }]),
      ack,
      reject: vi.fn(),
    };
    controller.abort();
    await drainLabDeviceAgentFrames({ client, receive, signal: controller.signal });
    expect(receive).not.toHaveBeenCalled();
    expect(ack).not.toHaveBeenCalled();
  });
});
