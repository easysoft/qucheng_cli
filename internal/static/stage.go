// Copyright (c) 2021-2022 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package static

import (
	"github.com/easysoft/qcadmin/common"
	"github.com/easysoft/qcadmin/internal/static/data"
	"github.com/easysoft/qcadmin/internal/static/deploy"
	"github.com/easysoft/qcadmin/internal/static/haogstls"
	"github.com/easysoft/qcadmin/internal/static/scripts"
)

func StageFiles() error {
	dataDir := common.GetDefaultDataDir()
	if err := data.Stage(dataDir); err != nil {
		return err
	}
	if err := deploy.Stage(dataDir); err != nil {
		return err
	}
	if err := scripts.Stage(dataDir); err != nil {
		return err
	}
	if err := haogstls.Stage(dataDir); err != nil {
		return err
	}
	return nil
}
