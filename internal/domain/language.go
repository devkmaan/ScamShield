package domain

import "strings"

type LanguageOption struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
}

func SupportedLanguages() []LanguageOption {
	return []LanguageOption{
		{Code: "en", Name: "English", NativeName: "English"},
		{Code: "hinglish", Name: "Hinglish", NativeName: "Hinglish"},
		{Code: "hi", Name: "Hindi", NativeName: "Hindi"},
		{Code: "bn", Name: "Bengali", NativeName: "Bangla"},
		{Code: "ta", Name: "Tamil", NativeName: "Tamil"},
		{Code: "te", Name: "Telugu", NativeName: "Telugu"},
		{Code: "mr", Name: "Marathi", NativeName: "Marathi"},
		{Code: "gu", Name: "Gujarati", NativeName: "Gujarati"},
		{Code: "kn", Name: "Kannada", NativeName: "Kannada"},
		{Code: "ml", Name: "Malayalam", NativeName: "Malayalam"},
		{Code: "pa", Name: "Punjabi", NativeName: "Punjabi"},
		{Code: "ur", Name: "Urdu", NativeName: "Urdu"},
	}
}

func NormalizeLanguage(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, "_", "-")
	aliases := map[string]string{
		"":          "en",
		"english":   "en",
		"eng":       "en",
		"hindi":     "hi",
		"hin":       "hi",
		"bangla":    "bn",
		"bengali":   "bn",
		"tamil":     "ta",
		"telugu":    "te",
		"marathi":   "mr",
		"gujarati":  "gu",
		"kannada":   "kn",
		"malayalam": "ml",
		"punjabi":   "pa",
		"urdu":      "ur",
	}
	if mapped, ok := aliases[value]; ok {
		return mapped
	}
	if IsSupportedLanguage(value) {
		return value
	}
	return "en"
}

func IsSupportedLanguage(raw string) bool {
	value := strings.ToLower(strings.TrimSpace(raw))
	for _, language := range SupportedLanguages() {
		if language.Code == value {
			return true
		}
	}
	return false
}

func LanguageName(code string) string {
	normalized := NormalizeLanguage(code)
	for _, language := range SupportedLanguages() {
		if language.Code == normalized {
			return language.Name
		}
	}
	return "English"
}
