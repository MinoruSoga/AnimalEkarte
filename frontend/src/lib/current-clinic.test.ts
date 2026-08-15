import { afterEach, describe, expect, it } from "vitest";

import {
  CURRENT_CLINIC_STORAGE_KEY,
  getStoredClinicId,
  normalizeClinicId,
  requireStoredClinicId,
  setStoredClinicId,
} from "./current-clinic";

afterEach(() => {
  localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
});

describe("current-clinic storage helpers", () => {
  it("空文字や空白だけの clinic_id を null として扱う", () => {
    expect(normalizeClinicId(null)).toBeNull();
    expect(normalizeClinicId("")).toBeNull();
    expect(normalizeClinicId("   ")).toBeNull();
    expect(normalizeClinicId(" 12 ")).toBe("12");
  });

  it("localStorage の空値を null に正規化する", () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "");
    expect(getStoredClinicId()).toBeNull();

    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, " 3 ");
    expect(getStoredClinicId()).toBe("3");
  });

  it("setStoredClinicId は trim して保存し空値は拒否する", () => {
    expect(setStoredClinicId("  5  ")).toBe(true);
    expect(getStoredClinicId()).toBe("5");
    expect(setStoredClinicId("   ")).toBe(false);
  });

  it("clinic_id が未選択なら requireStoredClinicId は例外を投げる", () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "");
    expect(() => requireStoredClinicId()).toThrow("クリニックが選択されていません");
  });
});
