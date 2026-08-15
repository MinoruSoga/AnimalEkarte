import { describe, expect, it } from "vitest";

import { paths } from "@/config/paths";

import { sidebarMenuSections } from "./sidebar-menu";

function findLstepSection() {
  const section = sidebarMenuSections.find((s) => s.title === "Lステップ連携");
  expect(section).toBeDefined();
  return section!;
}

function findLstepRootItem() {
  const section = findLstepSection();
  const root = section.items.find((item) => item.label === "Lステップ連携");
  expect(root).toBeDefined();
  expect(root!.subItems).toBeDefined();
  return root!;
}

function findLstepSubItem(label: string) {
  const root = findLstepRootItem();
  const sub = root.subItems!.find((item) => item.label === label);
  expect(sub, `expected Lステップ subItem "${label}"`).toBeDefined();
  return sub!;
}

describe("Lステップ連携 sidebar nav honesty (R-06 / R-07)", () => {
  it("paths.lstep.deliveryMonitor が /lstep/delivery-monitor を公開する", () => {
    expect(paths.lstep.deliveryMonitor.path).toBe("/lstep/delivery-monitor");
    expect(paths.lstep.deliveryMonitor.getHref()).toBe("/lstep/delivery-monitor");
  });

  it("配信監視 subItem が menu にあり分析レポートと同 resource を要求する (R-06)", () => {
    const deliveryMonitor = findLstepSubItem("配信監視");
    const analytics = findLstepSubItem("分析レポート");

    expect(deliveryMonitor.path).toBe(paths.lstep.deliveryMonitor.getHref());
    expect(deliveryMonitor.path).toBe("/lstep/delivery-monitor");
    // ResourceLstepAnalytics without importing generated/models (boundary allowlist)
    expect(deliveryMonitor.resource).toBe(analytics.resource);
  });

  it("タグ管理 subItem の resource が分析レポートと一致する (R-07)", () => {
    const tags = findLstepSubItem("タグ管理");
    const analytics = findLstepSubItem("分析レポート");

    expect(tags.path).toBe(paths.lstep.tags.getHref());
    expect(tags.resource).toBe(analytics.resource);
  });

  it("分析レポート path が paths.lstep.analytics と一致する (mirror)", () => {
    const analytics = findLstepSubItem("分析レポート");

    expect(analytics.path).toBe(paths.lstep.analytics.getHref());
  });
});
