import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Hospitalization } from "../api/transforms";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";

// EMR-301/EMR-319 レビュー対応: header は HospitalizationResponse の pet サマリ
// (transform 済み petBirthDate/petIsDeceased) を使い、/pets/:id の追加取得は行わない。
// owners:view 権限を持たない入院閲覧者でもバッジが出ることをこの依存排除で担保する。

function makeHospitalization(
  petIsDeceased: boolean,
  overrides: Partial<Hospitalization> = {},
): Hospitalization {
  const base: Hospitalization = {
    id: "42",
    hospitalizationNo: "42",
    ownerName: "山田",
    // Router 外で描画するテストのためリンク生成に使う ownerId は未設定。
    ownerId: undefined,
    petName: "タロウ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-20",
    endDate: "2026-07-25",
    status: "入院中",
    cageId: undefined,
    petId: "10",
    doctorId: undefined,
    doctorName: undefined,
    petIsDeceased,
    petBirthDate: undefined,
    memo: undefined,
    ownerRequest: undefined,
    staffNotes: undefined,
  };
  return { ...base, ...overrides };
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

describe("HospitalizationPatientHeader 担当医表示 (EMR-352)", () => {
  it("doctorName 未設定なら担当医ラベルは1回だけ表示し未設定と識別する", () => {
    render(<HospitalizationPatientHeader hospitalization={makeHospitalization(false)} />);

    // PatientInfoCard が staffLabel="担当医" を付与するため、呼び出し側の
    // fallback 値にラベルを含めると「担当医担当医未設定」と二重表示になる。
    expect(screen.getByText("担当医未設定")).toBeInTheDocument();
    expect(screen.queryByText(/担当医.*担当医/)).not.toBeInTheDocument();
  });

  it("doctorName 設定済みなら担当医ラベル1回と氏名を表示する", () => {
    render(
      <HospitalizationPatientHeader
        hospitalization={makeHospitalization(false, { doctorName: "田中 一郎" })}
      />,
    );

    expect(screen.getByText("担当医田中 一郎")).toBeInTheDocument();
    expect(screen.queryByText(/担当医.*担当医/)).not.toBeInTheDocument();
  });
});

describe("HospitalizationPatientHeader 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは Cake バッジを表示する", () => {
    render(
      <HospitalizationPatientHeader
        hospitalization={makeHospitalization(false, { petBirthDate: "2019-10-02" })}
      />,
    );

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2019-09-02" as const],
    ["birth_date 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    render(
      <HospitalizationPatientHeader
        hospitalization={makeHospitalization(false, { petBirthDate: birthDate })}
      />,
    );

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("petIsDeceased の個体には誕生月でも Cake バッジを表示しない", () => {
    render(
      <HospitalizationPatientHeader
        hospitalization={makeHospitalization(true, { petBirthDate: "2019-10-02" })}
      />,
    );

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
