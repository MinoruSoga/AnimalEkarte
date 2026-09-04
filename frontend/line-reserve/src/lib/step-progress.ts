// SD-16: トリミング分岐の有無によらず、フロー全体で一貫した ProgressDots の current/total を算出する。
// 分岐前後でステップ順序が変わるため、フロー種別ごとに全ステップの並びを明示し、
// 各ページはこの並びの中の自分の位置だけを指定すればよいようにする。

export type FlowStepKey =
  | "customerInfo"
  | "courseSelect"
  | "trimmingCourseSelect"
  | "trimmingOptionSelect"
  | "staffSelect"
  | "dateSelect"
  | "timeSelect"
  | "request"
  | "confirm";

// 完了ページ（CompletePage）は ProgressDots を表示しないが、進捗の母数には含める。
const GENERAL_FLOW_STEPS: FlowStepKey[] = [
  "customerInfo",
  "courseSelect",
  "staffSelect",
  "dateSelect",
  "timeSelect",
  "request",
  "confirm",
];

const TRIMMING_FLOW_STEPS: FlowStepKey[] = [
  "customerInfo",
  "courseSelect",
  "trimmingCourseSelect",
  "trimmingOptionSelect",
  "staffSelect",
  "dateSelect",
  "timeSelect",
  "request",
  "confirm",
];

const COMPLETE_STEP_COUNT = 1;

export interface StepProgress {
  current: number;
  total: number;
}

export function getStepProgress(step: FlowStepKey, isTrimming: boolean): StepProgress {
  const order = isTrimming ? TRIMMING_FLOW_STEPS : GENERAL_FLOW_STEPS;
  const index = order.indexOf(step);

  if (index === -1) {
    throw new Error(`getStepProgress: unknown step "${step}" for isTrimming=${isTrimming}`);
  }

  return {
    current: index + 1,
    total: order.length + COMPLETE_STEP_COUNT,
  };
}
