package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"goland-ddns/internal/update_dns"
	"goland-ddns/pkg/cloudflare"
	"goland-ddns/pkg/config"
	httpPkg "goland-ddns/pkg/http"
)

// updateDDNS represents the binance command
var updateDDNS = &cobra.Command{
	Use:   "update_dns",
	Short: "Update local ip from local to cloudflare via api token",
	Long:  `Simple command for update dns ip to cloudflare via api token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := zap.S()

		logger.Infof("update dns to cloudflare with config: %+v", config.ENV.Cloudflare.Domain)

		requestClient := httpPkg.NewAPIRequestClient(logger)

		api, err := cloudflare.NewCloudflareClient(logger, requestClient)
		if err != nil {
			return err
		}

		updateDNS := update_dns.NewUpdateDns(logger, api, requestClient)
		return updateDNS.Run()

	},
}

func init() {
	rootCmd.AddCommand(updateDDNS)
}
