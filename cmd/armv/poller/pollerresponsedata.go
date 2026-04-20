
package poller

type PollerResponseData struct {
	RespBody       []byte
	RespStatusCode int
	RespStatus     string
}

func NewPollerResponseData(respBody []byte, respStatusCode int, respStatus string) PollerResponseData {

	return PollerResponseData{
		RespBody:       respBody,
		RespStatusCode: respStatusCode,
		RespStatus:     respStatus,
	}
}
