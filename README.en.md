# yogo - CLI Tool for Bulk Add/Remove JSON Array Items
Automatically modify string arrays inside JSON configuration files. Optimized for Docker `daemon.json`, project `package.json`, and various service configs. Supports nested JSON paths, auto creation of missing nodes, and dry-run preview mode.

## ✨ Core Features
1. **Basic Array Operations**: `--include` bulk add items, `--exclude` bulk remove items, with automatic deduplication and alphabetical sorting
2. **Nested JSON Paths**: Multi-level paths via `--ns`, supporting plain keys, numeric array indexes `[number]`, and custom object keys `[text]`
3. **Auto Create Missing Nodes**: `--create-missing` generates missing parent objects and empty target arrays
4. **Safe Preview Mode**: `--dryrun` prints modified JSON only without writing to disk, preventing accidental corruption of system configs
5. **Debug Logging**: `--debug` outputs full arguments, file paths, and step-by-step array processing logs
6. **Quick Info Commands**: `-v/--version` print version, `-h/--help` full usage guide, `--print-feature` list all built-in capabilities
7. **Simplified Invocation**: The subcommand `edit-json-arr` can be omitted; the tool falls back to it automatically

## 📌 Full Parameter Reference
### Global Base Flags
| Flag | Short Alias | Description | Default | Example |
|------|-------------|-------------|---------|---------|
| --file | - | Path to target JSON file; workspace is ignored if absolute path is provided | N/A | /etc/docker/daemon.json, daemon.json |
| --workspace | -w | Working directory, only effective for relative file paths | ./ | --workspace ./config |
| --name | - | Name of target array field (required) | N/A | registry-mirrors, insecure-registries |
| --ns | - | Nested JSON hierarchy path | Empty string | docker.config, list[0].env[MIRROR] |
| --ns-sep | - | Separator for nested path segments | . | --ns-sep / |
| --sep | - | Delimiter separating multiple values in include/exclude | , | --sep \| |
| --create-missing | - | Auto create missing parent objects and empty arrays | false | --create-missing |
| --dryrun | - | Preview changes without writing to disk | false | --dryrun |
| --debug | - | Enable full verbose debug logs | false | --debug |
| --default-text | - | Fallback JSON content if target file does not exist | {} | --default-text '{"arr":[]}' |

### Info-only Flags (Highest Priority, Program exits immediately after execution)
| Flag | Short Alias | Description | Default | Example |
|------|-------------|-------------|---------|---------|
| --version | -v | Print short version string only | false | -v |
| --help | -h | Print complete usage documentation | false | -h |
| --print-feature | - | Print full list of all supported built-in features | false | --print-feature |

### Subcommand Compatibility Flag
| Argument | Description | Default | Example |
|----------|-------------|---------|---------|
| edit-json-arr | Explicit subcommand, optional to omit | Auto fallback | yogo edit-json-arr --file test.json |

## 🚀 Most Common Command Examples
### 1. Basic Information Lookup
```bash
# Show version
bin/yogo -v

# Print full help document
bin/yogo -h

# List all supported features
bin/yogo --print-feature
```

### 2. Manipulate Docker daemon.json (Most Typical Use Case)
#### Bulk add registry mirrors with dry-run preview (no write)
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--create-missing \
--include "mirrors.sohu.com,hub-mirror.c.163.com" \
--dryrun --debug
```

#### Delete specific mirror entry and write changes directly
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--exclude "mirrors.sohu.com"
```

#### Simultaneously add new entries and remove multiple old entries
```bash
sudo bin/yogo \
--file /etc/docker/daemon.json \
--name registry-mirrors \
--create-missing \
--include "https://new.mirror.test" \
--exclude "http://mirrors.sohu.com/,https://dockerproxy.com" \
--dryrun
```

#### Reload Docker service to apply config changes
```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 3. Basic Operations on Flat JSON Files
```bash
# Short syntax (recommended, no explicit subcommand)
bin/yogo --file app.json --name urls --create-missing --include "a.com,b.com" --dryrun

# Full explicit subcommand compatible syntax
bin/yogo edit-json-arr --file app.json --name urls --exclude "a.com"
```

### 4. Operate Simple Nested Paths
Original JSON structure:
```json
{
  "docker": {
    "config": {
      "mirrors": []
    }
  }
}
```
Execution command:
```bash
bin/yogo \
--file daemon.json \
--ns "docker.config" \
--name mirrors \
--create-missing \
--include "https://mirror.ustc.edu.cn"
```

### 5. Complex Mixed Nested Paths (Array Index + Custom Object Key)
Original JSON structure:
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
Execution command:
```bash
bin/yogo \
--file app.json \
--ns "config.list[0].env[MIRROR]" \
--name urls \
--create-missing \
--include "https://test-mirror.com"
```

### 6. Custom Separator `|` for Multi-value Lists
```bash
bin/yogo \
--file test.json \
--name arr \
--sep "|" \
--include "test1.com|test2.com" \
--exclude "old1.com|old2.com" \
--dryrun
```

### 7. Auto Initialize When Target File Does Not Exist
```bash
bin/yogo \
--file new.json \
--name list \
--create-missing \
--default-text '{"name":"demo"}' \
--include "item1,item2"
```

## 📁 Supported JSON Array Rules
1. Only **string arrays** are supported; numeric, boolean, and object arrays are not processed
2. Automatic deduplication: duplicate input values will not be stored repeatedly
3. Automatic sorting: final array entries are sorted lexicographically
4. Exact string matching for deletion via `--exclude`: wildcards and regular expressions are not supported at present

## 🐳 Docker Container Deployment
```bash
# Compile static binary
go build -o bin/yogo ./cmd/yogo/main.go

# Mount host Docker config directory and binary into container
docker run -it --rm -v /etc/docker:/etc/docker:ro -v $(pwd)/bin:/app/bin alpine sh

# Modify daemon.json inside container
/app/bin/yogo \
--file /etc/docker/daemon.json \
--name insecure-registries \
--create-missing \
--include "test.mirror.com" \
--dryrun
```

## ⚠️ Important Notes
1. **System Config Permissions**: `/etc/docker/daemon.json` is a protected system file; `sudo` is required for read/write operations
2. **Path Resolution Logic**: The `--workspace` flag is ignored entirely if an absolute path is passed to `--file`
3. **Exact Match Requirement**: Deletion uses full literal string matching; trailing slashes and whitespace will alter matching results
4. **Array Type Restriction**: Only pure string slices `[]string` are handled; nested arrays, number and object arrays are unsupported
5. **File Permission Mask**: Written files default to `0644` permission; reload related services after modifying system configurations
6. **Missing Nested Nodes**: `--create-missing` must be supplied when intermediate path nodes do not exist, otherwise an error will be thrown

## 🛡️ Highlighted Advantages
- ✅ Dual invocation modes: shorthand no-subcommand syntax + full explicit subcommand compatibility for legacy scripts
- ✅ Intelligent path handling: distinguishes absolute and relative file paths to avoid standard library root path truncation bugs
- ✅ Zero-config basic usage: only two mandatory flags (`--file`, `--name`) required for core operations
- ✅ DevOps & Operations friendly: dry-run preview and full debug logs reduce risk of breaking production configurations
- ✅ Full nested path support: unified syntax for plain keys, numeric array indexes, and custom object bracket keys
- ✅ Container optimized: static compiled binary with zero external runtime dependencies, fits lightweight base images
- ✅ Automation script ready: fully non-interactive input, ideal for CI/CD pipelines and batch operation scripts