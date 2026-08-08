import type { Workflow } from "@/lib/types/workflow";

export const mockWorkflows: Workflow[] = [
  {
    name: "subdomain-enum",
    kind: "module",
    description: "Enumerate subdomains using multiple tools and resolve alive hosts",
    tags: ["recon", "subdomain"],
    file_path: "/workflows/subdomain-enum.yaml",
    params: [
      { name: "target", required: true, default: "", generator: "" },
      { name: "output", required: false, default: "/tmp/subdomains", generator: "" },
      { name: "threads", required: false, default: "10", generator: "" },
    ],
    required_params: ["target"],
    step_count: 9,
    module_count: 0,
    checksum: "abc123",
    indexed_at: new Date("2024-06-15").toISOString(),
  },
  {
    name: "vulnerability-scan",
    kind: "module",
    description: "Scan targets for common vulnerabilities using nuclei templates",
    tags: ["vuln", "security"],
    file_path: "/workflows/vulnerability-scan.yaml",
    params: [
      { name: "target", required: true, default: "", generator: "" },
      { name: "severity", required: false, default: "critical,high", generator: "" },
    ],
    required_params: ["target"],
    step_count: 6,
    module_count: 0,
    checksum: "def456",
    indexed_at: new Date("2024-06-20").toISOString(),
  },
  {
    name: "http-probe",
    kind: "module",
    description: "Probe HTTP/HTTPS endpoints and collect response data",
    tags: ["http", "probe"],
    file_path: "/workflows/http-probe.yaml",
    params: [
      { name: "input", required: true, default: "", generator: "" },
    ],
    required_params: ["input"],
    step_count: 4,
    module_count: 0,
    checksum: "ghi789",
    indexed_at: new Date("2024-05-10").toISOString(),
  },
  {
    name: "full-recon",
    kind: "flow",
    description: "Complete reconnaissance workflow combining subdomain enum, HTTP probe, and vulnerability scan",
    tags: ["recon", "full", "flow"],
    file_path: "/workflows/full-recon.yaml",
    params: [
      { name: "target", required: true, default: "", generator: "" },
    ],
    required_params: ["target"],
    step_count: 15,
    module_count: 3,
    checksum: "jkl012",
    indexed_at: new Date("2024-07-01").toISOString(),
  },
  {
    name: "screenshot-capture",
    kind: "module",
    description: "Capture screenshots of live web pages",
    tags: ["screenshot", "visual"],
    file_path: "/workflows/screenshot-capture.yaml",
    params: [
      { name: "input", required: true, default: "", generator: "" },
    ],
    required_params: ["input"],
    step_count: 3,
    module_count: 0,
    checksum: "mno345",
    indexed_at: new Date("2024-04-15").toISOString(),
  },
  {
    name: "tech-detection",
    kind: "module",
    description: "Detect technologies and frameworks used by target applications",
    tags: ["tech", "fingerprint"],
    file_path: "/workflows/tech-detection.yaml",
    params: [
      { name: "input", required: true, default: "", generator: "" },
    ],
    required_params: ["input"],
    step_count: 5,
    module_count: 0,
    checksum: "pqr678",
    indexed_at: new Date("2024-06-01").toISOString(),
  },
  {
    name: "mock-workflow",
    kind: "module",
    description: "Demo mock workflow to preview React Flow editor",
    tags: ["demo", "test"],
    file_path: "/workflows/mock-workflow.yaml",
    params: [],
    required_params: [],
    step_count: 5,
    module_count: 0,
    checksum: "stu901",
    indexed_at: new Date("2024-07-01").toISOString(),
  },
];

// Sample YAML content for the workflow editor
export const sampleWorkflowYaml = `name: subdomain-enum
kind: module
description: "Enumerate subdomains using multiple tools and resolve alive hosts"

params:
  - name: target
    required: true
  - name: output
    default: "/tmp/subdomains"
  - name: threads
    default: "10"

dependencies:
  commands:
    - subfinder
    - amass
    - httpx

steps:
  - name: setup
    type: bash
    command: mkdir -p {{Output}}

  - name: enumerate
    type: parallel
    parallel_steps:
      - name: subfinder
        type: bash
        command: subfinder -d {{Target}} -o {{Output}}/subfinder.txt

      - name: amass
        type: bash
        command: amass enum -passive -d {{Target}} -o {{Output}}/amass.txt

  - name: merge-results
    type: bash
    commands:
      - cat {{Output}}/*.txt | sort -u > {{Output}}/all-subdomains.txt

  - name: http-probe
    type: bash
    command: cat {{Output}}/all-subdomains.txt | httpx -o {{Output}}/alive.txt
`;
