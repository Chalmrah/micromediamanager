package main

import (
	"context"
	"fmt"
	"gopkg.in/vansante/go-ffprobe.v2"
	"time"
)

func getVideoCodec(filePath string) (string, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	probeData, err := ffprobe.ProbeURL(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("error probing file: %w", err)
	}

	stream := probeData.FirstVideoStream()
	if stream == nil {
		return "", fmt.Errorf("no video stream found in %s", filePath)
	}

	return stream.CodecName, nil
}
