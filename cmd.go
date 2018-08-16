package cmd

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
	"qorsam/config/system"
	"github.com/aghape/aghape"
)

func SiteCommand(command *cobra.Command, run ...func(cmd *cobra.Command, site qor.SiteInterface, args []string)) *cobra.Command {
	Args := command.Args
	command.Args = func(cmd *cobra.Command, args []string) (err error) {
		err = cobra.MinimumNArgs(1)(cmd, args)
		if err == nil && system.Sites.Get(args[0]) == nil {
			return fmt.Errorf("Site %q does not exists.\n", args[0])
		}

		if Args != nil {
			return Args(cmd, args[1:])
		}
		return
	}
	if len(run) == 1 {
		command.Run = func(cmd *cobra.Command, args []string) {
			run[0](cmd, system.Sites.Get(args[0]), args[1:])
		}
	}

	UseParts := strings.Split(command.Use, " ")
	command.Use = strings.Join(append([]string{UseParts[0], "SITE_NAME"}, UseParts[1:]...), " ")
	return command
}

func SitesCommand(command *cobra.Command, run ...func(cmd *cobra.Command, site qor.SiteInterface, args []string) error) *cobra.Command {
	Args := command.Args
	command.Args = func(cmd *cobra.Command, args []string) (err error) {
		err = cobra.MinimumNArgs(1)(cmd, args)
		if err != nil {
			return
		}

		if args[0] != "*" && system.Sites.Get(args[0]) == nil {
			return fmt.Errorf("Site %q does not exists.\n", args[0])
		}

		if Args != nil {
			return Args(cmd, args[1:])
		}
		return
	}
	if len(run) == 1 {
		command.RunE = func(cmd *cobra.Command, args []string) error {
			siteName := args[0]
			args = args[1:]

			callSite := func(site qor.SiteInterface) error {
				defer func() {
					site.EachDB(func(db *qor.DB) bool {
						db.Raw.Close()
						return true
					})
				}()
				err := run[0](cmd, site, args)
				if err != nil {
					return errwrap.Wrap(err, "Site %q", site.Name())
				}
				return nil
			}

			if siteName == "*" {
				system.Sites.Each(callSite)
			} else {
				site := system.Sites.Get(args[0])
				callSite(site)
			}
		}
	}

	UseParts := strings.Split(command.Use, " ")
	command.Use = strings.Join(append([]string{UseParts[0], "[SITE_NAME]"}, UseParts[1:]...), " ")
	return command
}