package api

// DeleteLeaseReq is the encoding/json compatible struct that represents the DELETE /lease request body
type DeleteLeaseReq struct {
	CloudProvider string `json:"cloud_provider"`
}
