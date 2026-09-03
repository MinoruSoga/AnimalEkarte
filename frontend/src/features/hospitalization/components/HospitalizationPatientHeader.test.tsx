import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { Hospitalization } from "../api/transforms";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";

function makeHospitalization(petIsDeceased: boolean): Hospitalization {
  return {
    id: "42",
    hospitalizationNo: "42",
    ownerName: "山田",
    petName: "タロウ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-20",
    endDate: "2026-07-25",
    status: "入院中",
    petIsDeceased,
  };
}

describe("HospitalizationPatientHeader", () => {
  it("死亡個体では既存のPatientInfoCard死亡マーカーを表示する", () => {
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(true)} />);

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
  });

  it("生存個体では死亡マーカーを表示しない", () => {
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(false)} />);

    expect(screen.queryByText("【死亡】")).not.toBeInTheDocument();
  });
});
