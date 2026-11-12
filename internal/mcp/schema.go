package mcp

// Descriptions for the asset query tools (written into Tool.Description for MCP clients)
const listAssetsToolDesc = `List ScopeSentry assets.

[Interface and request body]
Consistent with each asset POST endpoint (e.g. /api/assets/asset); the body is models.SearchRequest:
- pageIndex, pageSize: pagination
- search: search expression string (corresponds to the frontend Csearch search box searchParams)
- filter: exact filter object (corresponds to frontend project/task dropdowns, statistic sidebar clicks, table column filters, merged before being sent)
- sort: sorting (ineffective for most types)
- sid: only for SensitiveResult, sensitive rule name

The MCP asset_type is mapped to the backend Index (e.g. asset->"asset"), then helper.GetSearchQuery builds the MongoDB query.

[Frontend parameter division of labor (Csearch.vue + asset pages)]
- search: the DSL the user types in the search box, e.g. domain=baidu && port==443
- filter.project: ElTreeSelect selected values, an array of project ObjectIDs (not project names)
- filter.task: ElSelect selected values, an array of task names (not task IDs); dynamic tags task=name may also write into filter
- filter may also come from: statistic sidebar clicks (port/service/app/icon), table column filters (statuscode/level/type/status etc.)
- Note: although the searchKeywordsData of each asset page has project hints, the backend SearchToMongoDB has not registered the project keyword; project can only go through filter

asset_type options:
- asset: Web/port assets
- RootDomain: root domains
- subdomain: subdomains
- app: mobile applications
- mp: mini programs
- UrlScan: URLs
- SensitiveResult: sensitive information
- DirScanResult: directory scan
- crawler: crawler
- vulnerability: vulnerabilities
- PageMonitoring: page monitoring
- IPAsset: aggregated IP assets
- SubdomainTakerResult: subdomain takeover

[search expression] (not SQL; a custom DSL parsed by helper.SearchToMongoDB)
- field=value : fuzzy match (case-insensitive)
- field=="value" : exact match; values containing spaces must be quoted with double quotes
- field!="value" : exclusion match
- expr1 && expr2 : AND
- expr1 || expr2 : OR
- (expr) : grouping
- Common search fields (all types, injected by SearchToMongoDB): tag->tags, task->taskName (task name), rootDomain
- project is not supported in search; it can only be passed via filter.project (value is a project ObjectID, obtained via list_projects / list_projects_data)

Available search keywords per asset_type -> MongoDB fields:
- asset: domain->host, ip, port, service, app->technologies, title, statuscode, icon->faviconmmh3, banner->metadata, type, body, header->rawheaders
- RootDomain: domain, icp, company
- subdomain: domain->host, ip, type, value
- app: name, icp, company, category, description, url, apk
- mp: name, icp, company, category, description, url
- UrlScan: url->output, input, source, resultId, type->outputtype (no statuscode search; HTTP status code only via filter.status)
- SensitiveResult: url, sname->sid, body, info->match, md5
- DirScanResult: url, statuscode->status, redirect->msg, length
- vulnerability: url, vulname, matched, request, response, level
- crawler: url, method, body, resultId
- PageMonitoring: url, hash, diff, response
- IPAsset: ip, domain->ports.server.domain, port->ports.port, service->ports.server.service, webServer->ports.server.webServer, app->ports.server.technologies
- SubdomainTakerResult: domain->input, value, type->cname, response

search examples:
- domain=baidu && port==443
- port==443 && service=nginx
- title="admin login" || body=admin

[filter exact filtering] (can be combined with search; helper.GetSearchQuery appends $and conditions)
A JSON object whose keys are filter dimensions and whose values are string arrays (multiple values for the same key are ORed, different keys are ANDed).
Only keys defined in filterKeyCache take effect; other keys are ignored.

filter key points:
- project: filter only; the value is an array of project ObjectIDs (the database stores IDs; some endpoints convert them to project names when displaying)
- task: available via search (task=="task name") or filter.task (an array of task names); the value is the name from list_tasks, not a task ID
- Multiple values for the same key are ORed, different keys are ANDed; keys not in filterKeyCache are ignored

Global filter keys -> MongoDB fields:
- project->project, port->port, service->service, app->technologies
- icon->faviconmmh3, statuscode->statuscode, status->status
- level->level, type->type, color->color, tags->tags
- task->taskName, sname->sid

Available filters per asset_type (keys not listed are ineffective for that collection or the field does not exist):
- asset: project, port, service, app, icon, statuscode, type, task, tags
- RootDomain: project, tags
- subdomain: project, type, task, tags
- app / mp: project, tags
- UrlScan: status (HTTP status code, not statuscode), tags
- SensitiveResult: status (1 unprocessed/2 processing/3 ignored/4 suspected/5 confirmed/6 processed), color, sname, tags
- DirScanResult: status (HTTP status code), tags
- crawler: project, task, tags
- vulnerability: project, level (critical/high/medium/low/info/unknown), status (1-6), task, tags
- PageMonitoring: tags
- IPAsset: project, port, service, app (nested ports fields, via aggregation queries)
- SubdomainTakerResult: tags

filter examples (project must be an ObjectID, not a project name):
- asset: {"project":["<project ObjectID>"],"port":["443"]}
- subdomain: {"project":["<project ObjectID>"],"type":["A"]}
- vulnerability: {"project":["<project ObjectID>"],"level":["high"]}

search + filter combined example:
- search: domain=baidu && port==443, filter: {"project":["<project ObjectID>"]}

Notes:
- Do not write project=... or project=="..." in search (invalid, or errors when combined with &&)
- For DirScanResult the HTTP status code can only be used in search as statuscode (e.g. statuscode==200)
- UrlScan has no statuscode search; filter results by status using filter.status
- SensitiveResult rule name filtering: use sname=rule name in search, or filter.sname

[sort]
- UrlScan / DirScanResult: {"length":"ascending"} for ascending; any other value (including "descending", "-1") means descending
- Other types: the server always sorts by time or _id; the sort parameter is usually ineffective

[sid parameter]
Only for SensitiveResult: pass the sensitive rule name (sid) to expand match details by rule (corresponds to clicking a rule name in the frontend)`

// createScanTemplateToolDesc scan template creation description
const createScanTemplateToolDesc = `Create a scan template.

A scan template consists of multiple [modules] (pipeline stages), each module mounts several [plugins],
and the module fields reference the plugin hash (not the plugin id, nor the plugin name).

Recommended creation flow:
1. list_plugin_modules to get all module names
2. list_plugins (optionally filtered by module) to get the hash and default parameter of the plugins available in each module
3. Use the modules parameter to specify module->plugin hash list (within the same module, executed in array order)
4. This tool automatically fills in Parameters using the plugin default parameters; use parameters to override if custom parameters are needed

Parameter description:
- name: template name, required
- modules: {module name: [plugin hash,...]}, e.g. {"SubdomainScan":["d60ba73c..."],"PortScan":["..."]}
- parameters: optional, {module name: {plugin hash: parameter string}}, overrides default parameters
- vullist: optional, list of nuclei POC template IDs (when VulnerabilityScan uses nuclei)
- template_json: optional, full ScanTemplate JSON, highest priority, for advanced customization

On success, the template id is returned and can be used directly for the template parameter of create_scan_task.

Common modules: TargetHandler, SubdomainScan, SubdomainSecurity, PortScanPreparation,
PortScan, PortFingerprint, AssetMapping, AssetHandle, URLScan, WebCrawler,
URLSecurity, DirScan, VulnerabilityScan, PassiveScan`

// createScanTaskToolDesc scan task creation description (consistent with the Web /api/task/add and common.Insert logic)
const createScanTaskToolDesc = `Create a scan task. name and node are required; template is the scan template ObjectID (obtained via list_scan_templates).

[targetSource] determines how the task targets are resolved (corresponds to internal/services/task/common/common.go):
- general: use the target field directly (domains, IPs, URLs separated by newlines/commas)
- project: read targets from the linked project; fill project (an array of project ObjectIDs), target is not needed
- asset: select from a search of the Web asset library, search required; project, filter, targetNumber optional
- RootDomain: select from a search of the root domain library, search required; project, filter, targetNumber optional
- subdomain: select from a search of the subdomain library, search required; project, filter, targetNumber optional
- UrlScan: select from a search of URL scan results, search required; project, filter, targetNumber optional
- assetSource / RootDomainSource / subdomainSource / UrlScanSource: create a task from the corresponding asset page
  - targetTp=search: filter targets using search + filter + project + targetNumber
  - targetTp=select: specify the asset ObjectID list using targetIds

[search] search expression, same syntax as list_assets (e.g. task=="some task name", domain=^example.com).
Example of continuing a scan from subdomains: targetSource=subdomain, search=task=="subdomain collection task name"

[filter] exact filter JSON, can be combined with search; filter.project is a project ObjectID.
[targetNumber] the maximum number of targets in search mode, 0 means no limit.
[targetIds] the list of selected asset ObjectIDs in select mode.

[Other parameters]
- allNode: automatically join all online nodes
- ignore / duplicates: ignore targets, deduplication strategy
- bindProject: bind a project (result ownership)
- scheduledTasks + cycleType/hour/minute/day/week: scheduled tasks

Two phases are recommended for complete root domain information collection: first use general + only SubdomainScan/SubdomainSecurity to scan root domains;
then use subdomain + search=task=="previous task name" to create the task for subsequent modules.`

// countAssetsToolDesc asset count statistics (corresponds to POST /api/assets/common/total)
const countAssetsToolDesc = `Count the assets matching the criteria (the same backend endpoint as the Web pagination "N items total", /api/assets/common/total).

The parameters are the same search/filter as list_assets, but without pagination; only total is returned.
asset_type values are the same as list_assets (asset, RootDomain, subdomain, vulnerability, etc.).

Example: count the subdomains under a project
{"asset_type":"subdomain","filter":{"project":["<project ObjectID>"]}}

Example: count the Web assets produced by a task
{"asset_type":"asset","search":"task==\"some task name\""}`
