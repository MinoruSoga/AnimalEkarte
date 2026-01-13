import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useHospitalizations } from "./useHospitalizations";
import { useMasterItems } from "../../master/hooks/useMasterItems";
import { HospitalizationFilterStatus, HOSPITALIZATION_FILTER_STATUS } from "../constants";

export const useHospitalizationList = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<HospitalizationFilterStatus>(HOSPITALIZATION_FILTER_STATUS.ACTIVE);
  const [viewMode, setViewMode] = useState<"list" | "board">("board");
  
  const { data: filteredHospitalizations, movePet, isLoading } = useHospitalizations(searchTerm, statusFilter);
  const { data: cages } = useMasterItems("cage");

  const handleNavigateToForm = (id?: string) => {
    navigate(id ? `/hospitalization/${id}` : "/hospitalization/select-pet");
  };

  return {
    searchTerm,
    setSearchTerm,
    statusFilter,
    setStatusFilter,
    viewMode,
    setViewMode,
    filteredHospitalizations,
    cages,
    movePet,
    handleNavigateToForm,
    isLoading
  };
};
