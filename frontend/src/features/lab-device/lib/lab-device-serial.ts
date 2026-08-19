import { isWebSerialSupported, labDevicePortStorageKey } from "./lab-device-board-model";

export interface LabDeviceSerialPortInfo {
  usbVendorId?: number;
  usbProductId?: number;
}

interface LabDeviceSerialPort {
  readable: ReadableStream<Uint8Array> | null;
  open: (options: { baudRate: number }) => Promise<void>;
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

async function readIdleFrame(port: LabDeviceSerialPort, baudRate: number): Promise<Uint8Array> {
  await port.open({ baudRate });
  const reader = port.readable?.getReader();
  if (!reader) {
    await port.close();
    throw new Error("serial port is not readable");
  }
  const chunks: Uint8Array[] = [];
  let idle = 0;
  try {
    while (idle < 8) {
      const next = reader.read();
      const timeout = new Promise<{ done: true; value: undefined }>((resolve) => {
        window.setTimeout(() => resolve({ done: true, value: undefined }), 250);
      });
      const { value, done } = await Promise.race([next, timeout]);
      if (value && value.length > 0) {
        chunks.push(value);
        idle = 0;
        continue;
      }
      if (chunks.length > 0) {
        idle += 1;
      }
      if (done && chunks.length > 0) {
        break;
      }
    }
  } finally {
    reader.releaseLock();
    await port.close();
  }
  const total = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
  const out = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks) {
    out.set(chunk, offset);
    offset += chunk.length;
  }
  return out;
}

export async function readLabDeviceSlot(
  slotKey: string,
  baudRate: number,
): Promise<Uint8Array | null> {
  const serial = serialApi();
  const stored = readStoredPortInfo(slotKey);
  if (!serial || !stored) {
    return null;
  }
  const ports = await serial.getPorts();
  const port = ports.find((candidate) => portsMatch(candidate, stored));
  if (!port) {
    return null;
  }
  return readIdleFrame(port, baudRate);
}

export function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}
