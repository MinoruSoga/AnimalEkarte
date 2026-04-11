import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Examination } from "@/types/generated/models";

export interface ExamGroupItem {
  id: number;
  name: string;
  result: string;
  unit: string;
  reference_value: string;
  status: string;
}

export interface ExamGroup {
  id: number;
  date: string;
  machine: string;
  items: ExamGroupItem[];
}

function transformExamination(exam: Examination): ExamGroup {
  return {
    id: exam.id,
    date: exam.date ? exam.date.slice(0, 16).replace("T", " ") : "-",
    machine: exam.machine,
    items: (exam.items ?? []).map((item) => ({
      id: item.id,
      name: item.name,
      result: item.result,
      unit: item.unit,
      reference_value: item.reference_value,
      status: item.status,
    })),
  };
}

const getRecordExaminations = async (
  petId: string,
): Promise<ExamGroup[]> => {
  const { data } = await axios.get<{ data: Examination[] }>("/v1/examinations", {
    params: { pet_id: Number(petId), limit: 100 },
  });
  return (data.data ?? []).map(transformExamination);
};

export const useGetRecordExaminations = (petId?: string) => {
  return useQuery({
    queryKey: ["examinations", "pet", petId],
    queryFn: () => getRecordExaminations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
