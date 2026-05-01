export interface SegmentTag {
  label: string;
  tagName: string;
}

export interface SegmentGroup {
  title: string;
  tags: SegmentTag[];
  gridCols: string;
}

export const CPM_SEGMENT_TAGS: SegmentTag[] = [
  { label: "CPM1 出会い", tagName: "CPM_01_出会い" },
  { label: "CPM2 これから", tagName: "CPM_02_これから" },
  { label: "CPM3 コア", tagName: "CPM_03_コア" },
  { label: "CPM4 スポット", tagName: "CPM_04_スポット" },
  { label: "CPM5 ノア", tagName: "CPM_05_ノア" },
];

export const LTV_SEGMENT_TAGS: SegmentTag[] = [
  { label: "LTV 上位20%", tagName: "LTV_上位20" },
];

export const VISIT_SEGMENT_TAGS: SegmentTag[] = [
  { label: "120日超", tagName: "VISIT_120日超" },
  { label: "180日超", tagName: "VISIT_180日超" },
  { label: "220日超", tagName: "VISIT_220日超" },
  { label: "240日超", tagName: "VISIT_240日超" },
];

export const SEGMENT_GROUPS: SegmentGroup[] = [
  {
    title: "CPM ステージ",
    tags: CPM_SEGMENT_TAGS,
    gridCols: "grid-cols-2 sm:grid-cols-3 lg:grid-cols-5",
  },
  {
    title: "LTV",
    tags: LTV_SEGMENT_TAGS,
    gridCols: "grid-cols-1",
  },
  {
    title: "休眠予備軍",
    tags: VISIT_SEGMENT_TAGS,
    gridCols: "grid-cols-2 sm:grid-cols-4",
  },
];
