/*
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

goog.provide('cardcpx.importer');
goog.provide('cardcpx.importer.module');

goog.require('cardcpx.allbox.module');
goog.require('cardcpx.autofillbtn.module');
goog.require('cardcpx.filesize.module');
goog.require('cardcpx.importer.Importer');
goog.require('cardcpx.importer.ImporterCtrl');
goog.require('cardcpx.takes.module');

/**
 * @ngInject
 * @private
 */
cardcpx.importer.config_ = function($routeProvider) {
  $routeProvider.when('/importer', {
      templateUrl: '/ui/views/importer/importer.html',
      controller: 'ImporterCtrl',
      controllerAs: 'importer'});
};

/** @type {!angular.Module} */
cardcpx.importer.module = angular.module('cardcpx.importer', [
    'ngCookies',
    'ngRoute',
    'ui.bootstrap',
    cardcpx.allbox.module.name,
    cardcpx.autofillbtn.module.name,
    cardcpx.filesize.module.name,
    cardcpx.takes.module.name]);

cardcpx.importer.module.controller(
    'ImporterCtrl', cardcpx.importer.ImporterCtrl);

cardcpx.importer.module.service('Importer', cardcpx.importer.Importer);

cardcpx.importer.module.config(cardcpx.importer.config_);
