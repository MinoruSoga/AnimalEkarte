import { Route, Routes } from "react-router-dom";
import { HospitalizationList } from "./HospitalizationList";
import { HospitalizationForm } from "./HospitalizationForm";
import { HospitalizationPetSelection } from "./HospitalizationPetSelection";
import { HospitalizationDetail } from "./HospitalizationDetail";

export const HospitalizationRoutes = () => {
  return (
    <Routes>
      <Route path="" element={<HospitalizationList />} />
      <Route path="select-pet" element={<HospitalizationPetSelection />} />
      <Route path="new" element={<HospitalizationForm />} />
      <Route path=":id" element={<HospitalizationDetail />} />
      <Route path=":id/edit" element={<HospitalizationForm />} />
    </Routes>
  );
};
