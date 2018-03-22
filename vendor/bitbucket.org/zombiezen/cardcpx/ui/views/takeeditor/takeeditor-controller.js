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

goog.provide('cardcpx.takeeditor.TakeEditorCtrl');

/**
 * List of takes.
 *
 * @param {!TakeStorage} TakeStorage
 * @param {!angular.$location} $location
 * @param {!Object} $routeParams
 * @param {!Object} take a take object fetched from the server
 * @constructor
 * @ngInject
 */
cardcpx.takeeditor.TakeEditorCtrl = function(
    TakeStorage, $location, $routeParams, take) {
  /**
   * @private
   * @type {!cardcpx.takes.TakeStorage}
   */
  this.storage_ = TakeStorage;

  /**
   * @private
   * @type {!angular.$location}
   */
  this.location_ = $location;

  /**
   * @export
   * @type {!cardcpx.takes.Id}
   */
  this.id = {scene: $routeParams['scene'], num: $routeParams['num']};

  /**
   * @export
   * @type {!cardcpx.takes.Take}
   */
  this.take = take;
};

/**
 * @export
 */
cardcpx.takeeditor.TakeEditorCtrl.prototype.save = function() {
  this.storage_.updateTake(this.id, this.take).
      then(angular.bind(this, function() {
        this.goToTakeList_();
      }));
};

/**
 * @export
 */
cardcpx.takeeditor.TakeEditorCtrl.prototype.destroy = function() {
  this.storage_.deleteTake(this.id).
      then(angular.bind(this, function() {
        this.goToTakeList_();
      }));
};

/**
 * @private
 */
cardcpx.takeeditor.TakeEditorCtrl.prototype.goToTakeList_ = function() {
  this.location_.url('/');
};
