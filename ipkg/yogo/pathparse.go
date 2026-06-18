package yogo

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"regexp"
)

// LoadJSON 读取文件，不存在则使用默认JSON初始化
func LoadJSON(path string, defaultJSON string, out *map[string]interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return json.Unmarshal([]byte(defaultJSON), out)
		}
		return err
	}
	return json.Unmarshal(data, out)
}

// SaveJSON 格式化写入JSON文件
func SaveJSON(path string, data map[string]interface{}) error {
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

// PrintJSON 格式化打印JSON（dryrun调试）
func PrintJSON(data map[string]interface{}) {
	buf, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(buf))
}

// PrintDebugConfig 打印完整配置JSON
func PrintDebugConfig(cfg Config) {
	fmt.Println("===== DEBUG CONFIG =====")
	b, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(b))
	fmt.Println("========================\n")
}

// ProcessArray 数组逻辑：去重、移除exclude、追加include、排序
func ProcessArray(arr []interface{}, include, exclude, reInclude, reExclude []string) []interface{} {
	// fmt.Printf("[DEBUG ProcessArray] raw arr len=%d\n", len(arr))
	// fmt.Printf("[DEBUG ProcessArray] include list: %#v\n", include)
	// fmt.Printf("[DEBUG ProcessArray] exclude list: %#v\n", exclude)
	// fmt.Printf("[DEBUG ProcessArray] regex-include patterns: %#v\n", reInclude)
	// fmt.Printf("[DEBUG ProcessArray] regex-exclude patterns: %#v\n", reExclude)

	set := make(map[string]bool)
	var allOld []string
	for _, item := range arr {
		if str, ok := item.(string); ok {
			set[str] = true
			allOld = append(allOld, str)
			// fmt.Printf("[DEBUG add old] %s\n", str)
		}
	}

	// 1. 精确删除
	for _, ex := range exclude {
		delete(set, ex)
		// fmt.Printf("[DEBUG exact del] %s\n", ex)
	}

	// 2. 正则批量删除
	for _, patStr := range reExclude {
		re, err := regexp.Compile(patStr)
		if err != nil {
			fmt.Printf("[WARN invalid regex %q skip: %v\n", patStr, err)
			continue
		}
		for _, s := range allOld {
			if re.MatchString(s) {
				delete(set, s)
				// fmt.Printf("[DEBUG regex del pat=%q match=%s\n", patStr, s)
			}
		}
	}

	// 3. 精确新增
	for _, inc := range include {
		set[inc] = true
		// fmt.Printf("[DEBUG exact add] %s\n", inc)
	}

	// 4. 正则新增：无意义（正则无法凭空生成字符串）
	// re-include 一般不用，只保留 re-exclude 做批量清理

	// 排序输出
	var list []string
	for k := range set {
		list = append(list, k)
	}
	sort.Strings(list)
	// fmt.Printf("[DEBUG final set keys] %#v\n", list)

	out := make([]interface{}, len(list))
	for i, s := range list {
		out[i] = s
	}
	return out
}