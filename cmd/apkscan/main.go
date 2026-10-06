package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"apkscan/internal/detector"
	"apkscan/internal/dex"
	"apkscan/internal/extractor"
	"apkscan/internal/manifest"
	"apkscan/internal/output"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s -h' for help.\n", appCmd)
		return 2
	}
	if opts.mode == modeHelp {
		printHelp(os.Stdout)
		return 0
	}

	apkPaths, err := resolveInputs(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	printBanner(os.Stderr) 

	if opts.mode == modeManifest {
		return runManifest(apkPaths[0])
	}
	return runScan(apkPaths)
}

func runScan(apkPaths []string) int {
	result, err := scan(apkPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return printJSON(output.FromResult(strings.Join(apkPaths, " + "), result))
}

func runManifest(path string) int {
	rep, err := analyzeManifest(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return printJSON(rep)
}

func printJSON(v any) int {
	if err := output.PrintJSON(v); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func withPathHint(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w (if the path contains spaces, wrap it in quotes)", err)
	}
	return err
}


func checkFile(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		return withPathHint(err)
	}
	if st.IsDir() {
		return fmt.Errorf("%s is a folder, not a file. Use -d to scan a folder", p)
	}
	return nil
}

func resolveInputs(opts options) ([]string, error) {
	switch opts.mode {
	case modeSingle, modeMulti, modeManifest:
		for _, p := range opts.paths {
			if err := checkFile(p); err != nil {
				return nil, err
			}
		}
		return opts.paths, nil

	case modeDir:
		dir := strings.TrimRight(opts.paths[0], `"`)
		st, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("cannot access folder: %w", withPathHint(err))
		}
		if !st.IsDir() {
			return nil, fmt.Errorf("%s is a file, not a folder. Use -f to scan a single APK", dir)
		}
		apks, err := findAPKsInDir(dir)
		if err != nil {
			return nil, err
		}
		if len(apks) == 0 {
			return nil, fmt.Errorf("no .apk files found in %s", dir)
		}
		return apks, nil
	}
	return nil, errors.New("internal error: unknown mode")
}

func analyzeManifest(path string) (output.ManifestReport, error) {
	info, err := manifest.Parse(path)
	if err != nil {
		return output.ManifestReport{}, fmt.Errorf("could not read the manifest of %s: %w", path, err)
	}
	return output.ManifestReport{
		APKPath:  path,
		Warnings: manifestWarnings(info),
		Manifest: info.Summary(),
	}, nil
}

func manifestWarnings(info *manifest.Info) []string {
	warnings := append([]string(nil), info.Warnings...)
	if info.Split != "" || !info.HasApplication {
		warnings = append(warnings, "This looks like a split APK without the app's own manifest settings; analyze the base APK (usually base.apk) instead.")
	}
	return warnings
}

func scan(apkPaths []string) (detector.Result, error) {
	prefixes := detector.ClassPrefixes()
	classHits := map[string]string{}
	var scanWarnings []string
	dexUnreadable := false
	var allNames []string

	for _, p := range apkPaths {
		info, err := extractor.Open(p)
		if err != nil {
			return detector.Result{}, fmt.Errorf("could not open %s: %w", p, withPathHint(err))
		}
		allNames = append(allNames, info.Names()...)

		err = info.ReadDexFiles(func(name string, data []byte, readErr error) {
			label := filepath.Base(p) + ":" + name
			if readErr != nil {
				dexUnreadable = true
				scanWarnings = append(scanWarnings, fmt.Sprintf("Could not read %s (%v); class-based signals were skipped for it.", label, readErr))
				return
			}
			if err := dex.CollectPrefixHits(data, prefixes, classHits); err != nil {
				dexUnreadable = true
				scanWarnings = append(scanWarnings, fmt.Sprintf("Could not parse %s as DEX (%v). The file may be packed, encrypted or corrupted, so class-based signals were skipped for it.", label, err))
			}
		})
		if err != nil {
			return detector.Result{}, fmt.Errorf("could not scan %s: %w", p, err)
		}
	}

	var manifestNames []string
	if dexUnreadable {
		for _, p := range apkPaths {
			if mi, err := manifest.Parse(p); err == nil {
				manifestNames = append(manifestNames, mi.ClassNames()...)
			}
		}
	}

	return detector.Detect(detector.Evidence{
		Files:         allNames,
		ClassHits:     classHits,
		ManifestNames: manifestNames,
		DexUnreadable: dexUnreadable,
		ScanWarnings:  scanWarnings,
	}), nil
}

func findAPKsInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var apks []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".apk") {
			apks = append(apks, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(apks)
	return apks, nil
}
