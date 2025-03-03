package logger

const (
	LogHost       = "host"
	LogPort       = "port"
	LogAddr       = "addr"
	LogRemote     = "remote" // aligned IPv4:Port "   192.168.0.42:1234 "
	LogFunc       = "func"   // RPC method name, REST resource path
	LogHTTPMethod = "httpMethod"
	LogHTTPStatus = "httpStatus"
	LogGRPCCode   = "grpcCode"
	LogEvent      = "ev"
	LogEventID    = "evID"
)
