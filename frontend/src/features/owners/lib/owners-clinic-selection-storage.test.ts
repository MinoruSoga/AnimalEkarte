import { beforeEach, describe, expect, it } from "vitest";
import {
  ownersClinicSelectionStorageKey,
  readOwnersClinicSelection,
  removeOwnersClinicSelection,
  writeOwnersClinicSelection,
} from "./owners-clinic-selection-storage";

// EMR-463: key `owners-list.clinics.v1.<user.id>`、保存内容は医院 ID 文字列配列のみ（合成値）。
describe("owners-clinic-selection-storage", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("key は user.id ごとで、医院 ID を含めない", () => {
    expect(ownersClinicSelectionStorageKey("10")).toBe("owners-list.clinics.v1.10");
  });

  it("保存した ID 配列だけを JSON で書き、そのまま読み戻す", () => {
    expect(writeOwnersClinicSelection("10", ["2", "3"])).toBe(true);
    expect(window.localStorage.getItem("owners-list.clinics.v1.10")).toBe('["2","3"]');
    expect(readOwnersClinicSelection("10")).toEqual(["2", "3"]);
  });

  it("別ユーザーの key は読まない", () => {
    writeOwnersClinicSelection("10", ["2"]);
    expect(readOwnersClinicSelection("11")).toBeNull();
    expect(readOwnersClinicSelection("")).toBeNull();
  });

  it('"9007199254740993" を文字列のまま保存・復元する', () => {
    writeOwnersClinicSelection("10", ["9007199254740993"]);
    expect(readOwnersClinicSelection("10")).toEqual(["9007199254740993"]);
  });

  it("正規形でない ID・重複は保存しない", () => {
    writeOwnersClinicSelection("10", ["1", "01", "0", "1", "x"]);
    expect(readOwnersClinicSelection("10")).toEqual(["1"]);
  });

  it.each([
    ["JSON でない", "1,2"],
    ["配列でない", '{"ids":["1"]}'],
    ["数値 ID", "[1,2]"],
    ["非正規形 ID", '["01"]'],
  ])("保存値が%sなら未保存として扱う", (_label, raw) => {
    window.localStorage.setItem("owners-list.clinics.v1.10", raw);
    expect(readOwnersClinicSelection("10")).toBeNull();
  });

  it("remove は自分の key だけを消す", () => {
    writeOwnersClinicSelection("10", ["1"]);
    writeOwnersClinicSelection("11", ["2"]);
    expect(removeOwnersClinicSelection("10")).toBe(true);
    expect(readOwnersClinicSelection("10")).toBeNull();
    expect(readOwnersClinicSelection("11")).toEqual(["2"]);
  });

  it("storage が例外を投げても throw せず未保存／失敗として扱う", () => {
    const throwing = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("quota");
      },
      removeItem: () => {
        throw new Error("blocked");
      },
    } as unknown as Storage;
    expect(readOwnersClinicSelection("10", throwing)).toBeNull();
    expect(writeOwnersClinicSelection("10", ["1"], throwing)).toBe(false);
    expect(removeOwnersClinicSelection("10", throwing)).toBe(false);
  });

  it("storage が無い環境では未保存として扱う", () => {
    expect(readOwnersClinicSelection("10", null)).toBeNull();
    expect(writeOwnersClinicSelection("10", ["1"], null)).toBe(false);
  });
});
