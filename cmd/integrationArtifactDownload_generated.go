// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type integrationArtifactDownloadOptions struct {
	APIServiceKey          string `json:"apiServiceKey,omitempty"`
	IntegrationFlowID      string `json:"integrationFlowId,omitempty"`
	IntegrationFlowVersion string `json:"integrationFlowVersion,omitempty"`
	DownloadPath           string `json:"downloadPath,omitempty"`
}

// IntegrationArtifactDownloadCommand Download integration flow runtime artefact
func IntegrationArtifactDownloadCommand() *cobra.Command {
	const STEP_NAME = "integrationArtifactDownload"

	metadata := integrationArtifactDownloadMetadata()
	var stepConfig integrationArtifactDownloadOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
		splunkClient = &splunk.Splunk{}
	}
	telemetryClient := &telemetry.Telemetry{}

	var createIntegrationArtifactDownloadCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Download integration flow runtime artefact",
		Long:  `With this step you can download a integration flow runtime artifact, which returns a zip file with the integration flow contents in to current workspace using the OData API. Learn more about the SAP Cloud Integration remote API for downloading an integration flow artifact [here](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US/d1679a80543f46509a7329243b595bdb.html).`,
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
			log.RegisterSecret(stepConfig.APIServiceKey)

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
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
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
			integrationArtifactDownload(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addIntegrationArtifactDownloadFlags(createIntegrationArtifactDownloadCmd, &stepConfig)
	return createIntegrationArtifactDownloadCmd
}

func addIntegrationArtifactDownloadFlags(cmd *cobra.Command, stepConfig *integrationArtifactDownloadOptions) {
	cmd.Flags().StringVar(&stepConfig.APIServiceKey, "apiServiceKey", os.Getenv("PIPER_apiServiceKey"), "Service key JSON string to access the Process Integration Runtime service instance of plan 'api'")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowID, "integrationFlowId", os.Getenv("PIPER_integrationFlowId"), "Specifies the ID of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowVersion, "integrationFlowVersion", os.Getenv("PIPER_integrationFlowVersion"), "Specifies the version of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.DownloadPath, "downloadPath", os.Getenv("PIPER_downloadPath"), "Specifies integration artifact download location.")

	cmd.MarkFlagRequired("apiServiceKey")
	cmd.MarkFlagRequired("integrationFlowId")
	cmd.MarkFlagRequired("integrationFlowVersion")
	cmd.MarkFlagRequired("downloadPath")
}

// retrieve step metadata
func integrationArtifactDownloadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "integrationArtifactDownload",
			Aliases:     []config.Alias{},
			Description: "Download integration flow runtime artefact",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "cpiApiServiceKeyCredentialsId", Description: "Jenkins secret text credential ID containing the service key to the Process Integration Runtime service instance of plan 'api'", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "apiServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "cpiApiServiceKeyCredentialsId",
								Param: "apiServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_apiServiceKey"),
					},
					{
						Name:        "integrationFlowId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationFlowId"),
					},
					{
						Name:        "integrationFlowVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationFlowVersion"),
					},
					{
						Name:        "downloadPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_downloadPath"),
					},
				},
			},
		},
	}
	return theMetaData
}
