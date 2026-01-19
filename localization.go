package main

// Localization strings loaded from american.txt, populated at runtime
var localizationMap = map[string]string{}

// LocalizeText attempts to localize a text key, returns original if not found
func LocalizeText(key string) string {
	if localized, ok := localizationMap[key]; ok {
		return localized
	}
	return key
}
