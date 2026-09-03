import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PatientContextHeader } from "./PatientContextHeader";

// PNG asset import を空文字列にスタブ
vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "" }));

// ImageWithFallback はシンプルな img でよい
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
  ownerName: "田中 太郎",
  petName: "ポチ",
  petNumber: "0001",
};

describe("PatientContextHeader", () => {
  it("飼主名・ペット名・ペット番号が表示される", () => {
    render(<PatientContextHeader {...baseProps} />);
    // Tooltip が同テキストを2箇所レンダリングするため getAllByText を使用
    expect(screen.getAllByText("田中 太郎").length).toBeGreaterThan(0);
    expect(screen.getAllByText("ポチ").length).toBeGreaterThan(0);
    expect(screen.getByText("#0001")).toBeInTheDocument();
  });

  it("birthDate があれば「生（年齢）」形式で表示される", () => {
    render(<PatientContextHeader {...baseProps} birthDate="2020-01-01" species="犬" />);
    // "2020-01-01生（N歳Mヶ月） / 犬" 形式のテキストが存在する
    // <span class="truncate"> と <div role="tooltip"> の両方に同テキストが出るため getAllByText で確認
    const elements = screen.getAllByText(/2020-01-01生/);
    expect(elements.length).toBeGreaterThan(0);
  });

  describe("カルテ確認用ペット属性", () => {
    it("犬の性別・去勢済・品種を表示する", () => {
      render(
        <PatientContextHeader
          {...baseProps}
          species="犬"
          gender="雄"
          neuteredDate="2020-06-01"
          breed="柴犬"
        />,
      );

      expect(screen.getByText("性別")).toBeInTheDocument();
      expect(screen.getByText("雄")).toBeInTheDocument();
      expect(screen.getByText("避妊去勢")).toBeInTheDocument();
      expect(screen.getByText("去勢済")).toBeInTheDocument();
      expect(screen.getByText("品種")).toBeInTheDocument();
      expect(screen.getByText("柴犬")).toBeInTheDocument();
    });

    it("猫の性別・避妊済・品種を表示する", () => {
      render(
        <PatientContextHeader
          {...baseProps}
          species="猫"
          gender="雌"
          neuteredDate="2021-04-15"
          breed="ミックス"
        />,
      );

      expect(screen.getByText("雌")).toBeInTheDocument();
      expect(screen.getByText("避妊済")).toBeInTheDocument();
      expect(screen.getByText("ミックス")).toBeInTheDocument();
    });

    it("性別不明で避妊去勢日があれば中立な済表現を表示する", () => {
      render(
        <PatientContextHeader
          {...baseProps}
          gender="不明"
          neuteredDate="2022-03-10"
          breed="ミックス"
        />,
      );

      expect(screen.getByText("不明")).toBeInTheDocument();
      expect(screen.getByText("避妊・去勢済")).toBeInTheDocument();
    });

    it("不明な性別はそのまま表示し、避妊去勢日と品種の未設定値を推測しない", () => {
      render(<PatientContextHeader {...baseProps} species="猫" gender="不明" breed="" />);

      expect(screen.getByText("不明")).toBeInTheDocument();
      expect(screen.getAllByText("—")).toHaveLength(2);
      expect(screen.queryByText("未去勢")).not.toBeInTheDocument();
      expect(screen.queryByText("未実施")).not.toBeInTheDocument();
    });

    it("属性 props が未指定なら既存呼び出し元へ属性欄を追加しない", () => {
      render(<PatientContextHeader {...baseProps} />);

      expect(screen.queryByText("性別")).not.toBeInTheDocument();
      expect(screen.queryByText("避妊去勢")).not.toBeInTheDocument();
      expect(screen.queryByText("品種")).not.toBeInTheDocument();
    });
  });

  describe("年齢表示の特性テスト（FE3-9: calcAgePartsAt 統合前後で出力を1文字も変えない）", () => {
    afterEach(() => {
      vi.useRealTimers();
    });

    it("誕生日 2026-01-15・基準日 2026-07-11 → 「5ヶ月」（1歳未満は年表記を省略する現行フォーマット）", () => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-07-11T12:00:00+09:00"));
      render(<PatientContextHeader {...baseProps} birthDate="2026-01-15" />);
      const elements = screen.getAllByText("2026-01-15生（5ヶ月）");
      expect(elements.length).toBeGreaterThan(0);
    });

    it("誕生日 2020-07-10・基準日 2026-07-11 → 「6歳0ヶ月」（現行実装の実測値。月0でも「Xヶ月」は省略しない）", () => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-07-11T12:00:00+09:00"));
      render(<PatientContextHeader {...baseProps} birthDate="2020-07-10" />);
      const elements = screen.getAllByText("2020-07-10生（6歳0ヶ月）");
      expect(elements.length).toBeGreaterThan(0);
    });
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

  it("長い飼主名が DOM に保持される（Tooltip で全文確認可能）", () => {
    const longOwnerName = "非常に長い飼主名前サンプル田中山田鈴木佐藤伊藤渡辺一二三四五六";
    render(<PatientContextHeader {...baseProps} ownerName={longOwnerName} />);
    // Tooltip が同テキストを2箇所レンダリングするため getAllByText を使用
    const elements = screen.getAllByText(longOwnerName);
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it("長いペット名が DOM に保持される（Tooltip で全文確認可能）", () => {
    const longPetName = "非常に長いペット名ポチタロウジロウハナコモモチャコロ";
    render(<PatientContextHeader {...baseProps} petName={longPetName} />);
    const elements = screen.getAllByText(longPetName);
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it("長い飼主名・ペット名でも Tooltip の content として完全な名前が保持される", () => {
    const longOwner = "超長飼主名前テスト一二三四五六七八九十";
    const longPet = "超長ペット名テスト一二三四五六七八九十";
    render(<PatientContextHeader {...baseProps} ownerName={longOwner} petName={longPet} />);
    // Tooltip の role="tooltip" 要素に完全な名前が含まれる（portal で常時 DOM に存在、非表示時は aria-hidden=true）
    const tooltips = screen.getAllByRole("tooltip", { hidden: true });
    const ownerTooltip = tooltips.find((el) => el.textContent === longOwner);
    const petTooltip = tooltips.find((el) => el.textContent === longPet);
    expect(ownerTooltip).toBeInTheDocument();
    expect(petTooltip).toBeInTheDocument();
  });
});
