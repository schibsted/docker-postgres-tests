/*
  @license
  Copyright 2014 Google Inc. All rights reserved.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

goog.provide('cardcpx');
goog.provide('cardcpx.module');

goog.require('cardcpx.importer.module');
goog.require('cardcpx.navbar.module');
goog.require('cardcpx.takeeditor.module');
goog.require('cardcpx.takelist.module');

/**
 * @param {!angular.$locationProvider} $locationProvider
 * @ngInject
 * @private
 */
cardcpx.config_ = function($locationProvider) {
  $locationProvider.
      html5Mode(true).
      hashPrefix('!');
};

/** @type {!angular.Module} */
cardcpx.module = angular.module('cardcpx', [
    cardcpx.importer.module.name,
    cardcpx.navbar.module.name,
    cardcpx.takeeditor.module.name,
    cardcpx.takelist.module.name]).config(cardcpx.config_);
