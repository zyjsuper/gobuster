package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdns"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

var cmdDNS *cobra.Command

func runDNS(cmd *cobra.Command, args []string) error {
	globalopts, pluginopts, err := parseDNSOptions()
	if err != nil {
		return fmt.Errorf("error on parsing arguments: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plugin, err := gobusterdns.NewGobusterDNS(globalopts, pluginopts)
	if err != nil {
		return fmt.Errorf("Error on creating gobusterdns: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			// caught CTRL+C
			if !globalopts.Quiet {
				fmt.Println("\n[!] Keyboard interrupt detected, terminating.")
			}
			cancel()
		}
	}()

	if err := cli.Gobuster(ctx, globalopts, plugin); err != nil {
		return fmt.Errorf("error on running goubster: %v", err)
	}
	return nil
}

func parseDNSOptions() (*libgobuster.Options, *gobusterdns.OptionsDNS, error) {
	globalopts, err := parseGlobalOptions()
	if err != nil {
		return nil, nil, err
	}
	plugin := gobusterdns.NewOptionsDNS()

	plugin.Domain, err = cmdDNS.Flags().GetString("domain")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for domain: %v", err)
	}

	plugin.ShowIPs, err = cmdDNS.Flags().GetBool("showips")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for showips: %v", err)
	}

	plugin.ShowCNAME, err = cmdDNS.Flags().GetBool("showcname")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for showcname: %v", err)
	}

	plugin.WildcardForced, err = cmdDNS.Flags().GetBool("wildcard")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for wildcard: %v", err)
	}

	plugin.Timeout, err = cmdDNS.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for timeout: %v", err)
	}

	plugin.Resolver, err = cmdDNS.Flags().GetString("resolver")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value for resolver: %v", err)
	}

	if plugin.Resolver != "" && runtime.GOOS == "windows" {
		return nil, nil, fmt.Errorf("currently can not set custom dns resolver on windows. See https://golang.org/pkg/net/#hdr-Name_Resolution")
	}

	return globalopts, plugin, nil
}

func init() {
	cmdDNS = &cobra.Command{
		Use:   "dns",
		Short: "Uses DNS subdomain bruteforcing mode",
		RunE:  runDNS,
	}

	cmdDNS.Flags().StringP("domain", "d", "", "The target domain")
	cmdDNS.Flags().BoolP("showips", "i", false, "Show IP addresses")
	cmdDNS.Flags().BoolP("showcname", "c", false, "Show CNAME records (cannot be used with '-i' option)")
	cmdDNS.Flags().DurationP("timeout", "", time.Second, "DNS resolver timeout")
	cmdDNS.Flags().BoolP("wildcard", "", false, "Force continued operation when wildcard found")
	cmdDNS.Flags().StringP("resolver", "r", "", "Use custom DNS server (format server.com or server.com:port)")
	if err := cmdDNS.MarkFlagRequired("domain"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}

	cmdDNS.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureGlobalOptions()
	}

	rootCmd.AddCommand(cmdDNS)
}
