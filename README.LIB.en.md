# yogo - Go Library Documentation for JSON Array Operations
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ymc-github/yogo)](https://goreportcard.com/report/github.com/ymc-github/yogo)

`ipkg/yogo` is a lightweight pure Go underlying utility library without heavy third-party dependencies. It provides full capabilities including JSON file read/write, nested path parsing, bulk array item add/remove, and CLI argument parsing. The upper-layer CLI program `cmd/yogo` is fully wrapped based on this library.

## ✨ Library Features
- Stateless pure function design with isolated input and output, friendly for unit testing
- Full support for multi-level nested JSON path parsing, compatible with special syntax for array indexes and custom object keys
- Built-in standardized array processing: deduplication, sorting, bulk addition and bulk deletion
- Intelligent differentiation between absolute/relative file paths to avoid root path loss bugs of standard library `filepath.Join`
- Auto initialization for missing files and auto creation of missing intermediate nodes
- Built-in debug print and formatted JSON output utility functions
- Decoupled argument parsing module, embeddable into any Go program without mandatory CLI dependency

## 📦 Import Library
```bash
go get github.com/ymc-github/yogo/ipkg/yogo
```
```go
import "github.com/ymc-github/yogo/ipkg/yogo"
```

## 📚 Module Structure Overview
1. `config.go`: CLI argument struct, argument parsing, helper/version/feature print functions
2. `jsonutil.go`: JSON file I/O, string splitting, core array processing logic
3. `pathparse.go`: Nested path parsing, path traversal, target array and parent object lookup

## 1. config.go API Reference
### Config Struct
```go
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
	Sep           string
	Ns            string
	NsSep         string
	CreateMissing bool
	Api           string
}
```

#### ParseFlags
Parse all CLI arguments, bind short & long aliases, return complete config instance
```go
func ParseFlags() Config
```

#### CheckAPI
Validate subcommand legitimacy; only `edit-json-arr` is allowed, returns error for invalid input
```go
func (cfg Config) CheckAPI() error
```

#### CheckRequired
Validate mandatory flag `--name`, returns error if missing
```go
func (cfg Config) CheckRequired() error
```

#### PrintHelp
Print complete CLI usage docs, path syntax, parameter table and business examples
```go
func PrintHelp()
```

#### PrintFeature
Print full list of built-in features, used for `--print-feature`
```go
func PrintFeature()
```

#### PrintDebugConfig
Format and output config JSON for debug mode
```go
func PrintDebugConfig(cfg Config)
```

## 2. jsonutil.go API Reference
### LoadJSON
Read JSON file; initialize empty root object with `defaultText` if file does not exist
```go
func LoadJSON(path string, defaultText string, out *map[string]interface{}) error
```

### SaveJSON
Write formatted indented JSON to file with file permission `0644`
```go
func SaveJSON(path string, root map[string]interface{}) error
```

### PrintJSON
Pretty-print map-style JSON, dedicated for dry-run preview
```go
func PrintJSON(root map[string]interface{})
```

### SplitBySep
Split string by specified separator, auto trim whitespace and filter empty segments
```go
func SplitBySep(s, sep string) []string
```

### ProcessArray
Core array handler integrating add, remove, deduplicate and sort logic
```go
func ProcessArray(arr []interface{}, include, exclude []string) []interface{}
```
Execution workflow:
1. Iterate original array and store items in set for auto deduplication
2. Remove all string entries matched in exclude list
3. Append all new items from include list
4. Sort strings lexicographically and convert back to `[]interface{}` for return

## 3. pathparse.go API Reference
### PathSegment Struct
Store metadata for a single path segment, distinguish plain key, numeric array index and custom object key
```go
type PathSegment struct {
	Key    string
	ArrIdx *int    // [number] numeric array index
	ObjKey *string // [text] custom object bracket key
}
```

#### ParseNsSegments
Split full nested path by separator, parse special `[number]` / `[string]` bracket syntax
```go
func ParseNsSegments(nsPath, nsSep string) ([]PathSegment, error)
```

#### GetTargetArray
Traverse nested path layer by layer to locate target array; return **parent object map, raw source array, error**
```go
func GetTargetArray(root map[string]interface{}, nsPath, nsSep, arrName string, createMissing bool) (map[string]interface{}, []interface{}, error)
```
Core logic:
1. Parse ns path segment by segment, create empty object nodes automatically if `createMissing` is enabled and nodes are missing
2. Support access via array index and custom object bracket key
3. Locate final parent object and read target array field specified by `--name`
4. Generate empty array and write to parent map if target array missing and auto creation enabled

## 🚀 Library Code Usage Examples
### Example 1: Modify root-level array insecure-registries in daemon.json
```go
package main

import (
	"fmt"
	"github.com/ymc-github/yogo/ipkg/yogo"
)

func main() {
	filePath := "/etc/docker/daemon.json"
	var root map[string]interface{}
	// Load JSON, fallback empty object if file missing
	err := yogo.LoadJSON(filePath, "{}", &root)
	if err != nil {
		fmt.Printf("Failed to load file: %v\n", err)
		return
	}

	// Locate root-level array, pass empty string for ns path
	parentObj, rawArr, err := yogo.GetTargetArray(root, "", ".", "insecure-registries", true)
	if err != nil {
		fmt.Printf("Failed to locate target array: %v\n", err)
		return
	}

	// Split add and remove value lists
	incList := yogo.SplitBySep("mirror.test.com,docker.ustc.edu.cn", ",")
	excList := yogo.SplitBySep("mirrors.sohu.com", ",")

	// Process array items
	newArr := yogo.ProcessArray(rawArr, incList, excList)
	// Write updated array back to parent object to sync JSON data
	parentObj["insecure-registries"] = newArr

	// Persist changes to disk
	err = yogo.SaveJSON(filePath, root)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}
	fmt.Println("Configuration updated successfully")
}
```

### Example 2: Operate multi-level nested JSON array
```go
package main

import (
	"fmt"
	"github.com/ymc-github/yogo/ipkg/yogo"
)

func main() {
	var root map[string]interface{}
	_ = yogo.LoadJSON("app.json", "{}", &root)

	// Nested path config.docker[0].mirror, target array field urls
	parent, arr, err := yogo.GetTargetArray(root, "config.docker[0].mirror", ".", "urls", true)
	if err != nil {
		fmt.Println("Path parsing failed:", err)
		return
	}

	newArr := yogo.ProcessArray(arr, []string{"new.mirror.com"}, []string{})
	parent["urls"] = newArr

	_ = yogo.SaveJSON("app.json", root)
	fmt.Println("Nested array modified successfully")
}
```

## 🧪 Unit Testing
```bash
# Run full unit test suite for library
go test -v ./ipkg/yogo
# Generate test coverage report
go test -cover ./ipkg/yogo
```

## Design Specifications
1. All low-level functions follow pure input-output pattern with zero global variables and side effects
2. Path parsing and file I/O modules are fully decoupled and reusable independently
3. Core array processing logic is isolated from CLI arguments, embeddable into any business program
4. Errors are returned hierarchically for upper layers to customize user prompt messages

## 📄 Open Source License
MIT License or Apache 2.0 License

## 📮 Repository Address
https://github.com/ymc-github/yogo