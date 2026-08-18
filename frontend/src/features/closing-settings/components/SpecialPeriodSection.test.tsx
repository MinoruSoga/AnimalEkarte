import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { SpecialPeriodSection } from "./SpecialPeriodSection";
import type { ClosingSpecialPeriod } from "@/types/generated/models";

const { mockCreateMutateAsync, mockDeleteMutateAsync, mockToastSuccess } = vi.hoisted(() => ({
  mockCreateMutateAsync: vi.fn(),
  mockDeleteMutateAsync: vi.fn(),
  mockToastSuccess: vi.fn(),
}));

vi.mock("sonner", () => ({ toast: { success: mockToastSuccess, error: vi.fn() } }));

vi.mock("../api/special-periods", () => ({
  useCreateSpecialPeriod: () => ({ mutateAsync: mockCreateMutateAsync }),
  useDeleteSpecialPeriod: () => ({ mutateAsync: mockDeleteMutateAsync }),
}));

vi.mock("@/components/shared/NavigationBlocker/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

function renderSection(ui: JSX.Element) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
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
  });

  it("periods が空のとき空状態メッセージを表示する", () => {
    renderSection(<SpecialPeriodSection periods={[]} />);
    expect(screen.getByText("特別期間は登録されていません")).toBeInTheDocument();
  });

  it("periods を一覧表示する", () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod()]} />);
    expect(screen.getByText("2026-12-29 〜 2027-01-03")).toBeInTheDocument();
    expect(screen.getByText(/区切り: 12:00 \/ 終了: 17:00/)).toBeInTheDocument();
  });

  it("削除ボタンで deleteMutation.mutateAsync が id で呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod({ id: 7 })]} />);
    fireEvent.click(
      screen.getByRole("button", {
        name: "2026-12-29から2027-01-03の特別期間を削除",
      }),
    );

    await waitFor(() => expect(mockDeleteMutateAsync).toHaveBeenCalledWith(7));
  });

  it("追加ボタンでフォームを表示し、キャンセルで閉じる", () => {
    renderSection(<SpecialPeriodSection periods={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));
    expect(screen.getByLabelText("開始日")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(screen.queryByLabelText("開始日")).not.toBeInTheDocument();
  });

  it("追加・キャンセル・削除を44px以上とし、削除対象を名前で識別できる", () => {
    renderSection(<SpecialPeriodSection periods={[makePeriod()]} />);

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
    renderSection(<SpecialPeriodSection periods={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
  });

  it("正常な期間（開始 < 終了、区切り < 終了時刻）で送信すると createMutation が呼ばれる", async () => {
    renderSection(<SpecialPeriodSection periods={[]} />);
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
    renderSection(<SpecialPeriodSection periods={[]} />);
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
    renderSection(<SpecialPeriodSection periods={[]} />);
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
});
