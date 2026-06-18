package yogo

import (
	"flag"
	"fmt"
)

// Config 全量命令行配置
type Config struct {
	Version       bool
	Help          bool
	PrintFeature  bool
	Workspace     string
	File          string
	DryRun        bool
	Debug         bool
	DefaultText   string
	Name          string
	Include       string
	Exclude       string
	RegexExclude string
	RegexInclude string
	Sep           string
	Ns            string
	NsSep         string
	CreateMissing bool
	Api           string
}

// ParseFlags 解析所有flag参数
func ParseFlags() Config {
	var cfg Config
	flag.BoolVar(&cfg.Version, "v", false, "info version")
	flag.BoolVar(&cfg.Version, "version", false, "info version")

	flag.BoolVar(&cfg.Help, "h", false, "info help")
	flag.BoolVar(&cfg.Help, "help", false, "info help")

	flag.BoolVar(&cfg.PrintFeature, "print-feature", false, "print all supported features")


	flag.StringVar(&cfg.Workspace, "w", "./", "set the workspace location")
	flag.StringVar(&cfg.Workspace, "workspace", "./", "set the workspace location")

	flag.StringVar(&cfg.File, "file", "package.json", "set the file location")
	flag.BoolVar(&cfg.DryRun, "dryrun", false, "set the dry-run mode")
	flag.BoolVar(&cfg.Debug, "debug", false, "set the debug mode")
	flag.StringVar(&cfg.DefaultText, "default-text", "{}", "set the default text for file")

	flag.StringVar(&cfg.Ns, "ns", "", "multi-level namespace path, support app[0] / app[config] / app[\"key\"]")
	flag.StringVar(&cfg.Name, "name", "", "target array field name (required)")
	flag.StringVar(&cfg.NsSep, "ns-sep", ".", "namespace path separator")
	flag.BoolVar(&cfg.CreateMissing, "create-missing", false, "auto create missing object nodes & target array field")

	flag.StringVar(&cfg.Include, "include", "", "values to add into array")
	flag.StringVar(&cfg.Exclude, "exclude", "", "values to remove from array")
	flag.StringVar(&cfg.Sep, "sep", ",", "separator for include/exclude list")
	
	flag.StringVar(&cfg.RegexExclude, "regex-exclude", "", "regex patterns to remove, split by --sep")
	flag.StringVar(&cfg.RegexInclude, "regex-include", "", "regex patterns to add, split by --sep")

	flag.Parse()
	return cfg
}

// PrintHelp 打印完整帮助文本
func PrintHelp() {
	helpText := `
Usage: yogo [edit-json-arr] [option]
  Two call styles supported:
  1. Omit api (auto fallback to edit-json-arr): yogo --ns "app[config]" --name keywords ...
  2. Explicit api: yogo edit-json-arr --ns "app[config]" --name keywords ...

Supported --ns syntax demo:
  1. Array numeric index: --ns "app[0]"
  2. Object key shorthand: --ns "app[config]"
  3. Quoted object key: --ns "app[\"config.name\"]"
  4. Mix multi-level: --ns "data.list[1].meta[tags]"
  5. Normal plain path: --ns "project.sub.config"
  6. Direct root field: no --ns

Command demo:
  # Omit api (recommended shorthand)
  yogo --ns "app[config]" --name keywords --create-missing --include "nano,utxt"
  # Full explicit api
  yogo edit-json-arr --ns "app[config]" --name keywords --create-missing --include "nano,utxt"

Rule:
  1. [number] = access array index
  2. [text] / ["text"] = access object property key
  3. --create-missing only auto create OBJECT nodes, cannot expand array
  4. array index out of range returns error

Options:
  -v,--version        boolean   info version (default:false)
  -h,--help           boolean   info help (default:false)
  --print-feature     boolean   print all supported features (default:false)
  -w,--workspace      string    set the workspace location (default:./)
  --file              string    set the file location (default:package.json)
  --dryrun            boolean   set the dry-run mode (default:false)
  --debug             boolean   set the debug mode (default:false)
  --default-text      string    set the default text for file (default:{})
  --ns                string    multi-level namespace path, support app[0] / app[config] syntax
  --name              string    target array field name, cooperate with --ns (required)
  --ns-sep            string    namespace path separator (default:.)
  --create-missing    boolean   auto create missing object nodes & target array (default:false)
  --include           string    values add to array
  --exclude           string    values remove from array
  --sep               string    split sep for include/exclude (default:,)
`
	fmt.Print(helpText)
}

// PrintFeature 输出全部支持特性清单
func PrintFeature() {
	featureText := `
=== YOGO edit-json-arr Supported Features ===
1. Call Mode
   - Shorthand: yogo [opts] (auto fallback api=edit-json-arr)
   - Full mode: yogo edit-json-arr [opts]
2. NS Path Syntax
   - Plain path: ns="app.config.sub"
   - Array index: ns="list[0].meta"
   - Object bracket key: ns="app[config]" / ns="app[\"key.name\"]"
   - Custom separator via --ns-sep
3. Array Operation
   - --include: append values auto deduplicate & sort
   - --exclude: remove specified values
   - --sep: custom split separator for include/exclude list
4. Auto Create
   - --create-missing: auto create missing object nodes & empty target array
5. File Control
   - --workspace: custom work dir, join with --file
   - --default-text: init content when file not exists
6. Debug & Dry Run
   - --dryrun: preview json without write file
   - --debug: print full parsed config
7. Aux Flags
   - -v/--version: print version
   - -h/--help: print usage help
   - --print-feature: show this feature list
`
	fmt.Print(featureText)
}
// CheckRequired 校验必填参数
func (cfg Config) CheckRequired() error {
	if cfg.Name == "" {
		return fmt.Errorf("--name is required, cooperate with --ns")
	}
	return nil
}

// CheckAPI 校验api仅支持edit-json-arr
func (cfg Config) CheckAPI() error {
	if cfg.Api != "edit-json-arr" {
		return fmt.Errorf("unsupported api '%s', only edit-json-arr allowed", cfg.Api)
	}
	return nil
}