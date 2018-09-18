package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "embeddemo_pro1",
	Short: "embeddemo_pro1",
	Run:   Run,
}
var listenURL string
var peerNodeURLs []string
var autoTrans bool

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	rootCmd.PersistentFlags().StringVar(&listenURL, "listen", "", "")
	rootCmd.PersistentFlags().StringArrayVar(&peerNodeURLs, "peer-nodes", nil, "")
	rootCmd.PersistentFlags().BoolVar(&autoTrans, "auto-trans", false, "auto trans state for demo")
}
func Run(cmd *cobra.Command, args []string) {
	node, err := NewNode()
	if err != nil {
		panic(err)
	}
	node.Startup()
}
func main() {
	rootCmd.Execute()
}
