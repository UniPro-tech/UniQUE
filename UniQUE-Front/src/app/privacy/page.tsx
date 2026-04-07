import { redirect } from "next/navigation";

export const metadata = {
  title: "プライバシーポリシー",
  description: "UniQUEのプライバシーポリシーページです。",
};

export default function Content() {
  redirect("https://wiki.uniproject.jp/share/6777b30da69c94671306a70c");
}
