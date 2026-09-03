import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import {
  createLabDeviceAgentClient,
  drainLabDeviceAgentFrames,
  type LabDeviceAgentClient,
  type LabDeviceAgentHealth,
} from "../lib/lab-device-agent";

const POLL_INTERVAL_MS = 750;
const MAX_RETRY_INTERVAL_MS = 30_000;

export interface LabDeviceAgentListenStatus {
  connected: boolean;
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
  lastErrorCategory: LabDeviceAgentHealth["lastErrorCategory"];
  degraded: boolean;
}

const disconnectedStatus: LabDeviceAgentListenStatus = {
  connected: false,
  openPorts: 0,
  configuredPorts: 0,
  pending: 0,
  rejected: 0,
  overflow: 0,
  inputOverflow: 0,
  portDiscoveryFailures: 0,
  portOpenFailures: 0,
  queueFailures: 0,
  portCloseFailures: 0,
  responseFailures: 0,
  lastErrorCategory: "none",
  degraded: false,
};

function toStatus(health: LabDeviceAgentHealth): LabDeviceAgentListenStatus {
  return {
    connected: true,
    openPorts: health.openPorts,
    configuredPorts: health.configuredPorts,
    pending: health.pending,
    rejected: health.rejected,
    overflow: health.overflow,
    inputOverflow: health.inputOverflow,
    portDiscoveryFailures: health.portDiscoveryFailures,
    portOpenFailures: health.portOpenFailures,
    queueFailures: health.queueFailures,
    portCloseFailures: health.portCloseFailures,
    responseFailures: health.responseFailures,
    lastErrorCategory: health.lastErrorCategory,
    degraded: health.status === "degraded",
  };
}

export function useLabDeviceAgentListen(input: {
  enabled: boolean;
  clinicId: string | null;
  consumerToken?: string;
  onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  client?: LabDeviceAgentClient;
}): LabDeviceAgentListenStatus {
  const [snapshot, setSnapshot] = useState<{
    clinicId: string | null;
    status: LabDeviceAgentListenStatus;
  }>({ clinicId: null, status: disconnectedStatus });
  const onFrameRef = useRef(input.onFrame);
  const enabledRef = useRef(input.enabled);
  useLayoutEffect(() => {
    onFrameRef.current = input.onFrame;
  }, [input.onFrame]);
  useLayoutEffect(() => {
    enabledRef.current = input.enabled;
  }, [input.enabled]);
  const client = useMemo(
    () =>
      input.client ??
      (input.consumerToken ? createLabDeviceAgentClient(input.consumerToken) : undefined),
    [input.client, input.consumerToken],
  );

  useEffect(() => {
    if (!input.enabled || !client) {
      return;
    }
    const controller = new AbortController();
    let timer: number | undefined;
    let retryInterval = POLL_INTERVAL_MS;
    const poll = async (): Promise<void> => {
      let nextInterval = POLL_INTERVAL_MS;
      try {
        if (input.clinicId === null) {
          return;
        }
        await client.claim(input.clinicId, controller.signal);
        const health = await client.health(controller.signal);
        if (controller.signal.aborted) {
          return;
        }
        setSnapshot({ clinicId: input.clinicId, status: toStatus(health) });
        const drained = await drainLabDeviceAgentFrames({
          client,
          signal: controller.signal,
          receive: (frame) => {
            if (enabledRef.current !== true) {
              return Promise.resolve();
            }
            return onFrameRef.current(frame);
          },
        });
        if (drained.retryableFailure) {
          retryInterval = Math.min(retryInterval * 2, MAX_RETRY_INTERVAL_MS);
          nextInterval = retryInterval;
        } else {
          retryInterval = POLL_INTERVAL_MS;
        }
      } catch {
        if (!controller.signal.aborted) {
          setSnapshot({ clinicId: input.clinicId, status: disconnectedStatus });
        }
        retryInterval = Math.min(retryInterval * 2, MAX_RETRY_INTERVAL_MS);
        nextInterval = retryInterval;
      }
      if (!controller.signal.aborted) {
        timer = window.setTimeout(() => void poll(), nextInterval);
      }
    };
    void poll();
    return () => {
      controller.abort();
      if (timer !== undefined) {
        window.clearTimeout(timer);
      }
    };
  }, [client, input.clinicId, input.enabled]);

  return input.enabled && snapshot.clinicId === input.clinicId
    ? snapshot.status
    : disconnectedStatus;
}
