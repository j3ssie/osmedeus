import { redirect } from "next/navigation";

export default function AssetsRedirectPage() {
  redirect("/inventory/workspaces");
}
