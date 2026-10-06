package detector

import (
	"sort"
	"strings"
)

const (
	minFrameworkConfidence = 20
	nativeNoLibsConfidence        = 70
	nativeWithLibsConfidence      = 50
	nativeUnreadableDexConfidence = 40
)

func matches(path string, s Signal) bool {
	switch s.MatchType {
	case MatchExact:
		return path == s.Pattern
	case MatchSuffix:
		return strings.HasSuffix(path, s.Pattern)
	case MatchContains:
		return strings.Contains(path, s.Pattern)
	case MatchPrefix:
		return strings.HasPrefix(path, s.Pattern)
	default:
		return false
	}
}

func findMatch(files []string, s Signal) (string, bool) {
	for _, f := range files {
		if matches(f, s) {
			return f, true
		}
	}
	return "", false
}

func containsFile(files []string, name string) bool {
	for _, f := range files {
		if f == name {
			return true
		}
	}
	return false
}

func hasPrefixFile(files []string, prefix string) bool {
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

func baseName(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func descriptorToClassName(desc string) string {
	desc = strings.TrimPrefix(desc, "L")
	desc = strings.TrimSuffix(desc, ";")
	return strings.ReplaceAll(desc, "/", ".")
}


func findClass(ev Evidence, prefix string) (name, source string, ok bool) {
	if desc, found := ev.ClassHits[prefix]; found {
		return descriptorToClassName(desc), "dex", true
	}
	dotted := descriptorToClassName(prefix)
	for _, n := range ev.ManifestNames {
		if strings.HasPrefix(n, dotted) {
			return n, "manifest", true
		}
	}
	return "", "", false
}

func ClassPrefixes() []string {
	seen := map[string]bool{}
	var prefixes []string
	for _, fw := range sortedFrameworks() {
		for _, s := range Rules[fw] {
			if s.MatchType == MatchClassPrefix && !seen[s.Pattern] {
				seen[s.Pattern] = true
				prefixes = append(prefixes, s.Pattern)
			}
		}
	}
	return prefixes
}

func isNativeLibSignal(s Signal) bool {
	return s.MatchType == MatchSuffix && strings.HasSuffix(s.Pattern, ".so")
}

func isKnownLib(name string) bool {
	for _, signals := range Rules {
		for _, s := range signals {
			if isNativeLibSignal(s) && name == s.Pattern {
				return true
			}
		}
	}
	return false
}


func expectsNativeLibs(fw Framework) bool {
	for _, s := range Rules[fw] {
		if isNativeLibSignal(s) {
			return true
		}
	}
	return false
}

func unrecognizedNativeLibs(files []string) []string {
	seen := map[string]bool{}
	var libs []string
	for _, f := range files {
		if !strings.HasPrefix(f, "lib/") || !strings.HasSuffix(f, ".so") {
			continue
		}
		name := baseName(f)
		if seen[name] || isKnownLib(name) {
			continue
		}
		seen[name] = true
		libs = append(libs, name)
	}
	sort.Strings(libs)
	return libs
}


func sortedFrameworks() []Framework {
	fws := make([]Framework, 0, len(Rules))
	for fw := range Rules {
		fws = append(fws, fw)
	}
	sort.Slice(fws, func(i, j int) bool { return fws[i] < fws[j] })
	return fws
}

func Detect(ev Evidence) Result {
	files := ev.Files
	rawScores := map[Framework]int{}
	maxScores := map[Framework]int{}
	matchedSignals := map[Framework][]string{}
	frameworks := sortedFrameworks()

	for _, fw := range frameworks {
		for _, sig := range Rules[fw] {
			if sig.MatchType == MatchClassPrefix {
				if name, source, ok := findClass(ev, sig.Pattern); ok {
					rawScores[fw] += sig.Weight
					matchedSignals[fw] = append(matchedSignals[fw], sig.Name+" → "+name+" ("+source+")")
				}
				continue
			}
			maxScores[fw] += sig.Weight
			if matchedPath, ok := findMatch(files, sig); ok {
				rawScores[fw] += sig.Weight
				matchedSignals[fw] = append(matchedSignals[fw], sig.Name+" → "+matchedPath)
			}
		}
	}

	confidences := map[Framework]int{}
	for _, fw := range frameworks {
		if maxScores[fw] == 0 {
			continue
		}
		c := (rawScores[fw] * 100) / maxScores[fw]
		if c > 100 {
			c = 100
		}
		confidences[fw] = c
	}

	best := Unknown
	bestConf := 0
	for _, fw := range frameworks {
		if confidences[fw] > bestConf {
			best = fw
			bestConf = confidences[fw]
		}
	}

	warnings := append([]string(nil), ev.ScanWarnings...)
	var unrecognized []string
	hasDex := containsFile(files, "classes.dex")

	if bestConf < minFrameworkConfidence && hasDex {
		best = Native
		unrecognized = unrecognizedNativeLibs(files)
		if len(unrecognized) == 0 {
			bestConf = nativeNoLibsConfidence
			matchedSignals[Native] = []string{"classes.dex present; no native libraries and no known framework signals found"}
		} else {
			bestConf = nativeWithLibsConfidence
			matchedSignals[Native] = []string{"classes.dex present; no known framework signals found"}
			warnings = append(warnings, "Native libraries that match no known framework were found. This may be custom JNI code, or an engine/framework that has no rules yet.")
		}
		if ev.DexUnreadable && bestConf > nativeUnreadableDexConfidence {
			bestConf = nativeUnreadableDexConfidence
		}
	}

	if !hasDex {
		warnings = append(warnings, "No classes.dex found. This looks like a split or config APK without code; include the base APK for a reliable result.")
	}
	if expectsNativeLibs(best) && !hasPrefixFile(files, "lib/") {
		warnings = append(warnings, "No native libraries (lib/) were found. If this app is installed as split APKs, the libraries are usually in a separate split_config ABI APK; pass all split files (or use -dir) for a reliable result.")
	}

	others := map[Framework]int{}
	for _, fw := range frameworks {
		if fw != best && confidences[fw] > 0 {
			others[fw] = confidences[fw]
		}
	}

	signals := matchedSignals[best]
	if signals == nil {
		signals = []string{}
	}

	return Result{
		Framework:              best,
		Confidence:             bestConf,
		Source:                 "rule-engine",
		SignalsMatched:         signals,
		AllScores:              confidences,
		OtherCandidates:        others,
		UnrecognizedNativeLibs: unrecognized,
		Warnings:               warnings,
	}
}
