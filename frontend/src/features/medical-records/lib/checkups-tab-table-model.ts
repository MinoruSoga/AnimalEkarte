import { todayJSTISO } from "@/lib/jst-date";
import { calculateNextDate } from "@/components/shared/NextScheduleField";

export interface AddCheckupFormState {
  checkup_type_id: string;
  date: string;
  next_schedule_type: string;
  next_date: string;
  doctor_id: string;
  result: string;
}

const EMPTY_ADD_FORM: AddCheckupFormState = {
  checkup_type_id: "",
  date: "",
  next_schedule_type: "1year",
  next_date: "",
  doctor_id: "",
  result: "",
};

function todayISODate(): string {
  return todayJSTISO();
}

export function makeDefaultCheckupAddForm(): AddCheckupFormState {
  const date = todayISODate();
  // EMR-431: 開いた時点で「当日＋既定間隔」の次回予定日を持たせる。
  // 日付に触れずに保存しても next_date が送られる（独立画面 PR665 と同じ挙動）。
  return {
    ...EMPTY_ADD_FORM,
    date,
    next_date: calculateNextDate(date, EMPTY_ADD_FORM.next_schedule_type),
  };
}
