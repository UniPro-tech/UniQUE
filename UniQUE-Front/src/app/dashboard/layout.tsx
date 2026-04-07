import { AppRouterCacheProvider } from "@mui/material-nextjs/v15-appRouter";
import { Session } from "@/classes/Session";
import BirthdateGuard from "@/components/BirthdateGuard";
import Drawer from "@/components/drawer";
import { generateCSRFToken } from "@/libs/csrf";

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
  const csrfToken = user ? generateCSRFToken(user.id) : "";
  return (
    <html lang="ja">
      <body className={`antialiased`}>
        <AppRouterCacheProvider options={{ enableCssLayer: true }}>
          <Drawer user={user?.toJson() || null} userRoles={roles}>
            {children}
          </Drawer>
        </AppRouterCacheProvider>
        <BirthdateGuard
          mustSetBirthdate={mustSetBirthdate}
          csrfToken={csrfToken}
          initialBirthdate={birthdate || ""}
        />
      </body>
    </html>
  );
}
