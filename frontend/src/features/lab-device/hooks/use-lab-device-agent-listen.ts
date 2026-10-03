import { useEffect, useLayoutEffect, useRef, useState } from "react";

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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function useLabDeviceAgentListen(input: {
  enabled: boolean;
  clinicId: string | null;
  consumerToken?: string;
  onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  onUnauthorized?: () => void;
  client?: LabDeviceAgentClient;
}): LabDeviceAgentListenStatus {
  const [snapshot, setSnapshot] = useState<{
    clinicId: string | null;
    status: LabDeviceAgentListenStatus;
  }>({ clinicId: null, status: disconnectedStatus });
  const onFrameRef = useRef(input.onFrame);
  const enabledRef = useRef(input.enabled);
  const onUnauthorizedRef = useRef(input.onUnauthorized);
  const consumerTokenRef = useRef(input.consumerToken);
  useLayoutEffect(() => {
    onFrameRef.current = input.onFrame;
  }, [input.onFrame]);
  useLayoutEffect(() => {
    enabledRef.current = input.enabled;
  }, [input.enabled]);
  useLayoutEffect(() => {
    onUnauthorizedRef.current = input.onUnauthorized;
  }, [input.onUnauthorized]);
  useLayoutEffect(() => {
    consumerTokenRef.current = input.consumerToken;
  }, [input.consumerToken]);
  const hasConsumerToken = input.consumerToken !== undefined;

  useEffect(() => {
    if (!input.enabled) {
      return;
    }
    // client は effect 内で構築する（capability getter が最新値を ref 経由で読む
    // ため render スコープの useMemo では react-hooks/refs に抵触する）。
    // deps に consumerToken 値自体ではなく hasConsumerToken のみを入れることで、
    // capability ローテーション時に effect が再実行されず claim も再走しない。
    const client =
      input.client ??
      (hasConsumerToken
        ? createLabDeviceAgentClient(() => consumerTokenRef.current ?? "")
        : undefined);
    if (!client) {
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
      } catch (error: unknown) {
        if (!controller.signal.aborted) {
          setSnapshot({ clinicId: input.clinicId, status: disconnectedStatus });
          // capability の期限切れ・他医院束縛を検知したら再取得を促す
          if (isRecord(error) && (error.status === 401 || error.status === 403)) {
            onUnauthorizedRef.current?.();
          }
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
  }, [input.client, hasConsumerToken, input.clinicId, input.enabled]);

  return input.enabled && snapshot.clinicId === input.clinicId
    ? snapshot.status
    : disconnectedStatus;
}
