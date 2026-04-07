"use client";

import { AuthorizationPageMode, useAuthPageMode } from "../Client";
import MigrationCard from "./MobileCards/MigrationCard";
import MultiFactorCard from "./MobileCards/MultiFactorCard";
import SignInCard from "./MobileCards/SignInCard";
import SignUpCard from "./MobileCards/SignUpCard";

export default function AuthCardMobile() {
  const mode = useAuthPageMode();

  switch (mode) {
    case AuthorizationPageMode.SignIn:
      return <SignInCard />;
    case AuthorizationPageMode.SignUp:
      return <SignUpCard />;
    case AuthorizationPageMode.Migration:
      return <MigrationCard />;
    case AuthorizationPageMode.MFA:
      return <MultiFactorCard />;
    default:
      return null;
  }
}
