import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { useGetMedicalRecords } from "@/features/medical-records";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import { useGetPetVaccinations } from "@/hooks/use-pet-vaccinations";
import { useGetPetCheckupResults } from "./use-pet-checkup-results";
import { formatJSTDate, todayJSTISO } from "@/lib/jst-date";
import {
  ResourceCheckups,
  ResourceExaminations,
  ResourceReservations,
  ResourceTrimming,
  ResourceVaccinations,
} from "@/types/generated/models";

import { useGetPetExaminations } from "../api/get-pet-examinations";
import { useGetPetTreatmentHistory } from "../api/get-pet-treatment-history";
import { useGetPetTrimmingHistory } from "../api/get-pet-trimming-history";

function addDaysISO(date: string, days: number): string {
  const instant = new Date(`${date}T00:00:00+09:00`);
  instant.setUTCDate(instant.getUTCDate() + days);
  return formatJSTDate(instant);
}

function useClinicalPermissions() {
  return {
    examination: usePermission(ResourceExaminations),
    vaccination: usePermission(ResourceVaccinations),
    checkup: usePermission(ResourceCheckups),
    trimming: usePermission(ResourceTrimming),
    reservation: usePermission(ResourceReservations),
  };
}

function useClinicalQueries(
  petId: string,
  today: string,
  permissions: ReturnType<typeof useClinicalPermissions>,
) {
  const medicalRecordsQuery = useGetMedicalRecords({
    petId,
    status: "finalized",
    page: 1,
    limit: HISTORY_FETCH_LIMIT,
    sort: "date",
    order: "desc",
  });
  const examinationsQuery = useGetPetExaminations(
    permissions.examination.canView ? petId : undefined,
  );
  const vaccinationsQuery = useGetPetVaccinations(
    permissions.vaccination.canView ? petId : undefined,
  );
  const checkupsQuery = useGetPetCheckupResults(
    permissions.checkup.canView ? petId : undefined,
  );
  const treatmentsQuery = useGetPetTreatmentHistory(petId, "all");
  const trimmingQuery = useGetPetTrimmingHistory(
    permissions.trimming.canView ? petId : undefined,
  );
  const reservationsQuery = useGetReservations({
    startDate: today,
    endDate: addDaysISO(today, 365),
    petId,
    enabled: permissions.reservation.canView,
  });

  return {
    medicalRecordsQuery,
    examinationsQuery,
    vaccinationsQuery,
    checkupsQuery,
    treatmentsQuery,
    trimmingQuery,
    reservationsQuery,
  };
}

export function useOwnerClinicalBriefingData(petId: string) {
  const permissions = useClinicalPermissions();
  const today = todayJSTISO();
  const queries = useClinicalQueries(petId, today, permissions);

  return { permissions, today, ...queries };
}

export type OwnerClinicalBriefingData = ReturnType<
  typeof useOwnerClinicalBriefingData
>;
