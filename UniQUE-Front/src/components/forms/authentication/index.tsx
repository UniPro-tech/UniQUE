"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ChevronDownIcon, InfoIcon } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, type SubmitHandler, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { type AuthenticationResponse, User } from "@/classes/User";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  // FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { GetErrorMessageClient, UniQUE_Error } from "@/errors/base";
import { cn } from "@/lib/utils";
import { Checkbox } from "../../ui/checkbox";
import { toastOption } from "../../ui/sonner";
import { migrateAction, type SignupFormState, signupAction } from "./action";

export function SigninForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(
    undefined,
  );

  const router = useRouter();

  const contactFormSchema = z.object({
    username: z.string().min(1, "ユーザーIDを入力してください。"),
    password: z.string().min(1, "パスワードを入力してください。"),
    isRemember: z.boolean(),
  });
  interface FormState {
    username: string;
    password: string;
    isRemember: boolean;
  }
  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<z.infer<typeof contactFormSchema>>({
    resolver: zodResolver(contactFormSchema),
    defaultValues: {
      isRemember: false,
    },
  });

  // onSubmits
  const onSubmit: SubmitHandler<FormState> = async (data: FormState) => {
    setIsSubmitting(true);
    let nexthop: string | undefined;
    toast.promise<AuthenticationResponse>(
      async () => {
        const res = await User.signin(data);
        if (res instanceof Error || res instanceof UniQUE_Error) {
          throw res;
        }
        return res;
      },
      {
        ...toastOption,
        loading: `送信中...`,
        success: (data: AuthenticationResponse) => {
          setErrorMessage(undefined);
          if (data.requireMfa) {
            nexthop = `/signin?mfa=1`;
            return `他段階認証が必要です。`;
          }
          nexthop = `/dashboard`;
          return `ログインに成功しました！`;
        },
        error: (error: Error) => {
          const errorMessage = GetErrorMessageClient(error);
          setErrorMessage(errorMessage);
          return `${errorMessage}`;
        },
        finally: () => {
          if (nexthop) router.push(nexthop);
          setIsSubmitting(false);
        },
      },
    );
  };

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">おかえりなさい！</CardTitle>
          <CardDescription>
            {/* Login with your Apple or Google account */}
            ユーザーIDとパスワードを入力してサインイン
          </CardDescription>
          {(errorMessage || errors.form || errors.root) && (
            <Alert variant={"destructive"} className="mt-5">
              <InfoIcon className="font-medium" />
              <AlertTitle className="font-medium">
                エラーが発生しました
              </AlertTitle>
              <AlertDescription>
                {errorMessage || errors.form?.message || errors.root?.message}
              </AlertDescription>
            </Alert>
          )}
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handleSubmit(onSubmit, (e) => {
              console.log("Invalid", e);
            })}
          >
            <FieldGroup>
              {/* 今は使わないのでコメントアウト
              <Field>
                <Button variant="outline" type="button">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                    <title>Apple Icon</title>
                    <path
                      d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701"
                      fill="currentColor"
                    />
                  </svg>
                  Login with Apple
                </Button>
                <Button variant="outline" type="button">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                    <title>Google Icon</title>
                    <path
                      d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
                      fill="currentColor"
                    />
                  </svg>
                  Login with Google
                </Button>
              </Field>
              <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
                Or continue with
              </FieldSeparator>
              */}
              <Field>
                <FieldLabel htmlFor="customid">ユーザーID</FieldLabel>
                <Input
                  autoComplete="username"
                  id="customid"
                  type="username"
                  placeholder="unipro-tarou"
                  required
                  {...register("username")}
                />
                {errors.username && (
                  <FieldDescription className="text-red-500">
                    {errors.username.message}
                  </FieldDescription>
                )}
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">パスワード</FieldLabel>
                  <Link
                    href="#"
                    className="ml-auto text-xs underline-offset-4 hover:underline"
                  >
                    パスワードをお忘れですか？
                  </Link>
                </div>
                <Input
                  autoComplete="current-password"
                  id="password"
                  placeholder="・・・・・・・・"
                  type="password"
                  required
                  {...register("password")}
                />
                {errors.password && (
                  <FieldDescription className="text-red-500">
                    {errors.password.message}
                  </FieldDescription>
                )}
              </Field>
              <Field orientation="horizontal">
                <Controller
                  name="isRemember"
                  control={control}
                  render={({ field }) => (
                    <Checkbox
                      id="remember"
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  )}
                />
                <FieldLabel htmlFor="password">
                  ログイン状態を保持する
                </FieldLabel>
              </Field>
              {errors.password && (
                <FieldDescription className="text-red-500">
                  {errors.password.message}
                </FieldDescription>
              )}
              <Field>
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "送信中..." : "サインイン"}
                </Button>
                <FieldDescription className="text-center">
                  まだ登録してない方は <Link href={"/signup"}>こちら</Link>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
      <AgreeTOS />
    </div>
  );
}

const MIN_AGE = new Date();
MIN_AGE.setFullYear(MIN_AGE.getFullYear() - 13); // 13年前の日付

export function SignupForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(
    undefined,
  );

  const router = useRouter();

  const contactFormSchema = z
    .object({
      displayName: z.string().min(1, "表示名は必須です"),
      customId: z
        .string()
        .min(1, "ユーザーIDは必須です")
        .min(3, "ユーザーIDは3文字以上で入力してください")
        .max(30, "ユーザーIDは30文字以下で入力してください")
        .regex(
          /^[a-zA-Z0-9_-]+$/,
          "ユーザーIDは英数字、_-のみで入力してください",
        ),
      externalEmail: z
        .email("メールアドレスの形式が正しくありません")
        .min(1, "メールアドレスは必須です"),
      password: z
        .string()
        .min(1, "パスワードは必須です")
        .min(8, "パスワードは8文字以上で入力してください"),
      confirmPassword: z.string().min(1, "パスワード（再入力）は必須です"),
      birthdate: z.string(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: "パスワードが一致しません",
      path: ["confirmPassword"],
    })
    .refine((data) => new Date(data.birthdate) < MIN_AGE, {
      message: "13才未満の方はご登録いただけません。",
      path: ["birthdate"],
    });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<z.infer<typeof contactFormSchema>>({
    resolver: zodResolver(contactFormSchema),
  });

  // onSubmits
  const onSubmit: SubmitHandler<SignupFormState> = async (
    data: SignupFormState,
  ) => {
    setIsSubmitting(true);
    let nexthop: string | undefined;
    toast.promise(
      async () => {
        const res = await signupAction(data);
        if (res instanceof Error || res instanceof UniQUE_Error) throw res;
      },
      {
        ...toastOption,
        loading: `送信中...`,
        success: () => {
          setErrorMessage(undefined);
          setIsSubmitted(true);
          return `メンバー登録申請を送信しました！`;
        },
        error: (error: Error) => {
          const errorMessage = GetErrorMessageClient(error);
          setErrorMessage(errorMessage);
          return `${errorMessage}`;
        },
        finally: () => {
          if (nexthop) router.push(nexthop);
          setIsSubmitting(false);
        },
      },
    );
  };

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        {!isSubmitted ? (
          <>
            <CardHeader className="text-center">
              <CardTitle className="text-xl">メンバー登録申請</CardTitle>
              <CardDescription>
                メンバー登録申請を行うには必要事項を入力してください。
              </CardDescription>
              <Alert variant={"default"} className="mt-5">
                <InfoIcon className="font-medium" />
                <AlertTitle className="font-medium">
                  2026年4月以前にメンバー登録した方へ
                </AlertTitle>
                <AlertDescription>
                  UniQUEには登録されておりませんので、
                  <Link href={"/migrate"}>こちら</Link>{" "}
                  から登録を移行してください。
                </AlertDescription>
              </Alert>
              {(errorMessage || errors.form || errors.root) && (
                <Alert variant={"destructive"} className="mt-5">
                  <InfoIcon className="font-medium" />
                  <AlertTitle className="font-medium">
                    エラーが発生しました
                  </AlertTitle>
                  <AlertDescription>
                    {errorMessage ||
                      errors.form?.message ||
                      errors.root?.message}
                  </AlertDescription>
                </Alert>
              )}
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)}>
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="displayName">表示名</FieldLabel>
                    <Input
                      autoComplete="name"
                      id="displayName"
                      type="text"
                      placeholder="ゆにぷろ太郎"
                      required
                      {...register("displayName")}
                    />
                    <FieldDescription
                      className={
                        errors.displayName ? "text-red-500" : undefined
                      }
                    >
                      {errors.displayName
                        ? errors.displayName.message
                        : "ニックネーム可・他のメンバーに公開されます。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="customId">ユーザーID</FieldLabel>
                    <Input
                      autoComplete="username"
                      id="customId"
                      type="text"
                      placeholder="unipro_tarou"
                      required
                      {...register("customId")}
                    />
                    <FieldDescription
                      className={errors.customId ? "text-red-500" : undefined}
                    >
                      {errors.customId
                        ? errors.customId.message
                        : "他人と被らず英数字と_-で3文字以上30文字以下で入力してください。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="externalEmail">
                      メールアドレス
                    </FieldLabel>
                    <Input
                      autoComplete="email"
                      id="externalEmail"
                      type="email"
                      placeholder="hogehoge@example.com"
                      required
                      {...register("externalEmail")}
                    />
                    <FieldDescription
                      className={
                        errors.externalEmail ? "text-red-500" : undefined
                      }
                    >
                      {errors.externalEmail
                        ? errors.externalEmail.message
                        : "他のメンバーには非公開です。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <Field className="grid grid-cols-2 gap-4">
                      <Field>
                        <FieldLabel htmlFor="password">パスワード</FieldLabel>
                        <Input
                          autoComplete="new-password"
                          id="password"
                          placeholder="・・・・・・・・"
                          type="password"
                          required
                          {...register("password")}
                        />
                      </Field>
                      <Field>
                        <FieldLabel htmlFor="confirm-password">
                          パスワードの確認
                        </FieldLabel>
                        <Input
                          autoComplete="new-password"
                          id="confirm-password"
                          placeholder="・・・・・・・・"
                          type="password"
                          required
                          {...register("confirmPassword")}
                        />
                      </Field>
                    </Field>
                    <FieldDescription
                      className={
                        errors.password || errors.confirmPassword
                          ? "text-red-500"
                          : undefined
                      }
                    >
                      {errors.password
                        ? errors.password.message
                        : errors.confirmPassword
                          ? errors.confirmPassword.message
                          : "8文字以上の半角英数"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="birthdate">生年月日</FieldLabel>
                    <BirthdateRequireReason />
                    <Input
                      autoComplete="bday"
                      id="birthdate"
                      type="date"
                      placeholder="2000/03/04"
                      required
                      {...register("birthdate")}
                    />
                    <FieldDescription
                      className={errors.birthdate ? "text-red-500" : undefined}
                    >
                      {errors.birthdate
                        ? errors.birthdate.message
                        : "デフォルトで他のメンバーには非公開です。未成年かどうかのみ公開されます。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <Button type="submit" disabled={isSubmitting}>
                      {isSubmitting ? "送信中..." : "メンバー登録を申請する"}
                    </Button>
                    <FieldDescription className="text-center">
                      すでに登録済みの方は{" "}
                      <Link href="/signin">サインイン</Link>
                    </FieldDescription>
                  </Field>
                </FieldGroup>
              </form>
            </CardContent>
          </>
        ) : (
          <SendedEmail />
        )}
      </Card>
      <AgreeTOS />
    </div>
  );
}

export function MigrateForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | undefined>(
    undefined,
  );

  const router = useRouter();

  const contactFormSchema = z
    .object({
      displayName: z.string().min(1, "表示名は必須です"),
      email: z
        .email("メールアドレスの形式が正しくありません")
        .regex(/^[a-zA-Z0-9._-]+@uniproject\.jp$/)
        .min(1, "メールアドレスは必須です"),
      externalEmail: z
        .email("メールアドレスの形式が正しくありません")
        .min(1, "メールアドレスは必須です"),
      password: z
        .string()
        .min(1, "パスワードは必須です")
        .min(8, "パスワードは8文字以上で入力してください"),
      confirmPassword: z.string().min(1, "パスワード（再入力）は必須です"),
      birthdate: z.string(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: "パスワードが一致しません",
      path: ["confirmPassword"],
    })
    .refine((data) => new Date(data.birthdate) < MIN_AGE, {
      message: "13才未満の方はご登録いただけません。",
      path: ["birthdate"],
    });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<z.infer<typeof contactFormSchema>>({
    resolver: zodResolver(contactFormSchema),
  });

  interface FormState {
    displayName: string;
    email: string;
    externalEmail: string;
    password: string;
    confirmPassword: string;
    birthdate: string;
  }

  const onSubmit: SubmitHandler<FormState> = (data: FormState) => {
    setIsSubmitting(true);
    let nexthop: string | undefined;
    toast.promise(
      async () => {
        const res = await migrateAction(data);
        if (res instanceof Error || res instanceof UniQUE_Error) throw res;
      },
      {
        ...toastOption,
        loading: `送信中...`,
        success: () => {
          setErrorMessage(undefined);
          setIsSubmitted(true);
          return `メンバー登録を移行しました！`;
        },
        error: (error: Error) => {
          const errorMessage = GetErrorMessageClient(error);
          setErrorMessage(errorMessage);
          return `${errorMessage}`;
        },
        finally: () => {
          if (nexthop) router.push(nexthop);
          setIsSubmitting(false);
        },
      },
    );
  };
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        {!isSubmitted ? (
          <>
            <CardHeader className="text-center">
              <CardTitle className="text-xl">アカウント移行</CardTitle>
              <CardDescription>
                2026年4月以前からのメンバー登録をUniQUEに移行します
              </CardDescription>
              <Alert variant={"default"} className="mt-5">
                <InfoIcon className="font-medium" />
                <AlertTitle className="font-medium">
                  メンバーの登録はできません
                </AlertTitle>
                <AlertDescription>
                  ここでは、<strong>2026年4月以前からのメンバーの</strong>
                  登録をUniQUEに移行するためのページです。
                  新規にメンバー登録をされる方は、
                  <Link href={"/signup"}>こちら</Link>{" "}
                  から登録申請してください。
                </AlertDescription>
              </Alert>
              {(errorMessage || errors.form || errors.root) && (
                <Alert variant={"destructive"} className="mt-5">
                  <InfoIcon className="font-medium" />
                  <AlertTitle className="font-medium">
                    エラーが発生しました
                  </AlertTitle>
                  <AlertDescription>
                    {errorMessage ||
                      errors.form?.message ||
                      errors.root?.message}
                  </AlertDescription>
                </Alert>
              )}
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)}>
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="displayName">表示名</FieldLabel>
                    <Input
                      autoComplete="name"
                      id="displayName"
                      type="text"
                      placeholder="ゆにぷろ太郎"
                      required
                      {...register("displayName")}
                    />
                    <FieldDescription
                      className={
                        errors.displayName ? "text-red-500" : undefined
                      }
                    >
                      {errors.displayName ? (
                        errors.displayName.message
                      ) : (
                        <>
                          <strong className="text-red-500">
                            新しく決めてください。
                          </strong>
                          ニックネーム可・他のメンバーに公開されます。
                        </>
                      )}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="email">内部メールアドレス</FieldLabel>
                    <Input
                      autoComplete="email"
                      id="email"
                      type="text"
                      placeholder="00a.hogehoge@uniproject.jp"
                      required
                      {...register("email")}
                    />
                    <FieldDescription
                      className={errors.email ? "text-red-500" : undefined}
                    >
                      {errors.email
                        ? errors.email.message
                        : "@uniproject.jpで終わるメールアドレスを入力してください。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="externalEmail">
                      外部メールアドレス
                    </FieldLabel>
                    <Input
                      autoComplete="email"
                      id="externalEmail"
                      type="email"
                      placeholder="hogehoge@example.com"
                      required
                      {...register("externalEmail")}
                    />
                    <FieldDescription
                      className={
                        errors.externalEmail ? "text-red-500" : undefined
                      }
                    >
                      {errors.externalEmail
                        ? errors.externalEmail.message
                        : "他のメンバーには非公開です。登録したメールアドレスを入力してください。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <Field className="grid grid-cols-2 gap-4">
                      <Field>
                        <FieldLabel htmlFor="password">パスワード</FieldLabel>
                        <Input
                          autoComplete="new-password"
                          id="password"
                          placeholder="・・・・・・・・"
                          type="password"
                          required
                          {...register("password")}
                        />
                      </Field>
                      <Field>
                        <FieldLabel htmlFor="confirm-password">
                          パスワードの確認
                        </FieldLabel>
                        <Input
                          autoComplete="new-password"
                          id="confirm-password"
                          placeholder="・・・・・・・・"
                          type="password"
                          required
                          {...register("confirmPassword")}
                        />
                      </Field>
                    </Field>
                    <FieldDescription
                      className={
                        errors.password || errors.confirmPassword
                          ? "text-red-500"
                          : undefined
                      }
                    >
                      {errors.password ? (
                        errors.password.message
                      ) : errors.confirmPassword ? (
                        errors.confirmPassword.message
                      ) : (
                        <>
                          8文字以上の半角英数で
                          <strong className="text-red-500">
                            新しく決めてください。
                          </strong>
                        </>
                      )}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="birthdate">生年月日</FieldLabel>
                    <BirthdateRequireReason />
                    <Input
                      autoComplete="bday"
                      id="birthdate"
                      type="date"
                      placeholder="2000/03/04"
                      required
                      {...register("birthdate")}
                    />
                    <FieldDescription
                      className={errors.birthdate ? "text-red-500" : undefined}
                    >
                      {errors.birthdate
                        ? errors.birthdate.message
                        : "デフォルトで他のメンバーには非公開です。未成年かどうかのみ公開されます。"}
                    </FieldDescription>
                  </Field>
                  <Field>
                    <Button type="submit" disabled={isSubmitting}>
                      {isSubmitting ? "送信中..." : "メンバー登録を移行する"}
                    </Button>
                    <FieldDescription className="text-center">
                      すでに登録済みの方は{" "}
                      <Link href="/signin">サインイン</Link>
                    </FieldDescription>
                  </Field>
                </FieldGroup>
              </form>
            </CardContent>
          </>
        ) : (
          <SendedEmail />
        )}
      </Card>
      <AgreeTOS />
    </div>
  );
}

function AgreeTOS() {
  return (
    <FieldDescription className="px-6 text-center [&_span]:inline-block">
      <span>続行することで</span>
      <span>我々の定める</span>{" "}
      <span>
        <Link href="/terms" className="underline">
          利用規約
        </Link>{" "}
        と
      </span>{" "}
      <span>
        <Link href="/club_statute" className="underline">
          サークル規約
        </Link>{" "}
        、
      </span>{" "}
      <span>
        <Link href="/privacy" className="underline">
          プライバシー・ポリシー
        </Link>{" "}
        に
      </span>
      <span>同意したと</span>
      <span>みなします。</span>
    </FieldDescription>
  );
}

function SendedEmail() {
  return (
    <CardHeader className="text-center flex flex-col justify-center items-center">
      <Image
        src="/assets/mail_open.png"
        alt="mail_open"
        width={150}
        height={0}
        className="mb-3"
      />
      <CardTitle className="text-xl">メールを送信しました</CardTitle>
      <CardDescription>
        ご入力いただいたメールアドレスに認証用のメールを送信しました。ご確認の上、登録を完了させてください。
      </CardDescription>
    </CardHeader>
  );
}

function BirthdateRequireReason() {
  return (
    <Collapsible className="rounded-md data-open:bg-muted">
      <CollapsibleTrigger
        render={
          <Button
            variant={"outline"}
            className="w-full"
            suppressHydrationWarning
          >
            生年月日が必要な理由
            <ChevronDownIcon className="ml-auto group-data-panel-open/button:rotate-180" />
          </Button>
        }
      />
      <CollapsibleContent className="flex flex-col items-start gap-2 p-2.5 text-sm">
        <div>
          当サークルではDiscordを使用しています。UniProjectサークル規約により、Discordの利用規約に記載されている年齢に達しているかどうか、また、未成年者保護の観点から確認させていただいております。
        </div>
        {/*
                    <Button size="xs">Learn More</Button>
                    */}
      </CollapsibleContent>
    </Collapsible>
  );
}
