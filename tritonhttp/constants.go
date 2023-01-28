package tritonhttp

import "time"

const (
	CONNECT_TIMEOUT time.Duration = 5 * time.Second
	SEND_TIMEOUT    time.Duration = 5 * time.Second
	RECV_TIMEOUT    time.Duration = 5 * time.Second
)
