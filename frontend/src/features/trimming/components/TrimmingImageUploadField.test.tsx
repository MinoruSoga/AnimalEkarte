import type { ChangeEvent } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { TrimmingImageUploadField } from "./TrimmingImageUploadField";

const TRIMMING_IMAGE_ACCEPT = "image/jpeg,image/png,image/gif";

function renderField(
  overrides: Partial<{
    label: string;
    preview: string | null;
    previewAlt: string;
    onImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
    onRemoveImage: () => void;
  }> = {},
) {
  const onImageChange = overrides.onImageChange ?? vi.fn();
  const onRemoveImage = overrides.onRemoveImage ?? vi.fn();
  render(
    <TrimmingImageUploadField
      label={overrides.label ?? "希望スタイル画像"}
      preview={overrides.preview === undefined ? null : overrides.preview}
      previewAlt={overrides.previewAlt ?? "Style preview"}
      onImageChange={onImageChange}
      onRemoveImage={onRemoveImage}
    />,
  );
  return { onImageChange, onRemoveImage };
}

function fileInputs(): HTMLInputElement[] {
  return [...document.querySelectorAll('input[type="file"]')] as HTMLInputElement[];
}

function pickerInput(): HTMLInputElement {
  const picker = fileInputs().find((input) => !input.hasAttribute("capture"));
  if (!picker) throw new Error("file picker input missing");
  return picker;
}

function captureInput(): HTMLInputElement {
  const camera = fileInputs().find((input) => input.getAttribute("capture") === "environment");
  if (!camera) throw new Error("capture input missing");
  return camera;
}

function expectAtLeast44px(element: Element) {
  const className = element.getAttribute("class") ?? "";
  const hasMinBox = className.includes("min-h-11") && className.includes("min-w-11");
  const hasHeight = /(?:^|\s)h-11(?:\s|$)/.test(className);
  expect(hasMinBox || hasHeight).toBe(true);
}

describe("TrimmingImageUploadField", () => {
  it("撮影は別の hidden input で、両 input の accept は BE 受理形式に揃える", () => {
    renderField();

    const picker = pickerInput();
    const camera = captureInput();
    expect(fileInputs()).toHaveLength(2);
    expect(picker).not.toBe(camera);
    expect(picker).toHaveAttribute("accept", TRIMMING_IMAGE_ACCEPT);
    expect(picker).not.toHaveAttribute("capture");
    expect(picker.multiple).toBe(false);
    expect(picker.className).toContain("sr-only");

    expect(camera).toHaveAttribute("capture", "environment");
    expect(camera).toHaveAttribute("accept", TRIMMING_IMAGE_ACCEPT);
    expect(camera.accept).not.toContain("pdf");
    expect(camera.accept).not.toContain("image/*");
    expect(picker.accept).not.toContain("image/*");
    expect(camera.multiple).toBe(false);
    expect(camera.className).toContain("hidden");
    expect(camera).toHaveAttribute("hidden");

    expect(screen.getByRole("button", { name: /^撮影$/ })).toBeInTheDocument();
    expect(screen.getByText("画像をアップロード")).toBeInTheDocument();
  });

  it("両方の input が同じ onImageChange を呼ぶ", async () => {
    const user = userEvent.setup();
    const onImageChange = vi.fn<(event: ChangeEvent<HTMLInputElement>) => void>();
    renderField({ onImageChange });

    const file = new File(["img"], "cut.jpg", { type: "image/jpeg" });
    await user.upload(pickerInput(), file);
    await user.upload(captureInput(), file);

    expect(onImageChange).toHaveBeenCalledTimes(2);
    expect(onImageChange.mock.calls[0][0].target).toBe(pickerInput());
    expect(onImageChange.mock.calls[1][0].target).toBe(captureInput());
    expect(onImageChange.mock.calls[0][0].target.files?.[0]?.name).toBe("cut.jpg");
    expect(onImageChange.mock.calls[1][0].target.files?.[0]?.name).toBe("cut.jpg");
  });

  it("ファイル選択と撮影はキーボードで操作でき、名前と44pxターゲットを持つ", async () => {
    const user = userEvent.setup();
    renderField();

    const picker = pickerInput();
    const fieldLabel = screen.getByText("希望スタイル画像");
    expect(fieldLabel).toHaveAttribute("for", picker.id);
    expect(screen.getByLabelText("希望スタイル画像")).toBe(picker);

    const uploadZone = screen.getByText("画像をアップロード").closest("label");
    expect(uploadZone).not.toBeNull();
    expect(uploadZone).toHaveAttribute("for", picker.id);
    expectAtLeast44px(uploadZone as Element);

    const captureButton = screen.getByRole("button", { name: /^撮影$/ });
    expect(captureButton).toHaveAccessibleName("撮影");
    expect(captureButton).toHaveTextContent("撮影");
    expect(captureButton).toHaveAttribute("type", "button");
    expect(captureButton.querySelector("svg")).toHaveAttribute("aria-hidden", "true");
    expectAtLeast44px(captureButton);

    const camera = captureInput();
    const cameraClick = vi.spyOn(camera, "click");
    await user.tab();
    expect(picker).toHaveFocus();
    await user.tab();
    expect(captureButton).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(cameraClick).toHaveBeenCalledTimes(1);
  });

  it("プレビュー中はアップロード枠を出さず、削除と撮影は残る", async () => {
    const user = userEvent.setup();
    const onRemoveImage = vi.fn();
    const onImageChange = vi.fn<(event: ChangeEvent<HTMLInputElement>) => void>();
    renderField({
      label: "完成画像",
      preview: "data:image/png;base64,abc",
      previewAlt: "Completed preview",
      onRemoveImage,
      onImageChange,
    });

    expect(screen.queryByText("画像をアップロード")).not.toBeInTheDocument();
    expect(screen.getByRole("img", { name: "Completed preview" })).toBeInTheDocument();
    expect(screen.getByText("完成画像")).toBeInTheDocument();

    const removeButton = screen.getByRole("button", { name: "画像を削除" });
    expectAtLeast44px(removeButton);
    await user.click(removeButton);
    expect(onRemoveImage).toHaveBeenCalledTimes(1);

    const captureButton = screen.getByRole("button", { name: /^撮影$/ });
    expectAtLeast44px(captureButton);
    const camera = captureInput();
    expect(camera).toHaveAttribute("accept", TRIMMING_IMAGE_ACCEPT);
    const file = new File(["img"], "done.jpg", { type: "image/jpeg" });
    await user.upload(camera, file);
    expect(onImageChange).toHaveBeenCalledTimes(1);
    expect(onImageChange.mock.calls[0][0].target).toBe(camera);
  });
});
