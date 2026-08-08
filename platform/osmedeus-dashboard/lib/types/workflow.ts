// YAML Workflow Types

export type WorkflowStepType =
  | "bash"
  | "parallel"
  | "parallel-steps"
  | "function"
  | "foreach"
  | "http"
  | "llm"
  | "remote-bash"
  | "container";

export interface WorkflowParam {
  name: string;
  required?: boolean;
  default?: string;
  generator?: string;
}

export interface WorkflowVariable {
  name: string;
  type: "domain" | "path" | "number" | "string";
  required: boolean;
}

export interface WorkflowDependencies {
  commands: string[];
  files: string[];
  variables: WorkflowVariable[];
}

export interface WorkflowReport {
  name: string;
  path: string;
  type: "text" | "csv" | "json";
  description: string;
}

export interface OnAction {
  action: "abort" | "continue" | "export" | "log" | "run";
  message?: string;
  name?: string;
  value?: unknown;
  condition?: string;
  type?: string;
  command?: string;
  functions?: string[];
  export?: Record<string, unknown>;
}

export interface DecisionRule {
  condition: string;
  next: string;
}

export interface SwitchDecisionTarget {
  goto: string;
}

export interface SwitchDecision {
  switch: string;
  cases: Record<string, SwitchDecisionTarget>;
  default?: SwitchDecisionTarget;
}

export type WorkflowDecision = DecisionRule[] | SwitchDecision;

export interface WorkflowFlowModule {
  name: string;
  path?: string;
  extends?: string;
  depends_on?: string[];
  condition?: string;
  params?: Record<string, unknown>;
  decision?: WorkflowDecision;
  on_success?: OnAction[];
  on_error?: OnAction[];
}

export interface WorkflowStep {
  name: string;
  type: WorkflowStepType;
  command?: string;
  depends_on?: string[];
  speed_args?: string;
  config_args?: string;
  input_args?: string;
  output_args?: string;
  commands?: string[];
  parallel_commands?: string[];
  function?: string;
  functions?: string[];
  parallel_functions?: string[];
  parallel_steps?: WorkflowStep[];
  step?: WorkflowStep; // For foreach inner step
  input?: string;
  variable?: string;
  threads?: number;
  timeout?: number;
  log?: string;
  pre_condition?: string;
  exports?: Record<string, string>;
  step_runner?: string;
  step_runner_config?: Record<string, unknown>;
  std_file?: string;
  step_remote_file?: string;
  host_output_file?: string;
  url?: string;
  method?: string;
  headers?: Record<string, string>;
  request_body?: string;
  messages?: Array<Record<string, unknown>>;
  tools?: Array<Record<string, unknown>>;
  tool_choice?: unknown;
  llm_config?: Record<string, unknown>;
  extra_llm_parameters?: Record<string, unknown>;
  is_embedding?: boolean;
  embedding_input?: string[];
  on_success?: OnAction[];
  on_error?: OnAction[];
  decision?: WorkflowDecision;
}

export interface ModuleWorkflowYaml {
  name: string;
  kind: "module";
  description?: string;
  params?: WorkflowParam[];
  dependencies?: WorkflowDependencies;
  reports?: WorkflowReport[];
  steps: WorkflowStep[];
}

export interface FlowWorkflowYaml {
  name: string;
  kind: "flow";
  description?: string;
  extends?: string;
  override?: Record<string, unknown>;
  params?: WorkflowParam[];
  dependencies?: WorkflowDependencies;
  reports?: WorkflowReport[];
  modules?: WorkflowFlowModule[];
  trigger?: unknown;
}

export type WorkflowYaml = ModuleWorkflowYaml | FlowWorkflowYaml;

// Workflow metadata for listing
export interface Workflow {
  name: string;
  kind: "module" | "flow";
  description: string;
  tags: string[];
  file_path: string;
  params: WorkflowParam[];
  required_params: string[];
  step_count: number;
  module_count: number;
  checksum: string;
  indexed_at: string;
}

// React Flow Node/Edge types
export type WorkflowNodeType =
  | "bash"
  | "parallel"
  | "parallel-steps"
  | "function"
  | "foreach"
  | "http"
  | "llm"
  | "remote-bash"
  | "container"
  | "module"
  | "trigger"
  | "start"
  | "end"
  | "override";

export interface WorkflowNodeData extends Record<string, unknown> {
  label: string;
  step: WorkflowStep | null;
  module: WorkflowFlowModule | null;
  triggers?: Record<string, unknown>[];
  status?: "idle" | "running" | "completed" | "failed";
}

export interface WorkflowEdgeData {
  condition?: string;
  isDecision?: boolean;
  isParallel?: boolean;
}
