import { useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import type { ClinicInfo } from "@/types";
import type { BackendClinic } from "@/features/hospital-settings/api/types";

function transformToClinicInfo(data: BackendClinic): ClinicInfo {
  return {
    name: data.name ?? "",
    postalCode: data.postal_code ?? "",
    address: data.address ?? "",
    phoneNumber: data.phone_number ?? "",
    faxNumber: data.fax_number,
    registrationNumber: data.registration_number,
    directorName: data.director_name,
    email: data.email,
    website: data.website,
    logoUrl: data.logo_url,
  };
}

const getClinics = async (): Promise<BackendClinic[]> => {
  const { data } = await axios.get<BackendClinic[]>("/v1/clinics");
  return data;
};

const updateClinic = async (id: string, info: ClinicInfo): Promise<BackendClinic> => {
  const { data } = await axios.patch<BackendClinic>(`/v1/clinics/${id}`, {
    name: info.name,
    postal_code: info.postalCode,
    address: info.address,
    phone_number: info.phoneNumber,
    fax_number: info.faxNumber,
    registration_number: info.registrationNumber,
    director_name: info.directorName,
    email: info.email,
    website: info.website,
    logo_url: info.logoUrl,
  });
  return data;
};

export function useClinicInfo() {
  const queryClient = useQueryClient();

  const { data: clinics = [], isLoading } = useQuery({
    queryKey: ["clinics"],
    queryFn: getClinics,
  });

  const firstClinic = clinics[0];
  const clinicInfo: ClinicInfo = firstClinic
    ? transformToClinicInfo(firstClinic)
    : {
        name: "",
        postalCode: "",
        address: "",
        phoneNumber: "",
      };

  const updateMutation = useMutation({
    mutationFn: ({ id, info }: { id: string; info: ClinicInfo }) =>
      updateClinic(id, info),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clinics"] });
    },
  });

  const updateClinicInfo = useCallback(
    (newInfo: ClinicInfo) => {
      if (!firstClinic) {
        toast.error("クリニック情報が取得できていません");
        return;
      }
      updateMutation.mutate(
        { id: String(firstClinic.id ?? 0), info: newInfo },
        {
          onSuccess: () => toast.success("病院情報を更新しました"),
          onError: () => toast.error("病院情報の保存に失敗しました"),
        }
      );
    },
    [firstClinic, updateMutation]
  );

  return {
    clinicInfo,
    updateClinicInfo,
    loading: isLoading,
  };
}
