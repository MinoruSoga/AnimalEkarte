/**
 * FE6-8: Pet 属性ラベル二重定義のドリフトガード
 *
 * owners/types (UI選択肢の配列) と lib/transforms/pet (BE⇔FE変換マップ) は
 * 同じ日本語ラベル群を別の目的（リテラル型の源泉／EN⇔JA変換）で手打ち二重定義している。
 * 目的が異なるため単純統合はできないが、片方だけ変更されると無音で乖離するため、
 * 値の集合が一致することをテストで固定する。
 */
import { describe, expect, test } from "vitest";
import { PET_GENDER_MAP, ACQUISITION_TYPE_MAP, DANGER_LEVEL_MAP } from "@/lib/transforms/pet";
import { PET_GENDER_VALUES, ACQUISITION_TYPE_VALUES, DANGER_LEVEL_VALUES } from "./index";

describe("owners の Pet 属性選択肢と transforms の変換マップの整合性", () => {
  test("owners の Pet 属性選択肢は transforms の変換マップと一致する", () => {
    expect(new Set(PET_GENDER_VALUES)).toEqual(new Set(Object.values(PET_GENDER_MAP)));
    expect(new Set(ACQUISITION_TYPE_VALUES)).toEqual(new Set(Object.values(ACQUISITION_TYPE_MAP)));
    expect(new Set(DANGER_LEVEL_VALUES)).toEqual(new Set(Object.values(DANGER_LEVEL_MAP)));
  });
});
