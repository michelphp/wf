// @Author: F.Michel
// @github: https://github.com/phpdevcommunity
package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var (
	workflows            []Workflow
	dockerComposeCommand = ""
	workflowDepth        = 0
)

func main() {
	currentDir, _ := os.Getwd()
	wfFiles, _ := filepath.Glob(filepath.Join(currentDir, "*.wf"))

	if len(wfFiles) == 0 {
		pterm.Warning.Println("No workflow files found")
	}

	workflows = []Workflow{}
	for _, wfFile := range wfFiles {
		for _, wf := range ParseContentToWorkFlowStruct(wfFile) {
			if len(wf.Lines) > 0 {
				workflows = append(workflows, wf)
			}
		}
	}

	Commands := []*cli.Command{}
	values := InitDefaultVariables()
	for _, wf := range workflows {
		Commands = append(Commands, &cli.Command{
			Name:  wf.Name,
			Usage: wf.Comment,
			Action: func(c *cli.Context) error {
				executeWorkflow(wf, values)
				return nil
			},
		})
	}

	if len(os.Args) <= 1 {
		pterm.DefaultBigText.WithLetters(
			pterm.NewLettersFromStringWithStyle("W", pterm.NewStyle(pterm.FgCyan)),
			pterm.NewLettersFromStringWithStyle("F", pterm.NewStyle(pterm.FgBlue)),
		).Render()
	}

	app := &cli.App{Commands: Commands}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func executeWorkflow(wf Workflow, values map[string]string) {
	fmt.Println()
	start := time.Now()

	depth := workflowDepth
	workflowDepth++
	defer func() { workflowDepth-- }()

	indent := ""
	if depth > 0 {
		indent = strings.Repeat("  ", depth)
	}

	if depth == 0 {
		pterm.DefaultHeader.
			WithFullWidth().
			WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
			WithMargin(1).
			Printfln("🚀 WORKFLOW: %s", strings.ToUpper(wf.Name))
	} else {
		pterm.FgYellow.Printfln("%s╭── 📦 Sub-Workflow: %s", indent, wf.Name)
	}

	for _, line := range wf.Lines {
		executeLine(line, values)
	}

	fmt.Println()
	elapsed := time.Since(start).Round(time.Millisecond)
	if depth == 0 {
		pterm.Success.Printfln("✨ Workflow '%s' completed in %s.", wf.Name, elapsed)
		fmt.Println()
	} else {
		pterm.FgYellow.Printfln("%s╰── ✔ Completed in %s", indent, elapsed)
	}
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func Touch(filename string) {
	fd, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fatal("Failed to touch file %s: %v", filename, err)
	}
	fd.Close()
}

func Copy(source string, dest string) {
	fd1, err := os.Open(source)
	if err != nil {
		fatal("Failed to open source file %s: %v", source, err)
	}
	defer fd1.Close()

	fd2, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fatal("Failed to create destination file %s: %v", dest, err)
	}
	defer fd2.Close()

	if _, err := io.Copy(fd2, fd1); err != nil {
		fatal("Failed to copy %s to %s: %v", source, dest, err)
	}
}

func executeLine(lineOriginal string, values map[string]string) {
	if lineOriginal == "" || strings.HasPrefix(lineOriginal, "#") {
		return
	}

	lineOriginal = ResolveVariables(values, lineOriginal)
	actionOriginal := strings.Split(lineOriginal, " ")[0]
	action := strings.ToLower(actionOriginal)
	line := strings.Replace(lineOriginal, actionOriginal, action, 1)

	switch action {
	case "set":
		line = strings.TrimPrefix(line, "set ")
		parts := strings.Split(line, "=")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			fatal("Invalid set command")
		}
		values[parts[0]] = parts[1]
		printLine(pterm.Info.Sprintf("Variable %s set to %s", parts[0], values[parts[0]]))

	case "run":
		run(strings.TrimPrefix(line, "run "), true)

	case "echo":
		msg := strings.TrimSpace(strings.TrimPrefix(line, "echo "))
		if strings.HasPrefix(msg, "\"") && strings.HasSuffix(msg, "\"") {
			msg = msg[1 : len(msg)-1]
		}
		printLine(pterm.DefaultBasicText.Sprint("  "+msg))

	case "exit":
		os.Exit(0)

	case "touch":
		fileToCreate := strings.TrimSpace(strings.TrimPrefix(line, "touch "))
		if fileToCreate == "" {
			fatal("Invalid touch command")
		}
		if FileExists(fileToCreate) {
			printLine(pterm.Info.Sprintf("%s already exists, skipping...", fileToCreate))
			return
		}
		Touch(fileToCreate)
		printLine(pterm.Success.Sprintf("%s created", fileToCreate))

	case "copy", "cp":
		line = strings.TrimPrefix(line, "copy ")
		parts := strings.Fields(line)
		if len(parts) < 2 {
			fatal("Invalid copy command")
		}
		fileOrFolder, destination := parts[0], parts[1]
		if !FileExists(fileOrFolder) {
			fatal("%s not found", fileOrFolder)
		}
		if FileExists(destination) {
			printLine(pterm.Info.Sprintf("%s already exists, skipping...", destination))
			return
		}
		Copy(fileOrFolder, destination)
		printLine(pterm.Success.Sprintf("%s copied to %s", fileOrFolder, destination))

	case "mkdir":
		parts := strings.Fields(line)
		if len(parts) < 2 {
			fatal("Invalid mkdir command")
		}
		folderName := parts[1]
		if FileExists(folderName) {
			printLine(pterm.Info.Sprintf("Folder %s already exists, skipping...", folderName))
			return
		}
		if err := os.MkdirAll(folderName, os.ModePerm); err != nil {
			fatal(err.Error())
		}
		printLine(pterm.Success.Sprintf("Folder '%s' created", folderName))

	case "set_permissions":
		args := strings.Fields(line)
		if len(args) < 3 {
			fatal("Invalid set_permissions command: %s", line)
		}
		folderName := args[1]
		permissions, err := strconv.ParseUint(args[2], 8, 32)
		if err != nil {
			fatal("Invalid permissions: %s", args[2])
		}
		fi, err := os.Stat(folderName)
		if err != nil {
			fatal("Failed to access '%s': %s", folderName, err)
		}
		err = filepath.Walk(folderName, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			return os.Chmod(path, os.FileMode(permissions))
		})
		if err != nil {
			fatal("Failed to set permissions on '%s': %s", folderName, err)
		}
		if fi.IsDir() {
			printLine(pterm.Info.Sprintf("Permissions 0%o applied recursively to '%s'", permissions, folderName))
		} else {
			printLine(pterm.Info.Sprintf("Permissions 0%o applied to '%s'", permissions, folderName))
		}

	case "sync_time":
		printLine(pterm.Info.Sprintf("System time: %s", time.Now().Format("2006-01-02 15:04:05 MST")))

	case "docker_compose":
		args := strings.Fields(line)
		if len(args) < 2 {
			fatal("Invalid execute command: %s", line)
		}
		run(fmt.Sprintf("%s %s", GetDockerComposeCommand(), strings.Join(args[1:], " ")), true)

	case "wf":
		args := strings.Fields(line)
		if len(args) < 2 {
			fatal("Invalid wf command: %s", line)
		}
		wfName := args[1]
		for _, wf := range workflows {
			if wf.Name == wfName {
				executeWorkflow(wf, values)
				break
			}
		}

	case "notify_success", "notify_error", "notify_warning", "notify_info":
		notifies := map[string]*pterm.PrefixPrinter{
			"notify_success": &pterm.Success,
			"notify_error":   &pterm.Error,
			"notify_warning": &pterm.Warning,
			"notify_info":    &pterm.Info,
		}
		printer := notifies[action]
		msg := strings.TrimSpace(strings.TrimPrefix(line, action))
		if strings.HasPrefix(msg, "\"") && strings.HasSuffix(msg, "\"") {
			msg = msg[1 : len(msg)-1]
		}
		printLine(printer.Sprint(msg))

	case "notify":
		msg := strings.TrimSpace(strings.TrimPrefix(line, "notify"))
		if strings.HasPrefix(msg, "\"") && strings.HasSuffix(msg, "\"") {
			msg = msg[1 : len(msg)-1]
		}
		printLine(pterm.DefaultBasicText.Sprint("  " + msg))

	default:
		fatal("Unknown command or invalid syntax : " + line)
	}
}

func run(line string, printCommand bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	var start time.Time
	indent := ""
	if workflowDepth > 1 {
		indent = strings.Repeat("  ", workflowDepth-1)
	}

	if printCommand {
		fmt.Println()
		pterm.FgCyan.Printfln("%s╭─ ▶ %s", indent, line)
		start = time.Now()
	}

	args := strings.Fields(line)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()

	if printCommand {
		prefixStr := pterm.FgCyan.Sprint(indent + "│  ")
		cmd.Stdout = NewPrefixWriter(prefixStr, os.Stdout)
		cmd.Stderr = NewPrefixWriter(prefixStr, os.Stderr)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()

	if printCommand {
		elapsed := time.Since(start).Round(time.Millisecond)
		if err != nil {
			pterm.FgRed.Printfln("%s╰─ ✘ Failed in %s", indent, elapsed)
		} else {
			pterm.FgGreen.Printfln("%s╰─ ✔ %s", indent, elapsed)
		}
	}

	if err != nil {
		fmt.Println()
		pterm.Error.Printfln("Command failed '%s': %v", line, err)
		os.Exit(1)
	}
}

func ParseContentToWorkFlowStruct(filename string) map[string]Workflow {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	workflowsMap := map[string]Workflow{}
	scanner := bufio.NewScanner(file)
	currentWorkflow := Workflow{Name: "main"}
	workflowsMap[currentWorkflow.Name] = currentWorkflow

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			end := strings.Index(line, "]")
			if end == -1 {
				continue
			}
			wfName := line[1:end]
			comment := ""
			after := strings.TrimSpace(line[end+1:])
			if strings.HasPrefix(after, "#") {
				comment = strings.TrimSpace(strings.TrimPrefix(after, "#"))
			}

			currentWorkflow = Workflow{Name: wfName, Comment: comment}
			workflowsMap[wfName] = currentWorkflow
			continue
		}

		currentWorkflow.Lines = append(currentWorkflow.Lines, line)
		workflowsMap[currentWorkflow.Name] = currentWorkflow
	}
	return workflowsMap
}

func InitDefaultVariables() map[string]string {
	return map[string]string{
		"IP_LOCAL":     GetLocalIP(),
		"CURRENT_PATH": getCurrentDir(),
	}
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return ""
}

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func ResolveVariables(values map[string]string, line string) string {
	for key, value := range values {
		line = strings.ReplaceAll(line, "${"+key+"}", value)
	}
	return strings.ReplaceAll(line, "${GENERATE_SECRET}", TokenGenerator(32))
}

func GetDockerComposeCommand() string {
	if dockerComposeCommand == "" {
		if err := exec.Command("docker", "compose", "--version").Run(); err != nil {
			dockerComposeCommand = "docker-compose"
		} else {
			dockerComposeCommand = "docker compose"
		}
	}
	return dockerComposeCommand
}

type Workflow struct {
	Name    string
	Comment string
	Lines   []string
}

type PrefixWriter struct {
	Prefix []byte
	Writer io.Writer
	isBOL  bool
}

func NewPrefixWriter(prefix string, writer io.Writer) *PrefixWriter {
	return &PrefixWriter{
		Prefix: []byte(prefix),
		Writer: writer,
		isBOL:  true,
	}
}

func (pw *PrefixWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if pw.isBOL {
			if _, err = pw.Writer.Write(pw.Prefix); err != nil {
				return n, err
			}
			pw.isBOL = false
		}
		if _, err = pw.Writer.Write([]byte{b}); err != nil {
			return n, err
		}
		n++
		if b == '\n' {
			pw.isBOL = true
		}
	}
	return n, nil
}

func printLine(text string) {
	if workflowDepth > 1 {
		fmt.Print(strings.Repeat("  ", workflowDepth-1))
	}
	fmt.Println(text)
}

func fatal(format string, a ...interface{}) {
	printLine(pterm.Error.Sprintf(format, a...))
	os.Exit(1)
}
