import { describe, expect, it } from "vitest";
import {
  getResourceForCardKey,
  GROUP_CARD_CONFIG,
  MASTER_SECTIONS,
} from "./master-settings-index-model";
import { CATEGORY_CONFIG } from "../constants/category-config";

describe("master-settings-index-model campaigns entry (V04-A01)", () => {
  it("master cards use the same public names as SidebarMenu", () => {
    expect(GROUP_CARD_CONFIG.clinic.label).toBe("医院マスタ");
    expect(GROUP_CARD_CONFIG.treatmentItems.label).toBe("診療項目マスタ");
    expect(GROUP_CARD_CONFIG.inquiry_template.label).toBe("問診テンプレートマスタ");
    expect(GROUP_CARD_CONFIG.paymentMethods.label).toBe("支払方法マスタ");
    expect(GROUP_CARD_CONFIG.campaigns.label).toBe("割引キャンペーンマスタ");
    expect(CATEGORY_CONFIG.animal_species.label).toBe("動物種類マスタ");
    expect(CATEGORY_CONFIG.reservationType.label).toBe("予約区分マスタ");
    expect(CATEGORY_CONFIG.staff.label).toBe("スタッフマスタ");
    expect(CATEGORY_CONFIG.permission_group.label).toBe("権限グループマスタ");
    expect(CATEGORY_CONFIG.merchandise_item.label).toBe("物販・フードマスタ");
    expect(CATEGORY_CONFIG.chief_complaint.label).toBe("主訴カテゴリマスタ");
  });

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

  it("バグ報告カードはサポート節にあり権限リソースを持たない（全スタッフ・全医院公開）", () => {
    // バグ報告は全医院共有の製品フィードバック基盤として意図的に権限ゲートなし
    expect(GROUP_CARD_CONFIG.bugReports).toMatchObject({
      label: "バグ報告",
      path: "/settings/bug-reports",
    });
    expect(GROUP_CARD_CONFIG.bugReports.resource).toBeUndefined();
    expect(getResourceForCardKey("bugReports")).toBeUndefined();
    const support = MASTER_SECTIONS.find((s) => s.title === "サポート");
    expect(support?.keys).toEqual(expect.arrayContaining(["bugReports"]));
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
