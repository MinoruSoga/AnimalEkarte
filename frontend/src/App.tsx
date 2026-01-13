import { DndProvider } from "react-dnd";
import { TouchBackend } from "react-dnd-touch-backend";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";

import Sidebar from "./components/Sidebar";
import { Dashboard } from "./features/dashboard/routes";
import { ReservationManagement } from "./features/reservations/routes";

// Features
import { OwnersList, OwnerForm } from "./features/owners/routes";
import { MedicalRecords, MedicalRecordForm, MedicalRecordPetSelection } from "./features/medical-records/routes";
import { HospitalizationRoutes } from "./features/hospitalization/routes";
import { TrimmingList, TrimmingForm, TrimmingPetSelection } from "./features/trimming/routes";
import { Examinations, ExaminationForm, ExaminationPetSelection } from "./features/examinations/routes";
import { Accounting, AccountingForm, AccountingPetSelection } from "./features/accounting/routes";
import { VaccinationList, VaccinationForm, VaccinationPetSelection } from "./features/vaccinations/routes";
import { Settings } from "./features/master/routes";
import { ClinicSettings } from "./features/clinic/routes";

import { Toaster } from "./components/ui/sonner";

const Layout = () => {
  return (
    <div className="flex w-full h-full bg-[#F7F6F3] overflow-hidden">
      <Sidebar />
      <div className="flex-1 overflow-hidden h-full">
        <Outlet />
      </div>
      <Toaster />
    </div>
  );
};

export default function App() {
  return (
    <DndProvider backend={TouchBackend} options={{ enableMouseEvents: true }}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Dashboard />} />
            
            {/* Owners Routes */}
            <Route path="/owners" element={<OwnersList />} />
            <Route path="/owners/new" element={<OwnerForm />} />
            <Route path="/owners/:id" element={<OwnerForm />} />
            
            <Route path="/reservations" element={<ReservationManagement />} />
            
            {/* Medical Records Routes */}
            <Route path="/medical-records" element={<MedicalRecords />} />
            <Route path="/medical-records/select-pet" element={<MedicalRecordPetSelection />} />
            <Route path="/medical-records/new" element={<MedicalRecordForm />} />
            <Route path="/medical-records/:id" element={<MedicalRecordForm />} />
            
            {/* Hospitalization Routes */}
            <Route path="/hospitalization/*" element={<HospitalizationRoutes />} />
            
            {/* Trimming Routes */}
            <Route path="/trimming" element={<TrimmingList />} />
            <Route path="/trimming/select-pet" element={<TrimmingPetSelection />} />
            <Route path="/trimming/new" element={<TrimmingForm />} />
            <Route path="/trimming/:id" element={<TrimmingForm />} />
            
            {/* Examination Routes */}
            <Route path="/examinations" element={<Examinations />} />
            <Route path="/examinations/select-pet" element={<ExaminationPetSelection />} />
            <Route path="/examinations/new" element={<ExaminationForm />} />
            <Route path="/examinations/:id" element={<ExaminationForm />} />
            
            {/* Accounting Routes */}
            <Route path="/accounting" element={<Accounting />} />
            <Route path="/accounting/select-pet" element={<AccountingPetSelection />} />
            <Route path="/accounting/new" element={<AccountingForm />} />
            <Route path="/accounting/:id" element={<AccountingForm />} />
            
            {/* Vaccination Routes */}
            <Route path="/vaccinations" element={<VaccinationList />} />
            <Route path="/vaccinations/select-pet" element={<VaccinationPetSelection />} />
            <Route path="/vaccinations/new" element={<VaccinationForm />} />
            <Route path="/vaccinations/:id" element={<VaccinationForm />} />

            {/* Settings Routes */}
            <Route path="/settings/clinic" element={<ClinicSettings />} />
            <Route path="/settings/service-type" element={<Settings category="serviceType" />} />
            <Route path="/settings/examination" element={<Settings category="examination" />} />
            <Route path="/settings/vaccine" element={<Settings category="vaccine" />} />
            <Route path="/settings/medicine" element={<Settings category="medicine" />} />
            <Route path="/settings/staff" element={<Settings category="staff" />} />
            <Route path="/settings/insurance" element={<Settings category="insurance" />} />
            <Route path="/settings/consultation" element={<Settings category="consultation" />} />
            <Route path="/settings/procedure" element={<Settings category="procedure" />} />
            <Route path="/settings/hospitalization" element={<Settings category="hospitalization" />} />
            <Route path="/settings/cage" element={<Settings category="cage" />} />
            
            {/* Fallback */}
            <Route path="*" element={
              <div className="flex-1 p-5 flex items-center justify-center">
                <p className="text-muted-foreground">ページが見つかりません</p>
              </div>
            } />
          </Route>
        </Routes>
      </BrowserRouter>
    </DndProvider>
  );
}
