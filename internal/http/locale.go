package http

import "liki/internal/i18n"

// langToLocale maps frontend language code to BCP 47 locale.
// Short codes "zh" and "hk" are legacy aliases for zh-Hans and zh-Hant.
func langToLocale(lang string) string {
	switch lang {
	case "zh", string(i18n.LangHans):
		return string(i18n.LangHans)
	case "hk", string(i18n.LangHant):
		return string(i18n.LangHant)
	case string(i18n.EN):
		return string(i18n.EN)
	default:
		return string(i18n.LangHans)
	}
}
