import { useState, useEffect, useCallback } from 'react';
import type { PageType, LiffSettings, LiffProfile } from './types/models';
import { liffApi } from './api/liff-api';
import { useLiff } from '@/shared-liff/use-liff';
import { Spinner } from '@/shared-liff/Spinner';
import { ErrorPage } from '@/shared-liff/ErrorPage';
import { useReservationFlow } from './hooks/use-reservation-flow';
import { getClinicId } from './lib/liff-config';
import { MaintenancePage } from './pages/MaintenancePage';
import { useReservationAppHandlers } from './app-flow-handlers';
import { ReservationNoticeBanner, ReservationPageSwitch } from './app-page-switch';

export function App() {
  const clinicId = getClinicId();

  const [page, setPage] = useState<PageType>('loading');
  const [settings, setSettings] = useState<LiffSettings | null>(null);
  const [profile, setProfile] = useState<LiffProfile | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(undefined);
  const [notice, setNotice] = useState<string | null>(null);
  const [completedReservationId, setCompletedReservationId] = useState<number>(0);
  const [completedNotes, setCompletedNotes] = useState<string>('');
  const [liffId, setLiffId] = useState<string>('');

  const {
    flow,
    setCustomerInfo,
    setCourse,
    setStaff,
    setDate,
    setTime,
    setRequestText,
    setTrimmingCourse,
    setTrimmingOptions,
    resetFlow,
  } = useReservationFlow();

  const isTrimming = flow.courseCategory === 'trimming';
  const { idToken, isReady, initError } = useLiff(liffId);
  const goTo = useCallback((nextPage: PageType) => setPage(nextPage), []);
  const handlers = useReservationAppHandlers(
    goTo,
    resetFlow,
    setNotice,
    setCompletedReservationId,
    setCompletedNotes,
  );

  useEffect(() => {
    if (!clinicId) return;

    liffApi.getSettings(clinicId)
      .then(s => {
        setSettings(s);
        if (s.status !== 'running') {
          setPage('maintenance');
          return;
        }
        setLiffId(s.liff_id);
      })
      .catch(() => {
        setErrorMessage('設定の取得に失敗しました');
        setPage('error');
      });
  }, [clinicId]);

  useEffect(() => {
    if (!isReady || !idToken || !clinicId) return;
    if (!settings || settings.status !== 'running') return;

    let cancelled = false;

    liffApi.getProfile(clinicId, idToken)
      .then(p => {
        if (cancelled) return;
        setProfile(p);
      })
      .catch(() => {
        // プロフィール取得失敗は無視してトップへ進む
      })
      .finally(() => {
        if (cancelled) return;
        setPage((current) => (current === 'error' || current === 'maintenance' ? current : 'top'));
      });

    return () => {
      cancelled = true;
    };
  }, [isReady, idToken, clinicId, settings]);

  if (!clinicId) {
    return <ErrorPage message="クリニックIDが見つかりません" showAction={false} />;
  }

  if (initError) {
    return (
      <ErrorPage message="LINEアプリの初期化に失敗しました。LINEアプリから再度お試しください。" />
    );
  }

  if (page === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-noah-teal-light">
        <div className="text-center">
          <Spinner size="sm" borderColorClassName="border-noah-teal" />
          <p className="text-noah-text-sub text-sm">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (page === 'error') {
    return <ErrorPage message={errorMessage} />;
  }

  if (page === 'maintenance') {
    return <MaintenancePage />;
  }

  if (!settings || !idToken) {
    return <ErrorPage message="初期化に失敗しました" />;
  }

  return (
    <>
      <ReservationNoticeBanner notice={notice} onDismiss={() => setNotice(null)} />
      <ReservationPageSwitch
        page={page}
        settings={settings}
        profile={profile}
        clinicId={clinicId}
        idToken={idToken}
        flow={flow}
        isTrimming={isTrimming}
        completedReservationId={completedReservationId}
        completedNotes={completedNotes}
        goTo={goTo}
        setCustomerInfo={setCustomerInfo}
        setCourse={setCourse}
        setStaff={setStaff}
        setDate={setDate}
        setTime={setTime}
        setRequestText={setRequestText}
        setTrimmingCourse={setTrimmingCourse}
        setTrimmingOptions={setTrimmingOptions}
        {...handlers}
      />
    </>
  );
}
