import { describe, expect, it, vi } from "vitest";

import { capturePageScreenshot } from "./capture-screenshot";

const { html2canvasMock, toBlobMock } = vi.hoisted(() => ({
  html2canvasMock: vi.fn(),
  toBlobMock: vi.fn(),
}));

vi.mock("html2canvas-pro", () => ({
  default: html2canvasMock,
}));

describe("capturePageScreenshot", () => {
  it("canvas から PNG File を生成する", async () => {
    toBlobMock.mockImplementation((cb: (b: Blob | null) => void) => {
      cb(new Blob(["png-bytes"], { type: "image/png" }));
    });
    html2canvasMock.mockResolvedValue({ toBlob: toBlobMock });

    const file = await capturePageScreenshot();

    expect(html2canvasMock).toHaveBeenCalledWith(document.body, expect.any(Object));
    expect(file).not.toBeNull();
    expect(file?.type).toBe("image/png");
    expect(file?.name.endsWith(".png")).toBe(true);
  });

  it("キャプチャ失敗時は null を返す（送信は継続可能）", async () => {
    html2canvasMock.mockRejectedValue(new Error("capture failed"));

    const file = await capturePageScreenshot();

    expect(file).toBeNull();
  });

  it("toBlob が null を返した場合も null を返す", async () => {
    toBlobMock.mockImplementation((cb: (b: Blob | null) => void) => cb(null));
    html2canvasMock.mockResolvedValue({ toBlob: toBlobMock });

    const file = await capturePageScreenshot();

    expect(file).toBeNull();
  });
});
