import { useAppSelector } from "@/store/hooks";

export default function ProgressBar() {
  const currentStep = useAppSelector((state) => state.createKyc.currentStep);
  const progress = (currentStep / 3) * 100;

  return (
    // <div className="max-w-4xl mx-auto px-6 py-8">
    //   <div className="relative mb-8">
    //     <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
    //       <div
    //         className="h-full bg-main transition-all duration-300 ease-out"
    //         style={{ width: `${progress}%` }}
    //       />
    //     </div>
    //   </div>

    // </div>

    <div className="px-6 py-4">
      <div className="h-1.5 bg-muted rounded-full overflow-hidden">
        <div
          className="h-full bg-primary transition-all duration-300 ease-out"
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
}
