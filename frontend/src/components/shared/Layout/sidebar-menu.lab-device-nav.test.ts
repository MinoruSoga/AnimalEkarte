import { describe, expect, it } from "vitest";

import { paths } from "@/config/paths";

import { sidebarMenuSections } from "./SidebarMenu";

function masterMenu() {
  const system = sidebarMenuSections.find((section) => section.title === "システム設定");
  const master = system?.items.find((item) => item.label === "マスタ設定");
  expect(master?.subItems).toBeDefined();
  return master!;
}

function findMasterChild(label: string) {
  const found = masterMenu().subItems!.find((item) => item.label === label);
  expect(found, `expected マスタ設定 child "${label}"`).toBeDefined();
  return found!;
}

function findNested(parentLabel: string, childLabel: string) {
  const parent = findMasterChild(parentLabel);
  const child = parent.subItems?.find((item) => item.label === childLabel);
  expect(child, `expected ${parentLabel} → ${childLabel}`).toBeDefined();
  return child!;
}

describe("検査マスタ sidebar nav", () => {
  it("マスタ設定のカルテ関連に検査種別タブを出す", () => {
    const exam = findNested("カルテ関連", "検査マスタ");
    expect(exam.path).toBe(paths.settings.examination.getHref());
    expect(exam.resource).toBe("master-medical");
  });
});

describe("検査機器マスタ sidebar nav", () => {
  it("マスタ設定のカルテ関連に lab-import 付きで出す", () => {
    const labDevice = findNested("カルテ関連", "検査機器マスタ");
    expect(labDevice.path).toBe(paths.settings.labDeviceItemMasters.getHref());
    expect(labDevice.resource).toBe("lab-import");
  });
});

describe("マスタ設定 sidebar 漏れ分", () => {
  it("主訴カテゴリをカルテ関連に出す", () => {
    const item = findNested("カルテ関連", "主訴カテゴリ");
    expect(item.path).toBe(paths.settings.interview.chiefComplaint.getHref());
    expect(item.resource).toBe("master-medical");
  });

  it("ケージマスタを入院・ケージ配下に出す", () => {
    const item = findNested("入院・ケージ", "ケージマスタ");
    expect(item.path).toBe(paths.settings.cage.getHref());
    expect(item.resource).toBe("master-hospitalization");
  });

  it("コース種別マスタをトリミング配下に出す", () => {
    const item = findNested("トリミング", "コース種別マスタ");
    expect(item.path).toBe(paths.settings.trimmingCourseType.getHref());
    expect(item.resource).toBe("master-trimming");
  });

  it("割引キャンペーンをマスタ設定に出す", () => {
    const item = findMasterChild("割引キャンペーン");
    expect(item.path).toBe(paths.settings.campaigns.getHref());
    expect(item.resource).toBe("accounting");
  });

  it("シフトテンプレートをマスタ設定に出す", () => {
    const item = findMasterChild("シフトテンプレート");
    expect(item.path).toBe(paths.settings.shiftTemplates.getHref());
    expect(item.resource).toBe("shifts");
  });
});

describe("同一飼主・ペット連携 sidebar nav", () => {
  it("システム設定に identity-links を出す", () => {
    const system = sidebarMenuSections.find((section) => section.title === "システム設定");
    const item = system?.items.find((entry) => entry.label === "同一飼主・ペット連携");
    expect(item?.path).toBe(paths.identityLinks.getHref());
    expect(item?.resource).toBe("identity-links");
  });
});
