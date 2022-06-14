// Copyright (c) 2021-2022 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.cn) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package experimental

import (
	"github.com/easysoft/qcadmin/internal/pkg/cli/helm"
	"github.com/spf13/cobra"
)

// HelmCommand helm command.
func HelmCommand() *cobra.Command {
	return helm.EmbedCommand()
}