import AuthenticationPageClient, {
  type AuthorizationFormState,
} from "./Client";

export enum AuthorizationPageMode {
  SignIn = "signin",
  SignUp = "signup",
  Migration = "migration",
  MFA = "mfa",
}

export interface SignInSideProps {
  initFormState?: AuthorizationFormState;
  mode?: AuthorizationPageMode;
}

export default function AuthenticationPage(props: SignInSideProps) {
  return (
    <AuthenticationPageClient
      mode={props.mode}
      initFormState={props.initFormState}
    />
  );
}
