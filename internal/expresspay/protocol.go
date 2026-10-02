package expresspay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
)

func VerifySign(got string, parts ...string) bool {
	h := md5.New()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	return got == hex.EncodeToString(h.Sum(nil))
}

type Response struct {
	XMLName   xml.Name `xml:"response"`
	OsmpTxnID string   `xml:"osmp_txn_id,omitempty"`
	PrvTxn    string   `xml:"prv_txn,omitempty"`
	Sum       string   `xml:"sum,omitempty"`
	Ccy       string   `xml:"ccy,omitempty"`
	Result    int      `xml:"result"`
	Comment   string   `xml:"comment,omitempty"`
}

func (r Response) Encode() []byte {
	body, _ := xml.MarshalIndent(r, "", "  ")
	return append([]byte(xml.Header), body...)
}
