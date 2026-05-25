import { InitializationForm } from "@/components/initialization-form";
import { redirect } from "next/navigation";
import prisma from "@/lib/prisma";

export default async function InitializePage() {
  const userCount = await prisma.user.count();
  const initialized = userCount > 0;

  if (initialized) {
    redirect("/");
  }

  return <InitializationForm />;
}
