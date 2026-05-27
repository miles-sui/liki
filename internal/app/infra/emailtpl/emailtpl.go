package emailtpl

import "fmt"

const baseURL = "https://25types.com"

// VerificationEmail returns subject and body for email verification.
func VerificationEmail(locale, token string) (subject, text string) {
	link := fmt.Sprintf("%s/%s/verify-email?token=%s", baseURL, locale, token)
	if locale == "zh-CN" {
		return "验证你的 25types 邮箱",
			fmt.Sprintf("请点击以下链接验证你的邮箱：\n\n%s\n\n此链接 24 小时内有效。\n\n— 25types", link)
	}
	return "Verify your 25types email",
		fmt.Sprintf("Please verify your email by clicking the link below:\n\n%s\n\nThis link expires in 24 hours.\n\n— 25types", link)
}

// PasswordResetEmail returns subject and body for password reset.
func PasswordResetEmail(locale, token string) (subject, text string) {
	link := fmt.Sprintf("%s/%s/reset-password?token=%s", baseURL, locale, token)
	if locale == "zh-CN" {
		return "25types 密码重置",
			fmt.Sprintf("请点击以下链接重置密码：\n\n%s\n\n此链接 15 分钟内有效。如果你没有请求重置密码，请忽略此邮件。\n\n— 25types", link)
	}
	return "25types Password Reset",
		fmt.Sprintf("Reset your password by clicking the link below:\n\n%s\n\nThis link expires in 15 minutes. If you didn't request this, please ignore.\n\n— 25types", link)
}

// BondNotification returns subject and body for bond notification.
func BondNotification(locale, otherName, creatorName string) (subject, text string) {
	profileURL := fmt.Sprintf("%s/%s/profile/%s", baseURL, locale, creatorName)
	if locale == "zh-CN" {
		return "有人通过你的链接完成了匹配 — 25types",
			fmt.Sprintf("%s 通过你的链接完成了一次匹配。\n\n查看你的匹配记录：%s\n\n— 25types", otherName, profileURL)
	}
	return "Someone matched with you on 25types",
		fmt.Sprintf("%s completed a bond match through your link.\n\nView your bonds: %s\n\n— 25types", otherName, profileURL)
}

// ThankYouEmail returns subject and body for donation thank-you.
func ThankYouEmail(locale string) (subject, text string) {
	if locale == "zh-CN" {
		return "感谢你的捐赠 — 25types",
			"感谢你对 25types 的慷慨支持！你的捐赠帮助我们保持服务免费开放给所有人。\n\n— 25types"
	}
	return "Thank you for your donation — 25types",
		"Thank you for your generous support of 25types! Your donation helps us keep the service free and open to everyone.\n\n— 25types"
}
