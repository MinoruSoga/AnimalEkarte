import { useState } from "react";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import { useMedicalRecordAddenda } from "../../hooks/use-medical-record-addenda";
import { isMedicalRecordFinalizedStatus } from "../../lib/medical-record-lock";
import { AddendumItem } from "./AddendumItem";
import { AddendumModal } from "./AddendumModal";

interface MedicalRecordAddendaProps {
  medicalRecordId: string;
  canEdit: boolean;
  recordStatus: string;
}

export function MedicalRecordAddenda({
  medicalRecordId,
  canEdit,
  recordStatus,
}: MedicalRecordAddendaProps) {
  if (!isMedicalRecordFinalizedStatus(recordStatus)) return null;

  return <MedicalRecordAddendaContent medicalRecordId={medicalRecordId} canEdit={canEdit} />;
}

interface ContentProps {
  medicalRecordId: string;
  canEdit: boolean;
}

function MedicalRecordAddendaContent({ medicalRecordId, canEdit }: ContentProps) {
  const { data: addenda = [], isLoading } = useMedicalRecordAddenda(medicalRecordId);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const sorted = [...addenda].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  );

  return (
    <section
      aria-label="追記一覧"
      className={`mx-4 mb-6 border-t ${C.borderLight} pt-4`}
      data-testid="medical-record-addenda"
    >
      <div className="flex items-center justify-between mb-3">
        <h2 className={`text-base font-medium ${C.text}`}>追記</h2>
        {canEdit ? (
          <Button
            type="button"
            variant="outline"
            className="h-8 text-sm px-3"
            onClick={() => setIsModalOpen(true)}
            data-testid="addendum-open-button"
          >
            追記する
          </Button>
        ) : null}
      </div>

      {isLoading ? (
        <p className={`text-sm ${C.text60}`}>読み込み中...</p>
      ) : sorted.length === 0 ? (
        <p className={`text-sm ${C.text60}`} data-testid="addenda-empty">
          追記はありません
        </p>
      ) : (
        <ul className="space-y-3 list-none">
          {sorted.map((addendum) => (
            <li key={addendum.id}>
              <AddendumItem addendum={addendum} />
            </li>
          ))}
        </ul>
      )}

      <AddendumModal
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        medicalRecordId={medicalRecordId}
      />
    </section>
  );
}
