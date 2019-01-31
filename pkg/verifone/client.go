package verifone

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/tiaguinho/gosoap"
)

const (
	MsgTypeGenSession = "VGGENERATESESSIONREQUEST"
)

type Config struct {
	SystemID   string
	SystemGUID string
	Passcode   string
	Url        string
}

type Client struct {
	config Config
	soap   *gosoap.Client
}

type ProcessMsgResult struct {
	MsgType string `xml:"MsgType"`
	MsgData []byte `xml:"MsgData"`
}

type ProcessMsgResponse struct {
	ProcessMsgResult ProcessMsgResult `xml:"ProcessMsgResult"`
}

type ErrorResponse struct {
	XMLName xml.Name `xml:"ERROR"`
	Code    int64    `xml:"CODE"`
	MsgText string   `xml:"MSGTXT"`
}

func NewClient(cfg Config) *Client {
	if cfg.Url == "" {
		cfg.Url = "https://txn-cst.cxmlpg.com/XML4/commideagateway.asmx?WSDL"
	}

	soap, err := gosoap.SoapClient(cfg.Url)
	if err != nil {
		panic(err)
	}

	return &Client{
		config: cfg,
		soap:   soap,
	}
}

func (this Client) getClientHeader() gosoap.Params {
	return gosoap.Params{
		"SystemGUID": this.config.SystemGUID,
		"SystemID":   this.config.SystemID,
		"Passcode":   this.config.Passcode,
	}
}

func (this Client) call(msgType string, msgData interface{}, target interface{}) (err error) {
	var body []byte
	body, err = xml.Marshal(msgData)
	if err != nil {
		return
	}

	//cdata := "<![CDATA[" + string(body) + "]]"
	cdata := string(body)

	params := gosoap.Params{
		"ClientHeader": this.getClientHeader(),
		"MsgType":      msgType,
		"MsgData":      cdata,
	}

	if err = this.soap.Call("ProcessMsg", params); err != nil {
		return
	}

	response := ProcessMsgResponse{}
	if err = this.soap.Unmarshal(&response); err != nil {
		return
	}

	// We have a response. First decode error, to see if something is up
	errResp := ErrorResponse{}
	if err = xml.Unmarshal(response.ProcessMsgResult.MsgData, &errResp); err != nil {
		return
	}

	if errResp.Code != 0 {
		return fmt.Errorf("[%d]: %s", errResp.Code, errResp.MsgText)
	}

	log.Println(string(response.ProcessMsgResult.MsgData))
	err = xml.Unmarshal(response.ProcessMsgResult.MsgData, &target)
	return
}