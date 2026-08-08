export interface UtilityFunction {
  name: string;
  description: string;
  return_type: string;
  parameters?: string;
  example?: string;
  tags?: string[];
}

export interface UtilityFunctionsResponse {
  functions: UtilityFunction[];
  total: number;
}

export interface EvalFunctionRequest {
  script: string;
  target?: string;
  params?: Record<string, string>;
}

export interface EvalFunctionResponse {
  result: unknown;
  rendered_script: string;
}
