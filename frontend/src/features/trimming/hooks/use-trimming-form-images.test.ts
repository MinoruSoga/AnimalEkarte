import { describe, it, expect, vi, beforeEach } from "vitest";
import { useState, type ChangeEvent } from "react";
import { renderHook, act, waitFor } from "@testing-library/react";
import { AxiosError } from "axios";
import type { TrimmingFormData } from "@/types/trimming";
import {
  buildCreateTrimmingRequest,
  buildUpdateTrimmingRequest,
  DEFAULT_TRIMMING_FORM_DATA,
} from "./trimming-form-utils";
import { useTrimmingFormHydration, useTrimmingFormImages } from "./use-trimming-form-helpers";

const { mockToast, mockUploadTrimmingImage } = vi.hoisted(() => ({
  mockToast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
  mockUploadTrimmingImage: vi.fn(),
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/upload-trimming-image", () => ({
  uploadTrimmingImage: mockUploadTrimmingImage,
}));

const STYLE_KEY = "trimming-images/1/style.jpg";
const COMPLETED_KEY = "trimming-images/1/done.jpg";
const STYLE_SIGNED = "https://signed.example/style.jpg";
const COMPLETED_SIGNED = "https://signed.example/done.jpg";

function imageChange(name: string): ChangeEvent<HTMLInputElement> {
  const input = document.createElement("input");
  const file = new File(["img"], name, { type: "image/jpeg" });
  Object.defineProperty(input, "files", { configurable: true, value: [file] });
  return { target: input } as ChangeEvent<HTMLInputElement>;
}

function useImageHarness(
  existing?: { styleImage?: string | null; completedImage?: string | null },
  isEdit = false,
) {
  const [overrides, setOverrides] = useState<Partial<TrimmingFormData>>({});
  const images = useTrimmingFormImages((updater) => {
    setOverrides((prev) => updater(prev));
  });
  useTrimmingFormHydration({
    isEdit,
    existingTrimming: isEdit ? existing : undefined,
    existingAppointmentTrimming: isEdit ? undefined : existing,
    setLocalOverrides: setOverrides,
    setStyleImagePreview: images.setStyleImagePreview,
    setCompletedImagePreview: images.setCompletedImagePreview,
  });
  return {
    formData: { ...DEFAULT_TRIMMING_FORM_DATA, ...overrides },
    ...images,
  };
}

describe("useTrimmingFormImages", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("選択直後にアップロードし、key を保持して signed URL をプレビューする", async () => {
    mockUploadTrimmingImage
      .mockResolvedValueOnce({ key: STYLE_KEY, signed_url: STYLE_SIGNED })
      .mockResolvedValueOnce({ key: COMPLETED_KEY, signed_url: COMPLETED_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
      await result.current.handleCompletedImageChange(imageChange("done.jpg"));
    });

    expect(mockUploadTrimmingImage).toHaveBeenCalledTimes(2);
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.formData.completedImage).toBe(COMPLETED_KEY);
    expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
    expect(result.current.completedImagePreview).toBe(COMPLETED_SIGNED);
    expect(String(result.current.formData.styleImage)).not.toContain("signed.example");

    const req = buildCreateTrimmingRequest(result.current.formData, 10, 9, undefined, undefined);
    expect(req.style_image).toBe(STYLE_KEY);
    expect(req.completed_image).toBe(COMPLETED_KEY);
  });

  it("signed URL が空なら data URL をプレビューにし、key だけを保持する", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: "" });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });

    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.styleImagePreview).toMatch(/^data:/);
  });

  it("アップロード失敗時は toast し、既存の key とプレビューを変えない", async () => {
    mockUploadTrimmingImage
      .mockResolvedValueOnce({ key: STYLE_KEY, signed_url: STYLE_SIGNED })
      .mockRejectedValueOnce(new AxiosError("network"));
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("next.jpg"));
    });

    expect(mockToast.error).toHaveBeenCalledWith(
      "画像アップロードに失敗しました。ネットワーク接続を確認してください。",
    );
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
  });

  it("後から成功した古いアップロードで新しい key を上書きしない", async () => {
    let resolveFirst: (value: { key: string; signed_url: string }) => void = () => {};
    mockUploadTrimmingImage
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve;
          }),
      )
      .mockResolvedValueOnce({ key: COMPLETED_KEY, signed_url: COMPLETED_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    let first: Promise<void> | undefined;
    await act(async () => {
      first = result.current.handleStyleImageChange(imageChange("first.jpg"));
      await result.current.handleStyleImageChange(imageChange("second.jpg"));
      resolveFirst({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
      await first;
    });

    expect(result.current.formData.styleImage).toBe(COMPLETED_KEY);
    expect(result.current.styleImagePreview).toBe(COMPLETED_SIGNED);
  });

  it("編集の署名 URL はプレビューのみに使い、未変更の再保存では画像を省略する", async () => {
    const { result } = renderHook(() =>
      useImageHarness({ styleImage: STYLE_SIGNED, completedImage: COMPLETED_SIGNED }, true),
    );

    await waitFor(() => {
      expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
    });
    expect(result.current.completedImagePreview).toBe(COMPLETED_SIGNED);
    expect(result.current.formData.styleImage).toBeNull();
    expect(result.current.formData.completedImage).toBeNull();

    const untouched = buildUpdateTrimmingRequest(result.current.formData);
    expect(untouched).not.toHaveProperty("style_image");
    expect(untouched).not.toHaveProperty("completed_image");
  });

  it("明示削除はプレビューを消し、空文字を update に載せる", async () => {
    const { result } = renderHook(() =>
      useImageHarness({ styleImage: STYLE_SIGNED, completedImage: COMPLETED_SIGNED }, true),
    );
    await waitFor(() => {
      expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
    });

    act(() => {
      result.current.removeStyleImage();
    });

    expect(result.current.styleImagePreview).toBeNull();
    expect(result.current.completedImagePreview).toBe(COMPLETED_SIGNED);
    expect(result.current.formData.styleImage).toBe("");
    expect(result.current.formData.completedImage).toBeNull();

    const update = buildUpdateTrimmingRequest(result.current.formData);
    expect(update.style_image).toBe("");
    expect(update).not.toHaveProperty("completed_image");
  });

  it("削除後に返ったアップロード結果では画像を復活させない", async () => {
    let resolveUpload: (value: { key: string; signed_url: string }) => void = () => {};
    mockUploadTrimmingImage.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const { result } = renderHook(() => useImageHarness());

    let pending: Promise<void> | undefined;
    act(() => {
      pending = result.current.handleCompletedImageChange(imageChange("done.jpg"));
    });
    act(() => {
      result.current.removeCompletedImage();
    });
    await act(async () => {
      resolveUpload({ key: COMPLETED_KEY, signed_url: COMPLETED_SIGNED });
      await pending;
    });

    expect(result.current.formData.completedImage).toBe("");
    expect(result.current.completedImagePreview).toBeNull();
    expect(mockToast.error).not.toHaveBeenCalled();
  });
});
