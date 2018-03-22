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

goog.provide('cardcpx.takes');
goog.provide('cardcpx.takes.TakeStorage');
goog.provide('cardcpx.takes.module');


/**
 * @typedef {{scene: string, num: string}}
 */
cardcpx.takes.Id;

/**
 * @typedef {{id: !cardcpx.takes.Id, clipName: string, select: boolean}}
 */
cardcpx.takes.Take;

/**
 * A database of takes.
 *
 * @param {!angular.$http} $http
 * @constructor
 * @ngInject
 */
cardcpx.takes.TakeStorage = function($http) {
  /**
   * @private
   * @type {!angular.$http}
   */
  this.http_ = $http;
};

/**
 * @param {?cardcpx.takes.Id} id
 * @return {!string}
 * @private
 */
cardcpx.takes.TakeStorage.prototype.url_ = function(id) {
  if (!id) {
    return '/take/';
  }
  return '/take/' + encodeURIComponent(id.scene) + '/' + encodeURIComponent(id.num);
};

/**
 * @param {!angular.$http.Response} response
 * @return {Object}
 * @private
 */
cardcpx.takes.TakeStorage.prototype.onResponse_ = function(response) {
  if (!this.isStatusOk_(response.status)) {
    throw "takes HTTP status " + response.status;
  }
  return response.data;
};

/**
 * @param {!angular.$http.Response} response
 * @private
 */
cardcpx.takes.TakeStorage.prototype.voidOnResponse_ = function(response) {
  this.onResponse_(response);
};

/**
 * @param {number} status
 * @return {boolean}
 */
cardcpx.takes.TakeStorage.prototype.isStatusOk_ = function(status) {
  return status >= 200 && status < 300;
};

/**
 * Fetch a sorted list of all takes in storage.
 * @return {angular.$q.Promise}
 */
cardcpx.takes.TakeStorage.prototype.listTakes = function() {
  return this.http_.get(this.url_(null)).
      then(angular.bind(this, this.onResponse_));
};

/**
 * Retrieve the take with the given ID.
 * @param {!cardcpx.takes.Id} id
 * @return {angular.$q.Promise}
 */
cardcpx.takes.TakeStorage.prototype.getTake = function(id) {
  return this.http_.get(this.url_(id)).
      then(angular.bind(this, this.onResponse_));
};

/**
 * Add a new take.
 * @param {!cardcpx.takes.Take} take
 * @return {angular.$q.Promise}
 */
cardcpx.takes.TakeStorage.prototype.insertTake = function(take) {
  return this.http_.post(this.url_(null), take).
      then(angular.bind(this, this.voidOnResponse_));
};

/**
 * Retrieve the take with the given ID.
 * @param {!cardcpx.takes.Id} id
 * @param {!cardcpx.takes.Take} take
 * @return {angular.$q.Promise}
 */
cardcpx.takes.TakeStorage.prototype.updateTake = function(id, take) {
  return this.http_.put(this.url_(id), take)
      .then(angular.bind(this, this.voidOnResponse_));
};

/**
 * Retrieve the take with the given ID.
 * @param {!cardcpx.takes.Id} id
 * @return {angular.$q.Promise}
 */
cardcpx.takes.TakeStorage.prototype.deleteTake = function(id) {
  return this.http_.delete(this.url_(id))
      .then(angular.bind(this, this.voidOnResponse_));
};

/** @type {!angular.Module} */
cardcpx.takes.module = angular.module('cardcpx.takes', []);

cardcpx.takes.module.service('TakeStorage', cardcpx.takes.TakeStorage);
