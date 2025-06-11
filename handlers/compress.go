package handlers

import (
	"fmt"
	"gophttp/http"
)

//go:generate stringer -type=CompressionAlgorithm
type CompressionAlgorithm string

const (
	IdentityCompression CompressionAlgorithm = "identity"
	BrotliCompression   CompressionAlgorithm = "br"
)

var compressions = map[CompressionAlgorithm]Handler{
	IdentityCompression: IdentityHandler{},
	BrotliCompression:   NewBrotliHandler(4),
}

type compressionHandler struct {
}

func (c compressionHandler) HandleRequest(ctx http.Context) error {
	//check accepted encodings and select appropriate handler(s) accordingly
	if !ctx.Request.Headers.HasHeader("Accept-Encoding") {
		return nil
	}
	header := ctx.Request.Headers["Accept-Encoding"]
	acceptedCompressions, err := http.ParseAcceptedQValues(header.Value)
	if err != nil {
		return err
	}
	bestFit := getPreferredAvailableCompression(acceptedCompressions)
	//TODO: trace logging
	fmt.Printf("chose %s for Accept-Encoding %s\n", bestFit, header.Value)
	if h, ok := compressions[bestFit]; ok {
		return h.HandleRequest(ctx)
	}
	return nil
}

func getPreferredAvailableCompression(acceptedCompressions map[string]float64) CompressionAlgorithm {
	retval := IdentityCompression
	qValue := -1.0
	for s, f := range acceptedCompressions {
		if _, ok := compressions[CompressionAlgorithm(s)]; ok && f > qValue {
			retval = CompressionAlgorithm(s)
		}
	}
	return retval
}

func NewCompressionHandler() Handler {
	return &compressionHandler{}
}
