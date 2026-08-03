package model

type SystemSettingUpdateDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type SystemSettingsUpdateReq struct {
	Settings []SystemSettingUpdateDTO `json:"settings"`
}
type SystemSettingDefaultDTO struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
type SystemSettingDefaultGroupsDTO struct {
	Basic        []SystemSettingDefaultDTO `json:"basic,omitempty"`
	Announcement []SystemSettingDefaultDTO `json:"announcement,omitempty"`
	Alipay       []SystemSettingDefaultDTO `json:"alipay,omitempty"`
	Wechat       []SystemSettingDefaultDTO `json:"wechat,omitempty"`
}
type SystemSettingsDefaultsResp struct {
	General   []SystemSettingDefaultDTO     `json:"general"`
	Operation SystemSettingDefaultGroupsDTO `json:"operation"`
	Payment   SystemSettingDefaultGroupsDTO `json:"payment"`
	OAuth     SystemSettingDefaultGroupsDTO `json:"oauth"`
}

func DefaultSystemSettings() SystemSettingsDefaultsResp {
	return SystemSettingsDefaultsResp{
		General:   []SystemSettingDefaultDTO{{"siteName", "API 网关", "general", "网站名称"}, {"siteDescription", "一站式 API 服务平台", "general", "网站描述"}, {"siteLogo", "", "general", "网站 Logo"}, {"contactEmail", "support@example.com", "general", "联系邮箱"}},
		Operation: SystemSettingDefaultGroupsDTO{Basic: []SystemSettingDefaultDTO{{"registrationEnabled", "true", "operation", "是否允许用户注册"}, {"defaultCredits", "1000", "operation", "新用户默认积分"}, {"inviteRewards", "100", "operation", "邀请奖励积分"}, {"maintenanceMode", "false", "operation", "是否开启维护模式"}}, Announcement: []SystemSettingDefaultDTO{{"announcementEnabled", "false", "operation", "是否启用公告"}, {"announcementContent", "", "operation", "公告内容"}}},
		Payment:   SystemSettingDefaultGroupsDTO{Basic: []SystemSettingDefaultDTO{{"alipayEnabled", "false", "payment", "是否开启支付宝支付"}, {"wechatEnabled", "false", "payment", "是否开启微信支付"}, {"creditPrice", "1", "payment", "每积分价格（元）"}, {"minRecharge", "10", "payment", "最低充值金额（元）"}, {"mockPaymentEnabled", "false", "payment", "是否启用模拟支付"}, {"mockPaymentAutoSuccess", "true", "payment", "模拟支付自动成功"}, {"mockPaymentDelay", "2000", "payment", "模拟支付延迟时间（毫秒）"}}, Alipay: []SystemSettingDefaultDTO{{"alipayAppId", "", "payment", "支付宝 AppID"}, {"alipayPrivateKey", "", "payment", "支付宝私钥"}, {"alipayPublicKey", "", "payment", "支付宝公钥"}, {"alipayNotifyUrl", "", "payment", "支付宝回调地址"}, {"alipayReturnUrl", "", "payment", "支付宝返回地址"}, {"alipaySandbox", "false", "payment", "支付宝沙箱模式"}}, Wechat: []SystemSettingDefaultDTO{{"wechatPayAppId", "", "payment", "微信支付 AppID"}, {"wechatPayMchId", "", "payment", "微信支付商户号"}, {"wechatPayApiKey", "", "payment", "微信支付 API 密钥"}, {"wechatPayPrivateKey", "", "payment", "微信支付私钥"}, {"wechatPayPublicKey", "", "payment", "微信支付公钥"}, {"wechatPayPaymentPublicKey", "", "payment", "微信支付平台公钥"}, {"wechatPayPublicKeyId", "", "payment", "微信支付公钥 ID"}, {"wechatPayNotifyUrl", "", "payment", "微信支付回调地址"}, {"wechatPayDebug", "false", "payment", "微信支付调试模式"}}},
		OAuth:     SystemSettingDefaultGroupsDTO{Basic: []SystemSettingDefaultDTO{{"oauthProviders", "[]", "oauth", "OAuth 提供商配置"}}},
	}
}
