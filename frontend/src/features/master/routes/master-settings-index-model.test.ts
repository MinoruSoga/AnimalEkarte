import { describe, expect, it } from "vitest";
import {
  getResourceForCardKey,
  GROUP_CARD_CONFIG,
  MASTER_SECTIONS,
} from "./master-settings-index-model";

describe("master-settings-index-model campaigns entry (V04-A01)", () => {
  it("campaigns カードが GROUP_CARD_CONFIG にあり /settings/campaigns と accounting resource を持つ", () => {
    // ResourceAccounting 定数 import は generated-model allowlist を増やすため文字列で固定
    expect(GROUP_CARD_CONFIG.campaigns).toMatchObject({
      label: "割引キャンペーンマスタ",
      path: "/settings/campaigns",
      resource: "accounting",
    });
    expect(getResourceForCardKey("campaigns")).toBe("accounting");
  });

  it("検査マスタカードがカルテ節にあり診療項目の検査タブへ行く", () => {
    expect(GROUP_CARD_CONFIG.examinationItems).toMatchObject({
      label: "検査マスタ",
      path: "/settings/treatment-items?tab=examination",
      resource: "master-medical",
    });
    expect(getResourceForCardKey("examinationItems")).toBe("master-medical");
    const chart = MASTER_SECTIONS.find((s) => s.title === "カルテ");
    expect(chart?.keys).toEqual(
      expect.arrayContaining(["examinationItems", "labDeviceItemMasters"]),
    );
    expect(chart!.keys.indexOf("examinationItems")).toBeLessThan(
      chart!.keys.indexOf("labDeviceItemMasters"),
    );
  });

  it("検査機器マスタカードがカルテ節にあり lab-import を要求する", () => {
    expect(GROUP_CARD_CONFIG.labDeviceItemMasters).toMatchObject({
      label: "検査機器マスタ",
      path: "/settings/lab-device-item-masters",
      resource: "lab-import",
    });
    expect(getResourceForCardKey("labDeviceItemMasters")).toBe("lab-import");
    const chart = MASTER_SECTIONS.find((s) => s.title === "カルテ");
    expect(chart?.keys).toEqual(expect.arrayContaining(["labDeviceItemMasters"]));
  });

  it("会計・商品セクションに paymentMethods と同列で campaigns が並ぶ", () => {
    const accounting = MASTER_SECTIONS.find((s) => s.title === "会計・商品");
    expect(accounting).toBeDefined();
    expect(accounting?.keys).toEqual(
      expect.arrayContaining(["paymentMethods", "campaigns", "closingTime"]),
    );
    const paymentIdx = accounting!.keys.indexOf("paymentMethods");
    const campaignsIdx = accounting!.keys.indexOf("campaigns");
    expect(campaignsIdx).toBeGreaterThan(paymentIdx);
  });
});
