package api

import (
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/oxtoacart/bpool"
  
	pb "github.com/smart-echo/micro/proto/api/v1"
	"github.com/smart-echo/micro/registry"
	"github.com/smart-echo/micro/selector"
)

var (
	// need to calculate later to specify useful defaults.
	bufferPool = bpool.NewSizedBufferPool(1024, 8)
)

func requestToProto(r *http.Request) (*pb.Request, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Error parsing form: %w", err)
	}

	req := &pb.Request{
		Path:   r.URL.Path,
		Method: r.Method,
		Header: make(map[string]*pb.Pair),
		Get:    make(map[string]*pb.Pair),
		Post:   make(map[string]*pb.Pair),
		Url:    r.URL.String(),
	}

	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		ct = "text/plain; charset=UTF-8" // default CT is text/plain
		r.Header.Set("Content-Type", ct)
	}

	// set the body:
	if r.Body != nil {
		switch ct {
		case "application/x-www-form-urlencoded":
			// expect form vals in Post data
		default:
			buf := bufferPool.Get()
			defer bufferPool.Put(buf)

			if _, err = buf.ReadFrom(r.Body); err != nil {
				return nil, err
			}

			req.Body = buf.String()
		}
	}

	// Set X-Forwarded-For if it does not exist
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior, ok := r.Header["X-Forwarded-For"]; ok {
			ip = strings.Join(prior, ", ") + ", " + ip
		}

		// Set the header
		req.Header["X-Forwarded-For"] = &pb.Pair{
			Key:    "X-Forwarded-For",
			Values: []string{ip},
		}
	}

	// Host is stripped from net/http Headers so let's add it
	req.Header["Host"] = &pb.Pair{
		Key:    "Host",
		Values: []string{r.Host},
	}

	// Get data
	for key, vals := range r.URL.Query() {
		header, ok := req.Get[key]
		if !ok {
			header = &pb.Pair{
				Key: key,
			}
			req.Get[key] = header
		}

		header.Values = vals
	}

	// Post data
	for key, vals := range r.PostForm {
		header, ok := req.Post[key]
		if !ok {
			header = &pb.Pair{
				Key: key,
			}
			req.Post[key] = header
		}

		header.Values = vals
	}

	for key, vals := range r.Header {
		header, ok := req.Header[key]
		if !ok {
			header = &pb.Pair{
				Key: key,
			}
			req.Header[key] = header
		}

		header.Values = vals
	}

	return req, nil
}

// strategy is a hack for selection.
func strategy(services []*registry.Service) selector.Strategy {
	return func(_ []*registry.Service) selector.Next {
		// ignore input to this function, use services above
		return selector.Random(services)
	}
}
