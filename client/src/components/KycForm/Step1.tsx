import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { RadioGroup, RadioGroupItem } from "../ui/radio-group";
import { Field, FieldContent, FieldLabel, FieldTitle } from "../ui/field";
import { useCountries } from "@/hooks/useContries";
import { Separator } from "../ui/separator";
import FormNavigation from "./FormNavigation";
import { updateFormData } from "@/store/kyc-slice";
import { idTypes } from "@/config/idtypes";
import {
  initiateSessionSchema,
  InitiateSessionValues,
} from "@/utils/validation/kycSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";

export default function InitiateSession() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { data: countries, isLoading, isError } = useCountries();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<InitiateSessionValues>({
    resolver: zodResolver(initiateSessionSchema),
    defaultValues: { IDType: "national_id", country: "" },
  });

  const onSubmit = (values: InitiateSessionValues) => {
    dispatch(updateFormData(values));
    navigate("/verify/upload-docs");
  };

  return (
    <div className="py-8 px-6 max-w-lg">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-xl font-semibold text-foreground tracking-tight">
          Choose your verification document
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          You must carry an official government-issued ID
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)}>
        {/* Document type */}
        <div className="mb-8">
          <label className="text-sm font-medium text-foreground mb-3 block">
            Document type
          </label>
          <Controller
            name="IDType"
            control={control}
            render={({ field }) => (
              <RadioGroup
                value={field.value}
                onValueChange={field.onChange}
                className="flex flex-col gap-2"
              >
                {idTypes.map(({ value, label, icon }) => (
                  <FieldLabel key={value} htmlFor={value}>
                    <Field
                      orientation="horizontal"
                      className="flex items-center gap-3 px-4 py-3 rounded-lg bg-muted hover:bg-muted/70 cursor-pointer transition-colors"
                    >
                      <div className="size-8 rounded-md bg-background flex items-center justify-center shrink-0 text-muted-foreground">
                        {icon}
                      </div>
                      <FieldContent className="flex-1">
                        <FieldTitle className="text-sm font-medium">
                          {label}
                        </FieldTitle>
                      </FieldContent>
                      <RadioGroupItem value={value} id={value} />
                    </Field>
                  </FieldLabel>
                ))}
              </RadioGroup>
            )}
          />
          {errors.IDType && (
            <p className="text-xs text-destructive mt-2">
              {errors.IDType.message}
            </p>
          )}
        </div>

        <Separator className="mb-8" />

        {/* Country */}
        <div className="mb-8">
          <label className="text-sm font-medium text-foreground mb-1 block">
            Country of document
          </label>
          <p className="text-xs text-muted-foreground mb-3">
            Select the country that issued your document
          </p>
          <Controller
            name="country"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger className="w-full max-w-sm">
                  <SelectValue
                    placeholder={
                      isLoading
                        ? "Loading countries..."
                        : isError
                          ? "Failed to load"
                          : "Select a country"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {countries?.map((country) => (
                      <SelectItem key={country.cca2} value={country.cca2}>
                        <span className="flex items-center gap-2">
                          <img
                            src={country.flags.svg}
                            alt={country.name.common}
                            className="w-4 h-3 object-cover rounded-sm"
                          />
                          {country.name.common}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
          {errors.country && (
            <p className="text-xs text-destructive mt-2">
              {errors.country.message}
            </p>
          )}
        </div>

        <FormNavigation onNext={handleSubmit(onSubmit)} />
      </form>
    </div>
  );
}
