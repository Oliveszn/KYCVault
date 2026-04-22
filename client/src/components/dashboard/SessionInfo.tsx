import { AdminSessionResponse } from "@/types/kyc";

type Props = {
  session: AdminSessionResponse;
};

export default function SessionInfo({ session }: Props) {
  const items = [
    { label: "Country", value: session.country },
    { label: "ID Type", value: session.id_type?.replace("_", " ") },
    { label: "Attempt", value: `#${session.attempt_number}` },
  ];

  return (
    <div className="grid grid-cols-3 gap-4">
      {items.map(({ label, value }) => (
        <div key={label} className="border rounded-xl p-4">
          <p className="text-xs text-muted-foreground">{label}</p>
          <p className="font-medium capitalize">{value}</p>
        </div>
      ))}
    </div>
  );
}
