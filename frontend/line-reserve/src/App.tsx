import { useState, useEffect, useCallback } from 'react';
import type { PageType, LiffSettings, LiffProfile, CustomerInfo } from './types/models';
import { liffApi } from './api/liff-api';
import { useLiff } from '@/shared-liff/use-liff';
import { Spinner } from '@/shared-liff/Spinner';
import { ErrorPage } from '@/shared-liff/ErrorPage';
import { useReservationFlow } from './hooks/use-reservation-flow';
import { getClinicId } from './lib/liff-config';

import { TopPage } from './pages/TopPage';
import { CustomerInfoPage } from './pages/CustomerInfoPage';
import { CourseSelectPage } from './pages/CourseSelectPage';
import { TrimmingCourseSelectPage } from './pages/TrimmingCourseSelectPage';
import { TrimmingOptionSelectPage } from './pages/TrimmingOptionSelectPage';
import { StaffSelectPage } from './pages/StaffSelectPage';
import { DateSelectPage } from './pages/DateSelectPage';
import { TimeSelectPage } from './pages/TimeSelectPage';
import { RequestPage } from './pages/RequestPage';
import { ConfirmPage } from './pages/ConfirmPage';
import { CompletePage } from './pages/CompletePage';
import { MyReservationsPage } from './pages/MyReservationsPage';
import { MaintenancePage } from './pages/MaintenancePage';

export function App() {
  const clinicId = getClinicId();

  const [page, setPage] = useState<PageType>('loading');
  const [settings, setSettings] = useState<LiffSettings | null>(null);
  const [profile, setProfile] = useState<LiffProfile | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(undefined);
  // FE5-21: 予約確定時の枠競合をブロッキングダイアログではなくインラインバナーで通知する
  const [notice, setNotice] = useState<string | null>(null);

  // 完了ページ用
  const [completedReservationId, setCompletedReservationId] = useState<number>(0);
  const [completedNotes, setCompletedNotes] = useState<string>('');

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

  // settings を取得して liffId を確定
  const [liffId, setLiffId] = useState<string>('');
  const { idToken, isReady, initError } = useLiff(liffId);

  // Step 1: settings 取得（認証不要）
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

  // Step 2: LIFF 初期化完了後にプロフィール取得 → トップページへ
  // settings 未取得・失敗時は進めない（VITE_LIFF_MOCK の即時 ready が page=error を上書きするのを防ぐ / BUG-505）
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
        // error/maintenance 確定後の ready レースでも sticky に保つ
        setPage((current) => (current === 'error' || current === 'maintenance' ? current : 'top'));
      });

    return () => {
      cancelled = true;
    };
  }, [isReady, idToken, clinicId, settings]);

  // ナビゲーション
  const goTo = useCallback((p: PageType) => setPage(p), []);

  const handleNewReservation = useCallback(() => {
    setNotice(null); // 直前フローの枠競合バナーを持ち越さない
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo]);

  const handleConfirm = useCallback((reservationId: number, notes: string) => {
    setNotice(null); // 予約成功後は枠競合バナーを残さない
    setCompletedReservationId(reservationId);
    setCompletedNotes(notes);
    goTo('step8');
  }, [goTo]);

  // 予約確定時に時間枠が既に埋まっていた場合のハンドラ
  const handleSlotTaken = useCallback((message: string, redirectStep: number) => {
    setNotice(message); // alert の代わりに通知バナー
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
    setNotice(null); // 直前フローの枠競合バナーを持ち越さない
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo]);

  // ページのレンダリング
  // BUG-014/027: clinic_id 欠落は URL 設定ミス。reload しても解消しないためアクション無し。
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

  let content: React.ReactElement;

  if (page === 'top') {
    content = (
      <TopPage
        settings={settings}
        onNewReservation={handleNewReservation}
        onMyReservations={() => goTo('my-reservations')}
      />
    );
  } else if (page === 'my-reservations') {
    content = (
      <MyReservationsPage
        clinicId={clinicId}
        idToken={idToken}
        onBack={() => goTo('top')}
      />
    );
  } else if (page === 'step1') {
    content = (
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
  } else if (page === 'step2') {
    content = (
      <CourseSelectPage
        clinicId={clinicId}
        idToken={idToken}
        onSelect={(courseId, courseName, category) => {
          setCourse(courseId, courseName, category);
          if (category === 'trimming') {
            goTo('step2b');
          } else {
            goTo('step3');
          }
        }}
        onBack={() => goTo('step1')}
      />
    );
  } else if (page === 'step2b') {
    content = (
      <TrimmingCourseSelectPage
        clinicId={clinicId}
        idToken={idToken}
        onSelect={(trimmingCourseId, trimmingCourseName) => {
          setTrimmingCourse(trimmingCourseId, trimmingCourseName);
          goTo('step2c');
        }}
        onBack={() => goTo('step2')}
      />
    );
  } else if (page === 'step2c') {
    content = (
      <TrimmingOptionSelectPage
        clinicId={clinicId}
        idToken={idToken}
        selectedOptionIds={flow.trimmingOptionIds}
        onNext={(optionIds) => {
          setTrimmingOptions(optionIds);
          goTo('step3');
        }}
        onBack={() => goTo('step2b')}
      />
    );
  } else if (page === 'step3') {
    content = (
      <StaffSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        showNoStaffOption={settings.show_no_staff_option}
        isTrimming={isTrimming}
        onSelect={(staffId, staffName) => {
          setStaff(staffId, staffName);
          goTo('step4');
        }}
        onBack={() => goTo(isTrimming ? 'step2c' : 'step2')}
      />
    );
  } else if (page === 'step4') {
    content = (
      <DateSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        staffId={flow.staffId}
        selectedDate={flow.date}
        bookingWindow={settings.booking_window}
        isTrimming={isTrimming}
        onSelect={setDate}
        onNext={() => goTo('step5')}
        onBack={() => goTo('step3')}
      />
    );
  } else if (page === 'step5') {
    content = (
      <TimeSelectPage
        clinicId={clinicId}
        idToken={idToken}
        courseId={flow.courseId ?? 0}
        staffId={flow.staffId}
        date={flow.date}
        isTrimming={isTrimming}
        onSelect={(startTime, endTime) => {
          setTime(startTime, endTime);
          goTo('step6');
        }}
        onBack={() => goTo('step4')}
      />
    );
  } else if (page === 'step6') {
    content = (
      <RequestPage
        requestExample={settings.request_example}
        initialText={flow.requestText}
        isTrimming={isTrimming}
        onNext={(text) => {
          setRequestText(text);
          goTo('step7');
        }}
        onBack={() => goTo('step5')}
      />
    );
  } else if (page === 'step7') {
    content = (
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
  } else if (page === 'step8') {
    content = (
      <CompletePage
        reservationId={completedReservationId}
        notes={completedNotes}
        flow={flow}
        onMyReservations={handleCompleteToMyReservations}
        onNewReservation={handleCompleteNewReservation}
      />
    );
  } else {
    content = <ErrorPage />;
  }

  return (
    <>
      {notice !== null ? (
        <div role="alert" className="mx-4 mt-4 flex items-start justify-between gap-3 rounded-lg border border-red-200 bg-red-50 px-4 py-3">
          <p className="flex-1 text-sm text-red-700">{notice}</p>
          <button
            type="button"
            onClick={() => setNotice(null)}
            aria-label="閉じる"
            className="font-bold leading-none text-red-700"
          >
            ×
          </button>
        </div>
      ) : null}
      {content}
    </>
  );
}
