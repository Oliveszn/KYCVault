import { fetchCountries } from "@/lib/countries";
import { Country } from "@/types/countries";
import { useQuery } from "@tanstack/react-query";

export const useCountries = () => {
  return useQuery<Country[]>({
    queryKey: ["countries"],
    queryFn: fetchCountries,
    staleTime: Infinity,
    gcTime: 1000 * 60 * 60,
  });
};
