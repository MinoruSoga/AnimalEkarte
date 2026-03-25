import { useState, useCallback, useTransition } from "react";

import type { TransformedClinic } from "../api/transforms";
import type { UpdateClinicRequest } from "../api/clinics";

interface ClinicFormData {
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber: string;
  registrationNumber: string;
  directorName: string;
  email: string;
  website: string;
  standardTaxRate: number;
  reducedTaxRate: number;
}

const DEFAULT_FORM: ClinicFormData = {
  name: "",
  postalCode: "",
  address: "",
  phoneNumber: "",
  faxNumber: "",
  registrationNumber: "",
  directorName: "",
  email: "",
  website: "",
  standardTaxRate: 0.1,
  reducedTaxRate: 0.08,
};

function clinicToFormData(clinic: TransformedClinic): ClinicFormData {
  return {
    name: clinic.name,
    postalCode: clinic.postalCode,
    address: clinic.address,
    phoneNumber: clinic.phoneNumber,
    faxNumber: clinic.faxNumber,
    registrationNumber: clinic.registrationNumber,
    directorName: clinic.directorName,
    email: clinic.email,
    website: clinic.website,
    standardTaxRate: clinic.standardTaxRate,
    reducedTaxRate: clinic.reducedTaxRate,
  };
}

function formDataToRequest(data: ClinicFormData): UpdateClinicRequest {
  return {
    name: data.name,
    postal_code: data.postalCode,
    address: data.address,
    phone_number: data.phoneNumber,
    fax_number: data.faxNumber,
    registration_number: data.registrationNumber,
    director_name: data.directorName,
    email: data.email,
    website: data.website,
    standard_tax_rate: data.standardTaxRate,
    reduced_tax_rate: data.reducedTaxRate,
  };
}

export function useClinicSettingsForm(clinic: TransformedClinic | null) {
  const [formData, setFormData] = useState<ClinicFormData>(() =>
    clinic ? clinicToFormData(clinic) : DEFAULT_FORM,
  );
  const [initialData, setInitialData] = useState<ClinicFormData>(() =>
    clinic ? clinicToFormData(clinic) : DEFAULT_FORM,
  );
  const [isSavePending, startSaveTransition] = useTransition();

  const isDirty = JSON.stringify(formData) !== JSON.stringify(initialData);

  const handleInputChange = useCallback(
    (field: keyof ClinicFormData, value: string) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleTaxRateChange = useCallback(
    (field: "standardTaxRate" | "reducedTaxRate", value: number) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const resetToClinic = useCallback((c: TransformedClinic) => {
    const data = clinicToFormData(c);
    setFormData(data);
    setInitialData(data);
  }, []);

  const toRequest = useCallback(
    () => formDataToRequest(formData),
    [formData],
  );

  return {
    formData,
    handleInputChange,
    handleTaxRateChange,
    isDirty,
    isSavePending,
    startSaveTransition,
    resetToClinic,
    toRequest,
  };
}
