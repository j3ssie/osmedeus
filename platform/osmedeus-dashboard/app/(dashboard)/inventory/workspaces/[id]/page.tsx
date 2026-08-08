import WorkspaceDetailClient from "@/components/assets/workspace-detail-client";

export async function generateStaticParams() {
  return [{ id: "default" }];
}

export default function Page({ params }: { params: { id: string } }) {
  return <WorkspaceDetailClient workspaceId={params.id} />;
}
