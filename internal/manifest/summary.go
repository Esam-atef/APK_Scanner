package manifest

import "fmt"

type Summary struct {
	Package          string `json:"package"`
	Split            string `json:"split,omitempty"`
	VersionName      string `json:"version_name,omitempty"`
	VersionCode      int64  `json:"version_code,omitempty"`
	MinSDK           int    `json:"min_sdk,omitempty"`
	TargetSDK        int    `json:"target_sdk,omitempty"`
	ApplicationClass string `json:"application_class,omitempty"`

	Permissions []string `json:"permissions"`
	Debuggable *bool `json:"debuggable,omitempty"`
	AllowBackup *bool `json:"allow_backup,omitempty"`
	ExportedComponents []ExportedComponent `json:"exported_components"`
	ComponentCounts    map[string]int      `json:"component_counts"`

	UsesCleartextTraffic  *bool `json:"uses_cleartext_traffic,omitempty"`
	NetworkSecurityConfig bool  `json:"network_security_config"`

	Findings []Finding `json:"findings"`
}

type ExportedComponent struct {
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Reason     string   `json:"reason"` 
	Permission string   `json:"permission,omitempty"`
	Launcher   bool     `json:"launcher,omitempty"`
	Actions    []string `json:"actions,omitempty"`
}

type Finding struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Note string `json:"note,omitempty"`
}

const (
	noteDebuggable = "Reportable in a pentest; not usually accepted in bug bounty programs."
	noteBackup     = "Bug bounty programs sometimes accept this and sometimes do not."
)

func (i *Info) Summary() *Summary {
	s := &Summary{
		Package:               i.Package,
		Split:                 i.Split,
		VersionName:           i.VersionName,
		VersionCode:           i.VersionCode,
		MinSDK:                i.MinSDK,
		TargetSDK:             i.TargetSDK,
		ApplicationClass:      i.ApplicationName,
		Permissions:           i.Permissions,
		Debuggable:            i.Debuggable,
		AllowBackup:           i.AllowBackup,
		ExportedComponents:    []ExportedComponent{},
		ComponentCounts:       map[string]int{},
		UsesCleartextTraffic:  i.UsesCleartextTraffic,
		NetworkSecurityConfig: i.NetworkSecurityConfig,
		Findings:              i.Findings(),
	}
	if s.Permissions == nil {
		s.Permissions = []string{}
	}
	for _, c := range i.Components {
		s.ComponentCounts[c.Type]++
		if c.IsExported() {
			s.ExportedComponents = append(s.ExportedComponents, ExportedComponent{
				Type:       c.Type,
				Name:       c.Name,
				Reason:     c.ExportReason(),
				Permission: c.Permission,
				Launcher:   c.Launcher,
				Actions:    sortedUnique(c.Actions),
			})
		}
	}
	return s
}


func (i *Info) Findings() []Finding {
	findings := []Finding{}
	if !i.HasApplication {
		return findings
	}

	if i.Debuggable != nil && *i.Debuggable {
		findings = append(findings, Finding{
			ID: "debuggable",
			Message: "android:debuggable is true. In a release build this is a vulnerability: a debugger can be attached " +
				"to the app and its data can be read. This tool cannot tell a debug build from a release build.",
			Note: noteDebuggable,
		})
	}

	switch {
	case i.AllowBackup == nil:
		findings = append(findings, Finding{
			ID:      "allow_backup",
			Message: "android:allowBackup is not set, which defaults to true: the app's data can be backed up and may be extracted over USB/ADB.",
			Note:    noteBackup,
		})
	case *i.AllowBackup:
		findings = append(findings, Finding{
			ID:      "allow_backup",
			Message: "android:allowBackup is true: the app's data can be backed up and may be extracted over USB/ADB.",
			Note:    noteBackup,
		})
	}

	unprotected := 0
	for _, c := range i.Components {
		if c.IsExported() && !c.Launcher && c.Permission == "" {
			unprotected++
		}
	}
	if unprotected > 0 {
		findings = append(findings, Finding{
			ID: "exported_components",
			Message: fmt.Sprintf("%d exported component(s) have no permission protection (launcher activities excluded). "+
				"A component is exported when it sets exported=true or has an intent-filter; see exported_components.", unprotected),
		})
	}

	return findings
}
