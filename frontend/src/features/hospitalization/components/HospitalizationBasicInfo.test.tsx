import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import type { MasterItem } from "@/types";

import { createInitialHospitalizationFormData } from "../hooks/use-hospitalization-form-model";
import type { HospitalizationFormData } from "../types";

import { HospitalizationBasicInfo } from "./HospitalizationBasicInfo";

const FORM_DATA: HospitalizationFormData = {
  hospitalizationType: "入院",
  ownerName: "山田 花子",
  species: "犬",
  petName: "ポチ",
  petNumber: "P-0001",
  petInsurance: "",
  petDetails: "",
  visit: "",
  nextVisit: "",
  weight: "",
  displayDate: "2026-10-07",
  endDate: "",
  memo: "",
  ownerRequest: "",
  staffNotes: "",
  doctorId: "",
  doctorName: "",
  cageId: "",
  isInsurance: false,
  insuranceCompanyName: "",
  insuranceNumber: "",
};

function renderBasicInfo(overrides: Partial<HospitalizationFormData> = {}) {
  return render(
    <MemoryRouter>
      <HospitalizationBasicInfo
        formData={{ ...FORM_DATA, ...overrides }}
        onChange={vi.fn()}
        cageItems={[]}
      />
    </MemoryRouter>,
  );
}

const cage = (id: string, name: string, overrides: Partial<MasterItem> = {}): MasterItem => ({
  id,
  name,
  category: "cage",
  price: 0,
  status: "active",
  ...overrides,
});

function makeFormData(overrides: Partial<HospitalizationFormData> = {}): HospitalizationFormData {
  return {
    hospitalizationType: "入院",
    ownerName: "",
    species: "",
    petName: "",
    petNumber: "",
    petInsurance: "",
    petDetails: "",
    visit: "",
    nextVisit: "",
    weight: "",
    displayDate: "2026-10-07",
    endDate: "",
    memo: "",
    ownerRequest: "",
    staffNotes: "",
    doctorId: "",
    doctorName: "",
    cageId: "",
    isInsurance: false,
    insuranceCompanyName: "",
    insuranceNumber: "",
    ...overrides,
  };
}

function renderBasicInfoWithCages(cageItems: MasterItem[], formData = makeFormData()) {
  render(
    <MemoryRouter>
      <HospitalizationBasicInfo formData={formData} onChange={vi.fn()} cageItems={cageItems} />
    </MemoryRouter>,
  );
}

describe("HospitalizationBasicInfo", () => {
  it("type group and period inputs have visible associated labels", () => {
    renderBasicInfo();
    expect(screen.getByRole("radiogroup", { name: "入院タイプ" })).toHaveAttribute(
      "aria-labelledby",
      "hospitalization-type-label",
    );
    expect(screen.getByText("開始日", { selector: "label" })).not.toHaveClass("sr-only");
    expect(screen.getByText("終了日", { selector: "label" })).not.toHaveClass("sr-only");
    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByLabelText("終了日")).toBeInTheDocument();
    expect(screen.getByText("期間").tagName).toBe("P");
    expect(screen.getByText("期間")).not.toHaveAttribute("for");
  });
  it("カードは DESIGN.md の24px余白・heading-3見出しに従う", () => {
    renderBasicInfo();

    const heading = screen.getByRole("heading", { name: "基本情報" });
    expect(heading).toHaveClass("text-heading-3", "font-bold");

    const card = heading.closest("div.rounded-lg");
    expect(card).toHaveClass("p-6");
    expect(card).not.toHaveClass("p-2");
  });

  it("「次回の予定」ブロックと「メモ」の間に12pxの余白を確保する", () => {
    renderBasicInfo();

    const nextScheduleBlock = screen.getByText("次回の予定").closest("div.space-y-2");
    expect(nextScheduleBlock).toHaveClass("mb-3");
  });

  it("メモ textarea は共有 Textarea 既定の text-sm(15px) に任せる", () => {
    renderBasicInfo();

    const memo = screen.getByRole("textbox", { name: "メモ" });
    expect(memo).toHaveClass("min-h-20");
    expect(memo).not.toHaveClass("text-base");
  });

  it("保険入力欄は text-base(16px) のまま据え置く", () => {
    renderBasicInfo({ isInsurance: true });

    expect(screen.getByRole("textbox", { name: "保険会社名" })).toHaveClass("text-base");
    expect(screen.getByRole("textbox", { name: "保険番号" })).toHaveClass("text-base");
  });
});

describe("HospitalizationBasicInfo ケージ選択肢 (EMR-306)", () => {
  it("連結メンバーを候補から除外し、代表ケージはメンバー名を ' + ' 連結したラベルで表示する", async () => {
    const user = userEvent.setup();
    renderBasicInfoWithCages([
      cage("10", "A-1", { description: "大型犬用" }),
      // A-3 を A-2 より先に置き、メンバー順が cageItems 順であることも証明する
      cage("12", "A-3", { mergedIntoCageId: "10" }),
      cage("11", "A-2", { mergedIntoCageId: "10" }),
      cage("20", "B-1", { description: "猫用" }),
    ]);

    await user.click(screen.getByRole("combobox", { name: "ケージ・個室" }));

    // 代表ケージ: 代表名 + メンバー名を ' + ' 連結（description suffix は付かない）
    expect(screen.getByRole("option", { name: "A-1 + A-3 + A-2" })).toBeInTheDocument();
    // 連結メンバー単体の選択肢は出さない（BE 422 への UX 防御）
    expect(screen.queryByRole("option", { name: "A-2" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "A-3" })).not.toBeInTheDocument();
    // 非連結ケージは従来どおり name（description）
    expect(screen.getByRole("option", { name: "B-1（猫用）" })).toBeInTheDocument();
  });

  it("フォーム現在値がメンバーケージを指す場合、そのケージは通常ラベルで候補に残す", async () => {
    const user = userEvent.setup();
    renderBasicInfoWithCages(
      [
        cage("10", "A-1"),
        cage("11", "A-2", { mergedIntoCageId: "10", description: "連結前の個室" }),
        cage("20", "B-1"),
      ],
      makeFormData({ cageId: "11" }),
    );

    // 登録後に連結された保存値を select が表示できるよう候補に残す
    const trigger = screen.getByRole("combobox", { name: "ケージ・個室" });
    expect(trigger).toHaveTextContent("A-2（連結前の個室）");

    await user.click(trigger);
    expect(screen.getByRole("option", { name: "A-2（連結前の個室）" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "A-1 + A-2" })).toBeInTheDocument();
  });
});

// EMR-353: 次回予定の選択種別は nextVisit の有無ではなく明示的なフォーム状態。
// 制御 formData を持つ実ハーネスで選択→再renderの回帰を検証する。
function ControlledBasicInfo({ initial }: { initial?: Partial<HospitalizationFormData> }) {
  const [formData, setFormData] = useState<HospitalizationFormData>({
    ...FORM_DATA,
    ...initial,
  });
  return (
    <HospitalizationBasicInfo
      formData={formData}
      onChange={(updates) => setFormData((prev) => ({ ...prev, ...updates }))}
      cageItems={[]}
    />
  );
}

function renderControlledBasicInfo(initial?: Partial<HospitalizationFormData>) {
  return render(
    <MemoryRouter>
      <ControlledBasicInfo initial={initial} />
    </MemoryRouter>,
  );
}

const nextScheduleTrigger = () => screen.getByRole("combobox", { name: "次回の予定" });
const nextDateInput = () => screen.getByLabelText("次回予定日");
const startDateInput = () => screen.getByLabelText("開始日");

async function selectScheduleOption(user: ReturnType<typeof userEvent.setup>, optionName: string) {
  await user.click(nextScheduleTrigger());
  await user.click(screen.getByRole("option", { name: optionName }));
}

function typeDateInto(input: HTMLElement, slashDate: string) {
  fireEvent.focus(input);
  fireEvent.change(input, { target: { value: slashDate } });
  fireEvent.blur(input);
}

describe("HospitalizationBasicInfo 次回予定の選択種別 (EMR-353)", () => {
  it.each([
    { option: "3週後", expected: "2026年10月28日" },
    { option: "4週後", expected: "2026年11月4日" },
    { option: "1年後", expected: "2027年10月7日" },
  ])(
    "プリセット %s の選択ラベルと基準日から計算した日付が再render後も維持される",
    async ({ option, expected }) => {
      const user = userEvent.setup();
      renderControlledBasicInfo({ displayDate: "2026-10-07" });

      await selectScheduleOption(user, option);

      expect(nextScheduleTrigger()).toHaveTextContent(option);
      expect(nextDateInput()).toHaveDisplayValue(new RegExp(expected));
    },
  );

  // R353-01: 新規フォームは次回予定未選択なので「以外（手動）」表示にする。
  // 未計算の「4週後」を既定表示にすると、同じ項目の再選択が Radix の no-op になり
  // 日付が生成されないため。初期値は createInitialHospitalizationFormData をそのまま使う。
  it("新規フォームの初期表示は以外（手動）で、4週後を選択すると計算日が入る", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({
      ...createInitialHospitalizationFormData(),
      displayDate: "2026-10-07",
    });

    expect(nextScheduleTrigger()).toHaveTextContent("以外（手動）");

    await selectScheduleOption(user, "4週後");

    expect(nextScheduleTrigger()).toHaveTextContent("4週後");
    expect(nextDateInput()).toHaveDisplayValue(/2026年11月4日/);
  });

  it("手動でプリセット計算日と同じ日付を入力しても選択種別を維持する", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "2026-10-07" });

    await selectScheduleOption(user, "3週後");
    typeDateInto(nextDateInput(), "2026/10/28");

    expect(nextScheduleTrigger()).toHaveTextContent("3週後");
    expect(nextDateInput()).toHaveDisplayValue(/2026年10月28日/);
  });

  it("手動で計算日と異なる日付を入力すると以外（手動）に切り替わる", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "2026-10-07" });

    await selectScheduleOption(user, "3週後");
    typeDateInto(nextDateInput(), "2026/11/15");

    expect(nextScheduleTrigger()).toHaveTextContent("以外（手動）");
    expect(nextDateInput()).toHaveDisplayValue(/2026年11月15日/);
  });

  it("プリセット選択中に開始日を変更すると計算日が追従する", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "2026-10-07" });

    await selectScheduleOption(user, "3週後");
    typeDateInto(startDateInput(), "2026/10/14");

    expect(nextScheduleTrigger()).toHaveTextContent("3週後");
    expect(nextDateInput()).toHaveDisplayValue(/2026年11月4日/);
  });

  it("以外（手動）の日付は開始日を変更しても保持される", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "2026-10-07" });

    await selectScheduleOption(user, "3週後");
    typeDateInto(nextDateInput(), "2026/12/01");
    typeDateInto(startDateInput(), "2026/10/14");

    expect(nextScheduleTrigger()).toHaveTextContent("以外（手動）");
    expect(nextDateInput()).toHaveDisplayValue(/2026年12月1日/);
  });

  it("基準日が空でもプリセット選択は保持され無効な日付を生成しない", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "" });

    await selectScheduleOption(user, "3週後");

    expect(nextScheduleTrigger()).toHaveTextContent("3週後");
    expect(nextDateInput()).toHaveValue("");
  });

  it("開始日を消去するとプリセット種別は保持しつつ日付は空になる", async () => {
    const user = userEvent.setup();
    renderControlledBasicInfo({ displayDate: "2026-10-07" });

    await selectScheduleOption(user, "3週後");
    typeDateInto(startDateInput(), "");

    expect(nextScheduleTrigger()).toHaveTextContent("3週後");
    expect(nextDateInput()).toHaveValue("");
  });

  it("既存日付だけを持つ formData は以外（手動）として表示する", () => {
    renderControlledBasicInfo({ nextVisit: "2026-12-25" });

    expect(nextScheduleTrigger()).toHaveTextContent("以外（手動）");
    expect(nextDateInput()).toHaveDisplayValue(/2026年12月25日/);
  });
});
