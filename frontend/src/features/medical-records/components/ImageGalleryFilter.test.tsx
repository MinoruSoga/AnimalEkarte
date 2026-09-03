import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";

import {
  ImageGalleryFilter,
  MAX_FILE_SIZE_BYTES,
  MAX_UPLOAD_BATCH_BYTES,
  MAX_UPLOAD_FILES,
} from "./ImageGalleryFilter";

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

function renderFilter(onFilesSelected = vi.fn()) {
  return {
    onFilesSelected,
    ...render(
      <ImageGalleryFilter
        searchTerm=""
        onSearchChange={vi.fn()}
        dateStart=""
        onDateStartChange={vi.fn()}
        dateEnd=""
        onDateEndChange={vi.fn()}
        sortOrder="desc"
        onSortOrderChange={vi.fn()}
        onFilesSelected={onFilesSelected}
        canUpload
      />,
    ),
  };
}

function makeFile(name: string, size: number, type = "image/jpeg"): File {
  // size を実体バイトなしで差し替え（50MiB 級の ArrayBuffer 確保を避ける）
  const file = new File([""], name, { type });
  Object.defineProperty(file, "size", { value: size });
  return file;
}

describe("ImageGalleryFilter — SEC-CS-F08 multi-file caps", () => {
  beforeEach(() => {
    vi.mocked(toast.error).mockClear();
  });

  it("件数が上限を超える選択は onFilesSelected を呼ばず toast する", async () => {
    const user = userEvent.setup();
    const { onFilesSelected } = renderFilter();
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input).toBeTruthy();

    const files = Array.from({ length: MAX_UPLOAD_FILES + 1 }, (_, i) =>
      makeFile(`img-${i}.jpg`, 1024),
    );
    await user.upload(input, files);

    expect(onFilesSelected).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(
      `一度にアップロードできるファイルは${MAX_UPLOAD_FILES}件までです`,
    );
  });

  it("合計バイトが上限を超える選択は onFilesSelected を呼ばず toast する", async () => {
    const user = userEvent.setup();
    const { onFilesSelected } = renderFilter();
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    // 2 files whose combined size exceeds batch cap but each is under per-file 10MiB.
    const half = Math.floor(MAX_UPLOAD_BATCH_BYTES / 2) + 1;
    const files = [makeFile("a.jpg", half), makeFile("b.jpg", half)];
    await user.upload(input, files);

    expect(onFilesSelected).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("合計ファイルサイズが上限（50MB）を超えています");
  });

  it("上限内の選択は onFilesSelected に渡す", async () => {
    const user = userEvent.setup();
    const { onFilesSelected } = renderFilter();
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const files = [makeFile("ok.jpg", 1024), makeFile("ok2.jpg", 2048)];
    await user.upload(input, files);

    expect(onFilesSelected).toHaveBeenCalledTimes(1);
    const passed = onFilesSelected.mock.calls[0][0] as File[];
    expect(passed).toHaveLength(2);
    expect(passed.map((f) => f.name)).toEqual(["ok.jpg", "ok2.jpg"]);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("混在する oversized がある選択は valid も送らず whole-batch で拒否する (SEC-CS-F08-R1)", async () => {
    const user = userEvent.setup();
    const { onFilesSelected } = renderFilter();
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    // one valid + one over per-file 10MiB → no partial upload
    const files = [makeFile("ok.jpg", 1024), makeFile("huge.jpg", MAX_FILE_SIZE_BYTES + 1)];
    await user.upload(input, files);

    expect(onFilesSelected).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(expect.stringContaining("huge.jpg"));
  });

  it("アップロード不可時はボタンを出さない", () => {
    render(
      <ImageGalleryFilter
        searchTerm=""
        onSearchChange={vi.fn()}
        dateStart=""
        onDateStartChange={vi.fn()}
        dateEnd=""
        onDateEndChange={vi.fn()}
        sortOrder="desc"
        onSortOrderChange={vi.fn()}
        onFilesSelected={vi.fn()}
        canUpload={false}
      />,
    );
    expect(screen.queryByRole("button", { name: "画像アップロード" })).not.toBeInTheDocument();
  });
});
