import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import { VaccinationForm } from "./VaccinationForm";

const noop = () => {};

const baseProps = {
  vaccineOptions: [] as { value: string; label: string }[],
  vaccineName: "",
  setVaccineName: noop,
  date: "",
  setDate: noop,
  supplemental: "",
  setSupplemental: noop,
  lot1: "",
  setLot1: noop,
  lot2: "",
  setLot2: noop,
  lot3: "",
  setLot3: noop,
  lot4: "",
  setLot4: noop,
  nextScheduleType: "1year",
  setNextScheduleType: noop,
  nextDate: "",
  setNextDate: noop,
  remarks: "",
  setRemarks: noop,
};

describe("VaccinationForm responsive layout", () => {
  it("接種基本情報とLOT番号はmobileで全幅1列、sm以上で既存の2列に戻る", () => {
    const { container } = render(
      <MemoryRouter>
        <VaccinationForm {...baseProps} />
      </MemoryRouter>,
    );

    const basicInfoGrid = screen.getByText("予防接種名").closest('[class*="grid-cols"]');
    const lotGrid = screen.getByText("LOT1").closest('[class*="grid-cols"]');
    const formColumn = container.firstElementChild;

    expect(formColumn).toHaveClass("col-span-1", "lg:col-span-3");
    expect(formColumn).not.toHaveClass("col-span-6");

    for (const grid of [basicInfoGrid, lotGrid]) {
      expect(grid).toHaveClass("w-full", "grid-cols-1", "sm:grid-cols-2");
      expect(grid).not.toHaveClass("grid-cols-2");
    }
  });
});

describe("VaccinationForm BUG-015 field errors", () => {
  it("fieldErrors.vaccineId / date を role=alert で表示する", () => {
    render(
      <MemoryRouter>
        <VaccinationForm
          {...baseProps}
          fieldErrors={{
            vaccineId: "ワクチン種別を選択してください",
            date: "接種日を入力してください",
          }}
        />
      </MemoryRouter>,
    );

    const alerts = screen.getAllByRole("alert");
    expect(alerts.map((el) => el.textContent)).toEqual([
      "ワクチン種別を選択してください",
      "接種日を入力してください",
    ]);
  });
});
