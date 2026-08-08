import { NextResponse } from "next/server";

export const dynamic = "force-static";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const workflow = body.flow || body.module || "subdomain-enum";
    const kind = body.flow ? "flow" : body.module ? "module" : "flow";
    const emptyTarget = Boolean(body.empty_target);
    const target = emptyTarget ? "" : body.target || "example.com";
    const priority = body.priority || "normal";
    return NextResponse.json({
      message: "Scan started",
      workflow,
      kind,
      target,
      priority,
      empty_target: emptyTarget,
      threads_hold: body.threads_hold,
      heuristics_check: body.heuristics_check,
      repeat: body.repeat,
      repeat_wait_time: body.repeat_wait_time,
    });
  } catch {
    return NextResponse.json({ error: true, message: "Invalid request" }, { status: 400 });
  }
}
