// React/Framework
import { useEffect } from "react";
import { useForm } from "react-hook-form";

// External
import { Building2, Save } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

// Feature API
import { useListClinics, useUpdateClinic } from "@/features/hospital-settings/api/clinics";

// Types
import type { UpdateClinicRequest } from "@/features/hospital-settings/api/clinics";

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
}

export function ClinicSettings() {
  const { data: clinics = [], isLoading } = useListClinics();
  const updateMutation = useUpdateClinic();

  const clinic = clinics[0] ?? null;

  const { register, handleSubmit, reset, formState: { isDirty } } = useForm<ClinicFormData>({
    defaultValues: {
      name: "",
      postalCode: "",
      address: "",
      phoneNumber: "",
      faxNumber: "",
      registrationNumber: "",
      directorName: "",
      email: "",
      website: "",
    },
  });

  useEffect(() => {
    if (clinic) {
      reset({
        name: clinic.name,
        postalCode: clinic.postalCode,
        address: clinic.address,
        phoneNumber: clinic.phoneNumber,
        faxNumber: clinic.faxNumber,
        registrationNumber: clinic.registrationNumber,
        directorName: clinic.directorName,
        email: clinic.email,
        website: clinic.website,
      });
    }
  }, [clinic, reset]);

  const onSubmit = (data: ClinicFormData) => {
    if (!clinic) return;
    const req: UpdateClinicRequest = {
      name: data.name,
      postal_code: data.postalCode,
      address: data.address,
      phone_number: data.phoneNumber,
      fax_number: data.faxNumber,
      registration_number: data.registrationNumber,
      director_name: data.directorName,
      email: data.email,
      website: data.website,
    };
    updateMutation.mutate(
      { id: clinic.id, req },
      { onSuccess: () => reset(data) },
    );
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <PageLayout title="病院情報設定" icon={<Building2 className="h-6 w-6" />} maxWidth="max-w-4xl">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Card>
          <CardHeader>
            <CardTitle>基本情報</CardTitle>
            <CardDescription>
              領収書や明細書、処方箋などに印字される病院情報です。
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-2">
              <Label htmlFor="name">病院名 <span className="text-red-500">*</span></Label>
              <Input id="name" {...register("name", { required: true })} placeholder="例: 八王子院" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="postalCode">郵便番号</Label>
                <Input id="postalCode" {...register("postalCode")} placeholder="例: 100-0001" />
              </div>
              <div className="grid gap-2 md:col-span-2">
                <Label htmlFor="address">住所</Label>
                <Input id="address" {...register("address")} placeholder="例: 東京都千代田区千代田1-1" />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="phoneNumber">電話番号</Label>
                <Input id="phoneNumber" {...register("phoneNumber")} placeholder="例: 03-1234-5678" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="faxNumber">FAX番号</Label>
                <Input id="faxNumber" {...register("faxNumber")} placeholder="例: 03-1234-5679" />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="registrationNumber">登録番号</Label>
              <Input id="registrationNumber" {...register("registrationNumber")} placeholder="例: 東京都獣医師会 第12345号" />
              <p className="text-sm text-gray-500">領収書などに記載される登録番号です。</p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="directorName">院長名</Label>
              <Input id="directorName" {...register("directorName")} placeholder="例: 山田 太郎" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="email">メールアドレス</Label>
                <Input id="email" type="email" {...register("email")} placeholder="info@example.com" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="website">WebサイトURL</Label>
                <Input id="website" {...register("website")} placeholder="https://example.com" />
              </div>
            </div>
          </CardContent>
          <CardFooter className="bg-gray-50 border-t p-4 flex justify-end">
            <Button type="submit" disabled={!isDirty || updateMutation.isPending}>
              <Save className="mr-2 h-4 w-4" />
              設定を保存
            </Button>
          </CardFooter>
        </Card>
      </form>
    </PageLayout>
  );
}
