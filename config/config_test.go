package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ==================== splitConfigFile 测试 ====================

func TestSplitConfigFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDir  string
		wantFile string
		wantExt  string
	}{
		{
			name:     "标准路径_yaml",
			input:    "config/app.yaml",
			wantDir:  "config",
			wantFile: "app",
			wantExt:  "yaml",
		},
		{
			name:     "标准路径_json",
			input:    "config/app.json",
			wantDir:  "config",
			wantFile: "app",
			wantExt:  "json",
		},
		{
			name:     "无目录",
			input:    "app.yaml",
			wantDir:  ".",
			wantFile: "app",
			wantExt:  "yaml",
		},
		{
			name:     "无扩展名",
			input:    "config/app",
			wantDir:  "config",
			wantFile: "app",
			wantExt:  "",
		},
		{
			name:     "多层目录",
			input:    filepath.Join("a", "b", "c", "config.yaml"),
			wantDir:  filepath.Join("a", "b", "c"),
			wantFile: "config",
			wantExt:  "yaml",
		},
		{
			name:     "文件名含多个点",
			input:    "config/app.v2.yaml",
			wantDir:  "config",
			wantFile: "app.v2",
			wantExt:  "yaml",
		},
		{
			name:     "隐藏文件_unix风格",
			input:    ".env",
			wantDir:  ".",
			wantFile: "",
			wantExt:  "env",
		},
		{
			name:     "隐藏文件_yaml风格",
			input:    "config/.app.yaml",
			wantDir:  "config",
			wantFile: ".app",
			wantExt:  "yaml",
		},
		{
			name:     "空字符串",
			input:    "",
			wantDir:  ".",
			wantFile: "",
			wantExt:  "",
		},
		{
			name:     "仅文件名",
			input:    "config.yaml",
			wantDir:  ".",
			wantFile: "config",
			wantExt:  "yaml",
		},
		{
			name:     "toml格式",
			input:    "settings.toml",
			wantDir:  ".",
			wantFile: "settings",
			wantExt:  "toml",
		},
		// ========== 边界条件 ==========
		{
			name:     "纯扩展名无文件名",
			input:    ".yaml",
			wantDir:  ".",
			wantFile: "",
			wantExt:  "yaml",
		},
		{
			name:     "unix绝对路径",
			input:    "/etc/app/config.yaml",
			wantDir:  filepath.FromSlash("/etc/app"),
			wantFile: "config",
			wantExt:  "yaml",
		},
		{
			name:     "当前目录相对路径前缀",
			input:    "./config/app.yaml",
			wantDir:  "config",
			wantFile: "app",
			wantExt:  "yaml",
		},
		{
			name:     "路径末尾带分隔符",
			input:    "config/",
			wantDir:  "config",
			wantFile: "config",
			wantExt:  "",
		},
		{
			name:     "含空格路径",
			input:    "my config/app config.yaml",
			wantDir:  "my config",
			wantFile: "app config",
			wantExt:  "yaml",
		},
		{
			name:     "双扩展名",
			input:    "archive/app.tar.gz",
			wantDir:  "archive",
			wantFile: "app.tar",
			wantExt:  "gz",
		},
		{
			name:     "当前目录单文件",
			input:    "./config.yaml",
			wantDir:  ".",
			wantFile: "config",
			wantExt:  "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotFile, gotExt := splitConfigFile(tt.input)
			if gotDir != tt.wantDir {
				t.Errorf("dir: got %q, want %q", gotDir, tt.wantDir)
			}
			if gotFile != tt.wantFile {
				t.Errorf("fileName: got %q, want %q", gotFile, tt.wantFile)
			}
			if gotExt != tt.wantExt {
				t.Errorf("extName: got %q, want %q", gotExt, tt.wantExt)
			}
		})
	}
}

// ==================== initViper / Init 测试 ====================

func TestInitViper_YAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `
server:
  port: 8080
  host: "localhost"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type ServerConf struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}
	type AppConfig struct {
		Server ServerConf `mapstructure:"server"`
	}

	var cfg AppConfig
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("port: got %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("host: got %q, want %q", cfg.Server.Host, "localhost")
	}
}

func TestInitViper_JSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.json")
	content := `{
  "server": {
    "port": 9090,
    "host": "127.0.0.1"
  }
}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type ServerConf struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}
	type AppConfig struct {
		Server ServerConf `mapstructure:"server"`
	}

	var cfg AppConfig
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper failed: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("port: got %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("host: got %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
}

func TestInitViper_FileNotFound(t *testing.T) {
	err := initViper("nonexistent/config.yaml", &struct{}{})
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestInitViper_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad.yaml")
	content := `invalid: yaml: [malformed`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := initViper(configPath, &struct{}{})
	if err == nil {
		t.Error("expected error for invalid yaml, got nil")
	}
}

func TestInitViper_NonPointerData(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `key: value`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		Key string `mapstructure:"key"`
	}
	// 传值而非指针
	err := initViper(configPath, cfg)
	if err == nil {
		t.Error("expected error for non-pointer data, got nil")
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `
database:
  dsn: "root:123456@tcp(127.0.0.1:3306)/test"
  max_conns: 100
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type DBConf struct {
		DSN      string `mapstructure:"dsn"`
		MaxConns int    `mapstructure:"max_conns"`
	}
	type AppConfig struct {
		Database DBConf `mapstructure:"database"`
	}

	var cfg AppConfig
	if err := Init(configPath, &cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if cfg.Database.DSN != "root:123456@tcp(127.0.0.1:3306)/test" {
		t.Errorf("dsn: got %q", cfg.Database.DSN)
	}
	if cfg.Database.MaxConns != 100 {
		t.Errorf("max_conns: got %d, want 100", cfg.Database.MaxConns)
	}
}

// ==================== 边界条件测试 ====================

func TestInitViper_TOML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	content := `
[server]
port = 3000
host = "0.0.0.0"
debug = true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type ServerConf struct {
		Port  int    `mapstructure:"port"`
		Host  string `mapstructure:"host"`
		Debug bool   `mapstructure:"debug"`
	}
	type AppConfig struct {
		Server ServerConf `mapstructure:"server"`
	}

	var cfg AppConfig
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper TOML failed: %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("port: got %d, want 3000", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("host: got %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if !cfg.Server.Debug {
		t.Error("debug: got false, want true")
	}
}

func TestInitViper_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(configPath, []byte(``), 0644); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		Key string `mapstructure:"key"`
	}
	// 空配置文件应可正常加载，字段保持零值
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper empty config failed: %v", err)
	}
	if cfg.Key != "" {
		t.Errorf("key: got %q, want empty string", cfg.Key)
	}
}

func TestInitViper_MultipleTypes(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `
app:
  name: "gonet"
  version: 1.5
  enabled: true
  tags:
    - "go"
    - "network"
  rate: 0.75
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type AppConf struct {
		Name    string   `mapstructure:"name"`
		Version float64  `mapstructure:"version"`
		Enabled bool     `mapstructure:"enabled"`
		Tags    []string `mapstructure:"tags"`
		Rate    float64  `mapstructure:"rate"`
	}
	type Config struct {
		App AppConf `mapstructure:"app"`
	}

	var cfg Config
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper multiple types failed: %v", err)
	}

	if cfg.App.Name != "gonet" {
		t.Errorf("name: got %q, want %q", cfg.App.Name, "gonet")
	}
	if cfg.App.Version != 1.5 {
		t.Errorf("version: got %v, want 1.5", cfg.App.Version)
	}
	if !cfg.App.Enabled {
		t.Error("enabled: got false, want true")
	}
	if len(cfg.App.Tags) != 2 || cfg.App.Tags[0] != "go" || cfg.App.Tags[1] != "network" {
		t.Errorf("tags: got %v, want [go network]", cfg.App.Tags)
	}
	if cfg.App.Rate != 0.75 {
		t.Errorf("rate: got %v, want 0.75", cfg.App.Rate)
	}
}

func TestInitViper_ExtraKeysIgnored(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `
name: "gonet"
version: "1.0"
extra_field: "should be ignored"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	}
	// 配置文件中多余的字段不应导致报错
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper with extra keys failed: %v", err)
	}
	if cfg.Name != "gonet" {
		t.Errorf("name: got %q, want %q", cfg.Name, "gonet")
	}
	if cfg.Version != "1.0" {
		t.Errorf("version: got %q, want %q", cfg.Version, "1.0")
	}
}

func TestInitViper_NilPointer(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `key: value`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// 传入 nil 指针应报错
	err := initViper(configPath, nil)
	if err == nil {
		t.Error("expected error for nil pointer, got nil")
	}
}

func TestInitViper_DeeplyNestedStruct(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := `
http:
  server:
    port: 8080
    read_timeout: 30
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	type ServerConf struct {
		Port        int `mapstructure:"port"`
		ReadTimeout int `mapstructure:"read_timeout"`
	}
	type HTTPConf struct {
		Server ServerConf `mapstructure:"server"`
	}
	type Config struct {
		HTTP HTTPConf `mapstructure:"http"`
	}

	var cfg Config
	if err := initViper(configPath, &cfg); err != nil {
		t.Fatalf("initViper deeply nested failed: %v", err)
	}
	if cfg.HTTP.Server.Port != 8080 {
		t.Errorf("port: got %d, want 8080", cfg.HTTP.Server.Port)
	}
	if cfg.HTTP.Server.ReadTimeout != 30 {
		t.Errorf("read_timeout: got %d, want 30", cfg.HTTP.Server.ReadTimeout)
	}
}
