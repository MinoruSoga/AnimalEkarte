import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  TrimmingCourseImageField,
  TrimmingCourseWeekdayField,
} from "./TrimmingCourseAttributeFields";

const upload = vi.hoisted(() => vi.fn());
const discard = vi.hoisted(() => vi.fn());

vi.mock("../api/upload-trimming-course-image", () => ({
  uploadTrimmingCourseImage: upload,
  discardTrimmingCourseImage: discard,
}));

describe("TrimmingCourseWeekdayField (EMR-457)", () => {
  it("shows every weekday when nothing is selected", () => {
    render(<TrimmingCourseWeekdayField value={[]} onChange={vi.fn()} />);

    expect(screen.getByText("全曜日")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "月" })).toHaveAttribute("aria-pressed", "false");
  });

  it("toggles a weekday on and off", () => {
    const onChange = vi.fn();
    const { rerender } = render(<TrimmingCourseWeekdayField value={[]} onChange={onChange} />);

    fireEvent.click(screen.getByRole("button", { name: "月" }));
    expect(onChange).toHaveBeenLastCalledWith([1]);

    rerender(<TrimmingCourseWeekdayField value={[1, 6]} onChange={onChange} />);
    expect(screen.getByRole("button", { name: "土" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByText("全曜日")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "月" }));
    expect(onChange).toHaveBeenLastCalledWith([6]);
  });

  it("does not forward row-label clicks to the first weekday button", () => {
    const onChange = vi.fn();
    render(<TrimmingCourseWeekdayField value={[0]} onChange={onChange} />);

    // 行ラッパーが label だとラベル部分のクリックが最初のボタン（日）へ転送される。
    // div 行なので転送されない。
    fireEvent.click(screen.getByText("予約可能な曜日"));
    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "日" })).toHaveAttribute("aria-pressed", "true");
  });
});

function selectFile(file: File) {
  const input = document.querySelector('input[type="file"]') as HTMLInputElement;
  fireEvent.change(input, { target: { files: [file] } });
}

describe("TrimmingCourseImageField (EMR-457)", () => {
  beforeEach(() => {
    upload.mockReset();
    discard.mockReset();
    discard.mockResolvedValue(undefined);
  });

  it("passes the uploaded key and preview url and reports the uploading state", async () => {
    upload.mockResolvedValue({ key: "trimming-images/1/a.jpg", signedUrl: "https://signed/a.jpg" });
    const onUploaded = vi.fn();
    const onUploadingChange = vi.fn();
    render(
      <TrimmingCourseImageField
        currentKey=""
        previewUrl=""
        hasImage={false}
        onUploaded={onUploaded}
        onRemove={vi.fn()}
        onUploadingChange={onUploadingChange}
      />,
    );

    selectFile(new File(["x"], "a.jpg", { type: "image/jpeg" }));

    await waitFor(() =>
      expect(onUploaded).toHaveBeenCalledWith("trimming-images/1/a.jpg", "https://signed/a.jpg"),
    );
    expect(onUploadingChange).toHaveBeenNthCalledWith(1, true);
    expect(onUploadingChange).toHaveBeenLastCalledWith(false);
  });

  it("keeps the current image and shows an error when the upload fails", async () => {
    upload.mockRejectedValue(new Error("network"));
    const onUploaded = vi.fn();
    const onUploadingChange = vi.fn();
    render(
      <TrimmingCourseImageField
        currentKey="trimming-images/1/old.jpg"
        previewUrl="https://signed/old.jpg"
        hasImage
        onUploaded={onUploaded}
        onRemove={vi.fn()}
        onUploadingChange={onUploadingChange}
      />,
    );

    selectFile(new File(["x"], "b.png", { type: "image/png" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "画像をアップロードできませんでした",
    );
    expect(onUploaded).not.toHaveBeenCalled();
    expect(screen.getByRole("img", { name: "コース画像" })).toHaveAttribute(
      "src",
      "https://signed/old.jpg",
    );
    expect(onUploadingChange).toHaveBeenLastCalledWith(false);
  });

  it("removes the image explicitly", () => {
    const onRemove = vi.fn();
    render(
      <TrimmingCourseImageField
        currentKey="trimming-images/1/old.jpg"
        previewUrl="https://signed/old.jpg"
        hasImage
        onUploaded={vi.fn()}
        onRemove={onRemove}
        onUploadingChange={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "画像を削除" }));
    expect(onRemove).toHaveBeenCalledTimes(1);
    // 保存済み（この編集中にアップロードしていない）key はクライアントから消さない
    expect(discard).not.toHaveBeenCalled();
  });

  it("does not forward row-label or preview clicks to the remove button", () => {
    const onRemove = vi.fn();
    render(
      <TrimmingCourseImageField
        currentKey="trimming-images/1/old.jpg"
        previewUrl="https://signed/old.jpg"
        hasImage
        onUploaded={vi.fn()}
        onRemove={onRemove}
        onUploadingChange={vi.fn()}
      />,
    );

    // 行ラッパーが label だと画像ラベル・プレビューのクリックが
    // 最初の対話要素（画像を削除ボタン）へ転送されて誤削除になる。
    // div 行なので転送されない。
    fireEvent.click(screen.getByText("画像"));
    fireEvent.click(screen.getByRole("img", { name: "コース画像" }));
    expect(onRemove).not.toHaveBeenCalled();
    expect(discard).not.toHaveBeenCalled();
  });

  it("discards an unsaved upload when it is replaced", async () => {
    upload
      .mockResolvedValueOnce({ key: "trimming-images/1/a.jpg", signedUrl: "https://signed/a.jpg" })
      .mockResolvedValueOnce({ key: "trimming-images/1/b.jpg", signedUrl: "https://signed/b.jpg" });
    const onUploaded = vi.fn();
    const props = {
      previewUrl: "",
      onUploaded,
      onRemove: vi.fn(),
      onUploadingChange: vi.fn(),
    };
    const { rerender } = render(
      <TrimmingCourseImageField currentKey="" hasImage={false} {...props} />,
    );

    selectFile(new File(["x"], "a.jpg", { type: "image/jpeg" }));
    await waitFor(() => expect(onUploaded).toHaveBeenCalledTimes(1));
    rerender(<TrimmingCourseImageField currentKey="trimming-images/1/a.jpg" hasImage {...props} />);

    selectFile(new File(["y"], "b.jpg", { type: "image/jpeg" }));
    await waitFor(() => expect(onUploaded).toHaveBeenCalledTimes(2));
    expect(discard).toHaveBeenCalledWith("trimming-images/1/a.jpg");
    expect(discard).toHaveBeenCalledTimes(1);
  });
});
