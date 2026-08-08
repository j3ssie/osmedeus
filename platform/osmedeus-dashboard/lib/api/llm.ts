import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  EmbeddingsRequest,
  EmbeddingsResponse,
} from "@/lib/types/llm";

export async function chatCompletion(
  request: ChatCompletionRequest
): Promise<ChatCompletionResponse> {
  if (isDemoMode()) {
    const lastUserMessage = request.messages
      ?.slice()
      .reverse()
      .find((m) => m.role === "user")?.content;
    const echo = typeof lastUserMessage === "string" ? lastUserMessage.slice(0, 200) : "";
    return {
      id: `demo-chat-${Date.now()}`,
      model: "demo",
      content: `Demo mode is enabled. LLM requests are not sent to the backend.\n\nInput: ${echo}`,
      finish_reason: "stop",
      tool_calls: undefined,
      usage: {
        prompt_tokens: 0,
        completion_tokens: 0,
        total_tokens: 0,
      },
    };
  }
  const res = await http.post(`${API_PREFIX}/llm/v1/chat/completions`, request);
  const data = res.data;
  return {
    id: data?.id || "",
    model: data?.model || "",
    content: data?.content ?? null,
    finish_reason: data?.finish_reason || "stop",
    tool_calls: data?.tool_calls,
    usage: {
      prompt_tokens: data?.usage?.prompt_tokens || 0,
      completion_tokens: data?.usage?.completion_tokens || 0,
      total_tokens: data?.usage?.total_tokens || 0,
    },
  };
}

export async function generateEmbeddings(
  request: EmbeddingsRequest
): Promise<EmbeddingsResponse> {
  if (isDemoMode()) {
    return {
      model: "demo",
      embeddings: (request.input ?? []).map(() => Array.from({ length: 16 }, () => 0)),
      usage: {
        prompt_tokens: 0,
        total_tokens: 0,
      },
    };
  }
  const res = await http.post(`${API_PREFIX}/llm/v1/embeddings`, request);
  const data = res.data;
  return {
    model: data?.model || "",
    embeddings: data?.embeddings || [],
    usage: {
      prompt_tokens: data?.usage?.prompt_tokens || 0,
      total_tokens: data?.usage?.total_tokens || 0,
    },
  };
}
