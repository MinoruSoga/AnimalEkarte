import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

import { CheckupsTab } from "./CheckupsTab";

// ── mock ──────────────────────────────────────────────────────────────────

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canCreate: true, canEdit: true, canDelete: true })),
}));

vi.mock("../../api/checkups", () => ({
  useGetCheckups: vi.fn(() => ({ data: [], isLoading: false })),
  useCreateCheckup: vi.fn(),
  useUpdateCheckup: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
  useDeleteCheckup: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllCheckupTypes: vi.fn(() => ({
    data: [{ id: "1", name: "定期健診" }],
  })),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: vi.fn(() => ({
    data: [
      { id: "10", name: "田中 医師", isActive: true, occupationName: "獣医師" },
      { id: "11", name: "鈴木 医師", isActive: true, occupationName: "獣医師" },
    ],
  })),
}));

import { useCreateCheckup } from "../../api/checkups";

// ── helpers ──────────────────────────────────────────────────────────────

function renderComponent() {
  return render(<CheckupsTab medicalRecordId="mr-1" />);
}

/** 追加フォームを開き、健診種別を選択した状態にする */
function openAddFormWithType() {
  fireEvent.click(screen.getByText("記録を追加"));
  // 健診種別セレクト：「選択」という option を持つ最初の combobox
  const selects = screen.getAllByRole("combobox");
  const typeSelect = selects.find(
    (s) => (s as HTMLSelectElement).querySelector("option[value='']")?.textContent === "選択"
  ) as HTMLSelectElement;
  fireEvent.change(typeSelect, { target: { value: "1" } });
  return selects;
}

// ── tests ─────────────────────────────────────────────────────────────────

describe("CheckupsTab — doctor field", () => {
  let mutateMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mutateMock = vi.fn();
    vi.mocked(useCreateCheckup).mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as ReturnType<typeof useCreateCheckup>);
  });

  it("追加フォームに担当医セレクトが表示される", () => {
    renderComponent();
    fireEvent.click(screen.getByText("記録を追加"));
    // 担当医 option と staff options が存在する
    expect(screen.getByRole("option", { name: "担当医" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "田中 医師" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "鈴木 医師" })).toBeInTheDocument();
  });

  it("担当医を選択して追加すると doctor_id が payload に含まれる", async () => {
    renderComponent();
    openAddFormWithType();

    // 担当医セレクト：「担当医」という option を持つ combobox
    const doctorSelect = screen.getAllByRole("combobox").find(
      (s) => (s as HTMLSelectElement).querySelector("option[value='']")?.textContent === "担当医"
    ) as HTMLSelectElement;
    expect(doctorSelect).toBeDefined();
    fireEvent.change(doctorSelect, { target: { value: "10" } });

    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateMock).toHaveBeenCalledWith(
        expect.objectContaining({ doctor_id: 10 }),
        expect.anything()
      );
    });
  });

  it("担当医未選択の場合 doctor_id は null で送信される", async () => {
    renderComponent();
    openAddFormWithType();

    // 担当医を選ばずにそのまま追加
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateMock).toHaveBeenCalledWith(
        expect.objectContaining({ doctor_id: null }),
        expect.anything()
      );
    });
  });
});
