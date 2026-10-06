package detector

type Framework string

const (
	Flutter      Framework = "Flutter"
	ReactNative  Framework = "React Native"
	Xamarin      Framework = "Xamarin / .NET MAUI"
	Unity        Framework = "Unity"
	Cordova      Framework = "Cordova"
	Capacitor    Framework = "Capacitor"
	NativeScript Framework = "NativeScript"
	Godot        Framework = "Godot"
	Native       Framework = "Native"
	Unknown      Framework = "Unknown"
)

type MatchType int

const (
	
	MatchSuffix MatchType = iota
	MatchExact
	MatchContains
	MatchPrefix
	MatchClassPrefix
)

type Signal struct {
	Name      string 
	Pattern   string
	MatchType MatchType
	Weight    int
}

type Evidence struct {
	Files []string 
	ClassHits map[string]string
	ManifestNames []string
	DexUnreadable bool
	ScanWarnings []string
}

type Result struct {
	Framework      Framework `json:"framework"`
	Confidence     int       `json:"confidence"` 
	Source         string    `json:"source"`    
	SignalsMatched []string  `json:"signals_matched"`

	AllScores map[Framework]int `json:"all_scores"`
	OtherCandidates map[Framework]int `json:"other_candidates,omitempty"`
	UnrecognizedNativeLibs []string `json:"unrecognized_native_libs,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
