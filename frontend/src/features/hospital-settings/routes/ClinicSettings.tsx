// React/Framework
import { useEffect } from "react";

// External
import { Building2, Save } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

// Feature API
import { useListClinics, useUpdateClinic } from "@/features/hospital-settings/api/clinics";

// Feature hooks
import { useClinicSettingsForm } from "../hooks/useClinicSettingsForm";

export function ClinicSettings() {
  const { data: clinics = [], isLoading } = useListClinics();
  const updateMutation = useUpdateClinic();

  const clinic = clinics[0] ?? null;

  const {
    formData,
    handleInputChange,
    isDirty,
    isSavePending,
    startSaveTransition,
    resetToClinic,
    toRequest,
  } = useClinicSettingsForm(clinic);

  useEffect(() => {
    if (clinic) {
      resetToClinic(clinic);
    }
  }, [clinic, resetToClinic]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!clinic) return;
    startSaveTransition(async () => {
      await updateMutation.mutateAsync({ id: clinic.id, req: toRequest() });
      if (clinic) resetToClinic(clinic);
    });
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <PageLayout
      title="病院情報設定"
      icon={<Building2 className="h-6 w-6" />}
      maxWidth="max-w-4xl"
    >
      <form onSubmit={handleSubmit}>
        <Card>
          <CardHeader>
            <CardTitle>基本情報</CardTitle>
            <CardDescription>
              領収書や明細書、処方箋などに印字される病院情報です。
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-2">
              <Label htmlFor="name">
                病院名 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => handleInputChange("name", e.target.value)}
                placeholder="例: 八王子院"
                required
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="postalCode">郵便番号</Label>
                <Input
                  id="postalCode"
                  value={formData.postalCode}
                  onChange={(e) =>
                    handleInputChange("postalCode", e.target.value)
                  }
                  placeholder="例: 100-0001"
                />
              </div>
              <div className="grid gap-2 md:col-span-2">
                <Label htmlFor="address">住所</Label>
                <Input
                  id="address"
                  value={formData.address}
                  onChange={(e) =>
                    handleInputChange("address", e.target.value)
                  }
                  placeholder="例: 東京都千代田区千代田1-1"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="phoneNumber">電話番号</Label>
                <Input
                  id="phoneNumber"
                  value={formData.phoneNumber}
                  onChange={(e) =>
                    handleInputChange("phoneNumber", e.target.value)
                  }
                  placeholder="例: 03-1234-5678"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="faxNumber">FAX番号</Label>
                <Input
                  id="faxNumber"
                  value={formData.faxNumber}
                  onChange={(e) =>
                    handleInputChange("faxNumber", e.target.value)
                  }
                  placeholder="例: 03-1234-5679"
                />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="registrationNumber">登録番号</Label>
              <Input
                id="registrationNumber"
                value={formData.registrationNumber}
                onChange={(e) =>
                  handleInputChange("registrationNumber", e.target.value)
                }
                placeholder="例: 東京都獣医師会 第12345号"
              />
              <p className="text-sm text-gray-500">
                領収書などに記載される登録番号です。
              </p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="directorName">院長名</Label>
              <Input
                id="directorName"
                value={formData.directorName}
                onChange={(e) =>
                  handleInputChange("directorName", e.target.value)
                }
                placeholder="例: 山田 太郎"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="email">メールアドレス</Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) =>
                    handleInputChange("email", e.target.value)
                  }
                  placeholder="info@example.com"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="website">WebサイトURL</Label>
                <Input
                  id="website"
                  value={formData.website}
                  onChange={(e) =>
                    handleInputChange("website", e.target.value)
                  }
                  placeholder="https://example.com"
                />
              </div>
            </div>
          </CardContent>
          <CardFooter className="bg-gray-50 border-t p-4 flex justify-end">
            <Button type="submit" disabled={!isDirty || isSavePending}>
              <Save className="mr-2 h-4 w-4" />
              設定を保存
            </Button>
          </CardFooter>
        </Card>
      </form>
    </PageLayout>
  );
}
