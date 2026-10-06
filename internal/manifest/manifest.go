package manifest

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/avast/apkparser"
)

type Component struct {
	Type            string 
	Name            string 
	Exported        *bool  
	Enabled         *bool
	Permission      string 
	HasIntentFilter bool
	Launcher        bool    
	Actions         []string 
}


func (c Component) IsExported() bool {
	if c.Enabled != nil && !*c.Enabled {
		return false
	}
	if c.Exported != nil {
		return *c.Exported
	}
	if c.Type == "provider" {
		return false
	}
	return c.HasIntentFilter
}


func (c Component) ExportReason() string {
	if c.Exported != nil && *c.Exported {
		return "exported=true"
	}
	return "intent-filter"
}

type Info struct {
	Package     string
	Split       string 
	VersionName string
	VersionCode int64
	MinSDK      int
	TargetSDK   int

	HasApplication        bool
	ApplicationName       string
	Debuggable            *bool
	AllowBackup           *bool
	UsesCleartextTraffic  *bool
	NetworkSecurityConfig bool

	Permissions []string
	Components  []Component
	MetaData    []string 

	Warnings []string
}


type collector struct {
	info *Info

	cur            int 
	inFilter       bool
	sawMain        bool
	sawLauncherCat bool
}

func newCollector() *collector {
	return &collector{info: &Info{}, cur: -1}
}

func (c *collector) EncodeToken(t xml.Token) error {
	switch tok := t.(type) {
	case xml.StartElement:
		c.start(tok)
	case xml.EndElement:
		c.end(tok)
	}
	return nil
}

func (c *collector) Flush() error { return nil }

func attrs(el xml.StartElement) map[string]string {
	m := make(map[string]string, len(el.Attr))
	for _, a := range el.Attr {
		m[a.Name.Local] = a.Value
	}
	return m
}

func boolAttr(a map[string]string, key string) *bool {
	switch a[key] {
	case "true":
		v := true
		return &v
	case "false":
		v := false
		return &v
	}
	return nil
}

func parseInt(s string) int64 {
	n, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return 0 
	}
	return n
}

func resolveClass(pkg, name string) string {
	switch {
	case name == "":
		return ""
	case strings.HasPrefix(name, "."):
		return pkg + name
	case !strings.Contains(name, "."):
		return pkg + "." + name
	}
	return name
}

func isComponentTag(name string) bool {
	switch name {
	case "activity", "activity-alias", "service", "receiver", "provider":
		return true
	}
	return false
}

func (c *collector) start(el xml.StartElement) {
	a := attrs(el)
	in := c.info

	switch name := el.Name.Local; {
	case name == "manifest":
		in.Package = a["package"]
		in.Split = a["split"]
		in.VersionName = a["versionName"]
		in.VersionCode = parseInt(a["versionCode"])

	case name == "uses-sdk":
		in.MinSDK = int(parseInt(a["minSdkVersion"]))
		in.TargetSDK = int(parseInt(a["targetSdkVersion"]))

	case name == "uses-permission" || name == "uses-permission-sdk-23":
		if p := a["name"]; p != "" {
			in.Permissions = append(in.Permissions, p)
		}

	case name == "application":
		in.HasApplication = true
		in.ApplicationName = resolveClass(in.Package, a["name"])
		in.Debuggable = boolAttr(a, "debuggable")
		in.AllowBackup = boolAttr(a, "allowBackup")
		in.UsesCleartextTraffic = boolAttr(a, "usesCleartextTraffic")
		in.NetworkSecurityConfig = a["networkSecurityConfig"] != ""

	case isComponentTag(name):
		perm := a["permission"]
		if perm == "" && a["readPermission"] != "" && a["writePermission"] != "" {
			perm = a["readPermission"] + " / " + a["writePermission"]
		}
		in.Components = append(in.Components, Component{
			Type:       name,
			Name:       resolveClass(in.Package, a["name"]),
			Exported:   boolAttr(a, "exported"),
			Enabled:    boolAttr(a, "enabled"),
			Permission: perm,
		})
		c.cur = len(in.Components) - 1

	case name == "intent-filter":
		if c.cur >= 0 {
			in.Components[c.cur].HasIntentFilter = true
			c.inFilter, c.sawMain, c.sawLauncherCat = true, false, false
		}

	case name == "action":
		if c.inFilter {
			act := a["name"]
			if act == "android.intent.action.MAIN" {
				c.sawMain = true
			}
			if act != "" && c.cur >= 0 {
				comp := &in.Components[c.cur]
				comp.Actions = append(comp.Actions, act)
			}
		}

	case name == "category":
		if c.inFilter && a["name"] == "android.intent.category.LAUNCHER" {
			c.sawLauncherCat = true
		}

	case name == "meta-data":
		if n := a["name"]; n != "" {
			in.MetaData = append(in.MetaData, n)
		}
	}
}

func (c *collector) end(el xml.EndElement) {
	name := el.Name.Local
	switch {
	case name == "intent-filter":
		if c.inFilter && c.cur >= 0 && c.sawMain && c.sawLauncherCat {
			c.info.Components[c.cur].Launcher = true
		}
		c.inFilter = false
	case isComponentTag(name):
		if c.cur >= 0 && c.info.Components[c.cur].Type == name {
			c.cur = -1
		}
	}
}

func Parse(apkPath string) (info *Info, err error) {
	defer func() {
		if r := recover(); r != nil {
			info, err = nil, fmt.Errorf("the manifest parser crashed on this file (%v)", r)
		}
	}()

	c := newCollector()
	zipErr, resErr, manErr := apkparser.ParseApk(apkPath, c)
	if zipErr != nil {
		return nil, fmt.Errorf("could not read the archive: %w", zipErr)
	}
	if manErr != nil {
		return nil, fmt.Errorf("could not parse AndroidManifest.xml: %w", manErr)
	}
	if resErr != nil && !errors.Is(resErr, os.ErrNotExist) {
		c.info.Warnings = append(c.info.Warnings,
			fmt.Sprintf("The resource table could not be parsed (%v); some manifest values may be unresolved.", resErr))
	}

	c.info.Permissions = sortedUnique(c.info.Permissions)
	return c.info, nil
}

func sortedUnique(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func (i *Info) ClassNames() []string {
	var names []string
	if i.ApplicationName != "" {
		names = append(names, i.ApplicationName)
	}
	for _, c := range i.Components {
		if c.Name != "" {
			names = append(names, c.Name)
		}
	}
	names = append(names, i.MetaData...)
	return sortedUnique(names)
}

func Pick(infos []*Info) *Info {
	for _, in := range infos {
		if in.Split == "" {
			return in
		}
	}
	if len(infos) > 0 {
		return infos[0]
	}
	return nil
}


func Packages(infos []*Info) []string {
	var pkgs []string
	for _, in := range infos {
		if in.Package != "" {
			pkgs = append(pkgs, in.Package)
		}
	}
	return sortedUnique(pkgs)
}
