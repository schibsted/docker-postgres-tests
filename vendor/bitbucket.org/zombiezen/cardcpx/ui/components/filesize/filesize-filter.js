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

goog.provide('cardcpx.filesize.fileSizeString');
goog.provide('cardcpx.filesize.fileSizeFilter');

/**
 * Convert bytes into a human-readable string like "1.5 MiB".
 *
 * @param {number} input number of bytes
 * @return {string} the human-readable string
 */
cardcpx.filesize.fileSizeString = function(input) {
  if (input == 1) {
    return input + ' byte';
  } else if (input < 1024) {
    return input.toFixed(0) + ' bytes';
  } else if (input < 1024 * 1024) {
    return (input / 1024).toFixed(1) + ' KiB';
  } else if (input < 1024 * 1024 * 1024) {
    return (input / 1024 / 1024).toFixed(1) + ' MiB';
  } else if (input < 1024 * 1024 * 1024 * 1024) {
    return (input / 1024 / 1024 / 1024).toFixed(1) + ' GiB';
  } else if (input < 1024 * 1024 * 1024 * 1024 * 1024) {
    return (input / 1024 / 1024 / 1024 / 1024).toFixed(1) + ' TiB';
  } else {
    return (input / 1024 / 1024 / 1024 / 1024 / 1024).toFixed(1) + ' PiB';
  }
};

/** @type {Function} */
cardcpx.filesize.fileSizeFilter = function() {
  return function(input) {
    if (!goog.isNumber(input)) {
      return '';
    }
    return cardcpx.filesize.fileSizeString(input);
  };
};
