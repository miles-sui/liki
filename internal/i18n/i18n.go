package i18n

import (
	"net/http"
	"strings"
)

type Lang string

const (
	ZH Lang = "zh"
	HK Lang = "hk"
	EN Lang = "en"
)

var messages = map[string]map[Lang]string{
	"err.order_not_found":           {ZH: "订单未找到", HK: "訂單未找到", EN: "Order not found"},
	"err.webhook_signature_invalid": {ZH: "webhook 签名验证失败", HK: "webhook 簽名驗證失敗", EN: "Webhook signature verification failed"},
	"err.missing_order_id":          {ZH: "缺少订单号", HK: "缺少訂單號", EN: "Missing order ID"},
	"err.order_not_paid":            {ZH: "订单未支付", HK: "訂單未支付", EN: "Order not paid"},
	"err.body_parse":                {ZH: "无法解析请求", HK: "無法解析請求", EN: "Unable to parse request"},
}

func T(lang Lang, key string) string {
	if m, ok := messages[key]; ok {
		if s, ok := m[lang]; ok {
			return s
		}
		if s, ok := m[HK]; ok {
			return s
		}
	}
	return key
}

func DetectLang(r *http.Request) Lang {
	h := r.Header.Get("Accept-Language")
	if h == "" {
		return HK
	}
	first := strings.SplitN(h, ",", 2)[0]
	first = strings.SplitN(first, ";", 2)[0]
	first = strings.TrimSpace(first)
	switch first {
	case "zh-CN", "zh-SG", "zh-Hans":
		return ZH
	default:
		if strings.HasPrefix(first, "zh") {
			return HK
		}
		return EN
	}
}
