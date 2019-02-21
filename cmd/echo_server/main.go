package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
)

var colorer = color.New()

var (
	host  string
	port  int
	root  string
	level int
)

var cmd = &cobra.Command{
	Use:   "echo_server",
	Short: "Echo: High performance, extensible, minimalist Go web framework",
	Long:  fmt.Sprintf(echo.Banner, colorer.Red("v"+echo.Version), colorer.Blue(echo.Website)),
	RunE: func(cmd *cobra.Command, args []string) error {
		e := echo.New()

		e.Logger.SetLevel(log.Lvl(level))
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(middleware.RequestID())
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   root,
			Browse: true,
		}))

		address := fmt.Sprintf("%s:%d", host, port)
		return e.Start(address)
	},
}

func init() {
	cw, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "host to listen on")
	cmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "port to listen on")
	cmd.PersistentFlags().StringVarP(&root, "root", "r", cw, "root directory to serve")
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
