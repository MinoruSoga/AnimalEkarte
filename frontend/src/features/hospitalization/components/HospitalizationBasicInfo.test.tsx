import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

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

describe("HospitalizationBasicInfo", () => {
  it("カードは DESIGN.md の24px余白・heading-3見出しに従う", () => {
    renderBasicInfo();

    const heading = screen.getByRole("heading", { name: "基本情報" });
    expect(heading).toHaveClass("text-heading-3", "font-bold");

    const card = heading.parentElement;
    expect(card).toHaveClass("p-6");
    expect(card).not.toHaveClass("p-2");
  });

  it("「次回の予定」ブロックと「メモ」の間に12pxの余白を確保する", () => {
    renderBasicInfo();

    const nextScheduleBlock = screen.getByText("次回の予定").parentElement;
    expect(nextScheduleBlock).toHaveClass("mb-3");
  });

  it("メモ textarea は共有 Textarea 既定の text-sm(15px) に任せる", () => {
    renderBasicInfo();

    const memo = screen.getByRole("textbox", { name: "メモ" });
    expect(memo).toHaveClass("min-h-[80px]");
    expect(memo).not.toHaveClass("text-base");
  });

  it("保険入力欄は text-base(16px) のまま据え置く", () => {
    renderBasicInfo({ isInsurance: true });

    expect(screen.getByRole("textbox", { name: "保険会社名" })).toHaveClass("text-base");
    expect(screen.getByRole("textbox", { name: "保険番号" })).toHaveClass("text-base");
  });
});
