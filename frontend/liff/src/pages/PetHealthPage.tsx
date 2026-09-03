import { useCallback, useEffect, useState } from "react";
import { fetchBrandSettings, fetchHealthCard } from "../api/liff-api";
import { useFetchState } from "@/shared-liff/use-fetch-state";
import { Spinner } from "@/shared-liff/Spinner";
import { ErrorPage } from "@/shared-liff/ErrorPage";
import { resolveClinicHeaderText } from "@/shared-liff/brand-tokens";

interface PetHealthPageProps {
  idToken: string;
  displayName: string;
  pictureUrl: string | null;
}

/**
 * BUG-014: clinic_id 欠落は再試行で解消しない恒久的な設定ミス。
 * useFetchState → resolveFetchError に載せると status 無し Error が汎用リトライ文言に潰れるため、
 * line-reserve と同様に fetch 前で専用文言へ分岐する（再試行ボタンも出さない）。
 * hooks 規則のため、欠落分岐と fetch 本体はコンポーネントを分ける。
 *
 * BUG-027: 失敗 chrome は shared-liff/ErrorPage に統一。
 */
export function PetHealthPage({ idToken, displayName, pictureUrl }: PetHealthPageProps) {
  const clinicId = new URLSearchParams(window.location.search).get("clinic_id") ?? "";

  if (!clinicId) {
    return <ErrorPage message="クリニックIDが見つかりません" showAction={false} />;
  }

  return (
    <PetHealthPageContent
      clinicId={clinicId}
      idToken={idToken}
      displayName={displayName}
      pictureUrl={pictureUrl}
    />
  );
}

interface PetHealthPageContentProps extends PetHealthPageProps {
  clinicId: string;
}

function PetHealthPageContent({
  clinicId,
  idToken,
  displayName,
  pictureUrl,
}: PetHealthPageContentProps) {
  const fetcher = useCallback(() => {
    return fetchHealthCard(idToken, clinicId);
  }, [idToken, clinicId]);

  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data, loading, error, retry } = useFetchState(fetcher, "健康記録の取得");

  // BUG-026: clinic brand from public settings; fail closed to empty (no fabricated name).
  const [clinicHeaderText, setClinicHeaderText] = useState("");
  useEffect(() => {
    let cancelled = false;
    fetchBrandSettings(clinicId)
      .then((settings) => {
        if (!cancelled) {
          setClinicHeaderText(resolveClinicHeaderText(settings.header_text));
        }
      })
      .catch(() => {
        if (!cancelled) {
          setClinicHeaderText("");
        }
      });
    return () => {
      cancelled = true;
    };
  }, [clinicId]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-liff-brand-bg">
        <div className="text-center">
          <Spinner />
          <p className="text-noah-text-muted text-sm">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error || !data) {
    const canRetry = error?.canRetry !== false;
    return (
      <ErrorPage
        message={error?.message ?? "不明なエラーが発生しました"}
        showAction={canRetry}
        onAction={canRetry ? retry : undefined}
        actionLabel="再試行"
      />
    );
  }

  return (
    <div className="min-h-screen bg-liff-brand-bg pb-10">
      {/* BUG-026: teal clinic brand header (aligned with line-reserve), owner secondary */}
      <header className="bg-liff-brand text-white px-4 py-4">
        {clinicHeaderText ? (
          <h1 className="text-lg font-bold tracking-tight text-center">{clinicHeaderText}</h1>
        ) : null}
        <div className={`flex items-center gap-3 ${clinicHeaderText ? "mt-3" : ""}`}>
          {pictureUrl ? (
            <img
              src={pictureUrl}
              alt={displayName}
              className="w-10 h-10 rounded-full object-cover border border-white/30"
            />
          ) : null}
          <div>
            <p className="text-xs text-white/70">飼い主</p>
            <p className="text-base font-semibold text-white">{data.owner_name || displayName}</p>
          </div>
        </div>
      </header>

      <div className="max-w-lg mx-auto px-4 mt-6 space-y-6">
        {data.pets.length === 0 ? (
          <p className="text-center text-noah-text-muted mt-10">ペット情報はありません</p>
        ) : null}

        {data.pets.map((pet) => (
          <div
            key={pet.pet_id}
            className="bg-white rounded-2xl border border-noah-border-light overflow-hidden"
          >
            {/* ペット基本情報 */}
            <div className="bg-liff-brand px-5 py-4">
              <h2 className="text-lg font-bold text-white">{pet.pet_name}</h2>
              <p className="text-liff-brand-subtle text-sm">
                {pet.species} / {pet.breed}
              </p>
            </div>

            <div className="px-5 py-4 space-y-3">
              {/* 最終来院日 */}
              <div className="flex justify-between text-sm">
                <span className="text-noah-text-muted">最終来院日</span>
                <span className="font-medium text-noah-text-strong">
                  {pet.last_visit_date ? pet.last_visit_date : "記録なし"}
                </span>
              </div>

              {/* ワクチン記録 */}
              {pet.vaccines.length > 0 ? (
                <div className="mt-3">
                  <p className="text-xs font-semibold text-noah-text-faint uppercase tracking-wide mb-2">
                    ワクチン記録
                  </p>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-noah-border-light">
                          <th className="text-left py-2 text-noah-text-muted font-medium">
                            ワクチン名
                          </th>
                          <th className="text-left py-2 text-noah-text-muted font-medium">
                            接種日
                          </th>
                          <th className="text-left py-2 text-noah-text-muted font-medium">
                            次回予定日
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {pet.vaccines.map((v, i) => (
                          <tr key={i} className="border-b border-noah-border-faint last:border-0">
                            <td className="py-2 text-noah-text-strong">{v.vaccine_name}</td>
                            <td className="py-2 text-noah-text-secondary">{v.vaccinated_at}</td>
                            <td className="py-2 text-liff-brand-dark">
                              {v.next_due_at ? v.next_due_at : "—"}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
