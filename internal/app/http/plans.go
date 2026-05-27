package http

func defaultPlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID:       "monthly",
			Name:     "月度无限",
			NameEn:   "Monthly Unlimited",
			Amount:   29.0,
			Interval: "month",
			Features: []string{"无限生成报告", "报告历史保存"},
		},
		{
			ID:       "yearly",
			Name:     "年度方案",
			NameEn:   "Yearly Plan",
			Amount:   99.0,
			Interval: "year",
			Features: []string{"无限报告", "起名/择日完整方案", "优先新功能"},
		},
	}
}
