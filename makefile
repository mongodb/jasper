buildDir := build
srcFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "*\#*")
testFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -path "*\#*")

_testPackages := ./ ./jrpc ./jrpc/internal

testArgs := -v
ifneq (,$(RUN_TEST))
testArgs += -run='$(RUN_TEST)'
endif
ifneq (,$(RUN_COUNT))
testArgs += -count='$(RUN_COUNT)'
endif
ifneq (,$(SKIP_LONG))
testArgs += -short
endif


compile:
	go build $(_testPackages)
race:
	@mkdir -p $(buildDir)
	go test $(testArgs) -race $(_testPackages) | tee $(buildDir)/race.sink.out
	@grep -s -q -e "^PASS" $(buildDir)/race.sink.out && ! grep -s -q "^WARNING: DATA RACE" $(buildDir)/race.sink.out
test:
	@mkdir -p $(buildDir)
	go test $(testArgs) -cover $(_testPackages) | tee $(buildDir)/test.sink.out
	@grep -s -q -e "^PASS" $(buildDir)/test.sink.out
coverage:$(buildDir)/cover.out
	@go tool cover -func=$< | sed -E 's%github.com/.*/jasper/%%' | column -t
coverage-html:$(buildDir)/cover.html

$(buildDir):$(srcFiles) compile
	@mkdir -p $@
$(buildDir)/cover.out:$(buildDir) $(testFiles) .FORCE
	go test $(testArgs)-coverprofile $@ -cover $(_testPackages)
$(buildDir)/cover.html:$(buildDir)/cover.out
	go tool cover -html=$< -o $@
.FORCE:


proto:
	@mkdir -p jrpc/internal
	protoc --go_out=plugins=grpc:jrpc/internal *.proto
clean:
	rm *.pb.go

vendor-clean:
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/mongodb/grip/vendor/golang.org/x/sys/
	rm -rf vendor/github.com/mongodb/grip/buildscripts/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/stretchr/testify
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/pkg/errors
