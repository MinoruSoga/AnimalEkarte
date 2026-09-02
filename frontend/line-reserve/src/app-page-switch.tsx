import type { CustomerInfo, LiffProfile, LiffSettings, PageType, ReservationFlow } from './types/models';
import { ErrorPage } from '@/shared-liff/ErrorPage';

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

export interface ReservationPageSwitchProps {
  page: PageType;
  settings: LiffSettings;
  profile: LiffProfile | null;
  clinicId: string;
  idToken: string;
  flow: ReservationFlow;
  isTrimming: boolean;
  completedReservationId: number;
  completedNotes: string;
  goTo: (page: PageType) => void;
  setCustomerInfo: (info: CustomerInfo) => void;
  setCourse: (id: number, name: string, category?: 'general' | 'trimming') => void;
  setStaff: (id: number, name: string) => void;
  setDate: (date: string) => void;
  setTime: (startTime: string, endTime: string) => void;
  setRequestText: (text: string) => void;
  setTrimmingCourse: (id: number, name: string) => void;
  setTrimmingOptions: (ids: number[]) => void;
  handleNewReservation: () => void;
  handleConfirm: (reservationId: number, notes: string) => void;
  handleSlotTaken: (message: string, redirectStep: number) => void;
  handleCompleteToMyReservations: () => void;
  handleCompleteNewReservation: () => void;
}

export function ReservationNoticeBanner({
  notice,
  onDismiss,
}: {
  notice: string | null;
  onDismiss: () => void;
}) {
  if (notice === null) return null;
  return (
    <div role="alert" className="mx-4 mt-4 flex items-start justify-between gap-3 rounded-lg border border-red-200 bg-red-50 px-4 py-3">
      <p className="flex-1 text-sm text-red-700">{notice}</p>
      <button
        type="button"
        onClick={onDismiss}
        aria-label="閉じる"
        className="font-bold leading-none text-red-700"
      >
        ×
      </button>
    </div>
  );
}

function renderAccountPages(props: ReservationPageSwitchProps) {
  if (props.page === 'top') {
    return (
      <TopPage
        settings={props.settings}
        onNewReservation={props.handleNewReservation}
        onMyReservations={() => props.goTo('my-reservations')}
      />
    );
  }
  if (props.page === 'my-reservations') {
    return (
      <MyReservationsPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        onBack={() => props.goTo('top')}
      />
    );
  }
  return null;
}

function renderEarlyReservationSteps(props: ReservationPageSwitchProps) {
  if (props.page === 'step1') {
    return (
      <CustomerInfoPage
        profile={props.profile}
        initialInfo={props.flow.customerInfo}
        onNext={(info: CustomerInfo) => {
          props.setCustomerInfo(info);
          props.goTo('step2');
        }}
        onBack={() => props.goTo('top')}
      />
    );
  }
  if (props.page === 'step2') {
    return (
      <CourseSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        onSelect={(courseId, courseName, category) => {
          props.setCourse(courseId, courseName, category);
          if (category === 'trimming') {
            props.goTo('step2b');
          } else {
            props.goTo('step3');
          }
        }}
        onBack={() => props.goTo('step1')}
      />
    );
  }
  if (props.page === 'step2b') {
    return (
      <TrimmingCourseSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        onSelect={(trimmingCourseId, trimmingCourseName) => {
          props.setTrimmingCourse(trimmingCourseId, trimmingCourseName);
          props.goTo('step2c');
        }}
        onBack={() => props.goTo('step2')}
      />
    );
  }
  if (props.page === 'step2c') {
    return (
      <TrimmingOptionSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        selectedOptionIds={props.flow.trimmingOptionIds}
        onNext={(optionIds) => {
          props.setTrimmingOptions(optionIds);
          props.goTo('step3');
        }}
        onBack={() => props.goTo('step2b')}
      />
    );
  }
  return null;
}

function renderLateReservationSteps(props: ReservationPageSwitchProps) {
  if (props.page === 'step3') {
    return (
      <StaffSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        courseId={props.flow.courseId ?? 0}
        showNoStaffOption={props.settings.show_no_staff_option}
        isTrimming={props.isTrimming}
        onSelect={(staffId, staffName) => {
          props.setStaff(staffId, staffName);
          props.goTo('step4');
        }}
        onBack={() => props.goTo(props.isTrimming ? 'step2c' : 'step2')}
      />
    );
  }
  if (props.page === 'step4') {
    return (
      <DateSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        courseId={props.flow.courseId ?? 0}
        staffId={props.flow.staffId}
        selectedDate={props.flow.date}
        bookingWindow={props.settings.booking_window}
        isTrimming={props.isTrimming}
        onSelect={props.setDate}
        onNext={() => props.goTo('step5')}
        onBack={() => props.goTo('step3')}
      />
    );
  }
  if (props.page === 'step5') {
    return (
      <TimeSelectPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        courseId={props.flow.courseId ?? 0}
        staffId={props.flow.staffId}
        date={props.flow.date}
        isTrimming={props.isTrimming}
        onSelect={(startTime, endTime) => {
          props.setTime(startTime, endTime);
          props.goTo('step6');
        }}
        onBack={() => props.goTo('step4')}
      />
    );
  }
  if (props.page === 'step6') {
    return (
      <RequestPage
        requestExample={props.settings.request_example}
        initialText={props.flow.requestText}
        isTrimming={props.isTrimming}
        onNext={(text) => {
          props.setRequestText(text);
          props.goTo('step7');
        }}
        onBack={() => props.goTo('step5')}
      />
    );
  }
  if (props.page === 'step7') {
    return (
      <ConfirmPage
        clinicId={props.clinicId}
        idToken={props.idToken}
        flow={props.flow}
        reservationNotice={props.settings.reservation_notice}
        cancelNotice={props.settings.cancel_notice}
        privacyPolicy={props.settings.privacy_policy}
        onConfirm={props.handleConfirm}
        onSlotTaken={props.handleSlotTaken}
        onBack={() => props.goTo('step6')}
      />
    );
  }
  if (props.page === 'step8') {
    return (
      <CompletePage
        reservationId={props.completedReservationId}
        notes={props.completedNotes}
        flow={props.flow}
        onMyReservations={props.handleCompleteToMyReservations}
        onNewReservation={props.handleCompleteNewReservation}
      />
    );
  }
  return <ErrorPage />;
}

export function ReservationPageSwitch(props: ReservationPageSwitchProps) {
  const account = renderAccountPages(props);
  if (account) return account;
  const early = renderEarlyReservationSteps(props);
  if (early) return early;
  return renderLateReservationSteps(props);
}
