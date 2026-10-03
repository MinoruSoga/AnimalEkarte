import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";

import { MedicalRecordImage } from "./MedicalRecordImage";
import { useGetMedicalRecordImages } from "../api/get-medical-record-images";
import type { ImageGalleryGroup } from "../api/get-medical-record-images";

vi.mock("../api/get-medical-record-images");

const apiMocks = vi.hoisted(() => ({ uploadMutate: vi.fn() }));

vi.mock("../api/medical-record-images", () => ({
  useCreateMedicalRecordImages: () => ({ mutate: apiMocks.uploadMutate, isPending: false }),
  useDeleteImage: () => ({ mutate: vi.fn() }),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() },
}));

const permissionState = {
  canView: true,
  canCreate: false,
  canDelete: false,
};

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permissionState,
}));

// src が非 null の場合は <img alt={name}> でレンダリングされ、
// label テキストは DOM に現れないため name の getByText が一意になる。
const IMAGE_GROUPS: ImageGalleryGroup[] = [
  {
    id: 1,
    date: "2026/01/01 10:00:00",
    images: [
      {
        id: 1,
        name: "レントゲン画像",
        src: "http://example.com/1.jpg",
        label: "レントゲン",
        mimeType: "image/jpeg",
      },
    ],
  },
  {
    id: 2,
    date: "2026/01/02 10:00:00",
    images: [
      {
        id: 2,
        name: "エコー検査",
        src: "http://example.com/2.jpg",
        label: "エコー",
        mimeType: "image/jpeg",
      },
    ],
  },
];

beforeEach(() => {
  permissionState.canView = true;
  permissionState.canCreate = false;
  permissionState.canDelete = false;
  apiMocks.uploadMutate.mockClear();
  vi.mocked(toast.error).mockClear();
  vi.mocked(toast.info).mockClear();
  vi.mocked(useGetMedicalRecordImages).mockReturnValue({
    data: IMAGE_GROUPS,
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMedicalRecordImages>);
});

const DROP_GUIDE_TEXT = "ここに画像ファイルをドロップしてアップロード";

/** jsdom は DragEvent/dataTransfer を持たないため、dataTransfer を差し込んだ素の Event を発火する */
function fireFileDragEvent(
  el: Element,
  type: "dragenter" | "dragleave" | "drop",
  files: File[] = [],
) {
  const event = new Event(type, { bubbles: true, cancelable: true });
  Object.defineProperty(event, "dataTransfer", {
    value: { files, types: ["Files"], items: [], dropEffect: "" },
  });
  fireEvent(el, event);
}

describe("MedicalRecordImage — カナ混同検索", () => {
  it("空の検索語は全グループを表示する", () => {
    render(<MedicalRecordImage medicalRecordId="123" />);
    expect(screen.getByText("レントゲン画像")).toBeInTheDocument();
    expect(screen.getByText("エコー検査")).toBeInTheDocument();
  });

  it("625px シェルと競合する min-h-[500px] を持たず局所スクロールする", () => {
    const { container } = render(<MedicalRecordImage medicalRecordId="123" />);
    // D&D 対応で ドロップゾーン wrapper > スクロール領域 の2層構造。
    // 局所スクロールは内側の .overflow-y-auto が担う。
    const scroller = container.querySelector(".overflow-y-auto") as HTMLElement;
    expect(scroller).not.toBeNull();
    expect(scroller.className).not.toContain("min-h-[500px]");
    expect(scroller.className).toContain("min-h-0");
  });

  it("ひらがな「れんとげん」でカタカナ画像名「レントゲン画像」にヒットする", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordImage medicalRecordId="123" />);

    await user.type(screen.getByLabelText("検索単語"), "れんとげん");

    await waitFor(() => {
      expect(screen.getByText("レントゲン画像")).toBeInTheDocument();
      expect(screen.queryByText("エコー検査")).not.toBeInTheDocument();
    });
  });

  it("カタカナ「レントゲン」でカタカナ画像名にヒットする (かな統一検索)", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordImage medicalRecordId="123" />);

    await user.type(screen.getByLabelText("検索単語"), "レントゲン");

    await waitFor(() => {
      expect(screen.getByText("レントゲン画像")).toBeInTheDocument();
      expect(screen.queryByText("エコー検査")).not.toBeInTheDocument();
    });
  });

  it("ひらがな「えこー」でカタカナ画像名「エコー検査」にヒットする", async () => {
    const user = userEvent.setup();
    render(<MedicalRecordImage medicalRecordId="123" />);

    await user.type(screen.getByLabelText("検索単語"), "えこー");

    await waitFor(() => {
      expect(screen.getByText("エコー検査")).toBeInTheDocument();
      expect(screen.queryByText("レントゲン画像")).not.toBeInTheDocument();
    });
  });
});

describe("MedicalRecordImage — SEC-CS-F14 死亡ペット", () => {
  it("作成権限があっても死亡ペットではアップロードボタンを出さない", () => {
    permissionState.canCreate = true;
    render(<MedicalRecordImage medicalRecordId="123" isPetDeceased />);

    expect(screen.queryByRole("button", { name: "画像アップロード" })).not.toBeInTheDocument();
  });

  it("作成権限があり生存ペットならアップロードボタンを出す", () => {
    permissionState.canCreate = true;
    render(<MedicalRecordImage medicalRecordId="123" isPetDeceased={false} />);

    expect(screen.getByRole("button", { name: "画像アップロード" })).toBeInTheDocument();
  });
});

describe("MedicalRecordImage — セクション見出し (BUG-018)", () => {
  it("画像タブのセクション見出しは「画像」であり「検査結果」ではない", () => {
    render(<MedicalRecordImage medicalRecordId="123" />);

    expect(screen.getByRole("heading", { level: 2, name: "画像" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { level: 2, name: "検査結果" })).not.toBeInTheDocument();
  });

  it("画像が空のとき empty-state は「画像がありません」のまま", () => {
    vi.mocked(useGetMedicalRecordImages).mockReturnValue({
      data: [],
      isLoading: false,
    } as unknown as ReturnType<typeof useGetMedicalRecordImages>);

    render(<MedicalRecordImage medicalRecordId="123" />);

    expect(screen.getByText("画像がありません")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 2, name: "画像" })).toBeInTheDocument();
  });
});

describe("MedicalRecordImage — 画像D&D取り込み", () => {
  it("ファイルドラッグ中はドロップガイドを表示し、ドロップで既存アップロードへファイルを渡す", () => {
    permissionState.canCreate = true;
    const { container } = render(<MedicalRecordImage medicalRecordId="123" />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireFileDragEvent(dropZone, "dragenter");
    expect(screen.getByText(DROP_GUIDE_TEXT)).toBeInTheDocument();

    const file = new File(["img"], "xray.jpg", { type: "image/jpeg" });
    fireFileDragEvent(dropZone, "drop", [file]);

    expect(apiMocks.uploadMutate).toHaveBeenCalledTimes(1);
    const passed = apiMocks.uploadMutate.mock.calls[0][0] as File[];
    expect(passed.map((f) => f.name)).toEqual(["xray.jpg"]);
    expect(screen.queryByText(DROP_GUIDE_TEXT)).not.toBeInTheDocument();
  });

  it("非対応形式のファイルは toast.error で拒否しアップロードしない", () => {
    permissionState.canCreate = true;
    const { container } = render(<MedicalRecordImage medicalRecordId="123" />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireFileDragEvent(dropZone, "drop", [new File(["x"], "memo.txt", { type: "text/plain" })]);

    expect(apiMocks.uploadMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(expect.stringContaining("memo.txt"));
  });

  it("作成権限がない場合はガイドを出さずドロップもアップロードしない", () => {
    permissionState.canCreate = false;
    const { container } = render(<MedicalRecordImage medicalRecordId="123" />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireFileDragEvent(dropZone, "dragenter");
    expect(screen.queryByText(DROP_GUIDE_TEXT)).not.toBeInTheDocument();

    fireFileDragEvent(dropZone, "drop", [new File(["i"], "a.jpg", { type: "image/jpeg" })]);
    expect(apiMocks.uploadMutate).not.toHaveBeenCalled();
  });

  it("dragleave でドロップガイドが消える", () => {
    permissionState.canCreate = true;
    const { container } = render(<MedicalRecordImage medicalRecordId="123" />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireFileDragEvent(dropZone, "dragenter");
    expect(screen.getByText(DROP_GUIDE_TEXT)).toBeInTheDocument();

    fireFileDragEvent(dropZone, "dragleave");
    expect(screen.queryByText(DROP_GUIDE_TEXT)).not.toBeInTheDocument();
  });
});

describe("MedicalRecordImage — isLocked 確定/権限ロック (EMR-216)", () => {
  it("ロック中は作成権限があってもアップロード・撮影ボタンを出さない", () => {
    permissionState.canCreate = true;
    render(<MedicalRecordImage medicalRecordId="123" isLocked />);

    expect(screen.queryByRole("button", { name: "画像アップロード" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "撮影" })).not.toBeInTheDocument();
  });

  it("ロック中はドロップガイドを出さず、ドロップしてもアップロード API を呼ばない", () => {
    permissionState.canCreate = true;
    const { container } = render(<MedicalRecordImage medicalRecordId="123" isLocked />);
    const dropZone = container.firstElementChild as HTMLElement;

    fireFileDragEvent(dropZone, "dragenter");
    expect(screen.queryByText(DROP_GUIDE_TEXT)).not.toBeInTheDocument();

    fireFileDragEvent(dropZone, "drop", [new File(["i"], "a.jpg", { type: "image/jpeg" })]);
    expect(apiMocks.uploadMutate).not.toHaveBeenCalled();
  });

  it("ロック中は削除権限があっても削除ボタンを出さない（画像の閲覧は維持）", () => {
    permissionState.canDelete = true;
    render(<MedicalRecordImage medicalRecordId="123" isLocked />);

    expect(screen.getByText("レントゲン画像")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /を削除$/ })).not.toBeInTheDocument();
  });

  it("ロック解除でアップロード・ドロップ導線が復帰する", () => {
    permissionState.canCreate = true;
    const { container, rerender } = render(<MedicalRecordImage medicalRecordId="123" isLocked />);
    expect(screen.queryByRole("button", { name: "画像アップロード" })).not.toBeInTheDocument();

    rerender(<MedicalRecordImage medicalRecordId="123" isLocked={false} />);
    expect(screen.getByRole("button", { name: "画像アップロード" })).toBeInTheDocument();

    const dropZone = container.firstElementChild as HTMLElement;
    fireFileDragEvent(dropZone, "drop", [new File(["i"], "a.jpg", { type: "image/jpeg" })]);
    expect(apiMocks.uploadMutate).toHaveBeenCalledTimes(1);
  });
});
