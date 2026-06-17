import { paths } from "@/config/paths";

/**
 * #158: 飼主レポートを別ウィンドウで開く。
 * 飼主一覧の行操作・カルテヘッダーの両導線で共有する(URL 構築の重複排除)。
 * セキュリティのため `noopener,noreferrer` を必ず付与する。
 */
export function openOwnerReport(ownerId: string, petId: string): void {
  const href = `${paths.owners.detail.report.getHref(ownerId)}?petId=${encodeURIComponent(petId)}`;
  window.open(href, "_blank", "noopener,noreferrer");
}
