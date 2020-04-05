package fastapi

var (
	globalMode Runmode
)

func SetMode(mode Runmode) {
	globalMode = mode
}

func GetMode() Runmode {
	return globalMode
}

var ContextType = struct {
	Text string
	JSON string
	Form string
}{
	Text: "text/plain",
	JSON: "application/json",
	Form: "application/x-www-form-urlencoded",
}
