"use client";

import { useAuthPageMode } from "../Client";
import MigrationCard from "./Cards/MigrationCard";
import MultiFactorCard from "./Cards/MultiFactorCard";
import SignInCard from "./Cards/SignInCard";
import SignUpCard from "./Cards/SignUpCard";

export default function AuthCard() {
  const mode = useAuthPageMode();

  switch (mode) {
    case "signin":
      return <SignInCard />;
    case "signup":
      return <SignUpCard />;
    case "migration":
      return <MigrationCard />;
    case "mfa":
      return <MultiFactorCard />;
    default:
      return null;
  }
}
