import type { RouteObject } from "react-router";

import { clinicalBusinessRoutes } from "./clinical-business-routes";
import { clinicalCareRoutes } from "./clinical-care-routes";
import { clinicalGeneralRoutes } from "./clinical-general-routes";

export const clinicalRoutes: RouteObject[] = [
  ...clinicalGeneralRoutes,
  ...clinicalCareRoutes,
  ...clinicalBusinessRoutes,
];
