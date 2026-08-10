import { paths } from "@/config/paths";

export type PostCreateOwnerNavigation =
  | { mode: "spa"; href: string }
  | { mode: "hard"; href: string; clinicId: string };

/**
 * BUG-010: 登録先医院がグローバル選択と異なる場合、詳細 GET は作成先 clinic の
 * X-Clinic-ID が必要なため hard navigation（storage 更新後のフルロード）を選ぶ。
 * 同一医院なら従来どおり SPA navigate。
 */
export function resolvePostCreateOwnerNavigation(args: {
  ownerId: string;
  targetClinicId?: string | null;
  currentClinicId?: string | null;
}): PostCreateOwnerNavigation {
  const href = paths.owners.detail.getHref(args.ownerId);
  const target = args.targetClinicId?.trim() || null;
  const current = args.currentClinicId?.trim() || null;
  if (target && current && target !== current) {
    return { mode: "hard", href, clinicId: target };
  }
  return { mode: "spa", href };
}
