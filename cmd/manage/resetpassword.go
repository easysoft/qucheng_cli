// Copyright (c) 2021-2022 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package manage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/easysoft/qcadmin/common"
	"github.com/easysoft/qcadmin/internal/app/config"
	"github.com/easysoft/qcadmin/internal/pkg/k8s"
	"github.com/easysoft/qcadmin/internal/pkg/util/factory"
	"github.com/ergoapi/util/color"
	"github.com/ergoapi/util/exnet"
	"github.com/ergoapi/util/expass"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Account string `json:"account"`
	} `json:"data"`
}

type Body struct {
	Password string `json:"password"`
}

func NewResetPassword(f factory.Factory) *cobra.Command {
	log := f.GetLog()
	var password string
	rp := &cobra.Command{
		Use:     "reset-password",
		Short:   "reset qucheng superadmin password",
		Aliases: []string{"rp", "re-pass"},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig()
			if cfg.APIToken == "" {
				k8sClient, err := k8s.NewSimpleClient()
				if err != nil {
					log.Errorf("k8s client err: %v", err)
					return
				}
				cneapiDeploy, err := k8sClient.GetDeployment(context.Background(), common.DefaultSystem, "qucheng", metav1.GetOptions{})
				if err != nil {
					log.Errorf("get k8s deploy err: %v", err)
					return
				}
				for _, e := range cneapiDeploy.Spec.Template.Spec.Containers[0].Env {
					if e.Name == "CNE_API_TOKEN" {
						cfg.APIToken = e.Value
						break
					}
				}
				cfg.SaveConfig()
			}

			if cfg.Domain == "" {
				cfg.Domain = fmt.Sprintf("%s:32379", exnet.LocalIPs()[0])
			}

			log.Debug("fetch api token")
			// 更新密码
			if len(password) == 0 {
				log.Warn("not found password, will generate random password")
				password = expass.SaltMd5Pass(cfg.APIToken, expass.RandomPassword(16))
			}
			log.Infof("update superadmin password: %s", "")
			client := req.C()
			if log.GetLevel() > logrus.InfoLevel {
				client = client.DevMode().EnableDumpAll()
			}
			var result Result
			resp, err := client.R().
				SetHeader("accept", "application/json").
				SetHeader("TOKEN", cfg.APIToken).
				SetBody(&Body{Password: password}).
				Post(fmt.Sprintf("http://%s/admin-resetpassword.html", cfg.Domain))
			if err != nil {
				log.Errorf("update password failed, reason: %v", err)
				return
			}
			if !resp.IsSuccess() {
				log.Errorf("update password failed, reason: bad response status %v", resp.Status)
				return
			}
			json.Unmarshal([]byte(resp.String()), &result)
			log.Donef("update superadmin %s password %s success.", color.SGreen(result.Data.Account), color.SGreen(password))
		},
	}
	rp.Flags().StringVarP(&password, "password", "p", "", "superadmin password")
	return rp
}
