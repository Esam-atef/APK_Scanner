package detector

var Rules = map[Framework][]Signal{
	Flutter: {
		{Name: "lib/*/libflutter.so", Pattern: "libflutter.so", MatchType: MatchSuffix, Weight: 60},
		{Name: "assets/flutter_assets/", Pattern: "assets/flutter_assets/", MatchType: MatchContains, Weight: 35},
		{Name: "lib/*/libapp.so (Flutter AOT snapshot)", Pattern: "libapp.so", MatchType: MatchSuffix, Weight: 5},
		{Name: "class io.flutter.*", Pattern: "Lio/flutter/", MatchType: MatchClassPrefix, Weight: 40},
	},
	ReactNative: {
		{Name: "assets/index.android.bundle", Pattern: "assets/index.android.bundle", MatchType: MatchExact, Weight: 55},
		{Name: "lib/*/libreactnativejni.so", Pattern: "libreactnativejni.so", MatchType: MatchSuffix, Weight: 30},
		{Name: "lib/*/libhermes.so", Pattern: "libhermes.so", MatchType: MatchSuffix, Weight: 15},
		{Name: "class com.facebook.react.*", Pattern: "Lcom/facebook/react/", MatchType: MatchClassPrefix, Weight: 40},
	},
	Xamarin: {
		{Name: "lib/*/libmonodroid.so", Pattern: "libmonodroid.so", MatchType: MatchSuffix, Weight: 40},
		{Name: "lib/*/libxamarin-app.so", Pattern: "libxamarin-app.so", MatchType: MatchSuffix, Weight: 20},
		{Name: "lib/*/libmonosgen-2.0.so", Pattern: "libmonosgen-2.0.so", MatchType: MatchSuffix, Weight: 20},
		{Name: "assemblies/", Pattern: "assemblies/", MatchType: MatchPrefix, Weight: 20},
		{Name: "class mono.*", Pattern: "Lmono/", MatchType: MatchClassPrefix, Weight: 30},
	},
	Unity: {
		{Name: "lib/*/libunity.so", Pattern: "libunity.so", MatchType: MatchSuffix, Weight: 60},
		{Name: "assets/bin/Data/", Pattern: "assets/bin/Data/", MatchType: MatchPrefix, Weight: 30},
		{Name: "lib/*/libil2cpp.so (IL2CPP backend)", Pattern: "libil2cpp.so", MatchType: MatchSuffix, Weight: 10},
		{Name: "class com.unity3d.player.*", Pattern: "Lcom/unity3d/player/", MatchType: MatchClassPrefix, Weight: 40},
	},
	Cordova: {
		{Name: "assets/www/cordova.js", Pattern: "assets/www/cordova.js", MatchType: MatchExact, Weight: 60},
		{Name: "assets/www/index.html", Pattern: "assets/www/index.html", MatchType: MatchExact, Weight: 40},
		{Name: "class org.apache.cordova.*", Pattern: "Lorg/apache/cordova/", MatchType: MatchClassPrefix, Weight: 40},
	},
	Capacitor: {
		{Name: "assets/capacitor.config.json", Pattern: "assets/capacitor.config.json", MatchType: MatchExact, Weight: 40},
		{Name: "assets/public/index.html", Pattern: "assets/public/index.html", MatchType: MatchExact, Weight: 40},
		{Name: "assets/capacitor.plugins.json", Pattern: "assets/capacitor.plugins.json", MatchType: MatchExact, Weight: 20},
		{Name: "class com.getcapacitor.*", Pattern: "Lcom/getcapacitor/", MatchType: MatchClassPrefix, Weight: 40},
	},
	NativeScript: {
		{Name: "lib/*/libNativeScript.so", Pattern: "libNativeScript.so", MatchType: MatchSuffix, Weight: 70},
		{Name: "assets/metadata/", Pattern: "assets/metadata/", MatchType: MatchPrefix, Weight: 30},
		{Name: "class com.tns.*", Pattern: "Lcom/tns/", MatchType: MatchClassPrefix, Weight: 40},
	},
	Godot: {
		{Name: "lib/*/libgodot_android.so", Pattern: "libgodot_android.so", MatchType: MatchSuffix, Weight: 70},
		{Name: "*.pck (Godot resource pack)", Pattern: ".pck", MatchType: MatchSuffix, Weight: 30},
		{Name: "class org.godotengine.godot.*", Pattern: "Lorg/godotengine/godot/", MatchType: MatchClassPrefix, Weight: 40},
	},
}
