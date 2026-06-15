// test app for demonstration

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PavelMkr/docline-new/pkg/docline"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	modeAutomatic = "automatic"
	modeHeuristic = "heuristic"
	modeNgram     = "ngram"
	modeOpenAI    = "openai"
)

type settings struct {
	Mode     string
	FilePath string

	MinCloneLength int
	MinGroupPower  int

	// automatic
	ConvertToDRL    bool
	ArchetypeLength int
	StrictFilter    bool

	// heuristic
	SimilarityThreshold float64

	// ngram
	MaxEdit  int
	MaxFuzzy int

	// openai
	Provider          string
	Model             string
	Prompt            string
	APIKey            string
	BaseURL           string
	Endpoint          string
	TimeoutSec        int
	ChunkMaxRunes     int
	ChunkOverlapRunes int
}

func defaultSettings() settings {
	return settings{
		Mode:           modeAutomatic,
		MinCloneLength: 20,
		MinGroupPower:  2,

		ConvertToDRL:    true,
		ArchetypeLength: 6,
		StrictFilter:    true,

		SimilarityThreshold: 0.72,
		MaxEdit:             3,
		MaxFuzzy:            2,

		Provider: "openai",
		Model:    "gpt-4o-mini",
		Prompt:   "Найди нечёткие повторы и сгруппируй по смыслу",
		TimeoutSec: 90,
	}
}

func main() {
	var (
		filePath = flag.String("file", "", "path to input file (if empty, choose in menu)")
		modeFlag = flag.String("mode", modeAutomatic, "finder mode: automatic|heuristic|ngram|openai")
		provider = flag.String("provider", "openai", "LLM provider (openai mode)")
		model    = flag.String("model", "gpt-4o-mini", "model (openai mode)")
		prompt   = flag.String("prompt", "Найди нечёткие повторы и сгруппируй по смыслу", "prompt (openai mode)")

		apiKey    = flag.String("api-key", "", "OpenAI API key (openai mode)")
		baseURL   = flag.String("base-url", "", "base URL (openai mode)")
		endpoint  = flag.String("endpoint", "", "endpoint override (openai mode)")
		timeout   = flag.Int("timeout-sec", 90, "request timeout (openai mode)")
		minPower  = flag.Int("min-group-power", 2, "minimum fragments per group")
		chunkMax  = flag.Int("chunk-max-runes", 0, "chunk size in runes (openai, 0=default)")
		chunkOvrl = flag.Int("chunk-overlap-runes", 0, "chunk overlap (openai, 0=default)")
	)
	flag.Parse()

	st := defaultSettings()
	st.FilePath = strings.TrimSpace(*filePath)
	st.Mode = normalizeMode(*modeFlag)
	st.MinGroupPower = *minPower

	st.Provider = strings.TrimSpace(*provider)
	st.Model = strings.TrimSpace(*model)
	st.Prompt = *prompt
	st.APIKey = strings.TrimSpace(*apiKey)
	st.BaseURL = strings.TrimSpace(*baseURL)
	st.Endpoint = strings.TrimSpace(*endpoint)
	st.TimeoutSec = *timeout
	st.ChunkMaxRunes = *chunkMax
	st.ChunkOverlapRunes = *chunkOvrl

	if st.APIKey == "" {
		st.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	st = menu(st)

	if strings.TrimSpace(st.FilePath) == "" {
		fatalf("file was not selected")
	}
	if _, err := os.Stat(st.FilePath); err != nil {
		fatalf("cannot access file %q: %v", st.FilePath, err)
	}

	d := docline.New(&docline.Config{
		ResultsDirectory:    "./results",
		DefaultReportFormat: "json",
		DefaultTokenizer:    "space",
		DefaultCloneFinder:  st.Mode,
	})

	groups, stats, cfg, meta, err := runAnalysis(d, st)
	if err != nil {
		fatalf("analyze failed: %v", err)
	}

	out := struct {
		Mode       string      `json:"mode"`
		SourceFile string      `json:"source_file"`
		Groups     interface{} `json:"groups"`
		Stats      interface{} `json:"stats"`
		Config     interface{} `json:"config"`
		Metadata   interface{} `json:"metadata,omitempty"`
	}{
		Mode:       st.Mode,
		SourceFile: st.FilePath,
		Groups:     groups,
		Stats:      stats,
		Config:     cfg,
		Metadata:   meta,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fatalf("marshal output: %v", err)
	}
	fmt.Println(string(data))
}

func runAnalysis(d *docline.Docline, st settings) (groups, stats, cfg, meta interface{}, err error) {
	switch st.Mode {
	case modeAutomatic:
		convert := st.ConvertToDRL
		archetype := st.ArchetypeLength
		strict := st.StrictFilter
		res, err := d.AnalyzeDocument(st.FilePath, modeAutomatic, docline.AutomaticConfig{
			MinCloneLength:  st.MinCloneLength,
			MinGroupPower:   st.MinGroupPower,
			ConvertToDRL:    &convert,
			ArchetypeLength: &archetype,
			StrictFilter:    &strict,
		})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return res.Groups, res.Statistics, res.Config, res.Metadata, nil
	case modeHeuristic:
		res, err := d.AnalyzeDocument(st.FilePath, modeHeuristic, docline.HeuristicConfig{
			MinCloneLength:      st.MinCloneLength,
			MinGroupPower:       st.MinGroupPower,
			SimilarityThreshold: st.SimilarityThreshold,
		})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return res.Groups, res.Statistics, res.Config, res.Metadata, nil
	case modeNgram:
		res, err := d.AnalyzeDocument(st.FilePath, modeNgram, docline.NgramConfig{
			MinCloneLength: st.MinCloneLength,
			MinGroupPower:  st.MinGroupPower,
			MaxEdit:        st.MaxEdit,
			MaxFuzzy:       st.MaxFuzzy,
		})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return res.Groups, res.Statistics, res.Config, res.Metadata, nil
	case modeOpenAI:
		res, err := d.AnalyzeDocument(st.FilePath, modeOpenAI, docline.OpenAIConfig{
			MinGroupPower:     st.MinGroupPower,
			Provider:          strings.TrimSpace(st.Provider),
			Model:             strings.TrimSpace(st.Model),
			Prompt:            st.Prompt,
			APIKey:            strings.TrimSpace(st.APIKey),
			BaseURL:           strings.TrimSpace(st.BaseURL),
			Endpoint:          strings.TrimSpace(st.Endpoint),
			TimeoutSec:        st.TimeoutSec,
			ChunkMaxRunes:     st.ChunkMaxRunes,
			ChunkOverlapRunes: st.ChunkOverlapRunes,
		})
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return res.Groups, res.Statistics, res.Config, res.Metadata, nil
	default:
		return nil, nil, nil, nil, fmt.Errorf("unknown mode %q", st.Mode)
	}
}

func menu(st settings) settings {
	in := bufio.NewReader(os.Stdin)

	for {
		printMenu(st)
		fmt.Fprint(os.Stderr, "Выбор: ")
		raw, _ := in.ReadString('\n')
		choice := strings.TrimSpace(raw)
		if choice == "" {
			continue
		}
		switch choice {
		case "0":
			fatalf("cancelled")
		case "1":
			picked, ok := pickFileInteractive(st.FilePath)
			if ok {
				st.FilePath = picked
			}
		case "2":
			st.Mode = promptMode(in, st.Mode)
		case "s", "S", "start", "Start", "старт", "Старт":
			if strings.TrimSpace(st.FilePath) == "" {
				fmt.Fprintln(os.Stderr, "Файл не выбран. Выберите файл пунктом 1.")
				continue
			}
			if _, err := os.Stat(st.FilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Файл недоступен: %v\n", err)
				continue
			}
			return st
		default:
			st = handleModeSetting(in, st, choice)
		}
	}
}

func handleModeSetting(in *bufio.Reader, st settings, choice string) settings {
	switch st.Mode {
	case modeAutomatic:
		switch choice {
		case "3":
			st.MinCloneLength = promptInt(in, "Min clone length", st.MinCloneLength, 1, 100000)
		case "4":
			st.MinGroupPower = promptInt(in, "Min group power", st.MinGroupPower, 2, 1000)
		case "5":
			st.ConvertToDRL = promptBool(in, "Convert to DRL", st.ConvertToDRL)
		case "6":
			st.ArchetypeLength = promptInt(in, "Archetype length", st.ArchetypeLength, 1, 1000)
		case "7":
			st.StrictFilter = promptBool(in, "Strict filter", st.StrictFilter)
		default:
			fmt.Fprintln(os.Stderr, "Неизвестный пункт меню.")
		}
	case modeHeuristic:
		switch choice {
		case "3":
			st.MinCloneLength = promptInt(in, "Min clone length", st.MinCloneLength, 1, 100000)
		case "4":
			st.MinGroupPower = promptInt(in, "Min group power", st.MinGroupPower, 2, 1000)
		case "5":
			st.SimilarityThreshold = promptFloat(in, "Similarity threshold", st.SimilarityThreshold, 0, 1)
		default:
			fmt.Fprintln(os.Stderr, "Неизвестный пункт меню.")
		}
	case modeNgram:
		switch choice {
		case "3":
			st.MinCloneLength = promptInt(in, "Min clone length", st.MinCloneLength, 1, 100000)
		case "4":
			st.MinGroupPower = promptInt(in, "Min group power", st.MinGroupPower, 2, 1000)
		case "5":
			st.MaxEdit = promptInt(in, "Max edit", st.MaxEdit, 0, 100)
		case "6":
			st.MaxFuzzy = promptInt(in, "Max fuzzy", st.MaxFuzzy, 0, 100)
		default:
			fmt.Fprintln(os.Stderr, "Неизвестный пункт меню.")
		}
	case modeOpenAI:
		switch choice {
		case "3":
			st.MinGroupPower = promptInt(in, "Min group power", st.MinGroupPower, 2, 1000)
		case "4":
			st.Provider = strings.ToLower(strings.TrimSpace(promptString(in, "Provider (openai|ollama)", st.Provider)))
			if st.Provider == "" {
				st.Provider = "openai"
			}
		case "5":
			st.Model = strings.TrimSpace(promptString(in, "Model", st.Model))
		case "6":
			st.Prompt = promptMultiline(in, "Prompt (закончить пустой строкой)", st.Prompt)
		case "7":
			st.BaseURL = strings.TrimSpace(promptString(in, "Base URL", st.BaseURL))
		case "8":
			st.Endpoint = strings.TrimSpace(promptString(in, "Endpoint", st.Endpoint))
		case "9":
			st.TimeoutSec = promptInt(in, "Timeout (sec)", st.TimeoutSec, 1, 3600)
		case "10":
			st.ChunkMaxRunes = promptInt(in, "Chunk max runes (0=default)", st.ChunkMaxRunes, 0, 10000000)
		case "11":
			st.ChunkOverlapRunes = promptInt(in, "Chunk overlap runes (0=default)", st.ChunkOverlapRunes, 0, 10000000)
		case "12":
			st.APIKey = strings.TrimSpace(promptString(in, "API key (empty = OPENAI_API_KEY)", st.APIKey))
			if st.APIKey == "" {
				st.APIKey = os.Getenv("OPENAI_API_KEY")
			}
		default:
			fmt.Fprintln(os.Stderr, "Неизвестный пункт меню.")
		}
	default:
		fmt.Fprintln(os.Stderr, "Неизвестный режим.")
	}
	return st
}

func printMenu(st settings) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "=== Docline analyzer ===")
	fmt.Fprintln(os.Stderr, "Выберите пункт и нажмите Enter (S = старт, 0 = выход).")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, " 1) Файл: %s\n", emptyAsDash(st.FilePath))
	fmt.Fprintf(os.Stderr, " 2) Режим: %s\n", st.Mode)

	switch st.Mode {
	case modeAutomatic:
		fmt.Fprintf(os.Stderr, " 3) Min clone length: %d\n", st.MinCloneLength)
		fmt.Fprintf(os.Stderr, " 4) Min group power: %d\n", st.MinGroupPower)
		fmt.Fprintf(os.Stderr, " 5) Convert to DRL: %s\n", boolLabel(st.ConvertToDRL))
		fmt.Fprintf(os.Stderr, " 6) Archetype length: %d\n", st.ArchetypeLength)
		fmt.Fprintf(os.Stderr, " 7) Strict filter: %s\n", boolLabel(st.StrictFilter))
	case modeHeuristic:
		fmt.Fprintf(os.Stderr, " 3) Min clone length: %d\n", st.MinCloneLength)
		fmt.Fprintf(os.Stderr, " 4) Min group power: %d\n", st.MinGroupPower)
		fmt.Fprintf(os.Stderr, " 5) Similarity threshold: %.2f\n", st.SimilarityThreshold)
	case modeNgram:
		fmt.Fprintf(os.Stderr, " 3) Min clone length: %d\n", st.MinCloneLength)
		fmt.Fprintf(os.Stderr, " 4) Min group power: %d\n", st.MinGroupPower)
		fmt.Fprintf(os.Stderr, " 5) Max edit: %d\n", st.MaxEdit)
		fmt.Fprintf(os.Stderr, " 6) Max fuzzy: %d\n", st.MaxFuzzy)
	case modeOpenAI:
		fmt.Fprintf(os.Stderr, " 3) Min group power: %d\n", st.MinGroupPower)
		fmt.Fprintf(os.Stderr, " 4) Provider: %s\n", st.Provider)
		fmt.Fprintf(os.Stderr, " 5) Model: %s\n", st.Model)
		fmt.Fprintf(os.Stderr, " 6) Prompt: %s\n", oneLinePreview(st.Prompt, 72))
		fmt.Fprintf(os.Stderr, " 7) Base URL: %s\n", emptyAsDash(st.BaseURL))
		fmt.Fprintf(os.Stderr, " 8) Endpoint: %s\n", emptyAsDash(st.Endpoint))
		fmt.Fprintf(os.Stderr, " 9) Timeout sec: %d\n", st.TimeoutSec)
		fmt.Fprintf(os.Stderr, "10) Chunk max runes: %d\n", st.ChunkMaxRunes)
		fmt.Fprintf(os.Stderr, "11) Chunk overlap runes: %d\n", st.ChunkOverlapRunes)
		fmt.Fprintf(os.Stderr, "12) API key: %s\n", apiKeyPreview(st.APIKey))
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "S) Старт анализа")
	fmt.Fprintln(os.Stderr, "0) Выход")
	fmt.Fprintln(os.Stderr, "")
}

func promptMode(in *bufio.Reader, current string) string {
	for {
		fmt.Fprintf(os.Stderr, "Режим (automatic/heuristic/ngram/openai) [%s]: ", current)
		s, _ := in.ReadString('\n')
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			return current
		}
		m := normalizeMode(s)
		if isKnownMode(m) {
			return m
		}
		fmt.Fprintln(os.Stderr, "Неизвестный режим. Допустимо: automatic, heuristic, ngram, openai.")
	}
}

func normalizeMode(m string) string {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case modeAutomatic, "auto":
		return modeAutomatic
	case modeHeuristic:
		return modeHeuristic
	case modeNgram, "n-gram":
		return modeNgram
	case modeOpenAI, "open-ai", "ai":
		return modeOpenAI
	default:
		return strings.ToLower(strings.TrimSpace(m))
	}
}

func isKnownMode(m string) bool {
	switch m {
	case modeAutomatic, modeHeuristic, modeNgram, modeOpenAI:
		return true
	default:
		return false
	}
}

func pickFileInteractive(current string) (string, bool) {
	startDir := resolveStartDir(current)
	p, ok := runFMStylePicker(startDir)
	if !ok {
		return "", false
	}
	abs, _ := filepath.Abs(p)
	return abs, true
}

func resolveStartDir(current string) string {
	cur := strings.TrimSpace(current)
	if cur == "" {
		if wd, err := os.Getwd(); err == nil {
			abs, _ := filepath.Abs(wd)
			return abs
		}
		abs, _ := filepath.Abs(".")
		return abs
	}
	if fi, err := os.Stat(cur); err == nil {
		if fi.IsDir() {
			abs, _ := filepath.Abs(cur)
			return abs
		}
		abs, _ := filepath.Abs(filepath.Dir(cur))
		return abs
	}
	abs, _ := filepath.Abs(filepath.Dir(cur))
	return abs
}

func runFMStylePicker(startDir string) (string, bool) {
	app := tview.NewApplication()

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetText("Выбор файла (↑/↓ перемещение, Enter открыть/выбрать, Backspace вверх, Esc отмена)")
	header.SetBorder(true)

	status := tview.NewTextView().SetDynamicColors(true)
	status.SetBorder(true)

	rootPath := startDir
	root := tview.NewTreeNode(rootPath).SetReference(rootPath).SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true)

	readDirFM(root, rootPath)

	var picked string

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref == nil {
			return
		}
		p := ref.(string)
		if isDirFM(p) {
			children := node.GetChildren()
			if len(children) == 0 {
				readDirFM(node, p)
				node.SetExpanded(true)
			} else {
				node.SetExpanded(!node.IsExpanded())
				if !node.IsExpanded() {
					node.ClearChildren()
				}
			}
			return
		}

		picked = p
		app.Stop()
	})

	tree.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyESC:
			picked = ""
			app.Stop()
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			parent := filepath.Dir(rootPath)
			if parent == rootPath {
				status.SetText("[yellow]Уже на верхнем уровне[-]")
				return nil
			}
			rootPath = parent
			newRoot := tview.NewTreeNode(rootPath).SetReference(rootPath).SetColor(tcell.ColorRed)
			readDirFM(newRoot, rootPath)
			tree.SetRoot(newRoot).SetCurrentNode(newRoot)
			return nil
		}
		return ev
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(header, 3, 1, false)
	flex.AddItem(tree, 0, 1, true)
	flex.AddItem(status, 3, 1, false)

	if err := app.SetRoot(flex, true).EnableMouse(false).Run(); err != nil {
		return "", false
	}
	if strings.TrimSpace(picked) == "" {
		return "", false
	}
	return picked, true
}

func readDirFM(target *tview.TreeNode, dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, file := range files {
		name := file.Name()
		full := filepath.Join(dir, name)
		node := tview.NewTreeNode(name).SetReference(full)
		if file.IsDir() {
			node.SetColor(tcell.ColorGreen)
			node.SetText("🗁 " + name)
		}
		target.AddChild(node)
	}
}

func isDirFM(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func promptString(in *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", label)
	}
	s, _ := in.ReadString('\n')
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

func promptInt(in *bufio.Reader, label string, def, min, max int) int {
	for {
		raw := promptString(in, label, fmt.Sprint(def))
		n, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ожидается целое число.")
			continue
		}
		if n < min || n > max {
			fmt.Fprintf(os.Stderr, "Ожидается число в диапазоне [%d..%d].\n", min, max)
			continue
		}
		return n
	}
}

func promptFloat(in *bufio.Reader, label string, def, min, max float64) float64 {
	for {
		raw := promptString(in, label, fmt.Sprintf("%.2f", def))
		n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ожидается число.")
			continue
		}
		if n < min || n > max {
			fmt.Fprintf(os.Stderr, "Ожидается число в диапазоне [%.2f..%.2f].\n", min, max)
			continue
		}
		return n
	}
}

func promptBool(in *bufio.Reader, label string, def bool) bool {
	defStr := "n"
	if def {
		defStr = "y"
	}
	for {
		raw := strings.ToLower(promptString(in, label+" (y/n)", defStr))
		switch raw {
		case "y", "yes", "да", "1", "true":
			return true
		case "n", "no", "нет", "0", "false":
			return false
		case "":
			return def
		default:
			fmt.Fprintln(os.Stderr, "Введите y или n.")
		}
	}
}

func promptMultiline(in *bufio.Reader, label, def string) string {
	fmt.Fprintln(os.Stderr, label)
	if def != "" {
		fmt.Fprintln(os.Stderr, "(по умолчанию ниже; Enter на пустой строке чтобы оставить как есть)")
		fmt.Fprintln(os.Stderr, "---")
		fmt.Fprintln(os.Stderr, def)
		fmt.Fprintln(os.Stderr, "---")
	}
	fmt.Fprintln(os.Stderr, "Введите текст. Пустая строка завершает ввод.")

	var lines []string
	for {
		fmt.Fprint(os.Stderr, "> ")
		s, _ := in.ReadString('\n')
		s = strings.TrimRight(s, "\r\n")
		if strings.TrimSpace(s) == "" {
			break
		}
		lines = append(lines, s)
	}
	if len(lines) == 0 {
		return def
	}
	return strings.Join(lines, "\n")
}

func boolLabel(v bool) string {
	if v {
		return "да"
	}
	return "нет"
}

func oneLinePreview(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if s == "" {
		return "-"
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func emptyAsDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func apiKeyPreview(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + "…" + s[len(s)-2:]
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
