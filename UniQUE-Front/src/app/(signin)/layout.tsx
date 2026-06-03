"use client";
import CssBaseline from "@mui/material/CssBaseline";
import Stack from "@mui/material/Stack";
import { SnackbarProvider } from "notistack";
import type * as React from "react";
import AppTheme from "@/components/mui-template/shared-theme/AppTheme";

export default function AuthLayout(props: { children: React.ReactNode }) {
  const { children } = props;
  return (
    <html lang="ja">
      <body>
        <AppTheme>
          <CssBaseline enableColorScheme />
          <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
            <Stack
              direction="column"
              sx={[
                {
                  alignItems: "center",
                  justifyContent: "center",
                  justifyItems: "center",
                  minHeight: "100vh",
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
                  p: { xs: 0, sm: 4 },
                  m: "auto",
                }}
              >
                {children}
              </Stack>
            </Stack>
          </SnackbarProvider>
        </AppTheme>
      </body>
    </html>
  );
}
