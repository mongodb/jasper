name := jasper
buildDir := build
srcFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "*\#*")
testFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -path "*\#*")
packages := $(name) cli rpc rest mock options

_testPackages := $(subst $(name),,$(foreach target,$(packages),./$(target)))
coverageOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage)
coverageHtmlOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage.html)

benchPattern := ./

.PHONY: benchmark
benchmark:
	@mkdir -p $(buildDir)
	go test $(testArgs) -bench=$(benchPattern) $(if $(RUN_TEST),, -run=^^$$) | tee $(buildDir)/bench.sink.out
	@grep -s -q -e "^PASS" $(buildDir)/bench.sink.out
compile:
	go build $(_testPackages)
compile-base:
	go build ./
build:compile


# convenience targets for runing tests and coverage tasks on a
# specific package.
test-%:$(buildDir)/output.%.test
	@grep -s -q -e "^PASS" $<
coverage-%:$(buildDir)/output.%.coverage
	@grep -s -q -e "^PASS" $(buildDir)/output.$*.test
html-coverage-%:$(buildDir)/output.%.coverage.html
	@grep -s -q -e "^PASS" $(buildDir)/output.$*.test
lint-%:$(buildDir)/output.%.lint
	@grep -v -s -q "^--- FAIL" $<
# end convienence targets


# start linting configuration
#   package, testing, and linter dependencies specified
#   separately. This is a temporary solution: eventually we should
#   vendorize all of these dependencies.
lintDeps := github.com/alecthomas/gometalinter
#   include test files and give linters 40s to run to avoid timeouts
lintArgs := --tests --deadline=13m --vendor
#   gotype produces false positives because it reads .a files which
#   are rarely up to date.
lintArgs += --disable="gotype" --disable="gosec" --disable="gocyclo" --enable="golint"
lintArgs += --skip="build"
#   enable and configure additional linters
lintArgs += --line-length=100 --dupl-threshold=150
#   some test cases are structurally similar, and lead to dupl linter
#   warnings, but are important to maintain separately, and would be
#   difficult to test without a much more complex reflection/code
#   generation approach, so we ignore dupl errors in tests.
lintArgs += --exclude="warning: duplicate of .*_test.go"
#   go lint warns on an error in docstring format, erroneously because
#   it doesn't consider the entire package.
lintArgs += --exclude="warning: package comment should be of the form \"Package .* ...\""
#   known issues that the linter picks up that are not relevant in our cases
lintArgs += --exclude="file is not goimported" # top-level mains aren't imported
lintArgs += --exclude="error return value not checked .defer.*"
# end linting configuration


# start test and coverage artifacts
#    This varable includes everything that the tests actually need to
#    run. (The "build" target is intentional and makes these targetsb
#    rerun as expected.)
testTimeout := -timeout=20m
testArgs := -v $(testTimeout)
ifneq (,$(RUN_TEST))
testArgs += -run='$(RUN_TEST)'
endif
ifneq (,$(RUN_COUNT))
testArgs += -count=$(RUN_COUNT)
endif
ifneq (,$(SKIP_LONG))
testArgs += -short
endif
ifneq (,$(DISABLE_COVERAGE))
testArgs += -cover
endif
ifneq (,$(RACE_DETECTOR))
testArgs += -race
endif
ifneq (,$(SKIP_LONG))
testArgs += -short
endif
# test execution and output handlers
$(buildDir)/:
	mkdir -p $@

$(buildDir)/output.%.test:$(buildDir)/ .FORCE
	go test $(testArgs) ./$(if $(subst $(name),,$*),$(subst -,/,$*),) | tee $@
	@! grep -s -q -e "^FAIL" $@ && ! grep -s -q "^WARNING: DATA RACE" $@
$(buildDir)/output.test:$(buildDir)/ .FORCE
	go test $(testArgs) ./... | tee $@
	@! grep -s -q -e "^FAIL" $@ && ! grep -s -q "^WARNING: DATA RACE" $@
$(buildDir)/output.%.coverage:$(buildDir)/ .FORCE
	go test $(testArgs) ./$(if $(subst $(name),,$*),$(subst -,/,$*),) -covermode=count -coverprofile $@ | tee $(buildDir)/output.$*.test
	@-[ -f $@ ] && go tool cover -func=$@ | sed 's%$(projectPath)/%%' | column -t
$(buildDir)/output.%.coverage.html:$(buildDir)/output.%.coverage
	go tool cover -html=$< -o $@
#  targets to generate gotest output from the linter.
lintDeps := $(addprefix $(GOPATH)/src/,$(lintDeps))
$(buildDir)/run-linter:cmd/run-linter/run-linter.go $(buildDir)/.lintSetup
	 go build -o $@ $<
$(buildDir)/.lintSetup:$(lintDeps)
	@-$(GOPATH)/bin/gometalinter --install >/dev/null && touch $@
$(buildDir)/output.%.lint:$(buildDir)/run-linter $(buildDir)/ .FORCE
	@./$< --output=$@ --lintArgs='$(lintArgs)' --packages='$*'
$(buildDir)/output.lint:$(buildDir)/run-linter $(buildDir)/ .FORCE
	@./$< --output="$@" --lintArgs='$(lintArgs)' --packages="$(packages)"
#  targets to process and generate coverage reports
# end test and coverage artifacts


# userfacing targets for basic build and development operations
proto:
	@mkdir -p rpc/internal
	protoc --go_out=plugins=grpc:rpc/internal *.proto
lint:$(foreach target,$(packages),$(buildDir)/output.$(target).lint)
test:build $(buildDir)/output.test
coverage:build $(coverageOutput)
coverage-html:build $(coverageHtmlOutput)
phony += lint lint-deps build build-race race test coverage coverage-html list-race list-tests
.PRECIOUS:$(coverageOutput) $(coverageHtmlOutput)
.PRECIOUS:$(foreach target,$(packages),$(buildDir)/output.$(target).test)
.PRECIOUS:$(foreach target,$(packages),$(buildDir)/output.$(target).lint)
.PRECIOUS:$(buildDir)/output.lint
# end front-ends

.FORCE:

clean:
	rm -rf *.pb.go

vendor-clean:
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/urfave/cli/
	rm -rf vendor/github.com/evergreen-ci/aviation/vendor/github.com/evergreen-ci/gimlet/
	rm -rf vendor/github.com/evergreen-ci/aviation/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/evergreen-ci/aviation/vendor/github.com/pkg/errors/
	rm -rf vendor/github.com/evergreen-ci/aviation/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/evergreen-ci/aviation/vendor/google.golang.org/grpc/
	rm -rf vendor/github.com/mongodb/amboy/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/mongodb/amboy/vendor/github.com/pkg/errors
	rm -rf vendor/github.com/mongodb/amboy/vendor/github.com/evergreen-ci/gimlet
	rm -rf vendor/github.com/mongodb/amboy/vendor/golang.org/x/tools
	rm -rf vendor/github.com/mongodb/amboy/vendor/github.com/urfave/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/go.mongodb.org/mongo-driver/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/gopkg.in/mgo.v2/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/github.com/pkg/errors/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/github.com/satori/go.uuid/
	rm -rf vendor/github.com/mongodb/ftdc/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/montanaflynn/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/pkg/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/mongodb/grip/vendor/golang.org/x/sys/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/pkg/
	rm -rf vendor/github.com/mongodb/grip/buildscripts/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/pkg/errors/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/mholt/archiver/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/mongodb/amboy/
	rm -rf vendor/github.com/tychoish/bond/vendor/github.com/satori/go.uuid/
	rm -rf vendor/github.com/tychoish/lru/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/tychoish/lru/vendor/github.com/pkg/errors/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/pkg/
	rm -rf vendor/github.com/mholt/archiver/rar.go
	rm -rf vendor/github.com/mholt/archiver/tarbz2.go
	rm -rf vendor/github.com/mholt/archiver/tarlz4.go
	rm -rf vendor/github.com/mholt/archiver/tarsz.go
	rm -rf vendor/github.com/mholt/archiver/tarxz.go
	find vendor/ -name "*.gif" -o -name "*.gz" -o -name "*.png" -o -name "*.ico" -o -name "*testdata*" | xargs rm -rf
	find vendor -type d -empty | xargs rm -rf
