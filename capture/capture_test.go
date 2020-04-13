package capture

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapture(t *testing.T) {
	dc := &DistractConfig{}
	_ = dc.setup()

	// nolint lll
	s := `2020-04-08 00:58:55,249  INFO 20478 --- [http-nio-10902-exec-647] [a77eb416c9604ca48dc7d3011ef33ab1] c.o.b.m.v.a.w.i.MonitorLogger            : AuthenticationController.customerVerify(..)[End]:1586278735249^_^AuthenticationController.customerVerify(..)=[{"operationDate":"2020-04-07 09:38:20","signAlgo":"HmacSHA256","signature":"NKSAdVYTiaPOqXcsrfM=","appId":"APP_949D6F949DEA1","api-version-id":"30628","operationName":"证书验证","platformId":"3","platformName":"云签章平台","deviceId":"DEV_B2B9180228184BF7A3B803FF1EA2297E","version":"1.0","operationResult":"成功"}]^_^{"message":"应用ID不存在","status":90002903}^_^0^_^false`

	c, err := dc.CaptureString([]string{s})
	assert.Nil(t, err)
	assert.Equal(t, s, c)
	assert.True(t, dc.IsEmpty())

	dc = &DistractConfig{
		SplitSeq:     3,
		Capture:      "",
		CaptureGroup: 0,
		AnchorStart:  "",
		AnchorEnd:    "",
		captureReg:   nil,
	}
	_ = dc.setup()

	c, err = dc.CaptureString(strings.SplitN(s, `^_^`, -1))
	assert.Nil(t, err)
	assert.Equal(t, `{"message":"应用ID不存在","status":90002903}`, c)

	dc = &DistractConfig{
		SplitSeq:     2,
		Capture:      "",
		CaptureGroup: 0,
		AnchorStart:  ")=[",
		AnchorEnd:    "]",
		captureReg:   nil,
	}
	_ = dc.setup()

	c, err = dc.CaptureString(strings.SplitN(s, `^_^`, -1))
	assert.Nil(t, err)
	// nolint lll
	assert.Equal(t, `{"operationDate":"2020-04-07 09:38:20","signAlgo":"HmacSHA256","signature":"NKSAdVYTiaPOqXcsrfM=","appId":"APP_949D6F949DEA1","api-version-id":"30628","operationName":"证书验证","platformId":"3","platformName":"云签章平台","deviceId":"DEV_B2B9180228184BF7A3B803FF1EA2297E","version":"1.0","operationResult":"成功"}`, c)

	dc = &DistractConfig{
		SplitSeq:     2,
		Capture:      "",
		CaptureGroup: 0,
		AnchorStart:  ")=[",
		AnchorEnd:    "",
		Cut:          ":-1",
		captureReg:   nil,
	}
	_ = dc.setup()

	c, err = dc.CaptureString(strings.SplitN(s, `^_^`, -1))
	assert.Nil(t, err)
	// nolint lll
	assert.Equal(t, `{"operationDate":"2020-04-07 09:38:20","signAlgo":"HmacSHA256","signature":"NKSAdVYTiaPOqXcsrfM=","appId":"APP_949D6F949DEA1","api-version-id":"30628","operationName":"证书验证","platformId":"3","platformName":"云签章平台","deviceId":"DEV_B2B9180228184BF7A3B803FF1EA2297E","version":"1.0","operationResult":"成功"}`, c)
}
