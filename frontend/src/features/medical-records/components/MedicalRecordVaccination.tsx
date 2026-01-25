// React/Framework
import React, { useState } from "react";

// Relative
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";

export function MedicalRecordVaccination() {
  const [vaccineName, setVaccineName] = useState("esophagitis");
  const [date, setDate] = useState("");
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleType] = useState("4weeks");
  const [nextDate, setNextDate] = useState("");
  const [remarks, setRemarks] = useState("");

  // Mock Data for History
  const historyItems = [
    { id: 1, name: "フィラリア薬", date: "24/4/6", next: "24/4/6" },
    { id: 2, name: "ノミダニ予防薬", date: "25/6/15", next: "-" },
    { id: 3, name: "ノミダニ予防薬", date: "25/1/15", next: "25/6/15" },
    { id: 4, name: "ノミダニ予防薬", date: "25/1/15", next: "25/6/15" },
    { id: 5, name: "狂犬病ワクチン", date: "25/1/20", next: "25/6/20" },
    { id: 6, name: "ジステンパーワクチン", date: "25/1/25", next: "25/6/25" },
    { id: 7, name: "パルボウイルスワクチン", date: "25/1/30", next: "25/6/30" },
    { id: 8, name: "猫ウイルス性鼻気管炎ワクチン", date: "25/1/35", next: "25/6/35" },
    { id: 9, name: "猫クラシウイルスワクチン", date: "25/1/40", next: "25/6/40" },
    { id: 10, name: "レプトスピラワクチン", date: "25/1/45", next: "25/6/45" },
    { id: 11, name: "コロナウイルスワクチン", date: "25/1/50", next: "25/6/50" },
    { id: 12, name: "フェライン・レトロウイルスワクチン", date: "25/1/55", next: "25/6/55" },
    { id: 13, name: "パルボウイルスキャリアワクチン", date: "25/1/60", next: "25/6/60" },
  ];

  return (
    <div className="grid grid-cols-12 gap-4 h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20 pr-1">
      {/* Left Column: Form */}
      <VaccinationForm
        vaccineName={vaccineName}
        setVaccineName={setVaccineName}
        date={date}
        setDate={setDate}
        supplemental={supplemental}
        setSupplemental={setSupplemental}
        lot1={lot1}
        setLot1={setLot1}
        lot2={lot2}
        setLot2={setLot2}
        lot3={lot3}
        setLot3={setLot3}
        lot4={lot4}
        setLot4={setLot4}
        nextScheduleType={nextScheduleType}
        setNextScheduleType={setNextScheduleType}
        nextDate={nextDate}
        setNextDate={setNextDate}
        remarks={remarks}
        setRemarks={setRemarks}
      />

      {/* Right Column: History */}
      <VaccinationHistory historyItems={historyItems} />
    </div>
  );
}
