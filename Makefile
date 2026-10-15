# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := pre-pr
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: pre-pr
pre-pr: backend-pre-pr app-pre-pr

.PHONY: backend-pre-pr
backend-pre-pr:
	@$(MAKE) -C backend pre-pr

.PHONY: app-pre-pr
app-pre-pr:
	@npm --prefix app ci
	@npm --prefix app run check
