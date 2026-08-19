import { useEffect, useMemo, useRef, useState } from "react";

import type { LabDeviceSlot } from "../api/lab-device";
import {
  isWebSerialSupported,
  labDeviceListenState,
  type LabDeviceListenState,
} from "../lib/lab-device-board-model";
import { readStoredPortInfo, startLabDeviceSlotListen } from "../lib/lab-device-serial";

export function useLabDeviceListen(input: {
  slots: LabDeviceSlot[];
  enabled: boolean;
  listenEpoch: number;
  onFrame: (deviceHint: string, bytes: Uint8Array) => Promise<void>;
}): Record<string, LabDeviceListenState> {
  const serialSupported = isWebSerialSupported();
  const [states, setStates] = useState<Record<string, LabDeviceListenState>>({});
  const onFrameRef = useRef(input.onFrame);
  onFrameRef.current = input.onFrame;
  const slotSignature = useMemo(
    () => input.slots.map((slot) => `${slot.key}:${slot.baud}:${slot.deviceHint}`).join("|"),
    [input.slots],
  );

  useEffect(() => {
    if (!input.enabled) {
      setStates({});
      return;
    }
    let stopped = false;
    const initial: Record<string, LabDeviceListenState> = {};
    for (const slot of input.slots) {
      initial[slot.key] = labDeviceListenState({
        serialSupported,
        hasStoredPort: readStoredPortInfo(slot.key) !== null,
        connected: false,
      });
    }
    setStates(initial);
    const stops = input.slots.map((slot) => startLabDeviceSlotListen({
      slotKey: slot.key,
      baudRate: slot.baud,
      isStopped: () => stopped,
      onState: (state) => {
        if (!stopped) {
          setStates((current) => ({ ...current, [slot.key]: state }));
        }
      },
      onFrame: (bytes) => onFrameRef.current(slot.deviceHint, bytes),
    }));
    return () => {
      stopped = true;
      stops.forEach((stop) => stop());
    };
  }, [input.enabled, input.listenEpoch, input.slots, serialSupported, slotSignature]);

  return states;
}
