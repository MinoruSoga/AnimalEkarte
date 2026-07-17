import { useAuth } from "@/hooks/use-auth";

/**
 * 現在選択中の医院名を返す共有フック。
 *
 * 印刷／帳票ヘッダー（#184 月次集計レポート・#153 レジ締めサマリー）で
 * 病院名を表示するために使用する。currentClinicId のメンバーシップが正本。
 * 会計帳票の誤帰属を避けるため、メンバーシップが見つからない場合は
 * メイン医院名へフォールバックせず空文字を返す（currentClinicId は
 * use-auth 側で所属医院に制約されるため、通常ヒットする）。
 */
export function useCurrentClinicName(): string {
  const { user, currentClinicId } = useAuth();
  if (!user) return "";
  const membership = user.clinics.find((c) => c.clinicId === currentClinicId);
  return membership?.clinicName ?? "";
}
