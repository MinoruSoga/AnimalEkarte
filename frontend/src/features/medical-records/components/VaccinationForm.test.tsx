import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import { VaccinationForm } from "./VaccinationForm";

const noop = () => {};

describe("VaccinationForm responsive layout", () => {
  it("接種基本情報とLOT番号はmobileで全幅1列、sm以上で既存の2列に戻る", () => {
    const { container } = render(
      <MemoryRouter>
        <VaccinationForm
          vaccineOptions={[]}
          vaccineName=""
          setVaccineName={noop}
          date=""
          setDate={noop}
          supplemental=""
          setSupplemental={noop}
          lot1=""
          setLot1={noop}
          lot2=""
          setLot2={noop}
          lot3=""
          setLot3={noop}
          lot4=""
          setLot4={noop}
          nextScheduleType="1year"
          setNextScheduleType={noop}
          nextDate=""
          setNextDate={noop}
          remarks=""
          setRemarks={noop}
        />
      </MemoryRouter>,
    );

    const basicInfoGrid = screen.getByText("予防接種名").closest('[class*="grid-cols"]');
    const lotGrid = screen.getByText("LOT1").closest('[class*="grid-cols"]');
    const formColumn = container.firstElementChild;

    expect(formColumn).toHaveClass("col-span-1", "lg:col-span-6");
    expect(formColumn).not.toHaveClass("col-span-6");

    for (const grid of [basicInfoGrid, lotGrid]) {
      expect(grid).toHaveClass("w-full", "grid-cols-1", "sm:grid-cols-2");
      expect(grid).not.toHaveClass("grid-cols-2");
    }
  });
});
