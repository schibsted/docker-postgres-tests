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

goog.provide('cardcpx.takelist');
goog.provide('cardcpx.takelist.module');

goog.require('cardcpx.takelist.TakeListCtrl');
goog.require('cardcpx.takes.module');

/**
 * @param {!angular.$routeProvider} $routeProvider
 * @ngInject
 * @private
 */
cardcpx.takelist.config_ = function($routeProvider) {
  $routeProvider.when('/', {
      templateUrl: '/ui/views/takelist/takelist.html',
      controller: 'TakeListCtrl',
      controllerAs: 'takeList'});
};

/** @type {!angular.Module} */
cardcpx.takelist.module = angular.module('cardcpx.takelist', [
    'ngRoute',
    'ui.bootstrap',
    cardcpx.takes.module.name]);

cardcpx.takelist.module.config(cardcpx.takelist.config_);

cardcpx.takelist.module.controller(
    'TakeListCtrl', cardcpx.takelist.TakeListCtrl);
