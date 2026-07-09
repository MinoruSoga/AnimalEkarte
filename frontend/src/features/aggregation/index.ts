// Public API for aggregation feature
export { AggregationDashboardPage } from "./routes/AggregationDashboardPage";
export { AggregationFilterPanel } from "./components/AggregationFilterPanel";
export { AggregationOwnerTable } from "./components/AggregationOwnerTable";

export type { AggregationTab } from "./routes/AggregationDashboardPage";
export type {
  AggregationSortField,
  AmountBasis,
  PeriodPreset,
  LastVisitBucket,
  AggregationOwner,
  AggregationParams,
  AggregationResponse,
} from "./api/get-aggregations";
