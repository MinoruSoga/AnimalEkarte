import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PetDeceasedDialog } from "./PetDeceasedDialog";

const { mockMutateAsync, mockToastError } = vi.hoisted(() => ({
  mockMutateAsync: vi.fn(),
  mockToastError: vi.fn(),
}));

vi.mock("@/hooks/use-record-pet-death", () => ({
  useRecordPetDeath: () => ({ mutateAsync: mockMutateAsync }),
}));

vi.mock("sonner", () => ({ toast: { error: mockToastError, success: vi.fn() } }));

const TODAY_JST = "2026-08-03";
const TOMORROW_JST = "2026-08-04";

function renderDialog(overrides: Partial<React.ComponentProps<typeof PetDeceasedDialog>> = {}) {
  const onOpenChange = overrides.onOpenChange ?? vi.fn();
  const onRecorded = overrides.onRecorded ?? vi.fn();
  render(
    <PetDeceasedDialog
      open={overrides.open ?? true}
      onOpenChange={onOpenChange}
      petId={overrides.petId ?? "42"}
      petName={overrides.petName ?? "ポチ"}
      petBreed={overrides.petBreed}
      petGender={overrides.petGender}
      petAge={overrides.petAge}
      canEdit={overrides.canEdit ?? true}
      onRecorded={onRecorded}
    />,
  );
  return { onOpenChange, onRecorded };
}

function submitForm() {
  fireEvent.click(screen.getByRole("button", { name: "死亡を記録する" }));
}

describe("PetDeceasedDialog", () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-08-03T14:30:00.000Z"));
    mockMutateAsync.mockReset();
    mockToastError.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("死亡日の初期値は当日日付になる", () => {
    renderDialog();
    const dateInput = screen.getByLabelText(/死亡日/) as HTMLInputElement;
    expect(dateInput.value).toBe(TODAY_JST);
    expect(dateInput.max).toBe(TODAY_JST);
  });

  it("死亡日を空にして実ボタンを押すとfield errorを表示し mutateAsync を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();
    const dateInput = screen.getByLabelText(/死亡日/);

    await user.clear(dateInput);
    await user.click(screen.getByRole("button", { name: "死亡を記録する" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("死亡日を入力してください");
    expect(dateInput).toHaveAttribute("aria-invalid", "true");
    expect(dateInput).toHaveAttribute("aria-describedby", alert.id);
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("未来の日付を指定すると検証エラーを表示し mutateAsync を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();
    const dateInput = screen.getByLabelText(/死亡日/);
    fireEvent.change(dateInput, { target: { value: TOMORROW_JST } });

    await user.click(screen.getByRole("button", { name: "死亡を記録する" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("未来の日付は指定できません");
    expect(dateInput).toHaveAttribute("aria-invalid", "true");
    expect(dateInput).toHaveAttribute("aria-describedby", alert.id);
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("有効な日付・理由で送信すると mutateAsync が正しい引数で呼ばれる", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    renderDialog({ petId: "42" });

    const reasonInput = screen.getByLabelText(/死亡理由/);
    fireEvent.change(reasonInput, { target: { value: "老衰" } });

    submitForm();

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        petId: "42",
        deceasedAt: TODAY_JST,
        deceasedReason: "老衰",
      });
    });
  });

  it("死亡理由が空文字の場合 deceasedReason は undefined として渡す（trim 済み）", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    renderDialog();

    submitForm();

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        petId: "42",
        deceasedAt: TODAY_JST,
        deceasedReason: undefined,
      });
    });
  });

  it("死亡理由が空白のみの場合も undefined として渡す", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    renderDialog();

    const reasonInput = screen.getByLabelText(/死亡理由/);
    fireEvent.change(reasonInput, { target: { value: "   " } });

    submitForm();

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ deceasedReason: undefined }),
      );
    });
  });

  it("登録成功時に onOpenChange(false) を呼びダイアログを閉じる", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    const { onOpenChange } = renderDialog();

    submitForm();

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });

  // BUG-407: バックエンドへの即時保存は既にこのダイアログ内で完結しているが、
  // 外側 PetEditModal のローカル formData（生死ラジオ・deceasedAt）へ結果を
  // 伝える手段が無いと、次に外側「更新」を押した際 status="生存" で
  // 上書きされ deceased_at のみ残る不整合を再現する。onRecorded がその橋渡し。
  it("登録成功時に onRecorded を保存内容付きで呼ぶ", async () => {
    mockMutateAsync.mockResolvedValueOnce(undefined);
    const { onRecorded } = renderDialog();

    const reasonInput = screen.getByLabelText(/死亡理由/);
    fireEvent.change(reasonInput, { target: { value: "老衰" } });

    submitForm();

    await waitFor(() => {
      expect(onRecorded).toHaveBeenCalledWith({
        deceasedAt: TODAY_JST,
        deceasedReason: "老衰",
      });
    });
  });

  it("mutateAsync が失敗した場合 onRecorded は呼ばない", async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error("network error"));
    const { onRecorded } = renderDialog();

    submitForm();

    await screen.findByText("死亡の記録に失敗しました");
    expect(onRecorded).not.toHaveBeenCalled();
  });

  it("mutateAsync が失敗した場合エラーメッセージを表示し onOpenChange は呼ばない", async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error("network error"));
    const { onOpenChange } = renderDialog();

    submitForm();

    expect(await screen.findByText("死亡の記録に失敗しました")).toBeInTheDocument();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("キャンセルボタンで onOpenChange(false) を呼ぶ（mutateAsync は呼ばない）", () => {
    const { onOpenChange } = renderDialog();

    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("取得済みform actionは最新の編集権限がfalseなら死亡記録mutationを発行しない", async () => {
    const props = {
      open: true,
      onOpenChange: vi.fn(),
      petId: "42",
      petName: "ポチ",
      canEdit: true,
    };
    const { rerender } = render(<PetDeceasedDialog {...props} />);
    const form = document.getElementById("pet-deceased-form") as HTMLFormElement;

    rerender(<PetDeceasedDialog {...props} canEdit={false} />);
    fireEvent.submit(form);

    await waitFor(() => expect(mockMutateAsync).not.toHaveBeenCalled());
  });
});
