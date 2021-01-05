RELEASE_FF_BRANCH ?= $(shell cat RELEASE_FF_BRANCH)
RELEASE_MAIN_BRANCH ?= $(shell cat RELEASE_MAIN_BRANCH)

.PHONY: release-ff
## Fast-forward from RELEASE_MAIN_BRANCH branch to branch defined in RELEASE_FF_BRANCH file in the dev repo
release-ff:
	${BUILD_HARNESS_EXTENSIONS_PATH}/modules/release-ff/bin/release_ff.sh $(RELEASE_FF_BRANCH) $(RELEASE_MAIN_BRANCH)
