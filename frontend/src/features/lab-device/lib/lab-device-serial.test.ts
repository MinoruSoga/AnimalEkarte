import { afterEach, describe, expect, it, vi } from "vitest";

import {
  LAB_DEVICE_IDLE_MS,
  LAB_DEVICE_IDLE_TICKS,
  LabDeviceIdleFrameBuffer,
  bytesToBase64,
  readLabDeviceFrames,
  readStoredPortInfo,
  requestLabDevicePort,
  startLabDeviceSlotListen,
  storePortInfo,
  type LabDeviceSerialReader,
} from "./lab-device-serial";

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  localStorage.clear();
  Reflect.deleteProperty(navigator, "serial");
});

describe("LabDeviceIdleFrameBuffer", () => {
  it("ignores empty input and has nothing to take", () => {
    const buffer = new LabDeviceIdleFrameBuffer();
    buffer.push(new Uint8Array());
    expect(buffer.take()).toBeNull();
  });

  it("does not emit until idle ticks accumulate after bytes", () => {
    const buffer = new LabDeviceIdleFrameBuffer();
    expect(buffer.tickIdle()).toBeNull();
    buffer.push(new Uint8Array([0x02, 0x41]));
    for (let i = 0; i < LAB_DEVICE_IDLE_TICKS - 1; i += 1) {
      expect(buffer.tickIdle()).toBeNull();
    }
    expect(Array.from(buffer.tickIdle() ?? [])).toEqual([0x02, 0x41]);
  });

  it("resets idle when more bytes arrive and concatenates the frame", () => {
    const buffer = new LabDeviceIdleFrameBuffer();
    buffer.push(new Uint8Array([0x02]));
    expect(buffer.tickIdle()).toBeNull();
    buffer.push(new Uint8Array([0x41, 0x03]));
    for (let i = 0; i < LAB_DEVICE_IDLE_TICKS - 1; i += 1) {
      expect(buffer.tickIdle()).toBeNull();
    }
    expect(Array.from(buffer.tickIdle() ?? [])).toEqual([0x02, 0x41, 0x03]);
    expect(buffer.tickIdle()).toBeNull();
  });
});

describe("lab device port setup", () => {
  it("returns null when Web Serial is unavailable", async () => {
    await expect(requestLabDevicePort("nx600")).resolves.toBeNull();
  });

  it("stores the selected port information", async () => {
    const info = { usbVendorId: 10, usbProductId: 20 };
    Object.defineProperty(navigator, "serial", {
      configurable: true,
      value: {
        getPorts: vi.fn(),
        requestPort: vi.fn(async () => ({ getInfo: () => info })),
      },
    });

    await expect(requestLabDevicePort("nx600")).resolves.toEqual(info);
    expect(readStoredPortInfo("nx600")).toEqual(info);
  });

  it("ignores unavailable or invalid stored port information", () => {
    expect(readStoredPortInfo("missing")).toBeNull();
    localStorage.setItem("lab-device-port:nx600", "not-json");
    expect(readStoredPortInfo("nx600")).toBeNull();
    vi.spyOn(Storage.prototype, "getItem").mockImplementationOnce(() => {
      throw new DOMException("blocked", "SecurityError");
    });
    expect(readStoredPortInfo("nx600")).toBeNull();
  });

  it("encodes raw bytes without text normalization", () => {
    expect(bytesToBase64(new Uint8Array([0x02, 0xff, 0x03]))).toBe("Av8D");
  });
});

describe("readLabDeviceFrames", () => {
  it("delivers buffered bytes when the serial stream ends before the idle timer", async () => {
    const reader: LabDeviceSerialReader = {
      read: vi
        .fn()
        .mockResolvedValueOnce({ done: false, value: new Uint8Array([0x02, 0x41, 0x03]) })
        .mockResolvedValueOnce({ done: true, value: undefined }),
      cancel: vi.fn(async () => {}),
    };
    const onFrame = vi.fn(async () => {});

    await readLabDeviceFrames(reader, new AbortController().signal, onFrame);

    expect(onFrame).toHaveBeenCalledWith(new Uint8Array([0x02, 0x41, 0x03]));
  });

  it("keeps one pending read and receives bytes sent after idle timeouts", async () => {
    vi.useFakeTimers();
    const controller = new AbortController();
    let resolveFirst!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    let resolveSecond!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    const firstRead = new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
      resolveFirst = resolve;
    });
    const secondRead = new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
      resolveSecond = resolve;
    });
    const reader: LabDeviceSerialReader = {
      read: vi.fn().mockReturnValueOnce(firstRead).mockReturnValueOnce(secondRead),
      cancel: vi.fn(async () => {
        resolveSecond({ done: true, value: undefined });
      }),
    };
    const onFrame = vi.fn(async (_bytes: Uint8Array): Promise<void> => {});
    const running = readLabDeviceFrames(reader, controller.signal, onFrame);

    await vi.advanceTimersByTimeAsync(LAB_DEVICE_IDLE_MS * 3);
    expect(reader.read).toHaveBeenCalledTimes(1);

    const raw = new Uint8Array([0x02, 0x41, 0x03]);
    resolveFirst({ done: false, value: raw });
    await vi.advanceTimersByTimeAsync(0);
    await vi.advanceTimersByTimeAsync(LAB_DEVICE_IDLE_MS * LAB_DEVICE_IDLE_TICKS - 1);

    expect(onFrame).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1);

    expect(onFrame).toHaveBeenCalledTimes(1);
    expect(onFrame).toHaveBeenCalledWith(raw);

    controller.abort();
    await running;
    expect(reader.cancel).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("cancels the pending read and reports frame delivery errors", async () => {
    vi.useFakeTimers();
    let resolvePending!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    const pendingRead = new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
      resolvePending = resolve;
    });
    const reader: LabDeviceSerialReader = {
      read: vi
        .fn()
        .mockResolvedValueOnce({ done: false, value: new Uint8Array([0x02, 0x03]) })
        .mockReturnValueOnce(pendingRead),
      cancel: vi.fn(async () => {
        resolvePending({ done: true, value: undefined });
      }),
    };
    const deliveryError = new Error("frame delivery failed");
    const running = readLabDeviceFrames(reader, new AbortController().signal, async () => {
      throw deliveryError;
    });
    const rejected = expect(running).rejects.toBe(deliveryError);

    await vi.advanceTimersByTimeAsync(LAB_DEVICE_IDLE_MS * LAB_DEVICE_IDLE_TICKS);

    await rejected;
    expect(reader.cancel).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("reports cancellation failures without leaving timers running", async () => {
    vi.useFakeTimers();
    const controller = new AbortController();
    const cancelError = new Error("cancel failed");
    const reader: LabDeviceSerialReader = {
      read: vi.fn(() => new Promise<ReadableStreamReadResult<Uint8Array>>(() => {})),
      cancel: vi.fn().mockRejectedValue(cancelError),
    };
    const running = readLabDeviceFrames(reader, controller.signal, async () => {});
    const rejected = expect(running).rejects.toBe(cancelError);

    controller.abort();

    await rejected;
    expect(reader.cancel).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("waits for cancellation to finish before completing", async () => {
    const controller = new AbortController();
    let resolveRead!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    let resolveCancel!: () => void;
    const reader: LabDeviceSerialReader = {
      read: vi.fn(
        () =>
          new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
            resolveRead = resolve;
          }),
      ),
      cancel: vi.fn(
        () =>
          new Promise<void>((resolve) => {
            resolveCancel = resolve;
          }),
      ),
    };
    let completed = false;
    const running = readLabDeviceFrames(reader, controller.signal, async () => {});
    void running.then(() => {
      completed = true;
    });

    controller.abort();
    resolveRead({ done: true, value: undefined });
    await Promise.resolve();
    await Promise.resolve();

    expect(completed).toBe(false);

    resolveCancel();
    await running;
    expect(completed).toBe(true);
  });
});

describe("startLabDeviceSlotListen", () => {
  it("finishes cancellation before releasing the reader and closing the port", async () => {
    const events: string[] = [];
    let resolveRead!: (result: ReadableStreamReadResult<Uint8Array>) => void;
    const reader = {
      read: vi.fn(
        () =>
          new Promise<ReadableStreamReadResult<Uint8Array>>((resolve) => {
            resolveRead = resolve;
          }),
      ),
      cancel: vi.fn(async () => {
        events.push("cancel-start");
        resolveRead({ done: true, value: undefined });
        await Promise.resolve();
        events.push("cancel-finish");
      }),
      releaseLock: vi.fn(() => {
        events.push("release");
      }),
    };
    const port = {
      readable: { getReader: () => reader },
      open: vi.fn(async () => {
        events.push("open");
      }),
      close: vi.fn(async () => {
        events.push("close");
      }),
      getInfo: () => ({ usbVendorId: 1, usbProductId: 2 }),
    };
    Object.defineProperty(navigator, "serial", {
      configurable: true,
      value: {
        getPorts: vi.fn(async () => [port]),
        requestPort: vi.fn(),
      },
    });
    storePortInfo("nx600", port.getInfo());
    const states: string[] = [];
    const stop = startLabDeviceSlotListen({
      slotKey: "nx600",
      baudRate: 9600,
      isStopped: () => false,
      onState: (state) => {
        states.push(state);
      },
      onFrame: async () => {},
    });
    await vi.waitFor(() => {
      expect(states).toContain("listening");
    });

    stop();

    await vi.waitFor(() => {
      expect(events).toContain("close");
    });
    expect(events).toEqual(["open", "cancel-start", "cancel-finish", "release", "close"]);
  });

  // FE-RC-035: open() 失敗時の原因情報を catch {} で握り潰さず onError へ渡す。
  it("open()が失敗した場合、原因情報を破棄せずonErrorへ渡す", async () => {
    const openError = new Error("port busy");
    const port = {
      readable: null,
      open: vi.fn(async () => {
        throw openError;
      }),
      close: vi.fn(async () => {}),
      getInfo: () => ({ usbVendorId: 1, usbProductId: 2 }),
    };
    Object.defineProperty(navigator, "serial", {
      configurable: true,
      value: {
        getPorts: vi.fn(async () => [port]),
        requestPort: vi.fn(),
      },
    });
    storePortInfo("nx600", port.getInfo());
    const states: string[] = [];
    const errors: unknown[] = [];
    const stop = startLabDeviceSlotListen({
      slotKey: "nx600",
      baudRate: 9600,
      isStopped: () => false,
      onState: (state) => {
        states.push(state);
      },
      onFrame: async () => {},
      onError: (error) => {
        errors.push(error);
      },
    });

    await vi.waitFor(() => {
      expect(errors).toContain(openError);
    });
    expect(states).toContain("disconnected");

    stop();
  });
});
