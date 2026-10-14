import { useMemo } from "react";
import { usePermission } from "@/hooks/use-permission";
import { RESOURCE_MASTER_PERMISSION } from "../constants/permission-resources";

import { useGetStaffs, useGetClinicsList, useGetAllStaffPermissionGroupMap } from "../api/staffs";
import { useGetPermissionGroups } from "../api/permission-groups";
import { useGetAllOccupations } from "../api/occupations";
import { useGetReservationTypes } from "../api/reservation-types";
import {
  buildGroupsByStaffId,
  buildStaffFilterProperties,
  buildStaffIds,
} from "../routes/staff-settings-model";

export function useStaffSettingsLookups() {
  const { data } = useGetStaffs();
  const { data: allOccupationsData } = useGetAllOccupations();
  const allOccupations = useMemo(() => allOccupationsData ?? [], [allOccupationsData]);
  const { canView } = usePermission(RESOURCE_MASTER_PERMISSION);
  const canViewGroups = canView === true;
  const { data: allGroupsData } = useGetPermissionGroups(canViewGroups);
  const allGroups = useMemo(
    () => (canViewGroups ? (allGroupsData ?? []) : []),
    [canViewGroups, allGroupsData],
  );
  const { data: allReservationTypesData } = useGetReservationTypes();
  const allReservationTypes = useMemo(
    () => allReservationTypesData ?? [],
    [allReservationTypesData],
  );
  const { data: allClinicsData } = useGetClinicsList("all");
  const allClinics = useMemo(() => allClinicsData ?? [], [allClinicsData]);
  const staffIds = useMemo(() => buildStaffIds(data), [data]);
  const { data: staffGroupMap } = useGetAllStaffPermissionGroupMap(staffIds, canViewGroups);
  const groupsByStaffId = useMemo(
    () =>
      buildGroupsByStaffId({
        staffGroupMap: canViewGroups ? staffGroupMap : undefined,
        groups: allGroups,
      }),
    [canViewGroups, staffGroupMap, allGroups],
  );
  const staffFilterProperties = useMemo(
    () => buildStaffFilterProperties(allOccupations),
    [allOccupations],
  );

  return {
    data,
    allOccupations,
    allGroups,
    allReservationTypes,
    allClinics,
    groupsByStaffId,
    staffFilterProperties,
  };
}
