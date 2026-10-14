import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createLabDeviceBrowserReceiver } from "./lab-device-browser-receiver";

function serialPort() {
  let controller: ReadableStreamDefaultController<Uint8Array> | undefined;
  const port = {
    readable: new ReadableStream<Uint8Array>({
      start(value) {
        controller = value;
      },
    }),
    open: vi.fn().mockResolvedValue(undefined),
    close: vi.fn().mockResolvedValue(undefined),
  };
  return {
    port,
    send: (bytes: number[]) => controller?.enqueue(new Uint8Array(bytes)),
    disconnect: () => controller?.close(),
  };
}

describe("createLabDeviceBrowserReceiver", () => {
  beforeEach(() =>
    Object.defineProperty(window, "isSecureContext", { configurable: true, value: true }),
  );
  afterEach(() => vi.useRealTimers());

  it("explicitly opens and independently receives identical-converter ports once each", async () => {
    vi.useFakeTimers();
    const left = serialPort();
    const right = serialPort();
    const serial = {
      requestPort: vi
        .fn()
        .mockResolvedValueOnce(left.port)
        .mockResolvedValueOnce(left.port)
        .mockResolvedValueOnce(right.port),
    };
    const onFrame = vi.fn().mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial,
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    const first = receiver.connect();
    await vi.waitFor(() => expect(left.port.open).toHaveBeenCalledTimes(1));
    await receiver.connect();
    const second = receiver.connect();
    await vi.waitFor(() => expect(right.port.open).toHaveBeenCalledTimes(1));
    left.send([0, 1, 255]);
    right.send([4, 5, 128]);
    await vi.advanceTimersByTimeAsync(2000);
    await vi.waitFor(() => expect(onFrame).toHaveBeenCalledTimes(2));
    expect(left.port.open).toHaveBeenCalledWith({
      baudRate: 9600,
      dataBits: 8,
      stopBits: 1,
      parity: "none",
      flowControl: "none",
    });
    expect(onFrame).toHaveBeenCalledWith({ payloadBase64: "AAH/", deviceHint: "auto" });
    expect(onFrame).toHaveBeenCalledWith({ payloadBase64: "BAWA", deviceHint: "auto" });
    await receiver.stop();
    await Promise.all([first, second]);
  });

  it("does not open a third port and reports cancelled picker without a fatal error", async () => {
    const one = serialPort();
    const two = serialPort();
    const serial = {
      requestPort: vi
        .fn()
        .mockResolvedValueOnce(one.port)
        .mockResolvedValueOnce(two.port)
        .mockRejectedValueOnce(new DOMException("cancel", "NotFoundError")),
    };
    const onStatus = vi.fn();
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus,
      isEnabled: () => true,
      serial,
      locks: { request: (_n, _o, cb) => cb({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    await receiver.connect();
    expect(one.port.open).toHaveBeenCalledTimes(1);
    expect(two.port.open).toHaveBeenCalledTimes(1);
    expect(onStatus.mock.calls.at(-1)?.[0].error).toBeNull();
    await receiver.stop();
  });

  it("rejects a third distinct SerialPort object for one receiver", async () => {
    const first = serialPort();
    const second = serialPort();
    const third = serialPort();
    const statuses: Array<{ error: string | null }> = [];
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: {
        requestPort: vi
          .fn()
          .mockResolvedValueOnce(first.port)
          .mockResolvedValueOnce(second.port)
          .mockResolvedValueOnce(third.port),
      },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    await receiver.connect();
    expect(first.port.open).toHaveBeenCalledTimes(1);
    expect(second.port.open).toHaveBeenCalledTimes(1);
    expect(third.port.open).not.toHaveBeenCalled();
    expect(statuses.at(-1)?.error).toContain("2台まで");
    await receiver.stop();
  });

  it("fails closed without an available origin lock", async () => {
    const item = serialPort();
    const onStatus = vi.fn();
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus,
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_n, _o, cb) => cb(null) },
    });
    await receiver.connect();
    expect(item.port.open).not.toHaveBeenCalled();
    expect(onStatus.mock.calls.at(-1)?.[0].error).toContain("別のタブ");
  });

  it("shows unsupported browser and HTTPS states without opening a port", async () => {
    const port = serialPort();
    const status = vi.fn();
    const unsupported = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: status,
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(port.port) },
      locks: null,
    });
    expect(status.mock.calls.at(-1)?.[0].supported).toBe(false);
    await unsupported.connect();
    expect(port.port.open).not.toHaveBeenCalled();
    await unsupported.stop();

    Object.defineProperty(window, "isSecureContext", { configurable: true, value: false });
    const insecure = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: status,
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(port.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await insecure.connect();
    expect(port.port.open).not.toHaveBeenCalled();
    await insecure.stop();
  });

  it("retains a retryable frame and delivers it once after backoff recovery", async () => {
    vi.useFakeTimers();
    const item = serialPort();
    const failure = Object.assign(new Error("offline"), { status: 503 });
    const onFrame = vi.fn().mockRejectedValueOnce(failure).mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send([9]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(onFrame).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1000);
    await vi.waitFor(() => expect(onFrame).toHaveBeenCalledTimes(2));
    await receiver.stop();
  });

  it.each([401, 403])("pauses after %s until explicit retry", async (statusCode) => {
    vi.useFakeTimers();
    const item = serialPort();
    const statuses: Array<{ authPaused: boolean }> = [];
    const onFrame = vi
      .fn()
      .mockRejectedValueOnce(Object.assign(new Error("unauthorized"), { status: statusCode }))
      .mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send([9]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(statuses.at(-1)?.authPaused).toBe(true);
    expect(onFrame).toHaveBeenCalledTimes(1);
    receiver.retry();
    await vi.waitFor(() => expect(onFrame).toHaveBeenCalledTimes(2));
    await receiver.stop();
  });

  it("moves a 400 frame into the bounded rejected count", async () => {
    vi.useFakeTimers();
    const item = serialPort();
    const statuses: Array<{ rejected: number; pending: number }> = [];
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn().mockRejectedValue(Object.assign(new Error("invalid"), { status: 400 })),
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send([9]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(statuses.at(-1)).toMatchObject({ rejected: 1, pending: 0 });
    await receiver.stop();
  });

  it("does not open a selected port when a pending picker resolves after stop", async () => {
    let resolvePicker!: (port: ReturnType<typeof serialPort>["port"]) => void;
    const picker = new Promise<ReturnType<typeof serialPort>["port"]>((resolve) => {
      resolvePicker = resolve;
    });
    const item = serialPort();
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: () => picker },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    const connecting = receiver.connect();
    await receiver.stop();
    resolvePicker(item.port);
    await connecting;
    expect(item.port.open).not.toHaveBeenCalled();
  });

  it("waits for a pending open to finish, then closes without starting a reader", async () => {
    let resolveOpen!: () => void;
    const opened = new Promise<void>((resolve) => {
      resolveOpen = resolve;
    });
    const reader = {
      read: vi.fn(() => new Promise<ReadableStreamReadResult<Uint8Array>>(() => {})),
      cancel: vi.fn(async () => {}),
      releaseLock: vi.fn(),
    };
    const port = {
      readable: { getReader: vi.fn(() => reader) } as unknown as ReadableStream<Uint8Array>,
      open: vi.fn(() => opened),
      close: vi.fn(async () => {}),
    };
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    const connecting = receiver.connect();
    await vi.waitFor(() => expect(port.open).toHaveBeenCalledTimes(1));
    const stopping = receiver.stop();
    resolveOpen();
    await Promise.all([connecting, stopping]);
    expect(port.close).toHaveBeenCalledTimes(1);
    expect(port.readable.getReader).not.toHaveBeenCalled();
  });

  it("cancels the active reader and releases it before closing the port", async () => {
    const events: string[] = [];
    let finishRead!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    const reader = {
      read: vi.fn(
        () =>
          new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
            finishRead = resolve;
          }),
      ),
      cancel: vi.fn(async () => {
        events.push("cancel");
        finishRead({ done: true, value: undefined });
      }),
      releaseLock: vi.fn(() => {
        events.push("release");
      }),
    };
    const port = {
      readable: { getReader: vi.fn(() => reader) } as unknown as ReadableStream<Uint8Array>,
      open: vi.fn(async () => {
        events.push("open");
      }),
      close: vi.fn(async () => {
        events.push("close");
      }),
    };
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await vi.waitFor(() => expect(reader.read).toHaveBeenCalledTimes(1));
    await receiver.stop();
    expect(events).toEqual(["open", "cancel", "release", "close"]);
    expect(port.close).toHaveBeenCalledTimes(1);
  });

  it("pauses readers visibly when the 100-frame pending bound is reached", async () => {
    vi.useFakeTimers();
    const item = serialPort();
    const statuses: Array<{ error: string | null; pending: number; overflow: number }> = [];
    const resolvers: Array<() => void> = [];
    const onFrame = vi.fn(() => new Promise<void>((resolve) => resolvers.push(resolve)));
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    for (let index = 0; index < 101; index += 1) {
      item.send([index % 255]);
      await vi.advanceTimersByTimeAsync(2000);
    }
    expect(statuses.at(-1)).toMatchObject({ pending: 100, overflow: 1 });
    expect(statuses.at(-1)?.error).toContain("再送");
    expect(onFrame).toHaveBeenCalledTimes(1);
    for (let index = 0; index < 100; index += 1) {
      resolvers[index]?.();
      await vi.advanceTimersByTimeAsync(0);
    }
    expect(onFrame.mock.calls.map(([frame]) => frame.payloadBase64)).toEqual(
      Array.from({ length: 100 }, (_, index) => btoa(String.fromCharCode(index))),
    );
    await receiver.stop();
  });

  it("honors one backoff timer when another port receives during retry delay", async () => {
    vi.useFakeTimers();
    const first = serialPort();
    const second = serialPort();
    const calls: Array<{ at: number; payload: string }> = [];
    const startedAt = Date.now();
    const onFrame = vi
      .fn()
      .mockImplementation(async ({ payloadBase64 }: { payloadBase64: string }) => {
        calls.push({ at: Date.now() - startedAt, payload: payloadBase64 });
        if (calls.length === 1) throw Object.assign(new Error("offline"), { status: 503 });
      });
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: {
        requestPort: vi.fn().mockResolvedValueOnce(first.port).mockResolvedValueOnce(second.port),
      },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    first.send([1]);
    await vi.advanceTimersByTimeAsync(500);
    second.send([2]);
    await vi.advanceTimersByTimeAsync(1500);
    expect(onFrame).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(500);
    expect(onFrame).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(500);
    expect(onFrame).toHaveBeenCalledTimes(3);
    expect(calls).toEqual([
      { at: 2000, payload: "AQ==" },
      { at: 3000, payload: "AQ==" },
      { at: 3000, payload: "Ag==" },
    ]);
    await receiver.stop();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("cancels one scheduled retry on explicit retry and clears timers on stop", async () => {
    vi.useFakeTimers();
    const item = serialPort();
    const onFrame = vi
      .fn()
      .mockRejectedValueOnce(Object.assign(new Error("offline"), { status: 503 }))
      .mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send([1]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(onFrame).toHaveBeenCalledTimes(1);
    item.send([2]);
    await vi.advanceTimersByTimeAsync(500);
    receiver.retry();
    await vi.waitFor(() => expect(onFrame).toHaveBeenCalledTimes(2));
    await vi.advanceTimersByTimeAsync(1500);
    expect(onFrame).toHaveBeenCalledTimes(3);
    await receiver.stop();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("keeps further frames behind exponential backoff after repeated failures", async () => {
    vi.useFakeTimers();
    const first = serialPort();
    const second = serialPort();
    const onFrame = vi
      .fn()
      .mockRejectedValueOnce(Object.assign(new Error("offline"), { status: 503 }))
      .mockRejectedValueOnce(Object.assign(new Error("offline"), { status: 503 }))
      .mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: {
        requestPort: vi.fn().mockResolvedValueOnce(first.port).mockResolvedValueOnce(second.port),
      },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    first.send([1]);
    await vi.advanceTimersByTimeAsync(500);
    second.send([2]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(onFrame).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(500);
    expect(onFrame).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(1500);
    expect(onFrame).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(500);
    expect(onFrame).toHaveBeenCalledTimes(4);
    expect(onFrame.mock.calls.map(([frame]) => frame.payloadBase64)).toEqual([
      "AQ==",
      "AQ==",
      "AQ==",
      "Ag==",
    ]);
    await receiver.stop();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("does not schedule retry or revive state when an in-flight request rejects after stop", async () => {
    vi.useFakeTimers();
    let rejectFrame!: (failure: unknown) => void;
    const onFrame = vi.fn(() => new Promise<void>((_resolve, reject) => (rejectFrame = reject)));
    const item = serialPort();
    const statuses: Array<{ pending: number; rejected: number; authPaused: boolean }> = [];
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send([1]);
    await vi.advanceTimersByTimeAsync(2000);
    const stopping = receiver.stop();
    await stopping;
    rejectFrame(Object.assign(new Error("late"), { status: 503 }));
    await vi.advanceTimersByTimeAsync(0);
    expect(onFrame).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
    await vi.advanceTimersByTimeAsync(30_000);
    expect(onFrame).toHaveBeenCalledTimes(1);
    expect(statuses.at(-1)).toMatchObject({ pending: 0, rejected: 0, authPaused: false });
  });

  it.each([400, 401, 403])(
    "does not revive queues or auth state after stop for late %s",
    async (statusCode) => {
      vi.useFakeTimers();
      let rejectFrame!: (failure: unknown) => void;
      const onFrame = vi.fn(() => new Promise<void>((_resolve, reject) => (rejectFrame = reject)));
      const item = serialPort();
      const statuses: Array<{ pending: number; rejected: number; authPaused: boolean }> = [];
      const receiver = createLabDeviceBrowserReceiver({
        clinicId: "clinic-a",
        onFrame,
        onStatus: (status) => statuses.push(status),
        isEnabled: () => true,
        serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
        locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
      });
      await receiver.connect();
      item.send([1]);
      await vi.advanceTimersByTimeAsync(2000);
      await receiver.stop();
      rejectFrame(Object.assign(new Error("late"), { status: statusCode }));
      await vi.advanceTimersByTimeAsync(0);
      expect(vi.getTimerCount()).toBe(0);
      expect(statuses.at(-1)).toMatchObject({ pending: 0, rejected: 0, authPaused: false });
    },
  );

  it("releases a port slot after EOF and keeps the other selected port receiving", async () => {
    vi.useFakeTimers();
    const first = serialPort();
    const second = serialPort();
    const replacement = serialPort();
    const onFrame = vi.fn().mockResolvedValue(undefined);
    const statuses: Array<{ error: string | null }> = [];
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: {
        requestPort: vi
          .fn()
          .mockResolvedValueOnce(first.port)
          .mockResolvedValueOnce(second.port)
          .mockResolvedValueOnce(replacement.port),
      },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    first.disconnect();
    await vi.waitFor(() => expect(first.port.close).toHaveBeenCalledTimes(1));
    await vi.waitFor(() => expect(statuses.at(-1)?.error).toContain("切断"));
    await receiver.connect();
    expect(replacement.port.open).toHaveBeenCalledTimes(1);
    second.send([7]);
    await vi.advanceTimersByTimeAsync(2000);
    await vi.waitFor(() =>
      expect(onFrame).toHaveBeenCalledWith({ payloadBase64: "Bw==", deviceHint: "auto" }),
    );
    await receiver.stop();
  });

  it("does not retain a failed opening port identity or consume an active slot", async () => {
    const failed = serialPort();
    failed.port.open.mockRejectedValueOnce(new Error("disconnected"));
    const first = serialPort();
    const second = serialPort();
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: {
        requestPort: vi
          .fn()
          .mockResolvedValueOnce(failed.port)
          .mockResolvedValueOnce(first.port)
          .mockResolvedValueOnce(second.port),
      },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    await receiver.connect();
    await receiver.connect();
    expect(failed.port.open).toHaveBeenCalledTimes(1);
    expect(first.port.open).toHaveBeenCalledTimes(1);
    expect(second.port.open).toHaveBeenCalledTimes(1);
    await receiver.stop();
  });

  it("shares the origin lock across receivers until the first receiver stops", async () => {
    let owned = false;
    const locks = {
      request: vi.fn(
        (_name, _options, callback: (lock: { name: string } | null) => Promise<void>) => {
          if (owned) return callback(null);
          owned = true;
          return callback({ name: "lock" }).finally(() => {
            owned = false;
          });
        },
      ),
    };
    const firstPort = serialPort();
    const secondPort = serialPort();
    const first = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(firstPort.port) },
      locks,
    });
    const second = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: vi.fn(),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(secondPort.port) },
      locks,
    });
    await first.connect();
    await second.connect();
    expect(firstPort.port.open).toHaveBeenCalledTimes(1);
    expect(secondPort.port.open).not.toHaveBeenCalled();
    await first.stop();
    await second.connect();
    expect(secondPort.port.open).toHaveBeenCalledTimes(1);
    await second.stop();
  });

  it("fails closed when the browser rejects the lock request", async () => {
    const item = serialPort();
    const statuses: Array<{ error: string | null }> = [];
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame: vi.fn(),
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: vi.fn().mockRejectedValue(new Error("lock denied")) },
    });
    await receiver.connect();
    expect(item.port.open).not.toHaveBeenCalled();
    expect(statuses.at(-1)?.error).toContain("排他制御");
    await receiver.stop();
  });

  it("reports frame overflow over 8KiB without sending an HTTP frame", async () => {
    vi.useFakeTimers();
    const item = serialPort();
    const statuses: Array<{ overflow: number }> = [];
    const onFrame = vi.fn().mockResolvedValue(undefined);
    const receiver = createLabDeviceBrowserReceiver({
      clinicId: "clinic-a",
      onFrame,
      onStatus: (status) => statuses.push(status),
      isEnabled: () => true,
      serial: { requestPort: vi.fn().mockResolvedValue(item.port) },
      locks: { request: (_name, _options, callback) => callback({ name: "lock" }) },
    });
    await receiver.connect();
    item.send(Array.from({ length: 8193 }, () => 65));
    await vi.advanceTimersByTimeAsync(2000);
    expect(statuses.at(-1)?.overflow).toBe(1);
    expect(onFrame).not.toHaveBeenCalled();
    await receiver.stop();
  });
});
