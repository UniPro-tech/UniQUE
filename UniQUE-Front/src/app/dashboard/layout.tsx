import { AppRouterCacheProvider } from "@mui/material-nextjs/v15-appRouter";
import { SnackbarProvider } from "notistack";
import { Session } from "@/classes/Session";
import BirthdateGuard from "@/components/BirthdateGuard";
import Drawer from "@/components/drawer";

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const session = await Session.getCurrent();
  const user = session ? await session.getUser() : null;
  const roles = user
    ? (await user.getRoles()).map((role) => role.toJson())
    : undefined;
  const birthdate = user?.profile.birthdate ? user.profile.birthdate : null;
  const mustSetBirthdate = Boolean(!birthdate);
  return (
    <html lang="ja">
      <body className={`antialiased`}>
        <AppRouterCacheProvider options={{ enableCssLayer: true }}>
          <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
            <Drawer user={user?.toJson() || null} userRoles={roles}>
              {children}
            </Drawer>
            <BirthdateGuard
              mustSetBirthdate={mustSetBirthdate}
              initialBirthdate={birthdate || ""}
            />
          </SnackbarProvider>
        </AppRouterCacheProvider>
      </body>
    </html>
  );
}
