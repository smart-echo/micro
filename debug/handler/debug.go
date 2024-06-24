// Package handler implements service debug handler embedded in micro services
package handler

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/smart-echo/micro/client"
	"github.com/smart-echo/micro/debug/log"
	"github.com/smart-echo/micro/debug/stats"
	"github.com/smart-echo/micro/debug/trace"
	pb "github.com/smart-echo/micro/proto/debug/v1"
)

// NewHandler returns an instance of the Debug Handler.
func NewHandler(c client.Client) *Debug {
	return &Debug{
		log:   log.DefaultLog,
		stats: stats.DefaultStats,
		trace: trace.DefaultTracer,
	}
}

var _ pb.DebugHandler = (*Debug)(nil)

type Debug struct {
	// must honor the debug handler
	pb.DebugHandler
	// the logger for retrieving logs
	log log.Log
	// the stats collector
	stats stats.Stats
	// the tracer
	trace trace.Tracer
}

func (d *Debug) Health(ctx context.Context, req *pb.HealthRequest, rsp *pb.HealthResponse) error {
	rsp.Status = "ok"
	return nil
}

func (d *Debug) MessageBus(ctx context.Context, stream pb.Debug_MessageBusStream) error {
	for {
		_, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}

		rsp := pb.BusMsg{
			Msg: "Request received!",
		}

		if err := stream.Send(&rsp); err != nil {
			return err
		}
	}
}

func (d *Debug) Stats(ctx context.Context, req *pb.StatsRequest, rsp *pb.StatsResponse) error {
	stats, err := d.stats.Read()
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return nil
	}

	// write the response values
	rsp.Timestamp = uint64(stats[0].Timestamp)
	rsp.Started = uint64(stats[0].Started)
	rsp.Uptime = uint64(stats[0].Uptime)
	rsp.Memory = stats[0].Memory
	rsp.Gc = stats[0].GC
	rsp.Threads = stats[0].Threads
	rsp.Requests = stats[0].Requests
	rsp.Errors = stats[0].Errors

	return nil
}

func (d *Debug) Trace(ctx context.Context, req *pb.TraceRequest, rsp *pb.TraceResponse) error {
	traces, err := d.trace.Read(trace.ReadTrace(req.Id))
	if err != nil {
		return err
	}

	for _, t := range traces {
		var typ pb.SpanType

		switch t.Type {
		case trace.SpanTypeRequestInbound:
			typ = pb.SpanType_INBOUND
		case trace.SpanTypeRequestOutbound:
			typ = pb.SpanType_OUTBOUND
		}

		rsp.Spans = append(rsp.Spans, &pb.Span{
			Trace:    t.Trace,
			Id:       t.Id,
			Parent:   t.Parent,
			Name:     t.Name,
			Started:  uint64(t.Started.UnixNano()),
			Duration: uint64(t.Duration.Nanoseconds()),
			Type:     typ,
			Metadata: t.Metadata,
		})
	}

	return nil
}

func (d *Debug) Log(ctx context.Context, req *pb.LogRequest, stream pb.Debug_LogStream) error {

	var options []log.ReadOption

	since := time.Unix(req.Since, 0)
	if !since.IsZero() {
		options = append(options, log.Since(since))
	}

	count := int(req.Count)
	if count > 0 {
		options = append(options, log.Count(count))
	}

	if req.Stream {
		// TODO: we need to figure out how to close the log stream
		// It seems like when a client disconnects,
		// the connection stays open until some timeout expires
		// or something like that; that means the map of streams
		// might end up leaking memory if not cleaned up properly
		lgStream, err := d.log.Stream()
		if err != nil {
			return err
		}
		defer lgStream.Stop()

		for record := range lgStream.Chan() {
			// copy metadata
			metadata := make(map[string]string)
			for k, v := range record.Metadata {
				metadata[k] = v
			}
			// send record
			if err := stream.Send(&pb.Record{
				Timestamp: record.Timestamp.Unix(),
				Message:   record.Message.(string),
				Metadata:  metadata,
			}); err != nil {
				return err
			}
		}

		// done streaming, return
		return nil
	}

	// get the log records
	records, err := d.log.Read(options...)
	if err != nil {
		return err
	}

	// send all the logs downstream
	for _, record := range records {
		// copy metadata
		metadata := make(map[string]string)
		for k, v := range record.Metadata {
			metadata[k] = v
		}
		// send record
		if err := stream.Send(&pb.Record{
			Timestamp: record.Timestamp.Unix(),
			Message:   record.Message.(string),
			Metadata:  metadata,
		}); err != nil {
			return err
		}
	}

	return nil
}
