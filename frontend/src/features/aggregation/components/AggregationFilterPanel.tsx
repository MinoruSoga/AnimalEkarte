import { C } from "@/lib/design-tokens";

import type { AggregationParams } from "../api/get-aggregations";
import type { AggregationTab } from "./aggregation-filter-panel-model";
import {
  AggregationSearchFilter,
  CPMStageFilter,
  LastVisitFilters,
  RevenueFilters,
  SortControls,
  VisitFilters,
} from "./AggregationFilterPanelParts";

interface AggregationFilterPanelProps {
  params: AggregationParams;
  onParamsChange: (params: Partial<AggregationParams>) => void;
  activeTab: AggregationTab;
}

export function AggregationFilterPanel({ params, onParamsChange, activeTab }: AggregationFilterPanelProps) {
  const inputClass = `h-11 ${C.borderMedium} ${C.text} bg-white text-base`;
  const labelClass = `text-xs ${C.text50}`;

  return (
    <div className="flex flex-wrap items-end gap-x-3 gap-y-2">
      <AggregationSearchFilter params={params} onParamsChange={onParamsChange} />

      {activeTab === "revenue" ? (
        <RevenueFilters
          params={params}
          inputClass={inputClass}
          labelClass={labelClass}
          onParamsChange={onParamsChange}
        />
      ) : null}

      {activeTab === "visit" ? (
        <VisitFilters
          params={params}
          inputClass={inputClass}
          labelClass={labelClass}
          onParamsChange={onParamsChange}
        />
      ) : null}

      {activeTab === "last_visit" ? (
        <LastVisitFilters
          params={params}
          inputClass={inputClass}
          labelClass={labelClass}
          onParamsChange={onParamsChange}
        />
      ) : null}

      <CPMStageFilter
        params={params}
        inputClass={inputClass}
        labelClass={labelClass}
        onParamsChange={onParamsChange}
      />

      <SortControls
        params={params}
        activeTab={activeTab}
        inputClass={inputClass}
        labelClass={labelClass}
        onParamsChange={onParamsChange}
      />
    </div>
  );
}
