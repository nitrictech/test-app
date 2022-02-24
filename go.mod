module github.com/asalkeld/test-app

go 1.16

require (
	github.com/google/uuid v1.3.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/nitrictech/go-sdk v0.8.1-rc.3.0.20220208211200-037427295f12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
)

replace github.com/nitrictech/go-sdk => github.com/asalkeld/go-sdk v0.8.1-0.20220228022220-78e8cfb0c5c0
