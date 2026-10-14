import { bytesToBase64, readLabDeviceFrames } from "./lab-device-serial";

const MAX_PORTS = 2;
const MAX_PENDING = 100;
const MAX_REJECTED = 100;
const LOCK_NAME = "lab-device-browser-receiver";

export interface LabDeviceBrowserReceiveStatus {
  supported: boolean;
  connected: boolean;
  openPorts: number;
  pending: number;
  rejected: number;
  overflow: number;
  authPaused: boolean;
  error: string | null;
}

interface LabDeviceBrowserSerialPort {
  readable: ReadableStream<Uint8Array> | null;
  open(options: {
    baudRate: number;
    dataBits: 8;
    stopBits: 1;
    parity: "none";
    flowControl: "none";
  }): Promise<void>;
  close(): Promise<void>;
}

interface LabDeviceBrowserSerial {
  requestPort(): Promise<LabDeviceBrowserSerialPort>;
}

interface LabDeviceBrowserLockManager {
  request<T>(
    name: string,
    options: { mode: "exclusive"; ifAvailable: true },
    callback: (lock: { name: string } | null) => Promise<T>,
  ): Promise<T>;
}

export function createLabDeviceBrowserReceiver(input: {
  clinicId: string;
  onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  onStatus: (status: LabDeviceBrowserReceiveStatus) => void;
  isEnabled: () => boolean;
  serial?: LabDeviceBrowserSerial | null;
  locks?: LabDeviceBrowserLockManager | null;
}): { connect(): Promise<void>; retry(): void; stop(): Promise<void> } {
  const serial = input.serial === undefined ? getSerial() : input.serial;
  const locks = input.locks === undefined ? getLocks() : input.locks;
  const ports = new Map<
    LabDeviceBrowserSerialPort,
    { close: () => Promise<void>; controller: AbortController; task?: Promise<void> }
  >();
  const opening = new Set<Promise<void>>();
  const pending: Uint8Array[] = [];
  const rejected: Uint8Array[] = [];
  let stopped = false;
  let stopTask: Promise<void> | undefined;
  let lockHeld = false;
  let releaseLock: (() => void) | undefined;
  let lockRequest: Promise<void> | undefined;
  let pumpActive = false;
  let retryTimer: number | undefined;
  let retryDelay = 1000;
  let overflow = 0;
  let authPaused = false;
  let error: string | null =
    !serial || !isSecurePage()
      ? "Web SerialにはHTTPSのChromeデスクトップ版が必要です"
      : locks === null
        ? "このブラウザでは複数タブ間の排他制御を利用できません"
        : null;
  let connectTask: Promise<void> | undefined;

  const status = () => {
    try {
      input.onStatus({
        supported: serial !== null && locks !== null && isSecurePage(),
        connected: ports.size > 0,
        openPorts: ports.size,
        pending: pending.length,
        rejected: rejected.length,
        overflow,
        authPaused,
        error,
      });
    } catch {
      error = "受信状態を表示できませんでした";
    }
  };
  const enabled = () => !stopped && input.isEnabled();
  const acquireLock = async (): Promise<boolean> => {
    if (lockHeld) return true;
    if (!locks) {
      error = "このブラウザでは複数タブ間の排他制御を利用できません";
      status();
      return false;
    }
    let resolveAcquired: (value: boolean) => void = () => undefined;
    const acquired = new Promise<boolean>((resolve) => {
      resolveAcquired = resolve;
    });
    const hold = new Promise<void>((resolve) => {
      releaseLock = resolve;
    });
    lockRequest = locks
      .request(LOCK_NAME, { mode: "exclusive", ifAvailable: true }, async (lock) => {
        if (!lock) {
          resolveAcquired(false);
          return;
        }
        lockHeld = true;
        resolveAcquired(true);
        await hold;
        lockHeld = false;
      })
      .catch(() => {
        resolveAcquired(false);
        error = "受信の排他制御を開始できません";
        status();
      });
    const result = await acquired;
    if (!result) error ??= "別のタブで検査機器を受信中です";
    status();
    return result;
  };
  const schedulePump = () => {
    if (retryTimer === undefined && !pumpActive && pending.length > 0 && !authPaused && enabled()) {
      void pump();
    }
  };
  const pump = async (): Promise<void> => {
    if (pumpActive) return;
    pumpActive = true;
    try {
      while (pending.length > 0 && enabled() && !authPaused) {
        const frame = pending[0];
        try {
          await input.onFrame({ payloadBase64: bytesToBase64(frame), deviceHint: "auto" });
          if (!enabled()) break;
          if (pending[0] === frame) pending.shift();
          retryDelay = 1000;
          error = null;
          status();
        } catch (failure: unknown) {
          if (!enabled()) break;
          const statusCode = failureStatus(failure);
          if (statusCode === 400) {
            if (rejected.length >= MAX_REJECTED) {
              authPaused = true;
              error = "判定失敗の保留上限です。送信を止めてサポートへ連絡してください";
              for (const entry of ports.values()) entry.controller.abort();
              status();
              break;
            }
            pending.shift();
            rejected.push(frame);
            error = "形式を判定できない検査電文があります。機器から再送してください";
            status();
            continue;
          }
          if (statusCode === 401 || statusCode === 403) {
            authPaused = true;
            error = "ログインまたは検査取込権限を確認し、再試行してください";
            status();
            break;
          }
          error = "通信を再試行しています。受信画面を開いたままにしてください";
          status();
          if (retryTimer === undefined) {
            retryTimer = window.setTimeout(() => {
              retryTimer = undefined;
              void pump();
            }, retryDelay);
          }
          retryDelay = Math.min(retryDelay * 2, 30_000);
          break;
        }
      }
    } finally {
      pumpActive = false;
      status();
    }
  };
  const enqueue = (frame: Uint8Array) => {
    if (!enabled()) return;
    if (
      pending.length >= MAX_PENDING ||
      pending.reduce((sum, entry) => sum + entry.length, 0) + frame.length > MAX_PENDING * 8 * 1024
    ) {
      overflow += 1;
      error = "未送信キュー上限のため次の検査電文は保持されませんでした。機器から再送してください";
      status();
      for (const entry of ports.values()) entry.controller.abort();
      return;
    }
    pending.push(frame.slice());
    status();
    schedulePump();
  };
  const listen = async (port: LabDeviceBrowserSerialPort, controller: AbortController) => {
    const stream = port.readable;
    if (!stream) throw new Error("unavailable");
    const reader = stream.getReader();
    try {
      await readLabDeviceFrames(
        reader,
        controller.signal,
        async (frame) => enqueue(frame),
        () => {
          overflow += 1;
          error = "8KiBを超える検査電文を破棄しました。機器から再送してください";
          status();
        },
      );
    } finally {
      try {
        reader.releaseLock();
      } catch {
        error = "ポートの読み取りを解放できませんでした";
      }
    }
  };
  const openPort = async (port: LabDeviceBrowserSerialPort): Promise<void> => {
    if (ports.has(port) || ports.size >= MAX_PORTS || !enabled()) return;
    const controller = new AbortController();
    let portClosed = false;
    let closeTask: Promise<void> | undefined;
    const entry: { controller: AbortController; close: () => Promise<void>; task?: Promise<void> } =
      {
        controller,
        close: () => {
          if (portClosed) return closeTask ?? Promise.resolve();
          portClosed = true;
          closeTask = port.close();
          return closeTask;
        },
      };
    ports.set(port, entry);
    let openingTask: Promise<void> | undefined;
    try {
      openingTask = port.open({
        baudRate: 9600,
        dataBits: 8,
        stopBits: 1,
        parity: "none",
        flowControl: "none",
      });
      opening.add(openingTask);
      await openingTask;
      opening.delete(openingTask);
      if (!enabled()) {
        await entry.close();
        ports.delete(port);
        status();
        return;
      }
      status();
      entry.task = listen(port, controller)
        .catch(() => {
          if (enabled()) {
            error = "ポートの受信が中断されました。接続を確認してください";
            status();
          }
        })
        .finally(async () => {
          const disconnected = enabled() && !controller.signal.aborted;
          controller.abort();
          if (disconnected) error = "ポートから切断されました。接続を確認して再接続してください";
          try {
            await entry.close();
          } catch {
            if (enabled()) error = "ポートを閉じられませんでした";
          }
          ports.delete(port);
          status();
        });
    } catch {
      if (enabled()) error = "選択したポートを開けませんでした。接続を確認して再試行してください";
      if (openingTask) opening.delete(openingTask);
      controller.abort();
      try {
        await entry.close();
      } catch {
        if (enabled()) error = "ポートを閉じられませんでした";
      }
      ports.delete(port);
      status();
    }
  };
  const connect = (): Promise<void> => {
    if (connectTask) return connectTask;
    if (!enabled()) return Promise.resolve();
    if (!serial || !isSecurePage()) {
      error = "Web SerialにはHTTPSのChromeデスクトップ版が必要です";
      status();
      return Promise.resolve();
    }
    let picker: Promise<LabDeviceBrowserSerialPort>;
    try {
      picker = serial.requestPort();
    } catch {
      error = "ポート選択を開始できません";
      status();
      return Promise.resolve();
    }
    connectTask = (async () => {
      try {
        const port = await picker;
        if (!enabled()) return;
        if (ports.has(port)) return;
        if (ports.size >= MAX_PORTS && !ports.has(port)) {
          error = "接続できるポートは2台までです";
          status();
          return;
        }
        if (!(await acquireLock()) || !enabled()) return;
        error = null;
        await openPort(port);
      } catch (failure: unknown) {
        if (isPickerCancelled(failure)) {
          status();
          return;
        }
        if (enabled()) {
          error = "ポートを選択できませんでした";
          status();
        }
      } finally {
        connectTask = undefined;
      }
    })();
    return connectTask;
  };
  status();
  return {
    connect,
    retry: () => {
      if (!enabled()) return;
      authPaused = false;
      error = null;
      if (retryTimer !== undefined) window.clearTimeout(retryTimer);
      retryTimer = undefined;
      status();
      schedulePump();
    },
    stop: () => {
      if (stopTask) return stopTask;
      stopped = true;
      stopTask = (async () => {
        if (retryTimer !== undefined) window.clearTimeout(retryTimer);
        retryTimer = undefined;
        for (const entry of ports.values()) entry.controller.abort();
        await Promise.all([...opening].map((task) => task.catch(() => undefined)));
        await Promise.all(
          [...ports.values()].map(async (entry) => {
            await entry.task?.catch(() => undefined);
            try {
              await entry.close();
            } catch {
              error = "ポートを閉じられませんでした";
            }
          }),
        );
        releaseLock?.();
        await lockRequest?.catch(() => undefined);
        ports.clear();
        pending.length = 0;
        status();
      })();
      return stopTask;
    },
  };
}

function getSerial(): LabDeviceBrowserSerial | null {
  if (typeof navigator === "undefined" || !("serial" in navigator)) return null;
  return (navigator as Navigator & { serial: LabDeviceBrowserSerial | undefined }).serial ?? null;
}

function getLocks(): LabDeviceBrowserLockManager | null {
  if (typeof navigator === "undefined" || !("locks" in navigator)) return null;
  return (
    (navigator as Navigator & { locks: LabDeviceBrowserLockManager | undefined }).locks ?? null
  );
}

function isSecurePage(): boolean {
  return typeof window !== "undefined" && window.isSecureContext;
}
function isPickerCancelled(error: unknown): boolean {
  return (
    typeof error === "object" && error !== null && "name" in error && error.name === "NotFoundError"
  );
}
function failureStatus(error: unknown): number | undefined {
  return typeof error === "object" &&
    error !== null &&
    "status" in error &&
    typeof error.status === "number"
    ? error.status
    : undefined;
}
