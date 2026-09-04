export const MEDICAL_RECORD_EXAM_ID_PARAM = "examId";

export function isTargetExamGroup(groupId: number, examId: string | null): boolean {
  return examId != null && examId !== "" && String(groupId) === examId;
}

export function orderExamGroupsForTarget<T extends { id: number }>(
  groups: T[],
  examId: string | null,
): T[] {
  if (!examId) return groups;
  const target = groups.filter((group) => String(group.id) === examId);
  const rest = groups.filter((group) => String(group.id) !== examId);
  return [...target, ...rest];
}
