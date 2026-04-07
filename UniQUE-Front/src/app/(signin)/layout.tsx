"use client";
import CssBaseline from "@mui/material/CssBaseline";
import Stack from "@mui/material/Stack";
import type * as React from "react";
import AppTheme from "@/components/mui-template/shared-theme/AppTheme";

export default function AuthLayout(props: { children: React.ReactNode }) {
  const { children } = props;
  return (
    <html lang="ja">
      <body>
        <AppTheme>
          <CssBaseline enableColorScheme />
          <Stack
            direction="column"
            component="main"
            justifyContent={"center"}
            justifyItems={"center"}
            alignItems={"center"}
            sx={[
              {
                justifyContent: "center",
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
                p: { xs: 2, sm: 4 },
                m: "auto",
              }}
            >
              {children}
            </Stack>
          </Stack>
        </AppTheme>
      </body>
    </html>
  );
}
