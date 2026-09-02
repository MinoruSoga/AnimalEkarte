import {
  isWebSerialSupported,
  labDeviceListenState,
  labDevicePortStorageKey,
  type LabDeviceListenState,
} from "./lab-device-board-model";

export const LAB_DEVICE_IDLE_TICKS = 8;
export const LAB_DEVICE_IDLE_MS = 250;
const LAB_DEVICE_REOPEN_MS = 2000;

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
    return this.take();
  }

  take(): Uint8Array | null {
    if (this.chunks.length === 0) {
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

export interface LabDeviceSerialReader {
  read: () => Promise<ReadableStreamReadResult<Uint8Array>>;
  cancel: () => Promise<void>;
}

function serialApi(): LabDeviceSerial | null {
  if (!isWebSerialSupported()) {
    return null;
  }
  return (navigator as Navigator & { serial: LabDeviceSerial }).serial;
}

// FE-RC-037: localStorage 由来の JSON は無検証で信用しない。形が合わなければ null（≒未保存扱い）。
function isLabDeviceSerialPortInfo(value: unknown): value is LabDeviceSerialPortInfo {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  const vendorOk = candidate.usbVendorId === undefined || typeof candidate.usbVendorId === "number";
  const productOk = candidate.usbProductId === undefined || typeof candidate.usbProductId === "number";
  return vendorOk && productOk;
}

export function readStoredPortInfo(slotKey: string): LabDeviceSerialPortInfo | null {
  let raw: string | null;
  try {
    raw = localStorage.getItem(labDevicePortStorageKey(slotKey));
  } catch {
    return null;
  }
  if (!raw) {
    return null;
  }
  try {
    const parsed: unknown = JSON.parse(raw);
    return isLabDeviceSerialPortInfo(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

export function storePortInfo(slotKey: string, info: LabDeviceSerialPortInfo): void {
  try {
    localStorage.setItem(labDevicePortStorageKey(slotKey), JSON.stringify(info));
  } catch {
    // Private browsing / quota exceeded: 次回接続時に再度権限ダイアログが出るのみで致命的ではない。
  }
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

export async function readLabDeviceFrames(
  reader: LabDeviceSerialReader,
  signal: AbortSignal,
  onFrame: (bytes: Uint8Array) => Promise<void>,
): Promise<void> {
  const buffer = new LabDeviceIdleFrameBuffer();
  let delivery = Promise.resolve();
  let deliveryError: unknown;
  let readError: unknown;
  let cancelPromise: Promise<void> | undefined;
  let idleTimer: number | undefined;
  const reading = new AbortController();
  const cancelRead = (): Promise<void> => {
    cancelPromise ??= reader.cancel();
    return cancelPromise;
  };
  const abortReading = () => reading.abort();
  if (signal.aborted) {
    reading.abort();
  } else {
    signal.addEventListener("abort", abortReading, { once: true });
  }
  const deliverBuffered = () => {
    const frame = buffer.take();
    if (!frame || frame.length === 0) {
      return;
    }
    delivery = delivery
      .then(() => onFrame(frame))
      .catch((error: unknown) => {
        deliveryError = error;
        reading.abort();
      });
  };
  const scheduleFrame = () => {
    if (idleTimer !== undefined) {
      window.clearTimeout(idleTimer);
    }
    idleTimer = window.setTimeout(() => {
      idleTimer = undefined;
      deliverBuffered();
    }, LAB_DEVICE_IDLE_MS * LAB_DEVICE_IDLE_TICKS);
  };
  try {
    while (!reading.signal.aborted) {
      const { done, value } = await readLabDeviceChunk(reader, reading.signal, cancelRead);
      if (done) {
        break;
      }
      if (value && value.length > 0) {
        buffer.push(value);
        scheduleFrame();
      }
    }
  } catch (error: unknown) {
    readError = error;
  } finally {
    if (idleTimer !== undefined) {
      window.clearTimeout(idleTimer);
    }
    if (!reading.signal.aborted) {
      deliverBuffered();
    }
    signal.removeEventListener("abort", abortReading);
    await delivery;
    if (cancelPromise !== undefined) {
      try {
        await cancelPromise;
      } catch (error: unknown) {
        readError ??= error;
      }
    }
  }
  if (deliveryError !== undefined) {
    throw deliveryError;
  }
  if (readError !== undefined) {
    throw readError;
  }
}

function readLabDeviceChunk(
  reader: LabDeviceSerialReader,
  signal: AbortSignal,
  cancelRead: () => Promise<void>,
): Promise<ReadableStreamReadResult<Uint8Array>> {
  if (signal.aborted) {
    return cancelRead().then(() => ({ done: true, value: undefined }));
  }
  return new Promise((resolve, reject) => {
    let settled = false;
    const finish = (result: ReadableStreamReadResult<Uint8Array>) => {
      if (settled) {
        return;
      }
      settled = true;
      signal.removeEventListener("abort", onAbort);
      resolve(result);
    };
    const fail = (error: unknown) => {
      if (settled) {
        return;
      }
      settled = true;
      signal.removeEventListener("abort", onAbort);
      reject(error);
    };
    const onAbort = () => {
      void cancelRead()
        .then(() => finish({ done: true, value: undefined }))
        .catch(fail);
    };
    signal.addEventListener("abort", onAbort, { once: true });
    void reader.read().then(finish).catch(fail);
  });
}

async function readOpenLoop(
  port: LabDeviceSerialPort,
  signal: AbortSignal,
  onFrame: (bytes: Uint8Array) => Promise<void>,
): Promise<void> {
  const reader = port.readable?.getReader();
  if (!reader) {
    throw new Error("serial port is not readable");
  }
  try {
    await readLabDeviceFrames(reader, signal, onFrame);
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
  /** FE-RC-035: open() 失敗・読み取り中断時の原因情報を破棄せず呼び出し側へ渡す（診断用途）。 */
  onError?: (error: unknown) => void;
}): () => void {
  let stopped = false;
  let activeRead: AbortController | undefined;

  const isStopped = () => stopped || input.isStopped();
  const notifyState = (state: LabDeviceListenState) => {
    try {
      input.onState(state);
    } catch {
      // The React consumer may already be unavailable during teardown.
    }
  };
  const stop = () => {
    stopped = true;
    activeRead?.abort();
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
        notifyState(nextState);
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
        continue;
      }
      if (!serial || !stored) {
        notifyState("disconnected");
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
        continue;
      }
      let matched: LabDeviceSerialPort | undefined;
      try {
        const ports = await serial.getPorts();
        matched = ports.find((candidate) => portsMatch(candidate, stored));
        if (!matched) {
          notifyState("disconnected");
          await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
          continue;
        }
        await matched.open(
          input.parity !== undefined
            ? { baudRate: input.baudRate, parity: input.parity }
            : { baudRate: input.baudRate },
        );
        activeRead = new AbortController();
        if (isStopped()) {
          activeRead.abort();
        }
        notifyState("listening");
        await readOpenLoop(matched, activeRead.signal, input.onFrame);
      } catch (error: unknown) {
        // FE-RC-035: open() 失敗と読み取り中断を区別できるよう、原因情報を破棄せず onError へ渡す。
        notifyState("disconnected");
        try {
          input.onError?.(error);
        } catch {
          // The React consumer may already be unavailable during teardown.
        }
      } finally {
        activeRead?.abort();
        activeRead = undefined;
        try {
          await matched?.close();
        } catch {
          // already closed
        }
      }
      if (!isStopped()) {
        await sleep(LAB_DEVICE_REOPEN_MS, isStopped);
      }
    }
  })().catch(() => {
    notifyState("disconnected");
  });

  return stop;
}

export function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}
