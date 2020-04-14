package capture

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint lll
func TestFilter(t *testing.T) {
	filtersStr := "contains AuController [End] | split by=^_^ keeps=1 | anchor start=[ | cut :-1"
	filters := ParseFilters(filtersStr)

	s := `2020-04-08 00:58:55,249  INFO 20478 --- [http-nio-10902-exec-647] c.o.MonitorLogger : AuController.customerVerify(..)[End]:1586278735249^_^AuController.customerVerify(..)=[{"operationDate":"2020-04-07 09:38:20","signature":"NKSAd","appId":"APP_9","platformId":"3","deviceId":"DEV_B2B","version":"1.0","operationResult":"成功"}]^_^{"message":"应用ID不存在","status":90002903}^_^0^_^false`
	assert.Equal(t, []string{
		`{"operationDate":"2020-04-07 09:38:20","signature":"NKSAd","appId":"APP_9","platformId":"3","deviceId":"DEV_B2B","version":"1.0","operationResult":"成功"}`},
		filters.Filter([]string{s}))
}
