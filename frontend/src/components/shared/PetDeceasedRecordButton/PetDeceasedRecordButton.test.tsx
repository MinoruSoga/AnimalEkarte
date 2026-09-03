import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import { PetDeceasedRecordButton } from "./PetDeceasedRecordButton";

const { mockMutate, mockMutateAsync } = vi.hoisted(() => ({
  mockMutate: vi.fn(),
  mockMutateAsync: vi.fn(),
}));

vi.mock("@/hooks/use-record-pet-death", () => ({
  useRecordPetDeath: () => ({ mutateAsync: mockMutateAsync }),
}));

vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: mockMutate, isPending: false }),
}));

const baseProps: React.ComponentProps<typeof PetDeceasedRecordButton> = {
  petId: "42",
  petName: "ポチ",
  petBreed: "柴犬",
  petGender: "male",
  birthDate: "2015-04-14",
  deceasedAt: null,
  canEdit: true,
};

function renderButton(
  overrides: Partial<React.ComponentProps<typeof PetDeceasedRecordButton>> = {},
) {
  render(<PetDeceasedRecordButton {...baseProps} {...overrides} />);
}

describe("PetDeceasedRecordButton (FE12-02 U3)", () => {
  it("生存ペットでも canEdit=false なら死亡登録buttonとdialogを表示しない", () => {
    renderButton({ petStatus: "生存", canEdit: false });

    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("生存ペットかつ canEdit=true ならdanger色のbuttonからdialogを開ける", async () => {
    const user = userEvent.setup();
    renderButton({ petStatus: "生存", canEdit: true });

    const button = screen.getByRole("button", { name: "死亡を記録" });
    expect(button).toHaveClass("text-sm", "underline", "hover:no-underline");
    expect(button.className).toContain(C.danger);
    expect(button.className).not.toContain(C.text50);
    expect(button.className).not.toContain(C.hoverText);

    await user.click(button);

    expect(screen.getByRole("dialog", { name: "死亡を記録する" })).toBeInTheDocument();
  });

  it("死亡statusなのに死亡日時がない場合は不整合を表示して再登録導線を閉じる", () => {
    renderButton({ petStatus: "死亡", deceasedAt: null, canEdit: true });

    expect(
      screen.getByText(
        "生死データに不整合があります（死亡ステータス・死亡日時未登録）。修復は管理者に依頼してください",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("status不明では死亡登録導線を閉じて管理者確認を促す", () => {
    renderButton({ petStatus: "不明", deceasedAt: null, canEdit: true });

    expect(
      screen.getByText("生死ステータスが不明です。死亡記録は管理者による確認後に行ってください"),
    ).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
  });

  it("死亡日時がある場合はpetStatusより優先して従来のBannerを表示する", () => {
    renderButton({
      petStatus: "死亡",
      deceasedAt: "2026-07-10T12:00:00+09:00",
      canEdit: true,
    });

    expect(screen.getByText(/2026年7月10日 永眠/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "死亡記録を解除" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
  });

  it.each([
    [undefined, "未指定"],
    [null, "null"],
  ])("petStatusが%sでは生存扱いせず死亡登録導線を閉じる（%s）", (petStatus) => {
    renderButton({ petStatus, canEdit: true });

    expect(
      screen.getByText("生死ステータスが不明です。死亡記録は管理者による確認後に行ってください"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "死亡を記録" })).not.toBeInTheDocument();
  });
});
