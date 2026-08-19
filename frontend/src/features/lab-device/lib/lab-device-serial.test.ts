import { describe, expect, it } from "vitest";

import { LAB_DEVICE_IDLE_TICKS, LabDeviceIdleFrameBuffer } from "./lab-device-serial";

describe("LabDeviceIdleFrameBuffer", () => {
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
