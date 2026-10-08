import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Hospitalization } from "../api/transforms";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";

// EMR-301: header は petId から Pet detail を引いて誕生月/死亡を判定する。
// このテストは QueryClientProvider なしで描画するためフックを固定する。
const getPetResult = vi.hoisted(() => ({
  current: {
    data: undefined as
      { birthDate?: string; status?: string; deceasedAt?: string | null } | undefined,
  },
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => getPetResult.current,
}));

function makeHospitalization(
  petIsDeceased: boolean,
  overrides: Partial<Hospitalization> = {},
): Hospitalization {
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
    petId: "10",
    ...overrides,
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

describe("HospitalizationPatientHeader 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
    getPetResult.current = { data: undefined };
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは Cake バッジを表示する", () => {
    getPetResult.current = { data: { birthDate: "2019-10-02", status: "生存" } };
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(false)} />);

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2019-09-02" as const],
    ["birth_date 未登録", undefined],
    ["Pet 未取得", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    getPetResult.current = { data: birthDate ? { birthDate, status: "生存" } : undefined };
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(false)} />);

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("wire petIsDeceased の個体には誕生月でも Cake バッジを表示しない", () => {
    getPetResult.current = { data: { birthDate: "2019-10-02", status: "生存" } };
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(true)} />);

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("取得 Pet が死亡判定（status=死亡）なら誕生月でも Cake バッジを表示しない", () => {
    getPetResult.current = { data: { birthDate: "2019-10-02", status: "死亡" } };
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(false)} />);

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
