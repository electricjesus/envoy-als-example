package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"

	envoyaccesslogdata "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	envoyaccesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	PORT    = ":8080"
	VERSION = "0.1.0"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Infof("Starting access log service version %s", VERSION)

	gs := grpc.NewServer()
	accessLogService := &accessLogserver{}
	accessLogService.Register(gs)

	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", PORT, err)
	}

	log.Infof("Start: access log service on port %s", PORT)
	go func() {
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()
	<-ctx.Done()

	log.Infof("Shutting down access log service on port %s", PORT)
	gs.GracefulStop()
	log.Info("Access log service stopped")
	if err := lis.Close(); err != nil {
		log.Errorf("Failed to close listener: %v", err)
	}
	log.Info("Listener closed, exiting")
	os.Exit(0)
}

type accessLogserver struct{}

var (
	_ envoyaccesslog.AccessLogServiceServer = &accessLogserver{}
)

func (s *accessLogserver) StreamAccessLogs(stream envoyaccesslog.AccessLogService_StreamAccessLogsServer) error {
	log.Info("Received new access log stream")

	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Receive the access log data from the stream
		logEntry, err := stream.Recv()
		if err == io.EOF {
			log.Info("Access log stream closed by client")
			return nil
		}
		if err != nil {
			log.Infof("Error receiving access log entry: %v", err)
			return err
		}

		log.Infof("Received access log entry: %v", logEntry)

		switch logEntry.LogEntries.(type) {
		case *envoyaccesslog.StreamAccessLogsMessage_HttpLogs:
			// Handle HTTP logs
			s.handleHttpLogs(logEntry)
		case *envoyaccesslog.StreamAccessLogsMessage_TcpLogs:
			// Handle TCP logs
			s.handleTcpLogs(logEntry)
		default:
			log.Warnf("Received unsupported log entry type: %T", logEntry.LogEntries)
		}
	}
}

// Process HTTP logs from the access log stream
func (s *accessLogserver) handleHttpLogs(logEntry *envoyaccesslog.StreamAccessLogsMessage) {
	log.Info("Processing HTTP access logs")
	for _, log := range logEntry.GetHttpLogs().LogEntry {
		s.handleHttpLogEntry(log)
	}
}

// Process a single HTTP access log entry
func (s *accessLogserver) handleHttpLogEntry(logEntry *envoyaccesslogdata.HTTPAccessLogEntry) {
	log.Printf("Received HTTP access log entry: %v", logEntry)
}

func (s *accessLogserver) handleTcpLogs(logEntry *envoyaccesslog.StreamAccessLogsMessage) {
	log.Info("Processing TCP access logs")
	for _, log := range logEntry.GetTcpLogs().LogEntry {
		s.handleTcpLogEntry(log)
	}
}

func (s *accessLogserver) handleTcpLogEntry(logEntry *envoyaccesslogdata.TCPAccessLogEntry) {
	log.Printf("Received TCP access log entry: %v", logEntry)
}

func (s *accessLogserver) Register(grpcServer *grpc.Server) {
	envoyaccesslog.RegisterAccessLogServiceServer(grpcServer, s)
	log.Info("Access log service registered")
}
