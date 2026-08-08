export interface LLMMessage {
  role: "system" | "user" | "assistant" | "tool";
  content: string;
}

export interface LLMTool {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: Record<string, unknown>;
  };
}

export interface ChatCompletionRequest {
  messages: LLMMessage[];
  model?: string;
  max_tokens?: number;
  temperature?: number;
  top_p?: number;
  tools?: LLMTool[];
  tool_choice?: string;
  response_format?: { type: "text" | "json_object" };
}

export interface ToolCall {
  id: string;
  type: "function";
  function: {
    name: string;
    arguments: string;
  };
}

export interface ChatCompletionResponse {
  id: string;
  model: string;
  content: string | null;
  finish_reason: string;
  tool_calls?: ToolCall[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

export interface EmbeddingsRequest {
  input: string[];
  model?: string;
}

export interface EmbeddingsResponse {
  model: string;
  embeddings: number[][];
  usage: {
    prompt_tokens: number;
    total_tokens: number;
  };
}
