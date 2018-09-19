package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var lu string
var pns []string

var rootCmd = &cobra.Command{
	Use:   "skeleton",
	Short: "skeleton",
	Run:   Run,
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	//file, _ := os.OpenFile("skeleton.log", os.O_CREATE|os.O_WRONLY, 0666)
	//log.SetOutput(file)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	rootCmd.Flags().StringVar(&lu, "listen-url", "", "communicate between peer nodes")
	rootCmd.Flags().StringSliceVar(&pns, "peer-node", nil, "peer node address:port")
}

func main() {
	rootCmd.Execute()
}

func Run(*cobra.Command, []string) {
	node, err := newNode(lu, pns)
	if err != nil {
		log.Fatal(err)
	}
	node.Startup()
}
