import {
  AppWindow,
  HomeIcon,
  IdCardIcon,
  RectangleEllipsis,
} from "lucide-react";
export const idTypes = [
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
