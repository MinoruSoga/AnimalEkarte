import { DndContext } from "@dnd-kit/core";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  DangerLevelHigh,
  DangerLevelLow,
  DangerLevelMedium,
  PetStatusAlive,
  PetStatusDeceased,
} from "@/types/generated/models";
import { C } from "@/lib/design-tokens";

import { AppointmentCard } from "./AppointmentCard";
import type { ReceptionAppointment } from "../api/types";

const { navigateMock } = vi.hoisted(() => ({
  navigateMock: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => navigateMock,
}));

afterEach(() => {
  navigateMock.mockClear();
});

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-05-29",
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
  petStatus: PetStatusAlive,
  petDangerLevel: DangerLevelLow,
};

function renderCard(
  appointment: ReceptionAppointment = baseAppointment,
  columnTitle = "受付済",
  onRecordOpen = vi.fn(),
  onCardClick = vi.fn(),
) {
  return render(
    <DndContext>
      <AppointmentCard
        appointment={appointment}
        columnTitle={columnTitle}
        onCardClick={onCardClick}
        onRecordOpen={onRecordOpen}
      />
    </DndContext>,
  );
}

describe("AppointmentCard", () => {
  it("通常予約のミニアクションを44px以上かつ折返し可能にする", () => {
    renderCard();

    const recordButton = screen.getByRole("button", { name: /ポチのカルテ/ });
    const accountingButton = screen.getByRole("button", { name: /ポチの会計/ });

    expect(recordButton).toHaveClass("min-h-11", "min-w-11");
    expect(accountingButton).toHaveClass("min-h-11", "min-w-11");
    expect(recordButton.parentElement).toHaveClass("flex-wrap");
  });

  it("入院予約のミニアクションを44px以上にする", () => {
    renderCard({
      ...baseAppointment,
      reservationType: "入院",
    });

    expect(screen.getByRole("button", { name: /ポチの会計/ })).toHaveClass("min-h-11", "min-w-11");
    expect(screen.getByRole("button", { name: /ポチの入院登録/ })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("通常カルテ遷移に appointmentId を query と state の両方で渡す", () => {
    const onRecordOpen = vi.fn();
    renderCard(baseAppointment, "受付済", onRecordOpen);

    fireEvent.click(screen.getByRole("button", { name: /ポチのカルテ/ }));

    expect(onRecordOpen).toHaveBeenCalledWith(baseAppointment, "受付済");
    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/new?petId=10&appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定の通常予約は appointmentId を保持してペット選択へ遷移する", () => {
    renderCard({
      ...baseAppointment,
      petId: "",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのカルテ/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/select-pet?appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("通常予約のカルテボタンは受付予約カラムでは表示しない", () => {
    renderCard(baseAppointment, "受付予約");

    expect(screen.queryByRole("button", { name: /ポチのカルテ/ })).not.toBeInTheDocument();
  });

  it("通常予約のカルテボタンは診療中カラムでは表示する", () => {
    const onRecordOpen = vi.fn();
    renderCard(baseAppointment, "診療中", onRecordOpen);

    expect(screen.getByRole("button", { name: /ポチのカルテ/ })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /ポチのカルテ/ }));

    expect(onRecordOpen).toHaveBeenCalledWith(baseAppointment, "診療中");
  });

  it("トリミング予約はカテゴリで判定し、施術遷移に appointmentId を渡す", () => {
    renderCard({
      ...baseAppointment,
      id: "202",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのトリミング記録/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/new?petId=10&appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定のトリミング予約は appointmentId を保持してペット選択へ遷移する", () => {
    renderCard({
      ...baseAppointment,
      id: "202",
      petId: "",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのトリミング記録/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("トリミング予約の施術ボタンは受付済カラム以外では表示しない", () => {
    renderCard(
      {
        ...baseAppointment,
        id: "202",
        reservationType: "シャンプーコース",
        reservationCategory: "trimming",
      },
      "診療中",
    );

    expect(screen.queryByRole("button", { name: /ポチのトリミング記録/ })).not.toBeInTheDocument();
  });

  // EMR-74: trimming の in_consultation（施術中）は受付済カラムに「施術中」バッジで示し、
  // 診療中・診察中とは一切表示しない。
  it("施術中のトリミング予約は「施術中」バッジを表示し「診療中」「診察中」を表示しない", () => {
    renderCard(
      {
        ...baseAppointment,
        id: "202",
        reservationType: "シャンプーコース",
        reservationCategory: "trimming",
        status: "in_consultation",
      },
      "受付済",
    );

    expect(screen.getByText("施術中")).toBeInTheDocument();
    expect(screen.queryByText("診療中")).not.toBeInTheDocument();
    expect(screen.queryByText("診察中")).not.toBeInTheDocument();
  });

  it("施術中でないトリミング予約は「施術中」バッジを表示しない", () => {
    renderCard(
      {
        ...baseAppointment,
        id: "202",
        reservationType: "シャンプーコース",
        reservationCategory: "trimming",
        status: "checked_in",
      },
      "受付済",
    );

    expect(screen.queryByText("施術中")).not.toBeInTheDocument();
  });

  it("一般予約の診療中カードは「施術中」バッジを表示しない", () => {
    renderCard(
      {
        ...baseAppointment,
        status: "in_consultation",
      },
      "診療中",
    );

    expect(screen.queryByText("施術中")).not.toBeInTheDocument();
  });

  it("入院予約では通常カルテボタンを表示しない", () => {
    renderCard({
      ...baseAppointment,
      id: "303",
      reservationType: "入院",
      reservationCategory: "general",
    });

    expect(screen.queryByRole("button", { name: /ポチのカルテ/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチの入院登録/ })).toBeInTheDocument();
  });

  it("ホテル予約では通常カルテボタンを表示せず入院登録導線を表示する", () => {
    renderCard({
      ...baseAppointment,
      id: "404",
      reservationType: "ホテル",
      reservationCategory: "general",
    });

    expect(screen.queryByRole("button", { name: /ポチのカルテ/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチの入院登録/ })).toBeInTheDocument();
  });

  it("死亡した一般診療の badge を表示し、card click・drag・カルテ・会計を抑止する", () => {
    const onCardClick = vi.fn();
    const deceasedAppointment = {
      ...baseAppointment,
      petStatus: PetStatusDeceased,
    };

    renderCard(deceasedAppointment, "受付済", vi.fn(), onCardClick);

    expect(screen.getByText("【死亡】")).toHaveClass(C.bgDanger, C.textWhite);
    expect(screen.queryByRole("button", { name: /ポチのカルテ/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /ポチの会計/ })).not.toBeInTheDocument();
    expect(screen.getByText("山田").closest("[aria-disabled='true']")).toBeInTheDocument();

    fireEvent.click(screen.getByText("山田"));
    expect(onCardClick).not.toHaveBeenCalled();
  });

  it("死亡したトリミング予約の施術・会計を抑止する", () => {
    renderCard({
      ...baseAppointment,
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
      petStatus: PetStatusDeceased,
    });

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /ポチのトリミング記録/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /ポチの会計/ })).not.toBeInTheDocument();
  });

  it("死亡した入院予約の会計・入院登録を抑止する", () => {
    renderCard({
      ...baseAppointment,
      reservationType: "入院",
      petStatus: PetStatusDeceased,
    });

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /ポチの会計/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /ポチの入院登録/ })).not.toBeInTheDocument();
  });

  it("保存済みの危険理由を click で開示し、再 click で閉じても card click を発火しない", async () => {
    const user = userEvent.setup();
    const onCardClick = vi.fn();
    renderCard(
      {
        ...baseAppointment,
        petDangerLevel: DangerLevelHigh,
        petDangerReason: "保定時に噛む",
      },
      "受付済",
      vi.fn(),
      onCardClick,
    );

    const trigger = screen.getByRole("button", {
      name: "ポチの危険理由を表示",
    });

    await user.click(trigger);

    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(onCardClick).not.toHaveBeenCalled();

    await user.click(trigger);

    await waitFor(() => {
      expect(screen.queryByText("保定時に噛む")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(onCardClick).not.toHaveBeenCalled();
  });

  it.each([
    ["undefined", undefined],
    ["空文字", ""],
    ["空白のみ", " \n\t "],
  ])("危険理由が%sの場合は理由未登録を表示する", async (_caseName, petDangerReason) => {
    const user = userEvent.setup();
    renderCard({
      ...baseAppointment,
      petDangerLevel: DangerLevelHigh,
      petDangerReason,
    });

    await user.click(
      screen.getByRole("button", {
        name: "ポチの危険理由を表示",
      }),
    );

    expect(await screen.findByText("理由未登録")).toBeInTheDocument();
  });

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])("危険 badge は%sで同じ trigger から開閉できる", async (_keyName, key) => {
    const user = userEvent.setup();
    renderCard({
      ...baseAppointment,
      petDangerLevel: DangerLevelHigh,
      petDangerReason: "診察台で噛む",
    });

    const trigger = screen.getByRole("button", {
      name: "ポチの危険理由を表示",
    });
    trigger.focus();

    await user.keyboard(key);

    expect(await screen.findByText("診察台で噛む")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");

    trigger.focus();
    await user.keyboard(key);

    await waitFor(() => {
      expect(screen.queryByText("診察台で噛む")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("危険度 high の警告 badge を表示し、既存のカルテ・会計 action は維持する", () => {
    renderCard({
      ...baseAppointment,
      petDangerLevel: DangerLevelHigh,
    });

    expect(screen.getByText("⚠ 危険")).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);
    expect(screen.queryByText("【死亡】")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチのカルテ/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチの会計/ })).toBeInTheDocument();
  });

  it("危険度 medium は黄色 ⚠ 注意 badge を出し、注意理由を card click せず開閉できる", async () => {
    const user = userEvent.setup();
    const onCardClick = vi.fn();
    renderCard(
      {
        ...baseAppointment,
        petDangerLevel: DangerLevelMedium,
        petDangerReason: "興奮しやすい",
      },
      "受付済",
      vi.fn(),
      onCardClick,
    );

    const trigger = screen.getByRole("button", { name: "ポチの注意理由を表示" });
    expect(trigger).toHaveTextContent("⚠ 注意");
    expect(trigger).toHaveClass(C.bgNotice, C.textNotice, C.borderNotice);
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();

    await user.click(trigger);
    expect(await screen.findByText("興奮しやすい")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(onCardClick).not.toHaveBeenCalled();
  });

  it("飼主が危険人物なら飼主名の横に ⚠ 危険人物 を出し card click は発火しない", async () => {
    const user = userEvent.setup();
    const onCardClick = vi.fn();
    renderCard({ ...baseAppointment, ownerIsDangerous: true }, "受付済", vi.fn(), onCardClick);

    const mark = screen.getByText("⚠ 危険人物");
    expect(mark).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);

    await user.click(mark);
    expect(onCardClick).not.toHaveBeenCalled();
  });

  it.each([
    ["false", false],
    ["未設定", undefined],
  ])("飼主の is_dangerous が%sなら危険人物マークを出さない", (_caseName, ownerIsDangerous) => {
    renderCard({ ...baseAppointment, ownerIsDangerous });

    expect(screen.queryByText("⚠ 危険人物")).not.toBeInTheDocument();
  });

  it("生存かつ危険度 low では badge を表示せず、既存 action を維持する", () => {
    renderCard();

    expect(screen.queryByText("【死亡】")).not.toBeInTheDocument();
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチのカルテ/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチの会計/ })).toBeInTheDocument();
  });

  it("pet sentinel が未判定の場合は badge を表示せず、既存 action を維持する", () => {
    renderCard({
      ...baseAppointment,
      petStatus: undefined,
      petDangerLevel: undefined,
    });

    expect(screen.queryByText("【死亡】")).not.toBeInTheDocument();
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチのカルテ/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ポチの会計/ })).toBeInTheDocument();
  });
});
