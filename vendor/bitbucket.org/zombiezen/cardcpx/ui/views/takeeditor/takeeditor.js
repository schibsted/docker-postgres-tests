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

goog.provide('cardcpx.takeeditor');
goog.provide('cardcpx.takeeditor.module');

goog.require('cardcpx.takeeditor.TakeEditorCtrl');
goog.require('cardcpx.takes.module');

/**
 * @param {!angular.$routeProvider} $routeProvider
 * @ngInject
 * @private
 */
cardcpx.takeeditor.config_ = function($routeProvider) {
  $routeProvider.when('/take/:scene/:num', {
    templateUrl: '/ui/views/takeeditor/takeeditor.html',
    controller: 'TakeEditorCtrl',
    controllerAs: 'takeEditor',
    resolve: {
      take: ['TakeStorage', '$route', function(TakeStorage, $route) {
        var params = $route.current.params;
        return TakeStorage.getTake({
          scene: params['scene'],
          num: params['num']});
      }]
    }
  });
};

/** @type {!angular.Module} */
cardcpx.takeeditor.module = angular.module('cardcpx.takeeditor', [
    'ngRoute',
    'ui.bootstrap',
    cardcpx.takes.module.name]);

cardcpx.takeeditor.module.config(cardcpx.takeeditor.config_);

cardcpx.takeeditor.module.controller(
    'TakeEditorCtrl', cardcpx.takeeditor.TakeEditorCtrl);
