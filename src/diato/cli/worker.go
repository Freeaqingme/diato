// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cli

import (
	//"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	//"diato/server"
	"diato/util/stop"

	"github.com/spf13/cobra"
	//gcfg "gopkg.in/gcfg.v1"
	"diato/worker"
)

var workerCmd = &cobra.Command{
	Use:   "internal-worker",
	Short: "Internal Worker. Should not be invoked manually",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the internal worker",
	RunE:  runWorker,
}

func init() {
	workerCmd.AddCommand(
		workerStartCmd,
	)
}

func runWorker(_ *cobra.Command, args []string) error {
	log.Printf("Starting Worker")

	//config := server.NewConfig()
	//err := gcfg.ReadFileInto(config, daemonOpts.ConfFile)
	//if err != nil {
	//	return fmt.Errorf("Could not parse configuration: %s", err.Error())
	//}
	//
	//if err = config.Validate(); err != nil {
	//	return fmt.Errorf("Could not parse configuration: %s", err.Error())
	//}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGQUIT)

	//s := server.NewServer(config)
	//
	//if err := s.Start(); err != nil {
	//	return fmt.Errorf("diato could not start: %s", err)
	//}

	w := worker.NewWorker()
	w.Start()

	select {
	case sig := <-signalCh:
		log.Printf("received signal '%s', exiting worker...", sig)
		stop.Stop()
	}

	log.Printf("Successfully ceased all worker operations")
	return nil
}
