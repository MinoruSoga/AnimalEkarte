import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { C } from "@/lib/design-tokens";
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

    expect(screen.getByText("次回予定 2025/10/10")).toBeInTheDocument();
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
    expect(screen.queryByText("次回予定 2025/10/10")).not.toBeInTheDocument();
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });

  it("保険・次回カードに固定 min-w-[120px] を付けない (BUG-458)", () => {
    const { container } = render(
      <PatientInfoCard {...baseProps} insuranceName="アニコム" nextVisitDate="2026/08/01" />,
    );
    const fixedMin = container.querySelector('[class*="min-w-[120px]"], [class*="min-w-[60px]"]');
    expect(fixedMin).toBeNull();
    expect(screen.getByText("次回予定 2026/08/01")).toBeInTheDocument();
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

describe("PatientInfoCard 特記マーク (EMR-173/231)", () => {
  it("ownerIsDangerous=true なら飼主名の横にアイコンマークを出す", () => {
    render(<PatientInfoCard {...baseProps} ownerIsDangerous />);

    expect(screen.getByRole("img", { name: "特記" })).toBeInTheDocument();
  });

  it("ownerIsDangerous 未指定なら特記マークを出さない", () => {
    render(<PatientInfoCard {...baseProps} />);

    expect(screen.queryByRole("img", { name: "特記" })).not.toBeInTheDocument();
  });

  it("petDangerLevel=高ならペット名の横にアイコンバッジを出しメモを開ける", async () => {
    const user = userEvent.setup();
    render(<PatientInfoCard {...baseProps} petDangerLevel="高" petDangerReason="保定時に噛む" />);

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    await user.click(trigger);
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
  });

  it("petDangerLevel=中ならアイコンバッジを出し、低・未指定は何も出さない", () => {
    const { rerender } = render(<PatientInfoCard {...baseProps} petDangerLevel="中" />);

    const mediumTrigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(mediumTrigger.textContent).toBe("");
    expect(mediumTrigger).toHaveClass(C.bgNotice, C.textBadgeYellow, C.borderNotice);

    rerender(<PatientInfoCard {...baseProps} petDangerLevel="低" />);
    expect(screen.queryByRole("button", { name: /の詳細を表示/ })).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} />);
    expect(screen.queryByRole("button", { name: /の詳細を表示/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("img", { name: "特記" })).not.toBeInTheDocument();
  });
});

describe("PatientInfoCard 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  it("今月(JST)誕生月の生存ペットはペット名横に Cake バッジを表示する", () => {
    render(<PatientInfoCard {...baseProps} birthDate="2018-10-05" />);

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05" as const],
    ["birthDate 未指定", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    render(<PatientInfoCard {...baseProps} birthDate={birthDate} />);

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=deceased）には誕生月でも Cake バッジを表示しない", () => {
    render(<PatientInfoCard {...baseProps} birthDate="2018-10-05" status="deceased" />);

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("PatientInfoCard 390px レスポンシブ (EMR-391/VQ-015)", () => {
  // VQ-015: shrink-0 の担当者コントロールの横で flex-1 サービス群が ~60px に潰れ、
  // 「診療種別」「トリミング」が1文字ずつ縦割れ、保険・次回カードも過圧縮されていた。
  // BUG-458 の固定 min-w を廃した方針は維持しつつ、レスポンシブな basis/折返しで
  // モバイルに使える幅を確保する。
  it("サービス情報群はモバイルで行全体(basis-full)を使い、lg 以上は従来の flex-1 配置", () => {
    render(<PatientInfoCard {...baseProps} staffName="佐藤" />);

    const serviceCol = screen.getByText("診療種別").parentElement;
    expect(serviceCol).not.toBeNull();
    const serviceGroup = serviceCol?.parentElement;
    expect(serviceGroup).toHaveClass("basis-full", "lg:basis-0");
  });

  it("短い固定ラベルと種別値は1文字ずつ割れない（whitespace-nowrap + shrink-0）", () => {
    render(<PatientInfoCard {...baseProps} reservationType="トリミング" />);

    const serviceCol = screen.getByText("診療種別").parentElement;
    expect(serviceCol).toHaveClass("whitespace-nowrap", "shrink-0");
  });

  it("担当者グループはモバイルで独立行(w-full)へ折り返し、lg 以上は従来の右寄せ", () => {
    render(<PatientInfoCard {...baseProps} staffName="佐藤" />);

    const staffGroup = screen.getByText("担当医佐藤").parentElement;
    expect(staffGroup).toHaveClass("w-full", "sm:w-auto", "sm:ml-auto");
  });

  it("保険名の意図的な truncate 省略は維持される（lg 以上の行内配置のみ。狭幅は省略しない）", () => {
    render(<PatientInfoCard {...baseProps} insuranceName="とても長い保険会社と契約プラン名" />);

    // EMR-412 R3/X5: <lg（820px 含む）では全幅折返し・省略なし。lg 以上で意図的な省略に戻す。
    expect(screen.getByText("とても長い保険会社と契約プラン名")).toHaveClass("lg:truncate");
    expect(screen.getByText("とても長い保険会社と契約プラン名")).not.toHaveClass("truncate");
  });
});

// EMR-412 X14 / R3: 担当者選択ボタンのアクセシブル名へ必須印を additive に反映する。
// 既定ラベル・staffButtonId・aria-invalid は維持する（見た目の * は aria-hidden のまま）。
describe("PatientInfoCard 担当者必須マーク (EMR-412 X14)", () => {
  const staffProps = { onStaffClick: () => {}, staffName: "佐藤", staffLabel: "担当医" };

  it("staffRequiredMark=true のとき担当者ボタンの aria-label 末尾に （必須） を付ける", () => {
    render(<PatientInfoCard {...baseProps} {...staffProps} staffRequiredMark />);

    expect(screen.getByRole("button", { name: "担当医佐藤（必須）" })).toBeInTheDocument();
  });

  it("staffRequiredMark=false/未指定 は既定の aria-label を維持し （必須） を付けない", () => {
    const { rerender } = render(<PatientInfoCard {...baseProps} {...staffProps} />);
    expect(screen.getByRole("button", { name: "担当医佐藤" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /（必須）/ })).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} {...staffProps} staffRequiredMark={false} />);
    expect(screen.getByRole("button", { name: "担当医佐藤" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /（必須）/ })).not.toBeInTheDocument();
  });

  it("rerender で staffRequiredMark を切替えても name が追従し id/aria-invalid/onClick を保持する", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    const { rerender } = render(
      <PatientInfoCard
        {...baseProps}
        onStaffClick={onClick}
        staffName="佐藤"
        staffLabel="担当医"
        staffButtonId="staff-select"
        staffInvalid
        staffRequiredMark
      />,
    );

    const requiredButton = screen.getByRole("button", { name: "担当医佐藤（必須）" });
    expect(requiredButton).toHaveAttribute("id", "staff-select");
    expect(requiredButton).toHaveAttribute("aria-invalid", "true");
    await user.click(requiredButton);
    expect(onClick).toHaveBeenCalledTimes(1);

    rerender(
      <PatientInfoCard
        {...baseProps}
        onStaffClick={onClick}
        staffName="佐藤"
        staffLabel="担当医"
        staffButtonId="staff-select"
        staffInvalid
      />,
    );

    expect(screen.queryByRole("button", { name: /（必須）/ })).not.toBeInTheDocument();
    const plainButton = screen.getByRole("button", { name: "担当医佐藤" });
    expect(plainButton).toHaveAttribute("id", "staff-select");
    expect(plainButton).toHaveAttribute("aria-invalid", "true");
    await user.click(plainButton);
    expect(onClick).toHaveBeenCalledTimes(2);
  });
});

// PR679: 提供されたマイクロチップ番号は付帯属性行ではなく、ペット名と同じ既存の名前行に表示する
// （旧 PatientContextHeader と同じ扱い。petDetails・体重などの付帯属性と混在させない）。
describe("PatientInfoCard マイクロチップ番号の行 (PR679)", () => {
  const microchipNumber = "392140000123456";
  const petDetails = "13歳7ヶ月 / 雄 / 不明";
  const chipProps = { microchipNumber, petDetails };

  it("テキスト表示でもペット名と同じ行にマイクロチップ番号を置く", () => {
    render(<PatientInfoCard {...baseProps} {...chipProps} />);

    const chip = screen.getByLabelText(`マイクロチップ番号 ${microchipNumber}`);
    const nameLeaf = screen.getByText(baseProps.petName, { selector: "span.truncate" });
    const row = nameLeaf.closest("div");

    expect(row).toContainElement(chip);
    expect(row).not.toContainElement(screen.getByText(petDetails));
    expect(row).not.toContainElement(screen.getByText(baseProps.weight));
  });

  it("ペット詳細リンク表示でも同じ行にマイクロチップ番号を置く", () => {
    render(
      <MemoryRouter>
        <PatientInfoCard {...baseProps} {...chipProps} petDetailHref="/owners/1/pets/2" />
      </MemoryRouter>,
    );

    const chip = screen.getByLabelText(`マイクロチップ番号 ${microchipNumber}`);
    const row = screen.getByLabelText("ペット詳細を開く").closest("div");

    expect(row).toContainElement(chip);
    expect(row).not.toContainElement(screen.getByText(petDetails));
    expect(row).not.toContainElement(screen.getByText(baseProps.weight));
  });

  it("表示書式（font-mono・truncate・aria-label）と全文保持を維持する", () => {
    render(<PatientInfoCard {...baseProps} microchipNumber={microchipNumber} />);

    const chip = screen.getByLabelText(`マイクロチップ番号 ${microchipNumber}`);
    expect(chip).toHaveClass("font-mono");
    expect(chip).toHaveClass("truncate");
    expect(chip).toHaveTextContent(microchipNumber);
    // 省略は CSS のみ。チップ本体と Tooltip 内容の両方に全文が残る
    expect(screen.getAllByText(microchipNumber).length).toBeGreaterThan(1);
  });

  it("未指定・空文字では表示せず、rerender で外しても取り残さない", () => {
    const { rerender } = render(<PatientInfoCard {...baseProps} />);
    expect(screen.queryByLabelText(/マイクロチップ番号/)).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} microchipNumber="" />);
    expect(screen.queryByLabelText(/マイクロチップ番号/)).not.toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} microchipNumber={microchipNumber} />);
    expect(screen.getByLabelText(`マイクロチップ番号 ${microchipNumber}`)).toBeInTheDocument();

    rerender(<PatientInfoCard {...baseProps} />);
    expect(screen.queryByLabelText(/マイクロチップ番号/)).not.toBeInTheDocument();
  });
});
