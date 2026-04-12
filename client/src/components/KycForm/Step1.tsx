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
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldLabel,
  FieldTitle,
} from "../ui/field";
import {
  AppWindow,
  HomeIcon,
  IdCardIcon,
  RectangleEllipsis,
} from "lucide-react";
import { useCountries } from "@/hooks/useContries";
import { Button } from "../ui/button";
import { Separator } from "../ui/separator";
import FormNavigation from "./FormNavigation";
import { nextStep } from "@/store/kyc-slice";

export default function InitiateSession() {
  const dispatch = useAppDispatch();
  const formData = useAppSelector((state) => state.createKyc.formData);
  const { data: countries, isLoading, isError } = useCountries();
  const currentStep = useAppSelector((state) => state.createKyc.currentStep);

  const idTypes = [
    {
      value: "national_id",
      label: "National ID Card",
      icon: <IdCardIcon />,
    },
    {
      value: "passport",
      label: "Passport",
      icon: <RectangleEllipsis />,
    },
    {
      value: "drivers_license",
      label: "Driver's License",
      icon: <AppWindow />,
    },
    {
      value: "residence_permit",
      label: "Residence Permit",
      icon: <HomeIcon />,
    },
  ];
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

      <form onSubmit={(e) => e.preventDefault()}>
        {/* Document type */}
        <div className="mb-8">
          <label className="text-sm font-medium text-foreground mb-3 block">
            Document type
          </label>
          <RadioGroup
            defaultValue="national_id"
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
          <Select>
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
                {isLoading && (
                  <SelectItem value="loading" disabled>
                    Loading...
                  </SelectItem>
                )}
                {isError && (
                  <SelectItem value="error" disabled>
                    Failed to load countries
                  </SelectItem>
                )}
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
        </div>

        {/* Submit */}
        {/* <Button type="submit" className="w-full max-w-sm">
          Continue
        </Button> */}

        <FormNavigation onNext={() => dispatch(nextStep())} />
      </form>
    </div>
  );
}
