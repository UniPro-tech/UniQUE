import Link from "next/link";
import { UniQUE_Logotype } from "@/components/brand/unique";

export default function SignupPage(props: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-muted p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link
          href="#"
          className="flex items-center gap-2 self-center font-medium text-xl h-16"
        >
          <UniQUE_Logotype className="h-16" />
        </Link>
        {props.children}
      </div>
    </div>
  );
}
