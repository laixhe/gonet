package cgibin

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type UserGetResponse struct {
	ErrCode        int      `json:"errcode"`           // 出错返回码，(0 成功)(-1 系统繁忙)
	ErrMsg         string   `json:"errmsg"`            // 返回码提示语
	Userid         string   `json:"userid"`            // 成员 UserID
	Name           string   `json:"name"`              // 成员名称
	Department     []int    `json:"department"`        // 成员所属部门 id 列表
	Order          []int    `json:"order"`             // 部门内的排序值(默认为0)数量必须和 department 一致，数值越大排序越前面
	Position       string   `json:"position"`          // 职务信息
	Mobile         string   `json:"mobile"`            // 手机号码
	Gender         string   `json:"gender"`            // 性别 0表示未定义 1表示男性 2表示女性
	Email          string   `json:"email"`             // 邮箱
	BizMail        string   `json:"biz_mail"`          // 企业邮箱
	IsLeaderInDept []int    `json:"is_leader_in_dept"` // 表示在所在的部门内是否为部门负责人，数量与 department 一致
	DirectLeader   []string `json:"direct_leader"`     // 直属上级 UserID，返回在应用可见范围内的直属上级列表，最多有1个直属上级
	Avatar         string   `json:"avatar"`            // 头像 url
	ThumbAvatar    string   `json:"thumb_avatar"`      // 头像缩略图 url
	Telephone      string   `json:"telephone"`         // 座机
	Alias          string   `json:"alias"`             // 别名
	Address        string   `json:"address"`           // 地址
	OpenUserid     string   `json:"open_userid"`       // 全局唯一
	MainDepartment int      `json:"main_department"`   // 主部门
	Extattr        struct {
		Attrs []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
		} `json:"attrs"`
	} `json:"extattr"` // 扩展属性
	Status           int    `json:"status"`            // 激活状态: 1=已激活，2=已禁用，4=未激活，5=退出企业
	QrCode           string `json:"qr_code"`           // 员工个人二维码
	ExternalPosition string `json:"external_position"` // 对外职务
	ExternalProfile  struct {
		ExternalCorpName string `json:"external_corp_name"`
		WechatChannels   struct {
			Nickname string `json:"nickname"`
			Status   int    `json:"status"`
		} `json:"wechat_channels"`
		ExternalAttr []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
			Miniprogram struct {
				Appid    string `json:"appid"`
				Pagepath string `json:"pagepath"`
				Title    string `json:"title"`
			} `json:"miniprogram,omitempty"`
		} `json:"external_attr"`
	} `json:"external_profile"` // 成员对外属性
}

// UserGet 读取成员
// DOC https://developer.work.weixin.qq.com/document/path/90196
// GET https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=ACCESS_TOKEN&userid=USERID
func UserGet(httpClient *resty.Client, accessToken, userid string) (*UserGetResponse, error) {
	httpResp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
			"userid":       userid,
		}).
		SetResult(&UserGetResponse{}).
		SetForceResponseContentType("application/json").
		Get("/cgi-bin/user/get")
	if err != nil {
		return &UserGetResponse{ErrCode: -1, ErrMsg: err.Error()}, err
	}
	if httpResp.IsSuccess() {
		resp, is := httpResp.Result().(*UserGetResponse)
		if is {
			if resp.ErrCode != 0 {
				return resp, fmt.Errorf("%d %s", resp.ErrCode, resp.ErrMsg)
			}
			return resp, nil
		}
	}
	return &UserGetResponse{
		ErrCode: httpResp.StatusCode(),
		ErrMsg:  httpResp.String(),
	}, errors.New(httpResp.String())
}
