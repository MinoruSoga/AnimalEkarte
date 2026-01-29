import type { RouteObject } from "react-router";

export const notFoundRoute: RouteObject = {
  path: "*",
  element: (
    <div className="flex-1 p-5 flex items-center justify-center">
      <p className="text-muted-foreground">ページが見つかりません</p>
    </div>
  ),
};
