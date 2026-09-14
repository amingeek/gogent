# راهنمای کامل سیستم پلاگین‌ها در gogent

در این مستند یاد می‌گیری که چطور پلاگین‌های جدید بنویسی. به‌عنوان مثال، قدم‌به‌قدم یک پلاگین **write** (نوشتن فایل) می‌سازیم که دقیقاً در مقابل پلاگین read قرار می‌گیرد.

---

## ۱) معماری سیستم پلاگین

کل سیستم پلاگین‌ها داخل پوشه `plugins/` و سه پارت اصلی زیر قرار دارد:

```
plugins/
├── plugin.go          ← دو اینترفیس اصلی (Plugin و Tool) و مدل‌ها
├── registry.go        ← حافظه‌ی ثبت پلاگین/تول و اجرای تول‌ها
├── helpers.go         ← توابع کمکی (خواندن فایل، بررسی وجود فایل...)
│
├── builtin/           ← یک پلاگین آماده (محاسبه‌گر + ساعت)
│   └── builtin.go
│
└── read/              ← پلاگین خواندن فایل (برای الگوبرداری)
    └── read_plugin.go
```

### جریان کار

1. پلاگین در متد `init()` خودش را ثبت می‌کند: `plugins.RegisterPlugin(...)`
2. در `main.go` فقط با یک import خالی پلاگین را وارد می‌کنیم: `_ "gogent/plugins/xxx"`
3. با اجرای برنامه، `init()` پلاگین را اجرا و `RegisterPlugin` آن را در `Registry` ثبت می‌کند.
4. تول‌های ثبت‌شده به مدل AI فرستاده می‌شوند تا AI تصمیم بگیرد کدام را صدا بزند.
5. وقتی AI یک تول را صدا می‌زند، `registry.ExecuteTool(...)` متد `Execute` آن تول را اجرا می‌کند.

---

## ۲) دو اینترفیس اصلی

همه‌چیز روی دو اینترفیس ساده بنا شده است (فایل `plugins/plugin.go`):

### اینترفیس `Plugin` — خود پلاگین

```go
type Plugin interface {
	Name() string                 // نام یکتا، مثلاً "write"
	Version() string              // نسخه، مثلاً "1.0.0"
	Description() string          // توضیح که پلاگین چه کار می‌کند
	GetTools() []Tool             // لیست تول‌هایی که این پلاگین ارائه می‌دهد
	Initialize(config map[string]interface{}) error // اگر تنظیمات لازم دارید
}
```

### اینترفیس `Tool` — یک ابزار/قابلیت

```go
type Tool interface {
	Name() string                              // نام یکتای تول، مثلاً "write_file"
	Description() string                       // چه زمانی AI باید این را صدا بزند
	Parameters() interface{}                   // تعریف ورودی‌ها به شکل JSON Schema
	Execute(args map[string]interface{}) (string, error) // بدنه‌ی اجرای تول
}
```

> نکته مهم: `Parameters()` **فرمت JSON Schema استاندارد** است که به مدل AI داده می‌شود تا بفهمد هر تول چه ورودی‌ای می‌گیرد. با حذف این بخش، AI نمی‌داند آرگومان درست چطور است.

---

## ۳) ساختار پوشه برای یک پلاگین جدید

برای هر پلاگین یک **پوشه‌ی جدا** زیر `plugins/` می‌سازیم:

```
plugins/write/              ← نام پوشه = نام پکیج Go
└── write_plugin.go         ← کد پلاگین
```

اسم پکیج باید با اسم پوشه یکی باشد (قانون Go). داخل پوشه فقط یک فایل هست ولی هرچه نیاز باشد (چند فایل `_test.go` و...) می‌توانی اضافه کنی.

---

## ۴) آموزش قدم‌به‌قدم: پلاگین write

می‌خواهیم پلاگینی بسازیم که این سه تول را داشته باشد:

| تول | کار |
|---|---|
| `write_file` | نوشتن کامل یک فایل (یا بازنویسی) |
| `append_file` | افزودن متن به انتهای یک فایل |
| `write_multiple_files` | نوشتن چند فایل در یک فراخوانی |

### قدم ۱ — ساخت پوشه

```bash
mkdir plugins/write
```

### قدم ۲ — نوشتن فایل پلاگین

فایل `plugins/write/write_plugin.go` را بساز. ابتدا پکیج و importها:

```go
package write

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogent/plugins"
)
```

### قدم ۳ — تعریف خودِ پلاگین

```go
type WritePlugin struct {
	initialized bool
	config      map[string]interface{}
}

func NewWritePlugin() *WritePlugin {
	return &WritePlugin{}
}

func (p *WritePlugin) Name() string        { return "write" }
func (p *WritePlugin) Version() string     { return "1.0.0" }
func (p *WritePlugin) Description() string { return "Create and modify files by writing or appending content" }

func (p *WritePlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.initialized = true
	return nil
}

// هر تولی که این پلاگین ارائه می‌دهد اینجا لیست می‌شود
func (p *WritePlugin) GetTools() []plugins.Tool {
	return []plugins.Tool{
		&WriteFileTool{},
		&AppendFileTool{},
		&WriteMultipleFilesTool{},
	}
}
```

### قدم ۴ — تول اول: `write_file`

```go
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write or overwrite the entire content of a file. Use when you need to create a new file or replace a file's content."
}

// تعریف ورودی‌ها با فرمت JSON Schema
func (t *WriteFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write (relative to workspace root)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The full content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

// بدنه‌ی اجرا: اینجا کار واقعی انجام می‌شود
func (t *WriteFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content must be a string")
	}

	// اطمینان از وجود پوشه‌ی والد
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("cannot create directory %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("cannot write file %s: %v", path, err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len([]byte(content)), path), nil
}
```

### قدم ۵ — تول دوم: `append_file`

```go
type AppendFileTool struct{}

func (t *AppendFileTool) Name() string { return "append_file" }
func (t *AppendFileTool) Description() string {
	return "Append text to the end of an existing or new file. Use when you need to add content without overwriting."
}
func (t *AppendFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to append to",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to append",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *AppendFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content must be a string")
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %v", path, err)
	}
	defer file.Close()

	n, err := file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("cannot append to file %s: %v", path, err)
	}

	return fmt.Sprintf("Successfully appended %d bytes to %s", n, path), nil
}
```

### قدم ۶ — تول سوم: `write_multiple_files`

```go
type WriteMultipleFilesTool struct{}

func (t *WriteMultipleFilesTool) Name() string { return "write_multiple_files" }
func (t *WriteMultipleFilesTool) Description() string {
	return "Write several files at once. Use when creating or updating multiple files in one step."
}
func (t *WriteMultipleFilesTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"files": map[string]interface{}{
				"type":        "array",
				"description": "List of files to write",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string", "description": "File path"},
						"content": map[string]interface{}{"type": "string", "description": "File content"},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		"required": []string{"files"},
	}
}

func (t *WriteMultipleFilesTool) Execute(args map[string]interface{}) (string, error) {
	filesInterface, ok := args["files"].([]interface{})
	if !ok {
		return "", fmt.Errorf("files must be an array")
	}

	var results []string
	for i, fileItem := range filesInterface {
		fileMap, ok := fileItem.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("files[%d] must be an object", i)
		}
		path, _ := fileMap["path"].(string)
		content, _ := fileMap["content"].(string)
		if path == "" {
			return "", fmt.Errorf("files[%d].path is required", i)
		}

		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				results = append(results, fmt.Sprintf("%s: ERROR %v", path, err))
				continue
			}
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			results = append(results, fmt.Sprintf("%s: ERROR %v", path, err))
			continue
		}
		results = append(results, fmt.Sprintf("Wrote %s (%d bytes)", path, len([]byte(content))))
	}

	return strings.Join(results, "\n"), nil
}
```

### قدم ۷ — ثبت خودکار پلاگین در `init`

پایین فایل، با `init()` پلاگین را در registry ثبت می‌کنیم:

```go
// وقتی پکیج import می‌شود، init خودکار اجرا و پلاگین ثبت می‌گردد
func init() {
	if err := plugins.RegisterPlugin(NewWritePlugin()); err != nil {
		fmt.Println("ERROR: failed to register write plugin:", err)
	}
}

// این خط فقط برای اطمینان نوع (compile-time check) است
var _ plugins.Plugin = (*WritePlugin)(nil)
```

### قدم ۸ — معرفی پلاگین به برنامه در `main.go`

فایل `main.go` را باز کن و import خالی پلاگین write را اضافه کن:

```go
package main

import (
	"gogent/core"
	_ "gogent/plugins/builtin"
	_ "gogent/plugins/read"
	_ "gogent/plugins/write"   // ← این خط را اضافه کن
)

func main() {
	core.AgentLoop()
}
```

### قدم ۹ — بیلد و تست

```bash
go build ./...
```

سپس برنامه را اجرا کن:

```bash
go run .
```

باید در آغاز برنامه نوشته شود:

```
Loaded 3 plugins:
  [builtin v1.0.0] ... (tools: calculator, get_current_time)
  [read v1.0.0] ... (tools: read_file, read_lines, ...)
  [write v1.0.0] ... (tools: write_file, append_file, ...)
```

و حالا می‌توانی از AI بخواهی: *"یک فایل hello.txt با محتوای hello world بساز"* و تول `write_file` فراخوانی می‌شود.

---

## ۵) خلاصه‌ی چک‌لیست ساخت هر پلاگین جدید

برای ساخت هر پلاگین جدید فقط این ۴ قدم را تکرار کن:

1. پوشه بساز: `mkdir plugins/نام-پلاگین`
2. فایل پلاگین را با اینترفیس `Plugin` بنویس (Name، Version، Description، GetTools، Initialize)
3. برای هر تول یک struct با اینترفیس `Tool` بنویس (Name، Description، Parameters، Execute)
4. در `init()` خود را ثبت کن و در `main.go` پلاگین را import خالی کن

---

## ۶) نکات و بهترین روش‌ها

### تبدیل نوع مقادیر از JSON

وقتی AI تول را صدا می‌زند، آرگومان‌ها از JSON می‌آیند و **اعداد همیشه float64** هستند. همیشه از یک helper مثل `toInt` استفاده کن (که در پلاگین read هست):

```go
func toInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	default:
		return 0, fmt.Errorf("value must be an integer, got %T", v)
	}
}
```

### توضیح خوب برای AI بنویس

متن `Description()` هر تول را طوری بنویس که AI بداند **کِی** باید صدا بزند. مثال:

- بد: `"Writes files"`
- خوب: `"Write or overwrite the entire content of a file. Use when you need to create a new file or replace a file's content."`

### نام یکتا باشد

اگر دو تول نام یکسان داشته باشند، ثبت دوم با خطا رد می‌شود. برای هر پلاگین از پیشوند استفاده کن (مثل `read_file`، `write_file`).

### پیشوند نام‌های پلاگین

- پلاگین `read` → تول‌ها: `read_file`، `read_lines`، `read_multiple_files`، `search_file`
- پلاگین `builtin` → تول‌ها: `calculator`، `get_current_time`
- پلاگین `write` → تول‌ها: `write_file`، `append_file`، `write_multiple_files`

### استفاده از helpers آماده

توابعی که در `plugins/helpers.go` هست را می‌توانی از داخل پلاگین‌ت صدا بزنی:

- `plugins.ReadFileContent(path)` — خواندن کل فایل
- `plugins.ReadFileLines(path, start, end)` — خواندن بازه‌ای از خطوط
- `plugins.SearchFileContent(path, pattern, ctx, caseSensitive)` — جستجو در فایل
- `plugins.FileExists(path)` — بررسی وجود فایل

### اگر پلاگین به فایل سیستم دسترسی محدود می‌خواهد

می‌توانی داخل `Initialize` پوشه‌ی ریشه یا مسیر مجاز را از `config` بخوانی و در تول‌ها، مسیر ورودی را با `filepath.Join` به آن محدود کنی.

---

## ۷) مرجع کامل توابع Registry

متدهای مفید روی `plugins.GetGlobalRegistry()`:

| متد | توضیح |
|---|---|
| `Register(plugin)` | ثبت پلاگین و تول‌هایش |
| `Unregister(name)` | حذف پلاگین |
| `GetTool(name)` | گرفتن یک تول با نام |
| `GetTools()` | لیست همه تول‌ها |
| `GetToolsForPrompt()` | تول‌ها در قالب JSON برای مدل |
| `GetPlugins()` | لیست پلاگین‌های ثبت‌شده |
| `GetPluginInfos()` | اطلاعات پلاگین‌ها (برای نمایش) |
| `DetectTools(prompt)` | تول‌های مرتبط با یک prompt |
| `ExecuteTool(name, argsJSON)` | اجرای یک تول با آرگومان JSON |