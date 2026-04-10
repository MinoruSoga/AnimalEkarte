import { useState, useEffect, useCallback } from 'react';
import type { PageType, LiffSettings, LiffProfile, CustomerInfo } from './types/models';
import { liffApi } from './api/liff-api';
import { useLiff } from './hooks/use-liff';
import { useReservationFlow } from './hooks/use-reservation-flow';
import { getClinicId } from './lib/liff-config';

import { TopPage } from './pages/TopPage';
import { CustomerInfoPage } from './pages/CustomerInfoPage';
import { CourseSelectPage } from './pages/CourseSelectPage';
import { StaffSelectPage } from './pages/StaffSelectPage';
import { DateSelectPage } from './pages/DateSelectPage';
import { TimeSelectPage } from './pages/TimeSelectPage';
import { RequestPage } from './pages/RequestPage';
import { ConfirmPage } from './pages/ConfirmPage';
import { CompletePage } from './pages/CompletePage';
import { MyReservationsPage } from './pages/MyReservationsPage';
import { ErrorPage } from './pages/ErrorPage';
import { MaintenancePage } from './pages/MaintenancePage';

export function App() {
  const clinicId = getClinicId();

  const [page, setPage] = useState<PageType>('loading');
  const [settings, setSettings] = useState<LiffSettings | null>(null);
  const [profile, setProfile] = useState<LiffProfile | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(undefined);

  // 完了ページ用
  const [completedReservationId, setCompletedReservationId] = useState<number>(0);
  const [completedNotes, setCompletedNotes] = useState<string>('');

  const { flow, setCustomerInfo, setCourse, setStaff, setDate, setTime, setRequestText, resetFlow } =
    useReservationFlow();

  // settings を取得して liffId を確定
  const [liffId, setLiffId] = useState<string>('');
  const { idToken, isReady, initError } = useLiff(liffId);

  // LIFF 初期化失敗時はエラーページへ
  useEffect(() => {
    if (initError) {
      setErrorMessage('LINEアプリの初期化に失敗しました。LINEアプリから再度お試しください。');
      setPage('error');
    }
  }, [initError]);

  // Step 1: settings 取得（認証不要）
  useEffect(() => {
    if (!clinicId) {
      setErrorMessage('クリニックIDが見つかりません');
      setPage('error');
      return;
    }

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

  // Step 2: LIFF 初期化完了後にプロフィール取得 → トップページへ
  useEffect(() => {
    if (!isReady || !idToken || !clinicId) return;

    liffApi.getProfile(clinicId, idToken)
      .then(p => {
        setProfile(p);
      })
      .catch(() => {
        // プロフィール取得失敗は無視してトップへ進む
      })
      .finally(() => {
        setPage('top');
      });
  }, [isReady, idToken, clinicId]);

  // ナビゲーション
  const goTo = useCallback((p: PageType) => setPage(p), []);

  const handleNewReservation = useCallback(() => {
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo]);

  const handleConfirm = useCallback((reservationId: number, notes: string) => {
    setCompletedReservationId(reservationId);
    setCompletedNotes(notes);
    goTo('step8');
  }, [goTo]);

  // 予約確定時に時間枠が既に埋まっていた場合のハンドラ
  const handleSlotTaken = useCallback((message: string, redirectStep: number) => {
    window.alert(message);
    const stepMap: Record<number, PageType> = {
      1: 'step1', 2: 'step2', 3: 'step3', 4: 'step4',
      5: 'step5', 6: 'step6', 7: 'step7',
    };
    goTo(stepMap[redirectStep] ?? 'step4');
  }, [goTo]);

  const handleCompleteToMyReservations = useCallback(() => {
    goTo('my-reservations');
  }, [goTo]);

  const handleCompleteNewReservation = useCallback(() => {
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo]);

  // ページのレンダリング
  if (page === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-noah-teal-light">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-noah-teal border-t-transparent rounded-full animate-spin mx-auto mb-3" aria-hidden="true" />
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

  if (page === 'top') {
    return (
      <TopPage
        settings={settings}
        onNewReservation={handleNewReservation}
        onMyReservations={() => goTo('my-reservations')}
      />
    );
  }

  if (page === 'my-reservations') {
    return (
      <MyReservationsPage
        clinicId={clinicId}
        idToken={idToken}
        onBack={() => goTo('top')}
      />
    );
  }

  if (page === 'step1') {
    return (
      <CustomerInfoPage
        profile={profile}
        initialInfo={flow.customerInfo}
        onNext={(info: CustomerInfo) => {
          setCustomerInfo(info);
          goTo('step2');
        }}
        onBack={() => goTo('top')}
      />
    );
  }

  if (page === 'step2') {
    return (
      <CourseSelectPage
        clinicId={clinicId}
        idToken={idToken}
        onSelect={(courseId, courseName) => {
          setCourse(courseId, courseName);
          goTo('step3');
        }}
        onBack={() => goTo('step1')}
      />
    );
  }

  if (page === 'step3') {
    return (
      <StaffSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        showNoStaffOption={settings.show_no_staff_option}
        onSelect={(staffId, staffName) => {
          setStaff(staffId, staffName);
          goTo('step4');
        }}
        onBack={() => goTo('step2')}
      />
    );
  }

  if (page === 'step4') {
    return (
      <DateSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        staffId={flow.staffId}
        selectedDate={flow.date}
        bookingWindow={settings.booking_window}
        onSelect={setDate}
        onNext={() => goTo('step5')}
        onBack={() => goTo('step3')}
      />
    );
  }

  if (page === 'step5') {
    return (
      <TimeSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        staffId={flow.staffId}
        date={flow.date}
        onSelect={(startTime, endTime) => {
          setTime(startTime, endTime);
          goTo('step6');
        }}
        onBack={() => goTo('step4')}
      />
    );
  }

  if (page === 'step6') {
    return (
      <RequestPage
        requestExample={settings.request_example}
        initialText={flow.requestText}
        onNext={(text) => {
          setRequestText(text);
          goTo('step7');
        }}
        onBack={() => goTo('step5')}
      />
    );
  }

  if (page === 'step7') {
    return (
      <ConfirmPage
        clinicId={clinicId}
        idToken={idToken}
        flow={flow}
        reservationNotice={settings.reservation_notice}
        cancelNotice={settings.cancel_notice}
        privacyPolicy={settings.privacy_policy}
        onConfirm={handleConfirm}
        onSlotTaken={handleSlotTaken}
        onBack={() => goTo('step6')}
      />
    );
  }

  if (page === 'step8') {
    return (
      <CompletePage
        reservationId={completedReservationId}
        notes={completedNotes}
        flow={flow}
        onMyReservations={handleCompleteToMyReservations}
        onNewReservation={handleCompleteNewReservation}
      />
    );
  }

  return <ErrorPage />;
}
