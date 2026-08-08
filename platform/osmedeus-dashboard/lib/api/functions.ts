import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type {
  UtilityFunction,
  UtilityFunctionsResponse,
  EvalFunctionRequest,
  EvalFunctionResponse,
} from "@/lib/types/functions";

// Mock data for demo mode
const mockFunctions: UtilityFunction[] = [
  { name: "trim", description: "Remove leading and trailing whitespace", return_type: "string", parameters: "(str)", example: 'trim("  hello  ")', tags: ["string"] },
  { name: "lower", description: "Convert string to lowercase", return_type: "string", parameters: "(str)", example: 'lower("HELLO")', tags: ["string"] },
  { name: "upper", description: "Convert string to uppercase", return_type: "string", parameters: "(str)", example: 'upper("hello")', tags: ["string"] },
  { name: "replace", description: "Replace substring in string", return_type: "string", parameters: "(str, old, new)", example: 'replace("hello", "l", "x")', tags: ["string"] },
  { name: "split", description: "Split string by delimiter", return_type: "[]string", parameters: "(str, delimiter)", example: 'split("a,b,c", ",")', tags: ["string"] },
  { name: "join", description: "Join array elements with delimiter", return_type: "string", parameters: "(arr, delimiter)", example: 'join(["a","b"], ",")', tags: ["string"] },
  { name: "contains", description: "Check if string contains substring", return_type: "bool", parameters: "(str, substr)", example: 'contains("hello", "ll")', tags: ["string"] },
  { name: "startsWith", description: "Check if string starts with prefix", return_type: "bool", parameters: "(str, prefix)", example: 'startsWith("hello", "he")', tags: ["string"] },
  { name: "endsWith", description: "Check if string ends with suffix", return_type: "bool", parameters: "(str, suffix)", example: 'endsWith("hello", "lo")', tags: ["string"] },
  { name: "readFile", description: "Read file contents", return_type: "string", parameters: "(path)", example: 'readFile("/tmp/test.txt")', tags: ["file"] },
  { name: "writeFile", description: "Write content to file", return_type: "bool", parameters: "(path, content)", example: 'writeFile("/tmp/out.txt", "data")', tags: ["file"] },
  { name: "appendFile", description: "Append content to file", return_type: "bool", parameters: "(path, content)", example: 'appendFile("/tmp/log.txt", "line")', tags: ["file"] },
  { name: "fileExists", description: "Check if file exists", return_type: "bool", parameters: "(path)", example: 'fileExists("/tmp/test.txt")', tags: ["file"] },
  { name: "deleteFile", description: "Delete a file", return_type: "bool", parameters: "(path)", example: 'deleteFile("/tmp/test.txt")', tags: ["file"] },
  { name: "listDir", description: "List directory contents", return_type: "[]string", parameters: "(path)", example: 'listDir("/tmp")', tags: ["file"] },
  { name: "httpGet", description: "Make HTTP GET request", return_type: "string", parameters: "(url)", example: 'httpGet("https://example.com")', tags: ["http"] },
  { name: "httpPost", description: "Make HTTP POST request", return_type: "string", parameters: "(url, body)", example: 'httpPost("https://api.example.com", "{}")', tags: ["http"] },
  { name: "resolveIP", description: "Resolve hostname to IP", return_type: "string", parameters: "(hostname)", example: 'resolveIP("example.com")', tags: ["network"] },
  { name: "checkPort", description: "Check if port is open", return_type: "bool", parameters: "(host, port)", example: 'checkPort("localhost", 80)', tags: ["network"] },
  { name: "base64Encode", description: "Encode string to base64", return_type: "string", parameters: "(str)", example: 'base64Encode("hello")', tags: ["encoding"] },
  { name: "base64Decode", description: "Decode base64 string", return_type: "string", parameters: "(str)", example: 'base64Decode("aGVsbG8=")', tags: ["encoding"] },
  { name: "urlEncode", description: "URL encode string", return_type: "string", parameters: "(str)", example: 'urlEncode("hello world")', tags: ["encoding"] },
  { name: "urlDecode", description: "URL decode string", return_type: "string", parameters: "(str)", example: 'urlDecode("hello%20world")', tags: ["encoding"] },
  { name: "md5", description: "Calculate MD5 hash", return_type: "string", parameters: "(str)", example: 'md5("hello")', tags: ["encoding"] },
  { name: "sha256", description: "Calculate SHA256 hash", return_type: "string", parameters: "(str)", example: 'sha256("hello")', tags: ["encoding"] },
  { name: "jsonParse", description: "Parse JSON string to object", return_type: "object", parameters: "(str)", example: 'jsonParse(\'{"key":"value"}\')', tags: ["data_query"] },
  { name: "jsonStringify", description: "Convert object to JSON string", return_type: "string", parameters: "(obj)", example: 'jsonStringify({"key":"value"})', tags: ["data_query"] },
  { name: "jsonGet", description: "Get value from JSON by path", return_type: "any", parameters: "(obj, path)", example: 'jsonGet(obj, "data.items[0]")', tags: ["data_query"] },
];

function normalizeTag(input: string): string {
  return input.trim().toLowerCase().replace(/\s+/g, "_");
}

export async function listUtilityFunctions(): Promise<UtilityFunctionsResponse> {
  if (isDemoMode()) {
    return { functions: mockFunctions, total: mockFunctions.length };
  }
  const res = await http.get(`${API_PREFIX}/functions/list`);
  const data = res.data as unknown;
  const functionsNode = (data as any)?.functions as unknown;
  if (Array.isArray(functionsNode)) {
    const funcs = functionsNode as UtilityFunction[];
    const total = Number.isFinite((data as any)?.total) ? Number((data as any)?.total) : funcs.length;
    return { functions: funcs, total };
  }
  if (functionsNode && typeof functionsNode === "object") {
    const byCategory = functionsNode as Record<string, UtilityFunction[]>;
    const funcs = Object.entries(byCategory).flatMap(([category, fns]) =>
      (fns ?? []).map((fn) => {
        const existingTags = Array.isArray((fn as any)?.tags) ? ((fn as any).tags as string[]) : undefined;
        const tags = existingTags && existingTags.length > 0 ? existingTags : [normalizeTag(category)];
        return { ...fn, tags };
      })
    );
    return { functions: funcs, total: funcs.length };
  }
  return { functions: [], total: 0 };
}

export async function evalUtilityFunction(
  input: EvalFunctionRequest
): Promise<EvalFunctionResponse> {
  const res = await http.post(`${API_PREFIX}/functions/eval`, input);
  const data = res.data as EvalFunctionResponse;
  return {
    result: (data as any)?.result,
    rendered_script: (data as any)?.rendered_script || input.script,
  };
}
