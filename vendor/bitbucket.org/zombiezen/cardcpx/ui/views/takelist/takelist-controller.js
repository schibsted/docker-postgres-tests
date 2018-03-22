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

goog.provide('cardcpx.takelist.TakeListCtrl');

/**
 * List of takes.
 *
 * @param {cardcpx.takes.TakeStorage} TakeStorage
 * @constructor
 * @ngInject
 */
cardcpx.takelist.TakeListCtrl = function(TakeStorage) {
  /**
   * @export
   * @type {Array.<cardcpx.takes.Take>}
   */
  this.takes = [];

  /**
   * @export
   * @type {Array.<{scene: string, takes: Array.<cardcpx.takes.Take>}>}
   */
  this.scenes = null;

  TakeStorage.listTakes().then(angular.bind(this, this.onTakesSuccess_));
};

/**
 * @param {Array.<cardcpx.takes.Take>} takes
 * @private
 */
cardcpx.takelist.TakeListCtrl.prototype.onTakesSuccess_ = function(takes) {
  this.takes = takes;
  this.scenes = [];
  var currScene = null;
  angular.forEach(this.takes, function(take) {
    if (!currScene || take.id.scene != currScene.scene) {
      currScene = {scene: take.id.scene, takes: []};
      this.scenes.push(currScene);
    }
    currScene.takes.push(take);
  }, this);
};
