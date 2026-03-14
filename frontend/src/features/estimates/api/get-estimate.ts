import { useQuery } from '@tanstack/react-query';
import { axios } from '@/lib/axios';
import type { Estimate } from '../types';
import { transformEstimate } from './transforms';
import type { BackendEstimate } from './types';

export async function getEstimate(id: string): Promise<Estimate> {
  const { data } = await axios.get<BackendEstimate>(`/v1/estimates/${id}`);
  return transformEstimate(data);
}

export function useEstimate(id: string | undefined) {
  return useQuery({
    queryKey: ['estimates', id],
    queryFn: () => getEstimate(id!),
    enabled: !!id,
  });
}
