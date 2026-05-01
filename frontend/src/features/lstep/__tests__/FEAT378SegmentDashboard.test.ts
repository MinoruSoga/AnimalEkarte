import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

/**
 * FEAT-378: CPM/LTV/休眠タグ ダッシュボード
 *
 * ソースコード検査による静的解析テスト。
 * 実レンダリングコストを避けつつ、仕様漏れを確実に検知する粒度で記述する。
 */

const cwd = process.cwd();

const segmentDashboardSource = readFileSync(
  resolve(cwd, 'src/features/lstep/components/SegmentDashboard.tsx'),
  'utf-8'
);

const segmentTagsSource = readFileSync(
  resolve(cwd, 'src/features/lstep/constants/segment-tags.ts'),
  'utf-8'
);

const autoTagPrefixesSource = readFileSync(
  resolve(cwd, 'src/constants/lstep-auto-tag-prefixes.ts'),
  'utf-8'
);

const drawerSource = readFileSync(
  resolve(cwd, 'src/features/lstep/components/TagOwnerListDrawer.tsx'),
  'utf-8'
);

const pageSource = readFileSync(
  resolve(cwd, 'src/features/lstep/pages/LstepTagManagementPage.tsx'),
  'utf-8'
);

const tagSummaryTableSource = readFileSync(
  resolve(cwd, 'src/features/lstep/components/TagSummaryTable.tsx'),
  'utf-8'
);

// -----------------------------------------------------------------------
// タグ定義
// -----------------------------------------------------------------------
describe('FEAT-378 SegmentDashboard — CPM タグ定義', () => {
  it('CPM 5タグがすべて定義されている', () => {
    expect(segmentTagsSource).toContain('CPM_01_出会い');
    expect(segmentTagsSource).toContain('CPM_02_これから');
    expect(segmentTagsSource).toContain('CPM_03_コア');
    expect(segmentTagsSource).toContain('CPM_04_スポット');
    expect(segmentTagsSource).toContain('CPM_05_ノア');
  });

  it('CPM_SEGMENT_TAGS が定数ファイルから export されている（テスト・他コンポーネントから参照可能）', () => {
    expect(segmentTagsSource).toContain('export const CPM_SEGMENT_TAGS');
  });
});

describe('FEAT-378 SegmentDashboard — LTV タグ定義', () => {
  it('LTV_上位20 が定義されている', () => {
    expect(segmentTagsSource).toContain('LTV_上位20');
  });

  it('LTV_SEGMENT_TAGS が export されている', () => {
    expect(segmentTagsSource).toContain('export const LTV_SEGMENT_TAGS');
  });
});

describe('FEAT-378 SegmentDashboard — VISIT タグ定義', () => {
  it('VISIT 4タグがすべて定義されている', () => {
    expect(segmentTagsSource).toContain('VISIT_120日超');
    expect(segmentTagsSource).toContain('VISIT_180日超');
    expect(segmentTagsSource).toContain('VISIT_220日超');
    expect(segmentTagsSource).toContain('VISIT_240日超');
  });

  it('VISIT_SEGMENT_TAGS が export されている', () => {
    expect(segmentTagsSource).toContain('export const VISIT_SEGMENT_TAGS');
  });
});

// -----------------------------------------------------------------------
// 件数集計ロジック
// -----------------------------------------------------------------------
describe('FEAT-378 SegmentDashboard — 件数集計', () => {
  it('countMap から tagName 完全一致で件数を取得し、該当なし時は 0 fallback する', () => {
    // countMap.get(seg.tagName) ?? 0 のパターンを確認
    expect(segmentDashboardSource).toContain('countMap.get(seg.tagName) ?? 0');
  });

  it('summaryTags から Map を構築している', () => {
    expect(segmentDashboardSource).toContain('new Map<string, number>()');
    expect(segmentDashboardSource).toContain('map.set(tag.tag_name, tag.owner_count)');
  });
});

// -----------------------------------------------------------------------
// ドロワー連携
// -----------------------------------------------------------------------
describe('FEAT-378 SegmentDashboard — ドロワー連携', () => {
  it('セグメントカードクリックで onViewOwners に tagName と count が渡る', () => {
    expect(segmentDashboardSource).toContain('onViewOwners(seg.tagName, count)');
  });
});

// -----------------------------------------------------------------------
// 自動管理タグ判定
// -----------------------------------------------------------------------
describe('FEAT-378 自動管理タグ — 新仕様プレフィックス', () => {
  it('CPM_ が自動管理タグエントリに含まれる', () => {
    expect(autoTagPrefixesSource).toContain('"CPM_"');
  });

  it('LTV_ が自動管理タグエントリに含まれる（新仕様大文字）', () => {
    expect(autoTagPrefixesSource).toContain('"LTV_"');
  });

  it('VISIT_ が自動管理タグエントリに含まれる', () => {
    expect(autoTagPrefixesSource).toContain('"VISIT_"');
  });

  it('PET_ が自動管理タグエントリに含まれる', () => {
    expect(autoTagPrefixesSource).toContain('"PET_"');
  });

  it('EXCL_ が自動管理タグエントリに含まれる', () => {
    expect(autoTagPrefixesSource).toContain('"EXCL_"');
  });

  it('既存の旧タグプレフィックスが維持されている', () => {
    expect(autoTagPrefixesSource).toContain('"cpm_"');
    expect(autoTagPrefixesSource).toContain('"dormant_"');
  });
});

// -----------------------------------------------------------------------
// TagSummaryTable — 自動タグの一括解除ボタン非表示
// -----------------------------------------------------------------------
describe('FEAT-378 TagSummaryTable — 自動タグは削除ボタンが非表示', () => {
  it('isAuto=true のタグには一括解除ボタンが表示されない (!isAuto && canDelete 条件)', () => {
    expect(tagSummaryTableSource).toContain('!isAuto && canDelete');
  });

  it('TagOwnerListDrawer 内でも自動管理タグは一括解除できない', () => {
    expect(drawerSource).toContain('isAutoManagedTag(tagName)');
    expect(drawerSource).toContain('canBulkDelete');
    expect(drawerSource).toContain('canDelete && !isAutoManagedTag(tagName)');
  });
});

// -----------------------------------------------------------------------
// CSV 改善
// -----------------------------------------------------------------------
describe('FEAT-378 TagOwnerListDrawer — CSV 改善', () => {
  it('buildOwnersCsv のヘッダーに tag_name が含まれる', () => {
    expect(drawerSource).toContain('owner_id,owner_name,tag_name,last_visit_date');
  });

  it('CSV 各行に tagName が埋め込まれる', () => {
    expect(drawerSource).toContain('buildOwnersCsv(allOwners, tagName)');
    expect(drawerSource).toContain('"${tagName}"');
  });

  it('CSV export 時に per_page: 100 のページングで全件取得する', () => {
    expect(drawerSource).toContain('const perPage = 100');
    expect(drawerSource).toContain('Math.ceil(ownerCount / perPage)');
    expect(drawerSource).toContain('page, per_page: perPage');
    expect(drawerSource).toContain('owners.push(...data.owners)');
  });

  it('ownerCount が 0 の場合は CSV ボタンが disabled になる', () => {
    expect(drawerSource).toContain('ownerCount === 0');
  });

  it('canExportCsv が false の場合は CSV ボタンが表示されない', () => {
    expect(drawerSource).toContain('canExportCsv ?');
    expect(drawerSource).toContain('canExportCsv = false');
  });
});

// -----------------------------------------------------------------------
// LstepTagManagementPage — SegmentDashboard 組み込み
// -----------------------------------------------------------------------
describe('FEAT-378 LstepTagManagementPage — SegmentDashboard 組み込み', () => {
  it('SegmentDashboard をインポートしている', () => {
    expect(pageSource).toContain('import { SegmentDashboard }');
  });

  it('SegmentDashboard が JSX に含まれる', () => {
    expect(pageSource).toContain('<SegmentDashboard');
  });

  it('TagOwnerListDrawer に canExportCsv が渡される', () => {
    expect(pageSource).toContain('canExportCsv={canView || canEdit}');
  });

  it('canView が usePermission から取得されている', () => {
    expect(pageSource).toContain('canView');
    expect(pageSource).toContain('usePermission(ResourceOwners)');
  });
});
