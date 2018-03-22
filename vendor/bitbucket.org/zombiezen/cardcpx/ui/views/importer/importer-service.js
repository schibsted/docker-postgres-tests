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

goog.provide('cardcpx.importer.Importer');

/**
 * Service for video imports.
 *
 * @param {!angular.$http} $http
 * @ngInject
 * @constructor
 */
cardcpx.importer.Importer = function($http) {
  /**
   * @type {!angular.$http}
   * @private
   */
  this.http_ = $http;
};

/**
 * Fetch the importer's current status.
 *
 * @return {!angular.$q.Promise}
 */
cardcpx.importer.Importer.prototype.getStatus = function() {
  return this.http_.get('/import').then(function(response) {
    if (response.status != 200) {
      throw "importer HTTP status " + response.status;
    }
    return response.data;
  });
};

/**
 * Call the server to start an import.
 *
 * @param {string} path directory to import from
 * @param {string} subdir subdirectory to store clips in
 * @param {Array.<Object>} items items to import
 * @return {!angular.$q.Promise}
 */
cardcpx.importer.Importer.prototype.startImport =
    function(path, subdir, items) {
  var postData = {'path': path, 'subdirectory': subdir, 'items': items};
  return this.http_.post('/import', postData).then(function(response) {
    if (response.status != 200) {
      throw "importer HTTP status " + response.status;
    }
    var data = response.data;
    if (data.code != 200) {
      throw "importer error code " + data.code + ": " + data.errorMessage;
    }
    return null;
  });
};
