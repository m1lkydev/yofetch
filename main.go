// Copyright (c) 2026 m1lkydev.
// Licensed under the MIT License.
// Contributors: AnmiTaliDev - security improvements

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"strconv"
	"strings"
	"unicode/utf8"

	"path/filepath"

	"github.com/spf13/pflag"
	lua "github.com/yuin/gopher-lua"
)

const (
	Version string = "v0.2.2-alpha"
)

type Config struct {
	Modules      []Module
	Separator    string
	Logo         string
	PaddingRight int
	PaddingLeft  int
	PaddingTop   int
	Mode         string
	Shell        []string
}

type Module struct {
	Text string
}

var (
	config     string
	configArgs string
	version    bool
	debug      bool
)

func formatText(s string) string {
	colors := map[string]string{
		"black":         "\033[30m",
		"red":           "\033[31m",
		"green":         "\033[32m",
		"yellow":        "\033[33m",
		"blue":          "\033[34m",
		"magenta":       "\033[35m",
		"cyan":          "\033[36m",
		"white":         "\033[37m",
		"black_light":   "\033[90m",
		"red_light":     "\033[91m",
		"green_light":   "\033[92m",
		"yellow_light":  "\033[93m",
		"blue_light":    "\033[94m",
		"magenta_light": "\033[95m",
		"cyan_light":    "\033[96m",
		"white_light":   "\033[97m",
	}

	styles := map[string]string{
		"reset":            "\033[0m",
		"bold":             "\033[1m",
		"faint":            "\033[2m",
		"italic":           "\033[3m",
		"underline":        "\033[4m",
		"blink":            "\033[5m",
		"blink_fast":       "\033[6m",
		"reverse":          "\033[7m",
		"hidden":           "\033[8m",
		"strikethrough":    "\033[9m",
		"double_underline": "\033[21m",
		"overline":         "\033[53m",
	}

	re := regexp.MustCompile(`\{(\w+)\.([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		category, value := submatches[1], submatches[2]

		switch category {
		case "color":
			if strings.HasPrefix(value, "#") {
				hex := strings.TrimPrefix(value, "#")
				if len(hex) != 6 {
					return match
				}

				r, _ := strconv.ParseUint(hex[0:2], 16, 8)
				g, _ := strconv.ParseUint(hex[2:4], 16, 8)
				b, _ := strconv.ParseUint(hex[4:6], 16, 8)

				return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
			} else if code, ok := colors[value]; ok {
				return code
			}

		case "style":
			if code, ok := styles[value]; ok {
				return code
			}
		}

		return match
	})
}

func echoMsg(str string, detail interface{}, type_ string) {
	var prefix, message, time_ string
	time_ = time.Now().Format("15:04:05.000000000")

	switch type_ {
	case "err":
		reason, ok := detail.(error)
		if !ok {
			fmt.Fprintf(os.Stderr, "expected error type for 'err', got %T\n", detail)
			return
		}
		prefix = fmt.Sprintf("{style.bold}{color.red_light}(!) [%s]\n%s:{style.reset}", time_, str)
		message = reason.Error()

	case "warn":
		msg, ok := detail.(string)
		if !ok {
			fmt.Fprintf(os.Stderr, "expected string type for 'warn', got %T\n", detail)
			return
		}
		prefix = fmt.Sprintf("{style.bold}{color.yellow_light}/!\\ [%s]\n%s:{style.reset}", time_, str)
		message = msg

	case "info":
		msg, ok := detail.(string)
		if !ok {
			fmt.Fprintf(os.Stderr, "expected string type for 'info', got %T\n", detail)
			return
		}
		prefix = fmt.Sprintf("{style.bold}{color.blue_light}(i) [%s]\n%s:{style.reset}", time_, str)
		message = msg

	case "debug":
		msg, ok := detail.(string)
		if !ok {
			fmt.Fprintf(os.Stderr, "expected string type for 'debug', got %T\n", detail)
			return
		}
		prefix = fmt.Sprintf("{style.bold}{color.cyan}(&) [%s]\n%s:{style.reset}", time_, str)
		message = msg

	default:
		fmt.Fprintf(os.Stderr, "unknown message type: %s\n", type_)
		return
	}

	if type_ == "debug" && !debug {
		return
	}

	fmt.Println(formatText(prefix))
	fmt.Fprintln(os.Stderr, message)
	fmt.Println()
}

func execCmd(cmd string, shell []string) (string, error) {
	var command *exec.Cmd

	echoMsg("Debug mode (go func 'execCmd')", "appending shell and command...", "debug")
	args := append(shell[1:], cmd)
	command = exec.Command(shell[0], args...)

	echoMsg("Debug mode (go func 'execCmd')", "executing command:\n"+cmd, "debug")
	out, err := command.CombinedOutput()
	trimmedOut := strings.TrimSpace(string(out))

	if err != nil {
		return trimmedOut, fmt.Errorf("%s: %w", trimmedOut, err)
	}

	echoMsg("Debug mode (go func 'execCmd')", "result of the command:\n"+trimmedOut, "debug")

	return trimmedOut, nil
}

func RegisterModule(L *lua.LState, name string, methods map[string]lua.LGFunction, fields map[string]lua.LValue) {
	L.PreloadModule(name, func(L *lua.LState) int {
		echoMsg("Debug mode (go func 'RegisterModule')", "creating the table...", "debug")
		module := L.NewTable()

		echoMsg("Debug mode (go func 'RegisterModule')", "setting funcs...", "debug")
		L.SetFuncs(module, methods)

		echoMsg("Debug mode (go func 'RegisterModule')", "adding fields...", "debug")
		for key, value := range fields {
			L.SetField(module, key, value)
		}
		L.Push(module)
		return 1
	})
}

func loadConfig(filename string, configArgs string) (*Config, error) {
	cleanPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("config file not accessible: %w", err)
	}

	L := lua.NewState()
	defer L.Close()

	cfg := &Config{
		PaddingRight: 2,
		PaddingLeft:  0,
		PaddingTop:   0,
		Mode:         "default",
		Shell:        []string{"/bin/sh", "-c"},
		Logo: `{style.bold}{color.green}      \  /{style.reset}
{style.bold}{color.green}       \/{style.reset}
{style.bold}{color.yellow}================{style.reset}
{style.bold}{color.yellow}\ -         -  /{style.reset}
{style.bold}{color.yellow} \     -      /{style.reset}
{style.bold}{color.yellow}  \          /{style.reset}
{style.bold}{color.yellow}   \    -   /{style.reset}
{style.bold}{color.yellow}    \      /{style.reset}
{style.bold}{color.yellow}     \ -  /{style.reset}
{style.bold}{color.yellow}      \  /{style.reset}
{style.bold}{color.yellow}       \/{style.reset}`,
		Separator: ": ",
	}

	echoMsg("Debug mode (go func 'loadConfig')", "setting the fields...", "debug")
	fields := map[string]lua.LValue{
		"version":     lua.LString(Version),
		"config_args": lua.LString(configArgs),
	}

	echoMsg("Debug mode (go func 'loadConfig')", "registering modules in lua...", "debug")
	RegisterModule(L, "Yofetch", map[string]lua.LGFunction{
		"exec": func(L *lua.LState) int {
			cmd := L.CheckString(1)

			echoMsg("Debug mode (lua func 'exec')", "executing go func 'execCmd' with arg:\n"+cmd, "debug")
			output, err := execCmd(cmd, cfg.Shell)
			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}

			L.Push(lua.LString(output))
			return 1
		},

		"logo": func(L *lua.LState) int {
			pathOrContent := L.CheckString(1)

			echoMsg("Debug mode (lua func 'logo')", "setting logo as:\n"+pathOrContent, "debug")

			if data, err := os.ReadFile(pathOrContent); err == nil {
				cfg.Logo = string(data)
			} else {
				cfg.Logo = pathOrContent
			}

			return 0
		},

		"padding": func(L *lua.LState) int {
			paddingRight := L.CheckInt(1)
			paddingLeft := L.CheckInt(2)
			paddingTop := L.CheckInt(3)

			echoMsg("Debug mode (lua func 'padding')", fmt.Sprintf("setting padding as:\nRight: %d, Left: %d, Top: %d",
				paddingRight, paddingLeft, paddingTop), "debug")

			if paddingRight < 0 || paddingLeft < 0 || paddingTop < 0 {
				L.RaiseError("padding cannot be negative")
			}

			cfg.PaddingRight = paddingRight
			cfg.PaddingLeft = paddingLeft
			cfg.PaddingTop = paddingTop
			return 0
		},

		"mode": func(L *lua.LState) int {
			mode := L.CheckString(1)

			echoMsg("Debug mode (lua func 'mode')", "setting mode as: "+mode, "debug")

			cfg.Mode = mode
			return 0
		},

		"print": func(L *lua.LState) int {
			text := L.CheckString(1)

			echoMsg("Debug mode (lua func 'print')", "printing the text:\n"+text, "debug")

			cfg.Modules = append(cfg.Modules, Module{
				Text: text,
			})
			return 0
		},

		"shell": func(L *lua.LState) int {
			table := L.CheckTable(1)

			echoMsg("Debug mode (lua func 'shell')", "setting the shell...", "debug")

			var shellname []string

			table.ForEach(func(_, value lua.LValue) {
				if str, ok := value.(lua.LString); ok {
					shellname = append(shellname, string(str))
				}
			})

			cfg.Shell = shellname
			return 0
		},
	}, fields)

	echoMsg("Debug mode (config)", "executing config file...", "debug")
	err = L.DoFile(cleanPath)

	if err != nil {
		return nil, fmt.Errorf("Lua config error:\n%w", err)
	}

	return cfg, nil
}

func buildInfo(cfg *Config) string {
	var info strings.Builder

	for _, module := range cfg.Modules {

		textFormatted := formatText(module.Text)
		info.WriteString(textFormatted + "\n")
	}

	return info.String()
}

func printLogoWithInfo(logo, info string, distance, paddingLeft, paddingTop int) {
	echoMsg("Debug mode (go func 'printLogoWithInfo')", "printing logo with info...", "debug")

	var ansiRegex = regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)
	type Line struct {
		Original string
		Clean    string
	}

	echoMsg("Debug mode (go func 'printLogoWithInfo')", "preparing the logo lines...", "debug")
	var logoLines []Line
	for i := 0; i < paddingTop; i++ {
		logoLines = append(logoLines, Line{
			Original: "",
			Clean:    "",
		})
	}
	for _, line := range strings.Split(strings.Trim(logo, "\n"), "\n") {
		clean := ansiRegex.ReplaceAllString(line, "")
		logoLines = append(logoLines, Line{
			Original: line,
			Clean:    clean,
		})
	}

	echoMsg("Debug mode (go func 'printLogoWithInfo')", "preparing the information lines...", "debug")
	var infoLines []Line
	for _, line := range strings.Split(strings.Trim(info, "\n"), "\n") {
		clean := ansiRegex.ReplaceAllString(line, "")
		infoLines = append(infoLines, Line{
			Original: line,
			Clean:    clean,
		})
	}

	echoMsg("Debug mode (go func 'printLogoWithInfo')", "calculating the maximum width of the logo...", "debug")
	maxLogoWidth := 0
	for _, line := range logoLines {
		if width := utf8.RuneCountInString(line.Clean); width > maxLogoWidth {
			maxLogoWidth = width
		}
	}

	maxLines := max(len(logoLines), len(infoLines))
	for i := 0; i < maxLines; i++ {
		var (
			logoLine Line
			infoLine Line
		)

		if i < len(logoLines) {
			logoLine = logoLines[i]
		}
		if i < len(infoLines) {
			infoLine = infoLines[i]
		}

		padding := maxLogoWidth - utf8.RuneCountInString(logoLine.Clean) + distance

		var output strings.Builder
		output.WriteString(strings.Repeat(" ", paddingLeft))
		output.WriteString(logoLine.Original)
		output.WriteString(strings.Repeat(" ", padding))
		output.WriteString(infoLine.Original)

		if !ansiRegex.MatchString(infoLine.Original) {
			output.WriteString("\033[0m")
		}

		output.WriteByte('\n')
		fmt.Print(output.String())
	}
	os.Stdout.Sync()
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		echoMsg("Error getting home directory", err, "err")
		os.Exit(1)
	}

	pflag.StringVarP(&config, "config", "c", filepath.Join(home, ".config", "yofetch", "config.lua"), "path to config")
	pflag.StringVarP(&configArgs, "config-args", "a", "", "args for config")
	pflag.BoolVarP(&version, "version", "v", false, "show version")
	pflag.BoolVarP(&debug, "debug", "d", false, "enable debug")

	pflag.Parse()

	if version {
		echoMsg("Debug mode (yofetch)", "showing version...", "debug")
		fmt.Println("────────────────────────")
		fmt.Printf(
			formatText("  {color.yellow}Yofetch {color.green}%s{style.reset}\n  © 2026 m1lkydev\n"),
			Version,
		)
		fmt.Println("────────────────────────")

		os.Exit(0)
	}

	echoMsg("Debug mode (yofetch config)", "loading config "+config, "debug")
	cfg, err := loadConfig(config, configArgs)
	if err != nil {
		echoMsg("Error loading config", err, "err")
		os.Exit(1)
	}

	echoMsg("Debug mode (go func 'formatText')", "formatting the logo...", "debug")
	logoFormatted := formatText(string(cfg.Logo))

	echoMsg("Debug mode (go func 'buildInfo')", "building information...", "debug")
	info := buildInfo(cfg)

	switch {
	case cfg.Logo == "":
		echoMsg("Debug mode (go func 'main')", "there is no logo, we just write the information", "debug")
		fmt.Println(info)
	case cfg.Mode == "default":
		echoMsg("Debug mode (go func 'main')", "mode installed as 'default'", "debug")
		printLogoWithInfo(logoFormatted, info, cfg.PaddingRight, cfg.PaddingLeft, cfg.PaddingTop)
	case cfg.Mode == "vertical":
		echoMsg("Debug mode (go func 'main')", "mode installed as 'vertical'", "debug")
		fmt.Println(logoFormatted)
		fmt.Println(info)
	default:
		echoMsg("Debug mode (go func 'main')", "mode not installed, uses 'default'", "debug")
		printLogoWithInfo(logoFormatted, info, cfg.PaddingRight, cfg.PaddingLeft, cfg.PaddingTop)
	}
}
