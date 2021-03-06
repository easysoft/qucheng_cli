package app

import (
	"context"
	"fmt"

	"github.com/easysoft/qcadmin/internal/app/debug"
	"github.com/easysoft/qcadmin/internal/pkg/k8s"
	"github.com/easysoft/qcadmin/internal/pkg/util/factory"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func NewCmdAppLogs(f factory.Factory) *cobra.Command {
	var previous bool
	log := f.GetLog()
	app := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "logs app",
		Args:    cobra.ExactArgs(1),
		Example: `q app exec http://console.efbb.haogs.cn/instance-view-39.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			apidebug := log.GetLevel() == logrus.DebugLevel
			log.Infof("start logs app: %s", url)
			appdata, err := debug.GetNameByURL(url, apidebug)
			if err != nil {
				return err
			}
			k8sClient, err := k8s.NewSimpleClient()
			if err != nil {
				log.Errorf("k8s client err: %v", err)
				return err
			}
			ctx := context.Background()
			podlist, _ := k8sClient.ListPods(ctx, "default", metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					"release": appdata.K8Name,
				}).String(),
			})
			if len(podlist.Items) < 1 {
				return fmt.Errorf("podnum %d,  app maybe not running", len(podlist.Items))
			}
			templates := &promptui.SelectTemplates{
				Label:    "{{ . }}?",
				Active:   "\U0001F449 {{ .Name | cyan }}",
				Inactive: "  {{ .Name | cyan }}",
				Selected: "\U0001F389 {{ .Name | red | cyan }}",
			}

			prompt := promptui.Select{
				Label:     "select pod",
				Items:     podlist.Items,
				Templates: templates,
				Size:      5,
			}
			it, _, _ := prompt.Run()
			return k8sClient.GetFollowLogs(ctx, "default", podlist.Items[it].Name, podlist.Items[it].Spec.Containers[0].Name, previous)
		},
	}
	app.Flags().BoolVarP(&previous, "previous", "p", false, " If true, print the logs for the previous instance of the container in a pod if it exists.")
	return app
}
