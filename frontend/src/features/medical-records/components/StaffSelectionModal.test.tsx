import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StaffSelectionModal } from "./StaffSelectionModal";
import { useGetStaffs } from "@/hooks/use-staffs";
import type { StaffItem } from "@/hooks/use-staffs";

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: vi.fn(),
}));

const KATAKANA_STAFFS: StaffItem[] = [
  { id: "1", name: "タナカ ハナコ", isActive: true, occupationName: "獣医師" },
  { id: "2", name: "サトウ イチロウ", isActive: true, occupationName: "看護師" },
];

const HIRAGANA_STAFFS: StaffItem[] = [
  { id: "1", name: "たなか はなこ", isActive: true, occupationName: "獣医師" },
  { id: "2", name: "さとう いちろう", isActive: true, occupationName: "看護師" },
];

beforeEach(() => {
  vi.mocked(useGetStaffs).mockReturnValue({ data: KATAKANA_STAFFS } as ReturnType<
    typeof useGetStaffs
  >);
});

function renderModal(props: { selectedStaffName?: string } = {}) {
  return render(
    <StaffSelectionModal
      open={true}
      selectedStaffName={props.selectedStaffName ?? ""}
      onSelect={vi.fn()}
      onOpenChange={vi.fn()}
    />,
  );
}

describe("StaffSelectionModal — カナ混同検索", () => {
  it("open=true で全スタッフが表示される", async () => {
    renderModal();
    expect(await screen.findByText("タナカ ハナコ")).toBeInTheDocument();
    expect(screen.getByText("サトウ イチロウ")).toBeInTheDocument();
  });

  it("ひらがなで検索するとカタカナ名のスタッフにヒットする", async () => {
    const user = userEvent.setup();
    renderModal();

    await screen.findByText("タナカ ハナコ");
    await user.type(screen.getByPlaceholderText("スタッフ名で検索..."), "たなか");

    await waitFor(() => {
      expect(screen.getByText("タナカ ハナコ")).toBeInTheDocument();
      expect(screen.queryByText("サトウ イチロウ")).not.toBeInTheDocument();
    });
  });

  it("カタカナで検索するとひらがな名のスタッフにヒットする", async () => {
    vi.mocked(useGetStaffs).mockReturnValue({ data: HIRAGANA_STAFFS } as ReturnType<
      typeof useGetStaffs
    >);

    const user = userEvent.setup();
    renderModal();

    await screen.findByText("たなか はなこ");
    await user.type(screen.getByPlaceholderText("スタッフ名で検索..."), "タナカ");

    await waitFor(() => {
      expect(screen.getByText("たなか はなこ")).toBeInTheDocument();
      expect(screen.queryByText("さとう いちろう")).not.toBeInTheDocument();
    });
  });

  it("スタッフを選択すると onSelect が呼ばれる", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(
      <StaffSelectionModal
        open={true}
        selectedStaffName=""
        onSelect={onSelect}
        onOpenChange={vi.fn()}
      />,
    );

    await screen.findByText("タナカ ハナコ");
    await user.click(screen.getByText("タナカ ハナコ"));

    expect(onSelect).toHaveBeenCalledWith("1", "タナカ ハナコ");
  });
});
