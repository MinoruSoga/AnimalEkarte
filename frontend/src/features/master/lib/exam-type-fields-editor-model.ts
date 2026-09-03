import type { ExamReferenceRange, ExamReferenceRangeInput } from "../api/exam-types-master";

export const QUALITATIVE_VALUES = ["(-)", "(±)", "(+)", "(++)", "(+++)"] as const;

type ReferenceRangeMode = "numeric" | "qualitative";

export interface ReferenceRangeDraft {
  animalSpeciesId: string;
  mode: ReferenceRangeMode;
  min: string;
  max: string;
  qualitativeMin?: string;
  qualitativeMax?: string;
}

function normalizeQualitativeValue(value: string): string {
  return value.replaceAll(/\s/g, "").replaceAll("（", "(").replaceAll("）", ")");
}

function qualitativeIndex(value: string): number {
  return QUALITATIVE_VALUES.indexOf(
    normalizeQualitativeValue(value) as (typeof QUALITATIVE_VALUES)[number],
  );
}

export function validateReferenceRangeDrafts(drafts: ReferenceRangeDraft[]): string | null {
  const species = new Set<string>();
  for (const draft of drafts) {
    if (species.has(draft.animalSpeciesId)) {
      return "同じ動物種の基準範囲は複数登録できません";
    }
    species.add(draft.animalSpeciesId);

    const hasNumeric =
      draft.mode === "numeric" && (draft.min.trim() !== "" || draft.max.trim() !== "");
    const rawQualitativeMin =
      draft.mode === "qualitative" ? draft.min : (draft.qualitativeMin ?? "");
    const rawQualitativeMax =
      draft.mode === "qualitative" ? draft.max : (draft.qualitativeMax ?? "");
    const hasQualitative = rawQualitativeMin.trim() !== "" || rawQualitativeMax.trim() !== "";

    if (hasNumeric && hasQualitative) {
      return "数値範囲と定性範囲は同時に指定できません";
    }

    if (hasNumeric) {
      const min = draft.min.trim() === "" ? undefined : Number(draft.min);
      const max = draft.max.trim() === "" ? undefined : Number(draft.max);
      if (
        (min !== undefined && !Number.isFinite(min)) ||
        (max !== undefined && !Number.isFinite(max))
      ) {
        return "数値範囲には有限の数値を入力してください";
      }
      if (min !== undefined && max !== undefined && min > max) {
        return "数値範囲の下限は上限以下にしてください";
      }
    }

    if (hasQualitative) {
      const minIndex =
        rawQualitativeMin.trim() === "" ? undefined : qualitativeIndex(rawQualitativeMin);
      const maxIndex =
        rawQualitativeMax.trim() === "" ? undefined : qualitativeIndex(rawQualitativeMax);
      if (minIndex === -1 || maxIndex === -1) {
        return "定性値は (-)、(±)、(+)、(++)、(+++) から選択してください";
      }
      if (minIndex !== undefined && maxIndex !== undefined && minIndex > maxIndex) {
        return "定性範囲の下限は上限以下にしてください";
      }
    }
  }
  return null;
}

export function buildReferenceRangeRequest(
  drafts: ReferenceRangeDraft[],
): ExamReferenceRangeInput[] {
  return drafts.map((draft) => {
    const base = { animal_species_id: Number(draft.animalSpeciesId) };
    if (draft.mode === "numeric") {
      return {
        ...base,
        ...(draft.min.trim() === "" ? {} : { ref_min: Number(draft.min) }),
        ...(draft.max.trim() === "" ? {} : { ref_max: Number(draft.max) }),
      };
    }
    return {
      ...base,
      ...(draft.min.trim() === "" ? {} : { qualitative_min: normalizeQualitativeValue(draft.min) }),
      ...(draft.max.trim() === "" ? {} : { qualitative_max: normalizeQualitativeValue(draft.max) }),
    };
  });
}

export function toReferenceRangeDraft(range: ExamReferenceRange): ReferenceRangeDraft {
  const qualitative = range.qualitativeMin !== undefined || range.qualitativeMax !== undefined;
  return {
    animalSpeciesId: range.animalSpeciesId,
    mode: qualitative ? "qualitative" : "numeric",
    min: qualitative ? (range.qualitativeMin ?? "") : String(range.refMin ?? ""),
    max: qualitative ? (range.qualitativeMax ?? "") : String(range.refMax ?? ""),
  };
}
