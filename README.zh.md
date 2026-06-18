# yogo - JSON数组批量增删命令行工具
自动化修改 JSON 配置内字符串数组，适配 Docker daemon.json、项目 package.json、各类服务配置文件，支持嵌套路径、自动创建节点、预览干跑。

## ✨ 核心功能
1. **基础数组操作**：`--include` 批量新增元素、`--exclude` 批量删除元素，自动去重升序排序
2. **嵌套JSON路径**：`--ns` 多层级路径，支持普通键、数组下标 `[number]`、对象自定义键 `[text]`
3. **缺失自动创建**：`--create-missing` 自动生成不存在的父对象与空数组
4. **安全预览**：`--dryrun` 仅输出变更JSON，不写入磁盘，避免误改系统配置
5. **调试日志**：`--debug` 打印完整入参、文件路径、数组全流程处理日志
6. **快捷信息指令**：`-v/--version` 打印版本、`-h/--help` 完整帮助、`--print-feature` 输出全部功能清单
7. **简化调用**：可省略子命令 `edit-json-arr`，程序自动兜底执行

## 📌 完整参数说明
### 全局基础参数
| 参数 | 短别名 | 说明 | 默认值 | 示例 |
|------|--------|------|--------|------|
| --file | - | 目标 JSON 文件路径，绝对路径自动忽略 workspace | - | /etc/docker/daemon.json、daemon.json |
| --workspace | -w | 工作目录，仅对相对文件路径生效 | ./ | --workspace ./config |
| --name | - | 目标数组字段名称（必填） | - | registry-mirrors、insecure-registries |
| --ns | - | 嵌套JSON层级路径 | 空 | docker.config、list[0].env[MIRROR] |
| --ns-sep | - | 嵌套路径层级分隔符 | . | --ns-sep / |
| --sep | - | include/exclude 多值分割符 | , | --sep \| |
| --create-missing | - | 自动创建缺失父对象、空数组 | false | --create-missing |
| --dryrun | - | 干跑预览，不写入磁盘 | false | --dryrun |
| --debug | - | 开启完整调试日志 | false | --debug |
| --default-text | - | 文件不存在时初始化JSON内容 | {} | --default-text '{"arr":[]}' |

### 信息查看专用参数（优先级最高，执行后直接退出）
| 参数 | 短别名 | 说明 | 默认值 | 示例 |
|------|--------|------|--------|------|
| --version | -v | 仅打印简洁版本号 | false | -v |
| --help | -h | 打印完整使用帮助文档 | false | -h |
| --print-feature | - | 打印全部内置功能特性清单 | false | --print-feature |

### 子命令兼容参数
| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| edit-json-arr | 显式子命令，可省略 | 自动兜底 | yogo edit-json-arr --file test.json |

## 🚀 最常用命令示例
### 1. 基础信息查看
```bash
# 查看版本
bin/yogo -v

# 完整帮助文档
bin/yogo -h

# 打印全部支持特性
bin/yogo --print-feature
```

### 2. Docker daemon.json 操作（最典型场景）
#### 批量新增镜像源，预览不写入
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--create-missing \
--include "mirrors.sohu.com,hub-mirror.c.163.com" \
--dryrun --debug
```

#### 删除指定镜像源，直接写入配置
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--exclude "mirrors.sohu.com"
```

#### 同时新增+批量删除多条配置
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name registry-mirrors \
--create-missing \
--include "https://new.mirror.test" \
--exclude "http://mirrors.sohu.com/,https://dockerproxy.com" \
--dryrun
```

#### 修改完成重载Docker生效
```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 3. 单层JSON文件基础操作
```bash
# 简写模式（推荐，省略子命令）
bin/yogo --file app.json --name urls --create-missing --include "a.com,b.com" --dryrun

# 完整显式子命令兼容写法
bin/yogo edit-json-arr --file app.json --name urls --exclude "a.com"
```

### 4. 普通嵌套路径操作
原始JSON结构：
```json
{
  "docker": {
    "config": {
      "mirrors": []
    }
  }
}
```
执行命令：
```bash
bin/yogo \
--file daemon.json \
--ns "docker.config" \
--name mirrors \
--create-missing \
--include "https://mirror.ustc.edu.cn"
```

### 5. 混合复杂嵌套路径（数组下标+对象自定义键）
原始JSON结构：
```json
{
  "config": {
    "list": [
      {
        "env": {
          "MIRROR": {
            "urls": []
          }
        }
      }
    ]
  }
}
```
执行命令：
```bash
bin/yogo \
--file app.json \
--ns "config.list[0].env[MIRROR]" \
--name urls \
--create-missing \
--include "https://test-mirror.com"
```

### 6. 自定义分割符 | 处理多值
```bash
bin/yogo \
--file test.json \
--name arr \
--sep "|" \
--include "test1.com|test2.com" \
--exclude "old1.com|old2.com" \
--dryrun
```

### 7. 文件不存在自动初始化
```bash
bin/yogo \
--file new.json \
--name list \
--create-missing \
--default-text '{"name":"demo"}' \
--include "item1,item2"
```

## 📁 支持的JSON数组规则
1. 仅处理**字符串类型数组**，数字、布尔、对象数组暂不支持
2. 自动去重：重复传入相同元素不会重复存入
3. 自动排序：处理完成后数组按字符串升序排列
4. 精确匹配：`--exclude` 使用完整字符串精确匹配删除，暂不支持通配符/正则

## 🐳 容器化部署（Docker）
```bash
# 编译二进制
go build -o bin/yogo ./cmd/yogo/main.go

# 挂载宿主机docker配置文件到容器内使用
docker run -it --rm -v /etc/docker:/etc/docker:ro -v $(pwd)/bin:/app/bin alpine sh

# 容器内操作daemon.json
/app/bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--create-missing \
--include "test.mirror.com" \
--dryrun
```

## ⚠️ 注意事项
1. **系统配置权限**：`/etc/docker/daemon.json` 属于系统保护文件，读写必须加 `sudo`
2. **路径规则**：`--file` 传入绝对路径时，`--workspace` 参数会直接失效
3. **元素匹配**：删除为完整字符串精确匹配，地址首尾斜杠、空格会影响匹配结果
4. **数组限制**：仅支持 `[]string` 格式数组，对象、数字、嵌套数组无法处理
5. **文件权限**：写入文件默认权限 `0644`，系统配置修改后建议重载对应服务
6. **嵌套层级**：多层路径缺失节点必须搭配 `--create-missing`，否则直接抛出错误

## 🛡️ 特性亮点
- ✅ 双调用模式：简写无命令 + 完整子命令兼容，适配新旧脚本
- ✅ 智能路径处理：自动区分绝对/相对文件路径，规避标准库路径丢失问题
- ✅ 零配置开箱即用：仅需 `--file`、`--name` 两个必填参数即可完成基础操作
- ✅ 运维友好：dryrun预览、debug全链路日志，降低线上配置误改风险
- ✅ 嵌套路径全兼容：普通键、数组下标、对象自定义键三种路径语法全覆盖
- ✅ 容器友好：静态编译二进制，无额外运行依赖，可直接放入轻量镜像
- ✅ 自动化脚本适配：无交互式输入，适合CI/CD、运维批量脚本调用
