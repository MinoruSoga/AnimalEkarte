import { isValidElement } from "react";
import { Navigate, type RouteObject } from "react-router";
import { describe, expect, it } from "vitest";

import { appRoutes } from "./app-routes";

const EXPECTED_REDIRECT_PATHS = [
  "/settings/consultation",
  "/settings/diagnosis-name",
  "/settings/diagnosis-type",
  "/settings/examination",
  "/settings/inquiry-template",
  "/settings/job-title",
  "/settings/procedure",
  "/settings/service-type",
  "/settings/shift-template",
  "/settings/trimming-course",
  "/settings/trimming-option",
  "/settings/vaccine",
];

interface LeafRoute {
  path: string;
  isRedirect: boolean;
}

function joinRoutePath(parentPath: string, childPath: string | undefined, index: boolean): string {
  if (index || childPath === undefined) return parentPath || "/";
  if (childPath === "*") return "*";
  if (childPath.startsWith("/")) return childPath;
  return `${parentPath === "/" ? "" : parentPath}/${childPath}`;
}

function flattenLeafRoutes(routes: RouteObject[], parentPath = ""): LeafRoute[] {
  return routes.flatMap((route) => {
    const path = joinRoutePath(parentPath, route.path, route.index === true);
    if (route.children?.length) return flattenLeafRoutes(route.children, path);
    return [{
      path,
      isRedirect: isValidElement(route.element) && route.element.type === Navigate,
    }];
  });
}

describe("main app route inventory", () => {
  it("86 product pages, 12 redirects, wildcard 1を重複なく維持する", () => {
    const leaves = flattenLeafRoutes(appRoutes);
    const wildcard = leaves.filter((route) => route.path === "*");
    const redirects = leaves.filter((route) => route.isRedirect).map((route) => route.path).toSorted();
    const pages = leaves.filter((route) => route.path !== "*" && !route.isRedirect);

    expect(pages).toHaveLength(86);
    expect(new Set(pages.map((route) => route.path)).size).toBe(86);
    expect(wildcard).toHaveLength(1);
    expect(redirects).toEqual(EXPECTED_REDIRECT_PATHS);
  });
});
