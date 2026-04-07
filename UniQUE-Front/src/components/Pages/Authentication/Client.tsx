"use client";

import CssBaseline from "@mui/material/CssBaseline";
import Stack from "@mui/material/Stack";
import {
  createContext,
  type Dispatch,
  type SetStateAction,
  useContext,
  useState,
} from "react";
import AppTheme from "@/components/mui-template/shared-theme/AppTheme";
import { isMobile } from "@/libs/window";
import AuthCard from "./components/AuthCard";
import AuthCardMobile from "./components/AuthCardMobile";
import Content from "./components/Content";

export enum AuthorizationPageMode {
  SignIn = "signin",
  SignUp = "signup",
  Migration = "migration",
  MFA = "mfa",
}

export interface SignInSideProps {
  initFormState?: AuthorizationFormState;
  mode?: AuthorizationPageMode;
  redirectTo?: string;
}

function SignInPC() {
  return (
    <Stack
      direction="column"
      component="main"
      sx={[
        {
          justifyContent: "center",
          height: "calc((1 - var(--template-frame-height, 0)) * 100%)",
          marginTop: "max(40px - var(--template-frame-height, 0px), 0px)",
          minHeight: "100%",
        },
        (theme) => ({
          "&::before": {
            content: '""',
            display: "block",
            position: "absolute",
            zIndex: -1,
            inset: 0,
            backgroundImage:
              "radial-gradient(ellipse at 50% 50%, hsl(210, 100%, 97%), hsl(0, 0%, 100%))",
            backgroundRepeat: "no-repeat",
            ...theme.applyStyles("dark", {
              backgroundImage:
                "radial-gradient(at 50% 50%, hsla(210, 100%, 16%, 0.5), hsl(220, 30%, 5%))",
            }),
          },
        }),
      ]}
    >
      <Stack
        direction={{ xs: "column-reverse", md: "row" }}
        sx={{
          justifyContent: "center",
          gap: { xs: 6, sm: 12 },
          p: 2,
          mx: "auto",
        }}
      >
        <Stack
          direction={{ xs: "column-reverse", md: "row" }}
          sx={{
            justifyContent: "center",
            gap: { xs: 6, sm: 12 },
            p: { xs: 2, sm: 4 },
            m: "auto",
          }}
        >
          <Content />
          <AuthCard />
        </Stack>
      </Stack>
    </Stack>
  );
}

export interface AuthorizationFormState {
  name?: string;
  username?: string;
  email?: string;
  externalEmail?: string;
  agreeToTerms?: boolean;
  rememberMe?: boolean;
}

export const InitialFormStateContext = createContext<
  AuthorizationFormState | undefined
>(undefined);
export const SetInitialFormStateContext = createContext<
  Dispatch<SetStateAction<AuthorizationFormState | undefined>> | undefined
>(undefined);
export const AuthPageModeContext = createContext(AuthorizationPageMode.SignIn);
export const SetAuthPageModeContext = createContext<
  Dispatch<SetStateAction<AuthorizationPageMode>> | undefined
>(undefined);
export const RedirectToContext = createContext<string | undefined>(undefined);

export const useInitialFormState = () => useContext(InitialFormStateContext);
export const useSetInitialFormState = () =>
  useContext(SetInitialFormStateContext);
export const useRedirectTo = () => useContext(RedirectToContext);
export const useAuthPageMode = () => useContext(AuthPageModeContext);
export const useSetAuthPageMode = () => useContext(SetAuthPageModeContext);

export default function AuthenticationPageClient(props: SignInSideProps) {
  const [mode, setMode] = useState(props.mode || AuthorizationPageMode.SignIn);
  const [initState, setInitState] = useState(props.initFormState || undefined);
  const redirectTo = props.redirectTo;
  return (
    <AppTheme>
      <CssBaseline enableColorScheme />
      {isMobile() ? (
        <InitialFormStateContext.Provider value={initState}>
          <SetInitialFormStateContext.Provider value={setInitState}>
            <AuthPageModeContext.Provider value={mode}>
              <SetAuthPageModeContext.Provider value={setMode}>
                <RedirectToContext.Provider value={redirectTo}>
                  <AuthCardMobile />
                </RedirectToContext.Provider>
              </SetAuthPageModeContext.Provider>
            </AuthPageModeContext.Provider>
          </SetInitialFormStateContext.Provider>
        </InitialFormStateContext.Provider>
      ) : (
        <InitialFormStateContext.Provider value={initState}>
          <SetInitialFormStateContext.Provider value={setInitState}>
            <AuthPageModeContext.Provider value={mode}>
              <SetAuthPageModeContext.Provider value={setMode}>
                <RedirectToContext.Provider value={redirectTo}>
                  <SignInPC />
                </RedirectToContext.Provider>
              </SetAuthPageModeContext.Provider>
            </AuthPageModeContext.Provider>
          </SetInitialFormStateContext.Provider>
        </InitialFormStateContext.Provider>
      )}
    </AppTheme>
  );
}
