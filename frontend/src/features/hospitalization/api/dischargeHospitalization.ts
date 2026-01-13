import { Hospitalization } from "../types";
import { format } from "date-fns";
import { updateHospitalization } from "./updateHospitalization";

export const dischargeHospitalization = async (id: string): Promise<Hospitalization> => {
    return updateHospitalization(id, {
        status: "退院済",
        endDate: format(new Date(), "yyyy-MM-dd")
    });
};
