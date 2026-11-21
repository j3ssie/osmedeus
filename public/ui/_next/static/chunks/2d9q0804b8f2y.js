(globalThis.TURBOPACK||(globalThis.TURBOPACK=[])).push(["object"==typeof document?document.currentScript:void 0,51673,e=>{"use strict";var t=e.i(55161),n=e.i(62280),r=e.i(72536),o=e.i(37364),i=e.i(57763);function a(){try{let e=window.localStorage.getItem("osmedeus_custom_workflows");if(!e)return{};let t=JSON.parse(e);if(!t||"object"!=typeof t)return{};let n={};return Object.entries(t).forEach(([e,t])=>{"string"!=typeof t||t.trim()&&(n[String(e)]=t)}),n}catch{return{}}}function s(){let e=a();return{...o.f,...e}}function l(){let e=[];return Object.entries(o.f).forEach(([t,n])=>{"string"==typeof n&&n.trim()&&e.push({id:t,content:n,source:"builtin"})}),Object.entries(a()).forEach(([t,n])=>{"string"==typeof n&&n.trim()&&e.push({id:t,content:n,source:"custom"})}),e}function c(e){let t=s()[e];if("string"==typeof t&&t.trim())return t;for(let{id:t,content:n}of l().slice().reverse()){let r={};try{r=i.z(n)||{}}catch{r={}}let o="string"==typeof r?.name?r.name.trim():"";if(o&&o===e||t===e)return n}return null}function u(){let e=l(),t=new Map,n=[];return e.forEach(({id:e,content:r,source:o})=>{let i=f(e,r),a=(i.name||"").trim()||e,s=t.get(a);if(!s){t.set(a,{wf:i,source:o}),n.push(a);return}"builtin"===s.source&&"custom"===o&&t.set(a,{wf:i,source:o})}),n.map(e=>t.get(e).wf)}function p(e){return Array.isArray(e)?e.filter(e=>"string"==typeof e).map(e=>e.trim()).filter(Boolean):"string"==typeof e?e.split(",").map(e=>e.trim()).filter(Boolean):[]}function d(e){let t=parseInt((e instanceof Error?e.message:"").split(":")[0]||"0",10);return Number.isFinite(t)?t:0}function m(){(0,r.d)(!0)}function f(e,t){let n,r={};try{r=i.z(t)||{}}catch{r={}}let o=Array.isArray(r?.steps)?r.steps:[],a=Array.isArray(r?.modules)?r.modules:[],s=r?.kind==="flow"?"flow":"module",l="string"==typeof r?.name?r.name:e,c="string"==typeof r?.description?r.description:"",u=((n=new Set(p(r?.tags))).add("mock-data"),Array.from(n)),d=Array.isArray(r?.params)?r.params:[];return{name:l,kind:s,description:c,tags:u,file_path:"",params:d,required_params:d.filter(e=>e?.required).map(e=>e?.name??""),step_count:o.length,module_count:a.length,checksum:"",indexed_at:new Date().toISOString()}}function g(){let e=new Set;return Object.values(s()).forEach(t=>{try{let n=i.z(t)||{};p(n?.tags).forEach(t=>e.add(t))}catch{}}),e.add("mock-data"),Array.from(e.values()).sort()}async function h(){if((0,r.f)())return u();let e=await t.Y.get(`${n.f}/workflows`);return(e.data?.data||[]).map(e=>({name:e.name??"",kind:"flow"===e.kind?"flow":"module",description:e.description??"",tags:Array.isArray(e.tags)?e.tags:[],file_path:e.file_path??"",params:Array.isArray(e.params)?e.params:[],required_params:Array.isArray(e.required_params)?e.required_params:[],step_count:e.step_count??0,module_count:e.module_count??0,checksum:e.checksum??"",indexed_at:e.indexed_at??""}))}async function y(e={}){let t=u().filter(t=>{if(e.kind&&t.kind!==e.kind)return!1;if(e.tags&&e.tags.length>0){let n=new Set((t.tags||[]).map(e=>String(e)));if(!e.tags.some(e=>n.has(e)))return!1}if(e.search&&e.search.trim()){let n=e.search.trim().toLowerCase();if(!`${t.name??""} ${t.description??""} ${(t.tags||[]).join(" ")}`.toLowerCase().includes(n))return!1}return!0}),n="number"==typeof e.offset?e.offset:0,r="number"==typeof e.limit?e.limit:t.length;return{items:t.slice(Math.max(0,n),Math.max(0,n)+Math.max(0,r)),pagination:{total:t.length,offset:n,limit:r}}}async function b(e={}){if((0,r.f)()){let t=(await h()).filter(t=>{if(e.kind&&t.kind!==e.kind)return!1;if(e.tags&&e.tags.length>0){let n=new Set((t.tags||[]).map(e=>String(e)));if(!e.tags.some(e=>n.has(e)))return!1}if(e.search&&e.search.trim()){let n=e.search.trim().toLowerCase();if(!`${t.name??""} ${t.description??""} ${(t.tags||[]).join(" ")}`.toLowerCase().includes(n))return!1}return!0}),n="number"==typeof e.offset?e.offset:0,r="number"==typeof e.limit?e.limit:t.length;return{items:t.slice(Math.max(0,n),Math.max(0,n)+Math.max(0,r)),pagination:{total:t.length,offset:n,limit:r}}}let o={};e.source&&(o.source=e.source),e.tags&&e.tags.length>0&&(o.tags=e.tags.join(",")),e.kind&&(o.kind=e.kind),e.search&&(o.search=e.search),"number"==typeof e.offset&&(o.offset=e.offset),"number"==typeof e.limit&&(o.limit=e.limit);try{let e=await t.Y.get(`${n.f}/workflows`,{params:o}),r=e.data?.data||[],i=e.data?.pagination||{total:r.length,offset:0,limit:r.length},a=r.map(e=>({name:e.name??"",kind:"flow"===e.kind?"flow":"module",description:e.description??"",tags:Array.isArray(e.tags)?e.tags.map(e=>String(e)):[],file_path:e.file_path??"",params:Array.isArray(e.params)?e.params:[],required_params:Array.isArray(e.required_params)?e.required_params:[],step_count:e.step_count??0,module_count:e.module_count??0,checksum:e.checksum??"",indexed_at:e.indexed_at??""}));return{items:a,pagination:{total:Number(i.total)||a.length,offset:Number(i.offset)||0,limit:Number(i.limit)||a.length}}}catch(t){if(0===d(t))return m(),y({kind:e.kind,tags:e.tags,search:e.search,offset:e.offset,limit:e.limit});throw t}}async function w(e){if((0,r.f)()){let t=c(e);return t?f(e,t):null}try{let r=(await t.Y.get(`${n.f}/workflows/${encodeURIComponent(e)}`,{params:{json:!0}})).data;if("string"==typeof r)return f(e,r);return{name:r.name??"",kind:"flow"===r.kind?"flow":"module",description:r.description??"",tags:Array.isArray(r.tags)?r.tags:[],file_path:r.file_path??"",params:Array.isArray(r.params)?r.params:[],required_params:Array.isArray(r.required_params)?r.required_params:[],step_count:Array.isArray(r.steps)?r.steps.length:r.step_count??0,module_count:r.module_count??0,checksum:r.checksum??"",indexed_at:r.indexed_at??""}}catch(n){let t=d(n);if(404===t)throw Error("WORKFLOW_NOT_FOUND");if(401===t)throw Error("UNAUTHORIZED");if(0===t){m();let t=c(e);return t?f(e,t):null}throw Error("REQUEST_FAILED")}}async function v(e){if((0,r.f)())return c(e);try{let r=await t.Y.get(`${n.f}/workflows/${encodeURIComponent(e)}`,{responseType:"text"});return"string"==typeof r.data?r.data:r.data?.yaml??null}catch(n){let t=d(n);if(404===t)throw Error("WORKFLOW_NOT_FOUND");if(401===t)throw Error("UNAUTHORIZED");if(0===t)return m(),c(e);throw Error("REQUEST_FAILED")}}async function k(){if((0,r.f)())return g();try{let e=await t.Y.get(`${n.f}/workflows/tags`),r=e.data?.tags||[];return Array.isArray(r)?r.map(e=>String(e)):[]}catch(e){if(0===d(e))return m(),g();throw e}}async function x(e=!1){let r=await t.Y.post(`${n.f}/workflows/refresh`,void 0,{params:e?{force:!0}:{}});return{message:r.data?.message||"",added:Number(r.data?.added||0),updated:Number(r.data?.updated||0),removed:Number(r.data?.removed||0),errors:Array.isArray(r.data?.errors)?r.data.errors:[]}}async function _(e,o){if(!e||!o.trim())return!1;if((0,r.f)())try{let t=window.localStorage.getItem("osmedeus_custom_workflows"),n=t?JSON.parse(t):{},r=n&&"object"==typeof n?n:{};return r[e]=o,window.localStorage.setItem("osmedeus_custom_workflows",JSON.stringify(r)),!0}catch{return!1}try{let r=e,a="module";try{let e=i.z(o)||{};"string"==typeof e?.name&&e.name.trim()&&(r=e.name.trim()),e?.kind==="flow"&&(a="flow")}catch{}let s=new FormData,l=`${r||e}.yaml`,c=new Blob([o],{type:"text/yaml"});return s.append("file",c,l),await t.Y.post(`${n.f}/workflow-upload`,s,{headers:{"Content-Type":"multipart/form-data"},params:{kind:a}}),!0}catch(t){if(0===d(t))return(0,r.d)(!0),_(e,o);return!1}}e.s(["S",0,y,"e",0,w,"X",0,k,"n",0,v,"a",0,h,"P",0,b,"M",0,x,"v",0,_])},37364,e=>{"use strict";let t={"test-complex-docker-workflow":`name: test-complex-docker-workflow
kind: module
description: Complex workflow demonstrating bash, function steps with docker step_runner

params:
  - name: target
    required: true
  - name: output_dir
    default: /tmp/osm-complex-test
  - name: threads
    default: "5"

steps:
  # Step 1: Setup - Create directories using function
  - name: setup-workspace
    type: function
    log: "Setting up workspace for {{target}}"
    function: createDir("{{output_dir}}")
    exports:
      workspace_created: "output"

  # Step 2: Create input file with bash
  - name: create-target-list
    type: bash
    log: "Creating target list for {{target}}"
    commands:
      - mkdir -p {{output_dir}}/targets
      - |
        cat > {{output_dir}}/targets/hosts.txt << 'EOF'
        sub1.{{target}}
        sub2.{{target}}
        api.{{target}}
        www.{{target}}
        admin.{{target}}
        EOF
    exports:
      target_file: "{{output_dir}}/targets/hosts.txt"

  # Step 3: Docker-based DNS resolution simulation
  - name: dns-resolve
    type: remote-bash
    log: "Resolving DNS for targets in Docker"
    timeout: 60
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      env:
        TARGET_DOMAIN: "{{target}}"
      volumes:
        - "{{output_dir}}:/workspace"
      workdir: /workspace
    command: |
      echo "Resolving DNS for $TARGET_DOMAIN"
      cat /workspace/targets/hosts.txt | while read host; do
        echo "$host -> 127.0.0.1" >> /workspace/dns-resolved.txt
      done
      echo "DNS resolution complete"
    exports:
      dns_output: "{{output_dir}}/dns-resolved.txt"

  # Step 4: Parallel docker commands - simulating port scanning
  - name: parallel-port-scan
    type: remote-bash
    log: "Running parallel port scans in Docker"
    timeout: 120
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    parallel_commands:
      - 'echo "Scanning ports 1-1000 on {{target}}" && sleep 1 && echo "Port 80 open" > /workspace/ports-1.txt'
      - 'echo "Scanning ports 1001-2000 on {{target}}" && sleep 1 && echo "Port 443 open" > /workspace/ports-2.txt'
      - 'echo "Scanning ports 2001-3000 on {{target}}" && sleep 1 && echo "Port 8080 open" > /workspace/ports-3.txt'
      - 'echo "Scanning ports 3001-4000 on {{target}}" && sleep 1 && echo "Port 3306 open" > /workspace/ports-4.txt'

  # Step 5: Merge port scan results
  - name: merge-port-results
    type: bash
    log: "Merging port scan results"
    command: cat {{output_dir}}/ports-*.txt > {{output_dir}}/all-ports.txt
    exports:
      ports_file: "{{output_dir}}/all-ports.txt"

  # Step 6: Function to check file existence
  - name: verify-ports-file
    type: function
    log: "Verifying ports file exists"
    function: fileExists("{{ports_file}}")
    exports:
      ports_verified: "output"

  # Step 7: Docker-based HTTP probing with parallel steps
  - name: http-probe-parallel
    type: parallel-steps
    log: "Running parallel HTTP probes"
    parallel_steps:
      - name: probe-http
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTP on port 80"
          echo "http://{{target}}:80 [200]" > /workspace/http-80.txt
      - name: probe-https
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTPS on port 443"
          echo "https://{{target}}:443 [200]" > /workspace/https-443.txt
      - name: probe-alt
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing alternate port 8080"
          echo "http://{{target}}:8080 [404]" > /workspace/http-8080.txt

  # Step 8: Foreach loop with docker - process each subdomain
  - name: process-subdomains
    type: foreach
    log: "Processing each subdomain"
    input: "{{output_dir}}/targets/hosts.txt"
    variable: subdomain
    threads: 3
    step:
      name: scan-subdomain
      type: remote-bash
      step_runner: docker
      step_runner_config:
        image: alpine:latest
        volumes:
          - "{{output_dir}}:/workspace"
      command: |
        echo "Scanning [[subdomain]]..."
        echo "[[subdomain]]: status=200, title=Example" >> /workspace/subdomain-results.txt

  # Step 9: Read results with function
  - name: read-subdomain-results
    type: function
    log: "Reading subdomain scan results"
    function: readFile("{{output_dir}}/subdomain-results.txt")
    exports:
      scan_results: "output"

  # Step 10: Decision based routing
  - name: check-results
    type: bash
    log: "Checking scan results"
    command: wc -l < {{output_dir}}/subdomain-results.txt
    exports:
      result_count: "output"
    decision:
      - condition: result_count == "0"
        next: "_end"
      - condition: result_count != "0"
        next: "generate-report"

  # Step 11: Generate final report in docker
  - name: generate-report
    type: remote-bash
    log: "Generating final report"
    timeout: 30
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    commands:
      - echo "=== Scan Report for {{target}} ===" > /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- DNS Results ---" >> /workspace/report.txt
      - cat /workspace/dns-resolved.txt >> /workspace/report.txt 2>/dev/null || echo "No DNS results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Open Ports ---" >> /workspace/report.txt
      - cat /workspace/all-ports.txt >> /workspace/report.txt 2>/dev/null || echo "No ports found" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Subdomain Results ---" >> /workspace/report.txt
      - cat /workspace/subdomain-results.txt >> /workspace/report.txt 2>/dev/null || echo "No subdomain results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "Report generated at $(date)" >> /workspace/report.txt
    exports:
      report_file: "{{output_dir}}/report.txt"

  # Step 12: Parallel functions to get file stats
  - name: get-file-stats
    type: function
    log: "Getting file statistics"
    parallel_functions:
      - fileLength("{{output_dir}}/report.txt")
      - fileExists("{{output_dir}}/all-ports.txt")
      - trim("  {{target}}  ")
    exports:
      file_stats: "output"

  # Step 13: Cleanup (optional - controlled by pre_condition)
  - name: cleanup-temp-files
    type: bash
    log: "Cleaning up temporary files"
    pre_condition: "false"
    command: rm -rf {{output_dir}}/ports-*.txt
    on_error:
      - action: log
        message: "Cleanup failed but continuing"
      - action: continue
`,"test-decision":`name: test-decision
kind: module
description: Test conditional step routing with decision

params:
  - name: target
    required: true

steps:
  - name: check-condition
    type: bash
    command: echo "{{target}}"
    exports:
      target_value: "output"
    decision:
      - condition: target_value == "skip"
        next: "_end"
      - condition: target_value == "jump"
        next: "final-step"

  - name: middle-step
    type: bash
    command: echo "middle executed"
    exports:
      middle_output: "output"

  - name: final-step
    type: bash
    command: echo "final executed"
    exports:
      final_output: "output"
`,"test-docker-flow":`name: test-docker-flow
kind: flow
description: Flow orchestrating multiple Docker-based security scanning modules

params:
  - name: target
    required: true
  - name: Output
    default: /tmp/osm-docker-flow
  - name: mode
    default: "full"
  - name: threads
    default: "10"
  - name: skip_vuln_scan
    default: "false"

modules:
  # Module 1: Initial reconnaissance
  - name: recon-module
    path: modules/test-docker-recon
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/recon"
      threads: "{{threads}}"
    on_success:
      - action: log
        message: "Reconnaissance completed for {{target}}"
      - action: export
        key: recon_complete
        value: "true"
    on_error:
      - action: log
        message: "Reconnaissance failed for {{target}}"
      - action: abort

  # Module 2: Subdomain enumeration (depends on recon)
  - name: subdomain-module
    path: modules/test-docker-subdomain
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/subdomains"
      wordlist: "/usr/share/wordlists/subdomains.txt"
    condition: "mode == 'full' || mode == 'subdomain'"
    on_success:
      - action: export
        key: subdomains_file
        value: "{{Output}}/subdomains/all.txt"

  # Module 3: Port scanning (parallel with subdomain)
  - name: portscan-module
    path: modules/test-docker-portscan
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/ports"
      port_range: "1-10000"
      rate: "1000"
    condition: "mode == 'full' || mode == 'portscan'"

  # Module 4: HTTP probing (depends on subdomain results)
  - name: httpx-module
    path: modules/test-docker-httpx
    depends_on:
      - subdomain-module
    params:
      input: "{{subdomains_file}}"
      output_dir: "{{Output}}/http"
      threads: "{{threads}}"
    on_success:
      - action: export
        key: alive_hosts
        value: "{{Output}}/http/alive.txt"
      - action: export
        key: httpx_json
        value: "{{Output}}/http/httpx.json"
    decision:
      - condition: "fileLength('{{Output}}/http/alive.txt') == 0"
        next: "report-module"

  # Module 5: Technology detection (depends on HTTP probe)
  - name: tech-detect-module
    path: modules/test-docker-techdetect
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/tech"

  # Module 6: Screenshot capture (parallel with tech detection)
  - name: screenshot-module
    path: modules/test-docker-screenshot
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/screenshots"
      threads: "5"

  # Module 7: Vulnerability scanning (conditional)
  - name: vulnscan-module
    path: modules/test-docker-scanning
    depends_on:
      - httpx-module
      - tech-detect-module
    params:
      target: "{{target}}"
      Output: "{{Output}}/vulns"
      severity: "critical,high,medium"
      threads: "{{threads}}"
    condition: "skip_vuln_scan != 'true'"
    on_error:
      - action: log
        message: "Vulnerability scan encountered errors but continuing"
      - action: continue

  # Module 8: Directory bruteforcing (optional - depends on mode)
  - name: dirbrute-module
    path: modules/test-docker-dirbrute
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/dirs"
      wordlist: "/usr/share/wordlists/common.txt"
      threads: "20"
    condition: "mode == 'full'"

  # Module 9: JavaScript analysis (depends on dir results)
  - name: js-analysis-module
    path: modules/test-docker-jsanalysis
    depends_on:
      - dirbrute-module
    params:
      input: "{{Output}}/dirs/js-files.txt"
      output_dir: "{{Output}}/js"
    condition: "mode == 'full'"

  # Module 10: Final report generation
  - name: report-module
    path: modules/test-docker-report
    depends_on:
      - screenshot-module
      - vulnscan-module
      - tech-detect-module
    params:
      target: "{{target}}"
      input_dir: "{{Output}}"
      output_dir: "{{Output}}/reports"
      format: "html,json,markdown"
    on_success:
      - action: log
        message: "Flow completed successfully for {{target}}"
      - action: notify
        message: "Security assessment complete: {{target}}"
`,"test-loop":`name: test-loop
kind: module
description: Test foreach loop with threading

params:
  - name: target
    required: true

steps:
  - name: create-input
    type: bash
    commands:
      - mkdir -p {{Output}}
      - printf 'one\\ntwo\\nthree\\nfour\\nfive\\n' > {{Output}}/items.txt

  - name: process-items
    type: foreach
    input: "{{Output}}/items.txt"
    variable: item
    threads: 2
    step:
      name: process-item
      type: bash
      command: echo "Processing [[item]] for {{target}}"
`,"comprehensive-flow-example":`# =============================================================================
# Flow Workflow: Comprehensive Example
# =============================================================================
# This file demonstrates ALL fields available in a flow-kind workflow.
# Flows orchestrate multiple modules with dependencies, conditions, and routing.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# Same as module workflows (kind, name, description, tags, params, etc.)
# -----------------------------------------------------------------------------

# kind: Workflow type - "flow" orchestrates multiple modules
kind: flow

# name: Unique identifier for this workflow (required)
name: comprehensive-flow-example

# description: Human-readable description
description: Demonstrates all flow-specific fields including modules, dependencies, conditions, and decisions

# tags: Comma-separated tags for filtering
tags: flow, comprehensive, example

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Parameters available to all modules in this flow
# -----------------------------------------------------------------------------
params:
  - name: threads
    default: "10"

  - name: timeout
    default: "3600"

  - name: scan_depth
    default: "normal"

  - name: output_format
    default: "json"

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Flow-level dependencies checked before any module executes
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - nmap
    - nuclei
    - httpx

  files:
    - /tmp

  variables:
    - name: Target
      type: domain
      required: true

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Reports aggregated from all modules in this flow
# -----------------------------------------------------------------------------
reports:
  - name: flow-summary
    path: "{{Output}}/flow-summary.json"
    type: json
    description: Aggregated results from all modules

  - name: vulnerabilities
    path: "{{Output}}/vulnerabilities.txt"
    type: text
    description: All discovered vulnerabilities

# -----------------------------------------------------------------------------
# PREFERENCES SECTION
# Flow-level preferences apply to all module executions
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: false
  heuristics_check: 'basic'

# -----------------------------------------------------------------------------
# MODULES SECTION (Flow-specific)
# Ordered list of module references to execute
# =============================================================================
modules:
  # ===========================================================================
  # Module Reference: Basic Configuration
  # ===========================================================================
  - # name: Display name for this module execution (required)
    name: reconnaissance

    # path: Path to the module YAML file (required)
    # Can be relative to workflows directory or absolute
    path: modules/recon.yaml

    # params: Parameters to pass to this module
    # Overrides module defaults and flow-level params
    params:
      threads: "20"  # Override flow-level threads
      output_dir: "{{Output}}/recon"

  # ===========================================================================
  # Module Reference: With Dependencies (depends_on)
  # ===========================================================================
  - name: port-scanning
    path: modules/portscan.yaml

    # depends_on: List of module names that must complete before this module runs
    # Creates a DAG (Directed Acyclic Graph) for execution order
    depends_on:
      - reconnaissance

    params:
      target_list: "{{Output}}/recon/subdomains.txt"
      threads: "{{threads}}"

  # ===========================================================================
  # Module Reference: With Condition
  # ===========================================================================
  - name: web-scanning
    path: modules/webscan.yaml

    depends_on:
      - port-scanning

    # condition: JavaScript expression - module only runs if evaluates to true
    # Can reference exported variables from previous modules
    condition: 'fileLength("{{Output}}/portscan/http-services.txt") > 0'

    params:
      input: "{{Output}}/portscan/http-services.txt"

  # ===========================================================================
  # Module Reference: With on_success Handler
  # ===========================================================================
  - name: vulnerability-scanning
    path: modules/vuln-scan.yaml

    depends_on:
      - web-scanning

    condition: 'fileExists("{{Output}}/webscan/endpoints.txt")'

    params:
      endpoints: "{{Output}}/webscan/endpoints.txt"
      timeout: "{{timeout}}"

    # on_success: Actions to execute when this module completes successfully
    on_success:
      # action: log - Log a message
      - action: log
        message: "Vulnerability scanning completed for {{Target}}"

      # action: export - Export a variable for subsequent modules
      - action: export
        name: vuln_scan_complete
        value: "true"

      # action: notify - Send a notification
      - action: notify
        notify: "Vulnerability scan finished for {{Target}}"

      # action: run - Execute a follow-up step
      - action: run
        type: bash
        command: 'echo "Vuln scan done" >> {{Output}}/flow-log.txt'

      # action: run with functions
      - action: run
        type: function
        functions:
          - 'log_info("Module completed successfully")'

  # ===========================================================================
  # Module Reference: With on_error Handler
  # ===========================================================================
  - name: exploit-verification
    path: modules/exploit-verify.yaml

    depends_on:
      - vulnerability-scanning

    condition: '{{vuln_scan_complete}} == "true"'

    params:
      vulns_file: "{{Output}}/vuln-scan/vulnerabilities.json"

    # on_error: Actions to execute when this module fails
    on_error:
      # action: log - Log error message
      - action: log
        message: "Exploit verification failed for {{Target}}"
        # condition: Only execute if this condition is true
        condition: 'true'

      # action: continue - Allow flow to continue despite error
      - action: continue
        message: "Continuing flow despite exploit verification failure"

      # action: abort - Stop the entire flow
      # (Usually with a condition so it doesn't always abort)
      - action: abort
        message: "Critical failure - aborting flow"
        condition: 'false'  # Only abort under specific conditions

      # action: notify - Alert on failure
      - action: notify
        notify: "Module failed: exploit-verification for {{Target}}"

      # action: export - Export error state
      - action: export
        name: exploit_verify_failed
        value: "true"

  # ===========================================================================
  # Module Reference: With Decision Routing
  # ===========================================================================
  - name: deep-scan
    path: modules/deep-scan.yaml

    depends_on:
      - vulnerability-scanning

    # decision: Conditional routing based on results
    # Determines which module to execute next based on conditions
    decision:
      # condition: JavaScript expression to evaluate
      # next: Module name to jump to, or "_end" to finish flow
      - condition: 'fileLength("{{Output}}/vuln-scan/critical.txt") > 0'
        next: notification-critical

      - condition: 'fileLength("{{Output}}/vuln-scan/high.txt") > 0'
        next: notification-high

      # Default case - continue to next module in list
      - condition: 'true'
        next: cleanup

    params:
      scan_depth: "{{scan_depth}}"

  # ===========================================================================
  # Module Reference: Notification branches (targets of decision routing)
  # ===========================================================================
  - name: notification-critical
    path: modules/notify.yaml

    # Note: This module can be jumped to via decision routing
    # It won't run in normal sequential flow unless explicitly in depends_on

    params:
      severity: critical
      message: "Critical vulnerabilities found for {{Target}}"
      channel: security-alerts

    on_success:
      - action: export
        name: notification_sent
        value: "critical"

  - name: notification-high
    path: modules/notify.yaml

    params:
      severity: high
      message: "High severity vulnerabilities found for {{Target}}"
      channel: security-team

    on_success:
      - action: export
        name: notification_sent
        value: "high"

  # ===========================================================================
  # Module Reference: Parallel Module Execution
  # Modules with same depends_on and no inter-dependencies run in parallel
  # ===========================================================================
  - name: ssl-analysis
    path: modules/ssl-check.yaml

    depends_on:
      - port-scanning  # Same dependency as web-scanning

    params:
      input: "{{Output}}/portscan/ssl-services.txt"

  - name: dns-analysis
    path: modules/dns-check.yaml

    depends_on:
      - reconnaissance  # Can run in parallel with port-scanning

    params:
      domains: "{{Output}}/recon/subdomains.txt"

  # ===========================================================================
  # Module Reference: Cleanup/Final Module
  # ===========================================================================
  - name: cleanup
    path: modules/cleanup.yaml

    # depends_on multiple modules - waits for all to complete
    depends_on:
      - vulnerability-scanning
      - exploit-verification
      - ssl-analysis
      - dns-analysis

    # condition with multiple checks
    condition: 'true'  # Always run cleanup

    params:
      output_dir: "{{Output}}"
      format: "{{output_format}}"

    on_success:
      - action: log
        message: "Flow completed successfully for {{Target}}"

      - action: notify
        notify: "Security scan flow completed for {{Target}}"

      - action: export
        name: flow_status
        value: "completed"

    on_error:
      - action: log
        message: "Cleanup failed but flow results are preserved"

      - action: continue
        message: "Flow complete despite cleanup issues"
`,"triggers-example":`# =============================================================================
# Flow Workflow: All Trigger Types Example
# =============================================================================
# This file demonstrates ALL trigger types available in osmedeus workflows.
# Triggers define when/how a workflow should automatically execute.
# Trigger types: cron, event, watch, manual
# =============================================================================

kind: flow
name: triggers-example
description: Demonstrates all trigger types with comprehensive field documentation
tags: triggers, automation, scheduled

# -----------------------------------------------------------------------------
# TRIGGERS SECTION
# Define automatic execution triggers for this workflow
# Multiple triggers can be defined; any triggered condition will start execution
# =============================================================================
trigger:
  # ===========================================================================
  # TRIGGER TYPE: cron
  # Schedule-based execution using cron expressions
  # ===========================================================================
  - # name: Identifier for this trigger (for logging and management)
    name: daily-scan

    # on: Trigger type - cron, event, watch, or manual
    on: cron

    # schedule: Cron expression defining when to run
    # Format: minute hour day-of-month month day-of-week
    # Examples:
    #   "0 0 * * *"     - Every day at midnight
    #   "0 */6 * * *"   - Every 6 hours
    #   "0 9 * * 1-5"   - 9 AM on weekdays
    #   "0 0 1 * *"     - First day of every month at midnight
    schedule: "0 2 * * *"  # Every day at 2 AM

    # input: Defines where the target input comes from for scheduled runs
    input:
      # type: Input source type - file, event_data, function, or param
      type: file

      # path: For "file" type - path to file containing targets (one per line)
      path: "/data/targets/active-targets.txt"

    # enabled: Whether this trigger is active
    # true = trigger is active and will fire
    # false = trigger is defined but disabled
    enabled: true

  # ---------------------------------------------------------------------------
  # Cron trigger with function-based input
  # ---------------------------------------------------------------------------
  - name: weekly-full-scan
    on: cron
    schedule: "0 0 * * 0"  # Every Sunday at midnight

    input:
      # type: function - Generate input dynamically using a function
      type: function

      # function: JavaScript function to generate/retrieve targets
      # Can use built-in functions like db queries, API calls, etc.
      function: 'get_targets_from_db("scope:production")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: event
  # Event-driven execution based on system events
  # Events follow topic format: <component>.<event_type>
  # ===========================================================================
  - name: webhook-trigger
    on: event

    # event: Event configuration for event triggers
    event:
      # topic: Event topic to subscribe to
      # Common topics:
      #   webhook.received    - External webhook received
      #   assets.new          - New asset discovered
      #   assets.changed      - Asset data changed
      #   db.change           - Database record changed
      #   watch.files         - File system change detected
      topic: webhook.received

      # filters: JavaScript expressions to filter events
      # Event data available as 'event' object with fields:
      #   event.name      - Event name
      #   event.source    - Event source
      #   event.data      - JSON payload (string)
      #   event.data_type - Type of data
      # All filters must evaluate to true for trigger to fire
      filters:
        - 'event.source == "github"'
        - 'event.name == "push"'

    # input: How to extract target from event data
    input:
      # type: event_data - Extract from event payload
      type: event_data

      # field: JSON path to extract from event.data
      # Uses dot notation for nested fields
      field: "repository.html_url"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger for new asset discovery
  # ---------------------------------------------------------------------------
  - name: new-asset-scan
    on: event

    event:
      topic: assets.new

      filters:
        # Filter for specific asset types
        - 'event.data_type == "subdomain"'
        # Filter by source tool
        - 'event.source == "subfinder" || event.source == "amass"'

    input:
      type: event_data
      field: "hostname"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger with function-based input extraction
  # ---------------------------------------------------------------------------
  - name: vuln-alert-trigger
    on: event

    event:
      topic: webhook.received

      filters:
        - 'event.name == "vulnerability_alert"'
        - 'JSON.parse(event.data).severity == "critical"'

    input:
      # type: function - Use function to parse/transform event data
      type: function

      # function: Transform event data to target format
      function: 'jq("{{event.data}}", ".affected_host")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: watch
  # File system watch - triggers when files change
  # ===========================================================================
  - name: targets-file-watch
    on: watch

    # path: File or directory path to watch for changes
    # Supports glob patterns in some implementations
    path: "/data/targets/new-targets.txt"

    # input: How to get targets when file changes
    input:
      type: file
      path: "/data/targets/new-targets.txt"

    enabled: true

  # ---------------------------------------------------------------------------
  # Watch trigger on directory
  # ---------------------------------------------------------------------------
  - name: input-directory-watch
    on: watch

    path: "/data/incoming/"

    input:
      # type: function - Process newly added files
      type: function
      function: 'get_new_files("/data/incoming/", "*.txt")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: manual
  # Explicit manual trigger control
  # Used to enable/disable CLI execution for this workflow
  # ===========================================================================
  - name: manual-execution
    on: manual

    # For manual triggers, enabled controls whether CLI can run this workflow
    # enabled: true  - Allow: osmedeus run -f triggers-example -t target
    # enabled: false - Block CLI execution (only scheduled/event triggers work)
    enabled: true

    # input: Default input for manual execution
    # This is optional; CLI -t flag overrides this
    input:
      # type: param - Use a parameter as input
      type: param

      # name: Parameter name to use as target
      name: Target

  # ---------------------------------------------------------------------------
  # Disabled manual trigger example
  # This workflow can ONLY be triggered via cron/events, not CLI
  # ---------------------------------------------------------------------------
  # Uncomment to see the effect:
  # - name: block-manual
  #   on: manual
  #   enabled: false

# -----------------------------------------------------------------------------
# PARAMS SECTION
# -----------------------------------------------------------------------------
params:
  - name: scan_type
    default: "standard"

  - name: threads
    default: "10"

# -----------------------------------------------------------------------------
# MODULES SECTION
# The actual workflow steps to execute when any trigger fires
# -----------------------------------------------------------------------------
modules:
  - name: initial-recon
    path: modules/recon.yaml
    params:
      threads: "{{threads}}"

  - name: scanning
    path: modules/scan.yaml
    depends_on:
      - initial-recon
    params:
      scan_type: "{{scan_type}}"

  - name: reporting
    path: modules/report.yaml
    depends_on:
      - scanning

    on_success:
      - action: notify
        notify: "Triggered scan completed for {{Target}}"
        # condition: Only notify for certain triggers
        condition: 'true'

      - action: export
        name: completed_at
        value: "{{currentDate()}}"
`,"docker-runner-example":`# =============================================================================
# Module Workflow: Docker Runner Configuration Example
# =============================================================================
# This file demonstrates all Docker runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: docker-runner-example
description: Demonstrates Docker runner configuration with all available fields
tags: docker, runner, container

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: docker

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # DOCKER-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # image: Docker image to use (required for docker runner)
  # Format: registry/image:tag or just image:tag
  image: ubuntu:22.04

  # env: Environment variables to set inside the container
  # Map of VAR_NAME: value
  env:
    MY_VAR: my-value
    API_KEY: "{{api_key}}"  # Can use template variables
    THREADS: "{{threads}}"

  # volumes: Volume mounts in docker format
  # Format: host_path:container_path[:options]
  # Options: ro (read-only), rw (read-write)
  volumes:
    - "/tmp/osmedeus:/data"
    - "{{Output}}:/output"
    - "/etc/hosts:/etc/hosts:ro"

  # network: Docker network mode
  # Options: bridge (default), host, none, container:<name>, or network name
  network: host

  # persistent: Container lifecycle mode
  # true = reuse the same container across steps (faster, state preserved)
  # false = ephemeral, create new container per step (isolated, clean state)
  persistent: true

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory inside the container/remote
  # Commands will execute in this directory
  workdir: /app

params:
  - name: api_key
    default: "demo-key"

  - name: threads
    default: "5"

steps:
  # ===========================================================================
  # Step using workflow-level runner (docker with ubuntu:22.04)
  # ===========================================================================
  - name: use-workflow-runner
    type: bash
    log: "Running in workflow-level Docker container"
    command: 'echo "Running inside ubuntu:22.04 container"'

  # ===========================================================================
  # Step with per-step Docker runner override
  # Uses different image than workflow-level config
  # ===========================================================================
  - name: step-with-runner-override
    type: bash
    log: "Running in step-specific Docker container"

    # step_runner: Override runner type for this step only
    # Options: host, docker, ssh
    step_runner: docker

    # step_runner_config: Override runner configuration for this step
    # Same structure as runner_config but applies only to this step
    step_runner_config:
      # Use a different image for this specific step
      image: python:3.11-slim

      env:
        PYTHONPATH: /app

      volumes:
        - "{{Output}}:/output:rw"

      network: bridge

      persistent: false

      workdir: /app

    command: 'python3 -c "print(\\"Running in Python container\\")"'

  # ===========================================================================
  # Remote-bash step type with Docker (explicit remote-bash type)
  # remote-bash is specifically for executing commands in remote environments
  # ===========================================================================
  - name: remote-bash-docker
    # type: remote-bash is specifically for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution in Docker"

    # step_runner: Required for remote-bash type - specifies execution environment
    # Must be "docker" or "ssh"
    step_runner: docker

    step_runner_config:
      image: alpine:latest
      workdir: /tmp

    # command/commands/parallel_commands: Same as bash step
    command: 'echo "Hello from Alpine container" > /tmp/output.txt'

    # step_remote_file: File path on remote (inside container) to copy after execution
    # This file will be copied from the container to the host
    step_remote_file: /tmp/output.txt

    # host_output_file: Local path where the remote file will be copied
    # Template variables are supported
    host_output_file: "{{Output}}/docker-output.txt"

  # ===========================================================================
  # Parallel commands in Docker container
  # ===========================================================================
  - name: docker-parallel-commands
    type: bash
    log: "Running parallel commands in Docker"
    step_runner: docker
    step_runner_config:
      image: ubuntu:22.04
      persistent: true

    parallel_commands:
      - 'sleep 2 && echo "Parallel job A completed"'
      - 'sleep 1 && echo "Parallel job B completed"'
      - 'sleep 3 && echo "Parallel job C completed"'

  # ===========================================================================
  # Foreach loop executing in Docker
  # ===========================================================================
  - name: docker-foreach
    type: foreach
    log: "Processing items in Docker containers"
    input: "{{Output}}/targets.txt"
    variable: target
    threads: 3

    step:
      name: process-in-docker
      type: bash
      step_runner: docker
      step_runner_config:
        image: curlimages/curl:latest
        network: host
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[target]]"'
      exports:
        http_status: "{{stdout}}"

  # ===========================================================================
  # Step running on host (override workflow's docker runner)
  # ===========================================================================
  - name: run-on-host
    type: bash
    log: "Running on host machine (overriding workflow runner)"

    # Override to run locally instead of in container
    step_runner: host

    command: 'echo "This runs directly on the host machine"'

  # ===========================================================================
  # Docker step with all structured arguments
  # ===========================================================================
  - name: docker-with-args
    type: bash
    log: "Docker step with structured arguments"
    step_runner: docker
    step_runner_config:
      image: nuclei:latest
      volumes:
        - "{{Output}}:/output"
        - "/root/nuclei-templates:/templates:ro"
      workdir: /output

    command: nuclei
    speed_args: '-rate-limit 100 -c {{threads}}'
    config_args: '-t /templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /output/nuclei-results.txt'

    step_remote_file: /output/nuclei-results.txt
    host_output_file: "{{Output}}/nuclei-results.txt"

    exports:
      nuclei_output: "{{Output}}/nuclei-results.txt"
`,"ssh-runner-example":`# =============================================================================
# Module Workflow: SSH Runner Configuration Example
# =============================================================================
# This file demonstrates all SSH runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: ssh-runner-example
description: Demonstrates SSH runner configuration with all available fields
tags: ssh, runner, remote

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: ssh

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # SSH-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # host: SSH hostname or IP address (required for ssh runner)
  # Can use template variables for dynamic targeting
  host: "{{ssh_host}}"

  # port: SSH port number
  # Default: 22
  port: 22

  # user: SSH username for authentication
  user: "{{ssh_user}}"

  # key_file: Path to SSH private key file for key-based authentication
  # Preferred over password authentication for security
  key_file: "{{ssh_key_path}}"

  # password: SSH password for password-based authentication
  # WARNING: Not recommended - use key_file instead when possible
  # Can use template variables or environment references
  # password: "{{ssh_password}}"

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory on the remote machine
  # Commands will execute in this directory
  workdir: /home/scanner/workspace

params:
  - name: ssh_host
    default: "192.168.1.100"
    required: true

  - name: ssh_user
    default: "scanner"
    required: true

  - name: ssh_key_path
    default: "~/.ssh/id_rsa"

  - name: threads
    default: "10"

steps:
  # ===========================================================================
  # Step using workflow-level SSH runner
  # ===========================================================================
  - name: setup-remote-workspace
    type: bash
    log: "Setting up workspace on remote SSH server"
    command: 'mkdir -p /home/scanner/workspace/results && echo "Workspace ready"'

  # ===========================================================================
  # Remote-bash step type with SSH (explicit remote-bash type)
  # remote-bash is specifically designed for remote execution scenarios
  # ===========================================================================
  - name: remote-bash-ssh
    # type: remote-bash is explicitly for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution via SSH"

    # step_runner: Required for remote-bash type - must be "docker" or "ssh"
    step_runner: ssh

    # step_runner_config: SSH configuration (inherits from workflow if not set)
    # Omitting this uses workflow-level runner_config
    step_runner_config:
      host: "{{ssh_host}}"
      port: 22
      user: "{{ssh_user}}"
      key_file: "{{ssh_key_path}}"
      workdir: /tmp

    # command: Command to execute on remote server
    command: 'hostname && whoami && pwd > /tmp/remote-info.txt'

    # step_remote_file: File on remote server to copy back to local host
    # This is useful for retrieving results from remote execution
    step_remote_file: /tmp/remote-info.txt

    # host_output_file: Local path where remote file will be copied
    host_output_file: "{{Output}}/remote-info.txt"

    exports:
      remote_file: "{{Output}}/remote-info.txt"

  # ===========================================================================
  # Step overriding SSH connection to different server
  # ===========================================================================
  - name: connect-to-secondary-server
    type: bash
    log: "Connecting to secondary server"

    # Override workflow runner with different SSH target
    step_runner: ssh

    step_runner_config:
      host: "192.168.1.101"  # Different server
      port: 2222             # Non-standard port
      user: admin
      key_file: "~/.ssh/secondary_key"
      workdir: /opt/scanner

    command: 'echo "Connected to secondary server" && uptime'

  # ===========================================================================
  # Multiple sequential commands via SSH
  # ===========================================================================
  - name: ssh-multiple-commands
    type: bash
    log: "Running multiple commands on remote"

    # commands: List of commands executed sequentially on remote
    commands:
      - 'echo "Step 1: Checking system"'
      - 'df -h'
      - 'echo "Step 2: Checking memory"'
      - 'free -m'
      - 'echo "Step 3: Checking processes"'
      - 'ps aux | head -10'

    std_file: "{{Output}}/system-check.txt"

  # ===========================================================================
  # Parallel commands on SSH (run concurrently on remote)
  # ===========================================================================
  - name: ssh-parallel-commands
    type: bash
    log: "Running parallel commands on remote SSH server"

    parallel_commands:
      - 'nmap -sS -p 80 {{Target}} > /tmp/port80.txt'
      - 'nmap -sS -p 443 {{Target}} > /tmp/port443.txt'
      - 'nmap -sS -p 22 {{Target}} > /tmp/port22.txt'

  # ===========================================================================
  # Run tool with structured arguments via SSH
  # ===========================================================================
  - name: ssh-nuclei-scan
    type: bash
    log: "Running nuclei scan via SSH"
    timeout: 3600

    command: nuclei
    speed_args: '-rate-limit 50 -c {{threads}}'
    config_args: '-t ~/nuclei-templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /home/scanner/workspace/nuclei-results.json -json'

    step_remote_file: /home/scanner/workspace/nuclei-results.json
    host_output_file: "{{Output}}/nuclei-results.json"

    exports:
      scan_results: "{{Output}}/nuclei-results.json"

  # ===========================================================================
  # Foreach loop with SSH execution
  # Processes multiple targets on remote server
  # ===========================================================================
  - name: ssh-foreach-targets
    type: foreach
    log: "Processing targets via SSH"

    # input: File containing targets (one per line)
    input: "{{Output}}/targets.txt"

    # variable: Loop variable accessed as [[variable]] in inner step
    variable: current_target

    # threads: Number of concurrent SSH executions
    threads: 5

    step:
      name: probe-target
      type: bash
      # Inner step inherits workflow-level SSH runner
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[current_target]]" 2>/dev/null || echo "failed"'
      exports:
        probe_result: "{{stdout}}"

  # ===========================================================================
  # Step running on local host (override workflow's SSH runner)
  # Useful for local processing of results retrieved from remote
  # ===========================================================================
  - name: process-results-locally
    type: bash
    log: "Processing results on local host"

    # Override to run locally instead of via SSH
    step_runner: host

    command: 'cat "{{Output}}/nuclei-results.json" | jq -r ".info.severity" | sort | uniq -c'

    exports:
      severity_summary: "{{stdout}}"

  # ===========================================================================
  # Function step (always runs locally, regardless of workflow runner)
  # Note: Function steps execute on the host running osmedeus, not remote
  # ===========================================================================
  - name: log-completion
    type: function
    log: "Logging scan completion"
    function: 'log_info("SSH scan completed for {{Target}}")'

  # ===========================================================================
  # Cleanup step on remote server
  # ===========================================================================
  - name: cleanup-remote
    type: bash
    log: "Cleaning up remote workspace"
    command: 'rm -rf /home/scanner/workspace/temp/* 2>/dev/null; echo "Cleanup complete"'

    on_success:
      - action: log
        message: "Remote cleanup completed successfully"

    on_error:
      - action: continue
        message: "Cleanup failed but continuing workflow"
`,"all-step-types-example":`# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: all-step-types-example

# description: Human-readable description of what this workflow does
description: Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  # name: Parameter identifier used in templates as {{param_name}}
  # default: Default value if not provided via CLI
  # required: If true, workflow fails without this value
  # generator: Function to generate value, e.g., uuid(), currentDate(), getEnvVar("KEY")
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"  # Can reference built-in variables
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()  # Generates a unique ID automatically

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  # commands: List of binaries/commands that must exist in PATH
  commands:
    - echo
    - curl

  # files: List of files/directories that must exist
  files:
    - /tmp

  # variables: Define variable requirements with type validation
  # Types: domain, path, number, file, string
  variables:
    - name: Target
      type: string
      required: true

  # functions_conditions: JavaScript expressions that must evaluate to true
  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  # name: Display name for the report
  # path: File path (can use templates like {{Output}})
  # type: Format type - text, csv, json, markdown, etc.
  # description: Human-readable description
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  # skip_workspace: Equivalent to --disable-workspace-creation
  skip_workspace: false

  # disable_notifications: Equivalent to --disable-notification
  disable_notifications: true

  # disable_logging: Equivalent to --disable-logging
  disable_logging: false

  # heuristics_check: Equivalent to --heuristics-check (none, basic, advanced)
  heuristics_check: 'basic'

  # ci_output_format: Equivalent to --ci-output-format
  ci_output_format: false

  # silent: Equivalent to --silent
  silent: false

  # repeat: Equivalent to --repeat
  repeat: false

  # repeat_wait_time: Equivalent to --repeat-wait-time (e.g., 30s, 1h, 2h30m)
  repeat_wait_time: '60s'

  # clean_up_workspace: Equivalent to --clean-up-workspace
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  # ===========================================================================
  # STEP TYPE: bash
  # Execute shell commands on the host (or configured runner)
  # ===========================================================================
  - name: bash-single-command
    # type: Step type - bash, function, parallel-steps, foreach, remote-bash, http, llm
    type: bash

    # pre_condition: JavaScript expression - step only runs if this evaluates to true
    pre_condition: 'true'

    # log: Custom log message displayed when step starts (supports templates)
    log: "Executing single bash command for {{Target}}"

    # timeout: Maximum execution time in seconds (0 = no timeout)
    timeout: 60

    # command: Single command to execute
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'

    # std_file: File path to save stdout/stderr output
    std_file: "{{Output}}/step1-output.txt"

    # exports: Variables to export for subsequent steps
    # Key = variable name, Value = extraction pattern or literal value
    exports:
      step1_result: "completed"

  # ---------------------------------------------------------------------------
  # Bash step with multiple sequential commands
  # ---------------------------------------------------------------------------
  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"

    # commands: List of commands executed sequentially
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  # ---------------------------------------------------------------------------
  # Bash step with parallel commands
  # ---------------------------------------------------------------------------
  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"

    # parallel_commands: List of commands executed concurrently
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  # ---------------------------------------------------------------------------
  # Bash step with structured arguments
  # Arguments are joined in order: command + speed + config + input + output
  # ---------------------------------------------------------------------------
  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"

    command: 'echo'

    # speed_args: Performance-related arguments (e.g., thread count, rate limits)
    speed_args: '-n'

    # config_args: Configuration arguments (e.g., config file paths)
    config_args: ''

    # input_args: Input-related arguments (e.g., input file, target)
    input_args: '"Structured arguments test"'

    # output_args: Output-related arguments (e.g., output file, format)
    output_args: ''

  # ===========================================================================
  # STEP TYPE: function
  # Execute built-in utility functions via Otto JavaScript runtime
  # ===========================================================================
  - name: function-single
    type: function
    log: "Executing single function"

    # function: Single function call (JavaScript expression)
    function: 'log_info("Processing {{Target}} in function step")'

  # ---------------------------------------------------------------------------
  # Function step with multiple sequential functions
  # ---------------------------------------------------------------------------
  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"

    # functions: List of functions executed sequentially
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  # ---------------------------------------------------------------------------
  # Function step with parallel functions
  # ---------------------------------------------------------------------------
  - name: function-parallel
    type: function
    log: "Executing functions in parallel"

    # parallel_functions: List of functions executed concurrently
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  # ===========================================================================
  # STEP TYPE: parallel-steps
  # Execute multiple complete steps in parallel
  # ===========================================================================
  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"

    # parallel_steps: List of Step objects executed concurrently
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'

      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'

      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  # ===========================================================================
  # STEP TYPE: foreach
  # Iterate over input lines, executing inner step for each
  # ===========================================================================
  - name: foreach-example
    type: foreach
    log: "Iterating over items"

    # input: File path or direct content to iterate over (one item per line)
    input: "{{Output}}/items.txt"

    # variable: Name for the loop variable, accessed as [[variable]] in inner step
    variable: item

    # threads: Number of concurrent iterations (default: 1 = sequential)
    threads: 5

    # step: The inner step to execute for each item (single Step object)
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  # ===========================================================================
  # STEP TYPE: http
  # Make HTTP requests to external APIs
  # ===========================================================================
  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30

    # url: Target URL for the request (required for http type)
    url: "https://httpbin.org/post"

    # method: HTTP method - GET, POST, PUT, DELETE, PATCH, etc.
    method: POST

    # headers: Map of HTTP headers to send
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value

    # request_body: Request body content (typically JSON for POST/PUT)
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }

    exports:
      http_response: "{{response.body}}"

  # ===========================================================================
  # STEP TYPE: llm
  # Make LLM API calls for AI-powered processing
  # ===========================================================================
  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120

    # messages: Conversation messages for chat completion
    # role: system, user, assistant, or tool
    # content: Message text (can be string or multimodal array)
    messages:
      - role: system
        content: "You are a security analysis assistant."

      - role: user
        # content can be a simple string or complex multimodal content
        content: "Analyze this target: {{Target}}"

    # tools: Function tools available to the LLM
    tools:
      - type: function  # Currently only "function" type supported
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          # parameters: JSON Schema defining function parameters
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target

    # tool_choice: How the model should choose tools
    # Can be: "auto", "none", "required", or {"type": "function", "function": {"name": "fn_name"}}
    tool_choice: auto

    # llm_config: Step-level LLM configuration overrides
    llm_config:
      # provider: Specific provider to use (overrides rotation)
      provider: openai

      # model: Model override for this step
      model: gpt-4

      # Generation parameters
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0

      # Request settings
      timeout: "60s"
      max_retries: 3
      stream: false

      # response_format: Control output format
      # type: "text", "json_object", or "json_schema"
      response_format:
        type: json_object

    # extra_llm_parameters: Additional provider-specific parameters
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0

    exports:
      llm_analysis: "{{response.content}}"

  # ---------------------------------------------------------------------------
  # LLM step for embeddings
  # ---------------------------------------------------------------------------
  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"

    # is_embedding: Flag to indicate this is an embedding request
    is_embedding: true

    # embedding_input: List of texts to generate embeddings for
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"

    llm_config:
      model: text-embedding-3-small

    exports:
      embeddings: "{{response.embeddings}}"

  # ===========================================================================
  # COMMON STEP FIELDS: on_success, on_error, decision
  # These fields are available on ALL step types
  # ===========================================================================
  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'

    # on_success: Actions to execute when step succeeds
    on_success:
      # action: Handler type - log, abort, continue, export, run, notify
      - action: log
        message: "Step completed successfully for {{Target}}"

      - action: export
        # name: Variable name to export
        name: success_flag
        # value: Value to export (can be string, number, or template)
        value: "true"

      - action: notify
        # notify: Notification message
        notify: "Step succeeded for {{Target}}"

      - action: run
        # type: Step type to run (bash or function)
        type: bash
        command: 'echo "Running follow-up command"'

      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'

    # on_error: Actions to execute when step fails
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        # condition: Only execute this action if condition evaluates to true
        condition: 'true'

      - action: notify
        notify: "Error in workflow for {{Target}}"

      # abort: Stops workflow execution immediately
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'  # Only abort under specific conditions

      # continue: Allows workflow to continue despite error
      - action: continue
        message: "Continuing despite error"

    # decision: Conditional routing to other steps or workflow end
    decision:
      # condition: JavaScript expression to evaluate
      # next: Step name to jump to, or "_end" to finish workflow
      - condition: '{{success_flag}} == "true"'
        next: final-step

      - condition: '{{success_flag}} != "true"'
        next: _end  # Special value to end workflow

  # ---------------------------------------------------------------------------
  # Final step
  # ---------------------------------------------------------------------------
  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`,"mock-all-step-types-example":`# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: mock-all-step-types-example

# description: Human-readable description of what this workflow does
description: Mock Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - echo
    - curl

  files:
    - /tmp

  variables:
    - name: Target
      type: string
      required: true

  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: true
  disable_logging: false
  heuristics_check: 'basic'
  ci_output_format: false
  silent: false
  repeat: false
  repeat_wait_time: '60s'
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  - name: bash-single-command
    type: bash
    pre_condition: 'true'
    log: "Executing single bash command for {{Target}}"
    timeout: 60
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'
    std_file: "{{Output}}/step1-output.txt"
    exports:
      step1_result: "completed"

  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"
    command: 'echo'
    speed_args: '-n'
    config_args: ''
    input_args: '"Structured arguments test"'
    output_args: ''

  - name: function-single
    type: function
    log: "Executing single function"
    function: 'log_info("Processing {{Target}} in function step")'

  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  - name: function-parallel
    type: function
    log: "Executing functions in parallel"
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'
      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'
      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  - name: foreach-example
    type: foreach
    log: "Iterating over items"
    input: "{{Output}}/items.txt"
    variable: item
    threads: 5
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30
    url: "https://httpbin.org/post"
    method: POST
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }
    exports:
      http_response: "{{response.body}}"

  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120
    messages:
      - role: system
        content: "You are a security analysis assistant."
      - role: user
        content: "Analyze this target: {{Target}}"
    tools:
      - type: function
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target
    tool_choice: auto
    llm_config:
      provider: openai
      model: gpt-4
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0
      timeout: "60s"
      max_retries: 3
      stream: false
      response_format:
        type: json_object
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0
    exports:
      llm_analysis: "{{response.content}}"

  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"
    is_embedding: true
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"
    llm_config:
      model: text-embedding-3-small
    exports:
      embeddings: "{{response.embeddings}}"

  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'
    on_success:
      - action: log
        message: "Step completed successfully for {{Target}}"
      - action: export
        name: success_flag
        value: "true"
      - action: notify
        notify: "Step succeeded for {{Target}}"
      - action: run
        type: bash
        command: 'echo "Running follow-up command"'
      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        condition: 'true'
      - action: notify
        notify: "Error in workflow for {{Target}}"
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'
      - action: continue
        message: "Continuing despite error"
    decision:
      - condition: '{{success_flag}} == "true"'
        next: final-step
      - condition: '{{success_flag}} != "true"'
        next: _end

  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`};e.s(["f",0,t])},57763,e=>{"use strict";var t=Symbol("NOT_RESOLVED");function n(e,t){return{tagName:e,nodeKind:"scalar",implicit:t.implicit??!1,matchByTagPrefix:t.matchByTagPrefix??!1,implicitFirstChars:t.implicitFirstChars??null,resolve:t.resolve,identify:t.identify,represent:t.represent??(e=>String(e)),representTagName:t.representTagName??(()=>e)}}function r(e,t){let n=void 0===t.finalize;return{tagName:e,nodeKind:"sequence",implicit:!1,matchByTagPrefix:t.matchByTagPrefix??!1,create:t.create,addItem:t.addItem,finalize:t.finalize??(e=>e),carrierIsResult:n,identify:t.identify,represent:t.represent??(e=>e),representTagName:t.representTagName??(()=>e)}}function o(e,t){let n=void 0===t.finalize;return{tagName:e,nodeKind:"mapping",implicit:!1,matchByTagPrefix:t.matchByTagPrefix??!1,create:t.create,addPair:t.addPair,has:t.has,keys:t.keys,get:t.get,finalize:t.finalize??(e=>e),carrierIsResult:n,identify:t.identify,represent:t.represent??(e=>e),representTagName:t.representTagName??(()=>e)}}var i=n("tag:yaml.org,2002:str",{resolve:e=>e,identify:e=>"string"==typeof e}),a=["","~","null","Null","NULL"],s=n("tag:yaml.org,2002:null",{implicit:!0,implicitFirstChars:["","~","n","N"],resolve:e=>-1!==a.indexOf(e)?null:t,identify:e=>null===e,represent:()=>"null"}),l=n("tag:yaml.org,2002:null",{implicit:!0,implicitFirstChars:["n"],resolve:(e,n)=>"null"===e||n&&""===e?null:t,identify:e=>null===e,represent:()=>"null"}),c=["","~","null","Null","NULL"],u=n("tag:yaml.org,2002:null",{implicit:!0,implicitFirstChars:["","~","n","N"],resolve:e=>-1!==c.indexOf(e)?null:t,identify:e=>null===e,represent:()=>"null"}),p=["true","True","TRUE"],d=["false","False","FALSE"],m=n("tag:yaml.org,2002:bool",{implicit:!0,implicitFirstChars:["t","T","f","F"],resolve:e=>-1!==p.indexOf(e)||-1===d.indexOf(e)&&t,identify:e=>"[object Boolean]"===Object.prototype.toString.call(e),represent:e=>e?"true":"false"}),f=["true"],g=["false"],h=n("tag:yaml.org,2002:bool",{implicit:!0,implicitFirstChars:["t","f"],resolve:e=>-1!==f.indexOf(e)||-1===g.indexOf(e)&&t,identify:e=>"[object Boolean]"===Object.prototype.toString.call(e),represent:e=>e?"true":"false"}),y=["true","True","TRUE","y","Y","yes","Yes","YES","on","On","ON"],b=["false","False","FALSE","n","N","no","No","NO","off","Off","OFF"],w=n("tag:yaml.org,2002:bool",{implicit:!0,implicitFirstChars:["y","Y","n","N","t","T","f","F","o","O"],resolve:e=>-1!==y.indexOf(e)||-1===b.indexOf(e)&&t,identify:e=>"[object Boolean]"===Object.prototype.toString.call(e),represent:e=>e?"true":"false"}),v=RegExp("^(?:0o[0-7]+|0x[0-9a-fA-F]+|[-+]?[0-9]+)$"),k=RegExp("^(?:[-+]?0b[0-1]+|[-+]?0o[0-7]+|[-+]?0x[0-9a-fA-F]+|[-+]?[0-9]+)$"),x=n("tag:yaml.org,2002:int",{implicit:!0,implicitFirstChars:["-","+",..."0123456789"],resolve:function(e,n){let r,o;if(n){if(!k.test(e))return t}else if(!v.test(e))return t;let i=(o=1,(("-"===(r=e)[0]||"+"===r[0])&&("-"===r[0]&&(o=-1),r=r.slice(1)),r.startsWith("0b"))?o*parseInt(r.slice(2),2):r.startsWith("0o")?o*parseInt(r.slice(2),8):r.startsWith("0x")?o*parseInt(r.slice(2),16):o*parseInt(r,10));return Number.isFinite(i)?i:t},identify:e=>Number.isInteger(e)&&!Object.is(e,-0)&&0>e.toString(10).indexOf("e"),represent:e=>e.toString(10)}),_=RegExp("^-?(?:0|[1-9][0-9]*)$"),S=RegExp("^(?:[-+]?0b[0-1]+|[-+]?0o[0-7]+|[-+]?0x[0-9a-fA-F]+|[-+]?[0-9]+)$"),C=n("tag:yaml.org,2002:int",{implicit:!0,implicitFirstChars:["-",..."0123456789"],resolve:function(e,n){let r,o;if(n){if(!S.test(e))return t}else if(!_.test(e))return t;let i=(o=1,(("-"===(r=e)[0]||"+"===r[0])&&("-"===r[0]&&(o=-1),r=r.slice(1)),r.startsWith("0b"))?o*parseInt(r.slice(2),2):r.startsWith("0o")?o*parseInt(r.slice(2),8):r.startsWith("0x")?o*parseInt(r.slice(2),16):o*parseInt(r,10));return Number.isFinite(i)?i:t},identify:e=>Number.isInteger(e)&&!Object.is(e,-0)&&0>e.toString(10).indexOf("e"),represent:e=>e.toString(10)}),A=RegExp("^(?:[-+]?0b[0-1_]+|[-+]?0[0-7_]+|[-+]?0x[0-9a-fA-F_]+|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+|[-+]?(?:0|[1-9][0-9_]*))$"),O=n("tag:yaml.org,2002:int",{implicit:!0,implicitFirstChars:["-","+",..."0123456789"],resolve:function(e){if(!A.test(e))return t;let n=function(e){let t=e.replace(/_/g,""),n=1;if(("-"===t[0]||"+"===t[0])&&("-"===t[0]&&(n=-1),t=t.slice(1)),t.startsWith("0b"))return n*parseInt(t.slice(2),2);if(t.startsWith("0x"))return n*parseInt(t.slice(2),16);if(t.includes(":")){let e=0;for(let n of t.split(":"))e=60*e+Number(n);return n*e}return"0"!==t&&"0"===t[0]?n*parseInt(t,8):n*parseInt(t,10)}(e);return Number.isFinite(n)?n:t},identify:e=>Number.isInteger(e)&&!Object.is(e,-0)&&0>e.toString(10).indexOf("e"),represent:e=>e.toString(10)}),T=RegExp("^(?:[-+]?[0-9]+(?:\\.[0-9]*)?(?:[eE][-+]?[0-9]+)?|[-+]?\\.[0-9]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),E=RegExp("^(?:[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),F=n("tag:yaml.org,2002:float",{implicit:!0,implicitFirstChars:["-","+",".",..."0123456789"],resolve:function(e){if(!T.test(e))return t;let n=e.toLowerCase(),r="-"===n[0]?-1:1;if("+-".includes(n[0])&&(n=n.slice(1)),".inf"===n)return 1===r?1/0:-1/0;if(".nan"===n)return 0/0;let o=r*parseFloat(n);return Number.isFinite(o)||E.test(e)?o:t},identify:e=>"number"==typeof e&&(!Number.isInteger(e)||Object.is(e,-0)||e.toString(10).indexOf("e")>=0),represent:function(e){if(isNaN(e))return".nan";if(1/0===e)return".inf";if(-1/0===e)return"-.inf";if(Object.is(e,-0))return"-0.0";let t=e.toString(10);return/^[-+]?[0-9]+e/.test(t)?t.replace("e",".e"):t}}),I=RegExp("^-?(?:0|[1-9][0-9]*)(?:\\.[0-9]*)?(?:[eE][-+]?[0-9]+)?$"),N=RegExp("^(?:[-+]?[0-9]+(?:\\.[0-9]*)?(?:[eE][-+]?[0-9]+)?|[-+]?\\.[0-9]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),$=n("tag:yaml.org,2002:float",{implicit:!0,implicitFirstChars:["-",..."0123456789"],resolve:function(e,n){if(n){if(!N.test(e))return t;let n=e.toLowerCase(),r="-"===n[0]?-1:1;if("+-".includes(n[0])&&(n=n.slice(1)),".inf"===n)return 1===r?1/0:-1/0;if(".nan"===n)return 0/0;let o=r*parseFloat(n);return Number.isFinite(o)?o:t}if(!I.test(e))return t;let r=Number(e);return Number.isFinite(r)?r:t},identify:e=>"number"==typeof e&&(!Number.isInteger(e)||Object.is(e,-0)||e.toString(10).indexOf("e")>=0),represent:function(e){if(isNaN(e))return".nan";if(1/0===e)return".inf";if(-1/0===e)return"-.inf";if(Object.is(e,-0))return"-0.0";let t=e.toString(10);return/^[-+]?[0-9]+e/.test(t)?t.replace("e",".e"):t}}),P=RegExp("^(?:[-+]?(?:(?:[0-9][0-9_]*)?\\.[0-9_]*)(?:[eE][-+][0-9]+)?|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\\.[0-9_]*|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),R=RegExp("^(?:[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"),q=n("tag:yaml.org,2002:float",{implicit:!0,implicitFirstChars:["-","+",".",..."0123456789"],resolve:function(e){if(!P.test(e))return t;let n=e.toLowerCase().replace(/_/g,""),r="-"===n[0]?-1:1;if("+-".includes(n[0])&&(n=n.slice(1)),".inf"===n)return 1===r?1/0:-1/0;if(".nan"===n)return 0/0;let o=0;if(n.includes(":")){for(let e of n.split(":"))o=60*o+Number(e);o*=r}else o=r*parseFloat(n);return Number.isFinite(o)||R.test(e)?o:t},identify:e=>"number"==typeof e&&(!Number.isInteger(e)||Object.is(e,-0)||e.toString(10).indexOf("e")>=0),represent:function(e){if(isNaN(e))return".nan";if(1/0===e)return".inf";if(-1/0===e)return"-.inf";if(Object.is(e,-0))return"-0.0";let t=e.toString(10);return/^[-+]?[0-9]+e/.test(t)?t.replace("e",".e"):t}}),M=n("tag:yaml.org,2002:merge",{implicit:!0,implicitFirstChars:["<"],resolve:(e,n)=>"<<"===e||n&&""===e?"<<":t,identify:()=>!1}),j=/^[A-Za-z0-9+/]*={0,2}$/,L=n("tag:yaml.org,2002:binary",{resolve:function(e){let n=e.replace(/\s/g,"");if(n.length%4!=0||!j.test(n))return t;let r=atob(n),o=new Uint8Array(r.length);for(let e=0;e<r.length;e++)o[e]=r.charCodeAt(e);return o},identify:e=>"[object Uint8Array]"===Object.prototype.toString.call(e),represent:function(e){let t="";for(let n=0;n<e.length;n++)t+=String.fromCharCode(e[n]);return btoa(t)}}),D=RegExp("^([0-9][0-9][0-9][0-9])-([0-9][0-9])-([0-9][0-9])$"),H=RegExp("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:[Tt]|[ \\t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \\t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?$");function K(e,t,n,r=0,o=0,i=0,a=0){let s=new Date(Date.UTC(e,t,n,r,o,i,a));return s.setUTCFullYear(e,t,n),s}var U=n("tag:yaml.org,2002:timestamp",{implicit:!0,implicitFirstChars:[..."0123456789"],resolve:function(e){let n=D.exec(e);if(null===n&&(n=H.exec(e)),null===n)return t;let r=+n[1],o=n[2]-1,i=+n[3];if(!n[4]){let e=K(r,o,i);return e.getUTCFullYear()!==r||e.getUTCMonth()!==o||e.getUTCDate()!==i?t:e}let a=+n[4],s=+n[5],l=+n[6],c=0;if(a>23||s>59||l>59)return t;if(n[7]){let e=n[7].slice(0,3);for(;e.length<3;)e+="0";c=+e}let u=K(r,o,i,a,s,l,c);if(u.getUTCFullYear()!==r||u.getUTCMonth()!==o||u.getUTCDate()!==i)return t;if(n[9]){let e=+n[10],r=+(n[11]||0);if(e>23||r>59)return t;let o=(60*e+r)*6e4;u.setTime(u.getTime()-("-"===n[9]?-o:o))}return u},identify:e=>e instanceof Date,represent:e=>e.toISOString()}),W=r("tag:yaml.org,2002:seq",{create:()=>[],addItem:(e,t)=>{e.push(t)},identify:Array.isArray});function B(e){if(null===e||"object"!=typeof e||Array.isArray(e))return!1;let t=Object.getPrototypeOf(e);return null===t||t===Object.prototype}function Y(e,t){let n={};for(let r of t)void 0!==e[r]&&(n[r]=e[r]);return n}var G=r("tag:yaml.org,2002:omap",{create:()=>({list:[],seen:new Set}),addItem:(e,t)=>{let n;if(t instanceof Map){if(1!==t.size)return"cannot resolve an ordered map item";n=t.keys().next().value}else{if(!B(t))return"cannot resolve an ordered map item";let e=Object.keys(t);if(1!==e.length)return"cannot resolve an ordered map item";n=e[0]}return e.seen.has(n)?"duplicate key in ordered map":(e.seen.add(n),e.list.push(t),"")},finalize:e=>e.list,identify:()=>!1}),z=r("tag:yaml.org,2002:pairs",{create:()=>[],addItem:(e,t)=>{if(t instanceof Map)return 1!==t.size?"cannot resolve a pairs item":(e.push(t.entries().next().value),"");if("[object Object]"!==Object.prototype.toString.call(t))return"cannot resolve a pairs item";let n=Object.keys(t);return 1!==n.length?"cannot resolve a pairs item":(e.push([n[0],t[n[0]]]),"")},identify:()=>!1}),V=o("tag:yaml.org,2002:map",{create:()=>({}),identify:B,represent:e=>{let t=new Map;for(let n of Object.keys(e))t.set(n,e[n]);return t},addPair:(e,t,n)=>{if(null!==t&&"object"==typeof t)return"object-based map does not support complex keys";let r=String(t);return"__proto__"===r?Object.defineProperty(e,r,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[r]=n,""},has:(e,t)=>(null===t||"object"!=typeof t)&&Object.prototype.hasOwnProperty.call(e,String(t)),keys:e=>Object.keys(e),get:(e,t)=>{let n=String(t);return Object.prototype.hasOwnProperty.call(e,n)?e[n]:null}}),J=o("tag:yaml.org,2002:set",{create:()=>new Set,identify:e=>e instanceof Set,represent:e=>{let t=new Map;for(let n of e)t.set(n,null);return t},addPair:(e,t,n)=>null!==n?"cannot resolve a set item":(e.add(t),""),has:(e,t)=>e.has(t),keys:e=>e.keys(),get:()=>null}),Q=class e{tags;implicitScalarTags;implicitScalarByFirstChar;implicitScalarAnyFirstChar;defaultScalarTag;defaultSequenceTag;defaultMappingTag;exact;prefix;constructor(e){let t=function(e){let t=[];for(let n of e){let e=t.length;for(let r=0;r<t.length;r++){let o=t[r];if(o.nodeKind===n.nodeKind&&o.tagName===n.tagName&&o.matchByTagPrefix===n.matchByTagPrefix){e=r;break}}t[e]=n}return t}(e),n=[],r={scalar:Object.create(null),sequence:Object.create(null),mapping:Object.create(null)},o={scalar:[],sequence:[],mapping:[]};for(let e of t){if("scalar"===e.nodeKind&&e.implicit){if(e.matchByTagPrefix)throw Error("Implicit scalar tags cannot match by tag prefix");n.push(e)}switch(e.nodeKind){case"scalar":e.matchByTagPrefix?o.scalar.push(e):r.scalar[e.tagName]=e;break;case"sequence":e.matchByTagPrefix?o.sequence.push(e):r.sequence[e.tagName]=e;break;case"mapping":e.matchByTagPrefix?o.mapping.push(e):r.mapping[e.tagName]=e}}let i=n.filter(e=>null===e.implicitFirstChars),a=new Set;for(let e of n)if(null!==e.implicitFirstChars)for(let t of e.implicitFirstChars)a.add(t);let s=new Map;for(let e of a)s.set(e,n.filter(t=>null===t.implicitFirstChars||-1!==t.implicitFirstChars.indexOf(e)));let l=r.scalar["tag:yaml.org,2002:str"];if(!l)throw Error("schema does not define the default scalar tag (tag:yaml.org,2002:str)");this.tags=t,this.implicitScalarTags=n,this.implicitScalarByFirstChar=s,this.implicitScalarAnyFirstChar=i,this.defaultScalarTag=l,this.defaultSequenceTag=r.sequence["tag:yaml.org,2002:seq"],this.defaultMappingTag=r.mapping["tag:yaml.org,2002:map"],this.exact=r,this.prefix=o}lookupScalarTag(e){let t=this.exact.scalar[e];if(t)return t;for(let t of this.prefix.scalar)if(e.startsWith(t.tagName))return t}lookupSequenceTag(e){let t=this.exact.sequence[e];if(t)return t;for(let t of this.prefix.sequence)if(e.startsWith(t.tagName))return t}lookupMappingTag(e){let t=this.exact.mapping[e];if(t)return t;for(let t of this.prefix.mapping)if(e.startsWith(t.tagName))return t}resolveImplicitScalarTag(e){for(let n of this.implicitScalarByFirstChar.get(e.charAt(0))??this.implicitScalarAnyFirstChar){let r=n.resolve(e,!1,n.tagName);if(r!==t)return{value:r,tag:n}}let n=this.defaultScalarTag;return{value:n.resolve(e,!1,n.tagName),tag:n}}withTags(...t){let n=[];for(let e of t)n=n.concat(e);return new e([...this.tags,...n])}},Z=new Q([i,W,V]);new Q([...Z.tags,l,h,C,$]);var X=new Q([...Z.tags,s,m,x,F]),ee=new Q([...Z.tags,u,w,O,q,U,M,L,G,z,J]).withTags({...O,resolve:(e,n,r)=>{let o=O.resolve(e,n,r);return o===t?x.resolve(e,n,r):o}},{...q,resolve:(e,n,r)=>{let o=q.resolve(e,n,r);return o===t?F.resolve(e,n,r):o}});function et(e){if(Array.isArray(e)){let t=Array.prototype.slice.call(e);for(let e=0;e<t.length;e++){if(Array.isArray(t[e]))return null;"object"==typeof t[e]&&"[object Object]"===Object.prototype.toString.call(t[e])&&(t[e]="[object Object]")}return String(t)}return"object"==typeof e&&"[object Object]"===Object.prototype.toString.call(e)?"[object Object]":String(e)}o("tag:yaml.org,2002:map",{create:()=>new Map,addPair:(e,t,n)=>(e.set(t,n),""),has:(e,t)=>e.has(t),keys:e=>e.keys(),get:(e,t)=>e.get(t),identify:e=>e instanceof Map||B(e),represent:e=>{if(e instanceof Map)return e;let t=new Map;for(let n of Object.keys(e))t.set(n,e[n]);return t}}),o("tag:yaml.org,2002:map",{create:()=>({}),identify:B,represent:e=>{let t=new Map;for(let n of Object.keys(e))t.set(n,e[n]);return t},addPair:(e,t,n)=>{let r=et(t);return null===r?"nested arrays are not supported inside keys":("__proto__"===r?Object.defineProperty(e,r,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[r]=n,"")},has:(e,t)=>{let n=et(t);return null!==n&&Object.prototype.hasOwnProperty.call(e,n)},keys:e=>Object.keys(e),get:(e,t)=>{let n=String(t);return Object.prototype.hasOwnProperty.call(e,n)?e[n]:null}});var en={maxLength:79,indent:1,linesBefore:3,linesAfter:2};function er(e,t,n,r,o){let i="",a="",s=Math.floor(o/2)-1;return r-t>s&&(t=r-s+(i=" ... ").length),n-r>s&&(n=r+s-(a=" ...").length),{str:i+e.slice(t,n).replace(/\t/g,"→")+a,pos:r-t+i.length}}function eo(e,t){return" ".repeat(Math.max(t-e.length,0))+e}function ei(e,t){let n="";return e.mark?(e.mark.name&&(n+=`in "${e.mark.name}" `),n+=`(${e.mark.line+1}:${e.mark.column+1})`,!t&&e.mark.snippet&&(n+=`

${e.mark.snippet}`),`${e.reason} ${n}`):e.reason}var ea=class e extends Error{reason;mark;constructor(e,t){super(),this.name="YAMLException",this.reason=e,this.mark=t,this.message=ei(this,!1),Error.captureStackTrace&&Error.captureStackTrace(this,this.constructor)}toString(e){return`${this.name}: ${ei(this,e)}`}static throwAt(t,n,r,o=""){let i=0,a=0;for(let e=0;e<n;e++){let n=t.charCodeAt(e);10===n?(i++,a=e+1):13===n&&(i++,10===t.charCodeAt(e+1)&&e++,a=e+1)}let s={name:o,buffer:t,position:n,line:i,column:n-a};throw s.snippet=function(e){let t;if(!e.buffer)return null;let n={...en},r=/\r?\n|\r|\0/g,o=[0],i=[],a=-1;for(;t=r.exec(e.buffer);)i.push(t.index),o.push(t.index+t[0].length),e.position<=t.index&&a<0&&(a=o.length-2);a<0&&(a=o.length-1);let s="",l=Math.min(e.line+n.linesAfter,i.length).toString().length,c=n.maxLength-(n.indent+l+3);for(let t=1;t<=n.linesBefore&&!(a-t<0);t++){let r=er(e.buffer,o[a-t],i[a-t],e.position-(o[a]-o[a-t]),c);s=`${" ".repeat(n.indent)}${eo((e.line-t+1).toString(),l)} | ${r.str}
${s}`}let u=er(e.buffer,o[a],i[a],e.position,c);s+=`${" ".repeat(n.indent)}${eo((e.line+1).toString(),l)} | ${u.str}
${"-".repeat(n.indent+l+3+u.pos)}^
`;for(let t=1;t<=n.linesAfter&&!(a+t>=i.length);t++){let r=er(e.buffer,o[a+t],i[a+t],e.position-(o[a]-o[a+t]),c);s+=`${" ".repeat(n.indent)}${eo((e.line+t+1).toString(),l)} | ${r.str}
`}return s.replace(/\n$/,"")}(s),new e(r,s)}};function es(e){switch(e){case 48:return"\0";case 97:return"\x07";case 98:return"\b";case 116:case 9:return"	";case 110:return"\n";case 118:return"\v";case 102:return"\f";case 114:return"\r";case 101:return"\x1b";case 32:return" ";case 34:return'"';case 47:return"/";case 92:return"\\";case 78:return"";case 95:return" ";case 76:return"\u2028";case 80:return"\u2029";default:return""}}var el=Array(256),ec=Array(256);for(let e=0;e<256;e++)el[e]=+!!es(e),ec[e]=es(e);function eu(e,t,n){let r=0;for(;t<n;){let n=e.charCodeAt(t);if(10===n)r++,t++;else if(13===n)r++,t++,10===e.charCodeAt(t)&&t++;else if(32===n||9===n)t++;else break}return{position:t,breaks:r}}function ep(e){return 1===e?" ":"\n".repeat(e-1)}function ed(e,t,n,r,o,i){let a=r<0?0:r,s=e.slice(t,n).replace(/\r\n?/g,"\n"),l=""===s?[]:(s.endsWith("\n")?s.slice(0,-1):s).split("\n"),c="",u=!1,p=0,d=!1;for(let e of l){let t=0;for(;t<a&&32===e.charCodeAt(t);)t++;if(r<0||t>=e.length){p++;continue}let n=e.slice(a),o=n.charCodeAt(0);i?32===o||9===o?(d=!0,c+="\n".repeat(u?1+p:p)):d?(d=!1,c+="\n".repeat(p+1)):0===p?u&&(c+=" "):c+="\n".repeat(p):c+="\n".repeat(u?1+p:p),c+=n,u=!0,p=0}return 3===o?c+="\n".repeat(u?1+p:p):2!==o&&u&&(c+="\n"),c}var em=Object.assign(Object.create(null),{"!":"!","!!":"tag:yaml.org,2002:"}),ef="-#;/?:@&=+$_.~*'()",eg=ef+"!,[]",eh="0123456789ABCDEF";function ey(e){return"%"+eh[e>>4]+eh[15&e]}function eb(e,t){let n=t?ef:eg,r="";for(let t=0;t<e.length;t++){let o=e.charCodeAt(t);if(o>=48&&o<=57||o>=65&&o<=90||o>=97&&o<=122||n.includes(e[t])){r+=e[t];continue}o>=55296&&o<=56319&&(o=65536+((1023&o)<<10)+(1023&e.charCodeAt(++t))),o<128?r+=ey(o):o<2048?r+=ey(192|o>>6)+ey(128|63&o):o<65536?r+=ey(224|o>>12)+ey(128|o>>6&63)+ey(128|63&o):r+=ey(240|o>>18)+ey(128|o>>12&63)+ey(128|o>>6&63)+ey(128|63&o)}return r}function ew(e){let t="",n=0;for(;n<e.length;){if(37!==e.charCodeAt(n)){t+=e[n++];continue}let r=parseInt(e.slice(n+1,n+3),16);if(n+=3,r>=192){let t=r<224?1:r<240?2:3;for(r&=(1<<6-t)-1;t-- >0;)r=r<<6|63&parseInt(e.slice(n+1,n+3),16),n+=3}r<65536?t+=String.fromCharCode(r):(r-=65536,t+=String.fromCharCode(55296|r>>10,56320|1023&r))}return t}function ev(e,t){if(e.startsWith("!<")&&e.endsWith(">"))return ew(e.slice(2,-1));let n=e.indexOf("!",1),r=-1===n?"!":e.slice(0,n+1);return ew(t?.[r]??em[r]??r)+ew(e.slice(r.length))}function ek(e){let t=e;return 33===t.charCodeAt(0)?(t=t.slice(1),`!${eb(t,!0)}`):"tag:yaml.org,2002:"===t.slice(0,18)?`!!${eb(t.slice(18),!0)}`:`!<${eb(t,!1)}>`}var ex={filename:"",schema:X,json:!1,maxTotalMergeKeys:1e4,maxAliases:-1};function e_(e,t){ea.throwAt(e.source,e.position,t,e.filename)}function eS(e,t,n,r){try{return n.finalize(r)}catch(n){if(n instanceof ea)throw n;ea.throwAt(e.source,t,n instanceof Error?n.message:String(n),e.filename)}}function eC(e,t,n){let r=-1===t.tagStart?"":e.source.slice(t.tagStart,t.tagEnd);return""===r||"!"===r?n:ev(r,e.tagHandlers)}function eA(e){return"mapping"===e.nodeKind}function eO(e){e.totalMergeKeys++,-1!==e.maxTotalMergeKeys&&e.totalMergeKeys>e.maxTotalMergeKeys&&e_(e,`merge keys exceeded maxTotalMergeKeys (${e.maxTotalMergeKeys})`)}function eT(e,t,n,r){for(let o of(eO(e),r.keys(n))){if(eO(e),t.tag.has(t.value,o))continue;let i=t.tag.addPair(t.value,o,r.get(n,o));i&&e_(e,i),t.overridable??=new Set,t.overridable.add(o)}}function eE(e,t,n){let r=e.frames[e.frames.length-1];if("document"===r.kind)r.value=t,r.hasValue=!0;else if("sequence"===r.kind){eA(n)&&e.nodeTags.set(t,n);let o=r.tag.addItem(r.value,t,r.index++);o&&e_(e,o)}else if(r.hasKey){let o=r.key;r.key=void 0,r.hasKey=!1,function(e,t,n,r,o){if(e.position=t.keyPosition,t.keyIsMerge){var i=e,a=t,s=r,l=o;if(i.position=a.keyPosition,eA(l))eT(i,a,s,l);else if("sequence"===l.nodeKind&&Array.isArray(s))for(let e of(s.length>100&&e_(i,"abnormal merge sequence size"),s)){let t=i.nodeTags.get(e);t||e_(i,"cannot merge mappings; the provided source object is unacceptable"),eT(i,a,e,t)}else e_(i,"cannot merge mappings; the provided source object is unacceptable");return}!e.json&&t.tag.has(t.value,n)&&!t.overridable?.has(n)&&e_(e,"duplicated mapping key");let c=t.tag.addPair(t.value,n,r);c&&e_(e,c),t.overridable?.delete(n)}(e,r,o,t,n)}else r.key=t,r.keyPosition=e.position,r.hasKey=!0,r.keyIsMerge="tag:yaml.org,2002:merge"===n.tagName}function eF(e,t,n,r,o){if(-1!==t.anchorStart){let i={value:n,tag:r,isValueFinal:o};return e.anchors.set(e.source.slice(t.anchorStart,t.anchorEnd),i),i}return null}var eI=Object.prototype.hasOwnProperty,eN=/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/,e$=/[,\[\]{}]/,eP=/^(?:!|!!|![0-9A-Za-z-]+!)$/,eR=String.raw`(?:%[0-9A-Fa-f]{2}|[0-9A-Za-z\-#;/?:@&=+$,_.!~*'()\[\]])`,eq=String.raw`(?:%[0-9A-Fa-f]{2}|[0-9A-Za-z\-#;/?:@&=+$.~*'()_])`,eM=RegExp(`^(?:${eR})*$`),ej=RegExp(`^(?:${eq})+$`),eL=RegExp(`^(?:!(?:${eR})*|${eq}(?:${eR})*)$`),eD={filename:"",maxDepth:100};function eH(e,t,n,r,o,i,a){e.events.push({type:2,start:t,anchorStart:n,anchorEnd:r,tagStart:o,tagEnd:i,style:a})}function eK(e,t,n,r,o,i,a){e.events.push({type:3,start:t,anchorStart:n,anchorEnd:r,tagStart:o,tagEnd:i,style:a})}function eU(e,t){e.events.splice(t.eventsLength,0,{type:3,start:t.position,anchorStart:-1,anchorEnd:-1,tagStart:-1,tagEnd:-1,style:2})}function eW(e,t,n,r,o,i,a,s,l=1,c=-1,u=!1){e.events.push({type:4,valueStart:t,valueEnd:n,anchorStart:r,anchorEnd:o,tagStart:i,tagEnd:a,style:s,chomping:l,indent:c,fast:u})}function eB(e){e.events.push({type:6})}function eY(e){eW(e,-1,-1,-1,-1,-1,-1,1)}function eG(){return{anchorStart:-1,anchorEnd:-1,tagStart:-1,tagEnd:-1}}function ez(e){return{position:e.position,line:e.line,lineStart:e.lineStart,lineIndent:e.lineIndent,firstTabInLine:e.firstTabInLine,eventsLength:e.events.length}}function eV(e,t){e.position=t.position,e.line=t.line,e.lineStart=t.lineStart,e.lineIndent=t.lineIndent,e.firstTabInLine=t.firstTabInLine,e.events.length=t.eventsLength}function eJ(e,t){ea.throwAt(e.input.slice(0,e.length),e.position,t,e.filename)}function eQ(e){return 10===e||13===e}function eZ(e){return 9===e||32===e}function eX(e){return eZ(e)||eQ(e)}function e0(e){return 0===e||eX(e)}function e1(e){return 44===e||91===e||93===e||123===e||125===e}function e2(e){10===e.input.charCodeAt(e.position)?e.position++:(e.position++,10===e.input.charCodeAt(e.position)&&e.position++),e.line++,e.lineStart=e.position,e.lineIndent=0,e.firstTabInLine=-1}function e3(e,t){let n=0,r=e.input.charCodeAt(e.position),o=e.position===e.lineStart||eX(e.input.charCodeAt(e.position-1));for(;0!==r;){for(;eZ(r);)o=!0,9===r&&-1===e.firstTabInLine&&(e.firstTabInLine=e.position),r=e.input.charCodeAt(++e.position);if(t&&o&&35===r)do r=e.input.charCodeAt(++e.position);while(!eQ(r)&&0!==r)if(!eQ(r))break;for(e2(e),n++,o=!0,r=e.input.charCodeAt(e.position);32===r;)e.lineIndent++,r=e.input.charCodeAt(++e.position)}return n}function e9(e,t=e.position){let n=e.input.charCodeAt(t);if((45===n||46===n)&&n===e.input.charCodeAt(t+1)&&n===e.input.charCodeAt(t+2)){let n=e.input.charCodeAt(t+3);return 0===n||eX(n)}return!1}function e4(e){e.position===e.lineStart&&65279===e.input.charCodeAt(e.position)&&(e.position++,e.lineStart=e.position)}function e6(e){if(e.position!==e.lineStart)return!1;if(e9(e))return!0;if(65279!==e.input.charCodeAt(e.position))return!1;let t=ez(e);e4(e),e3(e,!0);let n=e.input.charCodeAt(e.position),r=e.position===e.lineStart&&(37===n||45===n&&e9(e));return eV(e,t),r}function e5(e){let t=e.input.charCodeAt(e.position);for(;0!==t&&!eQ(t);)t=e.input.charCodeAt(++e.position)}function e8(e,t,n){let r=eN.exec(e.input.slice(t,n));null!==r&&(e.position=t+r.index+r[0].length-1,eJ(e,"the stream contains non-printable characters"))}function e7(e,t){e3(e,!1),e.lineIndent<t&&eJ(e,"deficient indentation")}function te(e,t){let n=e.line;e3(e,!0),(e.line>n&&e.lineIndent<t||-1!==e.firstTabInLine&&e.lineIndent<t)&&eJ(e,"deficient indentation")}function tt(e,t,n){if(-1!==e.firstTabInLine||45!==e.input.charCodeAt(e.position)||!e0(e.input.charCodeAt(e.position+1)))return!1;for(eH(e,e.position,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,1);45===e.input.charCodeAt(e.position)&&e0(e.input.charCodeAt(e.position+1));){-1!==e.firstTabInLine&&(e.position=e.firstTabInLine,eJ(e,"tab characters must not be used in indentation"));let n=e.line;e.position++;let r=e3(e,!0)>0;if(-1!==e.firstTabInLine&&45===e.input.charCodeAt(e.position)&&e0(e.input.charCodeAt(e.position+1))&&eJ(e,"bad indentation of a sequence entry"),r&&e.lineIndent<=t?eY(e):tr(e,t,3,!1,!0),e3(e,!0),e.lineIndent<t||e.position>=e.length)break;e.lineIndent>t&&eJ(e,"bad indentation of a sequence entry"),e.line===n&&45===e.input.charCodeAt(e.position)&&e0(e.input.charCodeAt(e.position+1))&&eJ(e,"bad indentation of a sequence entry")}return eB(e),!0}function tn(e,t,n,r){let o=!1,i=!1,a=!1,s=!1;if(-1!==e.firstTabInLine)return!1;let l=e.input.charCodeAt(e.position);for(;0!==l;){o||-1===e.firstTabInLine||(e.position=e.firstTabInLine,eJ(e,"tab characters must not be used in indentation"));let c=e.input.charCodeAt(e.position+1),u=e.line;if((63===l||58===l)&&e0(c))a||(eK(e,e.position,r.anchorStart,r.anchorEnd,r.tagStart,r.tagEnd,1),a=!0),63===l?(o&&eY(e),i=!0,o=!0):(o||(eY(e),i=!0),o=!1),e.position+=1,s=!0;else{o&&(eY(e),o=!1);let t=ez(e);if(!tr(e,n,2,!1,!0))break;if(e.line===u){for(l=e.input.charCodeAt(e.position);eZ(l);)l=e.input.charCodeAt(++e.position);if(58===l)e0(l=e.input.charCodeAt(++e.position))||eJ(e,"a whitespace character is expected after the key-value separator within a block mapping"),a||(e.events.splice(t.eventsLength,0,{type:3,start:t.position,anchorStart:r.anchorStart,anchorEnd:r.anchorEnd,tagStart:r.tagStart,tagEnd:r.tagEnd,style:1}),a=!0),i=!0,o=!1,s=!1;else if(i)eJ(e,"expected ':' after a mapping key");else{if(-1!==r.anchorStart||-1!==r.tagStart)return eV(e,t),!1;return!0}}else if(i)eJ(e,"can not read a block mapping entry; a multiline key may not be an implicit key");else{if(-1!==r.anchorStart||-1!==r.tagStart)return eV(e,t),!1;return!0}}if(tr(e,t,4,!0,s)&&(s=!1),!o&&s&&(eY(e),s=!1),e3(e,!0),l=e.input.charCodeAt(e.position),(e.line===u||e.lineIndent>t)&&0!==l)eJ(e,"bad indentation of a mapping entry");else if(e.lineIndent<t)break}return!!i&&(o&&eY(e),a&&eB(e),!0)}function tr(e,t,n,r,o,i=!0){e.depth>=e.maxDepth&&eJ(e,`nesting exceeded maxDepth (${e.maxDepth})`),e.depth++;let a=1,s=!1,l=!1,c=null,u=eG(),p=4===n||3===n,d=p,m=p;if(r&&e3(e,!0)&&(s=!0,a=e.lineIndent>t?1:e.lineIndent===t?0:-1),1===a)for(;;){let r=e.input.charCodeAt(e.position),o=ez(e);if(s&&1!==a&&(33===r||38===r))break;if(s&&m&&(-1!==u.tagStart||-1!==u.anchorStart)&&(33===r||38===r)){let n=ez(e),r=t+1;if(tn(e,e.position-e.lineStart,r,u)&&e.events[n.eventsLength]?.type===3)return e.depth--,!0;eV(e,n)}if(s&&(33===r&&-1!==u.tagStart||38===r&&-1!==u.anchorStart)||!function(e,t,n){let r;if(33!==e.input.charCodeAt(e.position))return!1;-1!==t.tagStart&&eJ(e,"duplication of a tag property");let o=e.position,i=!1,a=!1,s="!",l=e.input.charCodeAt(++e.position);60===l?(i=!0,l=e.input.charCodeAt(++e.position)):33===l&&(a=!0,s="!!",l=e.input.charCodeAt(++e.position));let c=e.position;if(i){for(;0!==l&&62!==l;)l=e.input.charCodeAt(++e.position);62!==l&&eJ(e,"unexpected end of the stream within a verbatim tag"),r=e.input.slice(c,e.position),e.position++}else{for(;0!==l&&!eX(l)&&!(n&&e1(l));)33===l&&(a?eJ(e,"tag suffix cannot contain exclamation marks"):(s=e.input.slice(c-1,e.position+1),eP.test(s)||eJ(e,"named tag handle cannot contain such characters"),a=!0,c=e.position+1)),l=e.input.charCodeAt(++e.position);r=e.input.slice(c,e.position),e$.test(r)&&eJ(e,"tag suffix cannot contain flow indicator characters")}return!r||(i?eM.test(r):ej.test(r))||eJ(e,`tag name cannot contain such characters: ${r}`),i||"!"===s||"!!"===s||eI.call(e.tagHandlers,s)||eJ(e,`undeclared tag handle "${s}"`),t.tagStart=o,t.tagEnd=e.position,!0}(e,u,1===n)&&!function(e,t){if(38!==e.input.charCodeAt(e.position))return!1;-1!==t.anchorStart&&eJ(e,"duplication of an anchor property"),e.position++;let n=e.position;for(;0!==e.input.charCodeAt(e.position)&&!eX(e.input.charCodeAt(e.position))&&!e1(e.input.charCodeAt(e.position));)e.position++;return e.position===n&&eJ(e,"name of an anchor node must contain at least one character"),t.anchorStart=n,t.anchorEnd=e.position,!0}(e,u))break;null===c&&(c=o),e3(e,!0)?(s=!0,d=m,a=e.lineIndent>t?1:e.lineIndent===t?0:-1):d=!1}if(d&&(d=s||o),1===a||4===n){let r=1===n||2===n?t:t+1,o=e.position-e.lineStart;if(1===a)if(d&&(tt(e,o,u)||tn(e,o,r,u))||function(e,t,n){let r=e.input.charCodeAt(e.position),o=123===r,i=e.position,a=!0;if(91!==r&&123!==r)return!1;let s=o?125:93;for(o?eK(e,i,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,2):eH(e,i,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,2),e.position++;0!==e.input.charCodeAt(e.position);){te(e,t);let n=e.input.charCodeAt(e.position);if(n===s)return e.position++,eB(e),!0;a?44===n&&eJ(e,"expected the node content, but found ','"):eJ(e,"missed comma between flow collection entries");let r=!1,i=!1;63===n&&eX(e.input.charCodeAt(e.position+1))&&(r=i=!0,e.position+=1,te(e,t));let l=e.line,c=ez(e),u=tr(e,t,1,!1,!0);te(e,t),n=e.input.charCodeAt(e.position),(o||i||e.line===l)&&58===n?(r=!0,e.position++,te(e,t),o||eU(e,c),u||eY(e),tr(e,t,1,!1,!0)||eY(e),te(e,t),o||eB(e)):o&&r?(u||eY(e),eY(e)):o?eY(e):r&&(eU(e,c),u||eY(e),eY(e),eB(e)),44===(n=e.input.charCodeAt(e.position))?(a=!0,e.position++):a=!1}eJ(e,"unexpected end of the stream within a flow collection")}(e,r,u))l=!0;else{let t=e.input.charCodeAt(e.position);if(null!==c&&i&&m&&!d&&124!==t&&62!==t){let t=ez(e),n=c.position-c.lineStart;eV(e,c),tn(e,n,r,eG())&&e.events[t.eventsLength]?.type===3?l=!0:eV(e,t)}!l&&(p&&function(e,t,n){let r=e.input.charCodeAt(e.position),o=1,i=-1,a=!1;if(124!==r&&62!==r)return!1;for(e.position++;0!==e.input.charCodeAt(e.position);){let n=e.input.charCodeAt(e.position),r=n>=48&&n<=57?n-48:-1;if(43===n||45===n)1!==o&&eJ(e,"repeat of a chomping mode identifier"),o=43===n?3:2,e.position++;else if(r>=0)0===r&&eJ(e,"bad explicit indentation width of a block scalar; it cannot be less than one"),a&&eJ(e,"repeat of an indentation width identifier"),i=t+r-1,a=!0,e.position++;else break}let s=!1;for(;eZ(e.input.charCodeAt(e.position));)s=!0,e.position++;s&&35===e.input.charCodeAt(e.position)&&e5(e),eQ(e.input.charCodeAt(e.position))?e2(e):0!==e.input.charCodeAt(e.position)&&eJ(e,"a line break is expected");let l=a?i:-1,c=0,u=e.position,p=e.position;for(;0!==e.input.charCodeAt(e.position);){let n=e.position,r=0;for(;32===e.input.charCodeAt(n+r);)r++;let o=e.input.charCodeAt(n+r);if(0===o){l>=0?r>l&&(p=n+r):r>0&&(p=n+r);break}if(e6(e))break;if(!a&&-1===l&&eQ(o)&&(c=Math.max(c,r)),!a&&-1===l&&!eQ(o)&&(9===o&&r<t&&(e.position=n+r,eJ(e,"tab characters must not be used in indentation")),r>=t&&r<c&&(e.position=n+r,eJ(e,"bad indentation of a mapping entry"))),-1===l&&0!==o&&!eQ(o)&&r<t){e.lineIndent=r,e.position=n+r;break}a||0===o||eQ(o)||-1!==l||(l=r);let i=-1===l?t+1:l;if(0!==o&&!eQ(o)&&r<i){e.lineIndent=r,e.position=n+r;break}e5(e),p=e.position,eQ(e.input.charCodeAt(e.position))&&(e2(e),p=e.position)}return e8(e,u,p),eW(e,u,p,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,124===r?4:5,o,l),!0}(e,r,u)||function(e,t,n){if(39!==e.input.charCodeAt(e.position))return!1;e.position++;let r=e.position,o=!0;for(;0!==e.input.charCodeAt(e.position);){let i=e.input.charCodeAt(e.position);if(39===i){if(39===e.input.charCodeAt(e.position+1)){o=!1,e.position+=2;continue}let t=e.position;return e.position++,eW(e,r,t,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,2,1,-1,o),!0}eQ(i)?(o=!1,e7(e,t)):e.position===e.lineStart&&e9(e)?eJ(e,"unexpected end of the document within a single quoted scalar"):9!==i&&i<32?eJ(e,"expected valid JSON character"):e.position++}eJ(e,"unexpected end of the stream within a single quoted scalar")}(e,r,u)||function(e,t,n){if(34!==e.input.charCodeAt(e.position))return!1;e.position++;let r=e.position,o=!0;for(;0!==e.input.charCodeAt(e.position);){let i=e.input.charCodeAt(e.position);if(34===i){let t=e.position;return e.position++,eW(e,r,t,n.anchorStart,n.anchorEnd,n.tagStart,n.tagEnd,3,1,-1,o),!0}if(92===i){o=!1;let n=e.input.charCodeAt(++e.position);if(eQ(n))e7(e,t);else if(48===n||97===n||98===n||116===n||9===n||110===n||118===n||102===n||114===n||101===n||32===n||34===n||47===n||92===n||78===n||95===n||76===n||80===n)e.position++;else{let t=120===n?2:117===n?4:8*(85===n);for(0===t&&eJ(e,"unknown escape sequence");t-- >0;)e.position++,0>function(e){if(e>=48&&e<=57)return e-48;let t=32|e;return t>=97&&t<=102?t-97+10:-1}(e.input.charCodeAt(e.position))&&eJ(e,"expected hexadecimal character");e.position++}}else eQ(i)?(o=!1,e7(e,t)):e.position===e.lineStart&&e9(e)?eJ(e,"unexpected end of the document within a double quoted scalar"):9!==i&&i<32?eJ(e,"expected valid JSON character"):e.position++}eJ(e,"unexpected end of the stream within a double quoted scalar")}(e,r,u)||function(e,t){var n;if(42!==e.input.charCodeAt(e.position))return!1;(-1!==t.anchorStart||-1!==t.tagStart)&&eJ(e,"alias node should not have any properties"),e.position++;let r=e.position;for(;0!==e.input.charCodeAt(e.position)&&!eX(e.input.charCodeAt(e.position))&&!e1(e.input.charCodeAt(e.position));)e.position++;return e.position===r&&eJ(e,"name of an alias node must contain at least one character"),n=e.position,e.events.push({type:5,anchorStart:r,anchorEnd:n}),!0}(e,u)||function(e,t,n,r){if(!function(e,t){let n=e.input.charCodeAt(e.position);if(0===n||eX(n)||35===n||38===n||42===n||33===n||124===n||62===n||39===n||34===n||37===n||64===n||96===n||e1(n))return!1;if(63===n||45===n){let n=e.input.charCodeAt(e.position+1);if(e0(n)||1===t&&e1(n))return!1}return!0}(e,n))return!1;let o=e.position,i=e.position,a=e.input.charCodeAt(e.position),s=1===n,l=!1;for(;0!==a&&!e6(e);){if(58===a){let t=e.input.charCodeAt(e.position+1);if(e0(t)||s&&e1(t))break}else if(35===a){if(eX(e.input.charCodeAt(e.position-1)))break}else if(s&&e1(a))break;else if(eQ(a)){let n=e.position,r=e.line,o=e.lineStart,i=e.lineIndent;if(e3(e,!1),e.lineIndent>=t){l=!0,a=e.input.charCodeAt(e.position);continue}e.position=n,e.line=r,e.lineStart=o,e.lineIndent=i;break}eZ(a)||(i=e.position+1),a=e.input.charCodeAt(++e.position)}return i!==o&&(e8(e,o,i),eW(e,o,i,r.anchorStart,r.anchorEnd,r.tagStart,r.tagEnd,1,1,-1,!l),!0)}(e,r,n,u))&&(l=!0)}else 0===a&&(l=d&&tt(e,o,u))}return p=p&&!l,!l&&(-1!==u.anchorStart||-1!==u.tagStart||p)&&(eW(e,-1,-1,u.anchorStart,u.anchorEnd,u.tagStart,u.tagEnd,1),l=!0),e.depth--,l||-1!==u.anchorStart||-1!==u.tagStart}var to={...eD,...ex},ti=Symbol("INVALID"),ta=Symbol("visit:break"),ts=Symbol("visit:skip");function tl(e,t){for(let n of e)if(n.contents&&function e(t,n,r){let o=n(t,r);if(o===ta)return!0;if(o===ts)return!1;let i=r.depth+1;switch(t.kind){case"sequence":for(let r of t.items)if(e(r,n,{depth:i,parent:t,isKey:!1}))return!0;break;case"mapping":for(let{key:r,value:o}of t.items)if(e(r,n,{depth:i,parent:t,isKey:!0})||e(o,n,{depth:i,parent:t,isKey:!1}))return!0}return!1}(n.contents,t,{depth:0,parent:null,isKey:!1}))return}var tc={applyQuoteFlowKeysOption:function(e){!e.presenterOptions.quoteFlowKeys||e.isKey&&e.flowOnly&&1===e.style&&(e.style=3)},doubleQuoteForInvisibles:function(e){1===e.style&&/[\t\x7F-\xA0\u2028\u2029\uFEFF\uFFFE\uFFFF]/.test(e.node.value)&&(e.style=3)},doubleQuoteWhitespaceOnly:function(e){1===e.style&&/^\s+$/.test(e.node.value)&&(e.style=3)},applyForceQuotesOption:function(e){!e.presenterOptions.forceQuotes||e.isKey||1!==e.style||e.node.tag===e.presenterOptions.schema.defaultScalarTag.tagName&&(e.style=e.node.value.includes("\n")?3:tu(e))},tryLongOrMultilineAsBlock:function(e){if(1!==e.style||e.isKey)return;let t=e.node.value,n=-1!==t.indexOf("\n");if((16&e.allowedStylesMask)==0){n&&(e.style=3);return}let r=e.presenterOptions.lineWidth;if(-1===r){n&&(e.style=4);return}let o=Math.max(Math.min(r,40),r-e.shiftOfContent),i=0,a=!1;for(;i<=t.length;){let e=t.length,n=t.indexOf("\n",i);-1!==n&&(e=n);let r=t.slice(i,e);if(r.length>o&&" "!==r[0]&&/ [^ \t]/.test(r)&&(a=!0),-1===n)break;i=n+1}a?e.style=5:n&&(e.style=4)},quoteInvalidPlain:function(e){1===e.style&&(2&e.allowedStylesMask)==0&&(e.style=tu(e))},fallbackToDoubleQuoted:function(e){(e.allowedStylesMask&1<<e.style)==0&&(e.style=3)}};function tu(e){return"single"===e.presenterOptions.quoteStyle&&(4&e.allowedStylesMask)!=0?2:3}var tp="[ \\t]",td="(?:(?!(?:[\\n\\r]|\\uFEFF))[\\x09\\x0A\\x0D\\x20-\\x7E\\x85\\xA0-\\uD7FF\\uE000-\\uFFFD\\u{10000}-\\u{10FFFF}])",tm=`(?:(?!${tp})${td})`,tf="[\\x09\\x20-\\uD7FF\\uE000-\\uFFFF\\u{10000}-\\u{10FFFF}]",tg="[-?:,\\[\\]{}#&*!|>'\"%@`]",th=`(?:(?![,\\[\\]{}])${tm})`,ty=`(?:(?:(?!${tg})${tm})|[?:-](?=${tm}))`,tb=`(?:(?:(?!${tg})${tm})|[?:-](?=${th}))`,tw=`(?:(?:(?![:#])${tm})|:(?=${tm}))#*`,tv=`(?:(?:(?![:#])${th})|:(?=${th}))#*`,tk=`(?:${tp}*${tw})*`,tx=`(?:${tp}*${tv})*`,t_=`${ty}#*${tk}`,tS=`${tb}#*${tx}`,tC=`\\n+${tw}${tk}`,tA=`\\n+${tv}${tx}`,tO=`${t_}(?:${tC})*`,tT=`${tS}(?:${tA})*`,tE=RegExp(`^(?:${tO})$`,"u"),tF=RegExp(`^(?:${tT})$`,"u"),tI=RegExp(`^(?:${t_})$`,"u"),tN=RegExp(`^(?:${tS})$`,"u"),t$=RegExp(`^(?:${tf})*$`,"u"),tP=RegExp(`^(?:${tf}|\\n)*$`,"u"),tR=RegExp(`^(?:${td}|\\n)*$`,"u"),tq=/^(?:---|\.\.\.)(?=$|[ \t\n\r])/,tM=/^(?:---|\.\.\.)(?=$|[ \t\n\r])/m;function tj(e,t){let n,r=e.indexOf("\n");if(-1===r)return e;let o=" ".repeat(t),i=e.slice(0,r),a=/(\n+)([^\n]*)/g;for(a.lastIndex=r;n=a.exec(e);){let e=n[1].length,t=n[2];i+="\n".repeat(e+1)+o+t}return i}function tL(e,t){let n=" ".repeat(t),r=0,o="",i=e.length;for(;r<i;){let t,a=e.indexOf("\n",r);-1===a?(t=e.slice(r),r=i):(t=e.slice(r,a+1),r=a+1),t.length&&"\n"!==t&&(o+=n),o+=t}return o}function tD(e,t,n){let r=/^\n* /.test(e)?String(n-t):"",o="\n"===e[e.length-1];return`${r}${o&&("\n"===e[e.length-2]||"\n"===e)?"+":o?"":"-"}
`}function tH(e){return"\n"===e[e.length-1]?e.slice(0,-1):e}function tK(e){return" "===e||"	"===e}function tU(e,t){let n,r;if(""===e||tK(e[0]))return e;let o=/ [^ \t]/g,i=0,a=0,s=0,l="";for(;n=o.exec(e);)(s=n.index)-i>t&&(r=a>i?a:s,l+=`
${e.slice(i,r)}`,i=r+1),a=s;return l+="\n",e.length-i>t&&a>i?l+=`${e.slice(i,a)}
${e.slice(a+1)}`:l+=e.slice(i),l.slice(1)}var tW=/["\\\x00-\x1F\x7F-\xA0\u2028\u2029\uD800-\uDFFF\uFEFF\uFFFE\uFFFF]/gu;function tB(e){switch(e){case"\0":return"\\0";case"\x07":return"\\a";case"\b":return"\\b";case"	":return"\\t";case"\n":return"\\n";case"\v":return"\\v";case"\f":return"\\f";case"\r":return"\\r";case"\x1b":return"\\e";case'"':return'\\"';case"\\":return"\\\\";case"":return"\\N";case" ":return"\\_";case"\u2028":return"\\L";case"\u2029":return"\\P"}let t=e.charCodeAt(0),n=t.toString(16).toUpperCase();return t<=255?`\\x${"0".repeat(2-n.length)}${n}`:`\\u${"0".repeat(4-n.length)}${n}`}var tY={indent:2,seqNoIndent:!1,seqInlineFirst:!0,lineWidth:80,flowBracketPadding:!1,flowSkipCommaSpace:!1,flowSkipColonSpace:!1,quoteFlowKeys:!1,quoteStyle:"single",forceQuotes:!1,scalarStyleRules:Object.keys(tc).map(e=>Reflect.get(tc,e)),tagBeforeAnchor:!1};function tG(e,t){return`
${" ".repeat(e.indent*t)}`}function tz(e,t,n,r){let o="";for(let i=0,a=n.items.length;i<a;i+=1){let a=tJ(e,t+1,n.items[i],n,{block:!0,compact:e.seqInlineFirst,isblockseq:!0}).text;r&&""===o||(o+=tG(e,t)),""===a||10===a.charCodeAt(0)?o+="-":o+="- ",o+=a}return o}function tV(e,t,n){return"alias"===t.kind||t.tagged||void 0!==t.anchor||e.indent<2&&n>0}function tJ(e,t,n,r,o){let i;if("alias"===n.kind)return e.openEnded=!1,{text:`*${n.anchor}`,noBody:!1};let{block:a=!1,iskey:s=!1,isblockseq:l=!1}=o,c=o.compact??!1,u=void 0!==n.anchor;tV(e,n,t)&&(c=!1);let p=n.tagged,d=a&&("mapping"===n.kind||"sequence"===n.kind)&&1===n.style&&0!==n.items.length;if("mapping"===n.kind)i=d?function(e,t,n,r){let o="";for(let i=0,a=n.items.length;i<a;i+=1){let a="";r&&""===o||(a+=tG(e,t));let{key:s,value:l}=n.items[i],c=("mapping"===s.kind||"sequence"===s.kind)&&1===s.style&&0!==s.items.length||"scalar"===s.kind&&(4===s.style||5===s.style),u=c?tJ(e,t+1,s,n,{block:!0,compact:!0,isblockseq:!tV(e,s,t+1)}):tJ(e,t+1,s,n,{block:!0,compact:!0,iskey:!0}),p=u.text,d="scalar"===s.kind&&-1!==s.value.indexOf("\n"),m=p.length>1024&&/^[\s\S]{1025}/u.test(p),f=c||d||m;f&&(p&&10===p.charCodeAt(0)?a+="?":a+="? "),a+=p,f&&(a+=tG(e,t));let g=tJ(e,t+1,l,n,{block:!0,compact:f,isblockseq:f&&!tV(e,l,t+1)}).text,h="scalar"===s.kind&&u.noBody&&(s.tagged||void 0!==s.anchor),y=!f&&("alias"===s.kind||h)?" ":"";""===g||10===g.charCodeAt(0)?a+=`${y}:`:a+=`${y}: `,a+=g,o+=a}return o}(e,t,n,c):function(e,t,n){let r="";for(let{key:o,value:i}of n.items){let a="";""!==r&&(a+=`,${!e.flowSkipCommaSpace?" ":""}`);let s=tJ(e,t,o,n,{iskey:!0}),l=s.text,c=tJ(e,t,i,n,{}).text,u=e.flowSkipColonSpace||""===c?"":" ",p="scalar"===o.kind&&s.noBody&&(o.tagged||void 0!==o.anchor),d="alias"===o.kind||p?" ":"";a+=`${l}${d}:${u}${c}`,r+=a}let o=e.flowBracketPadding&&""!==r?" ":"";return`{${o}${r}${o}}`}(e,t,n);else if("sequence"===n.kind)i=d?e.seqNoIndent&&!l&&t>0?tz(e,t-1,n,c):tz(e,t,n,c):function(e,t,n){let r="";for(let o=0,i=n.items.length;o<i;o+=1){let i=tJ(e,t,n.items[o],n,{}).text;o>0&&(r+=`,${!e.flowSkipCommaSpace?" ":""}`),r+=i}let o=e.flowBracketPadding&&n.items.length>0?" ":"";return`[${o}${r}${o}]`}(e,t,n);else{let o={node:n,parent:r,level:t,isKey:s,flowOnly:!a,shiftOfParent:0===t?-1:e.indent*(t-1),shiftOfContent:e.indent*Math.max(1,t),shiftOfFirstLine:0===t?0:e.indent*t,presenterOptions:e,allowedStylesMask:0,style:n.style},l=8;for(let t of(function(e){let t=e.node.value;if(""!==t){if(!(e.isKey?e.flowOnly?tN:tI:e.flowOnly?tF:tE).test(t)||0===e.shiftOfFirstLine&&tq.test(t))return!1;if(0===e.shiftOfContent){let e=t.indexOf("\n");if(-1!==e){let n=t.slice(e+1);if(tM.test(n))return!1}}}let n=e.presenterOptions.schema.resolveImplicitScalarTag(t).tag.tagName;return(!!e.node.tagged||n===e.node.tag)&&(!!e.node.tagged||"="!==t||n!==e.presenterOptions.schema.defaultScalarTag.tagName)}(o)&&(l|=2),function(e){let t=e.node.value;if(!(e.isKey?t$:tP).test(t)||/[ \t]\n|\n[ \t]/.test(t))return!1;if(!e.isKey&&0===e.shiftOfContent){let e=t.indexOf("\n");if(-1!==e&&tM.test(t.slice(e+1)))return!1}return!0}(o)&&(l|=4),function(e){if(e.flowOnly||!tR.test(e.node.value))return!1;let t=e.shiftOfContent-e.shiftOfParent;return!(t<1||t>9&&/^\n* /.test(e.node.value)||0===e.shiftOfContent&&tM.test(e.node.value))}(o)&&(l=16|l|32),o.allowedStylesMask=l,e.scalarStyleRules))t(o);i=function(e){var t,n,r,o,i;switch(e.style){case 1:return tj((t=e).node.value,t.shiftOfContent);case 2:return n=e,`'${tj(n.node.value,n.shiftOfContent).replace(/'/g,"''")}'`;case 4:let a;return"|"+tD(a=(r=e).node.value,r.shiftOfParent,r.shiftOfContent)+tH(tL(a,r.shiftOfContent));case 5:let s,l,c;return s=(o=e).node.value,l=o.presenterOptions.lineWidth,c=1/0,-1!==l&&(c=Math.max(Math.min(l,40),l-o.shiftOfContent)),">"+tD(s,o.shiftOfParent,o.shiftOfContent)+tH(tL(function(e,t){let n,r,o=/(\n+)([^\n]*)/g,i=e.indexOf("\n");-1===i&&(i=e.length),o.lastIndex=i;let a=tU(e.slice(0,i),t),s="\n"===e[0]||tK(e[0]);for(;r=o.exec(e);){let e=r[1],o=r[2];n=""!==o&&tK(o[0]),a+=e+(s||n||""===o?"":"\n")+tU(o,t),s=n}return a}(s,c),o.shiftOfContent));case 3:return i=e,`"${i.node.value.replace(tW,tB)}"`}}(o),e.openEnded=(4===o.style||5===o.style)&&("\n"===n.value||n.value.endsWith("\n\n")),p=n.tagged||""===i&&o.flowOnly&&r?.kind==="sequence"&&!u||1!==o.style&&n.tag!==e.defaultScalarTagName}"mapping"!==n.kind&&"sequence"!==n.kind||d||(e.openEnded=!1),d&&c&&t>0&&e.indent>2&&(i=`${" ".repeat(e.indent-2)}${i}`);let m=""===i,f=i;if(p||u){let t=[],r=p?n.tagged?n.tag:ek(n.tag):null,o=u?`&${n.anchor}`:null;e.tagBeforeAnchor?(null!==r&&t.push(r),null!==o&&t.push(o)):(null!==o&&t.push(o),null!==r&&t.push(r));let a=""===i||10===i.charCodeAt(0)?"":" ";f=`${t.join(" ")}${a}${i}`}return{text:f,noBody:m}}var tQ={...tY,schema:ee,skipInvalid:!1,noRefs:!1,flowLevel:-1,sortKeys:!1,transform:()=>{}};function tZ(e,t){let n=String(e),r=String(t);return n<r?-1:+(n>r)}e.s(["w",0,function(e,t={}){let n={...tQ,...t},r=function(e,t,n={}){let r,o,i,a,s=function e(t,n){if(!t.noRefs&&null!==n&&"object"==typeof n){let e=t.refs.get(n);if(e)return void 0===e.anchor&&(e.anchor=`ref_${t.refCounter++}`),{kind:"alias",anchor:e.anchor}}let r=function(e,t){for(let n=0,r=e.representTypes.length;n<r;n+=1){let{tag:r,implicitTag:o}=e.representTypes[n];if(r.identify(t)){let e;return e=r.matchByTagPrefix?r.representTagName(t):r.tagName,{tag:r,tagName:e,implicitTag:o}}}return null}(t,n);if(!r){if(void 0===n||t.skipInvalid)return ti;throw new ea(`unacceptable kind of an object to dump ${Object.prototype.toString.call(n)}`)}let{tag:o,tagName:i,implicitTag:a}=r,s=a?i:ek(i);if("scalar"===o.nodeKind)return{kind:"scalar",tag:s,tagged:!a,style:1,value:o.represent(n)};if("sequence"===o.nodeKind){let r=o.represent(n),i={kind:"sequence",tag:s,tagged:!a,style:1,items:[]};t.noRefs||t.refs.set(n,i);for(let n=0,o=r.length;n<o;n+=1){let o=e(t,r[n]);o===ti&&void 0===r[n]&&(o=e(t,null)),o!==ti&&i.items.push(o)}return i}let l=o.represent(n),c={kind:"mapping",tag:s,tagged:!a,style:1,items:[]};for(let[r,o]of(t.noRefs||t.refs.set(n,c),l)){let n=e(t,r);if(n===ti)continue;let i=e(t,o);i!==ti&&c.items.push({key:n,value:i})}return c}({representTypes:(r=new Set([t.defaultScalarTag,t.defaultSequenceTag,t.defaultMappingTag].filter(e=>void 0!==e)),o=t.implicitScalarTags,i=t.tags.filter(e=>!("scalar"===e.nodeKind&&e.implicit)&&!r.has(e)),a=t.tags.filter(e=>r.has(e)),[...o.map(e=>({tag:e,implicitTag:!0})),...i.map(e=>({tag:e,implicitTag:!1})),...a.map(e=>({tag:e,implicitTag:!0}))]),noRefs:n.noRefs??!1,skipInvalid:n.skipInvalid??!1,refs:new Map,refCounter:0},e);return[{contents:s===ti?null:s,directives:[]}]}(e,n.schema,{noRefs:n.noRefs,skipInvalid:n.skipInvalid});if(n.flowLevel>=0&&tl(r,(e,t)=>{if(!(t.depth<n.flowLevel))return("sequence"===e.kind||"mapping"===e.kind)&&(e.style=2),ts}),n.sortKeys){let e=!0===n.sortKeys?tZ:n.sortKeys;tl(r,t=>{"mapping"===t.kind&&t.items.sort((t,n)=>e("scalar"===t.key.kind?t.key.value:"","scalar"===n.key.kind?n.key.value:""))})}return n.transform(r),function(e,t){var n;let r,o=((r={...tY,...t}).flowSkipColonSpace&&(r.quoteFlowKeys=!0),{...r,defaultScalarTagName:r.schema.defaultScalarTag.tagName,openEnded:!1}),i="",a=!1;for(let t=0;t<e.length;t+=1){let r=e[t];o.openEnded=!1;let s=function(e){let t="";for(let n of e.directives){if("yaml"===n.kind){t+=`%YAML ${n.version}
`;continue}let{handle:e,prefix:r}=n;t+=`%TAG ${e} ${r}
`}return t}(r),l=""!==s,c=r.explicitStart||l||t>0&&!a;if(i+=s,null===r.contents)c&&(i+="---\n");else if(c){let e=tJ(o,0,r.contents,null,{block:!0,compact:!0}).text,t=""===e?"":l||("sequence"===(n=r.contents).kind||"mapping"===n.kind)&&1===n.style&&0!==n.items.length&&!n.tagged&&void 0===n.anchor?"\n":" ";i+=`---${t}${e}
`}else i+=tJ(o,0,r.contents,null,{block:!0,compact:!0}).text+"\n";(a=r.explicitEnd||o.openEnded)&&(i+="...\n")}return i}(r,{...Y(n,Object.keys(tY)),schema:n.schema})},"z",0,function(e,n){let r=function(e,n={}){let r={...to,...n},o=String(e),i=Object.keys(eD),a=Object.keys(ex);return function(e,n){let r={...ex,...n,events:e,documents:[],eventIndex:0,position:0,frames:[],anchors:new Map,nodeTags:new Map,tagHandlers:Object.create(null),totalMergeKeys:0,aliasCount:0};for(;r.eventIndex<r.events.length;){let e=r.events[r.eventIndex++];switch(r.position="tagStart"in e&&-1!==e.tagStart?e.tagStart:"anchorStart"in e&&-1!==e.anchorStart?e.anchorStart:"valueStart"in e&&-1!==e.valueStart?e.valueStart:"start"in e?e.start:0,e.type){case 1:for(let t of(r.anchors=new Map,r.nodeTags=new Map,r.aliasCount=0,r.tagHandlers=Object.create(null),e.directives))"tag"===t.kind&&(r.tagHandlers[t.handle]=t.prefix);r.frames.push({kind:"document",position:r.position,value:void 0,hasValue:!1});break;case 4:{let{value:n,tag:o}=function(e,n){let r=function(e,t){if(-1===t.valueStart)return"";let{valueStart:n,valueEnd:r}=t;if(t.fast)return e.slice(n,r);switch(t.style){case 2:return function(e,t,n){let r="",o=t,i=t,a=t;for(;o<n;){let t=e.charCodeAt(o);if(39===t)r+=e.slice(i,o)+"'",o+=2,i=a=o;else if(10===t||13===t){r+=e.slice(i,a);let t=eu(e,o,n);r+=ep(t.breaks),o=i=a=t.position}else o++,32!==t&&9!==t&&(a=o)}return r+e.slice(i,n)}(e,n,r);case 3:return function(e,t,n){let r="",o=t,i=t,a=t;for(;o<n;){let t=e.charCodeAt(o);if(92===t){r+=e.slice(i,o),o++;let t=e.charCodeAt(o);if(10===t||13===t)o=eu(e,o,n).position;else if(t<256&&el[t])r+=ec[t],o++;else{var s,l;let n=120===t?2:117===t?4:8,i=0;for(;n>0;n--)o++,i=(i<<4)+((s=e.charCodeAt(o))>=48&&s<=57?s-48:(32|s)-97+10);r+=(l=i)<=65535?String.fromCharCode(l):String.fromCharCode((l-65536>>10)+55296,(l-65536&1023)+56320),o++}i=a=o}else if(10===t||13===t){r+=e.slice(i,a);let t=eu(e,o,n);r+=ep(t.breaks),o=i=a=t.position}else o++,32!==t&&9!==t&&(a=o)}return r+e.slice(i,n)}(e,n,r);case 4:return ed(e,n,r,t.indent,t.chomping,!1);case 5:return ed(e,n,r,t.indent,t.chomping,!0);default:return function(e,t,n){let r="",o=t,i=t,a=t;for(;o<n;){let t=e.charCodeAt(o);if(10===t||13===t){r+=e.slice(i,a);let t=eu(e,o,n);r+=ep(t.breaks),o=i=a=t.position}else o++,32!==t&&9!==t&&(a=o)}return r+e.slice(i,a)}(e,n,r)}}(e.source,n),o=-1===n.tagStart?"":e.source.slice(n.tagStart,n.tagEnd),i=e.schema.defaultScalarTag;if(""!==o){if("!"===o)return{value:r,tag:i};let n=ev(o,e.tagHandlers),a=e.schema.lookupScalarTag(n);if(a){let o=a.resolve(r,!0,n);return o===t&&e_(e,`cannot resolve a node with !<${n}> explicit tag`),{value:o,tag:a}}let s=e.schema.lookupMappingTag(n)??e.schema.lookupSequenceTag(n);if(s){""!==r&&e_(e,`cannot resolve a node with !<${n}> explicit tag`);let t=s.create(n);return{value:s.carrierIsResult?t:eS(e,e.position,s,t),tag:s}}e_(e,`unknown scalar tag !<${n}>`)}return 1===n.style?e.schema.resolveImplicitScalarTag(r):{value:i.resolve(r,!1,i.tagName),tag:i}}(r,e);eF(r,e,n,o,!0),eE(r,n,o);break}case 2:{let t=eC(r,e,"tag:yaml.org,2002:seq"),n=r.schema.lookupSequenceTag(t);n||e_(r,`unknown sequence tag !<${t}>`);let o=n.create(t),i=eF(r,e,o,n,n.carrierIsResult);r.frames.push({kind:"sequence",position:r.position,value:o,tag:n,anchor:i,index:0});break}case 3:{let t=eC(r,e,"tag:yaml.org,2002:map"),n=r.schema.lookupMappingTag(t);n||e_(r,`unknown mapping tag !<${t}>`);let o=n.create(t),i=eF(r,e,o,n,n.carrierIsResult);r.frames.push({kind:"mapping",position:r.position,value:o,tag:n,anchor:i,key:void 0,keyPosition:r.position,hasKey:!1,keyIsMerge:!1,overridable:null});break}case 5:{-1!==r.maxAliases&&++r.aliasCount>r.maxAliases&&e_(r,`aliases exceeded maxAliases (${r.maxAliases})`);let t=r.source.slice(e.anchorStart,e.anchorEnd),n=r.anchors.get(t);n||e_(r,`unidentified alias "${t}"`),n.isValueFinal||e_(r,`recursive alias "${t}" is not supported for tag ${n.tag.tagName} because it uses finalize()`),eE(r,n.value,n.tag);break}case 6:{let e=r.frames.pop();if("mapping"===e.kind&&e.hasKey&&(r.position=e.keyPosition,e_(r,"incomplete mapping pair in event stream")),"document"===e.kind)r.documents.push(e.value);else{let t=e.tag.carrierIsResult?e.value:eS(r,e.position,e.tag,e.value);e.anchor&&(e.anchor.value=t,e.anchor.isValueFinal=!0),eE(r,t,e.tag)}}}}return r.documents}(function(e,t){let n=e.length,r={...eD,...t,input:`${e}\0`,length:n,position:0,line:0,lineStart:0,lineIndent:0,firstTabInLine:-1,depth:0,directives:[],tagHandlers:Object.create(null),events:[]},o=e.indexOf("\0");for(-1!==o&&ea.throwAt(e,o,"null byte is not allowed in input",r.filename);r.position<r.length&&(e4(r),e3(r,!0),!(r.position>=r.length));){let e=r.position;!function(e){var t;e.directives=[],e.tagHandlers=Object.create(null);let n=!1;for(e3(e,!0);function(e){if(e.lineIndent>0||37!==e.input.charCodeAt(e.position))return!1;e.position++;let t=e.position;for(;0!==e.input.charCodeAt(e.position)&&!eX(e.input.charCodeAt(e.position));)e.position++;let n=e.input.slice(t,e.position),r=[];for(0===n.length&&eJ(e,"directive name must not be less than one character in length");0!==e.input.charCodeAt(e.position)&&!eQ(e.input.charCodeAt(e.position));){for(;eZ(e.input.charCodeAt(e.position));)e.position++;if(35===e.input.charCodeAt(e.position)||eQ(e.input.charCodeAt(e.position))||0===e.input.charCodeAt(e.position))break;let t=e.position;for(;0!==e.input.charCodeAt(e.position)&&!eX(e.input.charCodeAt(e.position));)e.position++;r.push(e.input.slice(t,e.position))}if(eQ(e.input.charCodeAt(e.position))&&e2(e),"YAML"===n){e.directives.some(e=>"yaml"===e.kind)&&eJ(e,"duplication of %YAML directive"),1!==r.length&&eJ(e,"YAML directive accepts exactly one argument");let t=/^([0-9]+)\.([0-9]+)$/.exec(r[0]);null===t&&eJ(e,"ill-formed argument of the YAML directive"),1!==parseInt(t[1],10)&&eJ(e,"unacceptable YAML version of the document"),e.directives.push({kind:"yaml",version:r[0]})}else if("TAG"===n){2!==r.length&&eJ(e,"TAG directive accepts exactly two arguments");let[t,n]=r;eP.test(t)||eJ(e,"ill-formed tag handle (first argument) of the TAG directive"),eI.call(e.tagHandlers,t)&&eJ(e,`there is a previously declared suffix for "${t}" tag handle`),eL.test(n)||eJ(e,"ill-formed tag prefix (second argument) of the TAG directive"),e.tagHandlers[t]=n,e.directives.push({kind:"tag",handle:t,prefix:n})}return!0}(e);)n=!0,e3(e,!0);let r=!1,o=!1,i=!0;if(0===e.lineIndent&&45===e.input.charCodeAt(e.position)&&45===e.input.charCodeAt(e.position+1)&&45===e.input.charCodeAt(e.position+2)&&e0(e.input.charCodeAt(e.position+3))){r=!0;let t=e.line;e.position+=3,e3(e,!0),i=e.line>t}else n&&eJ(e,"directives end mark is expected");let a=e.events.length;if(!r&&e.position===e.lineStart&&46===e.input.charCodeAt(e.position)&&e9(e)){e.position+=3,e3(e,!0);return}if(t=r,e.events.push({type:1,explicitStart:t,explicitEnd:!1,directives:e.directives}),tr(e,e.lineIndent-1,4,!1,i,i)||eY(e),e3(e,!0),e.position===e.lineStart&&e9(e)&&(o=46===e.input.charCodeAt(e.position))){let t=e.line;e.position+=3,e3(e,!0),e.line===t&&e.position<e.length&&eJ(e,"end of the stream or a document separator is expected")}let s=e.events[a];s?.type===1&&(s.explicitEnd=o),eB(e),o||!(e.position<e.length)||e6(e)||eJ(e,"end of the stream or a document separator is expected")}(r),r.position===e&&eJ(r,"can not read a document")}return r.events}(o,Y(r,i)),{...Y(r,a),source:o})}(e,n);if(0===r.length)throw new ea("expected a document, but the input is empty");if(1===r.length)return r[0];throw new ea("expected a single document in the stream, but found more")}])}]);