package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/app"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/asset"
	assetCommon "github.com/Autumn-27/ScopeSentry/internal/services/assets/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/crawler"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/dirscan"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/ip"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/mp"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/page_monitoring"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/root_domain"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/sensitive"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/subdomain"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/url"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/vulnerability"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	"github.com/Autumn-27/ScopeSentry/internal/services/plugin"
	"github.com/Autumn-27/ScopeSentry/internal/services/project"
	taskCommon "github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/template"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type deps struct {
	projectService    project.Service
	taskService       task.Service
	taskCommonService taskCommon.Service
	templateService   template.Service
	assetService      asset.Service
	rootDomainService root_domain.Service
	subdomainService  subdomain.Service
	appService        app.Service
	mpService         mp.Service
	urlService        url.Service
	sensitiveService  sensitive.Service
	dirscanService    dirscan.Service
	crawlerService    crawler.Service
	vulnService       vulnerability.Service
	ipService         ip.Service
	pageMonService    page_monitoring.Service
	commonService     assetCommon.Service
	nodeService       node.Service
	pluginService     plugin.Service
}

var d = &deps{
	projectService:    project.NewService(),
	taskService:       task.NewService(),
	taskCommonService: taskCommon.NewService(),
	templateService:   template.NewService(),
	assetService:      asset.NewService(),
	rootDomainService: root_domain.NewService(),
	subdomainService:  subdomain.NewService(),
	appService:        app.NewService(),
	mpService:         mp.NewService(),
	urlService:        url.NewService(),
	sensitiveService:  sensitive.NewService(),
	dirscanService:    dirscan.NewService(),
	crawlerService:    crawler.NewService(),
	vulnService:       vulnerability.NewService(),
	ipService:         ip.NewService(),
	pageMonService:    page_monitoring.NewService(),
	commonService:     assetCommon.NewService(),
	nodeService:       node.NewService(),
	pluginService:     plugin.NewService(),
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects grouped by tag",
	}, listProjects)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects_data",
		Description: "Get a paginated list of projects with search support",
	}, listProjectsData)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project",
		Description: "Get project details by project ID",
	}, getProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new project. name and a valid target are required; tag, template, node etc. are optional. If the name already exists and the normalized target still contains targets, only that project's asset scope is replaced.",
	}, createProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "Get a paginated list of scan tasks",
	}, listTasks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Get task details by task ID",
	}, getTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scan_templates",
		Description: "Get a paginated list of scan templates",
	}, listScanTemplates)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_scan_template",
		Description: "Get scan template details by template ID",
	}, getScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugin_modules",
		Description: "Get all module names (pipeline stages) of scan templates. A scan template is composed of these modules, and each module has several plugins attached.",
	}, listPluginModules)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugins",
		Description: "Get available scan plugins, optionally filtered by module. Returns each plugin's hash, name, module and default parameter. When creating a scan template, the module fields reference plugin hashes.",
	}, listPlugins)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_template",
		Description: createScanTemplateToolDesc,
	}, createScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_task",
		Description: createScanTaskToolDesc,
	}, createScanTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Description: listAssetsToolDesc,
	}, listAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "count_assets",
		Description: countAssetsToolDesc,
	}, countAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset_detail",
		Description: "Get asset details; asset_type supports asset or vulnerability",
	}, getAssetDetail)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_asset_tag",
		Description: "Add a tag to an asset",
	}, addAssetTag)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nodes",
		Description: "Get the list of scan nodes; when online_only=true only online nodes are returned",
	}, listNodes)
}

type listProjectsDataInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Fuzzy search keyword for project name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Items per page, default 20"`
}

func listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectsByTag(c)
	if err != nil {
		return errorResult("Failed to get project list", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listProjectsData(ctx context.Context, _ *mcp.CallToolRequest, input listProjectsDataInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectsData(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("Failed to get project data", err)
	}
	return jsonToolResult(result)
}

type getProjectInput struct {
	ID string `json:"id" jsonschema:"Project MongoDB ObjectID"`
}

func getProject(ctx context.Context, _ *mcp.CallToolRequest, input getProjectInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectContent(c, input.ID)
	if err != nil {
		return errorResult("Failed to get project details", err)
	}
	if result == nil {
		return errorResult("Project does not exist", nil)
	}
	return jsonToolResult(result)
}

type createProjectInput struct {
	Name           string   `json:"name" jsonschema:"Project name, required"`
	Tag            string   `json:"tag,omitempty" jsonschema:"Project tag, used for grouping"`
	Target         string   `json:"target" jsonschema:"Scan targets, required. Separate by newlines or commas; domain/IP/URL etc. are supported"`
	Template       string   `json:"template,omitempty" jsonschema:"Associated scan template ID"`
	Node           []string `json:"node,omitempty" jsonschema:"List of specified scan node names"`
	AllNode        bool     `json:"allNode,omitempty" jsonschema:"Whether to use all nodes"`
	Ignore         string   `json:"ignore,omitempty" jsonschema:"Ignore list of targets, same format as target"`
	Duplicates     string   `json:"duplicates,omitempty" jsonschema:"Deduplication strategy"`
	ScheduledTasks bool     `json:"scheduledTasks,omitempty" jsonschema:"Whether to enable scheduled scanning"`
	Hour           int      `json:"hour,omitempty" jsonschema:"Scheduled scan interval (hours), only effective when scheduledTasks is enabled"`
}

func createProject(ctx context.Context, _ *mcp.CallToolRequest, input createProjectInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" || strings.TrimSpace(input.Target) == "" {
		return errorResult("name and target cannot be empty", nil)
	}
	p := &models.Project{
		Name:           input.Name,
		Tag:            input.Tag,
		Target:         input.Target,
		Template:       input.Template,
		Node:           input.Node,
		AllNode:        input.AllNode,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Tp:             "project",
	}
	c := ginContext(ctx)
	id, replaced, err := d.projectService.ApplyProjectScope(c, p)
	if err != nil {
		return errorResult("Failed to create project", err)
	}
	if replaced {
		return jsonToolResult(ginH{
			"success":  true,
			"message":  "Project scope has been replaced",
			"id":       id,
			"replaced": true,
		})
	}
	return jsonToolResult(ginH{
		"success":  true,
		"message":  "Project created successfully",
		"id":       id,
		"replaced": false,
	})
}

type listTasksInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Fuzzy search on task name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Items per page, default 20"`
}

func listTasks(ctx context.Context, _ *mcp.CallToolRequest, input listTasksInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	tasks, total, err := d.taskService.List(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("Failed to get task list", err)
	}
	return jsonToolResult(ginH{"list": tasks, "total": total})
}

type getTaskInput struct {
	ID string `json:"id" jsonschema:"Task MongoDB ObjectID"`
}

func getTask(ctx context.Context, _ *mcp.CallToolRequest, input getTaskInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := d.taskService.GetTaskDetail(c, input.ID)
	if err != nil {
		return errorResult("Failed to get task details", err)
	}
	if result == nil {
		return errorResult("Task does not exist", nil)
	}
	return jsonToolResult(result)
}

type listScanTemplatesInput struct {
	Query     string `json:"query,omitempty" jsonschema:"Fuzzy search on template name"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting from 1, default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Items per page, default 20"`
}

func listScanTemplates(ctx context.Context, _ *mcp.CallToolRequest, input listScanTemplatesInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	result, err := d.templateService.List(ctx, input.PageIndex, input.PageSize, input.Query)
	if err != nil {
		return errorResult("Failed to get template list", err)
	}
	return jsonToolResult(result)
}

type getScanTemplateInput struct {
	ID string `json:"id" jsonschema:"Scan template MongoDB ObjectID"`
}

func getScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input getScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	result, err := d.templateService.Detail(c, input.ID)
	if err != nil {
		return errorResult("Failed to get template details", err)
	}
	return jsonToolResult(result)
}

type createScanTemplateInput struct {
	Name         string                       `json:"name" jsonschema:"Template name, required"`
	Modules      map[string][]string          `json:"modules,omitempty" jsonschema:"Mapping from module name to plugin hash list. Key is the module name (get it via list_plugin_modules), value is the array of plugin hashes to enable under that module (get them via list_plugins; executed in array order). Example {\"SubdomainScan\":[\"d60ba73c...\"]}"`
	Parameters   map[string]map[string]string `json:"parameters,omitempty" jsonschema:"Optional; overrides plugin run parameters. Structure: module name -> plugin hash -> parameter string. Defaults to the plugin's own parameters when omitted"`
	VulList      []string                     `json:"vullist,omitempty" jsonschema:"Optional; nuclei POC template ID list, only effective when VulnerabilityScan uses nuclei"`
	TemplateJSON string                       `json:"template_json,omitempty" jsonschema:"Optional; a full ScanTemplate JSON. When provided it takes precedence over modules, for advanced customization"`
}

func createScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input createScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" && input.TemplateJSON == "" {
		return errorResult("Provide at least one of name or template_json", nil)
	}

	var tmpl *models.ScanTemplate

	switch {
	case input.TemplateJSON != "":
		tmpl = &models.ScanTemplate{}
		if err := json.Unmarshal([]byte(input.TemplateJSON), tmpl); err != nil {
			return errorResult("Invalid template_json format", err)
		}
		if tmpl.Name == "" {
			tmpl.Name = input.Name
		}
	case len(input.Modules) > 0:
		built, err := buildTemplateFromModules(ctx, input)
		if err != nil {
			return errorResult(err.Error(), nil)
		}
		tmpl = built
	default:
		tmpl = &models.ScanTemplate{Name: input.Name}
	}

	if tmpl.Name == "" {
		return errorResult("Template name cannot be empty", nil)
	}

	id, err := d.templateService.Save(ctx, "", tmpl)
	if err != nil {
		return errorResult("Failed to create template", err)
	}
	return jsonToolResult(ginH{"success": true, "id": id, "message": "Template created successfully; it can be used in create_scan_task's template parameter"})
}

// buildTemplateFromModules assembles a scan template from module -> plugin-hash lists and backfills plugin default parameters
func buildTemplateFromModules(ctx context.Context, input createScanTemplateInput) (*models.ScanTemplate, error) {
	validModule := make(map[string]bool, len(constants.PLUGINSMODULES))
	for _, m := range constants.PLUGINSMODULES {
		validModule[m] = true
	}

	templateMap := map[string]any{
		"name": input.Name,
	}
	paramMap := map[string]map[string]string{}

	for module, hashes := range input.Modules {
		if !validModule[module] {
			return nil, fmt.Errorf("Invalid module name: %s (use list_plugin_modules to get valid modules)", module)
		}
		if len(hashes) == 0 {
			continue
		}
		templateMap[module] = hashes

		modParams := map[string]string{}
		for _, hash := range hashes {
			// Prefer parameters provided by the caller
			if input.Parameters != nil {
				if mp, ok := input.Parameters[module]; ok {
					if v, ok := mp[hash]; ok {
						modParams[hash] = v
						continue
					}
				}
			}
			// Otherwise backfill the plugin's default parameters
			plg, err := d.pluginService.GetPluginByHash(ctx, hash)
			if err != nil || plg == nil {
				return nil, fmt.Errorf("Plugin hash does not exist: %s (module %s)", hash, module)
			}
			if plg.Module != module {
				return nil, fmt.Errorf("Plugin %s(hash=%s) belongs to module %s and cannot be placed in module %s", plg.Name, hash, plg.Module, module)
			}
			modParams[hash] = plg.Parameter
		}
		if len(modParams) > 0 {
			paramMap[module] = modParams
		}
	}

	if len(paramMap) > 0 {
		templateMap["Parameters"] = paramMap
	}
	if len(input.VulList) > 0 {
		templateMap["vullist"] = input.VulList
	}

	data, err := json.Marshal(templateMap)
	if err != nil {
		return nil, fmt.Errorf("Failed to assemble template: %w", err)
	}
	tmpl := &models.ScanTemplate{}
	if err := json.Unmarshal(data, tmpl); err != nil {
		return nil, fmt.Errorf("Failed to assemble template: %w", err)
	}
	return tmpl, nil
}

type createScanTaskInput struct {
	Name           string              `json:"name" jsonschema:"Task name, required, must be unique"`
	Target         string              `json:"target,omitempty" jsonschema:"Scan targets, required when targetSource is general; separate by newlines or commas"`
	Node           []string            `json:"node" jsonschema:"List of node names that execute the scan, required"`
	Template       string              `json:"template,omitempty" jsonschema:"Scan template ObjectID, required to run the scan"`
	AllNode        bool                `json:"allNode,omitempty" jsonschema:"Whether to automatically include all online nodes"`
	Ignore         string              `json:"ignore,omitempty" jsonschema:"Ignore list of targets, same format as target"`
	Duplicates     string              `json:"duplicates,omitempty" jsonschema:"Deduplication strategy, e.g. None"`
	Project        []string            `json:"project,omitempty" jsonschema:"List of associated project ObjectIDs, required when targetSource is project"`
	TargetSource   string              `json:"targetSource,omitempty" jsonschema:"Target source; one of general, project, asset, RootDomain, subdomain, UrlScan and their Source suffixes; default general. See the tool description for details"`
	TargetTp       string              `json:"targetTp,omitempty" jsonschema:"Selection method when the source is Source: search or select"`
	Search         string              `json:"search,omitempty" jsonschema:"search expression to filter targets from the asset library; same syntax as list_assets. See the tool description for details"`
	Filter         map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter, can be combined with search; filter.project is a project ObjectID"`
	TargetNumber   int                 `json:"targetNumber,omitempty" jsonschema:"Maximum number of targets in search mode; 0 means no limit"`
	TargetIds      []string            `json:"targetIds,omitempty" jsonschema:"List of selected asset ObjectIDs in select mode"`
	BindProject    string              `json:"bindProject,omitempty" jsonschema:"Bound project ObjectID that owns the scan results"`
	ScheduledTasks bool                `json:"scheduledTasks,omitempty" jsonschema:"Whether to create it as a scheduled task"`
	Hour           int                 `json:"hour,omitempty" jsonschema:"Scheduled task interval - hours"`
	Minute         int                 `json:"minute,omitempty" jsonschema:"Scheduled task interval - minutes"`
	Day            int                 `json:"day,omitempty" jsonschema:"Scheduled task interval - days"`
	Week           int                 `json:"week,omitempty" jsonschema:"Scheduled task interval - weeks (weekly cycle)"`
	CycleType      string              `json:"cycleType,omitempty" jsonschema:"Cycle type, e.g. nhours, daily, weekly"`
}

func createScanTask(ctx context.Context, _ *mcp.CallToolRequest, input createScanTaskInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" || len(input.Node) == 0 {
		return errorResult("name and node cannot be empty", nil)
	}
	c := ginContext(ctx)
	exists, err := d.taskService.CheckTaskNameExists(c, input.Name)
	if err != nil {
		return errorResult("Failed to check task name", err)
	}
	if exists {
		return errorResult("Task name already exists", nil)
	}

	targetSource := input.TargetSource
	if targetSource == "" {
		targetSource = "general"
	}

	taskModel := &models.Task{
		Name:           input.Name,
		Target:         input.Target,
		Node:           input.Node,
		Template:       input.Template,
		AllNode:        input.AllNode,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		Project:        input.Project,
		TargetSource:   targetSource,
		TargetTp:       input.TargetTp,
		Search:         input.Search,
		TargetNumber:   input.TargetNumber,
		TargetIds:      input.TargetIds,
		BindProject:    input.BindProject,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Minute:         input.Minute,
		Day:            input.Day,
		Week:           input.Week,
		CycleType:      input.CycleType,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		taskModel.Filter = filter
	}
	taskID, err := d.taskCommonService.Insert(ctx, taskModel)
	if err != nil {
		return errorResult("Failed to create scan task", err)
	}
	return jsonToolResult(ginH{"success": true, "id": taskID})
}

type listAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required. One of asset, RootDomain, subdomain, app, mp, UrlScan, SensitiveResult, DirScanResult, crawler, vulnerability, PageMonitoring, IPAsset, SubdomainTakerResult"`
	PageIndex        int                 `json:"pageIndex,omitempty" jsonschema:"Page number, starting from 1, default 1"`
	PageSize         int                 `json:"pageSize,omitempty" jsonschema:"Items per page, default 20"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search expression, not SQL; see the tool description for syntax"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter JSON, combinable with search. filter.project is a project ObjectID (from list_projects, not the project name); filter.task is the task name. See the tool description for details"`
	Sort             map[string]string   `json:"sort,omitempty" jsonschema:"Sort order. Only UrlScan/DirScanResult support length: ascending for ascending, other values for descending"`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive-info rule name, only effective when asset_type is SensitiveResult"`
}

type countAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required. Same values as list_assets"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search expression, same syntax as list_assets"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filter JSON, same syntax as list_assets; filter.project is a project ObjectID"`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive-info rule name, only effective when asset_type is SensitiveResult"`
}

func countAssets(ctx context.Context, _ *mcp.CallToolRequest, input countAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	query := models.SearchRequest{
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	total, err := d.commonService.TotalData(ctx, &query)
	if err != nil {
		return errorResult("Failed to count assets", err)
	}
	return jsonToolResult(ginH{"total": total})
}

func listAssets(ctx context.Context, _ *mcp.CallToolRequest, input listAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := models.SearchRequest{
		PageIndex:        input.PageIndex,
		PageSize:         input.PageSize,
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sort:             input.Sort,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	c := ginContext(ctx)
	data, err := queryAssets(ctx, c, index, query)
	if err != nil {
		return errorResult("Failed to query assets", err)
	}
	return jsonToolResult(data)
}

type getAssetDetailInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type. Detail lookup supports asset or vulnerability"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID; for vulnerability type pass the hash value"`
}

func getAssetDetail(ctx context.Context, _ *mcp.CallToolRequest, input getAssetDetailInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id cannot be empty", nil)
	}
	c := ginContext(ctx)
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	switch index {
	case "asset":
		result, err := d.assetService.GetAssetByID(c, input.ID)
		if err != nil {
			return errorResult("Failed to get asset details", err)
		}
		if result == nil {
			return errorResult("Asset does not exist", nil)
		}
		return jsonToolResult(result)
	case "vulnerability":
		result, err := d.vulnService.GetVulnerabilityDetailByHash(c, input.ID)
		if err != nil {
			return errorResult("Failed to get vulnerability details", err)
		}
		return jsonToolResult(result)
	default:
		return errorResult("asset_type only supports detail lookup for asset or vulnerability", nil)
	}
}

type addAssetTagInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type, same as list_assets' asset_type"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID"`
	Tag       string `json:"tag" jsonschema:"Name of the tag to add"`
}

func addAssetTag(ctx context.Context, _ *mcp.CallToolRequest, input addAssetTagInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.ID == "" || input.Tag == "" {
		return errorResult("id and tag cannot be empty", nil)
	}
	c := ginContext(ctx)
	req := &models.TagRequest{Type: index, ID: input.ID, Tag: input.Tag}
	if err := d.commonService.AddTag(c, req); err != nil {
		return errorResult("Failed to add tag", err)
	}
	return jsonToolResult(ginH{"success": true})
}

type listNodesInput struct {
	OnlineOnly bool `json:"online_only,omitempty" jsonschema:"When true only online nodes are returned; default false returns all"`
}

func listNodes(ctx context.Context, _ *mcp.CallToolRequest, input listNodesInput) (*mcp.CallToolResult, any, error) {
	result, err := d.nodeService.GetNodeData(ctx, input.OnlineOnly)
	if err != nil {
		return errorResult("Failed to get node list", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listPluginModules(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return jsonToolResult(ginH{"modules": constants.PLUGINSMODULES})
}

type listPluginsInput struct {
	Module string `json:"module,omitempty" jsonschema:"Filter by module name; leave empty to return all scan plugins. Get module names via list_plugin_modules"`
	Search string `json:"search,omitempty" jsonschema:"Fuzzy search on plugin name (only applies when module is not specified)"`
}

type pluginBrief struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Parameter string `json:"parameter"`
	Type      string `json:"type"`
	Status    bool   `json:"status"`
}

func toPluginBriefs(plugins []models.Plugin) []pluginBrief {
	briefs := make([]pluginBrief, 0, len(plugins))
	for _, p := range plugins {
		// Skip server plugins; they are not part of the scan pipeline
		if p.Type == "server" {
			continue
		}
		briefs = append(briefs, pluginBrief{
			Hash:      p.Hash,
			Name:      p.Name,
			Module:    p.Module,
			Parameter: p.Parameter,
			Type:      p.Type,
			Status:    p.Status,
		})
	}
	return briefs
}

func listPlugins(ctx context.Context, _ *mcp.CallToolRequest, input listPluginsInput) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)

	if input.Module != "" {
		plugins, err := d.pluginService.ListByModule(c, input.Module)
		if err != nil {
			return errorResult("Failed to get plugin list", err)
		}
		return jsonToolResult(ginH{"list": toPluginBriefs(plugins)})
	}

	resp, err := d.pluginService.List(c, &models.PluginListRequest{
		PageIndex: 1,
		PageSize:  500,
		Search:    input.Search,
	})
	if err != nil {
		return errorResult("Failed to get plugin list", err)
	}
	return jsonToolResult(ginH{"list": toPluginBriefs(resp.List), "total": resp.Total})
}

func queryAssets(ctx context.Context, c *gin.Context, index string, query models.SearchRequest) (any, error) {
	switch index {
	case "asset":
		list, err := d.assetService.GetAssets(c, query)
		return ginH{"list": list}, err
	case "RootDomain":
		return d.rootDomainService.GetRootDomainData(c, query)
	case "subdomain":
		list, err := d.subdomainService.GetSubdomains(ctx, query)
		return ginH{"list": list}, err
	case "app":
		return d.appService.GetAppData(c, query)
	case "mp":
		return d.mpService.GetMPData(c, query)
	case "UrlScan":
		list, err := d.urlService.GetURLs(ctx, query)
		return ginH{"list": list}, err
	case "SensitiveResult":
		list, err := d.sensitiveService.GetSensitiveInfo(ctx, query)
		return ginH{"list": list}, err
	case "DirScanResult":
		list, err := d.dirscanService.List(ctx, query)
		return ginH{"list": list}, err
	case "crawler":
		list, err := d.crawlerService.GetCrawlers(ctx, query)
		return ginH{"list": list}, err
	case "vulnerability":
		return d.vulnService.GetVulnerabilities(ctx, query)
	case "PageMonitoring":
		return d.pageMonService.GetResult(c, query)
	case "IPAsset":
		return d.ipService.GetIPAssets(c, query)
	case "SubdomainTakerResult":
		return querySubdomainTaker(ctx, query)
	default:
		return nil, fmt.Errorf("Unsupported asset type: %s", index)
	}
}

func querySubdomainTaker(ctx context.Context, query models.SearchRequest) (any, error) {
	takerService := subdomain.NewTakerService()
	c := ginContext(ctx)
	return takerService.GetSubdomainTakerData(c, query)
}

func normalizeAssetIndex(assetType string) (string, error) {
	assetType = strings.TrimSpace(assetType)
	if assetType == "" {
		return "", fmt.Errorf("asset_type cannot be empty")
	}
	aliases := map[string]string{
		"asset":                "asset",
		"web":                  "asset",
		"rootdomain":           "RootDomain",
		"root_domain":          "RootDomain",
		"root-domain":          "RootDomain",
		"subdomain":            "subdomain",
		"app":                  "app",
		"mp":                   "mp",
		"miniprogram":          "mp",
		"mini_program":         "mp",
		"url":                  "UrlScan",
		"urlscan":              "UrlScan",
		"sensitive":            "SensitiveResult",
		"sensitiveresult":      "SensitiveResult",
		"sensitive_result":     "SensitiveResult",
		"dirscan":              "DirScanResult",
		"dir_scan":             "DirScanResult",
		"dirscanresult":        "DirScanResult",
		"directory":            "DirScanResult",
		"crawler":              "crawler",
		"vulnerability":        "vulnerability",
		"vuln":                 "vulnerability",
		"pagemonitoring":       "PageMonitoring",
		"page_monitoring":      "PageMonitoring",
		"ip":                   "IPAsset",
		"ipasset":              "IPAsset",
		"subdomaintaker":       "SubdomainTakerResult",
		"subdomain_taker":      "SubdomainTakerResult",
		"subdomaintakerresult": "SubdomainTakerResult",
	}
	key := strings.ToLower(assetType)
	if v, ok := aliases[key]; ok {
		return v, nil
	}
	if _, ok := aliases[strings.ReplaceAll(key, "-", "_")]; ok {
		return aliases[strings.ReplaceAll(key, "-", "_")], nil
	}
	// Allow passing a MongoDB collection name directly
	valid := []string{"asset", "RootDomain", "subdomain", "app", "mp", "UrlScan",
		"SensitiveResult", "DirScanResult", "crawler", "vulnerability",
		"PageMonitoring", "IPAsset", "SubdomainTakerResult"}
	for _, v := range valid {
		if v == assetType {
			return v, nil
		}
	}
	return "", fmt.Errorf("Unsupported asset_type: %s", assetType)
}

type ginH map[string]any
