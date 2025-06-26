package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	otelaccesslog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

var _ otelaccesslog.LogsServiceServer = &accessLogserver{}

func (s *accessLogserver) Export(context context.Context, req *otelaccesslog.ExportLogsServiceRequest) (*otelaccesslog.ExportLogsServiceResponse, error) {
	log.Info("Receiving export logs request")
	for _, entry := range req.GetResourceLogs() {
		log.Infof("Log Record: %s", entry.String())
		for _, logRecord := range entry.GetScopeLogs() {
			log.WithFields(log.Fields{
				"scope":      logRecord.GetScope().String(),
				"num_logs":   len(logRecord.GetLogRecords()),
				"schema_url": logRecord.GetSchemaUrl(),
				"_record":    logRecord.String(),
			}).Infof("scope log")
			for _, rec := range logRecord.GetLogRecords() {
				log.Infof("Log Record: %s", rec.String())
			}
		}
	}
	return &otelaccesslog.ExportLogsServiceResponse{}, nil
}
