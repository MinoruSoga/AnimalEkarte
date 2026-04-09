// LIFF App 型定義

export interface LiffSettings {
  liff_id: string;
  clinic_name: string;
  header_text: string;
  phone_number: string;
  status: 'running' | 'stopped';
  request_example: string;
  show_no_staff_option: boolean;
  booking_window: number; // 予約可能な日数（今日から何日後まで）
}

export interface OwnerPet {
  id: number;
  name: string;
  animal_species?: { name: string };
}

export interface LiffProfile {
  line_user_id: string;
  display_name: string;
  additional_fields: {
    name?: string;
    phone?: string;
    owner_name?: string;
    pets?: Array<{ name: string; type: string; is_new?: boolean }>;
  };
  // 自動紐付けされたオーナー情報（紐付け済みの場合のみ存在）
  owner?: {
    owner_name: string;
    phone: string;
    pets?: OwnerPet[];
  };
}

export interface Course {
  id: number;
  name: string;
  short_name: string;
  show_short_name: boolean;
  duration_minutes: number;
  reservation_comment: string;
  reservation_image_url: string;
  sort_order: number;
}

export interface Staff {
  id: number;
  name: string;
  description: string;
}

export interface AvailableDate {
  date: string; // "YYYY-MM-DD"
  available: boolean;
}

export interface AvailableTime {
  start_time: string; // "HHMM"
  end_time: string;   // "HHMM"
  display_time: string; // "HH:MM"
}

export interface Reservation {
  id: number;
  course_name: string;
  staff_name: string;
  date: string;       // "YYYY-MM-DD"
  start_time: string; // "HHMM"
  end_time: string;   // "HHMM"
  status: 'confirmed' | 'cancelled' | 'completed';
  notes: string;
  created_at: string;
}

export interface PetSelection {
  name: string;
  type: string;
  isNew?: boolean;   // true = 自由入力で追加した新規ペット
}

export interface CustomerFields {
  customer_name?: string;
  phone?: string;
  owner_name?: string;
  pets?: Array<{ name: string; type: string; is_new?: boolean }>;
}

export interface CreateReservationBody {
  course_id: number;
  staff_id: number;
  date: string;           // "YYYY-MM-DD"
  start_time: string;     // "HHMM"
  end_time: string;       // "HHMM"
  customer_fields: CustomerFields;
  request_text: string;
}

export interface CustomerInfo {
  name: string;
  phone: string;
  ownerName: string;
  pets: PetSelection[];
}

export interface ReservationFlow {
  customerInfo: CustomerInfo;
  courseId: number | null;
  courseName: string;
  staffId: number; // 0 = 指名なし
  staffName: string;
  date: string;      // "YYYY-MM-DD"
  startTime: string; // "HHMM"
  endTime: string;   // "HHMM"
  requestText: string;
}

export type PageType =
  | 'loading'
  | 'error'
  | 'maintenance'
  | 'top'
  | 'my-reservations'
  | 'step1'
  | 'step2'
  | 'step3'
  | 'step4'
  | 'step5'
  | 'step6'
  | 'step7'
  | 'step8';
