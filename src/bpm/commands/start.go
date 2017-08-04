// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations
// under the License.

package commands

import (
	"bpm/config"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	configPath string
)

func init() {
	startCommand.Flags().StringVarP(&procName, "process", "p", "", "The optional process name.")
	startCommand.Flags().StringVarP(&configPath, "config", "c", "", "The optional process config path.")
	RootCmd.AddCommand(startCommand)
}

var startCommand = &cobra.Command{
	Long:     "Starts a BOSH Process",
	RunE:     start,
	Short:    "Starts a BOSH Process",
	Use:      "start <job-name>",
	PreRunE:  startPre,
	PostRunE: startPost,
}

func startPre(cmd *cobra.Command, args []string) error {
	if err := validateInput(args); err != nil {
		return err
	}

	if err := setupBpmLogs("start"); err != nil {
		return err
	}

	return acquireLifecycleLock()
}

func startPost(cmd *cobra.Command, args []string) error {
	return releaseLifecycleLock()
}

func start(cmd *cobra.Command, _ []string) error {
	logger.Info("starting")
	defer logger.Info("complete")

	if configPath == "" {
		procConfigFile := fmt.Sprintf("%s.yml", bpmCfg.ProcName())
		configPath = filepath.Join(bpmCfg.ProcConfigDir(), procConfigFile)
	}

	procCfg, err := config.ParseProcessConfig(configPath)
	if err != nil {
		logger.Error("failed-to-parse-config", err)
		return err
	}

	runcLifecycle := newRuncLifecycle()

	err = runcLifecycle.StartJob(bpmCfg, procCfg)
	if err != nil {
		logger.Error("failed-to-start", err)

		removeErr := runcLifecycle.RemoveJob(bpmCfg)
		if removeErr != nil {
			logger.Error("failed-to-cleanup", removeErr)
		}

		return err
	}

	return nil
}
