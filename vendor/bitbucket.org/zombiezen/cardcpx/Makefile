# Copyright 2014 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

JSCOMPILER=java -jar build/closure-compiler.jar
ANGULAR=third_party/angular
GOTOOL?=go
GOPATH=$(CURDIR)/build/gopath

JS_MIN_LIB=\
	$(ANGULAR)/angular.min.js \
	$(ANGULAR)/angular-cookies.min.js \
	$(ANGULAR)/angular-route.min.js \
	third_party/ui-bootstrap/dist/ui-bootstrap-tpls-0.11.0.min.js

JS_LIB=\
	$(ANGULAR)/angular.js \
	$(ANGULAR)/angular-cookies.js \
	$(ANGULAR)/angular-route.js \
	third_party/ui-bootstrap/dist/ui-bootstrap-tpls-0.11.0.js

JS_SRC=\
	third_party/closure-library/closure/goog/base.js \
	ui/components/takes/takes.js \
	ui/components/filesize/filesize-filter.js \
	ui/components/filesize/filesize.js \
	ui/components/autofillbtn/autofillbtn.js \
	ui/components/allbox/allbox.js \
	ui/views/navbar/navbar-controller.js \
	ui/views/navbar/navbar.js \
	ui/views/takelist/takelist-controller.js \
	ui/views/takelist/takelist.js \
	ui/views/takeeditor/takeeditor-controller.js \
	ui/views/takeeditor/takeeditor.js \
	ui/views/importer/importer-service.js \
	ui/views/importer/importer-controller.js \
	ui/views/importer/importer.js \
	ui/app.js

GO_PACKAGE_BASES=\
	bitbucket.org/zombiezen/cardcpx \
	bitbucket.org/zombiezen/webapp \
	github.com/gorilla/context \
	github.com/gorilla/mux

all: cardcpx ui/js.js

cardcpx: *.go */*.go $(addprefix $(GOPATH)/src/, $(GO_PACKAGE_BASES))
	$(GOTOOL) build -o $@

ui/js.js: $(JS_MIN_LIB) build/compiled_js.js
	cat $^ > $@

build/compiled_js.js: $(JS_SRC) build/closure-compiler.jar | build
	$(JSCOMPILER) \
	    --angular_pass \
	    --compilation_level=ADVANCED_OPTIMIZATIONS \
	    --closure_entry_point=cardcpx.module \
	    --externs=third_party/closure-compiler/contrib/externs/angular-1.2.js \
	    --generate_exports \
	    --remove_unused_prototype_props_in_externs=false \
	    --export_local_property_definitions \
	    --js_output_file=$@ \
	    --property_renaming=OFF \
	    $(JS_SRC)

build/uncompiled_js.js: $(JS_LIB) $(JS_SRC) | build
	cat $^ > $@

$(GOPATH)/src/bitbucket.org/zombiezen/cardcpx: | $(GOPATH)/src/bitbucket.org/zombiezen
	ln -s $(CURDIR) $@

$(GOPATH)/src/bitbucket.org/zombiezen/webapp: | $(GOPATH)/src/bitbucket.org/zombiezen
	ln -s $(CURDIR)/third_party/webapp $@

$(GOPATH)/src/github.com/gorilla/context: | $(GOPATH)/src/github.com/gorilla
	ln -s $(CURDIR)/third_party/context $@

$(GOPATH)/src/github.com/gorilla/mux: | $(GOPATH)/src/github.com/gorilla
	ln -s $(CURDIR)/third_party/mux $@

$(GOPATH)/src/bitbucket.org/zombiezen: | $(GOPATH)/src
	mkdir -p $@

$(GOPATH)/src/github.com/gorilla: | $(GOPATH)/src
	mkdir -p $@

$(GOPATH)/src:
	mkdir -p $@

build/closure-compiler.jar: | build
	ant -Dcompiler-jarfile="$(CURDIR)/$@" -f third_party/closure-compiler/build.xml jar

build:
	mkdir $@

clean:
	ant -f third_party/closure-compiler/build.xml clean
	rm -rf build
	rm -f cardcpx ui/js.js

test: testgo testjs

testgo:
	$(GOTOOL) test ./httputil ./importer ./multiwriter ./natsort ./netutil ./takedata ./video

testjs:
	node_modules/karma/bin/karma start --single-run --browsers PhantomJS

.PHONY: all clean build/closure-compiler.jar test testgo testjs
