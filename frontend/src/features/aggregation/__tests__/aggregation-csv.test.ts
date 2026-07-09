import { describe, it, expect } from 'vitest';
import { buildCsvContent } from '../components/aggregation-csv';
import type { AggregationOwner } from '../api/get-aggregations';
import type { AggregationTab } from '../routes/AggregationDashboardPage';

const baseOwner: AggregationOwner = {
  owner_id: 'o1',
  owner_name: 'NoComma太郎',
  total_visit_count: 10,
  annual_visit_count: 5,
  last_visit_date: '2026-04-20',
  first_visit_date: '2024-01-15',
  annual_amount: 150000,
  billing_count: 8,
  period_visit_count: 5,
  days_since_last_visit: 7,
  last_visit_bucket: 'within_3m',
  total_amount: 500000,
  cpm_stage: 'cpm_core',
};

const ALL_TABS: AggregationTab[] = ['revenue', 'visit', 'last_visit'];

describe('buildCsvContent (ISSUE-180)', () => {
  it('places a cpm_stage column right after owner_name on every tab', () => {
    for (const tab of ALL_TABS) {
      const csv = buildCsvContent([baseOwner], tab);
      const header = csv.split('\n')[0];
      expect(header).toContain('cpm_stage');
      expect(header).toMatch(/owner_name,cpm_stage/);
    }
  });

  it('outputs the short CPM label as the cpm_stage value', () => {
    const csv = buildCsvContent([baseOwner], 'revenue');
    const dataRow = csv.split('\n')[1];
    expect(dataRow).toContain('Core');
  });

  it('outputs an empty cpm_stage cell when the stage is missing', () => {
    const noCpm: AggregationOwner = { ...baseOwner, cpm_stage: undefined };
    const csv = buildCsvContent([noCpm], 'revenue');
    const [headerLine, dataLine] = csv.split('\n');
    const cpmIndex = headerLine.split(',').indexOf('cpm_stage');
    expect(cpmIndex).toBeGreaterThanOrEqual(0);
    // owner_name にカンマを含めていないため列分割が安全。
    expect(dataLine.split(',')[cpmIndex]).toBe('');
  });

  it('still escapes owner_name with embedded quotes/commas', () => {
    const tricky: AggregationOwner = {
      ...baseOwner,
      owner_name: '田中,"太郎"',
    };
    const csv = buildCsvContent([tricky], 'revenue');
    const dataRow = csv.split('\n')[1];
    expect(dataRow).toContain('"田中,""太郎"""');
  });
});
