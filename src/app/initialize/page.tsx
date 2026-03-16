import { checkSystemInitialized } from "@/app/actions/system";
import { InitializationForm } from "@/components/initialization-form";
import { redirect } from "next/navigation";

export default async function InitializePage() {
  const { initialized } = await checkSystemInitialized();

  if (initialized) {
    redirect("/");
  }

  return <InitializationForm />;
}
