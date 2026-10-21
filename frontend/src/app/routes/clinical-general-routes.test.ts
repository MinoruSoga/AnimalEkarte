import { isValidElement } from "react";
import { describe, expect, it } from "vitest";

import { RequirePermission } from "@/components/shared/RequirePermission";
import { ResourceOwners } from "@/types/generated/models";

import { clinicalGeneralRoutes } from "./clinical-general-routes";

describe("BUG-20260906-003 aggregation route", () => {
  it("/aggregation は element と lazy Component を同一 route に置かない", () => {
    const route = clinicalGeneralRoutes.find((entry) => entry.path === "/aggregation");
    expect(route).toBeDefined();
    expect(route?.lazy).toBeUndefined();
    expect(route?.element).toBeTruthy();
    expect(route?.children?.[0]?.index).toBe(true);
    expect(route?.children?.[0]?.lazy).toEqual(expect.any(Function));
  });
});

describe("EMR-463 /owners route structure", () => {
  const ownersRoute = clinicalGeneralRoutes.find((entry) => entry.path === "/owners");
  const ownersChildren = ownersRoute?.children ?? [];

  it("親 route は選択医院 view wrapper を持たず errorElement を維持する", () => {
    expect(ownersRoute).toBeDefined();
    expect(ownersRoute?.element).toBeUndefined();
    expect(ownersRoute?.errorElement).toBeTruthy();
  });

  it("一覧 index は親 guard なしの lazy leaf で、guard は lazy 内の Component が担う", () => {
    const indexRoute = ownersChildren.find((child) => child.index === true);
    expect(indexRoute).toBeDefined();
    expect(indexRoute?.lazy).toEqual(expect.any(Function));
    // BUG-20260906-003: 同一 route に element と lazy Component を置かない
    expect(indexRoute?.element).toBeUndefined();
  });

  it("new は owners:view の内側に owners:create を要求する", () => {
    const newRoute = ownersChildren.find((child) => child.path === "new");
    expect(newRoute).toBeDefined();
    expect(newRoute?.lazy).toBeUndefined();
    const viewGuard = newRoute?.element;
    if (!isValidElement(viewGuard)) throw new Error("new route element is missing");
    expect(viewGuard.type).toBe(RequirePermission);
    expect(viewGuard.props.resource).toBe(ResourceOwners);
    const createGuard = viewGuard.props.children;
    if (!isValidElement(createGuard)) throw new Error("new route create guard is missing");
    expect(createGuard.type).toBe(RequirePermission);
    expect(createGuard.props.resource).toBe(ResourceOwners);
    expect(createGuard.props.action).toBe("create");
    const leaf = newRoute?.children?.[0];
    expect(leaf?.index).toBe(true);
    expect(leaf?.lazy).toEqual(expect.any(Function));
  });

  it(":id は owners:view の element と lazy index child を持つ", () => {
    const idRoute = ownersChildren.find((child) => child.path === ":id");
    expect(idRoute).toBeDefined();
    expect(idRoute?.lazy).toBeUndefined();
    const viewGuard = idRoute?.element;
    if (!isValidElement(viewGuard)) throw new Error(":id route element is missing");
    expect(viewGuard.type).toBe(RequirePermission);
    expect(viewGuard.props.resource).toBe(ResourceOwners);
    const leaf = idRoute?.children?.[0];
    expect(leaf?.index).toBe(true);
    expect(leaf?.lazy).toEqual(expect.any(Function));
    expect(leaf?.element).toBeUndefined();
  });
});
