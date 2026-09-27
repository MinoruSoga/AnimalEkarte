import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { PatientInfoCard } from "./PatientInfoCard";

vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "/pet.png" }));

vi.mock("@/components/shared/Feedback", () => ({
  ImageWithFallback: ({
    src,
    alt,
    className,
  }: {
    src: string;
    alt: string;
    className?: string;
  }) => <img src={src} alt={alt} className={className} />,
}));

const baseProps = {
  ownerName: "山田 太郎",
  petName: "ポチ",
  petNumber: "0001",
  weight: "5.0kg",
};

afterEach(() => {
  vi.useRealTimers();
});

describe("PatientInfoCard petDetails (BUG-006)", () => {
  it("未指定時は固定の「9才5ヶ月 / メス / 避妊済」を出さず不明を表示する", () => {
    render(<PatientInfoCard {...baseProps} />);
    expect(screen.getByText("不明")).toBeInTheDocument();
    expect(screen.queryByText("9歳5ヶ月 / メス / 避妊済")).not.toBeInTheDocument();
  });

  it("渡した petDetails をそのまま表示する", () => {
    render(<PatientInfoCard {...baseProps} petDetails="13歳7ヶ月 / 雄 / 不明" />);
    expect(screen.getByText("13歳7ヶ月 / 雄 / 不明")).toBeInTheDocument();
    expect(screen.queryByText("9歳5ヶ月 / メス / 避妊済")).not.toBeInTheDocument();
  });
});

describe("PatientInfoCard next visit alert", () => {
  it("次回予定日が過去なら日付に加えて非色依存の「期限切れ」表示を出す", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-21T12:00:00+09:00"));

    render(<PatientInfoCard {...baseProps} nextVisitDate="2025/10/10" />);

    expect(screen.getByText("次回 2025/10/10")).toBeInTheDocument();
    expect(screen.getByText("期限切れ")).toBeInTheDocument();
  });

  it("次回予定日が30日より先なら期限アラートを出さない", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-21T12:00:00+09:00"));

    render(<PatientInfoCard {...baseProps} nextVisitDate="2026/10/10" />);

    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });

  it("次回予定日を渡さない画面では仮の日付や期限アラートを表示しない", () => {
    render(<PatientInfoCard {...baseProps} />);

    expect(screen.queryByText("次回 -")).not.toBeInTheDocument();
    expect(screen.queryByText("次回 2025/10/10")).not.toBeInTheDocument();
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });

  it("保険・次回カードに固定 min-w-[120px] を付けない (BUG-458)", () => {
    const { container } = render(
      <PatientInfoCard {...baseProps} insuranceName="アニコム" nextVisitDate="2026/08/01" />,
    );
    const fixedMin = container.querySelector('[class*="min-w-[120px]"], [class*="min-w-[60px]"]');
    expect(fixedMin).toBeNull();
    expect(screen.getByText("次回 2026/08/01")).toBeInTheDocument();
  });

  it("実在しない日付や区切り混在は期限判定へ渡さない", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-21T12:00:00+09:00"));

    const { rerender } = render(<PatientInfoCard {...baseProps} nextVisitDate="2025/02/31" />);
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} nextVisitDate="2025/02-01" />);
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
  });
});

describe("PatientInfoCard 危険マーク (EMR-173)", () => {
  it("ownerIsDangerous=true なら飼主名の横に ⚠ 危険人物 を出す", () => {
    render(<PatientInfoCard {...baseProps} ownerIsDangerous />);

    expect(screen.getByText("⚠ 危険人物")).toBeInTheDocument();
  });

  it("ownerIsDangerous 未指定なら危険人物マークを出さない", () => {
    render(<PatientInfoCard {...baseProps} />);

    expect(screen.queryByText("⚠ 危険人物")).not.toBeInTheDocument();
  });

  it("petDangerLevel=高ならペット名の横に ⚠ 危険 badge を出し理由を開ける", async () => {
    const user = userEvent.setup();
    render(<PatientInfoCard {...baseProps} petDangerLevel="高" petDangerReason="保定時に噛む" />);

    const trigger = screen.getByRole("button", { name: "ポチの危険理由を表示" });
    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
  });

  it("petDangerLevel=中なら黄色 ⚠ 注意 badge を出し、低・未指定は何も出さない", () => {
    const { rerender } = render(<PatientInfoCard {...baseProps} petDangerLevel="中" />);

    expect(screen.getByRole("button", { name: "ポチの注意理由を表示" })).toHaveTextContent(
      "⚠ 注意",
    );
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} petDangerLevel="低" />);
    expect(screen.queryByText("⚠ 注意")).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} />);
    expect(screen.queryByText("⚠ 危険")).not.toBeInTheDocument();
    expect(screen.queryByText("⚠ 注意")).not.toBeInTheDocument();
  });
});
