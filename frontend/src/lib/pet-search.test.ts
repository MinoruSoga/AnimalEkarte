import { describe, it, expect } from "vitest";
import { matchesPetSearch } from "./pet-search";
import type { Pet } from "@/types";

function makePet(overrides: Partial<Pet> = {}): Pet {
  return {
    id: "1",
    ownerId: "10",
    ownerNumber: 10,
    ownerName: "",
    ownerNameKana: undefined,
    petNumber: "",
    name: "",
    petNameKana: undefined,
    species: "",
    breed: "",
    color: "",
    food: "",
    environment: "",
    remarks: "",
    ...overrides,
  } as Pet;
}

describe("matchesPetSearch — カナ正規化検索", () => {
  it("ひらがなでカタカナの ownerNameKana にヒットする (ぴ → ピーター)", () => {
    const pet = makePet({ ownerNameKana: "ピーター" });
    expect(matchesPetSearch(pet, "ぴ")).toBe(true);
  });

  it("カタカナでカタカナの ownerNameKana にヒットする (ピ → ピーター)", () => {
    const pet = makePet({ ownerNameKana: "ピーター" });
    expect(matchesPetSearch(pet, "ピ")).toBe(true);
  });

  it("ひらがなでカタカナの petNameKana にヒットする", () => {
    const pet = makePet({ petNameKana: "ポチ" });
    expect(matchesPetSearch(pet, "ぽ")).toBe(true);
  });

  it("ownerName の部分一致 (漢字・通常)", () => {
    const pet = makePet({ ownerName: "山田花子" });
    expect(matchesPetSearch(pet, "山田")).toBe(true);
  });

  it("pet.name の部分一致", () => {
    const pet = makePet({ name: "ポチ" });
    expect(matchesPetSearch(pet, "ポチ")).toBe(true);
  });

  it("species の部分一致", () => {
    const pet = makePet({ species: "犬" });
    expect(matchesPetSearch(pet, "犬")).toBe(true);
  });

  it("ownerNumber の一致", () => {
    const pet = makePet({ ownerNumber: 123 });
    expect(matchesPetSearch(pet, "123")).toBe(true);
  });

  it("空文字ではすべての行がヒット", () => {
    const pet = makePet({ ownerName: "山田" });
    expect(matchesPetSearch(pet, "")).toBe(true);
  });

  it("関係のない文字列ではヒットしない", () => {
    const pet = makePet({ ownerName: "山田", ownerNameKana: "ヤマダ", name: "ポチ", petNameKana: "ポチ" });
    expect(matchesPetSearch(pet, "ねこ")).toBe(false);
  });

  it("ownerNameKana が未設定でも安全に動作する", () => {
    const pet = makePet({ ownerNameKana: undefined });
    expect(matchesPetSearch(pet, "ぴ")).toBe(false);
  });

  it("petNameKana が未設定でも安全に動作する", () => {
    const pet = makePet({ petNameKana: undefined });
    expect(matchesPetSearch(pet, "ぽ")).toBe(false);
  });
});
