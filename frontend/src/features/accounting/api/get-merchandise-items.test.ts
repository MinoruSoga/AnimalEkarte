import { describe, expect, it } from "vitest";

import { parseExactDecimal } from "./get-merchandise-items";

// EMR-530: master の exact decimal 文字列 → 課税行wire（number）変換の境界。
// 無損失に表現できない字句は undefined（fail closed）で、丸めて伝播しない。
describe("parseExactDecimal", () => {
  it("passes through exactly representable decimals", () => {
    expect(parseExactDecimal("0.1")).toBe(0.1);
    expect(parseExactDecimal("0.10")).toBe(0.1);
    expect(parseExactDecimal("0.08")).toBe(0.08);
    expect(parseExactDecimal("0")).toBe(0);
    expect(parseExactDecimal("1")).toBe(1);
  });

  it("keeps small nonzero rates that String() renders in exponential form", () => {
    expect(parseExactDecimal("0.0000001")).toBe(0.0000001);
  });

  it("fails closed for decimals binary64 cannot represent", () => {
    expect(parseExactDecimal("0.10000000000000001")).toBeUndefined();
    expect(parseExactDecimal("0.100000000000000000001")).toBeUndefined();
  });

  it("fails closed for out-of-contract tokens", () => {
    expect(parseExactDecimal("abc")).toBeUndefined();
    expect(parseExactDecimal("-0.1")).toBeUndefined();
    expect(parseExactDecimal("1e-7")).toBeUndefined();
  });
});
