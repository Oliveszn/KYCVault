import { Country } from "@/types/countries";

export const fetchCountries = async (): Promise<Country[]> => {
  const res = await fetch(
    "https://restcountries.com/v3.1/all?fields=name,cca2,flags",
  );
  if (!res.ok) throw new Error("Failed to fetch countries");
  const data: Country[] = await res.json();
  return data.sort((a, b) => a.name.common.localeCompare(b.name.common));
};
