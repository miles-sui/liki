package i18n

import (
	"net/http"
	"strings"
)

type Lang string

const (
	LangHans Lang = "zh-Hans"
	LangHant Lang = "zh-Hant"
	EN      Lang = "en"
)

var messages = map[string]map[Lang]string{
	"err.order_not_found":           {LangHans: "订单未找到", LangHant: "訂單未找到", EN: "Order not found"},
	"err.webhook_signature_invalid": {LangHans: "webhook 签名验证失败", LangHant: "webhook 簽名驗證失敗", EN: "Webhook signature verification failed"},
	"err.missing_order_id":          {LangHans: "缺少订单号", LangHant: "缺少訂單號", EN: "Missing order ID"},
	"err.order_not_paid":            {LangHans: "订单未支付", LangHant: "訂單未支付", EN: "Order not paid"},
	"err.body_parse":                {LangHans: "无法解析请求", LangHant: "無法解析請求", EN: "Unable to parse request"},
}

func T(lang Lang, key string) string {
	if m, ok := messages[key]; ok {
		if s, ok := m[lang]; ok {
			return s
		}
		if s, ok := m[LangHant]; ok {
			return s
		}
	}
	return key
}

func DetectLang(r *http.Request) Lang {
	h := r.Header.Get("Accept-Language")
	if h == "" {
		return LangHant
	}
	first := strings.SplitN(h, ",", 2)[0]
	first = strings.SplitN(first, ";", 2)[0]
	first = strings.TrimSpace(first)
	switch first {
	case "zh-CN", "zh-SG", "zh-Hans":
		return LangHans
	default:
		if strings.HasPrefix(first, "zh") {
			return LangHant
		}
		return EN
	}
}
