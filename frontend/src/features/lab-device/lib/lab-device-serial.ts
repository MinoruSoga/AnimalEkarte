import {
  isWebSerialSupported,
  labDeviceListenState,
  labDevicePortStorageKey,
  type LabDeviceListenState,
} from "./lab-device-board-model";

export const LAB_DEVICE_IDLE_TICKS = 8;
export const LAB_DEVICE_IDLE_MS = 250;
export const LAB_DEVICE_REOPEN_MS = 2000;

export class LabDeviceIdleFrameBuffer {
  private chunks: Uint8Array[] = [];
  private idleTicks = 0;

  push(bytes: Uint8Array): void {
    if (bytes.length === 0) {
      return;
    }
    this.chunks.push(bytes);
    this.idleTicks = 0;
  }

  tickIdle(maxIdle = LAB_DEVICE_IDLE_TICKS): Uint8Array | null {
    if (this.chunks.length === 0) {
      return null;
    }
    this.idleTicks += 1;
    if (this.idleTicks < maxIdle) {
      return null;
    }
    const total = this.chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const out = new Uint8Array(total);
    let offset = 0;
    for (const chunk of this.chunks) {
      out.set(chunk, offset);
      offset += chunk.length;
    }
    this.chunks = [];
    this.idleTicks = 0;
    return out;
  }
}

export interface LabDeviceSerialPortInfo {
  usbVendorId?: number;
  usbProductId?: number;
}

export type LabDeviceSerialParity = "none" | "even" | "odd";

interface LabDeviceSerialPort {
  readable: ReadableStream<Uint8Array> | null;
  open: (options: { baudRate: number; parity?: LabDeviceSerialParity }) => Promise<void>;
  close: () => Promise<void>;
  getInfo: () => LabDeviceSerialPortInfo;
}

interface LabDeviceSerial {
  getPorts: () => Promise<LabDeviceSerialPort[]>;
  requestPort: () => Promise<LabDeviceSerialPort>;
}

function serialApi(): LabDeviceSerial | null {
  if (!isWebSerialSupported()) {
    return null;
  }
  return (navigator as Navigator & { serial: LabDeviceSerial }).serial;
}

export function readStoredPortInfo(slotKey: string): LabDeviceSerialPortInfo | null {
  const raw = localStorage.getItem(labDevicePortStorageKey(slotKey));
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as LabDeviceSerialPortInfo;
  } catch {
    return null;
  }
}

export function storePortInfo(slotKey: string, info: LabDeviceSerialPortInfo): void {
  localStorage.setItem(labDevicePortStorageKey(slotKey), JSON.stringify(info));
}

export async function requestLabDevicePort(slotKey: string): Promise<LabDeviceSerialPortInfo | null> {
  const serial = serialApi();
  if (!serial) {
    return null;
  }
  const port = await serial.requestPort();
  const info = port.getInfo();
  storePortInfo(slotKey, info);
  return info;
}

function portsMatch(port: LabDeviceSerialPort, stored: LabDeviceSerialPortInfo): boolean {
  return port.getInfo().usbVendorId === stored.usbVendorId
    && port.getInfo().usbProductId === stored.usbProductId;
}

async function sleep(ms: number, isStopped: () => boolean): Promise<void> {
  const started = Date.now();
  while (!isStopped() && Date.now() - started < ms) {
    await new Promise((resolve) => {
      window.setTimeout(resolve, Math.min(250, ms));
    });
  }
}

async function readOpenLoop(
  port: LabDeviceSerialPort,
  isStopped: () => boolean,
  onFrame: (bytes: Uint8Array) => Promise<void>,
): Promise<void> {
  const reader = port.readable?.getReader();
  if (!reader) {
    throw new Error("serial port is not readable");
  }
  const buffer = new LabDeviceIdleFrameBuffer();
  try {
    while (!isStopped()) {
      const next = reader.read();
      const timeout = new Promise<{ done: true; value: undefined }>((resolve) => {
        window.setTimeout(() => resolve({ done: true, value: undefined }), LAB_DEVICE_IDLE_MS);
      });
      const { value } = await Promise.race([next, timeout]);
      if (value && value.length > 0) {
        buffer.push(value);
        continue;
      }
      const frame = buffer.tickIdle();
      if (frame && frame.length > 0) {
        await onFrame(frame);
      }
    }
  } finally {
    try {
      reader.releaseLock();
    } catch {
      // already released by cancel
    }
  }
}

export function startLabDeviceSlotListen(input: {
  slotKey: string;
  baudRate: number;
  parity?: LabDeviceSerialParity;
  isStopped: () => boolean;
  onState: (state: LabDeviceListenState) => void;
  onFrame: (bytes: Uint8Array) => Promise<void>;
}): () => void {
  let stopped = false;
  let port: LabDeviceSerialPort | undefined;

  const isStopped = () => stopped || input.isStopped();
  const stop = () => {
    stopped = true;
    void (async () => {
      try {
        await port?.close();
      } catch {
        // already closed
      }
    })();
  };

  void (async () => {
    while (!isStopped()) {
      const serial = serialApi();
      const stored = readStoredPortInfo(input.slotKey);
      const nextState = labDeviceListenState({
        serialSupported: serial !== null,
        hasStoredPort: stored !== null,
        connected: false,
      });
      if (nextState !== "disconnected") {
        input.onState(nextState);
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
        continue;
      }
      if (!serial || !stored) {
        input.onState("disconnected");
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
        continue;
      }
      const ports = await serial.getPorts();
      const matched = ports.find((candidate) => portsMatch(candidate, stored));
      if (!matched) {
        input.onState("disconnected");
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
        continue;
      }
      port = matched;
      try {
        await matched.open(
          input.parity !== undefined
            ? { baudRate: input.baudRate, parity: input.parity }
            : { baudRate: input.baudRate },
        );
        input.onState("listening");
        await readOpenLoop(matched, isStopped, input.onFrame);
      } catch {
        input.onState("disconnected");
      } finally {
        try {
          await matched.close();
        } catch {
          // already closed
        }
        port = undefined;
      }
      if (!isStopped()) {
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
      }
    }
  })();

  return stop;
}

export function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}
