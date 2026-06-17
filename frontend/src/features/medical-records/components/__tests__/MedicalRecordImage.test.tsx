import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { MedicalRecordImage } from "../MedicalRecordImage";
import { useGetMedicalRecordImages } from "../../api/get-medical-record-images";
import type { ImageGalleryGroup } from "../../api/get-medical-record-images";

vi.mock("../../api/get-medical-record-images");

vi.mock("../../api/medical-record-images", () => ({
  useCreateMedicalRecordImages: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteImage: () => ({ mutate: vi.fn() }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: false, canDelete: false }),
}));

// src が非 null の場合は <img alt={name}> でレンダリングされ、
// label テキストは DOM に現れないため name の getByText が一意になる。
const IMAGE_GROUPS: ImageGalleryGroup[] = [
  {
    id: 1,
    date: "2026/01/01 10:00:00",
    images: [
      { id: 1, name: "レントゲン画像", src: "http://example.com/1.jpg", label: "レントゲン", mimeType: "image/jpeg" },
    ],
  },
  {
    id: 2,
    date: "2026/01/02 10:00:00",
    images: [
      { id: 2, name: "エコー検査", src: "http://example.com/2.jpg", label: "エコー", mimeType: "image/jpeg" },
    ],
  },
];

beforeEach(() => {
  vi.mocked(useGetMedicalRecordImages).mockReturnValue({
    data: IMAGE_GROUPS,
    isLoading: false,
  } as unknown as ReturnType<typeof useGetMedicalRecordImages>);
});

describe("MedicalRecordImage — カナ混同検索", () => {
  it("空の検索語は全グループを表示する", () => {
    render(<MedicalRecordImage medicalRecordId="123" />);
    expect(screen.getByText("レントゲン画像")).toBeInTheDocument();
    expect(screen.getByText("エコー検査")).toBeInTheDocument();
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
