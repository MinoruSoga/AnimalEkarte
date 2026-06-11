import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PatientContextHeader } from "../PatientContextHeader";

// PNG asset import を空文字列にスタブ
vi.mock(
  "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png",
  () => ({ default: "" }),
);

// ImageWithFallback はシンプルな img でよい
vi.mock("@/components/shared/Feedback", () => ({
  ImageWithFallback: ({ src, alt, className }: { src: string; alt: string; className?: string }) => (
    <img src={src} alt={alt} className={className} />
  ),
}));

const baseProps = {
  ownerName: "田中 太郎",
  petName: "ポチ",
  petNumber: "0001",
};

describe("PatientContextHeader", () => {
  it("飼主名・ペット名・ペット番号が表示される", () => {
    render(<PatientContextHeader {...baseProps} />);
    expect(screen.getByText("田中 太郎")).toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
    expect(screen.getByText("#0001")).toBeInTheDocument();
  });

  it("birthDate があれば「生（年齢）」形式で表示される", () => {
    render(
      <PatientContextHeader
        {...baseProps}
        birthDate="2020-01-01"
        species="犬"
      />,
    );
    // "2020-01-01生（N歳Mヶ月） / 犬" 形式のテキストが存在する
    // <span class="truncate"> と <div role="tooltip"> の両方に同テキストが出るため getAllByText で確認
    const elements = screen.getAllByText(/2020-01-01生/);
    expect(elements.length).toBeGreaterThan(0);
  });

  it("weight があれば Nkg 形式で表示される", () => {
    render(<PatientContextHeader {...baseProps} weight="3.2kg" />);
    expect(screen.getByText("3.2kg")).toBeInTheDocument();
  });

  it("visitCount があれば「来院 N 回」が表示される", () => {
    render(<PatientContextHeader {...baseProps} visitCount={5} />);
    expect(screen.getByText(/来院 5 回/)).toBeInTheDocument();
  });

  it("visitCount が 0 のとき「来院 N 回」は表示されない", () => {
    render(<PatientContextHeader {...baseProps} visitCount={0} />);
    expect(screen.queryByText(/来院/)).not.toBeInTheDocument();
  });

  it("status='deceased' で【死亡】バッジが表示される", () => {
    render(<PatientContextHeader {...baseProps} status="deceased" />);
    expect(screen.getByText("【死亡】")).toBeInTheDocument();
  });

  it("contextControls が渡されると表示される", () => {
    render(
      <PatientContextHeader
        {...baseProps}
        contextControls={<button type="button">アクション</button>}
      />,
    );
    expect(screen.getByRole("button", { name: "アクション" })).toBeInTheDocument();
  });

  it("contextControls 未指定でもクラッシュしない", () => {
    expect(() => render(<PatientContextHeader {...baseProps} />)).not.toThrow();
  });

  it("onOwnerClick が渡されると飼主名がボタンになる", async () => {
    const onOwnerClick = vi.fn();
    render(<PatientContextHeader {...baseProps} onOwnerClick={onOwnerClick} />);
    const ownerButton = screen.getByRole("button", { name: "田中 太郎" });
    await userEvent.click(ownerButton);
    expect(onOwnerClick).toHaveBeenCalledOnce();
  });
});
