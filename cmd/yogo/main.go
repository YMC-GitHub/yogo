package main

import (
	"fmt"
	"os"
	"path/filepath"
	"flag"

	"github.com/ymc-github/yogo/ipkg/yogo"
)

func main() {
	cfg := yogo.ParseFlags()

	if cfg.Version {
		fmt.Println("v1.2.0")
		return
	}
	if cfg.PrintFeature {
		yogo.PrintFeature()
		return
	}
	if cfg.Help {
		yogo.PrintHelp()
		return
	}

	posArgs := flag.Args()
	if len(posArgs) == 0 {
		// 无位置参数，自动兜底 api=edit-json-arr
		cfg.Api = "edit-json-arr"
	} else {
		// 传入了第一个位置参数，校验api合法性
		cfg.Api = posArgs[0]
		if err := cfg.CheckAPI(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// 校验必填参数 --name
	if err := cfg.CheckRequired(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// 拼接完整文件路径
	var targetFile string
	if filepath.IsAbs(cfg.File) {
		// --file 传入绝对路径，直接使用，不拼接 workspace
		targetFile = cfg.File
	} else {
		// 相对路径才拼接工作目录
		targetFile = filepath.Join(cfg.Workspace, cfg.File)
	}
	if cfg.Debug {
		yogo.PrintDebugConfig(cfg)
		fmt.Printf("targetFile: %s\n", targetFile)
	}

	// 加载JSON文件
	var root map[string]interface{}
	if err := yogo.LoadJSON(targetFile, cfg.DefaultText, &root); err != nil {
		fmt.Fprintf(os.Stderr, "load json failed: %v\n", err)
		os.Exit(1)
	}

	// 定位目标数组
	// targetArrPtr, err := yogo.GetTargetArray(root, cfg.Ns, cfg.NsSep, cfg.Name, cfg.CreateMissing)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "locate array failed(ns=%s,name=%s): %v\n", cfg.Ns, cfg.Name, err)
	// 	os.Exit(1)
	// }

	// // 解析增减列表
	// includeList := yogo.SplitBySep(cfg.Include, cfg.Sep)
	// excludeList := yogo.SplitBySep(cfg.Exclude, cfg.Sep)
	// newArr := yogo.ProcessArray(*targetArrPtr, includeList, excludeList)
	// *targetArrPtr = newArr

	// 接收父对象map + 原始数组
	parentObj, rawArr, err := yogo.GetTargetArray(root, cfg.Ns, cfg.NsSep, cfg.Name, cfg.CreateMissing)
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate array failed(ns=%s,name=%s): %v\n", cfg.Ns, cfg.Name, err)
		os.Exit(1)
	}

	includeList := yogo.SplitBySep(cfg.Include, cfg.Sep)
	excludeList := yogo.SplitBySep(cfg.Exclude, cfg.Sep)

	reIncList := yogo.SplitBySep(cfg.RegexInclude, cfg.Sep)
	reExcList := yogo.SplitBySep(cfg.RegexExclude, cfg.Sep)
	
	newArr := yogo.ProcessArray(rawArr, includeList, excludeList, reIncList, reExcList)
	// newArr := yogo.ProcessArray(rawArr, includeList, excludeList)
	parentObj[cfg.Name] = newArr

	// 干跑模式仅打印，不落地文件
	if cfg.DryRun {
		fmt.Println("=== DRY RUN, NO FILE WRITE ===")
		yogo.PrintJSON(root)
		return
	}

	// 写入更新后的JSON
	if err := yogo.SaveJSON(targetFile, root); err != nil {
		fmt.Fprintf(os.Stderr, "write json failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("successfully updated file: %s\n", targetFile)
}