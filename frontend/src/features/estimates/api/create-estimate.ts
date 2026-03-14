import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { axios } from '@/lib/axios';
import type { Estimate } from '../types';
import { transformEstimate } from './transforms';
import type { BackendEstimate, CreateEstimateRequest } from './types';

export async function createEstimate(req: CreateEstimateRequest): Promise<Estimate> {
  const { data } = await axios.post<BackendEstimate>('/v1/estimates', req);
  return transformEstimate(data);
}

export function useCreateEstimate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createEstimate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['estimates'] });
      toast.success('見積書を作成しました');
    },
    onError: (error: Error) => {
      toast.error(error.message || '見積書の作成に失敗しました');
    },
  });
}
