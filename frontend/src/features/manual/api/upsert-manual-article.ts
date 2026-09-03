/**
 * upsert-manual-article.ts — マニュアル記事の編集保存 API
 *
 * 管理者権限（ResourceManualEdit）が必要。
 * 編集内容を DB にオーバーライド版として保存。
 */

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import type { ManualCategory } from "../lib/manual-index";

import type { ManualArticleOverride } from "./get-manual-articles";

interface UpsertManualArticleParams {
  category: ManualCategory;
  slug: string;
  title: string;
  order_value: number;
  section: string;
  body_markdown: string;
}

async function upsertManualArticle(
  params: UpsertManualArticleParams,
): Promise<ManualArticleOverride> {
  const { category, slug, ...body } = params;
  const { data } = await axios.put<ManualArticleOverride>(
    `/v1/manual/articles/${category}/${slug}`,
    body,
  );
  return data;
}

export function useUpsertManualArticle() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: upsertManualArticle,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.manualArticles.all() });
      toast.success("マニュアルを保存しました");
    },
    onError: (error) => {
      handleApiError(error, "マニュアル保存");
    },
  });
}
