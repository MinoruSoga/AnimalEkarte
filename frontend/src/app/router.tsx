import { createBrowserRouter } from "react-router";

import { Layout } from "@/components/shared/Layout";
import { RootErrorBoundary } from "@/components/errors/RouteErrorBoundary";

import { dashboardRoutes } from "./routes/dashboard";
import { ownersRoutes } from "./routes/owners";
import { reservationsRoutes } from "./routes/reservations";
import { medicalRecordsRoutes } from "./routes/medical-records";
import { hospitalizationRoutes } from "./routes/hospitalization";
import { trimmingRoutes } from "./routes/trimming";
import { examinationsRoutes } from "./routes/examinations";
import { accountingRoutes } from "./routes/accounting";
import { vaccinationsRoutes } from "./routes/vaccinations";
import { inventoryRoutes } from "./routes/inventory";
import { settingsRoutes } from "./routes/settings";
import { notFoundRoute } from "./routes/not-found";

export const router = createBrowserRouter([
  {
    element: <Layout />,
    errorElement: <RootErrorBoundary />,
    children: [
      ...dashboardRoutes,
      ...ownersRoutes,
      ...reservationsRoutes,
      ...medicalRecordsRoutes,
      ...hospitalizationRoutes,
      ...trimmingRoutes,
      ...examinationsRoutes,
      ...accountingRoutes,
      ...vaccinationsRoutes,
      ...inventoryRoutes,
      ...settingsRoutes,
      notFoundRoute,
    ],
  },
]);
