import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { axios } from '@/lib/axios';
import type { Estimate } from '../types';
import { transformEstimate } from './transforms';
import type { BackendEstimate, UpdateEstimateRequest } from './types';

interface UpdateEstimateParams {
  id: string;
  data: UpdateEstimateRequest;
}

export async function updateEstimate({ id, data }: UpdateEstimateParams): Promise<Estimate> {
  const { data: responseData } = await axios.patch<BackendEstimate>(`/v1/estimates/${id}`, data);
  return transformEstimate(responseData);
}

export function useUpdateEstimate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateEstimate,
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['estimates'] });
      queryClient.invalidateQueries({ queryKey: ['estimates', id] });
      toast.success('見積書を更新しました');
    },
    onError: (error: Error) => {
      toast.error(error.message || '見積書の更新に失敗しました');
    },
  });
}
