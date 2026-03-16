import { MainLayout } from "@/components/main-layout";
import { ConsoleLayout } from "@/components/console-layout";
import { ConsoleGuard } from "@/components/console-guard";

export default function ConsoleRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <MainLayout>
      <ConsoleGuard>
        <ConsoleLayout>{children}</ConsoleLayout>
      </ConsoleGuard>
    </MainLayout>
  );
}
