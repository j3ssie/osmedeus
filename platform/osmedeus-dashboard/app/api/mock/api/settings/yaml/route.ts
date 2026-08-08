import { NextResponse } from "next/server";
import { mockSettingsYaml } from "@/lib/mock/data/settings-yaml";

export const dynamic = "force-static";

export async function GET() {
  return new NextResponse(mockSettingsYaml, {
    status: 200,
    headers: {
      "Content-Type": "text/yaml",
    },
  });
}

export async function PUT(request: Request) {
  try {
    const yaml = await request.text();

    if (!yaml || !yaml.trim()) {
      return NextResponse.json(
        { error: true, message: "YAML content cannot be empty" },
        { status: 400 }
      );
    }

    // In a real implementation, this would validate and save the YAML
    // For mock purposes, we just return a success response
    return NextResponse.json({
      message: "Configuration updated successfully",
      path: "/home/user/osmedeus-base/osm-settings.yaml",
      backup: "/home/user/osmedeus-base/osm-settings.yaml.backup",
    });
  } catch {
    return NextResponse.json(
      { error: true, message: "Invalid YAML configuration" },
      { status: 400 }
    );
  }
}
