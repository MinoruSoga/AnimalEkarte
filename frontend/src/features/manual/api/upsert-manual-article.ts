/**
 * upsert-manual-article.ts — マニュアル記事の編集保存 hook（EMR-513）
 *
 * 保存は契約wire adapter (`./manual-article-edit`) へ統合した:
 * PUT /v1/manual/articles/{category}/{slug} + receipt照合（対象/版一致のみ成功）。
 * 旧wire（operation_key無・number ID・応答を一覧DTOとして扱う形式）は廃止。
 *
 * 成功/409 code別/結果不明（同operation_keyでの照会）/4xx,5xx の分岐は
 * 呼び出し側（ManualEditor）で行う。このhookは保存入口（mutation wrapper）のみを担う。
 */

import { useMutation } from "@tanstack/react-query";

import {
  saveManualArticleEdit,
  type ManualArticleSaveOutcome,
  type SaveManualArticleEditParams,
} from "./manual-article-edit";

export type { ManualArticleSaveOutcome, SaveManualArticleEditParams };

export function useUpsertManualArticle() {
  return useMutation({
    mutationFn: saveManualArticleEdit,
  });
}
