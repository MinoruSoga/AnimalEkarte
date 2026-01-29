import type { RouteObject } from "react-router";
import { Dashboard } from "@/features/dashboard/routes";

export const dashboardRoutes: RouteObject[] = [
  { path: "/", element: <Dashboard /> },
];
