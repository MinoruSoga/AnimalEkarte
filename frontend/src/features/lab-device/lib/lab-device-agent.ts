export const LAB_DEVICE_AGENT_URL = "http://127.0.0.1:17654";

export interface LabDeviceAgentHealth {
  status: "running" | "degraded";
  openPorts: number;
  configuredPorts: number;
  pending: number;
  rejected: number;
  overflow: number;
  inputOverflow: number;
  portDiscoveryFailures: number;
  portOpenFailures: number;
  queueFailures: number;
  portCloseFailures: number;
  responseFailures: number;
  lastErrorCategory:
    | "none"
    | "discovery_failed"
    | "port_open_failed"
    | "queue_write_failed"
    | "port_close_failed"
    | "response_write_failed";
}

interface LabDeviceAgentFrame {
  id: string;
  payloadBase64: string;
  receivedAt: string;
}

export interface LabDeviceAgentClient {
  claim: (clinicId: string, signal?: AbortSignal) => Promise<void>;
  health: (signal?: AbortSignal) => Promise<LabDeviceAgentHealth>;
  frames: (signal?: AbortSignal) => Promise<LabDeviceAgentFrame[]>;
  ack: (id: string, signal?: AbortSignal) => Promise<void>;
  reject: (id: string, signal?: AbortSignal) => Promise<void>;
}

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

async function requestJSON(
  fetcher: Fetcher,
  path: string,
  signal?: AbortSignal,
  headers?: HeadersInit,
): Promise<unknown> {
  const response = await fetcher(`${LAB_DEVICE_AGENT_URL}${path}`, {
    method: "GET",
    signal,
    headers,
  });
  if (!response.ok) {
    throw Object.assign(new Error(`lab device agent request failed: ${response.status}`), {
      status: response.status,
    });
  }
  return response.json() as Promise<unknown>;
}

async function decideFrame(
  fetcher: Fetcher,
  id: string,
  decision: "ack" | "reject",
  signal?: AbortSignal,
  headers?: HeadersInit,
): Promise<void> {
  const response = await fetcher(
    `${LAB_DEVICE_AGENT_URL}/frames/${encodeURIComponent(id)}/${decision}`,
    { method: "POST", signal, headers },
  );
  if (!response.ok) {
    throw new Error(`lab device agent decision failed: ${response.status}`);
  }
}

export function createLabDeviceAgentClient(
  consumerToken: string,
  fetcher: Fetcher = fetch,
): LabDeviceAgentClient {
  let owner = "";
  let clinicId = "";
  const consumerHeaders = (): HeadersInit => {
    if (owner === "" || clinicId === "") {
      throw new Error("lab device agent consumer is not claimed");
    }
    return {
      "X-Clinic-ID": clinicId,
      "X-Lab-Device-Owner": owner,
      "X-Lab-Device-Consumer-Token": consumerToken,
    };
  };
  return {
    claim: async (nextClinicId, signal) => {
      const response = await fetcher(`${LAB_DEVICE_AGENT_URL}/claim`, {
        method: "POST",
        signal,
        headers: {
          "Content-Type": "application/json",
          "X-Lab-Device-Consumer-Token": consumerToken,
          ...(owner !== "" && clinicId === nextClinicId ? { "X-Lab-Device-Owner": owner } : {}),
        },
        body: JSON.stringify({ clinic_id: nextClinicId }),
      });
      if (!response.ok) {
        throw new Error(`lab device agent claim failed: ${response.status}`);
      }
      const value: unknown = await response.json();
      if (!isRecord(value) || typeof value.owner !== "string" || value.owner === "") {
        throw new Error("invalid lab device agent response");
      }
      owner = value.owner;
      clinicId = nextClinicId;
    },
    health: async (signal) => {
      const value = await requestJSON(fetcher, "/health", signal);
      if (
        !isRecord(value) ||
        (value.status !== "running" && value.status !== "degraded") ||
        !isNonNegativeInteger(value.open_ports) ||
        !isNonNegativeInteger(value.configured_ports) ||
        !isRecord(value.queue) ||
        !isNonNegativeInteger(value.queue.pending) ||
        !isNonNegativeInteger(value.queue.rejected) ||
        !isNonNegativeInteger(value.queue.overflow) ||
        !isNonNegativeInteger(value.input_overflow) ||
        !isNonNegativeInteger(value.port_discovery_failures_total) ||
        !isNonNegativeInteger(value.port_open_failures_total) ||
        !isNonNegativeInteger(value.queue_failures_total) ||
        !isNonNegativeInteger(value.port_close_failures_total) ||
        !isNonNegativeInteger(value.response_failures_total) ||
        (value.last_error_category !== "none" &&
          value.last_error_category !== "discovery_failed" &&
          value.last_error_category !== "port_open_failed" &&
          value.last_error_category !== "queue_write_failed" &&
          value.last_error_category !== "port_close_failed" &&
          value.last_error_category !== "response_write_failed")
      ) {
        throw new Error("invalid lab device agent response");
      }
      return {
        status: value.status,
        openPorts: value.open_ports,
        configuredPorts: value.configured_ports,
        pending: value.queue.pending,
        rejected: value.queue.rejected,
        overflow: value.queue.overflow,
        inputOverflow: value.input_overflow,
        portDiscoveryFailures: value.port_discovery_failures_total,
        portOpenFailures: value.port_open_failures_total,
        queueFailures: value.queue_failures_total,
        portCloseFailures: value.port_close_failures_total,
        responseFailures: value.response_failures_total,
        lastErrorCategory: value.last_error_category,
      };
    },
    frames: async (signal) => {
      let value: unknown;
      try {
        value = await requestJSON(fetcher, "/frames", signal, consumerHeaders());
      } catch (error: unknown) {
        if (isRecord(error) && error.status === 409) {
          owner = "";
        }
        throw error;
      }
      if (!isRecord(value) || !Array.isArray(value.frames)) {
        throw new Error("invalid lab device agent response");
      }
      return value.frames.map((frame: unknown) => {
        if (
          !isRecord(frame) ||
          typeof frame.id !== "string" ||
          typeof frame.payload_base64 !== "string" ||
          typeof frame.received_at !== "string"
        ) {
          throw new Error("invalid lab device agent response");
        }
        return {
          id: frame.id,
          payloadBase64: frame.payload_base64,
          receivedAt: frame.received_at,
        };
      });
    },
    ack: (id, signal) => decideFrame(fetcher, id, "ack", signal, consumerHeaders()),
    reject: (id, signal) => decideFrame(fetcher, id, "reject", signal, consumerHeaders()),
  };
}

export async function drainLabDeviceAgentFrames(input: {
  client: LabDeviceAgentClient;
  receive: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  signal?: AbortSignal;
}): Promise<{ retryableFailure: boolean }> {
  const frames = await input.client.frames(input.signal);
  for (const frame of frames) {
    if (input.signal?.aborted) {
      return { retryableFailure: false };
    }
    try {
      await input.receive({ payloadBase64: frame.payloadBase64, deviceHint: "auto" });
      if (input.signal) {
        await input.client.ack(frame.id, input.signal);
      } else {
        await input.client.ack(frame.id);
      }
    } catch (error: unknown) {
      const status = isRecord(error) && typeof error.status === "number" ? error.status : undefined;
      if (status === 400) {
        if (input.signal) {
          await input.client.reject(frame.id, input.signal);
        } else {
          await input.client.reject(frame.id);
        }
      } else {
        return { retryableFailure: true };
      }
    }
  }
  return { retryableFailure: false };
}
