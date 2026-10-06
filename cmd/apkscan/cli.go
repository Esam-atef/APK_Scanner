package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	appName    = "APK Scanner"
	appVersion = "0.1.0"
	appTagline = "Detect the framework behind an Android APK"
	appCmd     = "apkscan"
)

type mode int

const (
	modeNone mode = iota
	modeSingle
	modeMulti
	modeDir
	modeManifest
	modeHelp
)

type options struct {
	mode  mode
	paths []string
}

const (
	kindNone     = ""
	kindUnknown  = "?"
	kindFile     = "f"
	kindMulti    = "m"
	kindDir      = "d"
	kindManifest = "a"
	kindHelp     = "h"
)

func classify(arg string) string {
	if len(arg) < 2 || arg[0] != '-' {
		return kindNone
	}
	name := arg[1:]
	name = strings.TrimPrefix(name, "-")
	switch name {
	case "f", "file":
		return kindFile
	case "m", "multi":
		return kindMulti
	case "d", "dir":
		return kindDir
	case "a", "manifest":
		return kindManifest
	case "h", "help":
		return kindHelp
	}
	return kindUnknown
}


func parseArgs(args []string) (options, error) {
	for _, a := range args {
		if classify(a) == kindHelp {
			return options{mode: modeHelp}, nil
		}
	}

	if len(args) == 0 {
		return options{}, errors.New("no input given. Use -f <apk>, -m <apk> <apk> ..., -d <folder>, or -a <apk>")
	}

	var opts options
	for i := 0; i < len(args); {
		kind := classify(args[i])
		switch kind {
		case kindNone:
			return options{}, fmt.Errorf("unexpected argument %q. Start with a flag: -f for one APK, -m for several, -d for a folder, -a for a manifest", args[i])
		case kindUnknown:
			return options{}, fmt.Errorf("unknown flag %q", args[i])
		}
		i++

		var values []string
		for i < len(args) && classify(args[i]) == kindNone {
			values = append(values, args[i])
			i++
		}

		if opts.mode != modeNone {
			return options{}, errors.New("only one of -f, -m, -d or -a can be used at a time")
		}

		const quoteTip = "If a path contains spaces, wrap it in quotes"
		switch kind {
		case kindFile:
			switch {
			case len(values) == 0:
				return options{}, errors.New("-f needs an APK path, e.g. -f app.apk")
			case len(values) > 1:
				return options{}, fmt.Errorf("-f takes a single APK but got %d values. Use -m to scan several APKs together. %s", len(values), quoteTip)
			}
			opts = options{mode: modeSingle, paths: values}

		case kindMulti:
			switch {
			case len(values) == 0:
				return options{}, errors.New("-m needs at least two APK paths, e.g. -m base.apk split_config.arm64_v8a.apk")
			case len(values) == 1:
				return options{}, errors.New("-m needs at least two APK files. Use -f to scan a single APK")
			}
			opts = options{mode: modeMulti, paths: values}

		case kindDir:
			switch {
			case len(values) == 0:
				return options{}, errors.New("-d needs a folder path, e.g. -d C:\\apks\\myapp")
			case len(values) > 1:
				return options{}, fmt.Errorf("-d takes a single folder but got %d values. %s", len(values), quoteTip)
			}
			opts = options{mode: modeDir, paths: values}

		case kindManifest:
			switch {
			case len(values) == 0:
				return options{}, errors.New("-a needs an APK path, e.g. -a app.apk")
			case len(values) > 1:
				return options{}, fmt.Errorf("-a takes a single APK but got %d values. %s", len(values), quoteTip)
			}
			opts = options{mode: modeManifest, paths: values}
		}
	}
	return opts, nil
}

func printBanner(w io.Writer) {
	lines := []string{appName + "  v" + appVersion, appTagline}
	width := 0
	for _, l := range lines {
		if len(l) > width {
			width = len(l)
		}
	}
	border := "+" + strings.Repeat("-", width+4) + "+"
	fmt.Fprintln(w, border)
	for _, l := range lines {
		fmt.Fprintf(w, "|  %-*s  |\n", width, l)
	}
	fmt.Fprintln(w, border)
	fmt.Fprintln(w)
}

type flagDoc struct {
	usage string
	desc  []string 
}

var flagDocs = []flagDoc{
	{"-f, --file <apk>", []string{"Scan a single APK file"}},
	{"-m, --multi <apk> <apk> ...", []string{
		"Scan several APK files together as one app",
		"(a base APK plus its split_config files)"}},
	{"-d, --dir <folder>", []string{
		"Scan every .apk file in a folder as one app",
		"(the folder should contain a single app)"}},
	{"-a, --manifest <apk>", []string{
		"Analyze the AndroidManifest.xml of an APK",
		"(permissions, debuggable, backup, exported components)"}},
	{"-h, --help", []string{"Show this help"}},
}

func printHelp(w io.Writer) {
	fmt.Fprintf(w, "%s detects which framework an Android app was built with (Flutter,\n", appName)
	fmt.Fprintln(w, "React Native, Unity, native Java/Kotlin, ...), or analyzes the app's manifest")
	fmt.Fprintln(w, "for security-relevant settings. The result is printed as JSON.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")

	width := 0
	for _, d := range flagDocs {
		if len(d.usage) > width {
			width = len(d.usage)
		}
	}
	for _, d := range flagDocs {
		for i, line := range d.desc {
			left := ""
			if i == 0 {
				left = d.usage
			}
			fmt.Fprintf(w, "  %-*s   %s\n", width, left, line)
		}
	}
}
