package main

import (
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func CreateClient(accessKeyId *string, accessKeySecret *string) (_client *alidns20150109.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	config.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	_client = &alidns20150109.Client{}
	_client, _err = alidns20150109.NewClient(config)
	return _client, _err
}

func main() {
	e := echo.New()
	e.GET("/ddns", func(c echo.Context) error {
		accessKeyId := c.QueryParam("accessKeyId")
		accessKeySecret := c.QueryParam("accessKeySecret")
		domain := c.QueryParam("domain")
		if accessKeyId == "" || accessKeySecret == "" || domain == "" {
			return c.String(http.StatusInternalServerError, "请检查请求参数")
		}
		client, err := CreateClient(tea.String(accessKeyId), tea.String(accessKeySecret))
		if err != nil {
			return c.String(http.StatusInternalServerError, "请检查accessKeyId和accessKeySecret")
		}

		domains := strings.Split(domain, ".")
		if len(domains) != 3 {
			return c.String(http.StatusInternalServerError, "请检查域名")
		}
		domainName := domains[1] + "." + domains[2]

		describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{}
		describeDomainRecordsRequest.SetDomainName(domainName)
		describeDomainRecordsResponse, err := client.DescribeDomainRecords(describeDomainRecordsRequest)
		if err != nil {
			return err
		}

		recordId := ""

		for _, record := range describeDomainRecordsResponse.Body.DomainRecords.Record {
			if *record.RR == domains[0] {
				recordId = *record.RecordId
				break
			}
		}

		if recordId == "" {
			addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{}
			addDomainRecordRequest.SetDomainName(domainName)
			addDomainRecordRequest.SetRR(domains[0])
			addDomainRecordRequest.SetValue(echo.ExtractIPDirect()(c.Request()))
			addDomainRecordRequest.SetType("A")
			addDomainRecordResponse, err := client.AddDomainRecord(addDomainRecordRequest)
			if err != nil {
				return err
			}
			return c.String(http.StatusOK, *addDomainRecordResponse.Body.RecordId)
		} else {
			updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{}
			updateDomainRecordRequest.SetRecordId(recordId)
			updateDomainRecordRequest.SetRR(domains[0])
			updateDomainRecordRequest.SetValue(echo.ExtractIPDirect()(c.Request()))
			updateDomainRecordRequest.SetType("A")
			updateDomainRecordResponse, err := client.UpdateDomainRecord(updateDomainRecordRequest)
			if err != nil {
				return err
			}
			return c.String(http.StatusOK, *updateDomainRecordResponse.Body.RecordId)
		}
	})
	e.Logger.Fatal(e.Start(":8000"))
}
