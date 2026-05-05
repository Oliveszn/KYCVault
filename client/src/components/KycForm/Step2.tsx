import { useState } from "react";
import { Separator } from "../ui/separator";
import FormNavigation from "./FormNavigation";
import {
  uploadDocumentSchema,
  UploadDocumentValues,
} from "@/utils/validation/kycSchema";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { updateFormData } from "@/store/kyc-slice";
import { useNavigate, useParams } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { useUploadDocument } from "@/hooks/useDocument";
import { toast } from "sonner";
import { idTypes } from "@/config/idtypes";
import { useSession } from "@/hooks/useKyc";
import { KYCDocument } from "@/types/document";
import DocumentSideInput, { SideFile } from "./Documentsideinput";

type SideFiles = {
  front: SideFile | null;
  back: SideFile | null;
};

export default function UploadDocument() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { id: sessionId } = useParams<{ id: string }>();

  const { data: session } = useSession(sessionId!);
  const uploadDocument = useUploadDocument();
  const formData = useAppSelector((state) => state.createKyc.formData);

  const existingFront = session?.documents?.find((d) => d.side === "front");
  const existingBack = session?.documents?.find((d) => d.side === "back");

  const country = session?.country || formData.country;
  const idType = session?.id_type || formData.IDType;
  const selectedIdType = idTypes.find((t) => t.value === idType);

  const [sideFiles, setSideFiles] = useState<SideFiles>({
    front: null,
    back: null,
  });

  const {
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<UploadDocumentValues>({
    resolver: zodResolver(uploadDocumentSchema),
    //Premark as valid if docs alreay exists, so form doesnt block submisiion
    defaultValues: {
      front: existingFront ? (true as any) : undefined,
      back: existingBack ? (true as any) : undefined,
    },
  });

  // Called by DocumentSideInput when a new file is selected or captured
  const handleFileChange = (side: "front" | "back", file: File) => {
    const preview = URL.createObjectURL(file);
    setValue(side, file, { shouldValidate: true });
    setSideFiles((prev) => ({ ...prev, [side]: { file, preview } }));
  };

  // Called by DocumentSideInput when the user removes a freshly selected file
  const handleClear = (side: "front" | "back") => {
    setSideFiles((prev) => ({ ...prev, [side]: null }));
    setValue(side, undefined as any, { shouldValidate: true });
  };

  const onSubmit = async (values: UploadDocumentValues) => {
    if (!sessionId) return;

    const uploads: Promise<KYCDocument>[] = [];

    if (values.front && !existingFront) {
      uploads.push(
        uploadDocument.mutateAsync({
          sessionId,
          file: values.front,
          side: "front",
        }),
      );
    }
    if (values.back && !existingBack) {
      uploads.push(
        uploadDocument.mutateAsync({
          sessionId,
          file: values.back,
          side: "back",
        }),
      );
    }

    // if both sides already exists and the user didnt replace any, advance
    if (uploads.length === 0) {
      navigate(`/kyc/sessions/${sessionId}/face`);
      return;
    }

    try {
      await Promise.all(uploads);
      dispatch(
        updateFormData({
          documents: { front: values.front, back: values.back },
        }),
      );
      toast.success("Documents uploaded successfully");
      navigate(`/kyc/sessions/${sessionId}/face`);
    } catch (err: any) {
      const message = err?.response?.data?.message || "Upload failed";
      toast.error(message);
    }
  };

  return (
    <div className="py-8 px-6 max-w-lg">
      <div className="mb-8">
        <h1 className="text-xl font-semibold text-foreground tracking-tight">
          Prepare your document
        </h1>
        <p className="text-sm text-muted-foreground mt-1 leading-relaxed">
          You'll need to scan both sides of your ID. Make sure you capture a
          clear and complete image.
        </p>
      </div>

      <div className="bg-muted rounded-lg px-4 py-3 mb-8 flex flex-col gap-1.5">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Country of document</span>
          <span className="font-medium text-foreground">{country || "—"}</span>
        </div>
        <Separator />
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Document type</span>
          <span className="font-medium text-foreground">
            {selectedIdType?.label || idType?.replace(/_/g, " ") || "—"}
          </span>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="flex flex-col gap-6 mb-8">
          {(["front", "back"] as const).map((side) => (
            <DocumentSideInput
              key={side}
              side={side}
              file={sideFiles[side]}
              existingDoc={side === "front" ? existingFront : existingBack}
              error={errors[side]?.message}
              onFileChange={handleFileChange}
              onClear={handleClear}
            />
          ))}
        </div>

        <FormNavigation
          onNext={handleSubmit(onSubmit)}
          isSubmitting={uploadDocument.isPending}
        />
      </form>
    </div>
  );
}
