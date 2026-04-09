import axios from 'axios';
import type {
  LiffSettings,
  LiffProfile,
  Course,
  Staff,
  AvailableDate,
  AvailableTime,
  Reservation,
  CreateReservationBody,
} from '../types/models';
import { API_BASE_URL } from '../lib/liff-config';

const httpClient = axios.create({
  baseURL: API_BASE_URL,
});

function authHeaders(idToken: string) {
  return { Authorization: `Bearer ${idToken}` };
}

export const liffApi = {
  getSettings: async (clinicId: string): Promise<LiffSettings> => {
    const res = await httpClient.get<LiffSettings>(`/api/liff/${clinicId}/settings`);
    return res.data;
  },

  getProfile: async (clinicId: string, idToken: string): Promise<LiffProfile> => {
    const res = await httpClient.get<LiffProfile>(`/api/liff/${clinicId}/profile`, {
      headers: authHeaders(idToken),
    });
    return res.data;
  },

  getCourses: async (clinicId: string, idToken: string): Promise<Course[]> => {
    const res = await httpClient.get<Course[]>(`/api/liff/${clinicId}/courses`, {
      headers: authHeaders(idToken),
    });
    return res.data;
  },

  getStaffs: async (clinicId: string, courseId: number, idToken: string): Promise<Staff[]> => {
    const res = await httpClient.get<Staff[]>(`/api/liff/${clinicId}/staffs`, {
      headers: authHeaders(idToken),
      params: { courseId },
    });
    return res.data;
  },

  getAvailableDates: async (
    clinicId: string,
    courseId: number,
    staffId: number,
    idToken: string,
  ): Promise<AvailableDate[]> => {
    const res = await httpClient.get<AvailableDate[]>(`/api/liff/${clinicId}/available-dates`, {
      headers: authHeaders(idToken),
      params: { courseId, staffId },
    });
    return res.data;
  },

  getAvailableTimes: async (
    clinicId: string,
    courseId: number,
    staffId: number,
    date: string,
    idToken: string,
  ): Promise<AvailableTime[]> => {
    const res = await httpClient.get<AvailableTime[]>(`/api/liff/${clinicId}/available-times`, {
      headers: authHeaders(idToken),
      params: { courseId, staffId, date },
    });
    return res.data;
  },

  createReservation: async (
    clinicId: string,
    body: CreateReservationBody,
    idToken: string,
  ): Promise<Reservation> => {
    const res = await httpClient.post<Reservation>(`/api/liff/${clinicId}/reservations`, body, {
      headers: authHeaders(idToken),
    });
    return res.data;
  },

  getMyReservations: async (clinicId: string, idToken: string): Promise<Reservation[]> => {
    const res = await httpClient.get<Reservation[]>(`/api/liff/${clinicId}/my-reservations`, {
      headers: authHeaders(idToken),
    });
    return res.data;
  },

  cancelReservation: async (clinicId: string, id: number, idToken: string): Promise<void> => {
    await httpClient.delete(`/api/liff/${clinicId}/my-reservations/${id}`, {
      headers: authHeaders(idToken),
    });
  },
};
