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

goog.provide('cardcpx.importer.ImporterCtrl');
goog.provide('cardcpx.importer.ImportState');
goog.provide('cardcpx.importer.ItemState');

/**
 * @enum
 */
cardcpx.importer.ImportState = {
  LOADING: 'loading',
  NO_DATA: 'noData',
  READY: 'ready',
  ACTIVE: 'active'
};

/**
 * Status of an item.
 * @enum
 */
cardcpx.importer.ItemState = {
  READY: 'ready',
  ERROR: 'error',
  ACTIVE: 'active',
  IMPORTED: 'imported'
};

/**
 * Number of milliseconds between status polls.
 *
 * @private
 * @const
 */
cardcpx.importer.ImporterCtrl.POLL_TIME_ = 500;

/**
 * Importer UI controller.
 *
 * @param {!cardcpx.importer.Importer} Importer
 * @param {!cardcpx.takes.TakeStorage} TakeStorage
 * @param {!Object} $routeParams
 * @param {!angular.$location} $location
 * @param {!angular.$http} $http
 * @param {!angular.$log} $log
 * @param {!angular.$timeout} $timeout
 * @param {!angular.$q} $q
 * @param {!angular.$cookies} $cookies
 * @constructor
 * @ngInject
 */
cardcpx.importer.ImporterCtrl = function(
    Importer, TakeStorage, $routeParams, $location,
    $http, $log, $timeout, $q, $cookies) {
  /**
   * @export
   * @type {!string}
   */
  this.path = $routeParams['path'];

  /**
   * @export
   * @type {!string}
   */
  this.newPath = this.path;

  if (!this.newPath && $cookies['lastImportPath']) {
    this.newPath = $cookies['lastImportPath'];
  }

  /**
   * @export
   * @type {!string}
   */
  this.subdir = '';

  /**
   * Whether or not an import is allowed.
   * @export
   * @type {boolean}
   */
  this.enabled = true;

  /**
   * @private
   * @type {boolean}
   */
  this.loading_ = true;

  /**
   * @export
   * @type {number}
   */
  this.progress = -1;

  /**
   * @export
   * @type {?Date}
   */
  this.eta = null;

  /**
   * @export
   * @type {Array.<Object>}
   */
  this.items = [];

  /**
   * @private
   * @type {!cardcpx.importer.Importer}
   */
  this.importer_ = Importer;

  /**
   * @private
   * @type {!cardcpx.takes.TakeStorage}
   */
  this.takeStorage_ = TakeStorage;

  /**
   * @private
   * @type {!angular.$location}
   */
  this.location_ = $location;

  /**
   * @private
   * @type {!angular.$http}
   */
  this.http_ = $http;

  /**
   * @private
   * @type {!angular.$log}
   */
  this.log_ = $log;

  /**
   * @private
   * @type {!angular.$timeout}
   */
  this.timeout_ = $timeout;

  /**
   * @private
   * @type {!angular.$q}
   */
  this.q_ = $q;

  /**
   * @private
   * @type {!angular.$cookies}
   */
  this.cookies_ = $cookies;

  this.load_();
};

/**
 * Open a new source in the importer.
 *
 * @param {!string} path the new path to open
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.openPath = function(path) {
  this.location_.search('path', path);
  this.loading_ = true;
  this.cookies_['lastImportPath'] = path;
};

/**
 * Returns the importer's state.
 * @return {cardcpx.importer.ImportState}
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.getState = function() {
  if (this.loading_) {
    return cardcpx.importer.ImportState.LOADING;
  } else if (this.items.length == 0) {
    return cardcpx.importer.ImportState.NO_DATA;
  } else if (this.progress >= 0) {
    return cardcpx.importer.ImportState.ACTIVE;
  } else {
    return cardcpx.importer.ImportState.READY;
  }
};

/**
 * Return whether there is an active import in-progress.
 * @return {boolean}
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.isActive = function() {
  return this.getState() == cardcpx.importer.ImportState.ACTIVE;
};

/**
 * @return {boolean}
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.anyChecked = function() {
  for (var i = 0; i < this.items.length; i++) {
    if (this.items[i].checked) {
      return true;
    }
  }
  return false;
};

/**
 * @return {boolean}
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.allChecked = function() {
  if (this.items.length == 0) {
    return false;
  }
  for (var i = 0; i < this.items.length; i++) {
    if (!this.items[i].checked) {
      return false;
    }
  }
  return true;
};

/**
 * Mark all items for importing.
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.checkAll = function() {
  angular.forEach(this.items, function(item) {
    item.checked = true;
  });
};

/**
 * Unmark all items for importing.
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.checkNone = function() {
  angular.forEach(this.items, function(item) {
    item.checked = false;
  });
};

/**
 * Load the clips from the server asynchronously.
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.load_ = function() {
  this.loading_ = true;
  this.refreshStatus_();
  if (this.path) {
    var sourcePromise = this.http_.get('/source', {params: {path: this.path}});
    var listPromise = this.takeStorage_.listTakes();

    this.q_.all({clips: sourcePromise, takes: listPromise}).
        then(angular.bind(this, function(val) {
          var clips = val.clips.data;
          var takes = val.takes;
          this.items = [];
          angular.forEach(clips, function(clip, i) {
            var state = cardcpx.importer.ItemState.READY;
            var checked = true;
            if (this.hasClip_(clip.name, takes)) {
              state = cardcpx.importer.ItemState.IMPORTED;
              checked = false;
            }
            this.items.push({
                checked: checked,
                state: state,
                clip: clip,
                scene: 'scratch',
                num: i + 1 + '',
                select: false});
          }, this);
          this.loading_ = false;
        }), angular.bind(this, function(data) {
          this.items = [];
          this.loading_ = false;
          this.log_.error('source load failed: ' + data);
        }));
  } else {
    this.loading_ = false;
  }
};

/**
 * @param {string} name
 * @param {Array.<Object>} takes
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.hasClip_ = function(name, takes) {
  for (var i = 0; i < takes.length; i++) {
    if (takes[i].clipName == name) {
      return true;
    }
  }
  return false;
};

/**
 * Copy the scene from an item to all of the subsequent items in the list.
 *
 * @param {number} i the index of the item to copy scene from
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.autofillScene = function(i) {
  if (i < 0 || i >= this.items.length) {
    return;
  }
  var scene = this.items[i].scene;
  for (var j = i + 1; j < this.items.length; j++) {
    this.items[j].scene = scene;
  }
};

/**
 * Set the take numbers of subsequent items to increasing numbers.
 *
 * @param {number} i the index of the item to start numbering from
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.autofillNum = function(i) {
  if (i < 0 || i >= this.items.length) {
    return;
  }
  var num = parseInt(this.items[i].num, 10);
  if (isNaN(num)) {
    return;
  }
  for (var j = i + 1; j < this.items.length; j++) {
    this.items[j].num = (num + j - i) + '';
  }
};

/**
 * Refresh import status.
 * @return {!angular.$q.Promise}
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.refreshStatus_ = function() {
  return this.importer_.getStatus().then(angular.bind(this, this.onStatus_));
};

/**
 * Handle new server status.
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.onStatus_ = function(st) {
  this.enabled = st.enabled;
  if (st.active) {
    this.progress = st.bytesCopied / st.bytesTotal;
    this.eta = Date.parse(st.eta);
  } else {
    this.progress = -1;
    this.eta = null;
  }
  angular.forEach(st.results, function(result) {
    var item = this.findItem_(result.clip.name);
    if (item == null) {
      return;
    }
    if (result.error) {
      item.state = cardcpx.importer.ItemState.ERROR;
    } else {
      item.state = cardcpx.importer.ItemState.IMPORTED;
    }
  }, this);
  if (st.pending && st.pending.length > 0) {
    var pendingItem = this.findItem_(st.pending[0].name);
    if (pendingItem) {
      pendingItem.state = cardcpx.importer.ItemState.ACTIVE;
    }
  }
};

/**
 * Find the item with the given clip name.
 * @param {string} clipName
 * @return {?Object}
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.findItem_ = function(clipName) {
  for (var i = 0; i < this.items.length; i++) {
    var item = this.items[i];
    if (item.clip.name == clipName) {
      return item;
    }
  }
  return null;
};

/**
 * Return the total size in bytes of the checked items.
 * @return {number}
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.checkedSize = function() {
  var n = 0;
  angular.forEach(this.items, function(item) {
    if (item.checked) {
      n += item.clip.totalSize;
    }
  });
  return n;
};

/**
 * Start import.
 * @export
 */
cardcpx.importer.ImporterCtrl.prototype.startImport = function() {
  this.importer_.startImport(this.path, this.subdir, this.getItemsToImport_());
  this.schedulePoll_();
};

/**
 * Return a list of items to import suitable for the importer RPC.
 * This will put select takes first in the list.
 * @return {Array.<Object>}
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.getItemsToImport_ = function() {
  var filteredItems = [];
  var sanitize = function(item) {
    return {
      clip: item.clip,
      scene: item.scene,
      num: item.num,
      select: item.select};
  };
  angular.forEach(this.items, function(item) {
    if (item.checked && item.select) {
      this.push(sanitize(item));
    }
  }, filteredItems);
  angular.forEach(this.items, function(item) {
    if (item.checked && !item.select) {
      this.push(sanitize(item));
    }
  }, filteredItems);
  return filteredItems;
};

/**
 * @private
 */
cardcpx.importer.ImporterCtrl.prototype.schedulePoll_ = function() {
  var imp = this;
  this.timeout_(function() {
    imp.refreshStatus_().then(function() {
      if (imp.progress >= 0) {
        imp.schedulePoll_();
      }
    });
  }, cardcpx.importer.ImporterCtrl.POLL_TIME_);
};
