import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { PetDeceasedBanner } from "./PetDeceasedBanner";

const { mockMutate } = vi.hoisted(() => ({ mockMutate: vi.fn() }));
const { capturedConfirm } = vi.hoisted(() => ({
  capturedConfirm: { current: undefined as (() => void) | undefined },
}));

vi.mock("@/components/shared/ConfirmDialog", () => ({
  ConfirmDialog: ({
    open,
    onConfirm,
  }: {
    open: boolean;
    onConfirm: () => void;
  }) => {
    capturedConfirm.current = onConfirm;
    return open ? <button onClick={onConfirm}>解除する</button> : null;
  },
}));

vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: mockMutate, isPending: false }),
}));

// FE4-9 fix 回帰テスト: `new Date(deceasedAt)` + ブラウザローカル TZ getter だった旧実装は
// 非 JST ブラウザで前日を表示するバグがあった。toJSTWallDate 経由に修正後、
// JST 相当の入力（このコンテナの既定 TZ は Asia/Tokyo）で現行出力と同一であることを固定する。
// TZ=America/New_York 等への vitest プロセス TZ 切替による分岐検証は計画上スコープ外
// （FE-refactor.md FE4-9 参照）— toJSTWallDate 自体の TZ 非依存動作は jst-date.test.ts が担保する。
describe("PetDeceasedBanner (FE4-9)", () => {
  it("JST 深夜 instant の deceasedAt を正しい日付（曜日表示なし）で表示する", () => {
    render(
      <PetDeceasedBanner deceasedAt="2026-07-11T00:00:00+09:00" petId="1" canEdit={false} />,
    );
    expect(screen.getByText("2026年7月11日 永眠")).toBeInTheDocument();
  });

  it("時刻付きの deceasedAt でも JST の日付で表示する", () => {
    render(
      <PetDeceasedBanner deceasedAt="2026-07-11T23:30:00+09:00" petId="1" canEdit={false} />,
    );
    expect(screen.getByText("2026年7月11日 永眠")).toBeInTheDocument();
  });

  // BUG-003: GET pets から取得した死亡理由を再表示する
  it("deceasedReason がある場合は死亡理由を表示する", () => {
    render(
      <PetDeceasedBanner
        deceasedAt="2026-07-11T00:00:00+09:00"
        deceasedReason="老衰"
        petId="1"
        canEdit={false}
      />,
    );
    expect(screen.getByText("老衰")).toBeInTheDocument();
    expect(screen.getByText(/死亡理由/)).toBeInTheDocument();
  });

  it("deceasedReason が無い場合は死亡理由行を出さない", () => {
    render(
      <PetDeceasedBanner deceasedAt="2026-07-11T00:00:00+09:00" petId="1" canEdit={false} />,
    );
    expect(screen.queryByText(/死亡理由/)).not.toBeInTheDocument();
  });
});

// BUG-407: 解除成功時に外側フォームのローカル formData（生死ラジオ・deceasedAt）へ
// 同期する onRevoked コールバックの回帰テスト。無いと死亡記録解除後も外側の
// 生死ラジオが「死亡」のまま残り、次に外側「更新」を押すと status="死亡" で
// 上書きされ、直前に解除したはずの deceased_at=null と矛盾する。
describe("PetDeceasedBanner (BUG-407)", () => {
  beforeEach(() => {
    mockMutate.mockReset();
  });

  it("解除確定成功時に onRevoked を呼ぶ", async () => {
    mockMutate.mockImplementation((_petId, options) => {
      options?.onSuccess?.();
      options?.onSettled?.();
    });
    const onRevoked = vi.fn();

    render(
      <PetDeceasedBanner
        deceasedAt="2026-07-11T00:00:00+09:00"
        petId="1"
        canEdit
        onRevoked={onRevoked}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "死亡記録を解除" }));
    fireEvent.click(screen.getByRole("button", { name: "解除する" }));

    await waitFor(() => expect(onRevoked).toHaveBeenCalledTimes(1));
  });

  it("取得済み解除callbackは最新の編集権限がfalseなら解除mutationを発行しない", () => {
    const props = {
      deceasedAt: "2026-07-11T00:00:00+09:00",
      petId: "1",
      canEdit: true,
    };
    const { rerender } = render(<PetDeceasedBanner {...props} />);

    fireEvent.click(screen.getByRole("button", { name: "死亡記録を解除" }));
    const confirm = capturedConfirm.current;
    expect(confirm).toBeDefined();

    rerender(<PetDeceasedBanner {...props} canEdit={false} />);
    confirm?.();

    expect(mockMutate).not.toHaveBeenCalled();
  });
});
