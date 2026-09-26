import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ReceptionDialogBody, ReceptionDialogFooter } from "./ReceptionDetailModalParts";
import type { ReceptionAppointment } from "../api/types";

const trimmingAppointment: ReceptionAppointment = {
  id: "1",
  time: "09:45",
  visitDate: "2026-05-29",
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "トリミング",
  reservationTypeId: "9",
  reservationCategory: "trimming",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

describe("ReceptionDialogFooter", () => {
  it("受付済のトリミング予約では汎用ステータス遷移せずトリミングカルテ作成へ遷移する", async () => {
    const onConfirm = vi.fn();
    const onCreateTrimming = vi.fn();

    render(
      <ReceptionDialogFooter
        currentStatus="受付済"
        appointment={trimmingAppointment}
        isTrimming={true}
        isHospitalization={false}
        isMedical={false}
        onConfirm={onConfirm}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={vi.fn()}
        onCreateTrimming={onCreateTrimming}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );

    await userEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    // EMR-74: トリミング記録作成では汎用の in_consultation 遷移（onConfirm）を発火しない
    expect(onConfirm).not.toHaveBeenCalled();
    expect(onCreateTrimming).toHaveBeenCalledOnce();
  });

  it("診療中のトリミング予約では施術記録ボタンを表示しない", () => {
    render(
      <ReceptionDialogFooter
        currentStatus="診療中"
        appointment={{ ...trimmingAppointment, status: "in_consultation" }}
        isTrimming={true}
        isHospitalization={false}
        isMedical={false}
        onConfirm={vi.fn()}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={vi.fn()}
        onCreateTrimming={vi.fn()}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: /施術記録/ })).not.toBeInTheDocument();
  });
});

describe("ReceptionDialogBody 危険マーク (EMR-173)", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  it("飼主が危険人物なら飼い主名の横に ⚠ 危険人物 を出す", () => {
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, ownerIsDangerous: true }}
        {...relatedPageProps}
      />,
    );

    expect(screen.getByText("山田").parentElement).toHaveTextContent("⚠ 危険人物");
  });

  it("飼主の is_dangerous が未設定なら危険人物マークを出さない", () => {
    render(<ReceptionDialogBody appointment={trimmingAppointment} {...relatedPageProps} />);

    expect(screen.queryByText("⚠ 危険人物")).not.toBeInTheDocument();
  });

  it("ペット危険度 high は患者情報のペット名横に ⚠ 危険 badge を出し理由を開ける", async () => {
    const user = userEvent.setup();
    render(
      <ReceptionDialogBody
        appointment={{
          ...trimmingAppointment,
          petDangerLevel: "high",
          petDangerReason: "保定時に噛む",
        }}
        {...relatedPageProps}
      />,
    );

    const trigger = screen.getByRole("button", { name: "ポチの危険理由を表示" });
    expect(trigger).toHaveTextContent("⚠ 危険");

    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
  });

  it("ペット危険度 medium は黄色 ⚠ 注意 badge を出し、low/未設定は何も出さない", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petDangerLevel: "medium" }}
        {...relatedPageProps}
      />,
    );

    const trigger = screen.getByRole("button", { name: "ポチの注意理由を表示" });
    expect(trigger).toHaveTextContent("⚠ 注意");
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();

    await user.click(trigger);
    expect(await screen.findByText("理由未登録")).toBeInTheDocument();

    rerender(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petDangerLevel: "low" }}
        {...relatedPageProps}
      />,
    );
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();
    expect(screen.queryByText("⚠ 注意")).not.toBeInTheDocument();
  });
});
