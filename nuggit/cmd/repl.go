package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/wenooij/jsong"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/edges"
	"github.com/wenooij/nuggit/graphs"
)

var replFlags struct {
	Graph  string
	Format string
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Create and edit Nuggit program graphs",
	RunE:  runRepl,
}

type replState struct {
	pendingEdits bool
	quit         bool
	rl           *readline.Instance
	r            *bufio.Reader
	filename     string
	undoStack    []*nuggit.Graph
	plan         nuggit.Plan
	b            graphs.Builder
}

func runRepl(cmd *cobra.Command, args []string) error {
	log.SetFlags(0)
	defer func() { log.Println("Goodbye!") }()

	rl, err := readline.New(":")
	if err != nil {
		return err
	}
	defer rl.Close()

	state := &replState{
		rl: rl,
		r:  &bufio.Reader{},
	}
	if replFlags.Graph != "" {
		state.filename = replFlags.Graph
		g, err := graphs.FromFile(state.filename)
		if err != nil {
			return err
		}
		state.undoStack = append(state.undoStack, g.Graph())
	} else {
		state.filename = "nuggit.json"
		state.undoStack = append(state.undoStack, new(graphs.Builder).Build())
	}

	for !state.quit {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				state.quit = true
				continue
			}
			log.Println(err)
			continue
		}
		line = strings.TrimSpace(line)
		if err := runCmd(state, line); err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
}

func runCmd(state *replState, cmd string) error {
	if cmd == "" {
		return nil
	}
	state.r.Reset(strings.NewReader(cmd))
	b, err := state.r.Peek(1)
	if err != nil {
		return err
	}
	switch b[0] {
	case 'c':
		return runCreate(state)
	case 'd':
		return runDelete(state)
	case 'e':
		return runEdit(state)
	case 'f':
		return runFetch(state)
	case 'h':
		return runHelp(state)
	case 'q':
		return runQuit(state)
	case 'o':
		return runOpen(state)
	case 'r':
		return runRead(state)
	case 'u':
		return runUpdate(state)
	case 'w':
		return runWrite(state)
	case 'x':
		return runDiff(state)
	default:
		return fmt.Errorf("bad sequence :%c", cmd[0])
	}
}

func runCreate(state *replState) error {
	state.pendingEdits = true
	cmd, err := consumeDelim(state.r, ' ', true)
	if err != nil {
		return err
	}
	switch cmd := strings.TrimSpace(string(cmd)); cmd {
	case "c", "cn":
		op, err := prompt(state.rl, "op", false)
		if err != nil {
			return err
		}
		state.b.Node(op)
	case "ce":
		key, err := prompt(state.rl, "key", true)
		if err != nil {
			return err
		}
		if key == "" {
			key = state.b.NextEdgeKey()
		}
		dst, err := prompt(state.rl, "dst", false)
		if err != nil {
			return err
		}
		src, err := prompt(state.rl, "src", false)
		if err != nil {
			return err
		}
		dstField, err := prompt(state.rl, "dstfield", true)
		if err != nil {
			return err
		}
		srcField, err := prompt(state.rl, "srcfield", true)
		if err != nil {
			return err
		}
		data, err := prompt(state.rl, "data", true)
		if err != nil {
			return err
		}
		var val any
		if data != "" {
			if err := json.Unmarshal([]byte(data), &val); err != nil {
				return err
			}
		}
		oldEdge, replaced := state.b.InsertEdge(key, dst, src, dstField, srcField, val)
		if replaced {
			log.Println(edges.Format(oldEdge))
		}
	case "cg":
	case "cx":
	default:
		return fmt.Errorf("unexpected sequence :%s", cmd)
	}

	return nil
}
func runRead(state *replState) error {
	state.r.Discard(1) // 'r'
	s, ok := consumeAnyWithDefault(state.r, "egnx", 'n')
	if !ok {
		return fmt.Errorf("unexpected sequence :r%c", s)
	}
	g := state.b.Build()
	switch s {
	case 'e':
		for _, e := range g.Edges {
			data, err := json.Marshal(e.Data)
			if err != nil {
				log.Printf("Failed to marshal edge: %v", err)
				break
			}
			log.Printf("%s\t%s", edges.Format(e), string(data))
		}
	case 'g':
		data, err := json.Marshal(g)
		if err != nil {
			log.Printf("Failed to marshal graph: %v", err)
			break
		}
		log.Println(string(data))
	case 'n':
		for _, n := range g.Nodes {
			data, err := json.Marshal(n.Data)
			if err != nil {
				log.Printf("Failed to marshal node: %v", err)
				continue
			}
			log.Printf("%s(%s)\t%s", n.Op, n.Key, string(data))
		}
	case 'x':
		fmt.Println("exchanges")
	default:
		panic("unreachable")
	}

	return nil
}
func runUpdate(state *replState) error {
	state.r.Discard(1) // 'u'
	s, ok := consumeAnyWithDefault(state.r, "egnx", 'n')
	if !ok {
		return fmt.Errorf("unexpected sequence :c%c", s)
	}
	create := consumeBytes(state.r, []byte("!"))

	_ = create

	switch s {
	case 'e':
	case 'g':
	case 'n':
		g := graphs.FromGraph(state.b.Build())
		defer func() { state.rl.Operation.SetPrompt(":") }()
		state.rl.Operation.SetPrompt("key:")
		key, err := state.rl.Readline()
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		n, ok := g.Nodes[key]
		if !ok {
			if !create {
				return fmt.Errorf("no node %q", key)
			}
			n = nuggit.Node{}
			g.Nodes[key] = n
		}
		state.rl.Operation.SetPrompt("path:")
		path, err := state.rl.Readline()
		if err != nil {
			return err
		}
		path = strings.TrimSpace(path)
		state.rl.Operation.SetPrompt("val:")
		val, err := state.rl.Readline()
		if err != nil {
			return err
		}
		v, err := jsong.NewDecoder(strings.NewReader(val)).Decode()
		if err != nil {
			return err
		}
		n.Data = jsong.Merge(n.Data, v, path, "")
		g.Nodes[key] = n
		state.b.Reset(g)
	case 'x':
	default:
		panic("unepexted state :c%c")
	}

	return nil
}
func runDelete(state *replState) error {
	state.r.Discard(1) // 'd'

	s, ok := consumeAnyWithDefault(state.r, "egnx", 'n')
	if !ok {
		return fmt.Errorf("unexpected sequence :d%c", s)
	}

	force := consumeBytes(state.r, []byte("!"))

	_ = force
	switch s {
	case 'e':
	case 'g':
		state.b.Init()
	case 'n':
		key, err := prompt(state.rl, "key", false)
		if err != nil {
			return err
		}
		n, es, ok := state.b.Delete(key)
		if ok {
			log.Printf("dn %s(%s)", n.Op, n.Key)
			for _, e := range es {
				log.Println("de", edges.Format(e))
			}
		}
	case 'x':
	default:
		panic("unreachable")
	}
	return nil
}
func runOpen(state *replState) error {
	state.r.Discard(1) // 'o'
	return nil
}
func runEdit(state *replState) error {
	state.r.Discard(1) // 'e'
	s, ok := consumeAnyWithDefault(state.r, "egnx", 'n')
	if !ok {
		return fmt.Errorf("unexpected sequence :e%c", s)
	}

	create := consumeBytes(state.r, []byte("!"))

	_ = create
	switch s {
	case 'e':
	case 'g':
	case 'n':
	case 'x':
	default:
		panic("unreachable")
	}
	return nil
}
func runFetch(state *replState) error { return nil }
func runHelp(state *replState) error  { return nil }
func runQuit(state *replState) error {
	state.r.Discard(1) // 'q'

	force := consumeBytes(state.r, []byte("!"))

	if state.pendingEdits && !force {
		return fmt.Errorf("write pending edits with :w first or use :q! to force")
	}
	state.quit = true
	return nil
}
func runWrite(state *replState) error {
	state.r.Discard(1) // 'w'
	state.pendingEdits = false
	b, err := state.r.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	if b[0] == 'q' {
		state.quit = true
	}
	return nil
}
func runDiff(state *replState) error {
	var prev, curr []byte
	if len(state.undoStack) > 0 {
		prevGraph := state.undoStack[len(state.undoStack)-1]
		var err error
		if prev, err = json.MarshalIndent(prevGraph, "", "  "); err != nil {
			return err
		}
	}
	currGraph := state.b.Build()
	curr, err := json.MarshalIndent(currGraph, "", "  ")
	if err != nil {
		return err
	}
	diff := cmp.Diff(json.RawMessage(prev), json.RawMessage(curr))
	if diff != "" {
		log.Println(diff)
	}
	return nil
}

func prompt(rl *readline.Instance, message string, allowEmpty bool) (string, error) {
	rl.Operation.SetPrompt(fmt.Sprint(message, ":"))
	defer func() { rl.Operation.SetPrompt(":") }()
	for {
		text, err := rl.Readline()
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text == "" && !allowEmpty {
			continue
		}
		return text, nil
	}
}

func consumeBytes(r *bufio.Reader, b []byte) bool {
	if bs, err := r.Peek(len(b)); err == nil && bytes.Equal(b, bs) {
		r.Discard(len(b))
		return true
	}
	return false
}

func consumeAny(r *bufio.Reader, chars string) (rune, bool) {
	v, _, err := r.ReadRune()
	if err == nil && strings.ContainsAny(string(v), chars) {
		return v, true
	}
	r.UnreadRune()
	return v, false
}

func consumeAnyWithDefault(r *bufio.Reader, chars string, defaultVal rune) (rune, bool) {
	v, _, err := r.ReadRune()
	if v == 0 {
		return defaultVal, true
	}
	if err == nil && strings.ContainsAny(string(v), chars) {
		return v, true
	}
	r.UnreadRune()
	return v, false
}

func consumeDelim(r *bufio.Reader, delim byte, allowEOF bool) ([]byte, error) {
	token, err := r.ReadBytes(delim)
	if err != nil && !(err == io.EOF && allowEOF) {
		return nil, err
	}
	return token, nil
}

func init() {
	fs := replCmd.Flags()
	fs.StringVarP(&replFlags.Graph, "graph", "g", "", "Graph to edit from a local file")
	rootCmd.AddCommand(replCmd)
}
