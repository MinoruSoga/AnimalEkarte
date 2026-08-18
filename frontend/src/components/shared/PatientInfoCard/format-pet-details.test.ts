import { describe, expect, it } from "vitest";

import { formatPatientPetDetails } from "./format-pet-details";

describe("formatPatientPetDetails (BUG-006)", () => {
  const asOf = new Date("2026-07-20T12:00:00+09:00");

  it("属性が異なる2頭で固定ダミーにならず正本の年齢・性別・去勢避妊を返す", () => {
    const mame = formatPatientPetDetails(
      {
        birthDate: "2012-12-20",
        gender: "雄",
        neuteredDate: undefined,
      },
      asOf,
    );
    const hana = formatPatientPetDetails(
      {
        birthDate: "2003-11-26",
        gender: "不明",
        neuteredDate: undefined,
      },
      asOf,
    );

    expect(mame).toBe("13歳7ヶ月 / 雄 / 不明");
    expect(hana).toBe("22歳7ヶ月 / 不明 / 不明");
    expect(mame).not.toBe(hana);
    expect(mame).not.toContain("メス");
    expect(mame).not.toContain("避妊済");
    expect(hana).not.toBe("9歳5ヶ月 / メス / 避妊済");
  });

  it("雄で去勢日ありは去勢済、雌で避妊日ありは避妊済", () => {
    expect(
      formatPatientPetDetails(
        { birthDate: "2020-01-01", gender: "雄", neuteredDate: "2021-01-01" },
        asOf,
      ),
    ).toBe("6歳6ヶ月 / 雄 / 去勢済");
    expect(
      formatPatientPetDetails(
        { birthDate: "2020-01-01", gender: "雌", neuteredDate: "2021-01-01" },
        asOf,
      ),
    ).toBe("6歳6ヶ月 / 雌 / 避妊済");
  });

  it("属性欠損は推測せず不明とし、固定値にフォールバックしない", () => {
    expect(formatPatientPetDetails({}, asOf)).toBe("不明 / 不明 / 不明");
    expect(
      formatPatientPetDetails({ birthDate: "not-a-date", gender: "", neuteredDate: "" }, asOf),
    ).toBe("不明 / 不明 / 不明");
    expect(formatPatientPetDetails({}, asOf)).not.toBe("9歳5ヶ月 / メス / 避妊済");
  });

  it("性別不明で去勢避妊日のみある場合は中立な済表現", () => {
    expect(
      formatPatientPetDetails(
        { birthDate: "2018-06-01", gender: "不明", neuteredDate: "2019-01-01" },
        asOf,
      ),
    ).toBe("8歳1ヶ月 / 不明 / 避妊・去勢済");
  });
});
