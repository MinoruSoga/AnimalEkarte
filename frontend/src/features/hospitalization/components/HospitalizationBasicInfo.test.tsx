import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import type { MasterItem } from "@/types";

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
