// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router";

// External
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";

// Internal
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout";
import { Settings } from "./Settings";
import { C } from "@/lib/design-tokens";

const TABS = [
  { value: "diagnosisCategory", label: "カテゴリ" },
  { value: "diagnosisName", label: "診断名" },
] as const;

export function DiagnosisSettings() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<string>("diagnosisCategory");

  return (
    <PageLayout
      title="診断マスタ"
      icon={<FolderTree className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList
            className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
          >
            {TABS.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
                  data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
                  data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
          {TABS.map((tab) => (
            <TabsContent key={tab.value} value={tab.value} className="mt-4">
              <Settings category={tab.value} embedded />
            </TabsContent>
          ))}
        </Tabs>
      </div>
    </PageLayout>
  );
}
