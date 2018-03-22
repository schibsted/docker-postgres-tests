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

goog.provide('cardcpx.navbar.NavbarCtrl');

/**
 * @param {!angular.Scope} $scope
 * @param {!angular.$route} $route
 * @constructor
 * @ngInject
 */
cardcpx.navbar.NavbarCtrl = function($scope, $route) {
  /**
   * @export
   * @type {!string}
   */
  this.active = '';

  /**
   * @export
   * @type {boolean}
   */
  this.collapsed = true;

  /**
   * @private
   * @type {angular.Scope}
   */
  this.scope = $scope;

  /**
   * @private
   * @type {!angular.$route}
   */
  this.route = $route;

  this.scope.$on('$routeChangeSuccess',
                 angular.bind(this, this.onRouteChange_));
};

/**
 * @param {!angular.Scope.Event} e the event to handle
 * @private
 */
cardcpx.navbar.NavbarCtrl.prototype.onRouteChange_ = function(e) {
  this.active = this.route.current.controller;
};

/**
 * @param {!string} controller name of the controller this item links to.
 * @export
 */
cardcpx.navbar.NavbarCtrl.prototype.itemClass = function(controller) {
  if (controller == this.active) {
    return 'active';
  } else {
    return '';
  }
};
