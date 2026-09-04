import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { SpecialPeriodSection } from "./SpecialPeriodSection";
import type { ClosingSpecialPeriod } from "@/types/generated/models";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { mockCreateMutateAsync, mockDeleteMutateAsync, mockToastSuccess, mockToastError } =
  vi.hoisted(() => ({
    mockCreateMutateAsync: vi.fn(),
    mockDeleteMutateAsync: vi.fn(),
    mockToastSuccess: vi.fn(),
    mockToastError: vi.fn(),
  }));

vi.mock("sonner", () => ({ toast: { success: mockToastSuccess, error: mockToastError } }));

vi.mock("../api/special-periods", () => ({
  useCreateSpecialPeriod: () => ({ mutateAsync: mockCreateMutateAsync }),
  useDeleteSpecialPeriod: () => ({ mutateAsync: mockDeleteMutateAsync }),
}));

vi.mock("@/components/shared/NavigationBlocker/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

function renderSection(ui: JSX.Element) {
  return render(ui, {
    wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter>,
  });
}

function makePeriod(overrides: Partial<ClosingSpecialPeriod> = {}): ClosingSpecialPeriod {
  return {
    id: 1,
    clinic_id: 1,
    start_date: "2026-12-29",
    end_date: "2027-01-03",
    am_pm_boundary: "12:00",
    pm_end: "17:00",
    note: "年末年始",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function fillAndSubmit(fields: {
  start_date: string;
  end_date: string;
  am_pm_boundary: string;
  pm_end: string;
}) {
  fireEvent.change(screen.getByLabelText("開始日"), { target: { value: fields.start_date } });
  fireEvent.change(screen.getByLabelText("終了日"), { target: { value: fields.end_date } });
  fireEvent.change(screen.getByLabelText("午前・午後 区切り時間"), {
    target: { value: fields.am_pm_boundary },
  });
  fireEvent.change(screen.getByLabelText("午後 終了時間"), { target: { value: fields.pm_end } });
  fireEvent.click(screen.getByRole("button", { name: "保存" }));
}

describe("SpecialPeriodSection", () => {
  beforeEach(() => {
    mockCreateMutateAsync.mockReset().mockResolvedValue(undefined);
    mockDeleteMutateAsync.mockReset().mockResolvedValue(undefined);
    mockToastSuccess.mockClear();
    mockToastError.mockClear();
  });

  it("periods が空のとき空状態メッセージを表示する", () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    expect(screen.getByText("特別期間は登録されていません")).toBeInTheDocument();
  });

  it("periods を一覧表示する", () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod()]} canEdit={true} />);
    expect(screen.getByText("2026-12-29 〜 2027-01-03")).toBeInTheDocument();
    expect(screen.getByText(/区切り: 12:00 \/ 終了: 17:00/)).toBeInTheDocument();
  });

  it("削除ボタンで deleteMutation.mutateAsync が id で呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod({ id: 7 })]} canEdit={true} />);
    fireEvent.click(
      screen.getByRole("button", {
        name: "2026-12-29から2027-01-03の特別期間を削除",
      }),
    );

    await waitFor(() => expect(mockDeleteMutateAsync).toHaveBeenCalledWith(7));
  });

  it("追加ボタンでフォームを表示し、キャンセルで閉じる", () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));
    expect(screen.getByLabelText("開始日")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(screen.queryByLabelText("開始日")).not.toBeInTheDocument();
  });

  it("追加・キャンセル・削除を44px以上とし、削除対象を名前で識別できる", () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod()]} canEdit={true} />);

    const addButton = screen.getByRole("button", { name: "新規登録" });
    expect(addButton).toHaveClass("min-h-11");
    const deleteButton = screen.getByRole("button", {
      name: "2026-12-29から2027-01-03の特別期間を削除",
    });
    expect(deleteButton).toHaveClass("size-11");

    fireEvent.click(addButton);
    expect(screen.getByRole("button", { name: "キャンセル" })).toBeInTheDocument();
  });

  it("新規登録はサイドパネルで開始日を入力できる", () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
  });

  it("正常な期間（開始 < 終了、区切り < 終了時刻）で送信すると createMutation が呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    fillAndSubmit({
      start_date: "2026-12-29",
      end_date: "2027-01-03",
      am_pm_boundary: "12:00",
      pm_end: "17:00",
    });

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          start_date: "2026-12-29",
          end_date: "2027-01-03",
          am_pm_boundary: "12:00",
          pm_end: "17:00",
        }),
      ),
    );
  });

  // ギャップ回帰テスト: start_date > end_date / am_pm_boundary > pm_end のクライアント側
  // 大小関係チェックは現状存在しない。フォームは required のみで、逆転した値でも
  // そのまま createMutation に渡ってしまう。このテストは「現状そのまま送信される」
  // 挙動を固定するものであり、バリデーションを追加する場合はこのテストの更新が必要になる。
  it("[既知のギャップ] start_date > end_date でもクライアント側で弾かれず createMutation が呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    fillAndSubmit({
      start_date: "2027-01-03",
      end_date: "2026-12-29", // 開始 > 終了（逆転）
      am_pm_boundary: "12:00",
      pm_end: "17:00",
    });

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ start_date: "2027-01-03", end_date: "2026-12-29" }),
      ),
    );
  });

  it("[既知のギャップ] am_pm_boundary > pm_end でもクライアント側で弾かれず createMutation が呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    fillAndSubmit({
      start_date: "2026-12-29",
      end_date: "2027-01-03",
      am_pm_boundary: "18:00", // 区切り > 終了（逆転）
      pm_end: "17:00",
    });

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ am_pm_boundary: "18:00", pm_end: "17:00" }),
      ),
    );
  });

  it("canEdit=false のとき formAction は mutate せず toast.error する", async () => {
    const { rerender } = renderSection(<SpecialPeriodSection periods={[]} canEdit={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    const form = screen.getByLabelText("開始日").closest("form");
    expect(form).not.toBeNull();

    rerender(<SpecialPeriodSection periods={[]} canEdit={false} />);

    fireEvent.change(screen.getByLabelText("開始日"), { target: { value: "2026-12-29" } });
    fireEvent.change(screen.getByLabelText("終了日"), { target: { value: "2027-01-03" } });
    fireEvent.change(screen.getByLabelText("午前・午後 区切り時間"), {
      target: { value: "12:00" },
    });
    fireEvent.change(screen.getByLabelText("午後 終了時間"), { target: { value: "17:00" } });

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
  });

  it("canEdit=false のとき delete は mutate せず toast.error する", async () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod({ id: 7 })]} canEdit={false} />);
    fireEvent.click(
      screen.getByRole("button", {
        name: "2026-12-29から2027-01-03の特別期間を削除",
      }),
    );

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mockDeleteMutateAsync).not.toHaveBeenCalled();
    expect(mockToastSuccess).not.toHaveBeenCalled();
  });
});
