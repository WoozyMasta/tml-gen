package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
)

// Options defines CLI arguments.
type Options struct {
	GameRoot  string   `short:"g" long:"game-root" required:"true" description:"Game root directory (absolute)"`
	Out       string   `short:"o" long:"out" default:"out" description:"Output dir"`
	Paths     []string `short:"p" long:"path" required:"true" description:"Path to scan: relative to game-root OR absolute inside game-root (repeatable)"`
	Skip      []string `short:"s" long:"skip" default:"animals" default:"characters" default:"data" default:"gear" default:"vehicles" default:"weapons" description:"Skip path prefixes (repeatable)"`
	Threshold int      `short:"n" long:"threshold" default:"75" description:"Min objects per library"`
	Force     bool     `short:"f" long:"force" description:"Delete output directory before writing"`
	Version   bool     `short:"v" long:"version" description:"Show version"`
}

func main() {
	var opt Options
	p := flags.NewParser(&opt, flags.Default|flags.PassDoubleDash)
	p.ShortDescription = "Template Library generator for TerrainBuilder (DayZ/Arma 3)."
	p.LongDescription = `Generates *.tml Template Libraries by scanning P:/ (or any game root).
Designed to automate DayZ/Arma 3 map setup and keep libraries easy to refresh.

Key behavior:
- Groups files by directory nodes with --threshold, but never bubbles up to the top-level group.
- Keeps <File> paths exactly as scanned (relative to game-root, original casing).
- Ensures global-unique <Name> across all libraries; only modifies on duplicates.
- Supports --skip prefix rules (relative to scan-root or game-root) to exclude subtrees.
- Auto colors and shapes libraries based on their type; unknown types use a hash color.`

	_, err := p.Parse()

	if opt.Version {
		PrintVersion()
		os.Exit(0)
	}

	if err != nil {
		if err == flags.ErrHelp {
			os.Exit(0)
		}
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if opt.Threshold <= 0 {
		fmt.Fprintln(os.Stderr, "threshold must be > 0")
		os.Exit(2)
	}

	// Normalize important paths upfront.
	opt.GameRoot = cleanAbs(opt.GameRoot)
	opt.Out = cleanAbs(opt.Out)

	// Prepare output directory (create or clean).
	prepareOut(opt.Out, opt.Force)

	if opt.GameRoot == "" {
		fmt.Fprintln(os.Stderr, "game-root is required")
		os.Exit(2)
	}
	if info, err := os.Stat(opt.GameRoot); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "bad game-root: %s\n", opt.GameRoot)
		os.Exit(2)
	}

	// Build normalized skip prefixes for fast matching during traversal.
	skipPrefixes, err := buildSkipPrefixes(opt.GameRoot, opt.Skip)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// Normalize scan roots to absolute under game-root
	scanRoots := make([]string, 0, len(opt.Paths))
	for _, p := range opt.Paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var abs string
		if filepath.IsAbs(p) {
			abs = cleanAbs(p)
			if !startsWithPathPrefix(abs, opt.GameRoot) {
				fmt.Fprintf(os.Stderr, "path not under game-root: %s\n", p)
				os.Exit(2)
			}
		} else {
			abs = cleanAbs(filepath.Join(opt.GameRoot, p))
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "bad scan path: %s\n", abs)
			os.Exit(2)
		}
		scanRoots = append(scanRoots, abs)
	}
	if len(scanRoots) == 0 {
		fmt.Fprintln(os.Stderr, "no valid --path provided")
		os.Exit(2)
	}

	// Build directory tree to compute grouping by threshold.
	tree := newNode("", nil)

	type item struct{ rel string }
	ch := make(chan item, 8192)

	var mu sync.Mutex
	recs := make([]Rec, 0, 20000)

	workers := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for it := range ch {
				// Insert each file path into the tree and record its node.
				segs := splitSegs(it.rel)
				if len(segs) == 0 {
					continue
				}

				dirSegs := segs[:len(segs)-1]
				mu.Lock()
				dirNode := insert(tree, dirSegs)
				recs = append(recs, Rec{RelPath: filepath.ToSlash(it.rel), DirNode: dirNode})
				mu.Unlock()
			}
		}()
	}

	for _, scanRoot := range scanRoots {
		if err := filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return err
			}

			// Compute both relative paths: to game-root and to the scan-root.
			relGame, err := filepath.Rel(opt.GameRoot, path)
			if err != nil {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			relScan, err := filepath.Rel(scanRoot, path)
			if err != nil {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			relGameTrimmed := ""
			if segs := splitSegs(relGame); len(segs) > 1 {
				relGameTrimmed = strings.Join(segs[1:], "/")
			}
			if matchSkip(relScan, skipPrefixes) || matchSkip(relGame, skipPrefixes) || matchSkip(relGameTrimmed, skipPrefixes) {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			// Only enqueue .p3d files.
			if d.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Ext(d.Name()), ".p3d") {
				return nil
			}

			ch <- item{rel: filepath.ToSlash(relGame)}
			return nil
		}); err != nil {
			fmt.Fprintln(os.Stderr, "walk error:", err)
			os.Exit(1)
		}
	}

	close(ch)
	wg.Wait()

	if len(recs) == 0 {
		fmt.Fprintln(os.Stderr, "no .p3d found")
		os.Exit(1)
	}

	// Group files by threshold-based directory nodes.
	groups := make(map[string][]string)
	for _, r := range recs {
		g := pickGroup(r.DirNode, opt.Threshold)
		key := nodeKey(g)
		groups[key] = append(groups[key], r.RelPath)
	}

	// Sort groups by key.
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Track unique names across all libraries.
	usedNames := make(map[string]struct{}, 4096)
	for _, k := range keys {
		lst := groups[k]
		sort.Strings(lst)

		// Normalize output naming to lowercase for files and library names.
		libName := strings.ToLower(k)
		outPath := filepath.Join(opt.Out, libName+".tml")

		// Color is derived from the (possibly mixed-case) logical library name.
		fill, outline := colorForLibrary(k)
		shape := shapeForLibrary(k)
		if err := writeTML(outPath, libName, lst, fill, outline, shape, usedNames); err != nil {
			fmt.Fprintln(os.Stderr, "write tml error:", err)
			os.Exit(1)
		}
	}

	fmt.Printf("game_root=%s p3d=%d groups=%d threshold=%d out=%s\n", opt.GameRoot, len(recs), len(groups), opt.Threshold, opt.Out)
}
