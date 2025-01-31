package splunk

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// Splunk SplunkHook provides a logrus hook which enables error logging to splunk platform.
// This is helpful in order to provide better monitoring and alerting on errors
// as well as the given error details can help to find the root cause of bugs.
type Splunk struct {
	levels        []logrus.Level
	tags          map[string]string
	splunkClient  piperhttp.Client
	correlationID string
	splunkDsn     string
	splunkIndex   string

	// boolean which forces to send all logs on error or none at all
	sendLogs bool

	// How big can be batch of messages
	postMessagesBatchSize int
}

func (s *Splunk) Initialize(correlationID, dsn, token, index string, sendLogs bool) error {
	log.Entry().Debugf("Initializing Splunk with DSN %v", dsn)

	if !strings.HasPrefix(token, "Splunk ") {
		token = "Splunk " + token
	}

	log.RegisterSecret(token)
	client := piperhttp.Client{}

	client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration:        5 * time.Second,
		Token:                     token,
		TransportSkipVerification: true,
		MaxRetries:                -1,
	})

	s.splunkClient = client
	s.splunkDsn = dsn
	s.splunkIndex = index
	s.correlationID = correlationID
	s.postMessagesBatchSize = 20000
	s.sendLogs = sendLogs

	return nil
}

func (s *Splunk) Send(telemetryData telemetry.Data, logCollector *log.CollectorHook) error {
	// Sends telemetry and or additionally logging data to Splunk
	preparedTelemetryData := s.prepareTelemetry(telemetryData)
	messagesLen := len(logCollector.Messages)
	// TODO: Logic for errorCategory (undefined, service, infrastructure)
	if telemetryData.ErrorCode == "0" || (telemetryData.ErrorCode == "1" && !s.sendLogs) {
		// Either Successful run, we only send the telemetry data, no logging information
		// OR Failure run and we do not want to send the logs
		err := s.tryPostMessages(preparedTelemetryData, []log.Message{})
		if err != nil {
			return errors.Wrap(err, "error while sending logs")
		}
		return nil
	} else {
		// ErrorCode indicates an error in the step, so we want to send all the logs with telemetry
		for i := 0; i < messagesLen; i += s.postMessagesBatchSize {
			upperBound := i + s.postMessagesBatchSize
			if upperBound > messagesLen {
				upperBound = messagesLen
			}
			err := s.tryPostMessages(preparedTelemetryData, logCollector.Messages[i:upperBound])
			if err != nil {
				return errors.Wrap(err, "error while sending logs")
			}
		}
	}
	return nil
}

func readCommonPipelineEnvironment(filePath string) string {

	// TODO: Dependent on a groovy step, which creates the folder.
	contentFile, err := ioutil.ReadFile(".pipeline/commonPipelineEnvironment/" + filePath)
	if err != nil {
		log.Entry().Warnf("Could not read %v file. %v", filePath, err)
		contentFile = []byte("N/A")
	}
	return string(contentFile)
}

// MonitoringData definition for monitoring
type MonitoringData struct {
	PipelineUrlHash string `json:"PipelineUrlHash,omitempty"`
	BuildUrlHash    string `json:"BuildUrlHash,omitempty"`
	StageName       string `json:"StageName,omitempty"`
	StepName        string `json:"StepName,omitempty"`
	ExitCode        string `json:"ExitCode,omitempty"`
	Duration        string `json:"Duration,omitempty"`
	ErrorCode       string `json:"ErrorCode,omitempty"`
	ErrorCategory   string `json:"ErrorCategory,omitempty"`
	CorrelationID   string `json:"CorrelationID,omitempty"`
	CommitHash      string `json:"CommitHash,omitempty"`
	Branch          string `json:"Branch,omitempty"`
	GitOwner        string `json:"GitOwner,omitempty"`
	GitRepository   string `json:"GitRepository,omitempty"`
}

func (s *Splunk) prepareTelemetry(telemetryData telemetry.Data) MonitoringData {

	return MonitoringData{
		PipelineUrlHash: telemetryData.PipelineURLHash,
		BuildUrlHash:    telemetryData.BuildURLHash,
		StageName:       telemetryData.StageName,
		StepName:        telemetryData.BaseData.StepName,
		ExitCode:        telemetryData.CustomData.ErrorCode,
		Duration:        telemetryData.CustomData.Duration,
		ErrorCode:       telemetryData.CustomData.ErrorCode,
		ErrorCategory:   telemetryData.CustomData.ErrorCategory,
		CorrelationID:   s.correlationID,
		CommitHash:      readCommonPipelineEnvironment("git/headCommitId"),
		Branch:          readCommonPipelineEnvironment("git/branch"),
		GitOwner:        readCommonPipelineEnvironment("github/owner"),
		GitRepository:   readCommonPipelineEnvironment("github/repository"),
	}
}

type Event struct {
	Messages  []log.Message  `json:"messages,omitempty"`  // messages
	Telemetry MonitoringData `json:"telemetry,omitempty"` // telemetryData
}
type Details struct {
	Host       string `json:"host"`                 // hostname
	Source     string `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      Event  `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

func (s *Splunk) SendPipelineStatus(pipelineTelemetryData telemetry.PipelineTelemetry, logFile *[]byte) error {
	// Sends telemetry and or additionally logging data to Splunk

	readLogFile := string(*logFile)
	splitted := strings.Split(readLogFile, "\n")
	messagesLen := len(splitted)

	log.Entry().Debugf("Sending %v messages to Splunk.", messagesLen)
	log.Entry().Debugf("Sending pipeline telemetry data to Splunk: %v", pipelineTelemetryData)
	s.postTelemetry(pipelineTelemetryData)

	if s.sendLogs {
		for i := 0; i < messagesLen; i += s.postMessagesBatchSize {
			upperBound := i + s.postMessagesBatchSize
			if upperBound > messagesLen {
				upperBound = messagesLen
			}
			err := s.postLogFile(pipelineTelemetryData, splitted[i:upperBound])
			if err != nil {
				return errors.Wrap(err, "error while sending logs")
			}
		}
	}
	return nil
}

type LogFileEvents struct {
	Messages  []string                    `json:"messages,omitempty"`  // messages
	Telemetry telemetry.PipelineTelemetry `json:"telemetry,omitempty"` // telemetryData
}
type DetailsLog struct {
	Host       string        `json:"host"`                 // hostname
	Source     string        `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string        `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string        `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      LogFileEvents `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

type DetailsTelemetry struct {
	Host       string                      `json:"host"`                 // hostname
	Source     string                      `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string                      `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string                      `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      telemetry.PipelineTelemetry `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

func (s *Splunk) postTelemetry(telemetryData telemetry.PipelineTelemetry) error {

	details := DetailsTelemetry{
		Host:       s.correlationID,
		SourceType: "_json",
		Index:      s.splunkIndex,
		Event:      telemetryData,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := ioutil.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}

func (s *Splunk) postLogFile(telemetryData telemetry.PipelineTelemetry, messages []string) error {

	event := LogFileEvents{
		Messages:  messages,
		Telemetry: telemetryData,
	}
	details := DetailsLog{
		Host:       s.correlationID,
		SourceType: "txt",
		Index:      s.splunkIndex,
		Event:      event,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := ioutil.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}

func (s *Splunk) tryPostMessages(telemetryData MonitoringData, messages []log.Message) error {

	event := Event{
		Messages:  messages,
		Telemetry: telemetryData,
	}
	details := Details{
		Host:       s.correlationID,
		SourceType: "_json",
		Index:      s.splunkIndex,
		Event:      event,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := ioutil.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}
