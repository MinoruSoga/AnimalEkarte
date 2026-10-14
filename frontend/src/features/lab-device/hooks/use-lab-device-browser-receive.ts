import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";

import {
  createLabDeviceBrowserReceiver,
  type LabDeviceBrowserReceiveStatus,
} from "../lib/lab-device-browser-receiver";

const EMPTY_STATUS: LabDeviceBrowserReceiveStatus = {
  supported: false,
  connected: false,
  openPorts: 0,
  pending: 0,
  rejected: 0,
  overflow: 0,
  authPaused: false,
  error: null,
};

export function useLabDeviceBrowserReceive(input: {
  enabled: boolean;
  clinicId: string | null;
  onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
}): LabDeviceBrowserReceiveStatus & {
  connect(): Promise<void>;
  retry(): void;
  stop(): Promise<void>;
} {
  const current = useRef(input);
  const receiver = useRef<ReturnType<typeof createLabDeviceBrowserReceiver> | null>(null);
  const scopedFrame = useRef<{ clinicId: string; onFrame: typeof input.onFrame } | null>(null);
  const [restart, setRestart] = useState(0);
  const [snapshot, setSnapshot] = useState<{
    clinicId: string | null;
    status: LabDeviceBrowserReceiveStatus;
  }>({ clinicId: null, status: EMPTY_STATUS });
  useLayoutEffect(() => {
    const previous = current.current;
    current.current = input;
    if (
      input.enabled &&
      input.clinicId !== null &&
      previous.enabled &&
      previous.clinicId === input.clinicId &&
      scopedFrame.current?.clinicId === input.clinicId
    ) {
      scopedFrame.current.onFrame = input.onFrame;
    } else if (input.enabled && input.clinicId !== null) {
      scopedFrame.current = { clinicId: input.clinicId, onFrame: input.onFrame };
    } else {
      scopedFrame.current = null;
    }
  }, [input]);

  useEffect(() => {
    const clinicId = input.clinicId;
    if (!input.enabled || clinicId === null) {
      receiver.current = null;
      return;
    }
    let active = true;
    const frameScope = scopedFrame.current;
    if (!frameScope || frameScope.clinicId !== clinicId) return;
    const instance = createLabDeviceBrowserReceiver({
      clinicId,
      onFrame: async (frame) => {
        if (!active || current.current.enabled !== true || current.current.clinicId !== clinicId)
          return;
        await frameScope.onFrame(frame);
      },
      onStatus: (status) => {
        if (active) setSnapshot({ clinicId, status });
      },
      isEnabled: () =>
        active && current.current.enabled === true && current.current.clinicId === clinicId,
    });
    receiver.current = instance;
    return () => {
      active = false;
      if (receiver.current === instance) receiver.current = null;
      void instance.stop();
    };
  }, [input.clinicId, input.enabled, restart]);

  const connect = useCallback(() => receiver.current?.connect() ?? Promise.resolve(), []);
  const retry = useCallback(() => receiver.current?.retry(), []);
  const stop = useCallback(async () => {
    const instance = receiver.current;
    receiver.current = null;
    await instance?.stop();
    const scope = current.current;
    if (scope.enabled && scope.clinicId !== null && receiver.current === null)
      setRestart((value) => value + 1);
  }, []);
  const status =
    input.enabled && snapshot.clinicId === input.clinicId ? snapshot.status : EMPTY_STATUS;
  return { ...status, connect, retry, stop };
}
