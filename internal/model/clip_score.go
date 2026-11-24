package model

type ClipScore struct {
	Hook         int    `json:"hook"`
	Clarity      int    `json:"clarity"`
	Completeness int    `json:"completeness"`
	Ending       int    `json:"ending"`
	Reason       string `json:"reason"`
	Total        int    `json:"total"`
}
