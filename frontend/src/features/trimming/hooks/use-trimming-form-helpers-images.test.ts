import { describe, it, expect, vi, beforeEach, beforeAll } from "vitest";
import { useState, StrictMode, createElement, type ChangeEvent } from "react";
import { renderHook, act, waitFor } from "@testing-library/react";
import { AxiosError } from "axios";
import type { TrimmingFormData } from "@/types/trimming";
import {
  buildCreateTrimmingRequest,
  buildUpdateTrimmingRequest,
  DEFAULT_TRIMMING_FORM_DATA,
} from "./trimming-form-utils";
import { useTrimmingFormHydration, useTrimmingFormImages } from "./use-trimming-form-helpers";

const { mockToast, mockUploadTrimmingImage, mockDeleteTrimmingImage, mockMarkDirty } = vi.hoisted(
  () => ({
    mockToast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
    mockUploadTrimmingImage: vi.fn(),
    mockDeleteTrimmingImage: vi.fn(),
    mockMarkDirty: vi.fn(),
  }),
);

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/upload-trimming-image", () => ({
  uploadTrimmingImage: mockUploadTrimmingImage,
}));

vi.mock("../api/delete-trimming-image", () => ({
  deleteTrimmingImage: mockDeleteTrimmingImage,
}));

const STYLE_KEY = "trimming-images/1/style.jpg";
const COMPLETED_KEY = "trimming-images/1/done.jpg";
const STYLE_SIGNED = "https://signed.example/style.jpg";
const COMPLETED_SIGNED = "https://signed.example/done.jpg";
const BLOB_URL = "blob:mock/style.jpg";

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
  }, mockMarkDirty);
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

beforeAll(() => {
  URL.createObjectURL = vi.fn((blob: Blob) => `blob:mock/${(blob as File).name}`);
  URL.revokeObjectURL = vi.fn();
});

describe("useTrimmingFormImages", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockDeleteTrimmingImage.mockResolvedValue(undefined);
  });

  it("選択直後に即時プレビューし、upload 完了で key と signed URL に置き換わる", async () => {
    let resolveUpload: (value: { key: string; signed_url: string }) => void = () => {};
    mockUploadTrimmingImage.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      void result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    // アップロード完了前から blob URL で即時表示される
    expect(result.current.styleImagePreview).toBe(BLOB_URL);
    expect(result.current.isImageUploading).toBe(true);
    // 選択時点で dirty（転送中の離脱も警告対象）
    expect(mockMarkDirty).toHaveBeenCalledTimes(1);

    await act(async () => {
      resolveUpload({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    });

    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
    expect(result.current.isImageUploading).toBe(false);
    // 署名URLへの切替で即時 blob は解放される
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(BLOB_URL);
  });

  it("2フィールド同時選択で key を保持して signed URL をプレビューする", async () => {
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

  it("signed URL が空なら blob プレビューを維持し、key だけを保持する", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: "" });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });

    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.styleImagePreview).toBe(BLOB_URL);
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
      { id: "画像アップロードに失敗しました。ネットワーク接続を確認してください。" },
    );
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(result.current.styleImagePreview).toBe(STYLE_SIGNED);
    expect(result.current.isImageUploading).toBe(false);
  });

  it("後から成功した古いアップロードで新しい key を上書きせず、stale キーを削除する", async () => {
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
    // 遅れて届いた stale 結果のオブジェクトは到達不能 → best-effort 削除
    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(STYLE_KEY);
    // stale blob も即座に解放される
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock/first.jpg");
  });

  it("置き換え時に未保存の先行アップロードキーを削除する", async () => {
    mockUploadTrimmingImage
      .mockResolvedValueOnce({ key: STYLE_KEY, signed_url: STYLE_SIGNED })
      .mockResolvedValueOnce({ key: COMPLETED_KEY, signed_url: COMPLETED_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("done.jpg"));
    });

    expect(result.current.formData.styleImage).toBe(COMPLETED_KEY);
    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(STYLE_KEY);
  });

  it("保存送信後の置き換えでは先行キーを削除しない（所有権がサーバーへ移行）", async () => {
    mockUploadTrimmingImage
      .mockResolvedValueOnce({ key: STYLE_KEY, signed_url: STYLE_SIGNED })
      .mockResolvedValueOnce({ key: COMPLETED_KEY, signed_url: COMPLETED_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    act(() => {
      result.current.markSubmittedKeys();
    });
    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("done.jpg"));
    });

    expect(result.current.formData.styleImage).toBe(COMPLETED_KEY);
    expect(mockDeleteTrimmingImage).not.toHaveBeenCalledWith(STYLE_KEY);
  });

  it("保存送信済みキーは削除ボタンでもクライアント削除しない", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    act(() => {
      result.current.markSubmittedKeys();
    });
    act(() => {
      result.current.removeStyleImage();
    });

    // フォーム値は空に更新されるが、送信済みキーは server 側が所有（参照中の
    // 可能性）→ DELETE は呼ばれない。orphan 側は BE の参照チェックで保護される。
    expect(result.current.formData.styleImage).toBe("");
    expect(result.current.styleImagePreview).toBeNull();
    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
  });

  it("確定 pre-commit 失敗で所有権が戻ったキーは削除ボタンでクライアント削除できる", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    let submitted!: ReturnType<typeof result.current.markSubmittedKeys>;
    act(() => {
      submitted = result.current.markSubmittedKeys();
    });
    // 4xx 応答 = サーバー未 commit が確定 → caller が所有権を戻す
    act(() => {
      result.current.restoreSubmittedKeys(submitted);
    });
    act(() => {
      result.current.removeStyleImage();
    });

    expect(result.current.formData.styleImage).toBe("");
    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(STYLE_KEY);
  });

  it("送信成功後は restore せずキーは server 所有のまま（クライアント削除しない）", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    act(() => {
      result.current.markSubmittedKeys();
    });
    // 応答無し/5xx = commit 状態不明 → restore しない（呼ばない）まま削除
    act(() => {
      result.current.removeStyleImage();
    });

    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
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

  it("明示削除はプレビューを消し、空文字を update に載せて dirty にする", async () => {
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
    expect(mockMarkDirty).toHaveBeenCalled();

    const update = buildUpdateTrimmingRequest(result.current.formData);
    expect(update.style_image).toBe("");
    expect(update).not.toHaveProperty("completed_image");
  });

  it("削除後に返ったアップロード結果では画像を復活させず、キーを削除する", async () => {
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
    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(COMPLETED_KEY);
  });

  it("未保存アップロード済みキーはアンマウント時に削除する", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const { result, unmount } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);

    unmount();

    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(STYLE_KEY);
  });

  it("保存送信済みキーはアンマウント時に削除しない", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const { result, unmount } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    act(() => {
      result.current.markSubmittedKeys();
    });

    unmount();

    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
  });

  it("後から失敗した新規アップロードは確定プレビューへ戻す（破棄済み blob を復元しない）", async () => {
    let resolveFirst: (value: { key: string; signed_url: string }) => void = () => {};
    mockUploadTrimmingImage
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve;
          }),
      )
      .mockRejectedValueOnce(new AxiosError("network"));
    const { result } = renderHook(() => useImageHarness());

    // 第1選択は in-flight のまま → 第2選択で stale 化。第2が失敗したとき
    // 復元先は「確定済みプレビュー」（ここでは未選択=null）であって、
    // stale になった第1の blob ではない。
    await act(async () => {
      void result.current.handleStyleImageChange(imageChange("first.jpg"));
    });
    expect(result.current.styleImagePreview).toBe("blob:mock/first.jpg");

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("second.jpg"));
    });

    expect(result.current.styleImagePreview).toBeNull();
    expect(result.current.formData.styleImage).toBeNull();

    // stale 化した第1が後から resolve しても破棄される
    await act(async () => {
      resolveFirst({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    });
    expect(mockDeleteTrimmingImage).toHaveBeenCalledWith(STYLE_KEY);
    expect(result.current.styleImagePreview).toBeNull();
  });

  it("signed_url 空で blob が確定プレビューになった直後の失敗はその blob へ戻す", async () => {
    mockUploadTrimmingImage
      .mockResolvedValueOnce({ key: STYLE_KEY, signed_url: "" })
      .mockRejectedValueOnce(new AxiosError("network"));
    const { result } = renderHook(() => useImageHarness());

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    // blob が確定プレビューとして生きている（revoke されていない）
    expect(result.current.styleImagePreview).toBe(BLOB_URL);
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);

    await act(async () => {
      await result.current.handleStyleImageChange(imageChange("next.jpg"));
    });

    // 失敗時は確定済み blob プレビューへ復帰する（revoke 済みでないため有効）
    expect(result.current.styleImagePreview).toBe(BLOB_URL);
    expect(result.current.formData.styleImage).toBe(STYLE_KEY);
    expect(URL.revokeObjectURL).not.toHaveBeenCalledWith(BLOB_URL);
  });
  it("unmount後の確定4xx復元は未保存画像を削除し繰り返し復元でも重複削除しない", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const hook = renderHook(() => useImageHarness());
    await act(async () => {
      await hook.result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    const submitted = hook.result.current.markSubmittedKeys();
    const restore = hook.result.current.restoreSubmittedKeys;
    hook.unmount();
    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
    act(() => {
      restore(submitted);
      restore(submitted);
    });
    expect(mockDeleteTrimmingImage).toHaveBeenCalledExactlyOnceWith(STYLE_KEY);
  });

  it("2slot同keyのunmount後4xx復元は削除を1回にまとめる", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const hook = renderHook(() => useImageHarness());
    await act(async () => {
      await hook.result.current.handleStyleImageChange(imageChange("style.jpg"));
      await hook.result.current.handleCompletedImageChange(imageChange("done.jpg"));
    });
    const submitted = hook.result.current.markSubmittedKeys();
    const restore = hook.result.current.restoreSubmittedKeys;
    hook.unmount();
    act(() => {
      restore(submitted);
    });
    expect(mockDeleteTrimmingImage).toHaveBeenCalledExactlyOnceWith(STYLE_KEY);
  });

  it("StrictModeのeffect再開後は生存中4xxを復元し最後のunmountで掃除する", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    const hook = renderHook(() => useImageHarness(), {
      wrapper: ({ children }) => createElement(StrictMode, null, children),
    });
    await act(async () => {
      await hook.result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    const submitted = hook.result.current.markSubmittedKeys();
    act(() => {
      hook.result.current.restoreSubmittedKeys(submitted);
    });
    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
    hook.unmount();
    expect(mockDeleteTrimmingImage).toHaveBeenCalledExactlyOnceWith(STYLE_KEY);
  });

  it("既存レコードの画像はunmount後4xx復元でもcleanup対象にしない", async () => {
    const hook = renderHook(() =>
      useImageHarness({ styleImage: STYLE_SIGNED, completedImage: COMPLETED_SIGNED }, true),
    );
    await waitFor(() => {
      expect(hook.result.current.styleImagePreview).toBe(STYLE_SIGNED);
    });
    const submitted = hook.result.current.markSubmittedKeys();
    const restore = hook.result.current.restoreSubmittedKeys;
    hook.unmount();
    act(() => {
      restore(submitted);
    });
    expect(mockDeleteTrimmingImage).not.toHaveBeenCalled();
  });

  it("shared画像のcleanupが参照409で拒否されても失敗を無視し二重削除しない", async () => {
    mockUploadTrimmingImage.mockResolvedValue({ key: STYLE_KEY, signed_url: STYLE_SIGNED });
    mockDeleteTrimmingImage.mockRejectedValue({ response: { status: 409 } });
    const hook = renderHook(() => useImageHarness());
    await act(async () => {
      await hook.result.current.handleStyleImageChange(imageChange("style.jpg"));
    });
    const submitted = hook.result.current.markSubmittedKeys();
    const restore = hook.result.current.restoreSubmittedKeys;
    hook.unmount();
    await act(async () => {
      restore(submitted);
      restore(submitted);
    });
    expect(mockDeleteTrimmingImage).toHaveBeenCalledExactlyOnceWith(STYLE_KEY);
    expect(mockToast.error).not.toHaveBeenCalled();
  });
});
