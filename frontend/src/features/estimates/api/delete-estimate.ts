import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { axios } from '@/lib/axios';
import { handleApiError } from '@/lib/handle-api-error';

export async function deleteEstimate(id: string): Promise<void> {
  await axios.delete(`/v1/estimates/${id}`);
}

export function useDeleteEstimate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteEstimate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['estimates'] });
      toast.success('見積書を削除しました');
    },
    onError: (error) => {
      handleApiError(error, "削除");
    },
  });
}
