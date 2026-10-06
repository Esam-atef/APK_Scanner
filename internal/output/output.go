package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Esam-atef/APK_Scanner/internal/detector"
	"github.com/Esam-atef/APK_Scanner/internal/manifest"
)

type Report struct {
	APKPath                string                     `json:"apk_path"`
	Framework              detector.Framework         `json:"framework"`
	Confidence             int                        `json:"confidence"`
	Source                 string                     `json:"source"`
	SignalsMatched         []string                   `json:"signals_matched"`
	OtherCandidates        map[detector.Framework]int `json:"other_candidates,omitempty"`
	UnrecognizedNativeLibs []string                   `json:"unrecognized_native_libs,omitempty"`
	Warnings               []string                   `json:"warnings,omitempty"`
}

type ManifestReport struct {
	APKPath  string            `json:"apk_path"`
	Warnings []string          `json:"warnings,omitempty"`
	Manifest *manifest.Summary `json:"manifest"`
}

func FromResult(apkPath string, r detector.Result) Report {
	return Report{
		APKPath:                apkPath,
		Framework:              r.Framework,
		Confidence:             r.Confidence,
		Source:                 r.Source,
		SignalsMatched:         r.SignalsMatched,
		OtherCandidates:        r.OtherCandidates,
		UnrecognizedNativeLibs: r.UnrecognizedNativeLibs,
		Warnings:               r.Warnings,
	}
}

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to build JSON output: %w", err)
	}
	return nil
}