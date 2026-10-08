import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Reservation } from "../types";
import { ExpiredTentativePanel } from "./ExpiredTentativePanel";

// EMR-308: 行単位で useGetPet → isPetDeceasedForClinicalWrite の死亡ゲートを行う。
// QueryClientProvider なしで描画するためフックを固定し、テストごとに差し替える。
const getPetResult = vi.hoisted(() => ({
  current: { data: undefined as { status?: string; deceasedAt?: string | null } | undefined },
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => getPetResult.current,
}));

/**
 * EMR-308 A3: 期限超過 tentative の actionable list。
 * カレンダーには出さず、ここで「再設定」（PUT intent）/「終了」（DELETE intent）を提供する。
 */

function makeExpiredTentative(overrides: Partial<Reservation> & { id: string }): Reservation {
  return {
    id: overrides.id,
    // sentinel: 意図日の JST 00:00。意図日は intendedDate に載る。
    start: new Date("2026-09-01T00:00:00+09:00"),
    end: new Date("2026-09-01T00:30:00+09:00"),
    ownerName: "山田",
    petName: "ポチ",
    petType: "犬",
    visitType: "revisit",
    type: "一般診察",
    category: "general",
    reservationTypeId: "1",
    doctor: "",
    doctorId: undefined,
    isDesignated: false,
    status: "pending",
    notes: undefined,
    petId: "10",
    ownerId: "20",
    source: "manual",
    reservationRoute: "record_shortcut",
    clinicId: undefined,
    isTentative: true,
    intendedDate: "2026-09-01",
    hadIntendedDate: true,
    sourceMedicalRecordId: undefined,
    ...overrides,
  };
}

describe("ExpiredTentativePanel (EMR-308 A3)", () => {
  beforeEach(() => {
    getPetResult.current = { data: undefined };
  });

  it("対象が0件なら何も描画しない", () => {
    const { container } = render(
      <ExpiredTentativePanel
        appointments={[]}
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("期限超過・未対応ラベル・件数・各行のペット/飼主/予定日を表示する", () => {
    render(
      <ExpiredTentativePanel
        appointments={[
          makeExpiredTentative({ id: "t1", petName: "ポチ", ownerName: "山田" }),
          makeExpiredTentative({ id: "t2", petName: "タマ", ownerName: "佐藤" }),
        ]}
        canConfirm
        canReset
        canEnd
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );

    const region = screen.getByRole("region", { name: "期限超過・未対応" });
    expect(region).toHaveTextContent("期限超過・未対応（2件）");
    expect(region).toHaveTextContent("ポチ（山田）");
    expect(region).toHaveTextContent("タマ（佐藤）");
    // 予定日は intendedDate を表示する（sentinel 時刻ではない）
    expect(within(region).getAllByText(/予定日: 2026-09-01/)).toHaveLength(2);
  });

  it("確定ボタンで onConfirm に対象 reservation を渡す（confirm-tentative モーダルへ）", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    const target = makeExpiredTentative({ id: "t1", petName: "ポチ" });
    render(
      <ExpiredTentativePanel
        appointments={[target]}
        canConfirm
        canReset
        canEnd
        onConfirm={onConfirm}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "ポチの次回予定を確定" }));
    expect(onConfirm).toHaveBeenCalledWith(target);
  });

  it("再設定ボタンで onReset に対象 reservation を渡す", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    const target = makeExpiredTentative({ id: "t1", petName: "ポチ" });
    render(
      <ExpiredTentativePanel
        appointments={[target]}
        canConfirm
        canReset
        canEnd
        onConfirm={vi.fn()}
        onReset={onReset}
        onEnd={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "ポチの次回予定を再設定" }));
    expect(onReset).toHaveBeenCalledWith(target);
  });

  it("終了ボタンで onEnd に対象 reservation を渡す", async () => {
    const user = userEvent.setup();
    const onEnd = vi.fn();
    const target = makeExpiredTentative({ id: "t1", petName: "ポチ" });
    render(
      <ExpiredTentativePanel
        appointments={[target]}
        canConfirm
        canReset
        canEnd
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={onEnd}
      />,
    );

    await user.click(screen.getByRole("button", { name: "ポチの次回予定を終了" }));
    expect(onEnd).toHaveBeenCalledWith(target);
  });

  it("canReset/canEnd が false のとき対応ボタンを出さない（fail-closed）", () => {
    render(
      <ExpiredTentativePanel
        appointments={[makeExpiredTentative({ id: "t1", petName: "ポチ" })]}
        canConfirm={false}
        canReset={false}
        canEnd={false}
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: /次回予定を確定/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /次回予定を再設定/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /次回予定を終了/ })).not.toBeInTheDocument();
  });

  it("死亡ペットの行では確定ボタンを出さない（再設定・終了は出す）", () => {
    // EMR-308 review fix: 確定は実予約化する臨床 write のため死亡ペットはゲート。
    // intent 操作（再設定/終了）は実予約を作らないため残す。
    getPetResult.current = {
      data: { status: "死亡", deceasedAt: "2026-09-10" },
    };
    render(
      <ExpiredTentativePanel
        appointments={[makeExpiredTentative({ id: "t1", petName: "ポチ" })]}
        canConfirm
        canReset
        canEnd
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: /次回予定を確定/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /次回予定を再設定/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /次回予定を終了/ })).toBeInTheDocument();
  });

  it("pet が未取得（不明）のときは確定ボタンを出す（BE が最終ゲート）", () => {
    getPetResult.current = { data: undefined };
    render(
      <ExpiredTentativePanel
        appointments={[makeExpiredTentative({ id: "t1", petName: "ポチ" })]}
        canConfirm
        onConfirm={vi.fn()}
        onReset={vi.fn()}
        onEnd={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: /次回予定を確定/ })).toBeInTheDocument();
  });
});
