import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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
  checkedInAt: undefined,
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

describe("ReceptionDialogBody 特記マーク (EMR-173/231)", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  it("特記対象の飼主は飼主名の横にアイコンマークを出す", () => {
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, ownerIsDangerous: true }}
        {...relatedPageProps}
      />,
    );

    expect(screen.getByRole("img", { name: "特記" })).toBeInTheDocument();
    expect(screen.getByText("山田").parentElement).not.toHaveTextContent("危険人物");
  });

  it("飼主の is_dangerous が未設定なら特記マークを出さない", () => {
    render(<ReceptionDialogBody appointment={trimmingAppointment} {...relatedPageProps} />);

    expect(screen.queryByRole("img", { name: "特記" })).not.toBeInTheDocument();
  });

  it("ペット特記レベル high は患者情報のペット名横にアイコン badge を出しメモを開ける", async () => {
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

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.textContent).toBe("");

    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
  });

  it("ペット特記レベル medium はアイコン badge を出し、low/未設定は何も出さない", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petDangerLevel: "medium" }}
        {...relatedPageProps}
      />,
    );

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.textContent).toBe("");

    await user.click(trigger);
    expect(await screen.findByText("内容未登録")).toBeInTheDocument();

    rerender(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petDangerLevel: "low" }}
        {...relatedPageProps}
      />,
    );
    expect(screen.queryByRole("button", { name: /の詳細を表示/ })).not.toBeInTheDocument();
    expect(screen.queryByText(/危険|注意/)).not.toBeInTheDocument();
  });
});

describe("ReceptionDialogBody ペット名", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  it("ownerId と petId があるペット名はテキストのボタンで詳細を開く", async () => {
    const user = userEvent.setup();
    const onOpenPetDetail = vi.fn();
    const petName = "ポチ<b>強調</b>";
    const { container } = render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petName, ownerId: "20", petId: "10" }}
        onOpenPetDetail={onOpenPetDetail}
        {...relatedPageProps}
      />,
    );

    const control = screen.getByRole("button", { name: petName });
    expect(control.tagName).toBe("BUTTON");
    expect(control.textContent).toBe(petName);
    expect(container.querySelector("b")).toBeNull();

    await user.click(control);
    expect(onOpenPetDetail).toHaveBeenCalledOnce();
  });

  it("ownerId が無いペット名はプレーンテキストのまま", () => {
    const petName = "ポチ<b>強調</b>";
    const { container } = render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petName, ownerId: "", petId: "10" }}
        onOpenPetDetail={vi.fn()}
        {...relatedPageProps}
      />,
    );

    const name = screen.getByText(petName);
    expect(name.tagName).toBe("SPAN");
    expect(name.textContent).toBe(petName);
    expect(container.querySelector("b")).toBeNull();
    expect(screen.queryByRole("button", { name: petName })).not.toBeInTheDocument();
  });

  it("petId が無いペット名はプレーンテキストのまま", () => {
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, ownerId: "20", petId: "" }}
        onOpenPetDetail={vi.fn()}
        {...relatedPageProps}
      />,
    );

    expect(screen.getByText("ポチ").tagName).toBe("SPAN");
    expect(screen.queryByRole("button", { name: /^ポチ$/ })).not.toBeInTheDocument();
  });
});

describe("ReceptionDialogBody 予約内容", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  it("見出しの下に予約内容の全文をテキストとして表示する", () => {
    const notes = "皮膚のかゆみ\n<b>重要</b>";
    const { container } = render(
      <ReceptionDialogBody appointment={{ ...trimmingAppointment, notes }} {...relatedPageProps} />,
    );

    const heading = screen.getByRole("heading", { name: "予約内容" });
    const body = heading.nextElementSibling;
    expect(body).toBeInstanceOf(HTMLParagraphElement);
    expect(
      heading.compareDocumentPosition(body as Node) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(body?.textContent).toBe(notes);
    expect(body).toHaveClass("whitespace-pre-wrap");
    expect(body).not.toHaveClass("truncate");
    expect(container.querySelector("b")).toBeNull();
  });

  it("予約内容が空または未設定なら見出しを出さない", () => {
    const { rerender } = render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, notes: "" }}
        {...relatedPageProps}
      />,
    );

    expect(screen.queryByRole("heading", { name: "予約内容" })).not.toBeInTheDocument();
    expect(screen.getByText("山田")).toBeInTheDocument();

    rerender(<ReceptionDialogBody appointment={trimmingAppointment} {...relatedPageProps} />);

    expect(screen.queryByRole("heading", { name: "予約内容" })).not.toBeInTheDocument();
    expect(screen.getByText("山田")).toBeInTheDocument();
  });
});

describe("ReceptionDialogBody 予防接種", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  it("canView が false なら予防接種を出さず予約内容とペット名リンクは残る", () => {
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, notes: "皮膚のかゆみ" }}
        onOpenPetDetail={vi.fn()}
        canViewVaccinations={false}
        onOpenVaccinations={vi.fn()}
        {...relatedPageProps}
      />,
    );

    expect(screen.queryByRole("button", { name: "予防接種" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "予約内容" })).toBeInTheDocument();
    expect(screen.getByText("皮膚のかゆみ")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ポチ" })).toBeInTheDocument();
  });

  it("canView が true の予防接種は作成ハンドラを呼ばず閲覧コールバックだけを呼ぶ", async () => {
    const user = userEvent.setup();
    const onOpenVaccinations = vi.fn();
    const onCreateMedicalRecord = vi.fn();
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, notes: "皮膚のかゆみ" }}
        canViewVaccinations
        onOpenVaccinations={onOpenVaccinations}
        {...relatedPageProps}
        onCreateMedicalRecord={onCreateMedicalRecord}
      />,
    );

    await user.click(screen.getByRole("button", { name: "予防接種" }));

    expect(onOpenVaccinations).toHaveBeenCalledOnce();
    expect(onCreateMedicalRecord).not.toHaveBeenCalled();
    expect(screen.getByRole("heading", { name: "予約内容" })).toBeInTheDocument();
  });
});
describe("ReceptionDialogBody 誕生月バッジ (EMR-301)", () => {
  const relatedPageProps = {
    isTrimming: false,
    onCreateMedicalRecord: vi.fn(),
    onCreateTrimming: vi.fn(),
    onCreateAccounting: vi.fn(),
    onCreateHospitalization: vi.fn(),
  };

  beforeEach(() => {
    vi.useFakeTimers();
    // JST 2026-10-15（当月 = 10月）
    vi.setSystemTime(new Date("2026-10-15T01:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("birth_date の月が当月のペット名横に Cake バッジを出し、生年月日を開示できる", () => {
    // fake timers 下は userEvent が凍結するため fireEvent で同期駆動する
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petBirthDate: "2018-10-05" }}
        {...relatedPageProps}
      />,
    );

    const trigger = screen.getByRole("button", { name: "今月誕生日" });
    fireEvent.click(trigger);
    expect(screen.getByText("生年月日: 2018-10-05")).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birth_date 未登録", undefined],
  ])("%s のペットには Cake バッジを出さない", (_label, petBirthDate) => {
    render(
      <ReceptionDialogBody
        appointment={{ ...trimmingAppointment, petBirthDate }}
        {...relatedPageProps}
      />,
    );

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペットには誕生月でも Cake バッジを出さない", () => {
    render(
      <ReceptionDialogBody
        appointment={{
          ...trimmingAppointment,
          petStatus: "deceased",
          petBirthDate: "2018-10-05",
        }}
        {...relatedPageProps}
      />,
    );

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
