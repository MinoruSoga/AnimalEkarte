import { describe, it, expect } from "vitest";
import { paths } from "./paths";

describe("paths（FE-RC-080: getHref の動的セグメントは encode する）", () => {
  it("通常の数値/UUID id は encode されても見た目が変わらない（後方互換）", () => {
    expect(paths.owners.detail.getHref(300588)).toBe("/owners/300588");
    expect(paths.owners.detail.getHref("a1b2c3")).toBe("/owners/a1b2c3");
    expect(paths.medicalRecords.detail.getHref(42)).toBe("/medical-records/42");
    expect(paths.estimates.edit.getHref(7)).toBe("/estimates/7/edit");
  });

  it("id に URL 予約文字（'/', '?', '#' 等）を含む値が渡っても、追加のパスセグメントや\nクエリを生成しない", () => {
    expect(paths.owners.detail.getHref("1/../2")).toBe("/owners/1%2F..%2F2");
    expect(paths.owners.detail.report.getHref("1?x=1")).toBe(
      "/owners/1%3Fx%3D1/report",
    );
    expect(paths.hospitalization.detail.getHref("1#frag")).toBe(
      "/hospitalization/1%23frag",
    );
  });

  it("manual.article.getHref は category/slug をそれぞれ独立して encode する", () => {
    expect(paths.manual.article.getHref("screens", "owner-detail")).toBe(
      "/manual/screens/owner-detail",
    );
    expect(paths.manual.article.getHref("a/b", "c/d")).toBe(
      "/manual/a%2Fb/c%2Fd",
    );
  });
});
