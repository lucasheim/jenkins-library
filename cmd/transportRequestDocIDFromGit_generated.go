// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type transportRequestDocIDFromGitOptions struct {
	GitFrom             string `json:"gitFrom,omitempty"`
	GitTo               string `json:"gitTo,omitempty"`
	ChangeDocumentLabel string `json:"changeDocumentLabel,omitempty"`
}

type transportRequestDocIDFromGitCommonPipelineEnvironment struct {
	custom struct {
		changeDocumentID string
	}
}

func (p *transportRequestDocIDFromGitCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "changeDocumentId", value: p.custom.changeDocumentID},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// TransportRequestDocIDFromGitCommand Retrieves change document ID from Git repository
func TransportRequestDocIDFromGitCommand() *cobra.Command {
	const STEP_NAME = "transportRequestDocIDFromGit"

	metadata := transportRequestDocIDFromGitMetadata()
	var stepConfig transportRequestDocIDFromGitOptions
	var startTime time.Time
	var commonPipelineEnvironment transportRequestDocIDFromGitCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	splunkClient := &splunk.Splunk{}
	telemetryClient := &telemetry.Telemetry{}
	provider, err := orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	var createTransportRequestDocIDFromGitCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Retrieves change document ID from Git repository",
		Long: `This step scans the commit messages of the Git repository for a pattern to retrieve the change document ID.
It is primarily made for the transportRequestUploadSOLMAN step to provide the change document ID by Git means.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			customTelemetryData := telemetry.CustomData{}
			customTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				customTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				customTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				customTelemetryData.Custom1Label = "PiperCommitHash"
				customTelemetryData.Custom1 = GitCommit
				customTelemetryData.Custom2Label = "PiperTag"
				customTelemetryData.Custom2 = GitTag
				customTelemetryData.Custom3Label = "Stage"
				customTelemetryData.Custom3 = provider.GetStageName()
				telemetryClient.SetData(&customTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			transportRequestDocIDFromGit(stepConfig, &customTelemetryData, &commonPipelineEnvironment)
			customTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestDocIDFromGitFlags(createTransportRequestDocIDFromGitCmd, &stepConfig)
	return createTransportRequestDocIDFromGitCmd
}

func addTransportRequestDocIDFromGitFlags(cmd *cobra.Command, stepConfig *transportRequestDocIDFromGitOptions) {
	cmd.Flags().StringVar(&stepConfig.GitFrom, "gitFrom", `origin/master`, "GIT starting point for retrieving the change document and transport request ID")
	cmd.Flags().StringVar(&stepConfig.GitTo, "gitTo", `HEAD`, "GIT ending point for retrieving the change document and transport request ID")
	cmd.Flags().StringVar(&stepConfig.ChangeDocumentLabel, "changeDocumentLabel", `ChangeDocument`, "Pattern used for identifying lines holding the change document ID. The GIT commit log messages are scanned for this label")

}

// retrieve step metadata
func transportRequestDocIDFromGitMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestDocIDFromGit",
			Aliases:     []config.Alias{},
			Description: "Retrieves change document ID from Git repository",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "gitFrom",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/from"}},
						Default:     `origin/master`,
					},
					{
						Name:        "gitTo",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/to"}},
						Default:     `HEAD`,
					},
					{
						Name:        "changeDocumentLabel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/changeDocumentLabel"}},
						Default:     `ChangeDocument`,
					},
				},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "custom/changeDocumentId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
