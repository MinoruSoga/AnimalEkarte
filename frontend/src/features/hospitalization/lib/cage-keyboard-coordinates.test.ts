import { describe, expect, it } from "vitest";
import { selectNextCageRect, type CageKeyboardRect } from "./cage-keyboard-coordinates";

const cage = (left: number, top: number): CageKeyboardRect => ({
  left,
  top,
  right: left + 100,
  bottom: top + 80,
});

describe("selectNextCageRect (BUG-005)", () => {
  const left = cage(0, 0);
  const right = cage(200, 0);
  const below = cage(0, 160);

  it("ArrowRight は右隣のケージを選ぶ", () => {
    expect(selectNextCageRect("ArrowRight", left, [left, right, below])).toEqual(right);
  });

  it("ArrowDown は下のケージを選ぶ", () => {
    expect(selectNextCageRect("ArrowDown", left, [left, right, below])).toEqual(below);
  });

  it("進行方向に候補が無ければ undefined", () => {
    expect(selectNextCageRect("ArrowLeft", left, [left, right])).toBeUndefined();
  });

  it("衝突矩形が無いときは先頭候補を返す", () => {
    expect(selectNextCageRect("ArrowRight", null, [right, below])).toEqual(right);
  });
});
