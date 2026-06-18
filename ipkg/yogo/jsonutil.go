package yogo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PathSegment 路径片段：支持数字下标 / 对象字符串key
type PathSegment struct {
	Key    string
	ArrIdx *int    // 非nil = 数组数字下标
	ObjKey *string // 非nil = 对象字符串键 [config] / ["config"]
}

var segRegex = regexp.MustCompile(`^(?P<key>.+)\[(?P<inner>.+)\]$`)
var numRegex = regexp.MustCompile(`^\d+$`)

// SplitBySep 通用分割工具，清理空与空格
func SplitBySep(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

// stripQuote 去除首尾单/双引号
func stripQuote(s string) string {
	s = strings.TrimSpace(s)
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		return s[1 : len(s)-1]
	}
	return s
}

// ParseNsSegments 解析 app[0] / app[config] / app["key"]
func ParseNsSegments(nsPath, nsSep string) ([]PathSegment, error) {
	rawParts := SplitBySep(nsPath, nsSep)
	var segs []PathSegment
	for _, part := range rawParts {
		match := segRegex.FindStringSubmatch(part)
		if len(match) != 3 {
			// 普通无括号key
			segs = append(segs, PathSegment{Key: part})
			continue
		}
		key := match[1]
		inner := stripQuote(match[2])

		// 判断括号内是数字下标 还是 对象key
		if numRegex.MatchString(inner) {
			idx, err := strconv.Atoi(inner)
			if err != nil {
				return nil, fmt.Errorf("invalid numeric index %s in %s", inner, part)
			}
			segs = append(segs, PathSegment{
				Key:    key,
				ArrIdx: &idx,
			})
		} else {
			// 对象字符串key [config]
			segs = append(segs, PathSegment{
				Key:    key,
				ObjKey: &inner,
			})
		}
	}
	return segs, nil
}

// GetTargetArray 遍历ns路径，定位目标数组指针
func GetTargetArray(root map[string]interface{}, nsPath, nsSep, arrName string, createMissing bool) (map[string]interface{}, []interface{}, error) {
	currentNode := interface{}(root)

	if nsPath != "" {
		segments, err := ParseNsSegments(nsPath, nsSep)
		if err != nil {
			return nil, nil, err
		}
		for _, seg := range segments {
			switch node := currentNode.(type) {
			case map[string]interface{}:
				val, exists := node[seg.Key]
				if !exists {
					if !createMissing {
						return nil, nil, fmt.Errorf("node key [%s] not exists, enable --create-missing to create object", seg.Key)
					}
					newObj := make(map[string]interface{})
					node[seg.Key] = newObj
					currentNode = newObj
					continue
				}

				if seg.ArrIdx != nil {
					arr, ok := val.([]interface{})
					if !ok {
						return nil, nil, fmt.Errorf("key [%s] is not array, cannot index %d", seg.Key, *seg.ArrIdx)
					}
					idx := *seg.ArrIdx
					if idx < 0 || idx >= len(arr) {
						return nil, nil, fmt.Errorf("array [%s] index %d out of range(len=%d)", seg.Key, idx, len(arr))
					}
					currentNode = arr[idx]
				} else if seg.ObjKey != nil {
					subObj, ok := val.(map[string]interface{})
					if !ok {
						return nil, nil, fmt.Errorf("key [%s] is not object, cannot access property [%s]", seg.Key, *seg.ObjKey)
					}
					subVal, subExists := subObj[*seg.ObjKey]
					if !subExists {
						if !createMissing {
							return nil, nil, fmt.Errorf("property [%s] under [%s] missing", *seg.ObjKey, seg.Key)
						}
						newObj := make(map[string]interface{})
						subObj[*seg.ObjKey] = newObj
						subVal = newObj
					}
					currentNode = subVal
				} else {
					currentNode = val
				}

			case []interface{}:
				return nil, nil, fmt.Errorf("cannot access key [%s] on raw array element", seg.Key)
			default:
				return nil, nil, fmt.Errorf("path segment [%s] hit non-object/non-array value", seg.Key)
			}
		}
	}

	finalObj, ok := currentNode.(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("final ns path node is not object, cannot get field [%s]", arrName)
	}

	arrVal, exists := finalObj[arrName]
	// fmt.Printf("DEBUG exist=%t, raw value type=%T, val=%#v\n", exists, arrVal, arrVal)
	if !exists {
		if !createMissing {
			return nil, nil, fmt.Errorf("array field [%s] not exists, enable --create-missing to auto create empty array", arrName)
		}
		newEmptyArr := make([]interface{}, 0)
		finalObj[arrName] = newEmptyArr
		return finalObj, newEmptyArr, nil
	}

	targetArr, ok := arrVal.([]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("field [%s] exists but is not json array", arrName)
	}
	// 返回父对象、原始数组
	return finalObj, targetArr, nil
}