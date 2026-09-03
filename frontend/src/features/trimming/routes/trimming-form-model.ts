import type { MasterItem } from "@/types";
import { filterActiveOrSelectedMasterItems } from "../hooks/trimming-form-utils";

export const TRIMMING_PRIORITY_FIELDS = ["staffId", "courseId"] as const;
export const TRIMMING_FORM_ID = "trimming-form";

export type TrimmingFormGate =
  | { kind: "new-pet-loading" }
  | { kind: "new-no-pet" }
  | { kind: "loading" }
  | { kind: "not-found" };

export function resolveTrimmingFormGate(input: {
  hasSelectedPet: boolean;
  mode: "new" | "edit";
  petId: string | null;
  isLoading: boolean;
  notFound: boolean;
}): TrimmingFormGate | null {
  if (!input.hasSelectedPet && input.mode === "new" && input.petId)
    return { kind: "new-pet-loading" };
  if (!input.hasSelectedPet && input.mode === "new") return { kind: "new-no-pet" };
  if (input.isLoading) return { kind: "loading" };
  if (input.notFound) return { kind: "not-found" };
  return null;
}

export type TrimmingSelectableItem = Omit<MasterItem, "id"> & { id: string };

export function decorateTrimmingCourses(
  coursesRaw: MasterItem[],
  courseTypeNameById: ReadonlyMap<string, string>,
  selectedCourseId: string,
): TrimmingSelectableItem[] {
  const named: TrimmingSelectableItem[] = coursesRaw.map((course) => {
    const typeName = course.courseTypeId ? courseTypeNameById.get(course.courseTypeId) : undefined;
    return {
      ...course,
      id: String(course.id),
      name: typeName ? `[${typeName}] ${course.name}` : course.name,
    };
  });
  return filterActiveOrSelectedMasterItems(named, selectedCourseId ? [selectedCourseId] : []);
}

export function decorateTrimmingOptions(
  optionsRaw: MasterItem[],
  selectedOptionIds: string[],
): TrimmingSelectableItem[] {
  const named: TrimmingSelectableItem[] = optionsRaw.map((option) => ({
    ...option,
    id: String(option.id),
  }));
  return filterActiveOrSelectedMasterItems(named, selectedOptionIds);
}
