import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("../../hooks/use-medical-record-addenda", () => ({
  useMedicalRecordAddenda: vi.fn(),
  useCreateMedicalRecordAddendum: vi.fn(),
}));

import { AddendumModal } from "./AddendumModal";
import { useCreateMedicalRecordAddendum } from "../../hooks/use-medical-record-addenda";
import { C } from "@/lib/design-tokens";

function renderModal(
  mutateAsync: ReturnType<typeof vi.fn> = vi.fn().mockResolvedValue({}),
  onOpenChange = vi.fn(),
) {
  vi.mocked(useCreateMedicalRecordAddendum).mockReturnValue({
    mutateAsync,
  } as unknown as ReturnType<typeof useCreateMedicalRecordAddendum>);
  return render(<AddendumModal open={true} onOpenChange={onOpenChange} medicalRecordId="1" />);
}

beforeEach(() => {
  vi.resetAllMocks();
});

describe("AddendumModal", () => {
  it("after_text 空でバリデーションエラーを表示する", async () => {
    const user = userEvent.setup();
    renderModal();

    await user.click(screen.getByText("追記を保存"));
    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("修正内容は必須です")).toBeInTheDocument();
    });
  });

  it("reason 空でバリデーションエラーを表示する", async () => {
    const user = userEvent.setup();
    renderModal();

    await user.type(screen.getByLabelText(/修正内容/), "内容あり");
    await user.click(screen.getByText("追記を保存"));
    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("修正理由は必須です")).toBeInTheDocument();
    });
    expect(screen.getByLabelText(/修正内容/)).toHaveValue("内容あり");
  });

  it("reason 501 文字でバリデーションエラーを表示する", async () => {
    const user = userEvent.setup({ delay: null });
    renderModal();

    await user.type(screen.getByLabelText(/修正内容/), "内容");
    await user.click(screen.getByLabelText(/修正理由/));
    await user.paste("あ".repeat(501));
    await user.click(screen.getByText("追記を保存"));
    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/500文字以内/)).toBeInTheDocument();
    });
  });

  it("送信成功で onOpenChange(false) が呼ばれる", async () => {
    const user = userEvent.setup();
    const mutateAsync = vi.fn().mockResolvedValue({});
    const onOpenChange = vi.fn();
    renderModal(mutateAsync, onOpenChange);

    await user.type(screen.getByLabelText(/修正内容/), "修正後テキスト");
    await user.type(screen.getByLabelText(/修正理由/), "誤記があったため");
    await user.click(screen.getByText("追記を保存"));
    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        after_text: "修正後テキスト",
        reason: "誤記があったため",
      });
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it("409 エラーで「確定済みカルテでない」エラーを表示する", async () => {
    const user = userEvent.setup();
    const mutateAsync = vi.fn().mockRejectedValue({ response: { status: 409 } });
    renderModal(mutateAsync);

    await user.type(screen.getByLabelText(/修正内容/), "内容");
    await user.type(screen.getByLabelText(/修正理由/), "理由");
    await user.click(screen.getByText("追記を保存"));
    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("確定済みカルテでないため追記できません")).toBeInTheDocument();
    });
  });

  it("キャンセルボタンで onOpenChange(false) が呼ばれる", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderModal(vi.fn(), onOpenChange);

    await user.click(screen.getByText("キャンセル"));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("『追記を保存』は brand と同じ primary teal + pill を使う", () => {
    renderModal();
    const button = screen.getByRole("button", { name: "追記を保存" });
    expect(button.className).toContain(C.bgActionPrimary);
    expect(button.className).toContain("rounded-full");
  });
});
