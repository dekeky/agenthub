package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/agenthub/internal/archive"
	"github.com/agenthub/internal/client"
	"github.com/agenthub/internal/config"
	"github.com/agenthub/internal/hub"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	cmd := firstCommand(args)
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		printUsage(os.Stdout)
		return 0
	}

	configPath := flagArg(args, "--config")
	cliCfg, err := config.LoadCLI(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	global := flag.NewFlagSet("global", flag.ExitOnError)
	global.String("config", "", "path to agenthub-cli.toml")
	baseURL := global.String("url", cliCfg.URL, "override hub URL from config")
	token := global.String("token", cliCfg.UploadToken, "override upload token from config")
	asJSON := global.Bool("json", false, "output JSON")
	global.SetOutput(os.Stderr)

	if err := global.Parse(args); err != nil {
		return 2
	}
	cmdArgs := global.Args()
	if len(cmdArgs) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	c := client.New(client.Config{
		BaseURL: *baseURL,
		Token:   *token,
	})
	out := &output{json: *asJSON, w: os.Stdout}

	switch cmdArgs[0] {
	case "list", "ls":
		return runList(c, out, cmdArgs[1:])
	case "categories", "cats":
		return runCategories(c, out, cmdArgs[1:])
	case "get":
		return runGet(c, out, cmdArgs[1:])
	case "file":
		return runFile(c, out, cmdArgs[1:])
	case "download", "dl":
		return runDownload(c, out, cmdArgs[1:])
	case "install":
		return runInstall(c, out, cmdArgs[1:])
	case "upload":
		return runUpload(c, out, cmdArgs[1:])
	case "update":
		return runUpdate(c, out, cmdArgs[1:])
	case "put-file", "write":
		return runPutFile(c, out, cmdArgs[1:])
	case "delete", "rm", "del":
		return runDelete(c, out, cmdArgs[1:])
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmdArgs[0])
		printUsage(os.Stderr)
		return 2
	}
}

func runList(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	category := fs.String("category", "", "filter by category (e.g. picoclaw, openclaw)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli list [--category CATEGORY]")
		return 2
	}

	agents, err := c.ListAgents(*category)
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"agents": agents, "total": len(agents)})
		return 0
	}
	if len(agents) == 0 {
		fmt.Fprintln(out.w, "No agents found.")
		return 0
	}
	for _, a := range agents {
		fmt.Fprintf(out.w, "%s\t%s\t%s\tlatest=%s\tupdated=%s\n",
			a.AgentName, a.Category, a.DisplayName, a.LatestVersion,
			a.UpdatedAt.Format("2006-01-02"))
	}
	return 0
}

func runCategories(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("categories", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli categories")
		return 2
	}

	categories, err := c.ListCategories()
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"categories": categories})
		return 0
	}
	for _, cat := range categories {
		fmt.Fprintln(out.w, cat)
	}
	return 0
}

func runGet(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	version := fs.String("version", "", "package version (default: latest)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli get <agentName> [--version VERSION]")
		return 2
	}

	detail, err := c.GetAgent(fs.Arg(0), *version)
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(detail)
		return 0
	}
	fmt.Fprintf(out.w, "agentName: %s\n", detail.AgentName)
	fmt.Fprintf(out.w, "category: %s\n", detail.Category)
	fmt.Fprintf(out.w, "displayName: %s\n", detail.DisplayName)
	fmt.Fprintf(out.w, "summary: %s\n", detail.Summary)
	fmt.Fprintf(out.w, "latestVersion: %s\n", detail.LatestVersion)
	fmt.Fprintf(out.w, "updatedAt: %s\n", detail.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(out.w, "versions: %s\n", strings.Join(detail.Versions, ", "))
	fmt.Fprintf(out.w, "files (%d):\n", len(detail.Files))
	for _, f := range detail.Files {
		if f.Dir {
			fmt.Fprintf(out.w, "  %s/\t(dir)\n", f.Path)
			continue
		}
		fmt.Fprintf(out.w, "  %s\t%d\n", f.Path, f.Size)
	}
	return 0
}

func runFile(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("file", flag.ContinueOnError)
	version := fs.String("version", "", "package version (default: latest)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli file <agentName> <path> [--version VERSION]")
		return 2
	}

	content, err := c.GetPackageFile(fs.Arg(0), *version, fs.Arg(1))
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{
			"agentName":    fs.Arg(0),
			"version": *version,
			"path":    fs.Arg(1),
			"content": string(content),
		})
		return 0
	}
	_, _ = out.w.Write(content)
	return 0
}

func runDownload(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("download", flag.ContinueOnError)
	version := fs.String("version", "", "package version (default: latest)")
	dest := fs.String("o", "", "output file path (default: <agentName>.zip)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli download <agentName> [-o FILE] [--version VERSION]")
		return 2
	}

	agentName := fs.Arg(0)
	target := *dest
	if target == "" {
		target = agentName + ".zip"
	}
	if err := c.DownloadAgent(agentName, *version, target); err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"agentName": agentName, "version": *version, "path": target})
		return 0
	}
	fmt.Fprintf(out.w, "downloaded %s\n", target)
	return 0
}

func runInstall(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	expectCategory := fs.String("expect-category", "", "require agent category (e.g. picoclaw, openclaw)")
	version := fs.String("version", "", "package version (default: latest)")
	dest := fs.String("dest", "", "install directory (default: ./agents/<agentName>)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli install [--expect-category CAT] [--dest DIR] [--version VERSION] <agentName>")
		return 2
	}

	agentName := fs.Arg(0)

	detail, err := c.GetAgent(agentName, "")
	if err != nil {
		out.err(err)
		return 1
	}
	if strings.TrimSpace(*expectCategory) != "" {
		want, err := hub.NormalizeCategory(*expectCategory)
		if err != nil {
			out.err(err)
			return 1
		}
		if detail.Category != want {
			out.err(fmt.Errorf("category mismatch: agent %q is %q, expected %q", agentName, detail.Category, want))
			return 1
		}
	}

	target := strings.TrimSpace(*dest)
	if target == "" {
		target = filepath.Join("agents", agentName)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		out.err(err)
		return 1
	}

	tmp, err := os.CreateTemp("", "agenthub-install-*.zip")
	if err != nil {
		out.err(err)
		return 1
	}
	tmpZip := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpZip)

	if err := c.DownloadAgent(agentName, *version, tmpZip); err != nil {
		out.err(err)
		return 1
	}
	if err := os.RemoveAll(target); err != nil {
		out.err(err)
		return 1
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		out.err(err)
		return 1
	}
	if err := archive.ExtractZipFile(tmpZip, target); err != nil {
		_ = os.RemoveAll(target)
		out.err(err)
		return 1
	}

	if out.json {
		out.writeJSON(map[string]any{
			"agentName": agentName,
			"category":  detail.Category,
			"version":   *version,
			"dest":      target,
			"installed": true,
		})
		return 0
	}
	fmt.Fprintf(out.w, "installed %s (%s) -> %s\n", agentName, detail.Category, target)
	return 0
}

func runUpload(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	category := fs.String("category", "picoclaw", "agent category (e.g. picoclaw, openclaw)")
	version := fs.String("version", "", "package version")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli upload [--category CATEGORY] [--version VERSION] <agentName> <zip-file>")
		return 2
	}

	meta, err := c.UploadAgent(fs.Arg(0), *category, *version, fs.Arg(1))
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"agent": meta})
		return 0
	}
	fmt.Fprintf(out.w, "uploaded %s@%s\n", meta.AgentName, meta.LatestVersion)
	return 0
}

func runUpdate(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	displayName := fs.String("display-name", "", "new display name")
	summary := fs.String("summary", "", "new summary")
	category := fs.String("category", "", "new category (e.g. picoclaw, openclaw)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli update [--display-name NAME] [--summary TEXT] [--category CAT] <agentName>")
		return 2
	}
	if strings.TrimSpace(*displayName) == "" && strings.TrimSpace(*summary) == "" && strings.TrimSpace(*category) == "" {
		fmt.Fprintln(os.Stderr, "error: at least one of --display-name, --summary, --category is required")
		return 2
	}

	meta, err := c.UpdateAgentMeta(fs.Arg(0), client.UpdateMetaRequest{
		DisplayName: strings.TrimSpace(*displayName),
		Summary:     strings.TrimSpace(*summary),
		Category:    strings.TrimSpace(*category),
	})
	if err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"agent": meta})
		return 0
	}
	fmt.Fprintf(out.w, "updated %s (%s)\n", meta.AgentName, meta.Category)
	return 0
}

func runPutFile(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("put-file", flag.ContinueOnError)
	version := fs.String("version", "", "package version (required)")
	fromFile := fs.String("file", "", "read content from local file (default: stdin)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli put-file <agentName> <path> --version VERSION [--file LOCAL]")
		return 2
	}
	if strings.TrimSpace(*version) == "" {
		fmt.Fprintln(os.Stderr, "error: --version is required")
		return 2
	}

	var content []byte
	var err error
	if strings.TrimSpace(*fromFile) != "" {
		content, err = os.ReadFile(*fromFile)
	} else {
		content, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		out.err(err)
		return 1
	}

	agentName := fs.Arg(0)
	filePath := fs.Arg(1)
	if err := c.UpdateFile(agentName, *version, filePath, content); err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{
			"agentName": agentName,
			"version":   *version,
			"path":      filePath,
			"updated":   true,
		})
		return 0
	}
	fmt.Fprintf(out.w, "updated %s@%s:%s\n", agentName, *version, filePath)
	return 0
}

func runDelete(c *client.Client, out *output, args []string) int {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthub-cli delete <agentName>")
		return 2
	}

	agentName := fs.Arg(0)
	if err := c.DeleteAgent(agentName); err != nil {
		out.err(err)
		return 1
	}
	if out.json {
		out.writeJSON(map[string]any{"agentName": agentName, "deleted": true})
		return 0
	}
	fmt.Fprintf(out.w, "deleted %s\n", agentName)
	return 0
}

type output struct {
	json bool
	w    io.Writer
}

func (o *output) writeJSON(v any) {
	enc := json.NewEncoder(o.w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func (o *output) err(err error) {
	if o.json {
		o.writeJSON(map[string]any{"error": err.Error()})
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}

func flagArg(args []string, name string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == name && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(args[i], name+"="); ok {
			return v
		}
	}
	return ""
}

func firstCommand(args []string) string {
	skipNext := false
	for i, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "--config" || a == "--url" || a == "--token" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "--config=") || strings.HasPrefix(a, "--url=") || strings.HasPrefix(a, "--token=") {
			continue
		}
		if a == "--json" || strings.HasPrefix(a, "-") {
			continue
		}
		return args[i]
	}
	return ""
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `AgentHub CLI — manage agent packages.

Usage:
  agenthub-cli [flags] <command> [args]

Config:
  Reads agenthub-cli.toml from the current directory (copy from agenthub-cli.example.toml).
  Use --config to specify another path, or set AGENTHUB_CLI_CONFIG.

Global flags:
  --config  Path to agenthub-cli.toml
  --url     Override [cli].url
  --token   Override [cli].upload_token (required for upload/update/put-file/delete)
  --json    Output JSON

Commands:
  list [--category CATEGORY]        List agents (optionally filter by category)
  categories                        List supported agent categories
  get <agentName> [--version VERSION]  Show agent detail and file tree
  file <agentName> <path>            Read a file from agent package
  download <agentName> [-o FILE]     Download agent package as ZIP
  install [--expect-category CAT] [--dest DIR] [--version VERSION] <agentName>
  upload [--category CATEGORY] [--version VERSION] <agentName> <zip-file>
  update [--display-name NAME] [--summary TEXT] [--category CAT] <agentName>
  put-file <agentName> <path> --version VERSION [--file LOCAL]
  delete <agentName>                 Delete agent from hub (requires upload_token)

Flag order: put all flags before positional arguments (<agentName>, paths).

Examples:
  agenthub-cli list --category picoclaw
  agenthub-cli get demo-weather
  agenthub-cli file demo-weather SKILL.md
  agenthub-cli download demo-weather -o demo.zip
  agenthub-cli install --expect-category picoclaw --dest ./agents/test test
  agenthub-cli upload --category openclaw --version 1.0.0 my-skill ./pkg.zip
  agenthub-cli update --display-name "Demo" --summary "..." demo-weather
  agenthub-cli put-file demo-weather SKILL.md --version 1.0.0 --file ./SKILL.md
  agenthub-cli delete demo-weather`)
}
